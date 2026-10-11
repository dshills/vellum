//go:build offscreen

package replayconsumer_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	. "github.com/dshills/vellum/drivers/cogent"
	consumer "github.com/dshills/vellum/examples/replay"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

var testContext = context.Background()

func manualConfig(t *testing.T, batch int) ManualConfig {
	t.Helper()
	clock, err := execution.NewManualClock(time.Unix(1730000000, 0).UTC(), 256, 256)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(clock.Close)
	executor, err := execution.NewManual(8)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(executor.Close)
	config := ManualConfig{Clock: clock, Executor: executor, OwnerBatch: batch, MaxWindows: 4, EventCapacity: 128, LogicalDPI: 160, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{R: 24, G: 24, B: 24, A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}}
	err = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf") {
			return nil
		}
		data, err := fs.ReadFile(fonts.Default, path)
		if err != nil {
			return err
		}
		h := sha256.Sum256(data)
		config.Fonts = append(config.Fonts, ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(h[:])})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return config
}
func fixtureManual(t *testing.T, batch int) (*Manual, *consumer.Fixture, ManualConfig) {
	t.Helper()
	config := manualConfig(t, batch)
	a, err := app.New(app.Config{QueueCapacity: 8, CompletionCapacity: 4})
	if err != nil {
		t.Fatal(err)
	}
	f, err := consumer.New(consumer.Config{Owner: a, Executor: config.Executor})
	if err != nil {
		t.Fatal(err)
	}
	config.Commands = f.Commands
	m, err := NewManual(testContext, a, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(testContext); err != nil {
			t.Error(err)
		}
	})
	return m, f, config
}
func settleManual(t *testing.T, m *Manual, batch int) ManualSnapshot {
	t.Helper()
	for range 512 {
		s, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
			if w.Blocked && !w.ReadyPaint {
				t.Fatal("blocked originating FIFO", w)
			}
		}
		if ready {
			r, err := m.Step(testContext, batch)
			if err != nil {
				t.Fatal(err)
			}
			if r.Events+r.OwnerCallbacks > batch {
				t.Fatal("step exceeded combined budget", r, batch)
			}
			continue
		}
		framed := false
		for _, w := range s.Windows {
			if w.LayoutPending {
				r, err := m.Frame(testContext, w.ID)
				if err != nil {
					t.Fatal(err)
				}
				framed = framed || r.Events > 0
			}
		}
		if framed {
			continue
		}
		return s
	}
	t.Fatal("bounded current-instant settle exhausted")
	return ManualSnapshot{}
}
func measured(t *testing.T, s ManualSnapshot, id string) (ManualWindowSnapshot, ManualElement) {
	t.Helper()
	for _, w := range s.Windows {
		for _, e := range w.Elements {
			if e.ID == id {
				return w, e
			}
		}
	}
	t.Fatal("no measured control", id, s)
	return ManualWindowSnapshot{}, ManualElement{}
}
func clickManual(t *testing.T, m *Manual, batch int, id string) {
	t.Helper()
	s := settleManual(t, m, batch)
	w, e := measured(t, s, id)
	if e.Bounds.Empty() {
		t.Fatal("empty measured control", id, e)
	}
	// Both primitive events use the same captured coordinates; actual Core focus,
	// hit-testing, disabled and modal routing decide what executes.
	position := e.Bounds.Min.Add(e.Bounds.Size().Div(2))
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: typ, Button: events.Left, Position: position}); err != nil {
			t.Fatal(err)
		}
	}
	settleManual(t, m, batch)
}
func typeManual(t *testing.T, m *Manual, batch int, window string, text string) {
	t.Helper()
	s := settleManual(t, m, batch)
	var incarnation uint64
	for _, w := range s.Windows {
		if w.ID == window {
			incarnation = w.Incarnation
		}
	}
	var modifiers key.Modifiers
	modifiers.SetFlag(true, key.Control)
	for _, input := range append([]ManualInput{{Type: events.KeyChord, Rune: 'a', Code: key.CodeA, Modifiers: modifiers}}, func() []ManualInput {
		result := []ManualInput{}
		for _, r := range text {
			result = append(result, ManualInput{Type: events.KeyChord, Rune: r})
		}
		return append(result, ManualInput{Type: events.KeyChord, Code: key.CodeReturnEnter})
	}()...) {
		input.Window = window
		input.WindowIncarnation = incarnation
		if err := m.Input(testContext, input); err != nil {
			t.Fatal(err)
		}
		settleManual(t, m, batch)
	}
}
func ownerSnapshot(t *testing.T, f *consumer.Fixture) consumer.Snapshot {
	t.Helper()
	var s consumer.Snapshot
	if err := f.Owner.Do(func() { s = f.Snapshot() }); err != nil {
		t.Fatal(err)
	}
	return s
}
func TestManualRealManagedProjectInput(t *testing.T) {
	m, f, config := fixtureManual(t, 7)
	initial, err := m.Snapshot(testContext)
	if err != nil || len(initial.Windows) != 1 || initial.Windows[0].Frames != 0 {
		t.Fatalf("construction silently framed %+v %v", initial, err)
	}
	if _, err = m.Capture(testContext, f.Window); !errors.Is(err, ErrManualLayout) {
		t.Fatal("uncaptured image accepted", err)
	}
	clickManual(t, m, 7, consumer.OpenProject)
	if s := ownerSnapshot(t, f); !s.Loaded || s.Property != "North Star" {
		t.Fatal("real pointer did not open fixture", s)
	}
	clickManual(t, m, 7, "property-1")
	typeManual(t, m, 7, f.Window, "Harbor")
	if s := ownerSnapshot(t, f); s.Property != "Harbor" || s.UndoCount != 1 {
		t.Fatal("real field/router Change failed", s)
	}
	clickManual(t, m, 7, consumer.SortTable)
	if s := ownerSnapshot(t, f); !s.Sorted || s.VisibleCells[0][0] != "Alpha" {
		t.Fatal("real sort failed", s)
	}
	clickManual(t, m, 7, "table-row-0")
	clickManual(t, m, 7, consumer.StartTask)
	if config.Executor.Snapshot().Pending != 1 {
		t.Fatal("worker was implicitly executed", config.Executor.Snapshot())
	}
	clickManual(t, m, 7, consumer.CancelTask)
	if _, err = config.Executor.StepNext(testContext); err != nil {
		t.Fatal(err)
	}
	settleManual(t, m, 7)
	s := ownerSnapshot(t, f)
	if len(s.Results) != 1 || s.Results[0].Outcome != "canceled" || s.WorkerCleanups != 1 {
		t.Fatal("reserved completion/explicit canceled worker failed", s)
	}
	clickManual(t, m, 7, consumer.Undo)
	if s = ownerSnapshot(t, f); s.Property != "North Star" || s.Dirty {
		t.Fatal("routed undo failed", s)
	}
	captured, err := m.Capture(testContext, f.Window)
	if err != nil {
		t.Fatal(err)
	}
	if captured.Bounds() != image.Rect(0, 0, consumer.ViewportWidth, consumer.ViewportHeight) {
		t.Fatal("software viewport mismatch", captured.Bounds())
	}
	original := captured.Pix[0]
	captured.Pix[0] ^= 0xff
	again, err := m.Capture(testContext, f.Window)
	if err != nil || again.Pix[0] != original {
		t.Fatal("capture aliases composer", err)
	}
	if !config.Clock.Now().Equal(time.Unix(1730000000, 0).UTC()) {
		t.Fatal("input/step/frame advanced clock")
	}
	t.Logf("actual managed software frame SHA256 %x; %d declared UI action kinds", sha256.Sum256(again.Pix), len(s.UIAttempts))
}

