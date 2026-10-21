package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

const recoveryFilename = "recovery.vellum"
const recoveryLockFilename = ".recovery.lock"

type recoveryConfig struct {
	Directory string
	Interval  time.Duration
	// Injected stores must honor cancellation and return from IO so Stop can join.
	// Production uses bounded local FileStore IO; lock release follows the join.
	Store document.Store
}

// placementRecovery belongs to one window opening. All model/prompt fields are
// owner-only. Its one worker receives detached captures through reserved owner
// completions, and performs all encoding and private disk work serially.
type placementRecovery struct {
	owner    *app.Application
	window   app.WindowRef
	model    *model
	codec    document.Codec[drawingFile]
	store    document.Store
	path     string
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	wake     chan struct{}
	// Written by the worker; Stop reads it only after done closes.
	storageReady  bool
	directoryLock *os.File
	stopMu        sync.Mutex
	stopped       bool
	stopErr       error

	checking   bool
	offer      *drawingFile
	loadErr    error
	removeNext bool
	initialRev uint64
}

// Call on the owner before the first View. Startup is an explicit decision;
// no existing copy is overwritten while its restore/discard dialog is open.
func newPlacementRecovery(a *app.Application, ref app.WindowRef, m *model, c recoveryConfig) (*placementRecovery, error) {
	if a == nil || m == nil || m.journal == nil || ref.Context().Err() != nil {
		return nil, errors.New("recovery: live owner, window and model required")
	}
	state := m.journal.Snapshot()
	if state.Dirty || state.Busy || m.prompt != "" {
		return nil, document.ErrBusy
	}
	if c.Directory == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}
		c.Directory = filepath.Join(base, "VellumPlacementInternal")
	}
	if !filepath.IsAbs(c.Directory) {
		return nil, errors.New("recovery: absolute private directory required")
	}
	if c.Interval == 0 {
		c.Interval = 15 * time.Second
	}
	if c.Interval < 0 {
		return nil, errors.New("recovery: positive interval required")
	}
	if c.Store == nil {
		c.Store = document.FileStore{}
	}
	ctx, cancel := context.WithCancel(ref.Context())
	r := &placementRecovery{owner: a, window: ref, model: m, codec: drawingCodec(m), store: c.Store, path: filepath.Join(c.Directory, recoveryFilename), interval: c.Interval, ctx: ctx, cancel: cancel, done: make(chan struct{}), wake: make(chan struct{}, 1), checking: true, initialRev: state.Revision}
	// Reserve startup delivery synchronously: failed admission must not leave
	// a modal gate with no worker able to resolve it.
	ticket, err := a.ReserveWindowCompletion(ref, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	m.prompt = "recovery"
	go r.run(ticket)
	return r, nil
}

func (r *placementRecovery) pending() bool {
	return r != nil && (r.checking || r.offer != nil || r.loadErr != nil)
}

// Request coalesces optional immediate checks (for example after Save/Open).
// Periodic checks continue independently; there is never a queue of snapshots.
func (r *placementRecovery) Request() {
	if r == nil {
		return
	}
	select {
	case r.wake <- struct{}{}:
	default:
	}
}

// Stop joins the worker before removing a clean/discarded copy, so an older
// in-flight write cannot resurrect it. Call after Run, outside the UI owner.
// A failed run or a close without an explicit clean/discard decision uses false.
// Process interruption never reaches this cleanup and retains the last copy.
// The first Stop call chooses removal; subsequent calls return its result and
// cannot delete a later instance's copy after releasing the directory lock.
func (r *placementRecovery) Stop(remove bool) error {
	if r == nil {
		return nil
	}
	r.stopMu.Lock()
	defer r.stopMu.Unlock()
	if r.stopped {
		return r.stopErr
	}
	r.cancel()
	<-r.done
	if remove && r.storageReady && r.directoryLock != nil {
		r.stopErr = removeRecoveryFile(r.path)
	}
	if r.directoryLock != nil {
		r.stopErr = errors.Join(r.stopErr, r.directoryLock.Close())
	}
	r.stopped = true
	return r.stopErr
}

