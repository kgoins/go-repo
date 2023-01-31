package codecs

type Codec[T any] interface {
	Marshal(val T) ([]byte, error)
	Unmarshal(data []byte) (T, error)
}

func NewDefaultCodec[T any]() Codec[T] {
	return JSONCodec[T]{}
}
