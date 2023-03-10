package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	gorepo "github.com/kgoins/go-repo"
)

type Foo struct {
	ID  string
	Bar string
	baz string
}

func (f Foo) GetID() string {
	return f.ID
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
		ID:  key,
		Bar: "baz",
		baz: "futz",
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
	r.NotEmpty(retVal.GetID())
	r.Empty(retVal.baz) // private vars not expected to be retained

	err = repo.Remove(key)
	r.NoError(err)

	_, found, err = repo.Get(key)
	r.NoError(err)
	r.False(found, "A value was found, but no value was expected")
}

// Tests basic CRUD ops against the given repo
func TestRepoGetAll(repo gorepo.Repo[Foo], t *testing.T) {
	r := require.New(t)

	f1 := Foo{ID: "1"}
	f2 := Foo{ID: "2"}
	f3 := Foo{ID: "3"}

	repo.Add(f1)
	repo.Add(f2)
	repo.Add(f3)

	count, err := repo.Count()
	r.NoError(err)
	r.Equal(int64(3), count)

	vals, err := repo.GetAll()
	r.NoError(err)
	r.Equal(int64(3), int64(len(vals)))
}
