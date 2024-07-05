package turl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/beihai0xff/turl/internal/tests"
	"github.com/beihai0xff/turl/internal/tests/mocks"
	"github.com/beihai0xff/turl/pkg/cache"
	"github.com/beihai0xff/turl/pkg/mapping"
	"github.com/beihai0xff/turl/pkg/storage"
	"github.com/beihai0xff/turl/pkg/tddl"
)

func TestService_Create(t *testing.T) {
	turl, err := newService(tests.GlobalConfig)
	require.NoError(t, err)

	t.Run("CreateNewURL", func(t *testing.T) {
		short, err := turl.Create(context.Background(), []byte("https://www.example.com"))
		require.NoError(t, err)
		require.NotNil(t, short)
	})

	t.Run("CreateInvalidURL", func(t *testing.T) {
		short, err := turl.Create(context.Background(), []byte("invalid_url"))
		require.Error(t, err)
		require.Nil(t, short)
	})
}

func TestService_Retrieve(t *testing.T) {
	require.NoError(t, tests.CreateTable(tddl.Sequence{}))
	require.NoError(t, tests.CreateTable(storage.TinyURL{}))

	turl, err := newService(tests.GlobalConfig)
	require.NoError(t, err)

	t.Run("RetrieveExistingURL", func(t *testing.T) {
		short, err := turl.Create(context.Background(), []byte("https://www.example.com"))
		require.NoError(t, err)
		got, err := turl.Retrieve(context.Background(), short)
		require.NoError(t, err)
		require.Equal(t, []byte("https://www.example.com"), got)
	})

	t.Run("RetrieveNonExistingURL", func(t *testing.T) {
		got, err := turl.Retrieve(context.Background(), []byte("zzzzzz"))
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		require.Nil(t, got)
	})
}

func TestService_Close(t *testing.T) {
	turl, err := newService(tests.GlobalConfig)
	require.NoError(t, err)

	t.Run("CloseService", func(t *testing.T) {
		err = turl.Close()
		require.NoError(t, err)
	})
}

func TestService_Create_failed(t *testing.T) {
	mockTDDL, mockCache, mockStorage := mocks.NewMockTDDL(t), mocks.NewMockCache(t), mocks.NewMockStorage(t)

	turl := &commandService{
		ttl:   time.Second,
		db:    mockStorage,
		cache: mockCache,
		seq:   mockTDDL,
	}

	testErr := errors.New("test error")

	t.Run("CreateFailedToGenerateSequence", func(t *testing.T) {
		mockTDDL.EXPECT().Next(mock.Anything).Return(uint64(0), testErr).Times(1)
		_, err := turl.Create(context.Background(), []byte("https://www.example.com"))
		require.ErrorIs(t, err, testErr)
	})

	t.Run("CreateFailedToInsertIntoDB", func(t *testing.T) {
		mockTDDL.EXPECT().Next(mock.Anything).Return(uint64(1), nil).Times(1)
		mockStorage.EXPECT().Insert(mock.Anything, uint64(1), []byte("https://www.example.com")).Return(testErr).Times(1)
		_, err := turl.Create(context.Background(), []byte("https://www.example.com"))
		require.ErrorIs(t, err, testErr)
	})

	t.Run("CreateFailedToSetCache", func(t *testing.T) {
		mockTDDL.EXPECT().Next(mock.Anything).Return(uint64(1), nil).Times(1)
		mockStorage.EXPECT().Insert(mock.Anything, uint64(1), []byte("https://www.example.com")).Return(nil)
		mockCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testErr).Times(1)
		_, err := turl.Create(context.Background(), []byte("https://www.example.com"))
		require.NoError(t, err)
	})
}

func TestService_Retrieve_failed(t *testing.T) {
	mockCache, mockStorage := mocks.NewMockCache(t), mocks.NewMockStorage(t)

	turl := &queryService{
		ttl:   time.Second,
		db:    mockStorage,
		cache: mockCache,
	}

	testErr := errors.New("test error")

	t.Run("RetrieveFailedToDecodeShortURL", func(t *testing.T) {
		got, err := turl.Retrieve(context.Background(), []byte("invalid_short_url"))
		require.ErrorIs(t, err, mapping.ErrInvalidInput)
		require.Nil(t, got)
	})

	t.Run("RetrieveFailedToGetFromCache", func(t *testing.T) {
		mockCache.EXPECT().Get(mock.Anything, "zzzzzz").Return(nil, testErr).Times(1)

		got, err := turl.Retrieve(context.Background(), []byte("zzzzzz"))
		require.ErrorIs(t, err, testErr)
		require.Nil(t, got)
	})

	t.Run("GetFailedToGetFromStorage", func(t *testing.T) {
		mockCache.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, cache.ErrCacheMiss).Times(1)
		mockStorage.EXPECT().GetTinyURLByID(mock.Anything, uint64(38068692543)).Return(nil, testErr).Times(1)

		got, err := turl.Retrieve(context.Background(), []byte("zzzzzz"))
		require.ErrorIs(t, err, testErr)
		require.Nil(t, got)
	})

	t.Run("RetrieveFailedToSetCache", func(t *testing.T) {
		mockCache.EXPECT().Get(mock.Anything, "zzzzzz").Return(nil, cache.ErrCacheMiss).Times(1)
		mockStorage.EXPECT().GetTinyURLByID(mock.Anything, uint64(38068692543)).Return(&storage.TinyURL{LongURL: []byte("https://www.example.com")}, nil).Times(1)
		mockCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testErr).Times(1)

		got, err := turl.Retrieve(context.Background(), []byte("zzzzzz"))
		require.NoError(t, err)
		require.Equal(t, []byte("https://www.example.com"), got)
	})
}

func Test_getDB(t *testing.T) {
	t.Run("GetDBSuccess", func(t *testing.T) {
		_, err := getDB(tests.GlobalConfig)
		require.NoError(t, err)
	})
	t.Run("GetDBDebug", func(t *testing.T) {
		tests.GlobalConfig.Debug = true
		_, err := getDB(tests.GlobalConfig)
		require.NoError(t, err)
	})

	t.Run("GetDBFailed", func(t *testing.T) {
		tests.GlobalConfig.MySQL.DSN = "invalid_dsn"
		_, err := getDB(tests.GlobalConfig)
		require.Error(t, err)
	})
}
