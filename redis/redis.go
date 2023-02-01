package redisrepo

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	gorepo "github.com/kgoins/go-repo"
	"github.com/kgoins/go-repo/codecs"
)

type RedisRepo[T gorepo.Identifiable] struct {
	rdb   redis.UniversalClient
	codec codecs.Codec[T]
}

var _ gorepo.Repo[gorepo.Identifiable] = &RedisRepo[gorepo.Identifiable]{}

func NewRepo[T gorepo.Identifiable](rdb redis.UniversalClient, c ...codecs.Codec[T]) *RedisRepo[T] {
	codec := codecs.NewDefaultCodec[T]()
	if len(c) > 0 {
		codec = c[0]
	}

	return &RedisRepo[T]{rdb, codec}
}

func (r *RedisRepo[T]) Close() error {
	return r.rdb.Close()
}

func (r *RedisRepo[T]) getCtx() context.Context {
	return context.Background()
}

func (r *RedisRepo[T]) isNotFound(err error) bool {
	return err == redis.Nil
}

func (r *RedisRepo[T]) deserialize(val string) (T, error) {
	return r.codec.Unmarshal([]byte(val))
}

func (r *RedisRepo[T]) serialize(val T) (string, error) {
	valBytes, err := r.codec.Marshal(val)
	if err != nil {
		return "", err
	}

	return string(valBytes), nil
}

func (r *RedisRepo[T]) GetAll() ([]T, error) {
	keys, err := r.rdb.Keys(r.getCtx(), "*").Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []T{}, nil
	}

	results, err := r.rdb.MGet(r.getCtx(), keys...).Result()
	if err != nil {
		return nil, err
	}

	vals := make([]T, 0, len(keys))
	for _, rawResult := range results {
		rStr, worked := rawResult.(string)
		if !worked {
			return nil, errors.New("unknown value retured from redis")
		}

		t, err := r.deserialize(rStr)
		if err != nil {
			return nil, err
		}

		vals = append(vals, t)
	}

	return vals, nil
}

func (r *RedisRepo[T]) Get(id string) (T, bool, error) {
	var t T

	val, err := r.rdb.Get(r.getCtx(), id).Result()
	if err != nil {
		if r.isNotFound(err) {
			return t, false, nil
		}
		return t, false, err
	}

	t, err = r.deserialize(val)
	return t, true, err
}

func (r *RedisRepo[T]) Count() (int64, error) {
	val, err := r.rdb.DBSize(r.getCtx()).Result()
	return val, err
}

func (r *RedisRepo[T]) Add(val T) error {
	valStr, err := r.serialize(val)
	if err != nil {
		return err
	}

	return r.rdb.Set(
		r.getCtx(),
		val.GetID(),
		valStr,
		0,
	).Err()
}

func (r *RedisRepo[T]) Remove(id string) error {
	_, err := r.rdb.Del(r.getCtx(), id).Result()
	if err != nil {
		if r.isNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}
