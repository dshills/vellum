//go:build darwin

package main

import (
	"bytes"
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/placement"
	"github.com/dshills/vellum/ui"
)

func recoveryPrivateDirectory(t *testing.T) string {
	t.Helper()
	path := t.TempDir()
	if err := os.Chmod(path, 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func startRecovery(t *testing.T, f *fileFixture, c recoveryConfig) *placementRecovery {
	t.Helper()
	if c.Directory == "" {
		c.Directory = recoveryPrivateDirectory(t)
	}
	if c.Interval == 0 {
		c.Interval = time.Hour
	}
	var r *placementRecovery
	f.owner(t, func() {
		var err error
		r, err = newPlacementRecovery(f.a, f.ref, f.m, c)
		if err != nil {
			t.Fatal(err)
		}
		f.m.recovery = r
	})
	t.Cleanup(func() {
		if err := r.Stop(false); err != nil {
			t.Error(err)
		}
	})
	recoveryWait(t, f, func() bool { return !r.checking })
	return r
}

func recoveryWait(t *testing.T, f *fileFixture, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := f.a.Drain(); err != nil {
			t.Fatal(err)
		}
		done := false
		f.owner(t, func() { done = ready() })
		if done {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("recovery did not settle")
}

func recoveryDiskMatches(path string, image drawingFile) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	got, err := drawingCodec(nil).Decode(data)
	return err == nil && reflect.DeepEqual(got, image)
}

func TestRecoveryRoundTripIsPrivateAndDoesNotAcknowledgeSave(t *testing.T) {
	directory := recoveryPrivateDirectory(t)
	f := fileModel(t, false)
	r := startRecovery(t, f, recoveryConfig{Directory: directory})
	f.place(t, "note", canvas.Point{X: 40, Y: -20})
	f.place(t, "block", canvas.Point{X: 240, Y: 120})
	var image drawingFile
	f.owner(t, func() {
		f.m.setSelection([]string{"note-1", "block-2"}, "block-2")
		if err := f.m.changeGrouping(false); err != nil {
			t.Fatal(err)
		}
		f.m.scene.View.Center = canvas.Point{X: 19.5, Y: -43.25}
		f.m.scene.View.Zoom = 1.75
		f.m.snap = false
		image, _ = drawingCodec(f.m).Capture()
	})
	before := f.m.journal.Snapshot()
	r.Request()
	recoveryWait(t, f, func() bool { return recoveryDiskMatches(r.path, image) })
	f.owner(t, func() {
		if after := f.m.journal.Snapshot(); after != before || !after.Dirty {
			t.Fatal("autosave changed history/saved state", before, after)
		}
		if s := f.m.files.State(); s.Path != "" || !s.Dirty || s.Busy {
			t.Fatal("recovery impersonated a file save", s)
		}
	})
	for _, path := range []string{directory, r.path} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm()&0077 != 0 {
			t.Fatal("copy is not private", path, info, err)
		}
	}
	if err := r.Stop(false); err != nil {
		t.Fatal(err)
	}
	if !recoveryDiskMatches(r.path, image) {
		t.Fatal("interruption-style shutdown removed recovery")
	}

	// A new process starts with its own clean, unpathed document.
	next := fileModel(t, false)
	recovered := startRecovery(t, next, recoveryConfig{Directory: directory})
	next.owner(t, func() {
		if !recovered.pending() || recovered.offer == nil || len(next.m.scene.Objects) != 0 {
			t.Fatal("startup installed without a decision")
		}
		next.m.scene.View.Width, next.m.scene.View.Height = 1234, 567
		if err := recovered.restore(); err != nil {
			t.Fatal(err)
		}
		got, _ := drawingCodec(next.m).Capture()
		if !reflect.DeepEqual(got, image) || recovered.pending() || next.m.prompt != "" {
			t.Fatal("restore differs", got, image)
		}
		s := next.m.journal.Snapshot()
		if !s.Dirty || s.UndoCount != 1 || s.UndoLabel != "Restore recovered drawing" || next.m.files.State().Path != "" {
			t.Fatal("restore was not one unsaved edit", s)
		}
		if next.m.scene.View.Width != 1234 || next.m.scene.View.Height != 567 {
			t.Fatal("recovery replaced receiving viewport")
		}
		if err := next.m.journal.Undo(); err != nil || len(next.m.scene.Objects) != 0 || next.m.journal.Snapshot().Dirty {
			t.Fatal("restore undo failed", err)
		}
		if next.m.next != image.NextID || next.m.nextGroup != image.NextGroup {
			t.Fatal("recovery undo regressed monotonic IDs")
		}
		if err := next.m.journal.Redo(); err != nil || !next.m.journal.Snapshot().Dirty {
			t.Fatal("restore redo failed", err)
		}
		if err := next.m.files.Save(next.ref.Context(), document.SaveOptions{Path: next.path}); err != nil {
			t.Fatal(err)
		}
	})
	next.wait(t)
	next.owner(t, func() {
		if s := next.m.files.State(); s.Last.Err != nil || s.Dirty || s.Path != next.path {
			t.Fatal("explicit save failed", s)
		}
	})
	// The successful document-session completion requests immediate cleanup.
	recoveryWait(t, next, func() bool { _, err := os.Stat(recovered.path); return errors.Is(err, os.ErrNotExist) })
}

func TestRecoveryVersion4RestoresLayerMetadataThroughUndoRedo(t *testing.T) {
	directory := recoveryPrivateDirectory(t)
	f := fileModel(t, false)
	r := startRecovery(t, f, recoveryConfig{Directory: directory})
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200})
	f.place(t, "connector", canvas.Point{X: 400})
	var image drawingFile
	var info map[string]layerMetadata
	f.owner(t, func() {
		f.m.setSelection([]string{"block-1", "note-2"}, "block-1")
		for _, change := range []func() error{
			func() error { return f.m.changeGrouping(false) },
			func() error { return f.m.renameLayer("group-1", "Assembly") },
			func() error { return f.m.renameLayer("block-1", "Housing") },
			func() error { return f.m.toggleLayerVisibility("group-1") },
			func() error { return f.m.toggleLayerLock("connector-3") },
		} {
			if err := change(); err != nil {
				t.Fatal(err)
			}
		}
		var err error
		image, err = drawingCodec(f.m).Capture()
		if err != nil || image.Version != 4 {
			t.Fatal("format4 capture", err, image.Version)
		}
		info = maps.Clone(f.m.layerInfo)
	})
	r.Request()
	recoveryWait(t, f, func() bool { return recoveryDiskMatches(r.path, image) })
	if err := r.Stop(false); err != nil {
		t.Fatal(err)
	}
	next := fileModel(t, false)
	restored := startRecovery(t, next, recoveryConfig{Directory: directory})
	next.owner(t, func() {
		if err := restored.restore(); err != nil {
			t.Fatal(err)
		}
		check := func() {
			if !maps.Equal(next.m.layerInfo, info) || !next.m.layerHidden("note-2") || !next.m.layerLocked("connector-3") || next.m.layerInfo["group-1"].Name != "Assembly" || next.m.layerInfo["block-1"].Name != "Housing" {
				t.Fatal("recovered layer names/visibility/locks differ", next.m.layerInfo)
			}
			if len(next.m.canvasForView().Objects) != 1 || len(next.m.scene.Objects) != 3 || !next.m.journal.Snapshot().Dirty {
				t.Fatal("recovered projection or dirty state")
			}
		}
		check()
		if err := next.m.journal.Undo(); err != nil || len(next.m.layerInfo) != 0 || len(next.m.scene.Objects) != 0 || next.m.journal.Snapshot().Dirty {
			t.Fatal("format4 restore undo", err)
		}
		if err := next.m.journal.Redo(); err != nil {
			t.Fatal(err)
		}
		check()
		got, err := drawingCodec(next.m).Capture()
		if err != nil || !reflect.DeepEqual(got, image) {
			t.Fatal("format4 redo capture differs", err, got)
		}
	})
}

