//go:build offscreen

package replayconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/replay"
)

var nativeContext = context.Background()

func nativeFixture(t *testing.T, batch int) *NativeBackend {
	t.Helper()
	b, err := NewNative(nativeContext, batch)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := b.Close(nativeContext); err != nil {
			t.Error(err)
		}
	})
	return b
}

type nativePilot struct {
	t      *testing.T
	b      *NativeBackend
	r      *replay.Runner
	record replay.Record
}

func newPilot(t *testing.T) *nativePilot {
	b := nativeFixture(t, 7)
	r, err := replay.NewRunner(replay.RunnerConfig{EventBatch: 7}, b)
	if err != nil {
		t.Fatal(err)
	}
	return &nativePilot{t: t, b: b, r: r, record: replay.Record{Manifest: b.Manifest(), Operations: []replay.Operation{}}}
}
func (p *nativePilot) window() cogent.ManualWindowSnapshot {
	p.t.Helper()
	s, err := p.b.Driver.Snapshot(nativeContext)
	if err != nil || len(s.Windows) != 1 {
		p.t.Fatal(s, err)
	}
	return s.Windows[0]
}
func (p *nativePilot) emit(o replay.Operation) replay.Result {
	p.t.Helper()
	o.At = p.b.Clock.Now().Sub(p.record.Manifest.Start)
	if o.Kind != replay.Worker {
		w := p.window()
		o.Window = w.ID
		o.WindowIncarnation = w.Incarnation
	}
	o.Sequence = uint64(len(p.record.Operations)) + 1
	p.record.Operations = append(p.record.Operations, o)
	// Pilot measures identities/positions once. Replay below runs the one frozen
	// resulting record; it never reauthors actions from a batch-specific model.
	one := o
	one.Sequence = 1
	got, err := p.r.Run(nativeContext, replay.Record{Manifest: p.record.Manifest, Operations: []replay.Operation{one}})
	if err != nil {
		p.t.Fatalf("pilot operation %d %s: %+v %v", o.Sequence, o.Kind, got, err)
	}
	return got
}
func (p *nativePilot) frame() { p.t.Helper(); p.emit(replay.Operation{Kind: replay.Frame}) }
func (p *nativePilot) capture(label string) {
	p.t.Helper()
	p.emit(replay.Operation{Kind: replay.Capture, Label: label})
}
func (p *nativePilot) click(id string) {
	p.t.Helper()
	w := p.window()
	if w.LayoutPending {
		p.t.Fatal("pilot target lacks explicit frame", id)
	}
	for _, e := range w.Elements {
		if e.ID != id {
			continue
		}
		if e.Bounds.Empty() || !e.Visible || !e.Enabled {
			p.t.Fatal("unavailable target", e)
		}
		pos := e.Bounds.Min.Add(e.Bounds.Size().Div(2))
		p.emit(replay.Operation{Kind: replay.PointerDown, Target: e.ID, TargetIncarnation: e.Incarnation, X: float64(pos.X), Y: float64(pos.Y), Button: 1})
		// Release uses the admitted physical point through the real router. Press
		// feedback may dirty the layout, so there is no stale target resolution.
		p.emit(replay.Operation{Kind: replay.PointerUp, X: float64(pos.X), Y: float64(pos.Y), Button: 1})
		p.frame()
		return
	}
	p.t.Fatal("pilot missing target", id)
}
func (p *nativePilot) key(value string, modifiers uint32) {
	p.t.Helper()
	p.emit(replay.Operation{Kind: replay.KeyChord, Key: value, Modifiers: modifiers})
	p.frame()
}
func authorNativeRecord(t *testing.T) replay.Record {
	t.Helper()
	p := newPilot(t)
	p.frame()
	p.capture("initial")
	p.click(OpenProject)
	p.capture("opened")
	p.click("property-1")
	p.key("a", 2)
	for _, r := range "Harbor" {
		p.key(string(r), 0)
	}
	p.key("Enter", 0)
	p.capture("edited")
	p.click(SortTable)
	p.click("table-row-0")
	p.capture("sorted-selected")
	p.click(StartTask)
	p.capture("worker-admitted")
	p.click(CancelTask)
	p.capture("cancel-requested")
	ids := p.b.Executor.PendingIDs()
	if len(ids) != 1 {
		t.Fatal("pilot pending worker", ids)
	}
	p.emit(replay.Operation{Kind: replay.Worker, WorkerID: uint64(ids[0])})
	p.frame()
	p.capture("worker-delivered")
	p.click(Undo)
	p.capture("undone")
	p.emit(replay.Operation{Kind: replay.Advance, Delta: time.Second})
	p.frame()
	p.capture("final")
	if err := p.record.Validate(replay.Limits{}); err != nil {
		t.Fatal(err)
	}
	return p.record
}

