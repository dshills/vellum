//go:build offscreen

package cogent

import (
	"context"
	"errors"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func TestShortcutRecorderNativeRoutingAndCommands(t *testing.T) {
	calls, commands, invalid := 0, 0, 0
	s := &ui.ShortcutRecorder{ID: "shortcut", Label: "Duplicate shortcut", Value: ui.Shortcut{Key: "D", Modifiers: ui.ShortcutCommand}}
	s.Change = func(e ui.ShortcutEvent) { calls++; s.Value = e.Value }
	s.Invalid = func(error) { invalid++ }
	s.Validate = func(v ui.Shortcut) error {
		if v.Key == "S" {
			return errors.New("already assigned")
		}
		return nil
	}
	view := func() ui.Node {
		n, err := s.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: s.ID, Sequence: 1}, Children: []ui.Node{n, {ID: "other", Kind: ui.Button, Text: "Other"}}}
	}
	m, a := simpleManual(t, manualConfig(t, 1332), view)
	unitSettle(t, m)
	w := m.implementation.find("one").state
	r := findRecord(w.root, s.ID)
	if r.control.shortcut == nil || !r.widget.AsWidget().ContainsFocus() {
		t.Fatal("recorder not bound/focused")
	}
	registry := command.New()
	if err := registry.Register(command.Command{ID: "duplicate", Label: "Duplicate", Shortcut: "Meta+K", Execute: func(context.Context) error { commands++; return nil }}); err != nil {
		t.Fatal(err)
	}
	bindCommandShortcutsScoped(w.body.Scene, a, w.window.Ref(), registry, nil)
	begin := func() {
		t.Helper()
		e := events.NewKey(events.Click, 0, key.CodeUnknown, 0)
		if !activateShortcutRecorder(r, e) {
			t.Fatal("not recorder")
		}
		unitSettle(t, m)
		if !s.Recording() {
			t.Fatal("not recording")
		}
	}
	send := func(rn rune, code key.Codes, mods key.Modifiers) *events.Key {
		t.Helper()
		e := events.NewKey(events.KeyChord, rn, code, mods)
		w.body.Scene.HandleEvent(e)
		return e
	}
	begin()
	e := send('k', key.CodeK, 1<<key.Meta)
	// A second event before owner delivery cannot propose twice or trigger commands.
	send('k', key.CodeK, 1<<key.Meta)
	unitSettle(t, m)
	if !e.IsHandled() || calls != 1 || commands != 0 || s.Value.Key != "K" || s.Recording() {
		t.Fatal("recording routing", e, calls, commands, s.Value)
	}
	send('k', key.CodeK, 1<<key.Meta)
	unitSettle(t, m)
	if commands != 1 {
		t.Fatal("idle recorder swallowed ordinary command", commands)
	}
	begin()
	send('s', key.CodeS, 1<<key.Meta)
	unitSettle(t, m)
	if calls != 1 || invalid != 1 || s.Value.Key != "K" {
		t.Fatal("conflict changed accepted value", calls, invalid, s.Value)
	}
	begin()
	e = send(0, key.CodeEscape, 0)
	unitSettle(t, m)
	if !e.IsHandled() || s.Recording() || calls != 1 {
		t.Fatal("escape cancellation")
	}
	begin()
	e = events.NewKey(events.KeyChord, 0, key.CodeTab, 0)
	if shortcutRecorderPreflight(w.body.Scene, e) || e.IsHandled() {
		t.Fatal("Tab was consumed")
	}
	unitSettle(t, m)
	if s.Recording() || calls != 1 {
		t.Fatal("Tab did not cancel")
	}
	// Clearing while idle uses the same Validate/Change path.
	e = send(0, key.CodeBackspace, 0)
	unitSettle(t, m)
	if !e.IsHandled() || calls != 2 || s.Value != (ui.Shortcut{}) {
		t.Fatal("Delete did not clear", calls, s.Value)
	}
}
func TestShortcutRecorderQueuedLifecycleRetirement(t *testing.T) {
	for _, mode := range []string{"disabled", "hidden", "removed", "replaced", "focus-lost", "window-inactive", "scope-closed", "closed"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			s := &ui.ShortcutRecorder{ID: "shortcut", Change: func(ui.ShortcutEvent) { calls++ }}
			view := func() ui.Node {
				n, err := s.View()
				if err != nil {
					panic(err)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: s.ID, Sequence: 1}, Children: []ui.Node{n}}
			}
			m, a := simpleManual(t, manualConfig(t, 1333), view)
			unitSettle(t, m)
			w := m.implementation.find("one").state
			r := findRecord(w.root, s.ID)
			s.Begin()
			_ = a.Invalidate("one")
			unitSettle(t, m)
			var closedScope *tabCloseScope
			if mode == "scope-closed" {
				closedScope = &tabCloseScope{}
				syncShortcutRecorderVisibility(r, true, []*tabCloseScope{closedScope})
			}
			e := events.NewKey(events.KeyChord, 'k', key.CodeK, 1<<key.Meta)
			if !shortcutRecorderPreflight(w.body.Scene, e) {
				t.Fatal("key not captured")
			}
			switch mode {
			case "disabled":
				r.control.disabled = true
				syncShortcutRecorder(r)
			case "hidden":
				syncShortcutRecorderVisibility(r, false, nil)
			case "removed":
				disposeShortcutRecorder(r)
			case "replaced":
				r.node.ShortcutRecorder = &ui.ShortcutRecorder{ID: s.ID, Change: func(ui.ShortcutEvent) { calls++ }}
				syncShortcutRecorder(r)
			case "focus-lost":
				cancelShortcutBinding(r.control.shortcut)
			case "window-inactive":
				cancelShortcutRecorderScene(w.body.Scene)
			case "scope-closed":
				closedScope.closed.Store(true)
			case "closed":
				if err := a.CloseWindow(w.window.Ref()); err != nil {
					t.Fatal(err)
				}
			}
			a.Drain()
			if calls != 0 {
				t.Fatal("retired input delivered", mode)
			}
		})
	}
}
func TestShortcutRecorderFocusAndActivationLifetime(t *testing.T) {
	calls := 0
	s := &ui.ShortcutRecorder{ID: "shortcut", Change: func(ui.ShortcutEvent) { calls++ }}
	view := func() ui.Node {
		n, _ := s.View()
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: s.ID, Sequence: 1}, Children: []ui.Node{n, {ID: "other", Kind: ui.Button, Text: "Other"}}}
	}
	m, a := simpleManual(t, manualConfig(t, 1334), view)
	unitSettle(t, m)
	w := m.implementation.find("one").state
	r := findRecord(w.root, s.ID)
	// Queued activation cannot leak a command before the visible recording prompt.
	click := events.NewKey(events.Click, 0, key.CodeUnknown, 0)
	activateShortcutRecorder(r, click)
	early := events.NewKey(events.KeyChord, 'q', key.CodeQ, 1<<key.Meta)
	if !shortcutRecorderPreflight(w.body.Scene, early) || !early.IsHandled() {
		t.Fatal("pending activation leaked command")
	}
	unitSettle(t, m)
	findRecord(w.root, "other").widget.AsWidget().SetFocus()
	unitSettle(t, m)
	if s.Recording() || calls != 0 {
		t.Fatal("focus loss did not cancel")
	}
	r.widget.AsWidget().SetFocus()
	unitSettle(t, m)
	activateShortcutRecorder(r, events.NewKey(events.Click, 0, key.CodeUnknown, 0))
	unitSettle(t, m)
	if !s.Recording() {
		t.Fatal("refocused recorder could not restart")
	}
	// An inactive originating window does not capture keys despite retained focus.
	activation := &windowActivation{publishedKnown: true, publishedFocused: false}
	w.host.SetProperty(windowActivationProperty, activation)
	e := events.NewKey(events.KeyChord, 'k', key.CodeK, 1<<key.Meta)
	if shortcutRecorderPreflight(w.body.Scene, e) || e.IsHandled() {
		t.Fatal("inactive window captured key")
	}
	a.Drain()
	if s.Recording() || calls != 0 {
		t.Fatal("inactive window retained recording")
	}
}