type recoveryCapture struct {
	image  drawingFile
	write  bool
	remove bool
}

func (r *placementRecovery) capture() recoveryCapture {
	if r.pending() || r.ctx.Err() != nil {
		return recoveryCapture{}
	}
	if r.removeNext {
		r.removeNext = false
		return recoveryCapture{remove: true}
	}
	m := r.model
	s := m.journal.Snapshot()
	if s.Busy || s.Invalid || m.fileBusy() || m.move != nil || m.pending != nil || m.palette.State().Active || m.prompt != "" {
		return recoveryCapture{}
	}
	if !s.Dirty {
		return recoveryCapture{remove: true}
	}
	image, err := r.codec.Capture()
	if err != nil {
		return recoveryCapture{} // Defer transient gesture/operation admission.
	}
	return recoveryCapture{image: image, write: true}
}

// onOwner never enters Do from a worker and has at most one outstanding
// completion. Cancellation aborts an undispatched callback and releases credit.
func (r *placementRecovery) onOwner(fn func()) error {
	if err := r.ctx.Err(); err != nil {
		return err
	}
	ticket, err := r.owner.ReserveWindowCompletion(r.window, nil)
	if err != nil {
		return err
	}
	defer ticket.Abort()
	if err := ticket.Publish(func() {
		if r.ctx.Err() == nil {
			fn()
		}
	}); err != nil {
		return err
	}
	select {
	case <-r.ctx.Done():
		return r.ctx.Err()
	case <-ticket.Done():
		return ticket.Err()
	}
}

func (r *placementRecovery) report(err error) {
	if err == nil || r.ctx.Err() != nil {
		return
	}
	_ = r.onOwner(func() {
		r.model.status = "Recovery copy failed; drawing kept: " + err.Error()
		_ = r.owner.Invalidate(r.window.ID())
	})
}

func (r *placementRecovery) run(startup *app.Completion) {
	defer close(r.done)
	defer startup.Abort()
	err := prepareRecoveryDirectory(filepath.Dir(r.path))
	if err == nil {
		r.directoryLock, err = lockRecoveryDirectory(filepath.Dir(r.path))
	}
	directoryErr := err
	r.storageReady = err == nil
	var image drawingFile
	found := false
	if err == nil {
		err = privateRecoveryTarget(r.path)
	}
	if err == nil {
		var data []byte
		data, err = r.store.Read(r.ctx, r.path, drawingByteLimit)
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		} else if err == nil {
			if len(data) > drawingByteLimit {
				err = document.ErrTooLarge
			} else {
				image, err = r.codec.Decode(data)
				found = err == nil
			}
		}
	}
	if startup.Publish(func() {
		if r.ctx.Err() != nil {
			return
		}
		r.checking = false
		if directoryErr != nil {
			// An unusable cache is not evidence of an interrupted drawing.
			// Continue editing without autosave instead of blocking startup.
			r.model.prompt = ""
			r.model.status = "Recovery unavailable; drawing remains usable: " + directoryErr.Error()
		} else if found {
			r.offer = &image
		} else if err == nil {
			r.model.prompt = ""
		} else {
			r.loadErr = err
		}
		_ = r.owner.Invalidate(r.window.ID())
	}) != nil {
		return
	}
	select {
	case <-r.ctx.Done():
		return
	case <-startup.Done():
		if startup.Err() != nil {
			return
		}
	}
	if directoryErr != nil {
		return
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	var last [32]byte
	haveLast := false
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
		case <-r.wake:
		}
		var c recoveryCapture
		if err := r.onOwner(func() { c = r.capture() }); err != nil {
			continue // Capacity pressure defers this tick; closure ends the loop.
		}
		if c.remove {
			err := removeRecoveryFile(r.path)
			if err == nil {
				haveLast = false
			}
			r.report(err)
		} else if c.write {
			data, err := r.codec.Encode(c.image)
			if err == nil && len(data) > drawingByteLimit {
				err = document.ErrTooLarge
			}
			digest := sha256.Sum256(data)
			if err == nil && (!haveLast || digest != last) {
				err = privateRecoveryTarget(r.path)
				if err == nil {
					_, err = r.store.Write(r.ctx, r.path, data, drawingByteLimit)
				}
				if err == nil {
					last, haveLast = digest, true
				}
			}
			r.report(err)
		}
	}
}