func TestActualNativeRecordReplaysAcrossEventBatches(t *testing.T) {
	record, err := NativeRecord()
	if err != nil {
		t.Fatal(err)
	}
	var expected []replay.Checkpoint
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			b := nativeFixture(t, batch)
			r, err := replay.NewRunner(replay.RunnerConfig{EventBatch: batch}, b)
			if err != nil {
				t.Fatal(err)
			}
			got, err := r.Run(nativeContext, record)
			if err != nil {
				t.Fatal(got, err)
			}
			if got.Applied != len(record.Operations) || !got.QuiescentNow {
				t.Fatal("record not complete/current-quiescent", got)
			}
			if len(got.Checkpoints) != 9 {
				t.Fatal("missing native captures", len(got.Checkpoints))
			}
			if expected == nil {
				expected = got.Checkpoints
			} else if !reflect.DeepEqual(expected, got.Checkpoints) {
				for i := range got.Checkpoints {
					if !reflect.DeepEqual(expected[i], got.Checkpoints[i]) {
						t.Errorf("capture %s differs: expected image %s got %s", got.Checkpoints[i].Label, expected[i].ImageSHA256, got.Checkpoints[i].ImageSHA256)
						t.Log(string(expected[i].Data))
						t.Log(string(got.Checkpoints[i].Data))
					}
				}
			}
			var final NativeCapture
			if err := json.Unmarshal(got.Checkpoints[len(got.Checkpoints)-1].Data, &final); err != nil {
				t.Fatal(err)
			}
			if final.Model.Property != "North Star" || !final.Model.Sorted || final.Model.Selected != "r-alpha" || final.Model.Dirty || final.Model.WorkerCleanups != 1 || len(final.Model.Results) != 1 || final.Model.Results[0].Outcome != "canceled" {
				t.Fatal("wrong native final model", final.Model)
			}
			if len(final.Model.UIAttempts) == 0 || len(final.Elements) == 0 || final.Viewport.Dx() != ViewportWidth || final.Viewport.Dy() != ViewportHeight {
				t.Fatal("missing actual UI/frame evidence", final)
			}
			for _, p := range got.Checkpoints {
				if len(p.ImageSHA256) != 64 {
					t.Fatal("missing image identity")
				}
			}
			if !b.Clock.Now().Equal(record.Manifest.Start.Add(time.Second)) {
				t.Fatal("unexpected scheduling time")
			}
			t.Logf("%d operations; %d capture comparisons; final image %s; progress %+v", got.Applied, len(got.Checkpoints), got.Checkpoints[len(got.Checkpoints)-1].ImageSHA256, got.Progress)
		})
	}
}

func TestBundledNativeRecordMatchesAuthoringFixture(t *testing.T) {
	// Deliberately separate from replay: authoring measures one pilot run, while
	// acceptance above always executes the immutable bundled record.
	got := authorNativeRecord(t)
	expected, err := NativeRecord()
	if err != nil || !reflect.DeepEqual(expected, got) {
		t.Fatal("bundled record no longer matches measured fixture", err)
	}
}

func TestNativeManifestAndLiveRevisionRefusals(t *testing.T) {
	b := nativeFixture(t, 1)
	r, err := replay.NewRunner(replay.RunnerConfig{}, b)
	if err != nil {
		t.Fatal(err)
	}
	m := b.Manifest()
	m.ThemeSHA256 = m.FixtureSHA256
	got, err := r.Run(nativeContext, replay.Record{Manifest: m})
	if !errors.Is(err, replay.ErrBackend) || got.Applied != 0 {
		t.Fatal(got, err)
	}
	m = b.Manifest()
	o := replay.Operation{Sequence: 1, Kind: replay.Command, Window: m.Windows[0].ID, WindowIncarnation: m.Windows[0].Incarnation, Action: OpenProject, CheckRevision: true, ExpectedRevision: 999}
	got, err = r.Run(nativeContext, replay.Record{Manifest: m, Operations: []replay.Operation{o}})
	if !errors.Is(err, ErrStale) || got.Applied != 0 {
		t.Fatal(got, err)
	}
	s, err := b.Model()
	if err != nil || s.Loaded {
		t.Fatal("refusal mutated model", s, err)
	}
}

