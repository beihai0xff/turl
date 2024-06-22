// Package turl implements the business logic of the tiny URL service.
package turl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/beiai0xff/turl/configs"
	"github.com/beiai0xff/turl/pkg/cache"
	"github.com/beiai0xff/turl/pkg/db/mysql"
	"github.com/beiai0xff/turl/pkg/db/redis"
	"github.com/beiai0xff/turl/pkg/mapping"
	"github.com/beiai0xff/turl/pkg/storage"
	"github.com/beiai0xff/turl/pkg/tddl"
	"github.com/beiai0xff/turl/pkg/validate"
)

// TinyURL represents the tiny URL service.
type TinyURL struct {
	c     *configs.ServerConfig
	db    storage.Storage
	cache cache.Interface
	seq   tddl.TDDL
}

// NewTinyURL creates a new TinyURL service.
func NewTinyURL(c *configs.ServerConfig) (*TinyURL, error) {
	db, err := mysql.New(c.MySQLConfig)
	if err != nil {
		return nil, err
	}

	t, err := tddl.New(db, c.TDDLConfig)
	if err != nil {
		return nil, err
	}

	cacheProxy, err := cache.NewProxy(c.CacheConfig)
	if err != nil {
		return nil, err
	}

	rdb := redis.Client(c.CacheConfig.RedisConfig)

	return &TinyURL{
		c:     c,
		db:    storage.New(db, rdb),
		cache: cacheProxy,
		seq:   t,
	}, nil
}

// Create creates a new tiny URL.
func (t *TinyURL) Create(ctx context.Context, long []byte) ([]byte, error) {
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
	if err = t.cache.Set(ctx, string(short), long, t.c.CacheConfig.RemoteCacheTTL); err != nil {
		slog.ErrorContext(ctx, "failed to set cache", slog.Any("error", err))
	}

	return short, nil
}

// Retrieve a tiny URL.
func (t *TinyURL) Retrieve(ctx context.Context, short []byte) ([]byte, error) {
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
			if cerr := t.cache.Set(ctx, string(short), long, t.c.CacheConfig.RemoteCacheTTL); cerr != nil {
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

// Close closes the TinyURL service.
func (t *TinyURL) Close() error {
	t.seq.Close()

	if err := t.db.Close(); err != nil {
		return err
	}

	return t.cache.Close()
}
