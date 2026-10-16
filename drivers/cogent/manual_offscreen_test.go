//go:build offscreen

package cogent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"

	"image/color"
	"io/fs"
	"runtime"

	"strings"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/system"
	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
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

func simpleManual(t *testing.T, config ManualConfig, view func() ui.Node) (*Manual, *app.Application) {
	t.Helper()
	a, err := app.New(app.Config{QueueCapacity: 4})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Open(app.Window{ID: "one", Title: "Manual unit", Width: 400, Height: 240, View: view}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(testContext); err != nil {
			t.Error(err)
		}
	})
	return m, a
}
func unitSettle(t *testing.T, m *Manual) {
	t.Helper()
	for range 128 {
		s, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
		}
		if ready {
			if _, err = m.Step(testContext, 7); err != nil {
				t.Fatal(err)
			}
			continue
		}
		pending := false
		for _, w := range s.Windows {
			if w.LayoutPending {
				if _, err = m.Frame(testContext, w.ID); err != nil {
					t.Fatal(err)
				}
				pending = true
			}
		}
		if !pending {
			return
		}
	}
	t.Fatal("unit settle bound exhausted")
}
func TestManualConfigurationIdentityAndRollback(t *testing.T) {
	config := manualConfig(t, 7)
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	view := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "unsafe", Kind: ui.Extension, Custom: struct{}{}}}}
	if err = a.Open(app.Window{ID: "one", Title: "Refused", Width: 400, Height: 240, View: func() ui.Node { return view }}); err != nil {
		t.Fatal(err)
	}
	if _, err = NewManual(testContext, a, config); !errors.Is(err, ErrManualUnsupported) {
		t.Fatal("unsupported extension admitted", err)
	}
	if a.Context().Err() != nil || len(a.Windows()) != 1 || a.Pending().DirtyWindows != 1 {
		t.Fatal("construction refusal changed domain registration/dirty readiness", a.Pending())
	}
	view = ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "button", Kind: ui.Button, Text: "Accepted"}}}
	m, err := NewManual(testContext, a, config)
	if err != nil {
		t.Fatal("rollback prevented valid remount", err)
	}
	defer m.Close(testContext)
	copied := &Manual{self: m.self, implementation: m.implementation}
	if _, err = copied.Snapshot(testContext); !errors.Is(err, ErrManualConfig) {
		t.Fatal("copied adapter accepted", err)
	}
	zero := &Manual{}
	if err = zero.Close(testContext); !errors.Is(err, ErrManualConfig) {
		t.Fatal("zero adapter accepted", err)
	}
	config.Fonts[0].Data[0] ^= 255 // Constructor retained detached valid font bytes.
	unitSettle(t, m)
	if _, err = m.Capture(testContext, "one"); err != nil {
		t.Fatal("font copy aliased caller mutation", err)
	}
	invalid := manualConfig(t, 7)
	invalid.Fonts[0].SHA256 = "wrong"
	a2, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a2.Close()
	if _, err = NewManual(testContext, a2, invalid); !errors.Is(err, ErrManualConfig) {
		t.Fatal("invalid font digest admitted", err)
	}
	invalid = manualConfig(t, 7)
	invalid.Fonts[0].Data = []byte("not a font")
	hash := sha256.Sum256(invalid.Fonts[0].Data)
	invalid.Fonts[0].SHA256 = hex.EncodeToString(hash[:])
	if _, err = NewManual(testContext, a2, invalid); !errors.Is(err, ErrManualConfig) {
		t.Fatal("malformed font admitted", err)
	}
}
func TestManualDescriptorControlLimitBeforeMount(t *testing.T) {
	config := manualConfig(t, 7)
	config.MaxControls = 1
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "one", Title: "Bounded", Width: 400, Height: 240, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "child", Kind: ui.Label}}}
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err = NewManual(testContext, a, config); !errors.Is(err, ErrManualConfig) {
		t.Fatal("control budget ignored", err)
	}
}
func TestManualFutureTimerActualOrigin(t *testing.T) {
	longs := 0
	config := manualConfig(t, 7)
	m, a := simpleManual(t, config, func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "button", Kind: ui.Button, Text: "Long press"}}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	r := w.state.root.children[0]
	r.widget.AsWidget().On(events.LongPressStart, func(events.Event) { longs++ })
	box := r.widget.AsWidget().Geom.TotalBBox
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: events.MouseDown, Button: events.Left, Position: box.Min.Add(box.Size().Div(2))}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	s, err := m.Snapshot(testContext)
	if err != nil || s.Clock.Future == 0 || s.Clock.Due != 0 || !s.Clock.CurrentIdle() || longs != 0 {
		t.Fatalf("future timer semantics %+v longs%d %v", s.Clock, longs, err)
	}
	if _, err = config.Clock.Advance(time.Second); err != nil {
		t.Fatal(err)
	}
	if longs != 0 {
		t.Fatal("timer ran widget outside origin")
	}
	unitSettle(t, m)
	if longs != 1 {
		t.Fatal("originating timer did not deliver exactly once", longs)
	}
	if !a.Pending().Idle() {
		t.Fatal("owner work remains", a.Pending())
	}
}
func TestManualFixedRegistrationAndOneWindowClose(t *testing.T) {
	config := manualConfig(t, 7)
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{"b": 0, "a": 0}
	for _, id := range []string{"b", "a"} {
		if err = a.Open(app.Window{ID: id, Title: id, Width: 400, Height: 220, View: func() ui.Node {
			return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "button", Kind: ui.Button, Text: id, Activate: func() { counts[id]++ }}}}
		}}); err != nil {
			t.Fatal(err)
		}
	}
	m, err := NewManual(testContext, a, config)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(testContext)
	unitSettle(t, m)
	s, err := m.Snapshot(testContext)
	if err != nil || s.Windows[0].ID != "b" || s.Windows[1].ID != "a" {
		t.Fatal("registration order not retained", s, err)
	}
	w := s.Windows[1]
	var e ManualElement
	for _, element := range w.Elements {
		if element.ID == "button" {
			e = element
		}
	}
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err = m.Input(testContext, ManualInput{Window: w.ID, WindowIncarnation: w.Incarnation, Type: typ, Button: events.Left, Position: e.Bounds.Min.Add(e.Bounds.Size().Div(2))}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if counts["a"] != 1 || counts["b"] != 0 {
		t.Fatal("window routing was shared", counts)
	}
	if err = a.CloseWindow(fixtureWindowRef(a, "a")); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	s, err = m.Snapshot(testContext)
	if err != nil || len(s.Windows) != 1 || s.Windows[0].ID != "b" || a.Context().Err() != nil {
		t.Fatalf("independent window retirement failed %+v %v", s, err)
	}
}

