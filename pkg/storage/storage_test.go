package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/internal/tests"
	"github.com/beihai0xff/turl/pkg/db/mysql"
)

func TestMain(m *testing.M) {
	tests.CreateTable(&TinyURL{})

	code := m.Run()
	tests.DropTable(&TinyURL{})

	os.Exit(code)
}

func TestNew(t *testing.T) {
	db, _ := mysql.New(&configs.MySQLConfig{DSN: tests.DSN})

	s := New(db)
	t.Cleanup(func() {
		s.Close()
	})

	require.NotNil(t, s)
}

func Test_newStorage(t *testing.T) {
	db, _ := mysql.New(&configs.MySQLConfig{DSN: tests.DSN})

	s := newStorage(db)
	t.Cleanup(func() {
		s.Close()
	})

	require.NotNil(t, s)
}

func Test_storage_GetTinyURLByID(t *testing.T) {
	db, _ := mysql.New(&configs.MySQLConfig{DSN: tests.DSN})

	short, long := uint64(20000), []byte("www.google.com")
	s, ctx := newStorage(db), context.Background()
	t.Cleanup(func() { s.Close() })

	require.NoError(t, s.Insert(ctx, short, long))
	got, err := s.GetTinyURLByID(ctx, short)
	require.NoError(t, err)
	require.Equal(t, long, got.LongURL)

	got, err = s.GetTinyURLByID(ctx, 100)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
