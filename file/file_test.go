package file_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kgoins/go-repo/file"
	"github.com/kgoins/go-repo/testutils"
)

func TestRepo(t *testing.T) {
	r := require.New(t)

	opts, err := file.NewOptions(
		"../testdata/filetest",
		"json",
	)
	r.NoError(err)
	r.True(filepath.IsAbs(opts.Directory))

	repo, err := file.NewRepo[testutils.Foo](opts)
	r.NoError(err)

	defer os.RemoveAll(opts.Directory)
	testutils.TestRepo(repo, t)
}

func TestRepoGetAll(t *testing.T) {
	r := require.New(t)

	opts, err := file.NewOptions(
		"../testdata/filetest",
		"json",
	)
	r.NoError(err)

	repo, err := file.NewRepo[testutils.Foo](opts)
	r.NoError(err)

	defer os.RemoveAll(opts.Directory)
	testutils.TestRepoGetAll(repo, t)
}
