// Package turl implements the business logic of the tiny URL service.
package turl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/beihai0xff/turl/app/turl/model"
	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/cache"
	"github.com/beihai0xff/turl/pkg/db/mysql"
	"github.com/beihai0xff/turl/pkg/mapping"
	"github.com/beihai0xff/turl/pkg/storage"
	"github.com/beihai0xff/turl/pkg/tddl"
	"github.com/beihai0xff/turl/pkg/validate"
)

// Service represents the tiny URL service interface.
type Service interface {
	Create(ctx context.Context, long []byte) (*model.TinyURL, error)
	GetByLong(ctx context.Context, long []byte) (*model.TinyURL, error)
	Retrieve(ctx context.Context, short []byte) ([]byte, error)
	Delete(ctx context.Context, short []byte) error
	Close() error
}

var _ Service = (*service)(nil)

type service struct {
	*commandService
	*queryService
}

func getDB(c *configs.ServerConfig) (*gorm.DB, error) {
	db, err := mysql.New(c.MySQL)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	} else if c.Debug {
		go func() {
			for range time.NewTicker(time.Second).C {
				slog.Info(fmt.Sprintf("mysql db stats %+v", sqlDB.Stats()))
			}
		}()
	}

	return db, nil
}

// newService creates a new commandService service.
func newService(c *configs.ServerConfig) (*service, error) {
	db, err := getDB(c)
	if err != nil {
		return nil, err
	}

	cacheProxy, err := cache.NewProxy(c.Cache)
	if err != nil {
		return nil, err
	}

	if c.Readonly {
		return &service{
			queryService: &queryService{
				ttl:   c.Cache.Redis.TTL,
				db:    storage.New(db),
				cache: cacheProxy,
			},
		}, nil
	}

	if db.AutoMigrate(tddl.Sequence{}, storage.TinyURL{}) != nil {
		return nil, err
	}

	t, err := tddl.New(db, c.TDDL)
	if err != nil {
		return nil, err
	}

	writeCacheProxy, err := cache.NewProxy(c.Cache)
	if err != nil {
		return nil, err
	}

	return &service{
		commandService: &commandService{
			ttl:   c.Cache.Redis.TTL,
			db:    storage.New(db),
			cache: writeCacheProxy,
			seq:   t,
		},
		queryService: &queryService{
			ttl:   c.Cache.Redis.TTL,
			db:    storage.New(db),
			cache: cacheProxy,
		},
	}, nil
}

// Close closes the command service.
func (s *service) Close() error {
	if s.commandService != nil {
		if err := s.commandService.Close(); err != nil {
			return err
		}
	}

	if s.queryService != nil {
		return s.queryService.Close()
	}

	return nil
}

// commandService represents the tiny URL service.
type commandService struct {
	ttl   time.Duration
	db    storage.Storage
	cache cache.Interface
	seq   tddl.TDDL
}

// Create creates a new tiny URL.
func (c *commandService) Create(ctx context.Context, long []byte) (*model.TinyURL, error) {
	if err := validate.Instance().VarCtx(ctx, string(long), "required,http_url"); err != nil {
		return nil, err
	}

	seq, err := c.seq.Next(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sequence: %w", err)
	}

	record, err := c.db.Insert(ctx, seq, long)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			slog.Error(fmt.Sprintf("failed to insert into db: %v, try to get from db", err),
				slog.Any("long url", long), slog.Int64("seq", int64(seq)))

			record, err = c.db.GetByLongURL(ctx, long)
			if err != nil {
				return nil, fmt.Errorf("failed to get from db: %w", err)
			}

			seq = record.Short
		} else {
			return nil, fmt.Errorf("failed to insert into db: %w", err)
		}
	}

	short := mapping.Base58Encode(seq)
	// set local cache and distributed cache, if failed, just log the error, not return err
	if err = c.cache.Set(ctx, string(short), long, c.ttl); err != nil {
		slog.ErrorContext(ctx, "failed to set cache", slog.Any("error", err))
	}

	return &model.TinyURL{
		ShortURL:  string(short),
		LongURL:   string(long),
		CreatedAt: record.CreatedAt,
		DeletedAt: record.DeletedAt,
	}, nil
}

func (c *commandService) Delete(ctx context.Context, short []byte) error {
	// decode and validate short URI
	seq, err := mapping.Base58Decode(short)
	if err != nil {
		return err
	}

	if err = c.db.Delete(ctx, seq); err != nil {
		return err
	}

	return c.cache.Del(ctx, string(short))
}

// Close closes the command service.
func (c *commandService) Close() error {
	c.seq.Close()

	if err := c.db.Close(); err != nil {
		return err
	}

	return c.cache.Close()
}

// queryService represents the query service.
type queryService struct {
	ttl   time.Duration
	db    storage.Storage
	cache cache.Interface
}

// Retrieve a tiny URL.
func (q *queryService) Retrieve(ctx context.Context, short []byte) ([]byte, error) {
	// decode and validate short URI
	seq, err := mapping.Base58Decode(short)
	if err != nil {
		return nil, err
	}

	// try to get from cache
	long, err := q.cache.Get(ctx, string(short))
	if err == nil {
		return long, nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		return nil, err
	}

	defer func() {
		if len(long) > 0 {
			// set local cache and distributed cache, if failed, just log the error, not return err
			if cerr := q.cache.Set(ctx, string(short), long, q.ttl); cerr != nil {
				slog.ErrorContext(ctx, "failed to set cache", slog.Any("error", err))
			}
		}
	}()

	// try to get from db
	res, err := q.db.GetByShortID(ctx, seq)
	if err != nil {
		return nil, err
	}

	long = res.LongURL

	return long, nil
}

// GetByLong returns the tiny URL by the long URL.
func (q *queryService) GetByLong(ctx context.Context, long []byte) (*model.TinyURL, error) {
	if err := validate.Instance().VarCtx(ctx, string(long), "required,http_url"); err != nil {
		return nil, err
	}

	record, err := q.db.GetByLongURL(ctx, long)
	if err != nil {
		return nil, err
	}

	return &model.TinyURL{
		ShortURL:  string(mapping.Base58Encode(record.Short)),
		LongURL:   string(record.LongURL),
		CreatedAt: record.CreatedAt,
		DeletedAt: record.DeletedAt,
	}, nil
}

// Close closes the command service.
func (q *queryService) Close() error {
	if err := q.db.Close(); err != nil {
		return err
	}

	return q.cache.Close()
}