func actionCount(s consumer.Snapshot, id string) int {
	for _, v := range s.UIAttempts {
		if v.ID == id {
			return v.Count
		}
	}
	return 0
}
func rawClick(t *testing.T, m *Manual, batch int, w ManualWindowSnapshot, p image.Point) {
	t.Helper()
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: typ, Button: events.Left, Position: p}); err != nil {
			t.Fatal(err)
		}
	}
	settleManual(t, m, batch)
}
func TestManualModalDisabledAndStaleRealRouter(t *testing.T) {
	m, f, _ := fixtureManual(t, 1)
	clickManual(t, m, 1, consumer.OpenProject)
	clickManual(t, m, 1, "property-disabled")
	s := settleManual(t, m, 1)
	w, field := measured(t, s, "property-1")
	if field.Enabled {
		t.Fatal("actual disabled field enabled in capture")
	}
	if err := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Target: field.ID, TargetIncarnation: field.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(err, ErrManualDisabled) {
		t.Fatal("disabled target accepted", err)
	}
	before := ownerSnapshot(t, f)
	rawClick(t, m, 1, w, field.Bounds.Min.Add(field.Bounds.Size().Div(2)))
	typeManual(t, m, 1, f.Window, "Ignored")
	if after := ownerSnapshot(t, f); after.Property != before.Property || actionCount(after, consumer.EditProperty) != actionCount(before, consumer.EditProperty) {
		t.Fatal("actual disabled hit/key route edited model", after)
	}
	clickManual(t, m, 1, "property-disabled")
	clickManual(t, m, 1, consumer.OpenModal)
	s = settleManual(t, m, 1)
	w, background := measured(t, s, consumer.SortTable)
	_, modal := measured(t, s, consumer.CloseModal)
	if background.Visible || !background.Enabled || !modal.Visible {
		t.Fatal("modal projection incorrect", background, modal)
	}
	before = ownerSnapshot(t, f)
	rawClick(t, m, 1, w, background.Bounds.Min.Add(background.Bounds.Size().Div(2)))
	if after := ownerSnapshot(t, f); after.Sorted || actionCount(after, consumer.SortTable) != actionCount(before, consumer.SortTable) {
		t.Fatal("modal did not intercept actual background hit", after)
	}
	// Real dialog ancestor keyboard handler dismisses Escape; no Dismiss call bypass.
	if err := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeEscape}); err != nil {
		t.Fatal(err)
	}
	settleManual(t, m, 1)
	if ownerSnapshot(t, f).Modal {
		t.Fatal("modal Escape did not dismiss")
	}
	s = settleManual(t, m, 1)
	w, oldField := measured(t, s, "property-1")
	clickManual(t, m, 1, consumer.RemountProperty)
	s = settleManual(t, m, 1)
	_, newField := measured(t, s, "property-2")
	if newField.Incarnation == oldField.Incarnation {
		t.Fatal("remount identity not actual", oldField, newField)
	}
	clickManual(t, m, 1, "property-2")
	_, newField = measured(t, settleManual(t, m, 1), "property-2")
	if !newField.Focused {
		t.Fatal("actual pointer refocus failed", newField)
	}
	if err := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Target: oldField.ID, TargetIncarnation: oldField.Incarnation, Type: events.MouseDown, Button: events.Left}); !errors.Is(err, ErrManualStale) {
		t.Fatal("removed target admitted", err)
	}
	if err := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation + 1, Type: events.KeyChord, Rune: 'x'}); !errors.Is(err, ErrManualStale) {
		t.Fatal("wrong window incarnation admitted", err)
	}
	if err := m.Close(testContext); err != nil {
		t.Fatal(err)
	}
	if f.Owner.Context().Err() == nil {
		t.Fatal("final managed closure did not close domain")
	}
	if err := m.Input(testContext, ManualInput{}); !errors.Is(err, ErrManualClosed) {
		t.Fatal("closed adapter accepted", err)
	}
}

