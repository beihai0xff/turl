// Package turl implements the business logic of the tiny URL service.
package turl

import (
	"github.com/beiai0xff/turl/pkg/cache"
	"github.com/beiai0xff/turl/pkg/storage"
)

// TinyURL represents the tiny URL service.
type TinyURL struct {
	db               storage.Storage
	distributedCache cache.Interface
	localCache       cache.Interface
}

// New creates a new TinyURL service.
func newTinyURL(db storage.Storage, dcache, lcache cache.Interface) *TinyURL {
	return &TinyURL{
		db:               db,
		distributedCache: dcache,
		localCache:       lcache,
	}
}

// // Create creates a new tiny URL.
// func (t *TinyURL) Create(longURL []byte) error {
// 	return nil
// }
//
// // Retrieve a tiny URL.
// func (t *TinyURL) Retrieve(short string) error {
// 	return nil
// }
//
// func (t *TinyURL) Close() error {
// 	if err := t.db.Close(); err != nil {
// 		return err
// 	}
//
// 	if err := t.distributedCache.Close(); err != nil {
// 		return err
// 	}
//
// 	return t.localCache.Close()
// }
