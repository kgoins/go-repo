package codecs

import "encoding/json"

type JSONCodec[T any] struct{}

// Ensure JSONCodec implements Codec
var _ Codec[interface{}] = JSONCodec[interface{}]{}

func (JSONCodec[T]) Marshal(val T) ([]byte, error) {
	valBytes, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	return valBytes, nil
}

func (JSONCodec[T]) Unmarshal(data []byte) (T, error) {
	var val T
	err := json.Unmarshal([]byte(data), &val)

	return val, err
}
