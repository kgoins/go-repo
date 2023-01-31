package gorepo

import "io"

type Identifiable interface {
	GetID() string
}

type Repo[T Identifiable] interface {
	io.Closer

	Get(id string) (T, bool, error)
	GetAll() ([]T, error)
	Count() (int64, error)

	// Duplicate Adds will overwrite with the latest value
	Add(entry T) error

	// Removing a non-existant id is a NOP
	Remove(id string) error
}