func TestManualOwnerArrivalCutoffAndFullOrdinaryQueue(t *testing.T) {
	config := manualConfig(t, 1)
	m, a := simpleManual(t, config, func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Queue"}}}
	})
	unitSettle(t, m)
	sequence := []string{}
	reserved, err := a.ReserveWindowCompletion(fixtureWindowRef(a, "one"), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{"A", "B", "C", "D"} {
		if err = a.PostWindow(fixtureWindowRef(a, "one"), func() { sequence = append(sequence, v) }); err != nil {
			t.Fatal(err)
		}
	}
	if err = a.PostWindow(fixtureWindowRef(a, "one"), func() {}); !errors.Is(err, app.ErrQueueFull) {
		t.Fatal("ordinary queue not bounded", err)
	}
	if err = reserved.Publish(func() { sequence = append(sequence, "R") }); err != nil {
		t.Fatal("completion blocked by full ordinary queue", err)
	}
	for range 128 {
		s, err := m.Snapshot(testContext)
		if err != nil {
			t.Fatal(err)
		}
		if s.Application.Ordinary+s.Application.ReadyCompletions == 0 {
			break
		}
		step, err := m.Step(testContext, 1)
		if err != nil {
			t.Fatal(err)
		}
		if step.Events+step.OwnerCallbacks > 1 {
			t.Fatal("originating event bypassed budget", step)
		}
	}
	if strings.Join(sequence, "") != "ABCDR" {
		t.Fatal("arrival order changed", sequence)
	}
	if !a.Pending().Idle() {
		t.Fatal("bounded owner work did not settle", a.Pending())
	}
}

