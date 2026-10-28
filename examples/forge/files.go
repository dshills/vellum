package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	vdoc "github.com/dshills/vellum/document"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/viewport"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

const studyLimit = 1 << 20

type studyFile struct {
	Version  int
	Document document
	Camera   viewport.Camera
	Next     uint64
}

func (m *workbench) captureStudy() studyFile {
	return studyFile{1, clone(m.state), m.scene.Camera, m.next}
}
func encodeStudy(s studyFile) ([]byte, error) {
	if e := validateStudy(s); e != nil {
		return nil, e
	}
	return json.MarshalIndent(s, "", "  ")
}
func decodeStudy(b []byte) (studyFile, error) {
	var s studyFile
	if len(b) > studyLimit {
		return s, vdoc.ErrTooLarge
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(&s); e != nil {
		return s, e
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return s, fmt.Errorf("trailing study data")
	}
	return s, validateStudy(s)
}
func validateStudy(s studyFile) error {
	if s.Version != 1 || s.Next < 8 || s.Next > 1<<53 || len(s.Document.Objects) > 48 || len(s.Document.Features) > 96 || s.Document.Marker < 0 || s.Document.Marker > len(s.Document.Features) {
		return fmt.Errorf("invalid study version or limits")
	}
	scene := viewport.Scene{Camera: s.Camera, Objects: s.Document.Objects, SelectedID: s.Document.Selected}
	if e := scene.Validate(); e != nil {
		return e
	}
	ids := map[string]bool{}
	for _, o := range s.Document.Objects {
		ids[o.ID] = true
		if o.ID == fmt.Sprintf("body-%d", s.Next) {
			return fmt.Errorf("next body identity collides")
		}
	}
	seen := map[string]bool{}
	for _, f := range s.Document.Features {
		if f.ID == "" || len(f.ID) > 256 || len(f.Label) > 1024 || !ids[f.Object] || seen[f.ID] {
			return fmt.Errorf("invalid feature")
		}
		seen[f.ID] = true
	}
	return nil
}
func (m *workbench) installStudy(s studyFile) error {
	if e := validateStudy(s); e != nil {
		return e
	}
	m.state = clone(s.Document)
	m.scene.Camera = s.Camera
	m.next = s.Next
	m.refresh()
	return nil
}
func (m *workbench) bindFiles() (err error) {
	defer func() {
		if err != nil && m.recovery != nil {
			m.recovery.stop(m)
		}
	}()
	var e error
	m.files, e = vdoc.New(vdoc.Config[studyFile]{Owner: m.application, Window: m.ref.ID(), Journal: m.journal, Filename: "Study", Extensions: []string{"vforge"}, MaxBytes: studyLimit, Codec: vdoc.Codec[studyFile]{Capture: func() (studyFile, error) { return m.captureStudy(), nil }, Encode: encodeStudy, Decode: decodeStudy, Install: m.installStudy}, OnComplete: func(r vdoc.Result) {
		if r.Err != nil {
			m.status = r.Err.Error()
		} else {
			m.status = string(r.Action) + " · " + filepath.Base(r.Path)
		}
		if r.Err == nil && m.recovery != nil && !m.journal.Snapshot().Dirty {
			m.recovery.offer(studyFile{})
		}
		m.changed()
	}})
	if e != nil {
		return e
	}
	for _, c := range []struct{ id, label, key string }{{"open", "Open…", "Meta+O"}, {"save", "Save", "Meta+S"}, {"save-as", "Save As…", "Meta+Shift+S"}} {
		m.register(c.id, c.label, c.key, func() error {
			if m.files.State().Busy || m.prompt != "" {
				return vdoc.ErrBusy
			}
			return nil
		}, func() error { return m.fileAction(c.id, false) })
	}
	m.recovery, e = openRecovery()
	if e != nil {
		return e
	}
	if b, err := (vdoc.FileStore{}).Read(context.Background(), m.recovery.path, studyLimit); err == nil {
		s, err := decodeStudy(b)
		if err != nil {
			return fmt.Errorf("recovery preserved: %w", err)
		}
		m.recovered = &s
		m.prompt = "recover"
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
func (m *workbench) fileAction(id string, discard bool) error {
	if m.files == nil {
		return app.ErrUnavailable
	}
	if id == "open" {
		e := m.files.Open(m.ref.Context(), vdoc.OpenOptions{DiscardChanges: discard})
		if errors.Is(e, vdoc.ErrUnsaved) {
			m.prompt = "open"
			m.changed()
			return nil
		}
		return e
	}
	return m.files.Save(m.ref.Context(), vdoc.SaveOptions{Choose: id == "save-as"})
}
func (m *workbench) requestClose() bool {
	if m.files == nil {
		return true
	}
	if m.files.State().Busy {
		m.files.Cancel()
		m.status = "Canceling file operation; close again when complete"
		m.changed()
		return false
	}
	if m.prompt == "recover" {
		return true
	}
	if m.journal.Snapshot().Dirty {
		m.prompt = "close"
		m.changed()
		return false
	}
	m.cleanClose = true
	return true
}
func (m *workbench) filePrompt() ui.Node {
	expected := m.prompt
	valid := func() bool {
		return expected != "" && m.prompt == expected && (m.application == nil || m.ref.Context().Err() == nil)
	}
	cancel := func() {
		if valid() {
			m.prompt = ""
			m.changed()
		}
	}
	label := "Discard and Open…"
	if expected == "close" {
		label = "Discard and Close"
	}
	if expected == "recover" {
		label = "Restore study"
	}
	return ui.Node{ID: "study-prompt", Kind: ui.Dialog, Text: map[bool]string{true: "Recover interrupted study?", false: "Unsaved study"}[expected == "recover"], Open: expected != "", Dismiss: func() {
		if expected != "recover" {
			cancel()
		}
	}, Children: []ui.Node{
		{ID: "study-prompt-info", Kind: ui.Label, Text: "Cancel to save your changes first. Recovery copies never replace saved studies."},
		button("study-confirm", label, func() {
			if !valid() {
				return
			}
			m.prompt = ""
			switch expected {
			case "recover":
				s := *m.recovered
				m.recovered = nil
				m.result(m.apply("Recover study", s.Document))
				m.scene.Camera = s.Camera
				m.next = s.Next
			case "close":
				m.cleanClose = true
				_ = m.application.CloseWindow(m.ref)
			case "open":
				m.result(m.fileAction("open", true))
			}
			m.changed()
		}),
		button("study-cancel", map[bool]string{true: "Discard recovery", false: "Cancel"}[expected == "recover"], func() {
			if !valid() {
				return
			}
			if expected == "recover" {
				m.recovered = nil
				_ = os.Remove(m.recovery.path)
			}
			cancel()
		})}}
}
func (m *workbench) fileToolbar() []cogent.TitlebarToolbarItem {
	if m.files == nil {
		return nil
	}
	return []cogent.TitlebarToolbarItem{{ID: "open", Label: "Open", Kind: ui.Button, PNG: pngIcon("cube", rgb(236, 168, 52)), Command: "forge.open"}, {ID: "save", Label: "Save", Kind: ui.Button, PNG: pngIcon("save", rgb(58, 122, 225)), Command: "forge.save"}}
}

// A single bounded worker owns atomic recovery writes. flock excludes another
// Forge instance; snapshots never share mutable slices with the UI owner.
type recoveryStore struct {
	path  string
	lock  *os.File
	queue chan studyFile
	done  chan struct{}
}

func openRecovery() (*recoveryStore, error) {
	base, e := os.UserCacheDir()
	if e != nil {
		return nil, e
	}
	dir := filepath.Join(base, "VellumForge")
	if e = os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	f, e := os.OpenFile(filepath.Join(dir, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		f.Close()
		return nil, fmt.Errorf("another Forge owns recovery: %w", e)
	}
	r := &recoveryStore{filepath.Join(dir, "recovery.vforge"), f, make(chan studyFile, 1), make(chan struct{})}
	go func() {
		defer close(r.done)
		for s := range r.queue {
			if s.Version == 0 {
				_ = os.Remove(r.path)
				continue
			}
			b, e := encodeStudy(s)
			if e == nil {
				_, _ = (vdoc.FileStore{}).Write(context.Background(), r.path, b, studyLimit)
			}
		}
	}()
	return r, nil
}
func (r *recoveryStore) offer(s studyFile) {
	select {
	case r.queue <- s:
	default:
		select {
		case <-r.queue:
		default:
		}
		r.queue <- s
	}
}
func (r *recoveryStore) stop(m *workbench) {
	if r == nil {
		return
	}
	m.recovery = nil
	close(r.queue)
	<-r.done
	if m.cleanClose {
		_ = os.Remove(r.path)
	}
	_ = syscall.Flock(int(r.lock.Fd()), syscall.LOCK_UN)
	_ = r.lock.Close()
}

func (m *workbench) syncRecovery() {
	if m.recovery == nil || m.prompt == "recover" {
		return
	}
	if m.journal.Snapshot().Dirty {
		m.recovery.offer(m.captureStudy())
	} else {
		m.recovery.offer(studyFile{})
	}
}
