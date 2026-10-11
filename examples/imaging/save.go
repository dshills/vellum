package main

import (
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

func savePNG(ctx context.Context, path string, img image.Image) error {
	return replacePNG(ctx, path, img, png.Encode)
}

// replacePNG writes in the destination directory and commits only a completed,
// synced, closed PNG. Encode/cancellation/IO failure preserves the existing file.
func replacePNG(ctx context.Context, path string, img image.Image, encode func(io.Writer, image.Image) error) error {
	if ctx == nil || path == "" || img == nil || encode == nil {
		return errors.New("PNG capture destination/image/context required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	mode := os.FileMode(0600)
	if info, err := os.Stat(path); err == nil {
		if !info.Mode().IsRegular() {
			return errors.New("PNG destination must be a regular file")
		}
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".vellum-imaging-*.png")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err = file.Chmod(mode); err == nil {
		err = encode(file, img)
	}
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		err = file.Sync()
	}
	err = errors.Join(err, file.Close())
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