func TestManualAbnormalConstructionRetiresAllWindows(t *testing.T) {
	for _, mode := range []string{"panic", "goexit"} {
		t.Run(mode, func(t *testing.T) {
			config := manualConfig(t, 7)
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			baseWindows := len(core.AllRenderWindows)
			var first system.Window
			if err = a.Open(app.Window{ID: "first", Title: "First", Width: 400, Height: 240, View: func() ui.Node { return ui.Node{ID: "first", Kind: ui.Label, Text: "first exists"} }}); err != nil {
				t.Fatal(err)
			}
			if err = a.Open(app.Window{ID: "second", Title: "Second", Width: 400, Height: 240, View: func() ui.Node {
				if len(core.AllRenderWindows) > baseWindows {
					first = core.AllRenderWindows[len(core.AllRenderWindows)-1].SystemWindow
				}
				if mode == "panic" {
					panic("second-view")
				}
				runtime.Goexit()
				return ui.Node{}
			}}); err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			var panicValue any
			go func() {
				defer close(done)
				defer func() { panicValue = recover() }()
				_, _ = NewManual(testContext, a, config)
			}()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("abnormal construction did not unwind")
			}
			if first == nil || !first.IsClosed() {
				t.Fatal("first real native window leaked", first)
			}
			if len(core.AllRenderWindows) != baseWindows {
				t.Fatal("real Core registry leaked", len(core.AllRenderWindows), baseWindows)
			}
			manualOwners.Lock()
			claimed := manualOwners.owners[a] != nil
			manualOwners.Unlock()
			if claimed {
				t.Fatal("manual owner claim leaked")
			}
			if a.Context().Err() == nil {
				t.Fatal("abnormal owner callback left domain alive")
			}
			if mode == "panic" && panicValue != "second-view" {
				t.Fatal("original normal Go panic was swallowed", panicValue)
			}
			if mode == "goexit" && panicValue != nil {
				t.Fatal("Goexit became a panic", panicValue)
			}
		})
	}
}
func TestManualClosedDomainRefusesImageAndEnqueueBeforeRetirement(t *testing.T) {
	config := manualConfig(t, 7)
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"one", "survivor"} {
		if err = a.Open(app.Window{ID: id, Title: id, Width: 400, Height: 240, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label, Text: id} }}); err != nil {
			t.Fatal(err)
		}
	}
	m, err := NewManual(testContext, a, config)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(testContext)
	unitSettle(t, m)
	if _, err = m.Capture(testContext, "one"); err != nil {
		t.Fatal(err)
	}
	if err = a.CloseWindow(fixtureWindowRef(a, "one")); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Capture(testContext, "one"); !errors.Is(err, ErrManualClosed) {
		t.Fatal("closed origin returned old image", err)
	}
	if err = m.RequestFrame(testContext, "one"); !errors.Is(err, ErrManualClosed) {
		t.Fatal("closed origin admitted paint", err)
	}
	s, err := m.Snapshot(testContext)
	if err != nil || len(s.Windows) != 2 || !s.Windows[0].DomainClosed || !s.Windows[0].RetirementPending {
		t.Fatalf("pending native retirement hidden %+v %v", s, err)
	}
	unitSettle(t, m)
	s, err = m.Snapshot(testContext)
	if err != nil || len(s.Windows) != 1 || s.Windows[0].ID != "survivor" {
		t.Fatalf("retired origin wasn't removed %+v %v", s, err)
	}
}

