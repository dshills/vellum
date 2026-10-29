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
	"strconv"
	"strings"
	"sync"
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
	scene := viewport.Scene{Camera: s.Camera, Objects: s.Document.Objects, SelectedIDs: s.Document.Selection, SelectedID: s.Document.Selected}
	if e := scene.Validate(); e != nil {
		return e
	}
	ids := map[string]bool{}
	for _, o := range s.Document.Objects {
		if o.Overlay || o.ScreenSize != 0 || strings.HasPrefix(o.ID, "forge-handle-") {
			return fmt.Errorf("reserved body identity")
		}
		if strings.HasPrefix(o.ID, "body-") {
			if n, e := strconv.ParseUint(strings.TrimPrefix(o.ID, "body-"), 10, 64); e == nil && n >= s.Next {
				return fmt.Errorf("next body identity not monotonic")
			}
		}
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
	m.isolatedIDs = nil
	m.renameID = ""
	m.renameEpoch++
	m.pasteKey = [32]byte{}
	m.pasteCount = 0
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
			m.lastStudy = r.Path
			m.recordRecent(r.Path)
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
	m.recovery, e = openRecovery(m.application)
	if e != nil {
		return e
	}
	if b, err := (vdoc.FileStore{}).Read(context.Background(), m.recovery.path, studyLimit); err == nil {
		s, err := decodeStudy(b)
		if err != nil {
			return fmt.Errorf("recovery preserved: %w", err)
		}
		m.recovered = &s
		m.showFilePrompt("recover", "")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	m.loadWorkspace()
	if m.recovered == nil && m.lastStudy != "" && !m.preferencesBlocked {
		if err := m.files.Open(m.ref.Context(), vdoc.OpenOptions{Path: m.lastStudy}); err != nil {
			m.status = "Restore: " + err.Error()
		}
	}
	m.bindWorkspaceCommands()
	m.bindRecentCommands()
	return nil
}
func (m *workbench) fileAction(id string, discard bool) error {
	if m.clipboardBusy || m.drag != nil || m.renameID != "" || m.recentOpen {
		return vdoc.ErrBusy
	}
	if m.files == nil {
		return app.ErrUnavailable
	}
	if id == "open" {
		return m.openStudy("", discard)
	}
	if id == "revert" {
		path := m.files.State().Path
		if path == "" {
			return fmt.Errorf("save the study before reverting")
		}
		m.showFilePrompt("revert", path)
		return nil
	}

	return m.files.Save(m.ref.Context(), vdoc.SaveOptions{Choose: id == "save-as"})
}
func (m *workbench) requestClose() bool {
	if m.renameID != "" {
		m.renameID = ""
		m.renameEpoch++
		m.changed()
		return false
	}
	if m.recentOpen {
		m.recentOpen = false
		m.changed()
		return false
	}
	if m.clipboardBusy {
		if m.clipboardCancel != nil {
			m.clipboardCancel()
		}
		m.status = "Canceling clipboard operation; close again when complete"
		m.changed()
		return false
	}
	if m.drag != nil {
		m.drag = nil
		m.refresh()
	}
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
		m.showFilePrompt("close", "")
		m.changed()
		return false
	}
	m.cleanClose = true
	return true
}
func (m *workbench) filePrompt() ui.Node {
	expected, epoch, path := m.prompt, m.fileEpoch, m.pendingStudyPath
	valid := func() bool {
		return expected != "" && m.prompt == expected && m.fileEpoch == epoch && m.pendingStudyPath == path && (m.application == nil || m.ref.Context().Err() == nil)
	}
	cancel := func() {
		if valid() {
			m.prompt = ""
			m.pendingStudyPath = ""
			m.fileEpoch++
			m.changed()
		}
	}
	label := "Discard and Open…"
	if expected == "close" {
		label = "Discard and Close"
	}
	if expected == "revert" {
		label = "Revert to Saved"
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
				m.result(m.restoreRecoveredStudy(s))
			case "close":
				m.cleanClose = true
				_ = m.application.CloseWindow(m.ref)
			case "open", "revert":
				m.pendingStudyPath = ""
				m.fileEpoch++
				m.result(m.openStudy(path, true))
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
	return []cogent.TitlebarToolbarItem{{ID: "open", Label: "Open", Kind: ui.Button, PNG: pngIcon("open", rgb(236, 168, 52)), Command: "forge.open"}, {ID: "save", Label: "Save", Kind: ui.Button, PNG: pngIcon("save", rgb(58, 122, 225)), Command: "forge.save"}}
}

// A single bounded worker owns atomic recovery writes. flock excludes another
// Forge instance; snapshots never share mutable slices with the UI owner.
type recoveryStore struct {
	path      string
	lock      *os.File
	queue     chan studyFile
	done      chan struct{}
	owner     *app.Application
	failureMu sync.Mutex
	failure   error
}

func openRecovery(owners ...*app.Application) (*recoveryStore, error) {
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
	r := &recoveryStore{path: filepath.Join(dir, "recovery.vforge"), lock: f, queue: make(chan studyFile, 1), done: make(chan struct{})}
	if len(owners) > 0 {
		r.owner = owners[0]
	}
	go func() {
		defer close(r.done)
		for s := range r.queue {
			r.process(s)
		}
	}()
	return r, nil
}

// process records recovery failures independently of ordinary status messages.
func (r *recoveryStore) process(s studyFile) {
	var err error
	if s.Version == 0 {
		err = os.Remove(r.path)
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
	} else {
		var b []byte
		b, err = encodeStudy(s)
		if err == nil {
			_, err = (vdoc.FileStore{}).Write(context.Background(), r.path, b, studyLimit)
		}
	}
	r.failureMu.Lock()
	changed := fmt.Sprint(r.failure) != fmt.Sprint(err)
	r.failure = err
	r.failureMu.Unlock()
	if changed && r.owner != nil {
		_ = r.owner.Invalidate("")
	}
}
func (r *recoveryStore) warning() string {
	if r == nil {
		return ""
	}
	r.failureMu.Lock()
	defer r.failureMu.Unlock()
	if r.failure != nil {
		return "Recovery unavailable: " + r.failure.Error()
	}
	return ""
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
	m.saveWorkspace(r.path)
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

func (m *workbench) restoreRecoveredStudy(s studyFile) error {
	if e := validateStudy(s); e != nil {
		return e
	}
	oldCamera, oldNext := m.scene.Camera, m.next
	m.scene.Camera = s.Camera
	m.next = s.Next
	if e := m.apply("Recover study", s.Document); e != nil {
		m.scene.Camera = oldCamera
		m.next = oldNext
		m.refresh()
		return e
	}
	return nil
}
