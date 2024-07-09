// Package storage provides the implementation of the storage interface.
package storage

import (
	"context"

	"gorm.io/gorm"
)

// Ensuring that *storage implements the Storage interface
var _ Storage = (*storage)(nil)

// Storage is an interface that defines the methods that a storage system must implement.
type Storage interface {
	// Insert adds a new TinyURL record to the storage.
	Insert(ctx context.Context, short uint64, longURL []byte) (*TinyURL, error)
	// GetByLongURL retrieves a TinyURL record by its original URL.
	GetByLongURL(ctx context.Context, long []byte) (*TinyURL, error)
	// GetByShortID retrieves a TinyURL record by its short ID.
	GetByShortID(ctx context.Context, short uint64) (*TinyURL, error)
	// Delete a short link by short id
	Delete(ctx context.Context, short uint64) error
	// Close closes the storage.
	Close() error
}

// TinyURL represents a shortened URL record.
type TinyURL struct {
	gorm.Model
	LongURL []byte `gorm:"type:VARCHAR(500);uniqueIndex;not null" json:"long_url"` // The original URL.
	Short   uint64 `gorm:"type:BIGINT;uniqueIndex;not null" json:"short"`          // The shortened URL ID.
}

// TableName returns the table name of the TinyURL model.
func (TinyURL) TableName() string {
	return "tiny_urls"
}

// storage is a concrete implementation of the Storage interface.
type storage struct {
	db *gorm.DB // Database client.
}

// New creates a new storage instance.
func New(db *gorm.DB) Storage {
	return newStorage(db)
}

// newStorage is a helper function that creates a new storage instance.
func newStorage(db *gorm.DB) *storage {
	return &storage{
		db: db,
	}
}

// Insert adds a new TinyURL record to the storage.
func (s *storage) Insert(ctx context.Context, short uint64, long []byte) (*TinyURL, error) {
	t := TinyURL{
		Short:   short,
		LongURL: long,
	}

	// Create a new record in the database.
	if err := s.db.WithContext(ctx).Create(&t).Error; err != nil {
		return nil, err
	}

	return &t, nil
}

// GetByShortID retrieves a TinyURL record by its short ID.
func (s *storage) GetByShortID(ctx context.Context, short uint64) (*TinyURL, error) {
	t := TinyURL{}
	// Query the database for the record.
	res := s.db.WithContext(ctx).Where("short = ?", short).Take(&t)

	if res.Error != nil {
		return nil, res.Error
	}

	return &t, nil
}

// GetByLongURL retrieves a TinyURL record by its original URL.
func (s *storage) GetByLongURL(ctx context.Context, long []byte) (*TinyURL, error) {
	t := TinyURL{}
	// Query the database for the record.
	res := s.db.WithContext(ctx).Where("long_url = ?", long).Take(&t)

	if res.Error != nil {
		return nil, res.Error
	}

	return &t, nil
}

// Delete a short link by short id
func (s *storage) Delete(ctx context.Context, short uint64) error {
	res := s.db.WithContext(ctx).Where("short = ?", short).Delete(&TinyURL{})

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Close closes the storage.
func (s *storage) Close() error {
	return nil
}