func TestNativeCaptureFrameAndStaleAdmission(t *testing.T) {
	b := nativeFixture(t, 1)
	manifest := b.Manifest()
	w := manifest.Windows[0]
	if err := b.CheckManifest(nativeContext, manifest); err != nil {
		t.Fatal(err)
	}
	capOp := replay.Operation{Kind: replay.Capture, Window: w.ID, WindowIncarnation: w.Incarnation, Label: "sample"}
	if _, err := b.Capture(nativeContext, capOp); !errors.Is(err, cogent.ErrManualLayout) {
		t.Fatal("unrendered capture admitted", err)
	}
	before, err := b.Driver.Snapshot(nativeContext)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Driver.RequestFrame(nativeContext, w.ID); err != nil {
		t.Fatal(err)
	}
	after, err := b.Driver.Snapshot(nativeContext)
	if err != nil {
		t.Fatal(err)
	}
	if after.Windows[0].Frames != before.Windows[0].Frames || !after.Clock.Now.Equal(before.Clock.Now) {
		t.Fatal("frame request dispatched/advanced", before, after)
	}
	r, _ := replay.NewRunner(replay.RunnerConfig{EventBatch: 1}, b)
	if _, err := r.RunUntilIdle(nativeContext); err != nil {
		t.Fatal(err)
	}
	s, err := b.Driver.Snapshot(nativeContext)
	if err != nil {
		t.Fatal(err)
	}
	var target cogent.ManualElement
	for _, e := range s.Windows[0].Elements {
		if e.ID == OpenProject {
			target = e
		}
	}
	if target.ID == "" {
		t.Fatal("no measured target")
	}
	p := target.Bounds.Min.Add(target.Bounds.Size().Div(2))
	op := replay.Operation{Kind: replay.PointerDown, Window: w.ID, WindowIncarnation: w.Incarnation, Target: target.ID, TargetIncarnation: target.Incarnation + 1, X: float64(p.X), Y: float64(p.Y), Button: 1}
	if err := b.Apply(nativeContext, op); !errors.Is(err, cogent.ErrManualStale) {
		t.Fatal("stale target admitted", err)
	}
	state, err := b.Model()
	if err != nil || state.Loaded || len(state.UIAttempts) != 0 {
		t.Fatal("refusal invoked widget callback", state, err)
	}
	if _, err := b.Capture(nativeContext, capOp); err != nil {
		t.Fatal("stable frame unavailable", err)
	}
	if err := b.Fixture.Owner.CloseWindow(w.ID); err != nil {
		t.Fatal(err)
	}
	now := b.Clock.Now()
	if err := b.Apply(nativeContext, replay.Operation{Kind: replay.Advance, Window: w.ID, WindowIncarnation: w.Incarnation, Delta: time.Second}); !errors.Is(err, app.ErrWindowClosed) && !errors.Is(err, app.ErrClosed) {
		t.Fatal("closed domain advance accepted", err)
	}
	if !b.Clock.Now().Equal(now) {
		t.Fatal("closed domain refusal advanced clock")
	}
}

func TestNativeWindowlessWorkerRefusesClosedOwner(t *testing.T) {
	b := nativeFixture(t, 1)
	m := b.Manifest()
	if err := b.CheckManifest(nativeContext, m); err != nil {
		t.Fatal(err)
	}
	var invokeErr error
	if err := b.Fixture.Owner.Do(func() {
		invokeErr = b.Fixture.Invoke(nativeContext, Action{ID: OpenProject})
		if invokeErr == nil {
			invokeErr = b.Fixture.Invoke(nativeContext, Action{ID: StartTask})
		}
	}); err != nil || invokeErr != nil {
		t.Fatal(err, invokeErr)
	}
	ids := b.Executor.PendingIDs()
	if len(ids) != 1 {
		t.Fatal(ids)
	}
	worker := replay.Operation{Sequence: 1, Kind: replay.Worker, WorkerID: uint64(ids[0])}
	if err := (replay.Record{Manifest: m, Operations: []replay.Operation{worker}}).Validate(replay.Limits{}); err != nil {
		t.Fatal(err)
	}
	b.Fixture.Owner.Close()
	before := b.Executor.Snapshot()
	now := b.Clock.Now()
	if err := b.Apply(nativeContext, worker); !errors.Is(err, app.ErrClosed) {
		t.Fatal("closed owner global worker admitted", err)
	}
	if !reflect.DeepEqual(before, b.Executor.Snapshot()) || !b.Clock.Now().Equal(now) || b.Fixture.cleanups.Load() != 0 {
		t.Fatal("refused worker executed/mutated", before, b.Executor.Snapshot())
	}
	// Executor remains caller-owned. Explicitly run its canceled accepted job
	// outside the closed backend to terminate that worker's own cleanup tail.
	if err := b.Executor.Step(nativeContext, ids[0]); err != nil {
		t.Fatal(err)
	}
	if b.Fixture.cleanups.Load() != 1 || !b.Executor.Snapshot().Idle() {
		t.Fatal("explicit worker teardown did not terminate")
	}
}
