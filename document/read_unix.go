//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package document

import (
	"os"
	"syscall"
)

// Nonblocking open prevents a replaced FIFO from trapping the worker before
// Read's regular-file check. This flag does not interrupt regular-file IO.
func openReadFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
