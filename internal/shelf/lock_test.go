package shelf

import (
	"testing"
	"time"
)

func TestAcquireWriteLockAcquireAndRelease(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	lock, err := AcquireWriteLock(root, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire lock failed: %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("release lock failed: %v", err)
	}

	lock2, err := AcquireWriteLock(root, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("re-acquire lock failed: %v", err)
	}
	if err := lock2.Release(); err != nil {
		t.Fatalf("release second lock failed: %v", err)
	}
}

func TestAcquireWriteLockTimeout(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	lock, err := AcquireWriteLock(root, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire lock failed: %v", err)
	}
	defer func() {
		_ = lock.Release()
	}()

	if _, err := AcquireWriteLock(root, 50*time.Millisecond); err == nil {
		t.Fatal("expected timeout while lock is held")
	}
}
