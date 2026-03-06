package cli

import (
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

const defaultWriteLockTimeout = 5 * time.Second

func withWriteLock(rootDir string, fn func() error) error {
	lock, err := shelf.AcquireWriteLock(rootDir, defaultWriteLockTimeout)
	if err != nil {
		return err
	}
	defer func() {
		_ = lock.Release()
	}()
	return fn()
}
