package mem

import (
	"sync"

	gorepo "github.com/kgoins/go-repo"
	"github.com/kgoins/go-repo/codecs"
)

type MemRepo[T gorepo.Identifiable] struct {
	m     *sync.Map
	codec codecs.Codec[T]
}

var _ gorepo.Repo[gorepo.Identifiable] = &MemRepo[gorepo.Identifiable]{}

func NewRepo[T gorepo.Identifiable](c ...codecs.Codec[T]) MemRepo[T] {
	codec := codecs.NewDefaultCodec[T]()
	if len(c) > 0 {
		codec = c[0]
	}

	return MemRepo[T]{
		m:     &sync.Map{},
		codec: codec,
	}
}

func (s MemRepo[T]) Get(id string) (val T, found bool, err error) {
	rawValIface, found := s.m.Load(id)
	if !found {
		return val, false, nil
	}

	// Will always be []byte due to Add
	rawVal := rawValIface.([]byte)

	val, err = s.codec.Unmarshal(rawVal)
	return val, true, err
}

func (s MemRepo[T]) GetAll() ([]T, error) {
	vals := []T{}

	s.m.Range(func(key, value any) bool {
		// Will always be []byte due to Add
		valBytes := value.([]byte)
		val, err := s.codec.Unmarshal(valBytes)
		if err != nil {
			return false
		}

		vals = append(vals, val)
		return true
	})

	return vals, nil
}

func (s MemRepo[T]) Count() (int64, error) {
	i := int64(0)

	s.m.Range(func(key, value any) bool {
		i++
		return true
	})

	return i, nil
}

func (s MemRepo[T]) Add(v T) error {
	data, err := s.codec.Marshal(v)
	if err != nil {
		return err
	}

	s.m.Store(v.GetID(), data)
	return nil
}

func (s MemRepo[T]) Remove(id string) error {
	s.m.Delete(id)
	return nil
}

func (s MemRepo[T]) Close() error {
	s.m = nil // forces GC on internal map
	return nil
}