func runMeasuredScenario(t *testing.T, batch int) (consumer.Snapshot, []ManualElement, [32]byte) {
	t.Helper()
	m, f, config := fixtureManual(t, batch)
	clickManual(t, m, batch, consumer.OpenProject)
	clickManual(t, m, batch, "property-1")
	typeManual(t, m, batch, f.Window, "Harbor")
	clickManual(t, m, batch, consumer.SortTable)
	clickManual(t, m, batch, "table-row-0")
	clickManual(t, m, batch, consumer.StartTask)
	clickManual(t, m, batch, consumer.CancelTask)
	if _, err := config.Executor.StepNext(testContext); err != nil {
		t.Fatal(err)
	}
	settleManual(t, m, batch)
	clickManual(t, m, batch, consumer.Undo)
	model := ownerSnapshot(t, f)
	s := settleManual(t, m, batch)
	img, err := m.Capture(testContext, f.Window)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.Close(testContext); err != nil {
		t.Fatal(err)
	}
	return model, s.Windows[0].Elements, sha256.Sum256(img.Pix)
}
func TestManualMeasuredBatchEquivalence(t *testing.T) {
	var model consumer.Snapshot
	var elements []ManualElement
	var digest [32]byte
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			current, measured, imageHash := runMeasuredScenario(t, batch)
			if batch == 1 {
				model, elements, digest = current, measured, imageHash
			} else {
				if !reflect.DeepEqual(model, current) {
					t.Fatalf("model mismatch by actual event batch\n%+v\n%+v", model, current)
				}
				if !reflect.DeepEqual(elements, measured) {
					t.Fatalf("measured layout/focus/selection mismatch\n%+v\n%+v", elements, measured)
				}
				if digest != imageHash {
					t.Fatalf("software image mismatch %x != %x", digest, imageHash)
				}
			}
			t.Logf("batch=%d actual frame sha256=%x", batch, imageHash)
		})
	}
}
func TestManualAdmissionCaptureAndTerminalWork(t *testing.T) {
	m, f, config := fixtureManual(t, 7)
	settleManual(t, m, 7)
	if _, err := NewManual(testContext, f.Owner, config); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("second adapter/native services takeover accepted", err)
	}
	canceled, cancel := context.WithCancel(testContext)
	cancel()
	before, _ := m.Snapshot(testContext)
	if err := m.Input(canceled, ManualInput{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	after, _ := m.Snapshot(testContext)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("refused input mutated readiness")
	}
	if err := f.Owner.Invalidate(f.Window); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Capture(testContext, f.Window); !errors.Is(err, ErrManualLayout) {
		t.Fatal("model-dirty capture accepted", err)
	}
	settleManual(t, m, 7)
	w, _ := measured(t, settleManual(t, m, 7), consumer.OpenProject)
	var control key.Modifiers
	control.SetFlag(true, key.Control)
	if err := m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeV, Modifiers: control}); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("unqualified OS paste admitted", err)
	}
	services, err := f.Owner.Services()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = services.ChooseFile(testContext, app.FileOptions{}); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("native chooser wasn't refused", err)
	}
	clickManual(t, m, 7, consumer.OpenProject)
	clickManual(t, m, 7, consumer.StartTask)
	if err = m.Close(canceled); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled cleanup admission mutated", err)
	}
	if err = m.Close(testContext); err != nil {
		t.Fatal(err)
	}
	if config.Executor.Snapshot().Pending != 1 {
		t.Fatal("adapter silently discarded accepted worker", config.Executor.Snapshot())
	}
	if _, err = config.Executor.StepNext(testContext); err != nil {
		t.Fatal(err)
	}
	s, err := m.Snapshot(testContext)
	if err != nil || !s.Closed || len(s.Windows) != 0 || s.Executor.Pending+s.Executor.Running != 0 || s.Application.Cleanup+s.Application.Reserved+s.Application.ReadyCompletions != 0 {
		t.Fatalf("closed cleanup/work not terminal %+v %v", s, err)
	}
}
func TestManualConcurrentAndReentrantRefusal(t *testing.T) {
	config := manualConfig(t, 7)
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	var m *Manual
	var reentrant error
	if err = a.Open(app.Window{ID: "one", Title: "Concurrency", Width: 400, Height: 220, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "button", Kind: ui.Button, Text: "Block", Activate: func() { _, reentrant = m.Snapshot(testContext); close(entered); <-release }}}}
	}}); err != nil {
		t.Fatal(err)
	}
	m, err = NewManual(testContext, a, config)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(testContext)
	s := settleManual(t, m, 7)
	w, e := measured(t, s, "button")
	p := e.Bounds.Min.Add(e.Bounds.Size().Div(2))
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err = m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: typ, Button: events.Left, Position: p}); err != nil {
			t.Fatal(err)
		}
	}
	finished := make(chan error, 1)
	go func() {
		for range 64 {
			r, err := m.Step(testContext, 7)
			if err != nil {
				finished <- err
				return
			}
			if r.Events+r.OwnerCallbacks == 0 {
				break
			}
		}
		finished <- nil
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("actual button event never entered")
	}
	if _, err = m.Snapshot(testContext); !errors.Is(err, ErrManualBusy) {
		t.Fatal("concurrent read was not refused", err)
	}
	if err = m.Close(testContext); !errors.Is(err, ErrManualBusy) {
		t.Fatal("concurrent close was not refused", err)
	}
	close(release)
	if err = <-finished; err != nil {
		t.Fatal(err)
	}
	if !errors.Is(reentrant, ErrManualBusy) {
		t.Fatal("callback reentry allowed", reentrant)
	}
}
