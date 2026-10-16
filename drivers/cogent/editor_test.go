package cogent

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/text/textcore"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent/internal/mailbox"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

func editorFixture(t *testing.T, text string) (*app.Application, *nativeWindow, *textWindow, *record, *editor.Model) {
	t.Helper()
	a, _ := app.New(app.Config{})
	_ = a.Open(app.Window{ID: "editor-unit", Title: "Editor unit", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }})
	body := core.NewBody("Editor unit")
	m, err := editor.New(text, editor.Config{})
	if err != nil {
		t.Fatal(err)
	}
	r, err := reconcile(body, nil, ui.Node{ID: "document", Kind: ui.MultilineText, Editor: m}, a, "editor-unit")
	if err != nil {
		t.Fatal(err)
	}
	w := &nativeWindow{body: body, window: fixtureWindow(a, "editor-unit"), root: r}
	var source events.Source
	w.source.Store(&source)
	tw, err := newTextWindow(a, &source.Deque)
	if err != nil {
		t.Fatal(err)
	}
	tw.window = w
	lifetime, _ := a.WindowContext(w.window.ID)
	tw.lifetime.Store(&textLifetime{ctx: lifetime})
	w.input.Store(tw)
	textWindows.Lock()
	textWindows.windows[w.window.ID] = tw
	textWindows.Unlock()
	if err := bindTextWindow(w); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeTextWindow(w); disposeControlTree(w.root); a.Close() })
	fixtureShadow(tw, r, m)
	return a, w, tw, r, m
}
func fixtureShadow(tw *textWindow, r *record, m *editor.Model) {
	s := m.Snapshot()
	tw.shadows[r.control.client] = textShadow{Text: s.VisibleText, Revision: s.Revision, Epoch: s.VisibleEpoch, SelectionEpoch: s.SelectionEpoch}
}
func fixtureInput(r *record, seq uint64, kind, before, next string, a, f, ms, me int) textinput.Update {
	return textinput.Update{Sequence: seq, ClientID: r.control.client, Kind: kind, BeforeText: before, Text: next, Anchor: a, Focus: f, MarkedStart: ms, MarkedEnd: me}
}
func TestManagedEditorCompositionHistoryExactBytes(t *testing.T) {
	a, _, tw, r, m := editorFixture(t, "A\r\n😀\n\n")
	original := m.Snapshot().Text
	_ = a.Do(func() { tw.apply(fixtureInput(r, 1, "mark", original, "Aに\r\n😀\n\n", 2, 2, 1, 2)) })
	if !m.Snapshot().Composing || m.Snapshot().Text != original || m.Snapshot().History.UndoCount != 0 {
		t.Fatal("preview changed committed bytes/history", m.Snapshot())
	}
	_ = a.Do(func() {
		tw.apply(fixtureInput(r, 2, "mark", m.Snapshot().VisibleText, "A日本\r\n😀\n\n", 3, 3, 1, 3))
		tw.apply(fixtureInput(r, 3, "unmark", m.Snapshot().VisibleText, m.Snapshot().VisibleText, 3, 3, -1, -1))
	})
	if m.Snapshot().Composing || m.Snapshot().History.UndoCount != 1 || m.Snapshot().Text != "A日本\r\n😀\n\n" {
		t.Fatal("zero-delta commit lost atomic transaction", m.Snapshot())
	}
	_ = a.Do(func() {
		if err := m.Journal().Undo(); err != nil {
			t.Fatal(err)
		}
		syncEditor(r)
	})
	if m.Snapshot().Text != original || r.widget.(*textcore.Editor).EditingState().Text != original {
		t.Fatal("undo lost exact CRLF/final blanks")
	}
	_ = a.Do(func() {
		if err := m.Journal().Redo(); err != nil {
			t.Fatal(err)
		}
		syncEditor(r)
		fixtureShadow(tw, r, m)
		tw.apply(fixtureInput(r, 4, "mark", m.Snapshot().VisibleText, "A取消\r\n😀\n\n", 3, 3, 1, 3))
		tw.apply(fixtureInput(r, 5, "cancel", m.Snapshot().VisibleText, "A日本\r\n😀\n\n", 3, 3, -1, -1))
	})
	if m.Snapshot().Text != "A日本\r\n😀\n\n" || m.Snapshot().Composing || m.Snapshot().History.UndoCount != 1 {
		t.Fatal("cancel changed history", m.Snapshot())
	}
}
func TestManagedEditorStaleRevisionAndRemount(t *testing.T) {
	a, w, tw, r, m := editorFixture(t, "base")
	old := fixtureInput(r, 1, "insert", "base", "base!", 5, 5, -1, -1)
	_ = a.Do(func() {
		if err := m.Journal().Replace(true, func() error { return m.Install("base", editor.Selection{}) }); err != nil {
			t.Fatal(err)
		}
		syncEditor(r)
		tw.apply(old)
	})
	if m.Snapshot().Text != "base" || tw.ack != 1 {
		t.Fatal("same-byte Install upgraded queued input", m.Snapshot())
	}
	fixtureShadow(tw, r, m)
	old = fixtureInput(r, 2, "insert", "base", "old", 3, 3, -1, -1)
	var next *record
	_ = a.Do(func() {
		var err error
		next, err = reconcile(w.body, r, ui.Node{ID: "replacement", Kind: ui.MultilineText, Editor: m}, a, "editor-unit")
		if err != nil {
			t.Fatal(err)
		}
		w.root = next
		if err := bindTextWindow(w); err != nil {
			t.Fatal(err)
		}
		tw.apply(old)
	})
	if next.control.client == r.control.client || m.Snapshot().Text != "base" {
		t.Fatal("old incarnation reached replacement")
	}
}
func TestManagedEditorLegacyBoundaryPrecedesUndoAndNative(t *testing.T) {
	a, w, tw, r, m := editorFixture(t, "")
	ed := r.widget.(*textcore.Editor)
	source := w.source.Load()
	source.KeyChord('x', key.CodeX, 0)
	source.KeyChord('z', key.CodeZ, 1<<key.Meta)
	for {
		e, ok := source.Deque.PollEvent()
		if !ok {
			break
		}
		if !handleTextEvent(w, e) {
			ed.HandleEvent(e)
			w.body.Scene.EventBoundary(e)
		}
	}
	if m.Snapshot().Text != "" || m.Snapshot().History.RedoCount != 1 {
		t.Fatal("queued legacy key crossed Undo", m.Snapshot())
	}
	if err := a.Do(func() { _ = m.Journal().Redo(); syncEditor(r); fixtureShadow(tw, r, m) }); err != nil {
		t.Fatal(err)
	}
	// The native record captured before a queued legacy key must be refused after
	// that key's boundary publishes its own transaction, never erase the key.
	source.KeyChord('y', key.CodeY, 0)
	if err := tw.pump.Send(textMessage{Kind: "input", Input: fixtureInput(r, 1, "insert", "x", "x!", 2, 2, -1, -1)}); err != nil {
		t.Fatal(err)
	}
	for {
		e, ok := w.source.Load().Deque.PollEvent()
		if !ok {
			break
		}
		if !handleTextEvent(w, e) {
			ed.HandleEvent(e)
			w.body.Scene.EventBoundary(e)
		}
	}
	if m.Snapshot().Text != "xy" || m.Snapshot().History.UndoCount != 2 {
		t.Fatal("native shadow erased prior legacy edit", m.Snapshot())
	}
}
func TestManagedEditorLargePasteAndCut(t *testing.T) {
	a, _, tw, r, m := editorFixture(t, "")
	next := strings.Repeat("😀\r\n", 1100)
	n := len([]rune(next))
	u, _ := textinput.UTF16Offset(next, n)
	_ = a.Do(func() {
		tw.apply(fixtureInput(r, 1, "insert", "", next, u, u, -1, -1))
		tw.apply(fixtureInput(r, 2, "insert", next, "", 0, 0, -1, -1))
	})
	if m.Snapshot().Text != "" || m.Snapshot().History.UndoCount != 2 {
		t.Fatal("large paste/cut refused", m.Snapshot())
	}
	_ = a.Do(func() {
		if err := m.Journal().Undo(); err != nil {
			t.Fatal(err)
		}
		syncEditor(r)
	})
	if m.Snapshot().Text != next {
		t.Fatal("large Cut undo lost exact bytes")
	}
}
func TestManagedEditorBindingRollbackAndBoundaryRestore(t *testing.T) {
	a, w, tw, r, m := editorFixture(t, "a")
	if err := a.Do(func() {
		if err := bindEditorModels(&record{children: []*record{r, {control: &controlState{editor: &editorBinding{model: m}}}}}); !errors.Is(err, ErrEditorBinding) {
			t.Fatal("multiple live binding admitted", err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if editorBindings[m] != r {
		t.Fatal("failed binding poisoned registry")
	}
	// A previously installed callback is retained and restored by this adapter.
	priorCalls := 0
	scene := core.NewBody("extra").Scene
	prior := func(events.Event) { priorCalls++ }
	scene.EventBoundary = prior
	tw.bindBoundary(scene)
	scene.EventBoundary(events.NewKey(events.KeyChord, 'q', key.CodeQ, 0))
	closeTextWindow(w)
	scene.EventBoundary(events.NewKey(events.KeyChord, 'q', key.CodeQ, 0))
	if priorCalls != 2 {
		t.Fatal("previous event boundary lost", priorCalls)
	}
}
func TestManagedEditorSettlementBoundCancellationRetirementAndAppClose(t *testing.T) {
	a, w, tw, r, m := editorFixture(t, "base")
	ctx, cancel := context.WithCancel(context.Background())
	callback := make(chan error, 2)
	// Reserve the request before any worker/Cocoa operation. An injected pending
	// request exercises public admission without needing an OS window in a unit test.
	tw.requests[1] = &textSettlement{binding: r, model: m, ctx: ctx, callback: func(err error) { callback <- err }}
	if err := a.Do(func() {
		if err := SettleEditor(ctx, a, w.window.ID, m, func(error) {}); !errors.Is(err, mailbox.ErrFull) {
			t.Fatal("unbounded settle admitted", err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	cancel()
	_ = a.Do(func() { tw.settled(1) })
	if err := <-callback; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	tw.requests[2] = &textSettlement{binding: r, model: m, ctx: context.Background(), callback: func(err error) { callback <- err }}
	r.control.retired = true
	_ = a.Do(func() { tw.settled(2) })
	if err := <-callback; !errors.Is(err, ErrTextInputStale) {
		t.Fatal(err)
	}
	r.control.retired = false
	tw.requests[3] = &textSettlement{binding: r, model: m, ctx: context.Background(), callback: func(err error) { callback <- err }}
	a.Close()
	closeTextWindow(w)
	select {
	case err := <-callback:
		t.Fatal("teardown called user after App.Close", err)
	case <-time.After(10 * time.Millisecond):
	}
}

func TestManagedEditorExternalGraphemeSelectionMapping(t *testing.T) {
	a, _, _, r, m := editorFixture(t, "Aé👩‍💻Z")
	for _, test := range []struct{ request, want editor.Selection }{{editor.Selection{Anchor: 2, Focus: 2}, editor.Selection{Anchor: 3, Focus: 3}}, {editor.Selection{Anchor: 5, Focus: 4}, editor.Selection{Anchor: 6, Focus: 3}}} {
		_ = a.Do(func() {
			if err := m.Select(test.request); err != nil {
				t.Fatal(err)
			}
			syncEditor(r)
		})
		if m.Snapshot().Selection != test.want {
			t.Fatal("external scalar selection not normalized once", m.Snapshot().Selection, test.want)
		}
		state := r.widget.(*textcore.Editor).EditingState()
		if state.Anchor != test.want.Anchor || state.Focus != test.want.Focus {
			t.Fatal("native/model selection disagrees", state)
		}
		before := m.Snapshot().SelectionEpoch
		_ = a.Do(func() { syncEditor(r) })
		if m.Snapshot().SelectionEpoch != before {
			t.Fatal("unrelated sync moved selection again")
		}
	}
}

func TestManagedEditorSettlementCapturesPriorNativeFIFO(t *testing.T) {
	a, w, tw, r, m := editorFixture(t, "A")
	if err := tw.pump.Send(textMessage{Kind: "input", Input: fixtureInput(r, 1, "mark", "A", "Aに", 2, 2, 1, 2)}); err != nil {
		t.Fatal(err)
	}
	ready := make(chan struct{})
	tw.settleDispatch = func(ctx context.Context, tw *textWindow, client, id uint64) error {
		if err := tw.pump.Send(textMessage{Kind: "input", Input: fixtureInput(r, 2, "unmark", "Aに", "Aに", 2, 2, -1, -1)}); err != nil {
			return err
		}
		err := tw.pump.Send(textMessage{Kind: "settle", Request: id})
		close(ready)
		return err
	}
	called := false
	_ = a.Do(func() {
		if err := SettleEditor(context.Background(), a, w.window.ID, m, func(err error) {
			if err != nil {
				t.Fatal(err)
			}
			called = true
			if m.Snapshot().Text != "Aに" || m.Snapshot().Composing || m.Snapshot().History.UndoCount != 1 {
				t.Fatal("save capture bypassed originating native FIFO", m.Snapshot())
			}
		}); err != nil {
			t.Fatal(err)
		}
	})
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("settle dispatch stuck")
	}
	for {
		e, ok := w.source.Load().Deque.PollEvent()
		if !ok {
			break
		}
		handleTextEvent(w, e)
	}
	if !called || len(tw.requests) != 0 {
		t.Fatal("settle did not complete owner callback")
	}
}

func TestManagedEditorNativeAdmissionStopsAtLogicalWindowClose(t *testing.T) {
	a, w, tw, _, _ := editorFixture(t, "A")
	if !tw.acceptsNativeInput() {
		t.Fatal("live client rejected")
	}
	// Keep another window live so this exercises logical window cancellation,
	// before Cocoa attachment destruction or whole-application cancellation.
	_ = a.Open(app.Window{ID: "other", Title: "Other", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "other", Kind: ui.Column} }})
	_ = a.Do(func() { _ = a.CloseWindow(w.window.Ref()) })
	if tw.closed.Load() || a.Context().Err() != nil || tw.acceptsNativeInput() {
		t.Fatal("logical close still admitted native mutation")
	}
}

func TestManagedEditorSameVisibleInstallRetiresPrivatePreedit(t *testing.T) {
	a, _, tw, r, m := editorFixture(t, "A")
	_ = a.Do(func() { tw.apply(fixtureInput(r, 1, "mark", "A", "Aに", 2, 2, 1, 2)) })
	_ = a.Do(func() {
		if err := m.CommitComposition(); err != nil {
			t.Fatal(err)
		}
		visible := m.Snapshot().VisibleText
		selection := m.Snapshot().Selection
		if err := m.Journal().Replace(true, func() error { return m.Install(visible, selection) }); err != nil {
			t.Fatal(err)
		}
		syncEditor(r)
	})
	if _, _, active := r.widget.(*textcore.Editor).NativeCompositionRange(); active {
		t.Fatal("same-visible Install retained private restoration baseline")
	}
	// A queued cancellation from the retired shadow cannot erase owner bytes.
	_ = a.Do(func() { tw.apply(fixtureInput(r, 2, "cancel", "Aに", "A", 0, 0, -1, -1)) })
	if m.Snapshot().Text != "Aに" || r.widget.(*textcore.Editor).EditingState().Text != "Aに" {
		t.Fatal("stale preedit cancellation erased installed document")
	}
}

func TestManagedEditorLogicalCloseRefusesAlreadyAdmittedInput(t *testing.T) {
	a, w, tw, r, m := editorFixture(t, "A")
	changes := 0
	r.node.Change = func(ui.ChangeEvent) { changes++ }
	_ = a.Open(app.Window{ID: "peer", Title: "Peer", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "peer", Kind: ui.Column} }})
	if err := tw.pump.Send(textMessage{Kind: "input", Input: fixtureInput(r, 1, "insert", "A", "closed", 6, 6, -1, -1)}); err != nil {
		t.Fatal(err)
	}
	_ = a.Do(func() { _ = a.CloseWindow(w.window.Ref()) })
	for {
		e, ok := w.source.Load().Deque.PollEvent()
		if !ok {
			break
		}
		handleTextEvent(w, e)
	}
	if m.Snapshot().Text != "A" || m.Snapshot().History.UndoCount != 0 || changes != 0 {
		t.Fatal("logical close delivered queued input", m.Snapshot(), changes)
	}
}

func TestManagedEditorShutdownWaitsForAcceptedNativeCommit(t *testing.T) {
	_, w, tw, r, _ := editorFixture(t, "A")
	if !tw.preflightInput(fixtureInput(r, 1, "insert", "A", "A!", 2, 2, -1, -1)) {
		t.Fatal("live preflight refused")
	}
	done := make(chan struct{})
	go func() { closeTextWindow(w); close(done) }()
	deadline := time.Now().Add(time.Second)
	for !tw.closed.Load() {
		if time.Now().After(deadline) {
			t.Fatal("teardown did not close admission")
		}
		time.Sleep(time.Millisecond)
	}
	select {
	case <-done:
		t.Fatal("teardown canceled native ticket before accepted mutation committed")
	default:
	}
	if err := tw.commitInput(); err != nil {
		t.Fatal("accepted native ticket invalidated", err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("teardown did not finish after commit")
	}
	if tw.pump.Stats().Queue.Outstanding != 0 || tw.pendingInput != nil {
		t.Fatal("accepted teardown leaked credit")
	}
	if tw.preflightInput(fixtureInput(r, 2, "insert", "A!", "bad", 3, 3, -1, -1)) {
		t.Fatal("retired native mutation admitted")
	}
}
func TestManagedEditorModalBoundaryRegistryRetiresPriorScenes(t *testing.T) {
	_, w, tw, r, _ := editorFixture(t, "A")
	calls := 0
	for i := 0; i < 30; i++ {
		scene := core.NewBody("modal cycle").Scene
		scene.EventBoundary = func(events.Event) { calls++ }
		r.widget.AsWidget().SetScene(scene)
		if err := bindTextWindow(w); err != nil {
			t.Fatal(err)
		}
		if len(tw.boundaries) != 2 {
			t.Fatal("retired modal Scene retained", len(tw.boundaries))
		}
		r.widget.AsWidget().SetScene(w.body.Scene)
		if err := bindTextWindow(w); err != nil {
			t.Fatal(err)
		}
		if len(tw.boundaries) != 1 {
			t.Fatal("closed modal Scene retained", len(tw.boundaries))
		}
		scene.EventBoundary(events.NewKey(events.KeyChord, 'q', key.CodeQ, 0))
	}
	if calls != 30 {
		t.Fatal("retired modal prior hooks not restored", calls)
	}
}

func TestManagedEditorNativeBudgetRefusedBeforeAllocation(t *testing.T) {
	for _, initial := range []string{"", strings.Repeat("x", textinput.MaxDocumentBytes+1)} {
		m, err := editor.New(initial, editor.Config{MaxBytes: 8 << 20})
		if err != nil {
			t.Fatal(err)
		}
		n := ui.Node{ID: "too-large", Kind: ui.MultilineText, Editor: m}
		if err := ui.Validate(n); err != nil {
			t.Fatalf("portable model must remain valid: %v", err)
		}
		body := core.NewBody("budget")
		count := body.NumChildren()
		if _, err := reconcile(body, nil, ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{n}}, nil, ""); err == nil {
			t.Fatal("oversized configuration accepted")
		}
		if body.NumChildren() != count {
			t.Fatal("failed preflight allocated widgets")
		}
		r := &record{node: n}
		if _, _, err := mountControl(body, r, nil, ""); err == nil {
			t.Fatal("direct mount accepted oversized configuration")
		}
		if r.control != nil || body.NumChildren() != count {
			t.Fatal("failed mount allocated private buffer")
		}
	}
}

func TestManagedEditorInitialPublicationSurvivesApplicationWake(t *testing.T) {
	a, w, tw, _, _ := editorFixture(t, "")
	source := w.source.Load()
	source.Custom("vellum.wake")
	var ordinary int
	for {
		e, ok := source.Deque.PollEvent()
		if !ok {
			break
		}
		if !handleTextEvent(w, e) {
			ordinary++
		}
	}
	if ordinary != 1 || tw.pump.Stats().LatestBytes != 0 {
		t.Fatalf("startup publication was replaced by application wake: ordinary=%d stats=%+v", ordinary, tw.pump.Stats())
	}
	// Future publication must reserve a fresh wake, not coalesce into a node
	// that an ordinary custom compressed out of the queue.
	_ = a.Do(func() { requestTextPublication(w.window.ID) })
	e, ok := source.Deque.PollEvent()
	if !ok || !handleTextEvent(w, e) || tw.pump.Stats().LatestBytes != 0 {
		t.Fatal("future publication permanently stranded")
	}
}

func TestManagedTextFieldNativeCompositionControlledChangeAndStaleGuard(t *testing.T) {
	a, w, tw, _, m := editorFixture(t, "document")
	var field *record
	calls := 0
	fieldNode := ui.Node{ID: "field", Kind: ui.TextField, Text: "A", Change: func(e ui.ChangeEvent) { calls++; field.node.Text = e.Text }}
	var err error
	field, err = reconcile(w.body, nil, fieldNode, a, w.window.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { disposeControlTree(field) })
	tw.clients[field.control.client] = field
	tw.shadows[field.control.client] = textShadow{Text: "A", Epoch: field.control.textEpoch}
	_ = a.Do(func() {
		tw.apply(fixtureInput(field, 1, "mark", "A", "Aに", 2, 2, 1, 2))
		tw.apply(fixtureInput(field, 2, "unmark", "Aに", "Aに", 2, 2, -1, -1))
	})
	if calls != 0 || field.node.Text != "A" {
		t.Fatal("native field preedit bypassed controlled Change")
	}
	field.widget.(*core.TextField).Send(events.FocusLost)
	a.Drain()
	if calls != 1 || field.node.Text != "Aに" {
		t.Fatalf("focus handoff lost controlled Change: calls=%d text=%q", calls, field.node.Text)
	}
	oldNode := fieldNode
	_ = a.Do(func() {
		if err := syncControl(field, oldNode); err != nil {
			t.Fatal(err)
		}
	})
	// An owner replacement seals the prior shadow even if the client stays mounted.
	_ = a.Do(func() {
		previous := field.node
		field.node.Text = "owner"
		if err := syncControl(field, previous); err != nil {
			t.Fatal(err)
		}
		tw.apply(fixtureInput(field, 3, "insert", "Aに", "stale", 5, 5, -1, -1))
	})
	if field.widget.(*core.TextField).EditingState().Text != "owner" || calls != 1 || m.Snapshot().History.UndoCount != 0 {
		t.Fatal("stale field client changed controlled value or peer document")
	}
}

func TestManagedEditorNativeMainReadinessDoesNotBlockOwnerOrActivateRetiredWindow(t *testing.T) {
	for _, retire := range []bool{false, true} {
		a, w, tw, _, _ := editorFixture(t, "")
		queued := make(chan func(), 1)
		returned := make(chan error, 1)
		go func() { returned <- a.Do(func() { tw.awaitNativeTextMain(func(f func()) { queued <- f }) }) }()
		select {
		case err := <-returned:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("main readiness waited under application owner")
		}
		if tw.nativeReady.Load() {
			t.Fatal("pre-loop callback activated native publication")
		}
		f := <-queued
		if retire {
			closeTextWindow(w)
		}
		f()
		if tw.nativeReady.Load() == retire {
			t.Fatalf("readiness after retired=%t was %t", retire, tw.nativeReady.Load())
		}
		if !retire {
			e, ok := w.source.Load().Deque.PollEvent()
			if !ok || !handleTextEvent(w, e) || tw.pump.Stats().LatestBytes != 0 {
				t.Fatal("main readiness did not queue publication retry")
			}
		}
	}
}
func TestManagedTextFieldNativeBudgetRefusedBeforeAllocation(t *testing.T) {
	n := ui.Node{ID: "too-large-field", Kind: ui.TextField, Text: strings.Repeat("x", textinput.MaxDocumentBytes+1)}
	if err := ui.Validate(n); err != nil {
		t.Fatal("portable field must remain valid", err)
	}
	body := core.NewBody("field budget")
	count := body.NumChildren()
	if _, err := reconcile(body, nil, n, nil, ""); err == nil {
		t.Fatal("oversized native field accepted")
	}
	r := &record{node: n}
	if _, _, err := mountControl(body, r, nil, ""); err == nil || r.control != nil || body.NumChildren() != count {
		t.Fatal("field refusal allocated a widget")
	}
}

func TestManagedTextFieldOversizedLegacyEditingRestoresControlledValueWithoutChange(t *testing.T) {
	a, w, _, _, m := editorFixture(t, "peer")
	calls := 0
	field, err := reconcile(w.body, nil, ui.Node{ID: "legacy-field", Kind: ui.TextField, Text: "controlled", Change: func(ui.ChangeEvent) { calls++ }}, a, w.window.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { disposeControlTree(field) })
	native := field.widget.(*core.TextField)
	if !native.ApplyNativeEdit("local", 5, 5, -1, -1, "insert") {
		t.Fatal("local fixture edit")
	}
	accepted := native.EditingState()
	field.control.fieldState = &accepted
	if !native.ApplyNativeEdit(strings.Repeat("x", textinput.MaxDocumentBytes+1), 0, 0, -1, -1, "insert") {
		t.Fatal("oversized legacy fixture edit")
	}
	_ = a.Do(func() {
		if !refuseOversizedFieldEditing(field) {
			t.Fatal("overcap legacy edit accepted")
		}
	})
	if native.EditingState().Text != "controlled" || field.control.fieldState == nil || *field.control.fieldState != native.EditingState() || calls != 0 || m.Snapshot().History.UndoCount != 0 || field.control.restore {
		t.Fatal("refused edit failed controlled state/cache restoration or changed callback/model")
	}
	_ = a.Do(func() { requireFieldSemanticState(t, field, native.EditingState()) })
	// A field is still controlled: after refusing all pending edit bytes, the
	// next valid edit can commit by focus loss using the normal Change path.
	if !native.ApplyNativeEdit("valid", 5, 5, -1, -1, "insert") {
		t.Fatal("valid editing did not resume")
	}
	native.Send(events.FocusLost)
	a.Drain()
	if calls != 1 {
		t.Fatal("valid field Change unusable after refusal", calls)
	}
	// A Change already queued before an originating budget boundary must refuse
	// oversized bytes on owner delivery as well; Drain may run in another window.
	if !native.ApplyNativeEdit(strings.Repeat("x", textinput.MaxDocumentBytes+1), 0, 0, -1, -1, "insert") {
		t.Fatal("oversized delivery fixture")
	}
	native.Send(events.FocusLost)
	a.Drain()
	if calls != 1 {
		t.Fatal("queued oversized field Change escaped refusal", calls)
	}
	_ = a.Do(func() {
		if err := syncControl(field, field.node); err != nil {
			t.Fatal(err)
		}
	})
	if native.EditingState().Text != "controlled" {
		t.Fatal("queued refusal did not restore descriptor")
	}
}

func TestManagedTextFieldRefreshPreservesPreeditAndQueuedDeltaShadow(t *testing.T) {
	a, w, tw, _, _ := editorFixture(t, "")
	field, err := reconcile(w.body, nil, ui.Node{ID: "refresh-field", Kind: ui.TextField, Text: "A"}, a, w.window.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { disposeControlTree(field) })
	tw.clients[field.control.client] = field
	tw.shadows[field.control.client] = textShadow{Text: "A", Epoch: field.control.textEpoch}
	_ = a.Do(func() {
		tw.apply(fixtureInput(field, 1, "mark", "A", "Aに", 2, 2, 1, 2))
		native := field.widget.(*core.TextField)
		native.Update()
		if native.EditingState().Text != "Aに" {
			t.Fatal("retained refresh erased private preedit")
		}
		if _, _, marked := native.NativeCompositionRange(); !marked {
			t.Fatal("refresh erased marked baseline")
		}
		u := textinput.Update{Sequence: 2, ClientID: field.control.client, Kind: "mark", Delta: true, ReplaceStart: 1, ReplaceEnd: 2, BeforeText: "に", Text: "日本", Anchor: 3, Focus: 3, MarkedStart: 1, MarkedEnd: 3}
		tw.apply(u)
		if native.EditingState().Text != "A日本" || tw.shadows[field.control.client].Text != "A日本" {
			t.Fatal("queued delta lost refreshed before-whole-shadow")
		}
		native.Update()
		tw.apply(fixtureInput(field, 3, "cancel", "A日本", "A", 0, 0, -1, -1))
		if native.EditingState().Text != "A" {
			t.Fatal("refresh replaced composition cancellation baseline")
		}
	})
}

func focusLossField(t *testing.T, a *app.Application, w *nativeWindow, tw *textWindow, id string, changes *[]string) *record {
	t.Helper()
	r, err := reconcile(w.body, nil, ui.Node{ID: id, Kind: ui.TextField, Text: "A", Change: func(e ui.ChangeEvent) { *changes = append(*changes, e.Text) }}, a, w.window.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { disposeControlTree(r) })
	tw.clients[r.control.client] = r
	tw.shadows[r.control.client] = textShadow{Text: "A", Epoch: r.control.textEpoch}
	return r
}

func TestManagedTextFieldFocusLossWaitsForAdmittedNativeFIFO(t *testing.T) {
	a, w, tw, _, model := editorFixture(t, "peer")
	var changes []string
	field := focusLossField(t, a, w, tw, "early-loss", &changes)
	tw.nativeReady.Store(true)
	tw.setActiveClient(field.control.client, true)
	source := w.source.Load()
	// The ordinary owner wake can move Go focus before these already admitted
	// native records are delivered. A legacy event must still divide batching.
	source.Custom("earlier-owner-wake")
	for i, update := range []textinput.Update{
		fixtureInput(field, 1, "mark", "A", "Aに", 2, 2, 1, 2),
		fixtureInput(field, 2, "mark", "Aに", "A日本", 3, 3, 1, 3),
		fixtureInput(field, 3, "unmark", "A日本", "A日本", 3, 3, -1, -1),
	} {
		if err := tw.pump.Send(textMessage{Kind: "input", Input: update}); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			source.Custom("legacy-separator")
		}
	}
	field.widget.(*core.TextField).Send(events.FocusLost)
	a.Drain()
	if len(changes) != 0 || !field.control.fieldFocusLost.Load() {
		t.Fatal("focus loss emitted an early controlled Change")
	}
	// This fixture models the native sequence acknowledgment gate, rather than
	// pretending a Cocoa mutation's admission already applied its Go record.
	deactivate := func(r *record) (bool, error) {
		if r != field || len(changes) != 0 {
			t.Fatal("incoming client activated or Change emitted before quiescence")
		}
		return tw.ack >= 3, nil
	}
	var separator bool
	for {
		e, ok := source.Deque.PollEvent()
		if !ok {
			break
		}
		if !handleTextEvent(w, e) && e.Type() == events.Custom && e.(*events.CustomEvent).Data == "legacy-separator" {
			separator = true
			_ = a.Do(func() {
				if tw.ack != 1 || tw.finishFieldFocusLosses(deactivate) {
					t.Fatal("partial marked state bypassed the native acknowledgment gate")
				}
			})
			a.Drain()
			if len(changes) != 0 {
				t.Fatal("partial preedit delivered a controlled Change")
			}
		}
	}
	if !separator || tw.ack != 3 {
		t.Fatal("fixture lost native/legacy ordering", separator, tw.ack)
	}
	_ = a.Do(func() {
		if !tw.finishFieldFocusLosses(deactivate) {
			t.Fatal("acknowledged native field did not finish")
		}
	})
	a.Drain()
	if len(changes) != 1 || changes[0] != "A日本" || field.control.fieldFocusLost.Load() || model.Snapshot().History.UndoCount != 0 {
		t.Fatal("handoff failed exact single Change or mutated peer history", changes)
	}
	_ = a.Do(func() {
		tw.finishFieldFocusLosses(func(*record) (bool, error) { t.Fatal("duplicate native deactivation"); return false, nil })
	})
	a.Drain()
	if len(changes) != 1 {
		t.Fatal("duplicate Change after quiescence")
	}
}

func TestManagedTextFieldFocusLossRetainsUnmarkAndRefocusSelection(t *testing.T) {
	a, w, tw, _, _ := editorFixture(t, "peer")
	var changes []string
	field := focusLossField(t, a, w, tw, "refocus-loss", &changes)
	tw.nativeReady.Store(true)
	tw.setActiveClient(field.control.client, true)
	native := field.widget.(*core.TextField)
	_ = a.Do(func() { tw.apply(fixtureInput(field, 1, "mark", "A", "A日本", 3, 3, 1, 3)) })
	native.Send(events.FocusLost)
	// A -> B -> A retains the loss boundary even though A is focused again.
	native.SetState(true, states.Focused)
	if !native.SetEditingSelection(3, 1) {
		t.Fatal("refocus selection fixture")
	}
	attempts := 0
	deactivate := func(r *record) (bool, error) {
		attempts++
		if attempts == 1 {
			// A full native admission queue must leave the loss pending, with
			// no Change and no acceptance or new-client activation.
			return false, nil
		}
		if tw.ack == 1 {
			if err := tw.pump.Send(textMessage{Kind: "input", Input: fixtureInput(field, 2, "unmark", "A日本", "A日本", 3, 3, -1, -1)}); err != nil {
				t.Fatal(err)
			}
			return false, nil
		}
		return tw.ack == 2, nil
	}
	for range 2 {
		_ = a.Do(func() {
			if tw.finishFieldFocusLosses(deactivate) || !field.control.fieldFocusLost.Load() {
				t.Fatal("unacknowledged unmark or credit refusal cleared pending loss")
			}
		})
	}
	a.Drain()
	if len(changes) != 0 {
		t.Fatal("mark committed before native unmark acknowledgment")
	}
	for {
		e, ok := w.source.Load().Deque.PollEvent()
		if !ok {
			break
		}
		handleTextEvent(w, e)
	}
	_ = a.Do(func() {
		if !tw.finishFieldFocusLosses(deactivate) {
			t.Fatal("unmark acknowledgment not finalized")
		}
	})
	a.Drain()
	state := native.EditingState()
	if len(changes) != 1 || changes[0] != "A日本" || state.Anchor != 3 || state.Focus != 1 || !native.StateIs(states.Focused) {
		t.Fatal("refocus lost selection or single Change", changes, state)
	}
}

func TestManagedTextFieldFocusLossFallbackAndRetirement(t *testing.T) {
	for _, scenario := range []string{"no-native-client", "pre-main", "closed-window", "retired-field"} {
		t.Run(scenario, func(t *testing.T) {
			a, w, tw, _, _ := editorFixture(t, "peer")
			var changes []string
			field := focusLossField(t, a, w, tw, scenario, &changes)
			native := field.widget.(*core.TextField)
			tw.nativeReady.Store(scenario != "pre-main")
			if scenario != "no-native-client" {
				tw.setActiveClient(field.control.client, true)
			}
			if !native.ApplyNativeEdit("legacy", 6, 6, -1, -1, "insert") {
				t.Fatal("fixture edit")
			}
			if scenario == "closed-window" {
				_ = a.Open(app.Window{ID: "surviving-window", Title: "Survivor", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "survivor", Kind: ui.Column} }})
				if err := a.CloseWindow(w.window.Ref()); err != nil {
					t.Fatal(err)
				}
			} else if scenario == "retired-field" {
				native.Send(events.FocusLost)
				disposeControl(field)
			}
			native.Send(events.FocusLost)
			a.Drain()
			if scenario == "retired-field" {
				if len(changes) != 0 || field.control.fieldFocusLost.Load() || field.control.fieldNativeActive.Load() {
					t.Fatal("retired incarnation retained pending native state or emitted Change")
				}
			} else if scenario == "closed-window" {
				if len(changes) != 0 || field.control.fieldFocusLost.Load() {
					t.Fatal("closed attachment emitted Change or stranded a native loss")
				}
			} else if len(changes) != 1 || changes[0] != "legacy" || field.control.fieldFocusLost.Load() {
				t.Fatal("non-native field loss did not retain immediate controlled semantics", changes)
			}
		})
	}
}

