// Package turl implements the business logic of the tiny URL service.
package turl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/cache"
	"github.com/beihai0xff/turl/pkg/db/mysql"
	"github.com/beihai0xff/turl/pkg/db/redis"
	"github.com/beihai0xff/turl/pkg/mapping"
	"github.com/beihai0xff/turl/pkg/storage"
	"github.com/beihai0xff/turl/pkg/tddl"
	"github.com/beihai0xff/turl/pkg/validate"
)

// Service represents the tiny URL service interface.
type Service interface {
	Create(ctx context.Context, long []byte) ([]byte, error)
	Retrieve(ctx context.Context, short []byte) ([]byte, error)
	Close() error
}

// tinyURLService represents the tiny URL service.
type tinyURLService struct {
	c     *configs.ServerConfig
	db    storage.Storage
	cache cache.Interface
	seq   tddl.TDDL
}

var _ Service = (*tinyURLService)(nil)

// newTinyURLService creates a new tinyURLService service.
func newTinyURLService(c *configs.ServerConfig) (*tinyURLService, error) {
	db, err := mysql.New(c.MySQL)
	if err != nil {
		return nil, err
	}

	if db.AutoMigrate(tddl.Sequence{}, storage.TinyURL{}) != nil {
		return nil, err
	}

	t, err := tddl.New(db, c.TDDL)
	if err != nil {
		return nil, err
	}

	cacheProxy, err := cache.NewProxy(c.Cache)
	if err != nil {
		return nil, err
	}

	rdb := redis.Client(c.Cache.Redis)

	return &tinyURLService{
		c:     c,
		db:    storage.New(db, rdb),
		cache: cacheProxy,
		seq:   t,
	}, nil
}

// Create creates a new tiny URL.
func (t *tinyURLService) Create(ctx context.Context, long []byte) ([]byte, error) {
	if err := validate.Instance().VarCtx(ctx, string(long), "required,http_url"); err != nil {
		return nil, err
	}

	seq, err := t.seq.Next(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sequence: %w", err)
	}

	if err = t.db.Insert(ctx, seq, long); err != nil {
		return nil, fmt.Errorf("failed to insert into db: %w", err)
	}

	short := mapping.Base58Encode(seq)

	// set local cache and distributed cache, if failed, just log the error, not return err
	if err = t.cache.Set(ctx, string(short), long, t.c.Cache.RemoteCacheTTL); err != nil {
		slog.ErrorContext(ctx, "failed to set cache", slog.Any("error", err))
	}

	return short, nil
}

// Retrieve a tiny URL.
func (t *tinyURLService) Retrieve(ctx context.Context, short []byte) ([]byte, error) {
	// validate short URL
	seq, err := mapping.Base58Decode(short)
	if err != nil {
		if errors.Is(err, mapping.ErrInvalidInput) {
			return nil, err
		}
	}

	// try to get from cache
	long, err := t.cache.Get(ctx, string(short))
	if err == nil {
		return long, nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		return nil, err
	}

	defer func() {
		if len(long) > 0 {
			// set local cache and distributed cache, if failed, just log the error, not return err
			if cerr := t.cache.Set(ctx, string(short), long, t.c.Cache.RemoteCacheTTL); cerr != nil {
				slog.ErrorContext(ctx, "failed to set cache", slog.Any("error", err))
			}
		}
	}()

	// try to get from db
	res, err := t.db.GetTinyURLByID(ctx, seq)
	if err != nil {
		return nil, err
	}

	long = res.LongURL

	return long, nil
}

// Close closes the tinyURLService service.
func (t *tinyURLService) Close() error {
	t.seq.Close()

	if err := t.db.Close(); err != nil {
		return err
	}

	return t.cache.Close()
}