func TestShortcutRecorderPendingBeginCanBeCanceled(t *testing.T) {
	for _, mode := range []string{"tab", "shift-tab", "escape", "focus-lost"} {
		t.Run(mode, func(t *testing.T) {
			s := &ui.ShortcutRecorder{ID: "shortcut", Change: func(ui.ShortcutEvent) { t.Error("unexpected value") }}
			view := func() ui.Node {
				n, _ := s.View()
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: s.ID, Sequence: 1}, Children: []ui.Node{n, {ID: "other", Kind: ui.Button, Text: "Other"}}}
			}
			m, _ := simpleManual(t, manualConfig(t, 1335), view)
			unitSettle(t, m)
			w := m.implementation.find("one").state
			r := findRecord(w.root, s.ID)
			activateShortcutRecorder(r, events.NewKey(events.Click, 0, key.CodeUnknown, 0))
			if mode == "focus-lost" {
				findRecord(w.root, "other").widget.AsWidget().SetFocusQuiet()
			} else {
				code, mods := key.CodeTab, key.Modifiers(0)
				if mode == "shift-tab" {
					mods = 1 << key.Shift
				}
				if mode == "escape" {
					code = key.CodeEscape
				}
				e := events.NewKey(events.KeyChord, 0, code, mods)
				consumed := shortcutRecorderPreflight(w.body.Scene, e)
				if mode == "escape" && (!consumed || !e.IsHandled()) {
					t.Fatal("pending Escape not consumed")
				}
				if mode != "escape" && (consumed || e.IsHandled()) {
					t.Fatal("pending Tab consumed")
				}
			}
			unitSettle(t, m)
			if s.Recording() {
				t.Fatal("queued begin survived cancellation", mode)
			}
		})
	}
}
func TestShortcutRecorderRebindsRetainedDialogScene(t *testing.T) {
	calls := 0
	s := &ui.ShortcutRecorder{ID: "shortcut", Change: func(ui.ShortcutEvent) { calls++ }}
	view := func() ui.Node {
		n, _ := s.View()
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: s.ID, Sequence: 1}, Children: []ui.Node{n}}
	}
	m, a := simpleManual(t, manualConfig(t, 1336), view)
	unitSettle(t, m)
	w := m.implementation.find("one").state
	r := findRecord(w.root, s.ID)
	b := r.control.shortcut
	original := r.widget.AsWidget().Scene
	s.Begin()
	syncShortcutRecorder(r)
	syncShortcutRecorderVisibility(r, true, nil)
	oldEpoch := b.scope.Load()
	// Dialog keeps the native child but moves it with WidgetBase.SetScene.
	modal := core.NewBody("Retained dialog transfer").Scene
	defer func() {
		r.widget.AsWidget().SetScene(original)
		syncShortcutRecorder(r)
		syncShortcutRecorderVisibility(r, true, nil)
	}()
	r.widget.AsWidget().SetScene(modal)
	syncShortcutRecorder(r)
	syncShortcutRecorderVisibility(r, true, nil)
	if s.Recording() || b.scope.Load() == oldEpoch || shortcutScene(original).active.Load() == b || shortcutScene(modal) == nil {
		t.Fatal("scene transfer retained old session/registration")
	}
	r.widget.AsWidget().SetFocusQuiet()
	syncShortcutRecorderVisibility(r, true, nil)
	s.Begin()
	syncShortcutRecorder(r)
	syncShortcutRecorderVisibility(r, true, nil)
	e := events.NewKey(events.KeyChord, 'k', key.CodeK, 1<<key.Meta)
	modal.HandleEvent(e)
	a.Drain()
	if !e.IsHandled() || calls != 1 {
		t.Fatal("modal scene preflight not installed", calls)
	}
	s.Begin()
	syncShortcutRecorder(r)
	syncShortcutRecorderVisibility(r, true, nil)
	cancelWindowShortcutRecorders(w.root)
	a.Drain()
	if s.Recording() {
		t.Fatal("window activation loss missed modal recorder")
	}
}

