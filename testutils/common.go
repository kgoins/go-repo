package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	gorepo "github.com/kgoins/go-repo.git"
)

type Foo struct {
	Bar string
	id  string
}

func (f Foo) GetID() string {
	return f.id
}

// Tests basic CRUD ops against the given repo
func TestRepo(repo gorepo.Repo[Foo], t *testing.T) {
	r := require.New(t)
	key := "1880"

	_, found, err := repo.Get(key)
	r.NoError(err)
	r.False(found, "A value was found, but no value was expected")

	err = repo.Remove(key)
	r.NoError(err)

	val := Foo{
		Bar: "baz",
		id:  key,
	}

	err = repo.Add(val)
	r.NoError(err)

	err = repo.Add(val)
	r.NoError(err)

	expected := val
	retVal, found, err := repo.Get(key)
	r.NoError(err)
	r.True(found, "No value was found, but should have been")
	r.Equal(expected.Bar, retVal.Bar)
	r.Empty(retVal.id) // private vars not expected to be retained

	err = repo.Remove(key)
	r.NoError(err)

	_, found, err = repo.Get(key)
	r.NoError(err)
	r.False(found, "A value was found, but no value was expected")
}
