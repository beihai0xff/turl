// Package pkg provides the implementation of the storage interface.
package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Ensuring that *storage implements the Storage interface
var _ Storage = (*storage)(nil)

// Storage is an interface that defines the methods that a storage system must implement.
type Storage interface {
	// Insert adds a new TinyURL record to the storage.
	Insert(ctx context.Context, short uint64, longURL []byte) error
	// GetTinyURLByID retrieves a TinyURL record by its ID.
	GetTinyURLByID(ctx context.Context, short uint64) (*TinyURL, error)
	// Close closes the storage.
	Close() error
}

// TinyURL represents a shortened URL record.
type TinyURL struct {
	gorm.Model
	LongURL []byte `gorm:"type:VARCHAR(500);not null" json:"long_url"` // The original URL.
	Short   uint64 `gorm:"type:BIGINT;index;not null" json:"short"`    // The shortened URL ID.
}

// storage is a concrete implementation of the Storage interface.
type storage struct {
	db  *gorm.DB              // Database client.
	rdb redis.UniversalClient // Redis client.
}

// New creates a new storage instance.
func New(db *gorm.DB, rdb redis.UniversalClient) Storage {
	return newStorage(db, rdb)
}

// newStorage is a helper function that creates a new storage instance.
func newStorage(db *gorm.DB, rdb redis.UniversalClient) *storage {
	return &storage{
		db:  db,
		rdb: rdb,
	}
}

// Insert adds a new TinyURL record to the storage.
func (s *storage) Insert(ctx context.Context, short uint64, long []byte) error {
	t := TinyURL{
		Short:   short,
		LongURL: long,
	}

	// Create a new record in the database.
	return s.db.WithContext(ctx).Create(&t).Error
}

// GetTinyURLByID retrieves a TinyURL record by its ID.
func (s *storage) GetTinyURLByID(ctx context.Context, short uint64) (*TinyURL, error) {
	t := TinyURL{}
	// Query the database for the record.
	res := s.db.WithContext(ctx).Where("short = ?", short).Take(&t)

	if res.Error != nil {
		return nil, res.Error
	}

	return &t, nil
}

// Close closes the storage.
func (s *storage) Close() error {
	return s.rdb.Close()
}