func TestShortcutRecorderRefusedCancellationAndUnpublishedBegin(t *testing.T) {
	t.Run("begun-before-refresh", func(t *testing.T) {
		s := &ui.ShortcutRecorder{ID: "shortcut", Change: func(ui.ShortcutEvent) { t.Error("unexpected value") }}
		view := func() ui.Node {
			n, _ := s.View()
			return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: s.ID, Sequence: 1}, Children: []ui.Node{n}}
		}
		m, a := simpleManual(t, manualConfig(t, 1337), view)
		unitSettle(t, m)
		w := m.implementation.find("one").state
		r := findRecord(w.root, s.ID)
		activateShortcutRecorder(r, events.NewKey(events.Click, 0, key.CodeUnknown, 0))
		a.Drain()
		if !s.Recording() {
			t.Fatal("fixture begin not delivered")
		}
		disposeShortcutRecorder(r)
		if s.Recording() {
			t.Fatal("unpublished begun session survived disposal")
		}
	})
	t.Run("refused-cancel", func(t *testing.T) {
		// This is a direct owner/adapter fixture: it asserts refusal plus capability
		// cleanup without recovering or clearing a terminal Manual window failure.
		a, _ := newAXTest(t, 1)
		host := core.NewFrame(core.NewBody("refused recorder cancellation"))
		s := &ui.ShortcutRecorder{ID: "shortcut", Change: func(ui.ShortcutEvent) { t.Error("unexpected value") }}
		n, _ := s.View()
		r, err := reconcile(host, nil, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		defer disposeControlTree(r)
		s.Begin()
		n, _ = s.View()
		r, err = reconcile(host, r, n, a, "ax")
		if err != nil {
			t.Fatal(err)
		}
		if err := a.Post(func() {}); err != nil {
			t.Fatal(err)
		}
		cancelShortcutBinding(r.control.shortcut)
		refusal := r.control.refused.Load()
		if refusal == nil || r.control.shortcut.cancelled.Load() == nil {
			t.Fatal("refusal/cancellation not retained")
		}
		if err := a.Drain(); err != nil {
			t.Fatal(err)
		}
		syncShortcutRecorder(r)
		if s.Recording() || r.control.shortcut.cancelled.Load() != nil || r.control.refused.Load() != refusal {
			t.Fatal("cancellation cleanup changed failure policy or resurrected recording")
		}
	})
}

