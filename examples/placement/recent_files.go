package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
)

const recentLimit, recentByteLimit = 10, 64 << 10
const recentFilename = "recent-drawings.json"

type recentConfig struct {
	Directory string
	// Injected stores must honor cancellation so Stop can join their worker.
	Store document.Store
}
type recentFile struct {
	Version int      `json:"version"`
	Paths   []string `json:"paths"`
}

// Only the bounded recent-list state is synchronized; no document/model values
// cross to this worker. One writer coalesces updates and owns all private IO.
type recentDrawings struct {
	mu                     sync.Mutex
	paths                  []string
	dirty, loading, closed bool
	lastErr                error
	ctx                    context.Context
	cancel                 context.CancelFunc
	wake, done             chan struct{}
	startup                *app.Completion
	path                   string
	store                  document.Store
}

func newRecentDrawings(a *app.Application, ref app.WindowRef, config recentConfig) (*recentDrawings, error) {
	if a == nil || ref.Context().Err() != nil {
		return nil, app.ErrWindowClosed
	}
	if config.Directory == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}
		config.Directory = filepath.Join(base, "VellumPlacementInternal")
	}
	if !filepath.IsAbs(config.Directory) {
		return nil, errors.New("recents: absolute private directory required")
	}
	if config.Store == nil {
		config.Store = document.FileStore{}
	}
	ticket, err := a.ReserveWindowCompletion(ref, nil)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ref.Context())
	r := &recentDrawings{ctx: ctx, cancel: cancel, wake: make(chan struct{}, 1), done: make(chan struct{}), startup: ticket,
		path: filepath.Join(config.Directory, recentFilename), store: config.Store, loading: true}
	go r.run(a, ref)
	return r, nil
}

func recentPath(path string) (string, error) {
	if len(path) == 0 || len(path) > 4096 || !utf8.ValidString(path) || strings.IndexByte(path, 0) >= 0 || !filepath.IsAbs(path) {
		return "", errors.New("recents: invalid absolute path")
	}
	return filepath.Clean(path), nil
}
func slicesContainsPath(paths []string, path string) bool { return slices.Contains(paths, path) }
func mergedRecent(newer, older []string) []string {
	result := make([]string, 0, recentLimit)
	for _, paths := range [][]string{newer, older} {
		for _, p := range paths {
			if !slices.Contains(result, p) {
				result = append(result, p)
				if len(result) == recentLimit {
					return result
				}
			}
		}
	}
	return result
}
func (r *recentDrawings) Paths() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.paths)
}
func (r *recentDrawings) Status() (loading bool, err error) {
	if r == nil {
		return false, app.ErrUnavailable
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loading, r.lastErr
}
func (r *recentDrawings) Record(path string) error {
	if r == nil {
		return app.ErrUnavailable
	}
	path, err := recentPath(path)
	if err != nil {
		return err
	}
	r.mu.Lock()
	if r.closed || r.ctx.Err() != nil {
		r.mu.Unlock()
		return app.ErrClosed
	}
	r.paths = mergedRecent([]string{path}, r.paths)
	r.dirty = true
	r.mu.Unlock()
	select {
	case r.wake <- struct{}{}:
	default:
	}
	return nil
}
func decodeRecents(data []byte) ([]string, error) {
	if len(data) > recentByteLimit || !utf8.Valid(data) {
		return nil, errors.New("recents: invalid cache size or encoding")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var f recentFile
	if err := d.Decode(&f); err != nil {
		return nil, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, errors.New("recents: trailing cache data")
	}
	if f.Version != 1 || len(f.Paths) > recentLimit {
		return nil, errors.New("recents: unsupported cache")
	}
	paths := make([]string, 0, len(f.Paths))
	for _, p := range f.Paths {
		clean, err := recentPath(p)
		if err != nil {
			return nil, err
		}
		if slices.Contains(paths, clean) {
			return nil, errors.New("recents: duplicate cache path")
		}
		paths = append(paths, clean)
	}
	return paths, nil
}
func recentPrivatePath(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return errors.New("recents: private directory required")
	}
	info, err = os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return errors.New("recents: private regular cache required")
	}
	return nil
}
func (r *recentDrawings) flush(ctx context.Context) {
	r.mu.Lock()
	if !r.dirty {
		r.mu.Unlock()
		return
	}
	paths := slices.Clone(r.paths)
	r.dirty = false
	r.mu.Unlock()
	data, err := json.Marshal(recentFile{Version: 1, Paths: paths})
	if err == nil && len(data) > recentByteLimit {
		err = document.ErrTooLarge
	}
	if err == nil {
		err = recentPrivatePath(r.path)
	}
	if err == nil {
		var written bool
		written, err = r.store.Write(ctx, r.path, data, recentByteLimit)
		if !written && err == nil {
			err = errors.New("recents: cache write not committed")
		}
	}
	r.mu.Lock()
	r.lastErr = err
	if err != nil {
		r.dirty = true
	}
	r.mu.Unlock()
}
func (r *recentDrawings) run(a *app.Application, ref app.WindowRef) {
	defer close(r.done)
	var paths []string
	err := recentPrivatePath(r.path)
	if err == nil {
		var data []byte
		data, err = r.store.Read(r.ctx, r.path, recentByteLimit)
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		} else if err == nil {
			paths, err = decodeRecents(data)
		}
	}
	r.mu.Lock()
	r.paths = mergedRecent(r.paths, paths)
	r.loading = false
	r.lastErr = err
	r.mu.Unlock()
	if err := r.startup.Publish(func() {
		if r.ctx.Err() == nil {
			_ = a.Invalidate("placement")
		}
	}); err != nil {
		r.startup.Abort()
	}
	for {
		select {
		case <-r.ctx.Done():
			// Join preserves the latest successful file operation even when window
			// retirement races its coalesced write. No document callback is emitted.
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			r.flush(ctx)
			cancel()
			return
		case <-r.wake:
			r.flush(r.ctx)
		}
	}
}
func (r *recentDrawings) Stop() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()
	r.cancel()
	r.startup.Abort()
	<-r.done
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastErr
}
