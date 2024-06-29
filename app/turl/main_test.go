package turl

import (
	"os"
	"testing"

	"github.com/beihai0xff/turl/internal/tests"
	"github.com/beihai0xff/turl/pkg/storage"
	"github.com/beihai0xff/turl/pkg/tddl"
)

func TestMain(m *testing.M) {
	if err := tests.CreateTable(tddl.Sequence{}); err != nil {
		panic(err)
	}
	if err := tests.CreateTable(storage.TinyURL{}); err != nil {
		panic(err)
	}

	exitCode := m.Run()

	tests.DropTable(tddl.Sequence{})
	tests.DropTable(storage.TinyURL{})

	os.Exit(exitCode)
}