// This drives the actual software window FIFO and Core focus traversal, rather
// than merely calling the recorder preflight. Physical keyboard input is separate.
func TestShortcutRecorderFIFOKeyboardNavigation(t *testing.T) {
	calls := 0
	s := &ui.ShortcutRecorder{ID: "shortcut"}
	s.Change = func(e ui.ShortcutEvent) { calls++; s.Value = e.Value }
	view := func() ui.Node {
		n, _ := s.View()
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: s.ID, Sequence: 1}, Children: []ui.Node{n, {ID: "other", Kind: ui.Button, Text: "Other"}}}
	}
	m, _ := simpleManual(t, manualConfig(t, 1338), view)
	unitSettle(t, m)
	w := m.implementation.find("one").state
	source := w.body.SystemEvents()
	source.KeyChord(0, key.CodeReturnEnter, 0)
	unitSettle(t, m)
	if !s.Recording() {
		t.Fatal("Return did not begin through window FIFO")
	}
	source.KeyChord('k', key.CodeK, 1<<key.Meta)
	unitSettle(t, m)
	if calls != 1 || s.Value.Key != "K" {
		t.Fatal("FIFO recording did not commit", calls, s.Value)
	}
	source.KeyChord(0, key.CodeReturnEnter, 0)
	unitSettle(t, m)
	if !s.Recording() {
		t.Fatal("Return did not restart")
	}
	source.KeyChord(0, key.CodeTab, 0)
	unitSettle(t, m)
	if s.Recording() || !findRecord(w.root, "other").widget.AsWidget().ContainsFocus() || calls != 1 {
		t.Fatal("Tab did not cancel and move focus")
	}
	source.KeyChord(0, key.CodeTab, 1<<key.Shift)
	unitSettle(t, m)
	if !findRecord(w.root, s.ID).widget.AsWidget().ContainsFocus() {
		t.Fatal("Shift-Tab did not restore recorder focus")
	}
	source.KeyChord(0, key.CodeBackspace, 0)
	unitSettle(t, m)
	if calls != 2 || s.Value != (ui.Shortcut{}) {
		t.Fatal("FIFO Delete did not clear", calls, s.Value)
	}
}