func TestRecoveryStartupRejectsCorruptAndOversizedCopiesWithoutMutation(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"corrupt", []byte(`{"format":"vellum-placement","version":2,"zoom":1,"foreign":true}`)},
		{"oversized", bytes.Repeat([]byte("x"), drawingByteLimit+1)},
		{"trailing", []byte(`{"format":"vellum-placement","version":2,"zoom":1} {}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := fileModel(t, false)
			directory := recoveryPrivateDirectory(t)
			path := filepath.Join(directory, recoveryFilename)
			if err := os.WriteFile(path, tc.data, 0600); err != nil {
				t.Fatal(err)
			}
			before := f.m.journal.Snapshot()
			r := startRecovery(t, f, recoveryConfig{Directory: directory})
			f.owner(t, func() {
				if r.offer != nil || r.loadErr == nil || !r.pending() || r.restore() == nil || len(f.m.scene.Objects) != 0 || f.m.journal.Snapshot() != before {
					t.Fatal("invalid copy mutated document or escaped decision gate")
				}
				if err := ui.Validate(r.promptNode()); err != nil {
					t.Fatal(err)
				}
				r.discard()
			})
			recoveryWait(t, f, func() bool { _, err := os.Stat(path); return errors.Is(err, os.ErrNotExist) })
			f.owner(t, func() {
				if r.pending() || f.m.prompt != "" || f.m.journal.Snapshot() != before {
					t.Fatal("discard changed clean document")
				}
			})
		})
	}
}

type recoveryStore struct {
	read  func(context.Context, string, int64) ([]byte, error)
	write func(context.Context, string, []byte, int64) (bool, error)
}

func (s recoveryStore) Read(ctx context.Context, path string, max int64) ([]byte, error) {
	if s.read != nil {
		return s.read(ctx, path, max)
	}
	return (document.FileStore{}).Read(ctx, path, max)
}
func (s recoveryStore) Write(ctx context.Context, path string, data []byte, max int64) (bool, error) {
	if s.write != nil {
		return s.write(ctx, path, data, max)
	}
	return (document.FileStore{}).Write(ctx, path, data, max)
}

func TestRecoveryWorkerJoinsBeforeCleanRemovalAndReleasesReservations(t *testing.T) {
	f := fileModel(t, false)
	entered := make(chan struct{})
	var writes atomic.Int32
	r := startRecovery(t, f, recoveryConfig{Store: recoveryStore{write: func(ctx context.Context, path string, data []byte, max int64) (bool, error) {
		writes.Add(1)
		close(entered)
		<-ctx.Done()
		// Model the cancellation race where replacement has already committed.
		written, err := (document.FileStore{}).Write(context.Background(), path, data, max)
		return written, errors.Join(err, ctx.Err())
	}}})
	f.place(t, "block", canvas.Point{})
	r.Request()
	recoveryWait(t, f, func() bool {
		select {
		case <-entered:
			return true
		default:
			return false
		}
	})
	for i := 0; i < 1000; i++ {
		r.Request()
	}
	if err := r.Stop(true); err != nil {
		t.Fatal(err)
	}
	if writes.Load() != 1 {
		t.Fatal("coalesced requests created concurrent work", writes.Load())
	}
	if _, err := os.Stat(r.path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("late write resurrected clean-close recovery", err)
	}
	if p := f.a.Pending(); !p.Idle() {
		t.Fatal("worker leaked completion credit", p)
	}
	entries, err := os.ReadDir(filepath.Dir(r.path))
	if err != nil || len(entries) != 1 || entries[0].Name() != recoveryLockFilename {
		t.Fatal("worker leaked temporary files", entries, err)
	}
}

func TestRecoveryWindowRetirementAbortsUndispatchedCapture(t *testing.T) {
	f := fileModel(t, false)
	r := startRecovery(t, f, recoveryConfig{})
	f.place(t, "block", canvas.Point{})
	r.Request()
	deadline := time.Now().Add(5 * time.Second)
	for f.a.Pending().ReadyCompletions == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if f.a.Pending().ReadyCompletions == 0 {
		t.Fatal("capture was not queued")
	}
	if err := f.a.CloseWindow(f.ref); err != nil {
		t.Fatal(err)
	}
	if err := r.Stop(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(r.path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("retired window wrote a capture", err)
	}
	if p := f.a.Pending(); !p.Idle() {
		t.Fatal("retired worker leaked credit", p)
	}
}

func TestRecoveryDefersGesturesPromptsAndFiles(t *testing.T) {
	f := fileModel(t, false)
	r := startRecovery(t, f, recoveryConfig{})
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		for _, busy := range []string{"move", "layer", "clipboard", "placement", "prompt"} {
			switch busy {
			case "move":
				f.m.move = &objectMove{}
			case "layer":
				f.m.layerDrag = &layerDrag{}
			case "clipboard":
				f.m.clipboardJob = &clipboardOperation{}
			case "placement":
				f.m.pending = &placement.Request[component]{}
			case "prompt":
				f.m.prompt = "close"
			}
			if c := r.capture(); c.write || c.remove {
				t.Fatal("captured busy operation", busy)
			}
			f.m.move, f.m.layerDrag, f.m.clipboardJob, f.m.pending = nil, nil, nil, nil
			f.m.prompt = ""
		}
		if err := f.m.files.Save(f.ref.Context(), document.SaveOptions{Path: f.path}); err != nil {
			t.Fatal(err)
		}
		if c := r.capture(); c.write || c.remove {
			t.Fatal("captured while file completion pending")
		}
	})
	f.wait(t)
	f.place(t, "note", canvas.Point{X: 200})
	r.Request()
	image, _ := drawingCodec(f.m).Capture()
	recoveryWait(t, f, func() bool { return recoveryDiskMatches(r.path, image) })
}

func TestRecoveryFailedReplacementPreservesPreviousCopyAndDirtyRevision(t *testing.T) {
	f := fileModel(t, false)
	var fail atomic.Bool
	r := startRecovery(t, f, recoveryConfig{Store: recoveryStore{write: func(ctx context.Context, path string, data []byte, max int64) (bool, error) {
		if fail.Load() {
			return false, errors.New("injected disk failure")
		}
		return (document.FileStore{}).Write(ctx, path, data, max)
	}}})
	f.place(t, "block", canvas.Point{})
	image, _ := drawingCodec(f.m).Capture()
	r.Request()
	recoveryWait(t, f, func() bool { return recoveryDiskMatches(r.path, image) })
	fail.Store(true)
	f.place(t, "note", canvas.Point{X: 200})
	before := f.m.journal.Snapshot()
	r.Request()
	recoveryWait(t, f, func() bool { return strings.Contains(f.m.status, "injected disk failure") })
	f.owner(t, func() {
		if f.m.journal.Snapshot() != before || !before.Dirty || f.m.files.State().Path != "" {
			t.Fatal("failed recovery changed document save state")
		}
	})
	if !recoveryDiskMatches(r.path, image) {
		t.Fatal("failed replacement destroyed prior copy")
	}
}

func TestRecoveryPeriodicCaptureAndUnchangedImageDeduplication(t *testing.T) {
	f := fileModel(t, false)
	var writes atomic.Int32
	r := startRecovery(t, f, recoveryConfig{Interval: 3 * time.Millisecond, Store: recoveryStore{write: func(ctx context.Context, path string, data []byte, max int64) (bool, error) {
		writes.Add(1)
		return (document.FileStore{}).Write(ctx, path, data, max)
	}}})
	f.place(t, "connector", canvas.Point{})
	image, _ := drawingCodec(f.m).Capture()
	recoveryWait(t, f, func() bool { return recoveryDiskMatches(r.path, image) })
	// Exercise subsequent timer ticks; each drained capture sees identical bytes.
	deadline := time.Now().Add(30 * time.Millisecond)
	for time.Now().Before(deadline) {
		f.a.Drain()
		time.Sleep(time.Millisecond)
	}
	if err := r.Stop(false); err != nil {
		t.Fatal(err)
	}
	if writes.Load() != 1 {
		t.Fatal("unchanged bytes were repeatedly written", writes.Load())
	}
}

func TestRecoveryStartupReservationRefusalLeavesModelUngated(t *testing.T) {
	f := fileModel(t, false)
	var tickets []*app.Completion
	for {
		ticket, err := f.a.ReserveWindowCompletion(f.ref, nil)
		if errors.Is(err, app.ErrCompletionCapacity) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		tickets = append(tickets, ticket)
	}
	defer func() {
		for _, ticket := range tickets {
			ticket.Abort()
		}
	}()
	f.owner(t, func() {
		r, err := newPlacementRecovery(f.a, f.ref, f.m, recoveryConfig{Directory: t.TempDir()})
		if !errors.Is(err, app.ErrCompletionCapacity) || r != nil || f.m.prompt != "" {
			t.Fatal("refused startup left an unresolvable gate", r, err, f.m.prompt)
		}
	})
}

func TestRecoveryReadRefusesSymlinkAndSharedFileBeforeStore(t *testing.T) {
	for _, kind := range []string{"symlink", "shared"} {
		t.Run(kind, func(t *testing.T) {
			f := fileModel(t, false)
			directory := recoveryPrivateDirectory(t)
			path := filepath.Join(directory, recoveryFilename)
			outside := filepath.Join(t.TempDir(), "original.vellum")
			data, err := drawingCodec(f.m).Encode(drawingFile{Format: "vellum-placement", Version: 2, Zoom: 1})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(outside, data, 0600); err != nil {
				t.Fatal(err)
			}
			if kind == "symlink" {
				if err := os.Symlink(outside, path); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(path, 0644); err != nil {
					t.Fatal(err)
				}
			}
			var reads atomic.Int32
			r := startRecovery(t, f, recoveryConfig{Directory: directory, Store: recoveryStore{read: func(ctx context.Context, path string, limit int64) ([]byte, error) {
				reads.Add(1)
				return (document.FileStore{}).Read(ctx, path, limit)
			}}})
			f.owner(t, func() {
				if reads.Load() != 0 || r.offer != nil || r.loadErr == nil || !r.pending() || len(f.m.scene.Objects) != 0 {
					t.Fatal("unsafe copy reached decoder/model")
				}
				r.discard()
			})
			recoveryWait(t, f, func() bool { _, err := os.Lstat(path); return errors.Is(err, os.ErrNotExist) })
			got, err := os.ReadFile(outside)
			if err != nil || !bytes.Equal(got, data) {
				t.Fatal("discard followed symlink/outside file", err)
			}
		})
	}
}

func TestRecoveryUnavailableDirectoryDoesNotGateEditing(t *testing.T) {
	f := fileModel(t, false)
	path := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	r := startRecovery(t, f, recoveryConfig{Directory: path})
	f.owner(t, func() {
		if r.pending() || f.m.prompt != "" || !strings.Contains(f.m.status, "Recovery unavailable") {
			t.Fatal("unavailable cache blocked drawing", f.m.status)
		}
	})
	f.place(t, "block", canvas.Point{})
	if err := r.Stop(true); err != nil {
		t.Fatal("unavailable cache cleanup should be a no-op", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "keep" {
		t.Fatal("unavailable path was removed", err)
	}
}

func TestRecoveryCloseHooksPreservePendingAndRemoveExplicitDiscard(t *testing.T) {
	t.Run("pending keeps copy", func(t *testing.T) {
		f := fileModel(t, false)
		directory := recoveryPrivateDirectory(t)
		path := filepath.Join(directory, recoveryFilename)
		data, err := drawingCodec(f.m).Encode(drawingFile{Format: "vellum-placement", Version: 2, Zoom: 1})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		r := startRecovery(t, f, recoveryConfig{Directory: directory})
		f.owner(t, func() {
			if !f.m.requestClose(f.a) || f.m.removeRecoveryOnClose || f.m.filePrompt(f.a, f.ref).Open {
				t.Fatal("pending close lost decision/copy policy")
			}
			if err := f.m.fileAction(f.a, f.ref, "open", true); !errors.Is(err, document.ErrBusy) || f.m.files.State().Busy {
				t.Fatal("direct file action bypassed pending recovery", err)
			}
		})
		if err := r.Stop(f.m.removeRecoveryOnClose); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, data) {
			t.Fatal("pending close deleted copy", err)
		}
	})
	t.Run("explicit dirty discard removes copy", func(t *testing.T) {
		f := fileModel(t, false)
		r := startRecovery(t, f, recoveryConfig{})
		f.place(t, "note", canvas.Point{})
		image, _ := drawingCodec(f.m).Capture()
		r.Request()
		recoveryWait(t, f, func() bool { return recoveryDiskMatches(r.path, image) })
		f.owner(t, func() {
			if f.m.requestClose(f.a) || f.m.prompt != "close" || f.m.removeRecoveryOnClose {
				t.Fatal("dirty close skipped explicit decision")
			}
			for _, node := range f.m.filePrompt(f.a, f.ref).Children {
				if node.ID == "file-discard" {
					node.Activate()
				}
			}
			if !f.m.removeRecoveryOnClose || f.ref.Context().Err() == nil {
				t.Fatal("discard did not approve cleanup/close")
			}
		})
		if err := r.Stop(f.m.removeRecoveryOnClose); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(r.path); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("explicit close left stale copy", err)
		}
	})
}

func TestRecoveryDirectoryLockIsExclusiveUntilStopAndStopIsIdempotent(t *testing.T) {
	directory := recoveryPrivateDirectory(t)
	first := fileModel(t, false)
	r := startRecovery(t, first, recoveryConfig{Directory: directory})
	first.place(t, "note", canvas.Point{X: 40})
	image, _ := drawingCodec(first.m).Capture()
	r.Request()
	recoveryWait(t, first, func() bool { return recoveryDiskMatches(r.path, image) })
	second := fileModel(t, false)
	blocked := startRecovery(t, second, recoveryConfig{Directory: directory, Interval: time.Millisecond})
	second.owner(t, func() {
		if blocked.pending() || second.m.prompt != "" || !strings.Contains(second.m.status, "in use") {
			t.Fatal("second instance did not remain editable with recovery unavailable", second.m.status)
		}
	})
	second.place(t, "block", canvas.Point{X: 200})
	blocked.Request()
	// A denied instance must neither overwrite nor delete the owner's copy.
	if err := blocked.Stop(true); err != nil {
		t.Fatal(err)
	}
	if !recoveryDiskMatches(r.path, image) {
		t.Fatal("denied instance changed the active owner's recovery")
	}
	if err := r.Stop(false); err != nil {
		t.Fatal(err)
	}
	third := fileModel(t, false)
	available := startRecovery(t, third, recoveryConfig{Directory: directory})
	third.owner(t, func() {
		if available.offer == nil || !reflect.DeepEqual(*available.offer, image) {
			t.Fatal("released lock did not permit a new owner to recover")
		}
	})
	// An old deferred Stop must not act on a slot already owned by a new run.
	if err := r.Stop(true); err != nil {
		t.Fatal(err)
	}
	if !recoveryDiskMatches(r.path, image) {
		t.Fatal("repeated old Stop deleted a new owner's recovery")
	}
}
