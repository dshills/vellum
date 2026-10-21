//go:build darwin

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// The lock inode is permanent: unlinking it could let a later process lock a
// different inode while an earlier process still owns the old one. Process
// death releases flock automatically, leaving the last recovery copy readable.
func lockRecoveryDirectory(directory string) (*os.File, error) {
	path := filepath.Join(directory, recoveryLockFilename)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		_ = f.Close()
		return nil, errors.New("recovery: private regular lock file required")
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, errors.New("recovery: directory is in use by another app instance")
		}
		return nil, fmt.Errorf("recovery: directory lock: %w", err)
	}
	return f, nil
}
