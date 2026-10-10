//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package document

import (
	"context"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestFIFODoesNotBlockBeforeRegularFileCheck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fifo")
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := (FileStore{}).Read(context.Background(), path, 32); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("FIFO accepted")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FIFO open blocked")
	}
}
