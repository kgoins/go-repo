package mem_test

import (
	"testing"

	"github.com/kgoins/go-repo.git/mem"
	"github.com/kgoins/go-repo.git/testutils"
)

func TestMemRepo(t *testing.T) {
	memRepo := mem.NewRepo[testutils.Foo]()
	defer memRepo.Close()

	testutils.TestRepo(memRepo, t)
}