func TestManualRequestFrameOnlyEnqueuesAndStepChargesDispatch(t *testing.T) {
	changed := 0
	config := manualConfig(t, 7)
	m, a := simpleManual(t, config, func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "value", Kind: ui.Label, Text: fmt.Sprint(changed)}}}
	})
	unitSettle(t, m)
	if err := a.PostWindow(fixtureWindowRef(a, "one"), func() { changed++; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	before, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.RequestFrame(testContext, "one"); err != nil {
		t.Fatal(err)
	}
	after, err := m.Snapshot(testContext)
	if err != nil || after.Windows[0].Frames != before.Windows[0].Frames || !after.Windows[0].ReadyPaint || after.Clock.Now != before.Clock.Now || after.Application != before.Application || changed != 0 {
		t.Fatalf("frame request dispatched hidden work before%+v after%+v value%d error%v", before, after, changed, err)
	}
	r, err := m.Step(testContext, 1)
	if err != nil || r.Events+r.OwnerCallbacks != 1 || r.OwnerCallbacks != 0 || r.Frames != 1 || changed != 0 {
		t.Fatalf("frame dispatch bypassed budget %+v value%d %v", r, changed, err)
	}
	if _, err = m.Capture(testContext, "one"); !errors.Is(err, ErrManualLayout) {
		t.Fatal("capture admitted pending owner/native work", err)
	}
	unitSettle(t, m)
	if changed != 1 {
		t.Fatal("explicit steps did not deliver owner work once", changed)
	}
	if _, err = m.Capture(testContext, "one"); err != nil {
		t.Fatal(err)
	}
}

func TestManualOneApplicationPerProcessAndReservationRelease(t *testing.T) {
	config := manualConfig(t, 7)
	view := func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label, Text: "Exclusive application"} }
	m, a := simpleManual(t, config, view)
	second, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	called := 0
	if err = second.Open(app.Window{ID: "second", Title: "Other app", Width: 400, Height: 240, View: func() ui.Node { called++; return view() }}); err != nil {
		t.Fatal(err)
	}
	nativeCount := len(core.AllRenderWindows)
	if _, err = NewManual(testContext, second, config); !errors.Is(err, ErrManualBusy) {
		t.Fatal("another managed app admitted", err)
	}
	if called != 0 || len(core.AllRenderWindows) != nativeCount || second.Context().Err() != nil {
		t.Fatal("process refusal constructed widgets or changed domain", called)
	}
	if err = m.Close(testContext); err != nil {
		t.Fatal(err)
	}
	if a.Context().Err() == nil {
		t.Fatal("application close was not terminal")
	}
	other, err := NewManual(testContext, second, config)
	if err != nil {
		t.Fatal("completed Close retained process claim", err)
	}
	if err = other.Close(testContext); err != nil {
		t.Fatal(err)
	}
}

func TestManualUnqualifiedManagerKeysRefuseBeforeRefreshOrAdmission(t *testing.T) {
	views := 0
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		views++
		return ui.Node{ID: "root", Kind: ui.Label, Text: "Qualified fixed scene"}
	})
	unitSettle(t, m)
	if err := a.Invalidate("one"); err != nil {
		t.Fatal(err)
	}
	before, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	viewCount := views
	for _, function := range []keymap.Functions{keymap.Copy, keymap.Cut, keymap.Paste, keymap.PasteHist, keymap.WinSnapshot, keymap.ZoomIn, keymap.ZoomOut, keymap.Refresh, keymap.Find, keymap.Menu, keymap.WinFocusNext} {
		chords := function.Chord()
		if chords == "" {
			t.Fatalf("test requires a mapped chord for %s", function)
		}
		for _, chord := range strings.Split(string(chords), "\n") {
			r, code, modifiers, err := key.Chord(chord).Decode()
			if err != nil || keymap.Of(key.NewChord(r, code, modifiers)) != function {
				t.Fatalf("chord reconstruction does not match actual routing: %s %v", chord, err)
			}
			for _, event := range []events.Types{events.KeyChord, events.KeyDown} {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Windows[0].Incarnation, Type: event, Rune: r, Code: code, Modifiers: modifiers}); !errors.Is(err, ErrManualUnsupported) {
					t.Fatalf("unqualified %s through %s admitted: %v", function, event, err)
				}
			}
		}
	}
	after, err := m.Snapshot(testContext)
	if err != nil || views != viewCount || !reflect.DeepEqual(before, after) {
		t.Fatalf("refused input changed view/owner/queue/frame/timer state: views %d/%d before %+v after %+v error %v", viewCount, views, before, after, err)
	}
	for _, function := range []keymap.Functions{keymap.Undo, keymap.Redo, keymap.FocusNext, keymap.FocusPrev, keymap.Enter, keymap.SelectAll, keymap.MoveLeft} {
		r, code, modifiers, err := key.Chord(strings.Split(string(function.Chord()), "\n")[0]).Decode()
		if err != nil || manualUnqualifiedInput(ManualInput{Type: events.KeyChord, Rune: r, Code: code, Modifiers: modifiers}) {
			t.Fatalf("qualified command/navigation %s refused: %v", function, err)
		}
	}
}

