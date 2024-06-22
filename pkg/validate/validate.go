// Package validate warp the validator package, and provides a singleton instance of validator.
package validate

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	once sync.Once
	v    *validator.Validate
)

// Instance returns a singleton instance of validator.
func Instance() *validator.Validate {
	once.Do(func() {
		v = validator.New(validator.WithRequiredStructEnabled())
	})

	return v
}
