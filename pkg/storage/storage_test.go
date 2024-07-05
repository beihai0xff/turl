package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

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
	db, _ := mysql.New(tests.GlobalConfig.MySQL)

	s := New(db)
	t.Cleanup(func() {
		s.Close()
	})

	require.NotNil(t, s)
}

func Test_newStorage(t *testing.T) {
	db, _ := mysql.New(tests.GlobalConfig.MySQL)

	s := newStorage(db)
	t.Cleanup(func() {
		s.Close()
	})

	require.NotNil(t, s)
}

func Test_storage_Insert(t *testing.T) {
	db, _ := mysql.New(tests.GlobalConfig.MySQL)

	long := []byte("www.Insert.com")
	s, ctx := newStorage(db), context.Background()
	t.Cleanup(func() { s.Close() })

	t.Run("Insert", func(t *testing.T) {
		require.NoError(t, s.Insert(ctx, uint64(20000), long))
	})

	t.Run("InsertDuplicateURL", func(t *testing.T) {
		require.ErrorIs(t, s.Insert(ctx, uint64(30000), long), gorm.ErrDuplicatedKey)
	})

	t.Run("InsertDuplicateShort", func(t *testing.T) {
		require.ErrorIs(t, s.Insert(ctx, uint64(20000), []byte("www.InsertDuplicateShort.com")), gorm.ErrDuplicatedKey)
	})
}

func Test_storage_GetTinyURLByID(t *testing.T) {
	db, _ := mysql.New(tests.GlobalConfig.MySQL)

	short, long := uint64(40000), []byte("www.GetByShortID.com")
	s, ctx := newStorage(db), context.Background()
	t.Cleanup(func() { s.Close() })

	require.NoError(t, s.Insert(ctx, short, long))
	got, err := s.GetByShortID(ctx, short)
	require.NoError(t, err)
	require.Equal(t, long, got.LongURL)

	got, err = s.GetByShortID(ctx, 100)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func Test_storage_GetByLongURL(t *testing.T) {
	db, _ := mysql.New(tests.GlobalConfig.MySQL)

	long := []byte("www.GetByLongURL.com")
	s, ctx := newStorage(db), context.Background()
	t.Cleanup(func() { s.Close() })

	t.Run("GetByLongURL", func(t *testing.T) {
		require.NoError(t, s.Insert(ctx, uint64(50000), long))

		got, err := s.GetByLongURL(ctx, long)
		require.NoError(t, err)
		require.Equal(t, long, got.LongURL)
		require.Equal(t, uint64(50000), got.Short)
	})

	t.Run("GetByLongURLNotFound", func(t *testing.T) {
		got, err := s.GetByLongURL(ctx, []byte("www.GetByLongURLNotFound.com"))
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		require.Nil(t, got)
	})
}
