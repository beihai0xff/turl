package turl

import (
	"os"
	"testing"

	"github.com/beiai0xff/turl/internal/tests"
	"github.com/beiai0xff/turl/pkg/storage"
	"github.com/beiai0xff/turl/pkg/tddl"
)

func TestMain(m *testing.M) {
	if err := tests.CreateTable(tddl.Sequence{}); err != nil {
		panic(err)
	}
	if err := tests.CreateTable(storage.TinyURL{}); err != nil {
		panic(err)
	}
	defer func() {
		tests.DropTDDLTable(tddl.Sequence{})
		tests.DropTDDLTable(storage.TinyURL{})
	}()

	os.Exit(m.Run())
}