func TestManualAccessibilityIndexUsesQualifiedIdentity(t *testing.T) {
	firstSelection, laterSelection := [2]int{1, 4}, [2]int{0, 0}
	nodes := []accessibilityNode{
		{ID: "field", DomainID: "field", Value: "first", Selected: true, TextSelection: &firstSelection, TextFocus: 1},
		{ID: "check", DomainID: "check", Value: true},
		{ID: "field", DomainID: "field", Value: "later", TextSelection: &laterSelection, TextFocus: 0},
		{ID: accessibilityOptionID("group", "check"), DomainID: "check", Value: false},
	}
	index := manualAccessibilityIndex(nodes)
	if len(index) != 3 || !reflect.DeepEqual(index["field"], nodes[0]) || !reflect.DeepEqual(index["check"], nodes[1]) {
		t.Fatalf("index aliased domain identities or changed duplicate-ID precedence: %+v", index)
	}
	if _, exists := index["missing"]; exists {
		t.Fatal("index invented a visible node")
	}
}

func TestManualCaptureKeepsLiveFieldTextAndDirectionalSelectionCoherent(t *testing.T) {
	changes := 0
	m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "field", Sequence: 1}, Children: []ui.Node{
			{ID: "label", Kind: ui.Label, Text: "Status"},
			{ID: "check", Kind: ui.Checkbox, Checked: true, Disabled: true},
			{ID: "field", Kind: ui.TextField, Text: "controlled", Change: func(ui.ChangeEvent) { changes++ }},
		}}
	})
	unitSettle(t, m)
	w := m.implementation.find("one")
	var fieldRecord *record
	var expected core.TextFieldEditingState
	if err := a.Do(func() {
		fieldRecord = findRecord(w.state.root, "field")
		field := fieldRecord.widget.(*core.TextField)
		stale := field.EditingState()
		fieldRecord.control.fieldState = &stale
		if !field.ApplyNativeEdit("A😀éאב", 4, 1, -1, -1, "insert") {
			t.Fatal("fixture editing state refused")
		}
		expected = field.EditingState()
		// Preserve an explicitly stale AX cache: capture must use the live
		// field value/selection after indexed semantic lookup, not this state.
		fieldRecord.control.fieldState = &stale
		nodes, _ := accessibilitySnapshot(w.state.root, nil)
		if manualAccessibilityIndex(nodes)[accessibilityManagedID("field")].Value != "controlled" {
			t.Fatal("fixture did not establish a stale published cache")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := m.implementation.captureWindow(testContext, w); err != nil {
		t.Fatal(err)
	}
	snapshot, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	elements := map[string]ManualElement{}
	for _, element := range snapshot.Windows[0].Elements {
		elements[element.ID] = element
	}
	field := elements["field"]
	if field.Value != expected.Text || field.Selection != [2]int{1, 4} || field.TextFocus != 1 || !field.Visible || !field.Focused || field.Parent != "root" || field.Incarnation != w.records[fieldRecord] {
		t.Fatalf("field capture mixed live/cached state or lost mapping: %+v expected %+v", field, expected)
	}
	if elements["label"].Value != "Status" || elements["check"].Value != true || elements["check"].Enabled || !elements["check"].Visible {
		t.Fatalf("indexed mapping changed nonfield semantics: %+v", elements)
	}
	if err := a.Do(func() {
		if changes != 0 || fieldRecord.control.fieldState == nil || *fieldRecord.control.fieldState != expected {
			t.Fatal("capture delivered a Change or did not refresh the same live state")
		}
	}); err != nil {
		t.Fatal(err)
	}
}
