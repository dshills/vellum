package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unicode/utf8"
)

const preferenceFilename = "editor-preferences.json"
const preferenceByteLimit = 4096
const (
	prefGridVisible uint8 = 1 << iota
	prefGridSpacing
	prefAspect
	prefSubtract
	prefAll = prefGridVisible | prefGridSpacing | prefAspect | prefSubtract
)

type preferenceValues struct {
	Version           int     `json:"version"`
	GridHidden        bool    `json:"grid_hidden"`
	GridSpacing       float64 `json:"grid_spacing"`
	AspectLocked      bool    `json:"aspect_locked"`
	SelectionSubtract bool    `json:"selection_subtract"`
}

func defaultPreferences() preferenceValues { return preferenceValues{Version: 1, GridSpacing: 20} }
func mergePreferences(base, newer preferenceValues, mask uint8) preferenceValues {
	if mask&prefGridVisible != 0 {
		base.GridHidden = newer.GridHidden
	}
	if mask&prefGridSpacing != 0 {
		base.GridSpacing = newer.GridSpacing
	}
	if mask&prefAspect != 0 {
		base.AspectLocked = newer.AspectLocked
	}
	if mask&prefSubtract != 0 {
		base.SelectionSubtract = newer.SelectionSubtract
	}
	return base
}
func decodePreferences(data []byte) (preferenceValues, error) {
	v := preferenceValues{}
	if len(data) > preferenceByteLimit || !utf8.Valid(data) {
		return v, errors.New("invalid preference size/encoding")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&v); err != nil {
		return v, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return v, errors.New("trailing preference data")
	}
	if v.Version != 1 || v.GridSpacing != gridUnit(v.GridSpacing) {
		return v, errors.New("unsupported preferences or grid spacing")
	}
	return v, nil
}

type preferenceConfig struct {
	Directory string
	Store     document.Store
}
type editorPreferences struct {
	mu            sync.Mutex
	values        preferenceValues
	changed       uint8
	dirty, closed bool
	lastErr       error
	ctx           context.Context
	cancel        context.CancelFunc
	wake, done    chan struct{}
	startup       *app.Completion
	path          string
	store         document.Store
}

func (m *model) currentPreferences() preferenceValues {
	return preferenceValues{1, m.gridHidden, m.spacing(), m.aspectLocked, m.selectionSubtract}
}
func (m *model) installPreferences(v preferenceValues) {
	m.gridHidden, m.gridSpacing, m.aspectLocked, m.selectionSubtract = v.GridHidden, v.GridSpacing, v.AspectLocked, v.SelectionSubtract
}
func (m *model) preferencesChanged(mask uint8) {
	m.preferenceTouched |= mask
	if m.preferences != nil {
		if err := m.preferences.Record(m.currentPreferences(), mask); err != nil {
			m.status = "Preferences not saved: " + err.Error()
		}
	}
}
func newEditorPreferences(a *app.Application, ref app.WindowRef, m *model, c preferenceConfig) (*editorPreferences, error) {
	if a == nil || ref.Context().Err() != nil {
		return nil, app.ErrWindowClosed
	}
	if m.preferences != nil {
		return nil, errors.New("preferences already bound")
	}
	if c.Directory == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		c.Directory = filepath.Join(dir, "VellumPlacementInternal")
	}
	if !filepath.IsAbs(c.Directory) {
		return nil, errors.New("preferences require an absolute private directory")
	}
	if c.Store == nil {
		c.Store = document.FileStore{}
	}
	ticket, err := a.ReserveWindowCompletion(ref, nil)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ref.Context())
	r := &editorPreferences{values: m.currentPreferences(), changed: m.preferenceTouched, dirty: m.preferenceTouched != 0, ctx: ctx, cancel: cancel, wake: make(chan struct{}, 1), done: make(chan struct{}), startup: ticket, path: filepath.Join(c.Directory, preferenceFilename), store: c.Store}
	m.preferences = r
	m.preferencesLoading = true
	go r.run(a, ref, m)
	return r, nil
}
func (r *editorPreferences) Record(v preferenceValues, mask uint8) error {
	if r == nil {
		return app.ErrUnavailable
	}
	if mask&^prefAll != 0 || v.Version != 1 || v.GridSpacing != gridUnit(v.GridSpacing) {
		return errors.New("invalid editor preferences")
	}
	r.mu.Lock()
	if r.closed || r.ctx.Err() != nil {
		r.mu.Unlock()
		return app.ErrClosed
	}
	r.values = mergePreferences(r.values, v, mask)
	r.changed |= mask
	r.dirty = true
	r.mu.Unlock()
	select {
	case r.wake <- struct{}{}:
	default:
	}
	return nil
}
func (r *editorPreferences) flush(ctx context.Context) {
	r.mu.Lock()
	if !r.dirty {
		r.mu.Unlock()
		return
	}
	v := r.values
	r.dirty = false
	r.mu.Unlock()
	data, err := json.Marshal(v)
	if err == nil {
		err = recentPrivatePath(r.path)
	}
	if err == nil {
		var written bool
		written, err = r.store.Write(ctx, r.path, data, preferenceByteLimit)
		if !written && err == nil {
			err = errors.New("preference write did not commit")
		}
	}
	r.mu.Lock()
	r.lastErr = err
	if err != nil {
		r.dirty = true
	}
	r.mu.Unlock()
}
func (r *editorPreferences) run(a *app.Application, ref app.WindowRef, m *model) {
	defer close(r.done)
	loaded := defaultPreferences()
	err := recentPrivatePath(r.path)
	if err == nil {
		var data []byte
		data, err = r.store.Read(r.ctx, r.path, preferenceByteLimit)
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		} else if err == nil {
			var v preferenceValues
			v, err = decodePreferences(data)
			if err == nil {
				loaded = v
			}
		}
	}
	r.mu.Lock()
	r.values = mergePreferences(loaded, r.values, r.changed)
	r.lastErr = err
	r.mu.Unlock()
	if e := r.startup.Publish(func() {
		if r.ctx.Err() != nil || ref.Context().Err() != nil || m.preferences != r {
			return
		}
		m.installPreferences(mergePreferences(loaded, m.currentPreferences(), m.preferenceTouched))
		m.preferencesLoading = false
		if err != nil {
			m.status = "Preferences unavailable; using defaults: " + err.Error()
		}
		_ = a.Invalidate("placement")
	}); e != nil {
		// A reserved completion cannot fail from ordinary queue pressure.
		// Refusal means this opening was retired, or Stop/close already aborted
		// it; normal close clears preferencesLoading before that revocation.
		r.startup.Abort()
	}
	for {
		select {
		case <-r.ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			r.flush(ctx)
			cancel()
			return
		case <-r.wake:
			r.flush(r.ctx)
		}
	}
}
func (r *editorPreferences) Stop() error {
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
func (m *model) cancelPreferenceStartup() {
	if m.preferencesLoading {
		m.preferencesLoading = false
		if r := m.preferences; r != nil {
			r.cancel()
			r.startup.Abort()
		}
	}
}