func TestShortcutRecorderModalEscapeAndTabWithAccessibility(t *testing.T) {
	a, _ := newAXTest(t, 16)
	original := core.NewBody("Modal keyboard origin")
	host := core.NewFrame(original)
	dismissed := 0
	s := &ui.ShortcutRecorder{ID: "shortcut", Change: func(ui.ShortcutEvent) { t.Error("unexpected proposal") }}
	descriptor := func() ui.Node {
		n, _ := s.View()
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "dialog", Kind: ui.Dialog, Text: "Preferences", Open: true, Dismiss: func() { dismissed++ }, Children: []ui.Node{n, {ID: "other", Kind: ui.Button, Text: "Other", Activate: func() {}}}}}}
	}
	n := descriptor()
	root, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	defer disposeControlTree(root)
	w := &nativeWindow{window: a.Windows()[0], body: original, host: host, root: root}
	bindAccessibilityKeyboard(w, a, nil)
	dialog := findRecord(root, "dialog")
	modal := prepareDialogBody(dialog)
	modal.NewDialog(dialog.widget)
	dialog.control.dialog.active = true
	sync := func() {
		t.Helper()
		n = descriptor()
		if err := syncRecord(root, n); err != nil {
			t.Fatal(err)
		}
	}
	sync()
	r := findRecord(root, s.ID)
	r.widget.AsWidget().SetFocusQuiet()
	sync()
	s.Begin()
	sync()
	escape := events.NewKey(events.KeyChord, 0, key.CodeEscape, 0)
	modal.Scene.HandleEvent(escape)
	a.Drain()
	sync()
	if !escape.IsHandled() || s.Recording() || dismissed != 0 {
		t.Fatal("recording Escape dismissed modal", dismissed, s.Recording())
	}
	s.Begin()
	sync()
	tab := events.NewKey(events.KeyChord, 0, key.CodeTab, 0)
	modal.Scene.HandleEvent(tab)
	a.Drain()
	// Headless scenes defer accessibility focus until Shown, just like initial
	// modal presentation. Deliver that production lifecycle hook explicitly.
	modal.Shown()
	sync()
	if !tab.IsHandled() || s.Recording() || !findRecord(root, "other").widget.AsWidget().ContainsFocus() || dismissed != 0 {
		t.Fatal("modal Tab did not cancel and move focus", dismissed, s.Recording(), "handled", tab.IsHandled(), "focused", accessibilityFocusedID(root), "order", ui.FocusOrder(accessibilityFocusProjection(root, "")), "target scene", findRecord(root, "other").widget.AsWidget().Scene == modal.Scene)
	}
	// After recording ends, Escape returns to the established modal dismissal path.
	idleEscape := events.NewKey(events.KeyChord, 0, key.CodeEscape, 0)
	modal.Scene.HandleEvent(idleEscape)
	a.Drain()
	if !idleEscape.IsHandled() || dismissed != 1 {
		t.Fatal("idle modal Escape not restored", dismissed)
	}
}
