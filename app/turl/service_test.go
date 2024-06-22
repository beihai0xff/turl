package turl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/beiai0xff/turl/internal/tests/mocks"
	"github.com/beiai0xff/turl/pkg/cache"
	"github.com/beiai0xff/turl/pkg/mapping"
	"github.com/beiai0xff/turl/pkg/storage"

	"github.com/beiai0xff/turl/configs"
	"github.com/beiai0xff/turl/internal/tests"
)

var testConfig = &configs.ServerConfig{
	Listen: "localhost",
	Port:   8080,
	TDDLConfig: &configs.TDDLConfig{
		Step:     100,
		StartNum: 10000,
		SeqName:  "tiny_url",
	},
	MySQLConfig: &configs.MySQLConfig{
		DSN: tests.DSN,
	},
	CacheConfig: &configs.CacheConfig{
		LocalCacheSize: 10,
		LocalCacheTTL:  10,
		RedisConfig: &configs.RedisConfig{
			Addr:        tests.RedisAddr,
			DialTimeout: time.Second,
		},
	},
}

func TestTinyURL_Create(t *testing.T) {
	turl, err := NewTinyURL(testConfig)
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

func TestTinyURL_Retrieve(t *testing.T) {
	turl, err := NewTinyURL(testConfig)
	require.NoError(t, err)

	t.Run("RetrieveExistingURL", func(t *testing.T) {
		short, _ := turl.Create(context.Background(), []byte("https://www.example.com"))
		_, err := turl.Retrieve(context.Background(), short)
		require.NoError(t, err)
	})

	t.Run("RetrieveNonExistingURL", func(t *testing.T) {
		_, err := turl.Retrieve(context.Background(), []byte("non_existing_url"))
		require.Error(t, err)
	})
}

func TestTinyURL_Close(t *testing.T) {
	turl, err := NewTinyURL(testConfig)
	require.NoError(t, err)

	t.Run("CloseService", func(t *testing.T) {
		err := turl.Close()
		require.NoError(t, err)
	})
}

func TestTinyURL_Create_failed(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTDDL, mockCache, mockStorage := mocks.NewMockTDDL(ctrl), mocks.NewMockCache(ctrl), mocks.NewMockStorage(ctrl)

	turl := &TinyURL{
		c:     testConfig,
		db:    mockStorage,
		cache: mockCache,
		seq:   mockTDDL,
	}

	testErr := errors.New("test error")

	t.Run("CreateFailedToGenerateSequence", func(t *testing.T) {
		mockTDDL.EXPECT().Next(gomock.Any()).Return(uint64(0), testErr)
		_, err := turl.Create(context.Background(), []byte("https://www.example.com"))
		require.ErrorIs(t, err, testErr)
	})

	t.Run("CreateFailedToInsertIntoDB", func(t *testing.T) {
		mockTDDL.EXPECT().Next(gomock.Any()).Return(uint64(1), nil)
		mockStorage.EXPECT().Insert(gomock.Any(), uint64(1), []byte("https://www.example.com")).Return(testErr)
		_, err := turl.Create(context.Background(), []byte("https://www.example.com"))
		require.ErrorIs(t, err, testErr)
	})

	t.Run("CreateFailedToSetCache", func(t *testing.T) {
		mockTDDL.EXPECT().Next(gomock.Any()).Return(uint64(1), nil)
		mockStorage.EXPECT().Insert(gomock.Any(), uint64(1), []byte("https://www.example.com")).Return(nil)
		mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testErr)
		_, err := turl.Create(context.Background(), []byte("https://www.example.com"))
		require.NoError(t, err)
	})
}

func TestTinyURL_Retrieve_failed(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTDDL, mockCache, mockStorage := mocks.NewMockTDDL(ctrl), mocks.NewMockCache(ctrl), mocks.NewMockStorage(ctrl)

	turl := &TinyURL{
		c:     testConfig,
		db:    mockStorage,
		cache: mockCache,
		seq:   mockTDDL,
	}

	testErr := errors.New("test error")

	t.Run("RetrieveFailedToDecodeShortURL", func(t *testing.T) {
		_, err := turl.Retrieve(context.Background(), []byte("invalid_short_url"))
		require.ErrorIs(t, err, mapping.ErrInvalidInput)

		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)
		mockStorage.EXPECT().GetTinyURLByID(gomock.Any(), gomock.Any()).Times(0)
		mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	})

	t.Run("RetrieveFailedToGetFromCache", func(t *testing.T) {
		mockCache.EXPECT().Get(gomock.Any(), "zzzzzz").Return(nil, testErr).Times(1)
		mockStorage.EXPECT().GetTinyURLByID(gomock.Any(), gomock.Any()).Times(0)
		mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		got, err := turl.Retrieve(context.Background(), []byte("zzzzzz"))
		require.ErrorIs(t, err, testErr)
		require.Nil(t, got)
	})

	t.Run("GetFailedToGetFromStorage", func(t *testing.T) {
		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, cache.ErrCacheMiss).Times(1)
		mockStorage.EXPECT().GetTinyURLByID(gomock.Any(), uint64(38068692543)).Return(nil, testErr).Times(1)
		mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		got, err := turl.Retrieve(context.Background(), []byte("zzzzzz"))
		require.ErrorIs(t, err, testErr)
		require.Nil(t, got)
	})

	t.Run("RetrieveFailedToSetCache", func(t *testing.T) {
		mockCache.EXPECT().Get(gomock.Any(), "zzzzzz").Return(nil, cache.ErrCacheMiss).Times(1)
		mockStorage.EXPECT().GetTinyURLByID(gomock.Any(), uint64(38068692543)).Return(&storage.TinyURL{LongURL: []byte("https://www.example.com")}, nil).Times(1)
		mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testErr).Times(1)

		got, err := turl.Retrieve(context.Background(), []byte("zzzzzz"))
		require.NoError(t, err)
		require.Equal(t, []byte("https://www.example.com"), got)
	})
}
