//go:build !(darwin || linux || freebsd || openbsd || netbsd || dragonfly)

package document

import "os"

func openReadFile(path string) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrInvalid}
	}
	return os.Open(path)
}
