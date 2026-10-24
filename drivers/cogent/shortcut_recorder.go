package cogent

import (
	"strings"
	"sync/atomic"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

const shortcutRecorderSceneProperty = "vellum.shortcut-recorder"

// Native handlers read only this immutable owner publication, atomic lifetime
// capabilities and their own window's focus. Model callbacks never walk native
// parents or inspect another window's focus.
type shortcutRecorderScope struct {
	recorder       *ui.ShortcutRecorder
	session        uint64
	recording      bool
	epoch          *tabCloseEpoch
	claimed        atomic.Bool
	beginRequested atomic.Bool
	started        atomic.Uint64
}
type shortcutRecorderBinding struct {
	r           *record
	button      *core.Button
	scene       atomic.Pointer[shortcutRecorderScene]
	nativeScene atomic.Pointer[core.Scene]
	scope       atomic.Pointer[shortcutRecorderScope]
	cancelled   atomic.Pointer[shortcutRecorderScope]
	managed     atomic.Bool
}
type shortcutRecorderScene struct {
	active atomic.Pointer[shortcutRecorderBinding]
}

func shortcutScene(scene *core.Scene) *shortcutRecorderScene {
	if scene == nil {
		return nil
	}
	s, _ := scene.Property(shortcutRecorderSceneProperty).(*shortcutRecorderScene)
	return s
}
func bindShortcutRecorder(r *record, button *core.Button) {
	if r.node.ShortcutRecorder == nil {
		return
	}
	b := &shortcutRecorderBinding{r: r, button: button}
	r.control.shortcut = b
	button.OnFocus(func(events.Event) {
		if state := b.scene.Load(); state != nil {
			state.active.Store(b)
		}
	})
	button.OnFocusLost(func(events.Event) { cancelShortcutBinding(b) })
	button.OnFirst(events.KeyChord, func(e events.Event) { shortcutRecorderPreflight(button.Scene, e) })
	syncShortcutRecorderScene(b)
}

// Retained dialog content moves between native scenes. Revoke the former scene
// before publishing a new preflight on the current native scene.
func syncShortcutRecorderScene(b *shortcutRecorderBinding) {
	scene := b.button.Scene
	if b.nativeScene.Load() == scene {
		return
	}
	if state := b.scene.Load(); state != nil {
		state.active.CompareAndSwap(b, nil)
	}
	if old := b.scope.Swap(nil); old != nil && (old.recording || old.beginRequested.Load()) {
		old.recorder.Cancel(shortcutScopeSession(old))
		requestControlRestore(b.r)
	}
	state := shortcutScene(scene)
	if state == nil && scene != nil {
		state = &shortcutRecorderScene{}
		scene.SetProperty(shortcutRecorderSceneProperty, state)
		scene.OnFirst(events.KeyChord, func(e events.Event) { shortcutRecorderPreflight(scene, e) })
	}
	b.nativeScene.Store(scene)
	b.scene.Store(state)
}

// Activation belongs to the window, while modal content has a different Scene.
// Walk retained descriptor records on the originating native owner; delivery
// still reads only immutable scopes and owner-held component state.
func cancelWindowShortcutRecorders(r *record) {
	if r == nil {
		return
	}
	if r.control != nil && r.control.shortcut != nil {
		b := r.control.shortcut
		if s := b.scope.Load(); s != nil && (s.recording || s.beginRequested.Load()) {
			cancelShortcutBinding(b)
		}
	}
	for _, child := range r.children {
		cancelWindowShortcutRecorders(child)
	}
}
func shortcutOwnerLive(r *record, b *shortcutRecorderBinding, s *shortcutRecorderScope) bool {
	return r != nil && r.control != nil && !r.control.retired && !r.control.disabled && r.control.shortcut == b && b.scope.Load() == s && s != nil && tabCloseEpochLive(s.epoch) && r.node.ShortcutRecorder == s.recorder
}

// activateShortcutRecorder replaces generic Button activation with an immutable
// capability, preventing a queued click from beginning a replacement component.
func activateShortcutRecorder(r *record, e events.Event) bool {
	b := r.control.shortcut
	if b == nil || !b.managed.Load() {
		return false
	}
	s := b.scope.Load()
	if s == nil || !tabCloseEpochLive(s.epoch) {
		return true
	}
	e.SetHandled()
	if !s.beginRequested.CompareAndSwap(false, true) {
		return true
	}
	admitted := postControl(r, func() {
		if shortcutOwnerLive(r, b, s) && r.node.Activate != nil {
			r.node.Activate()
			s.started.Store(s.recorder.Session())
			requestControlRestore(r)
		}
	})
	if !admitted {
		s.beginRequested.Store(false)
	}
	return true
}
func shortcutScopeSession(s *shortcutRecorderScope) uint64 {
	if started := s.started.Load(); started != 0 {
		return started
	}
	return s.session
}
func cancelShortcutBinding(b *shortcutRecorderBinding) {
	if b == nil {
		return
	}
	if state := b.scene.Load(); state != nil {
		state.active.CompareAndSwap(b, nil)
	}
	s := b.scope.Swap(nil)
	if s == nil {
		return
	}
	b.cancelled.Store(s)
	postControl(b.r, func() { settleShortcutCancellation(b, s) })
}

// A refused cancellation post must not resurrect recording on refresh. Retain
// only its latest capability; originating reconciliation settles it on the owner.
func settleShortcutCancellation(b *shortcutRecorderBinding, s *shortcutRecorderScope) {
	s.recorder.Cancel(shortcutScopeSession(s))
	b.cancelled.CompareAndSwap(s, nil)
	if !b.r.control.retired && b.r.node.ShortcutRecorder == s.recorder {
		requestControlRestore(b.r)
	}
}

// Called on originating native activation loss. Logical focus remains retained,
// but recording cannot survive a window switch and capture the next window's key.
func cancelShortcutRecorderScene(scene *core.Scene) {
	if state := shortcutScene(scene); state != nil {
		cancelShortcutBinding(state.active.Load())
	}
}
func syncShortcutRecorder(r *record) {
	b := r.control.shortcut
	if b == nil && r.node.ShortcutRecorder != nil {
		if button, ok := r.widget.(*core.Button); ok {
			bindShortcutRecorder(r, button)
			b = r.control.shortcut
		}
	}
	if b == nil {
		return
	}
	if cancelled := b.cancelled.Swap(nil); cancelled != nil {
		settleShortcutCancellation(b, cancelled)
	}
	syncShortcutRecorderScene(b)
	b.managed.Store(r.node.ShortcutRecorder != nil)
	s := b.scope.Load()
	recorder := r.node.ShortcutRecorder
	if recorder == nil || r.control.disabled {
		if s != nil {
			b.scope.Store(nil)
			if s.recording || s.beginRequested.Load() {
				s.recorder.Cancel(shortcutScopeSession(s))
				requestControlRestore(r)
			}
		}
		return
	}
	if s != nil && (s.recorder != recorder || s.session != recorder.Session() || s.recording != recorder.Recording()) {
		b.scope.Store(nil)
		if s.recorder != recorder && (s.recording || s.beginRequested.Load()) {
			s.recorder.Cancel(shortcutScopeSession(s))
		}
	}
}
func syncShortcutRecorderVisibility(r *record, visible bool, scopes []*tabCloseScope) {
	if r.control == nil || r.control.shortcut == nil {
		return
	}
	b := r.control.shortcut
	recorder := r.node.ShortcutRecorder
	s := b.scope.Load()
	if !visible || r.control.disabled || recorder == nil {
		if s != nil {
			b.scope.Store(nil)
			if s.recording || s.beginRequested.Load() {
				s.recorder.Cancel(shortcutScopeSession(s))
				requestControlRestore(r)
			}
		}
		return
	}
	if s == nil || s.recorder != recorder || s.session != recorder.Session() || s.recording != recorder.Recording() || !tabCloseSameScopes(s.epoch.scopes, scopes) {
		b.scope.Store(&shortcutRecorderScope{recorder: recorder, session: recorder.Session(), recording: recorder.Recording(), epoch: &tabCloseEpoch{scopes: append([]*tabCloseScope(nil), scopes...)}})
	}
	// Native SetFocus may precede initial visibility publication.
	if b.button.ContainsFocus() {
		if state := b.scene.Load(); state != nil {
			state.active.Store(b)
		}
	}
}
func disposeShortcutRecorder(r *record) {
	if r == nil || r.control == nil || r.control.shortcut == nil {
		return
	}
	b := r.control.shortcut
	if cancelled := b.cancelled.Swap(nil); cancelled != nil {
		cancelled.recorder.Cancel(shortcutScopeSession(cancelled))
	}
	if state := b.scene.Load(); state != nil {
		state.active.CompareAndSwap(b, nil)
	}
	if s := b.scope.Swap(nil); s != nil && (s.recording || s.beginRequested.Load()) {
		s.recorder.Cancel(shortcutScopeSession(s))
	}
}
func shortcutFromEvent(e events.Event) ui.Shortcut {
	var mods ui.ShortcutModifiers
	for _, m := range []struct {
		native   key.Modifiers
		portable ui.ShortcutModifiers
	}{{key.Control, ui.ShortcutControl}, {key.Alt, ui.ShortcutOption}, {key.Shift, ui.ShortcutShift}, {key.Meta, ui.ShortcutCommand}} {
		if e.Modifiers()&(1<<m.native) != 0 {
			mods |= m.portable
		}
	}
	names := map[key.Codes]string{key.CodeSpacebar: "Space", key.CodeReturnEnter: "Return", key.CodeKeypadEnter: "Return", key.CodeTab: "Tab", key.CodeEscape: "Escape", key.CodeBackspace: "Delete", key.CodeDelete: "ForwardDelete", key.CodeLeftArrow: "Left", key.CodeRightArrow: "Right", key.CodeUpArrow: "Up", key.CodeDownArrow: "Down", key.CodeHome: "Home", key.CodeEnd: "End", key.CodePageUp: "PageUp", key.CodePageDown: "PageDown"}
	name := names[e.KeyCode()]
	if name == "" {
		if e.KeyRune() != 0 {
			name = strings.ToUpper(string(e.KeyRune()))
		} else {
			name = strings.TrimPrefix(e.KeyCode().String(), "Code")
		}
	}
	return ui.Shortcut{Key: name, Modifiers: mods}
}

// Native focused-window preflight runs before application command handlers.
// Returning true means the event has been consumed; plain Tab cancels and passes
// through so normal focus traversal still runs. Modifiers alone keep recording.
func shortcutRecorderPreflight(scene *core.Scene, e events.Event) bool {
	if e.IsHandled() {
		return true
	}
	state := shortcutScene(scene)
	if state == nil {
		return false
	}
	b := state.active.Load()
	if b == nil {
		return false
	}
	s := b.scope.Load()
	if s == nil || !tabCloseEpochLive(s.epoch) || b.nativeScene.Load() != scene || !b.button.ContainsFocus() {
		return false
	}
	if !nativeWindowActive(b.r) {
		cancelShortcutBinding(b)
		return false
	}
	mods := e.Modifiers() & ^((1 << key.CapsLock) | (1 << key.NumLock))
	plain := mods == 0
	tab := e.KeyCode() == key.CodeTab && (plain || mods == 1<<key.Shift)
	clear := plain && (e.KeyCode() == key.CodeBackspace || e.KeyCode() == key.CodeDelete)
	// A clicked recorder owns command chords while its owner publishes the
	// recording prompt. These early keys are consumed rather than escaping to
	// an unrelated application command. Recording begins with the visible prompt.
	if !s.recording && s.beginRequested.Load() {
		if tab || plain && e.KeyCode() == key.CodeEscape {
			cancelShortcutBinding(b)
		}
		if !tab {
			e.SetHandled()
		}
		return !tab
	}
	if !s.recording && !clear {
		return false
	}
	if key.CodeIsModifier(e.KeyCode()) {
		e.SetHandled()
		return true
	}
	cancel := s.recording && (tab || plain && e.KeyCode() == key.CodeEscape)
	if cancel {
		cancelShortcutBinding(b)
		if !tab {
			e.SetHandled()
		}
		return !tab
	}
	if !tab {
		e.SetHandled()
	}
	if !s.claimed.CompareAndSwap(false, true) {
		return !tab
	}
	value := shortcutFromEvent(e)
	if clear {
		value = ui.Shortcut{}
	}
	admitted := postControl(b.r, func() {
		if !shortcutOwnerLive(b.r, b, s) {
			return
		}

		session := s.session
		if !s.recording {
			if !s.recorder.Begin() {
				return
			}
			session = s.recorder.Session()
		}
		_ = s.recorder.Record(session, value)
		requestControlRestore(b.r)
	})
	if !admitted {
		s.claimed.Store(false)
	}
	return !tab
}
