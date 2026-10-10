package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var ErrTooLarge = errors.New("document: byte limit exceeded")

// Store implementations are worker-safe. Write reports whether replacement
// occurred separately from its error; a post-rename error cannot undo a save.
type Store interface {
	Read(context.Context, string, int64) ([]byte, error)
	Write(context.Context, string, []byte, int64) (bool, error)
}

type FileStore struct{}

func validateLimit(max int64) error {
	if max < 1 || max > MaximumMaxBytes {
		return fmt.Errorf("document: byte limit must be 1..%d", MaximumMaxBytes)
	}
	return nil
}
func validatePath(path string) error {
	if !filepath.IsAbs(path) || len(path) > 1<<20 || !utf8.ValidString(path) || strings.ContainsRune(path, 0) {
		return fmt.Errorf("document: absolute valid file path required")
	}
	return nil
}
func fileAdmission(ctx context.Context, path string, max int64) error {
	if ctx == nil {
		return fmt.Errorf("document: context required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateLimit(max); err != nil {
		return err
	}
	return validatePath(path)
}
func (FileStore) Read(ctx context.Context, path string, max int64) ([]byte, error) {
	if err := fileAdmission(ctx, path, max); err != nil {
		return nil, err
	}
	f, err := openReadFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("document: regular file required")
	}
	if info.Size() > max {
		return nil, ErrTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if int64(len(data)) > max {
		return nil, ErrTooLarge
	}
	return data, nil
}

type atomicFileOps struct {
	syncFile      func(*os.File) error
	rename        func(string, string) error
	syncDirectory func(*os.File) error
}

var defaultAtomicFileOps = atomicFileOps{syncFile: func(f *os.File) error { return f.Sync() }, rename: os.Rename, syncDirectory: func(f *os.File) error { return f.Sync() }}

func (FileStore) Write(ctx context.Context, path string, data []byte, max int64) (bool, error) {
	return writeAtomic(ctx, path, data, max, defaultAtomicFileOps)
}
func writeAtomic(ctx context.Context, path string, data []byte, max int64, ops atomicFileOps) (bool, error) {
	if err := fileAdmission(ctx, path, max); err != nil {
		return false, err
	}
	if int64(len(data)) > max {
		return false, ErrTooLarge
	}
	mode := os.FileMode(0600)
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return false, fmt.Errorf("document: save target must be a regular file")
		}
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return false, err
	}
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return false, err
	}
	defer directory.Close()
	f, err := os.CreateTemp(filepath.Dir(path), ".vellum-document-*")
	if err != nil {
		return false, err
	}
	temporary := f.Name()
	defer os.Remove(temporary)
	defer f.Close()
	if err = f.Chmod(mode); err != nil {
		return false, err
	}
	if err = ctx.Err(); err != nil {
		return false, err
	}
	if _, err = f.Write(data); err != nil {
		return false, err
	}
	if err = ops.syncFile(f); err != nil {
		return false, err
	}
	if err = f.Close(); err != nil {
		return false, err
	}
	if err = ctx.Err(); err != nil {
		return false, err
	}
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return false, fmt.Errorf("document: save target is no longer a regular file")
	} else if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err = ops.rename(temporary, path); err != nil {
		return false, err
	}
	if err = ops.syncDirectory(directory); err != nil {
		return true, err
	}
	if err = ctx.Err(); err != nil {
		return true, err
	}
	return true, nil
}
