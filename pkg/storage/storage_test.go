package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/beiai0xff/turl/pkg/db/mysql"
	"github.com/beiai0xff/turl/pkg/db/redis"
	"github.com/beiai0xff/turl/test"
)

func TestMain(m *testing.M) {
	db, _ := mysql.New(test.DSN)

	db.AutoMigrate(&TinyURL{})

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	db, _ := mysql.New(test.DSN)
	rdb := redis.Client(test.RedisAddr)

	s := New(db, rdb)
	t.Cleanup(func() {
		s.Close()
	})

	require.NotNil(t, s)
}

func Test_newStorage(t *testing.T) {
	db, _ := mysql.New(test.DSN)
	rdb := redis.Client(test.RedisAddr)

	s := newStorage(db, rdb)
	t.Cleanup(func() {
		s.Close()
	})

	require.NotNil(t, s)
}

func Test_storage_GetTinyURLByID(t *testing.T) {
	db, _ := mysql.New(test.DSN)
	rdb := redis.Client(test.RedisAddr)

	short, long := uint64(10000), []byte("www.google.com")
	s, ctx := newStorage(db, rdb), context.Background()
	t.Cleanup(func() { s.Close() })

	require.NoError(t, s.Insert(ctx, short, long))
	got, err := s.GetTinyURLByID(ctx, short)
	require.NoError(t, err)
	require.Equal(t, long, got.LongURL)

	got, err = s.GetTinyURLByID(ctx, 100)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
