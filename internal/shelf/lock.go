package shelf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	writeLockFilename      = ".write.lock"
	writeLockPollInterval  = 25 * time.Millisecond
	writeLockStaleDuration = 10 * time.Minute
)

type WriteLock struct {
	path string
}

func AcquireWriteLock(rootDir string, timeout time.Duration) (*WriteLock, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	lockPath := filepath.Join(rootDir, ".shelf", writeLockFilename)
	deadline := time.Now().Add(timeout)

	for {
		if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create lock directory: %w", err)
		}
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = fmt.Fprintf(f, "pid=%d\ncreated_at=%s\n", os.Getpid(), time.Now().Local().Format(time.RFC3339))
			_ = f.Close()
			return &WriteLock{path: lockPath}, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("failed to acquire write lock: %w", err)
		}
		if stale, staleErr := isStaleLock(lockPath); staleErr == nil && stale {
			_ = os.Remove(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("write lock timeout: %s", lockPath)
		}
		time.Sleep(writeLockPollInterval)
	}
}

func (l *WriteLock) Release() error {
	if l == nil || strings.TrimSpace(l.path) == "" {
		return nil
	}
	err := os.Remove(l.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func isStaleLock(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "created_at=") {
			raw := strings.TrimSpace(strings.TrimPrefix(line, "created_at="))
			ts, parseErr := time.Parse(time.RFC3339, raw)
			if parseErr != nil {
				return true, nil
			}
			return time.Since(ts) > writeLockStaleDuration, nil
		}
	}
	_, statErr := os.Stat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return false, nil
		}
		return false, statErr
	}
	return true, nil
}