func prepareRecoveryDirectory(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return errors.New("recovery: private directory required")
	}
	return nil
}

func privateRecoveryTarget(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return errors.New("recovery: private regular file required")
	}
	return nil
}

func removeRecoveryFile(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
		return errors.New("recovery: recovery target is not a file")
	}
	if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

// Restore is one explicit undoable document edit, not a document.Session Open.
// It keeps the original path untouched and remains dirty until a real Save.
func (r *placementRecovery) restore() error {
	if r == nil || r.offer == nil || r.checking || r.ctx.Err() != nil {
		return errors.New("recovery: no available copy")
	}
	m := r.model
	if m.fileBusy() || m.move != nil || m.palette.State().Active || m.journal.Snapshot().Revision != r.initialRev {
		return document.ErrBusy
	}
	before, err := r.codec.Capture()
	if err != nil {
		return err
	}
	after := *r.offer
	install := func(image drawingFile) error {
		// Counters remain monotonic across undo branches, as with other edits.
		image.NextID = max(image.NextID, m.next)
		image.NextGroup = max(image.NextGroup, m.nextGroup)
		if image.NextGroup != 0 && image.Version < 3 {
			image.Version = 3
		}
		return r.codec.Install(image)
	}
	if err := m.journal.Apply(history.Change{Label: "Restore recovered drawing", Apply: func() error { return install(after) }, Revert: func() error { return install(before) }}); err != nil {
		return err
	}
	r.offer, r.loadErr = nil, nil
	m.prompt, m.status = "", "Recovered drawing; save to keep it as a document"
	r.Request()
	return r.owner.Invalidate(r.window.ID())
}

func (r *placementRecovery) discard() {
	if r == nil || r.checking || r.ctx.Err() != nil {
		return
	}
	r.offer, r.loadErr, r.removeNext = nil, nil, true
	r.model.prompt, r.model.status = "", "Recovery copy discarded"
	r.Request()
	_ = r.owner.Invalidate(r.window.ID())
}

func (r *placementRecovery) promptNode() ui.Node {
	if r == nil {
		return ui.Node{ID: "recovery-decision", Kind: ui.Dialog}
	}
	text := "A private recovery copy is available after an interrupted session. Restore it as an unsaved drawing, or discard the copy."
	if r.checking {
		text = "Checking for a private recovery copy…"
	} else if r.loadErr != nil {
		text = fmt.Sprintf("The recovery copy could not be loaded (%v). The current drawing is unchanged. Discard the unreadable copy to continue, or close to keep it for later.", r.loadErr)
	}
	return ui.Node{ID: "recovery-decision", Kind: ui.Dialog, Text: "Recover drawing", Open: r.pending(), Dismiss: func() {
		r.model.status = "Choose Restore or Discard to continue, or Close to keep the copy"
		_ = r.owner.Invalidate(r.window.ID())
	}, Children: []ui.Node{
		{ID: "recovery-description", Kind: ui.Label, Text: text},
		{ID: "recovery-restore", Kind: ui.Button, Text: "Restore Drawing", Disabled: r.checking || r.offer == nil, Activate: func() {
			if err := r.restore(); err != nil {
				r.model.status = err.Error()
				_ = r.owner.Invalidate(r.window.ID())
			}
		}},
		{ID: "recovery-discard", Kind: ui.Button, Text: "Discard Recovery", Disabled: r.checking, Activate: r.discard},
		{ID: "recovery-close", Kind: ui.Button, Text: "Close; Keep Recovery", Activate: func() { _ = r.owner.CloseWindow(r.window) }},
	}}
}