func TestManagedTextFieldFocusLossDoesNotActivateInterveningField(t *testing.T) {
	a, w, tw, _, _ := editorFixture(t, "peer")
	var aChanges, bChanges []string
	first := focusLossField(t, a, w, tw, "first-native", &aChanges)
	second := focusLossField(t, a, w, tw, "intervening-legacy", &bChanges)
	tw.nativeReady.Store(true)
	tw.setActiveClient(first.control.client, true)
	_ = a.Do(func() { tw.apply(fixtureInput(first, 1, "mark", "A", "A日本", 3, 3, 1, 3)) })
	first.widget.(*core.TextField).Send(events.FocusLost)
	if !second.widget.(*core.TextField).ApplyNativeEdit("legacy B", 8, 8, -1, -1, "insert") {
		t.Fatal("second field fixture")
	}
	second.widget.(*core.TextField).Send(events.FocusLost)
	a.Drain()
	if len(aChanges) != 0 || len(bChanges) != 1 || second.control.fieldFocusLost.Load() {
		t.Fatal("never-active field did not retain legacy semantics", aChanges, bChanges)
	}
	_ = a.Do(func() {
		if tw.finishFieldFocusLosses(func(r *record) (bool, error) {
			if r != first {
				t.Fatal("activated an intervening field to finalize its loss")
			}
			return false, errors.New("transient native error")
		}) || !first.control.fieldFocusLost.Load() {
			t.Fatal("native error cleared pending field loss")
		}
		tw.apply(fixtureInput(first, 2, "unmark", "A日本", "A日本", 3, 3, -1, -1))
		if !tw.finishFieldFocusLosses(func(r *record) (bool, error) { return r == first && tw.ack == 2, nil }) {
			t.Fatal("first field never finalized")
		}
	})
	a.Drain()
	if len(aChanges) != 1 || aChanges[0] != "A日本" || len(bChanges) != 1 || first.control.fieldNativeActive.Load() {
		t.Fatal("multiple focus losses changed clients or duplicate values", aChanges, bChanges)
	}
	// A pending loss at logical closure is retired, even if a different window
	// keeps the application running and the native event loop has not torn down.
	tw.setActiveClient(first.control.client, true)
	first.widget.(*core.TextField).Send(events.FocusLost)
	_ = a.Open(app.Window{ID: "other-live", Title: "Other", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "other", Kind: ui.Column} }})
	if err := a.CloseWindow(w.window.Ref()); err != nil {
		t.Fatal(err)
	}
	_ = a.Do(func() {
		if tw.finishFieldFocusLosses(func(*record) (bool, error) {
			t.Fatal("closed lifetime dispatched native publication")
			return true, nil
		}) || first.control.fieldFocusLost.Load() {
			t.Fatal("logical closure retained pending field callback")
		}
	})
	a.Drain()
	if len(aChanges) != 1 {
		t.Fatal("closed window delivered pending Change")
	}
}

func TestManagedTextFieldRepublishDoesNotTemporarilyDeactivate(t *testing.T) {
	a, w, tw, _, _ := editorFixture(t, "peer")
	var changes []string
	field := focusLossField(t, a, w, tw, "same-client", &changes)
	tw.setActiveClient(field.control.client, true)
	started, done, ended := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var observedInactive atomic.Bool
	go func() {
		defer close(ended)
		close(started)
		for {
			select {
			case <-done:
				return
			default:
				if !field.control.fieldNativeActive.Load() {
					observedInactive.Store(true)
				}
			}
		}
	}()
	<-started
	err := a.Do(func() {
		for range 100000 {
			tw.setActiveClient(field.control.client, true)
		}
	})
	close(done)
	<-ended
	if err != nil {
		t.Fatal("same-client owner publication", err)
	}
	if observedInactive.Load() {
		t.Fatal("same-client publication exposed a false inactive snapshot to focus callbacks")
	}
	// A real accepted deactivation must still publish its false snapshot.
	if err := a.Do(func() { tw.setActiveClient(field.control.client, false) }); err != nil {
		t.Fatal("owner deactivation", err)
	}
	if field.control.fieldNativeActive.Load() {
		t.Fatal("actual native deactivation retained active snapshot")
	}
}
