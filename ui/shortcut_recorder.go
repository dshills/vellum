package ui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ShortcutModifiers describe a single local shortcut. Command is the macOS
// Command key; Control remains the distinct Control key.
type ShortcutModifiers uint8

const (
	ShortcutControl ShortcutModifiers = 1 << iota
	ShortcutOption
	ShortcutShift
	ShortcutCommand
)

// Shortcut is one key and its modifiers, not a global registration. Keys are
// uppercase printable ASCII or one of the documented named keys. The zero value
// means no shortcut. Plain function keys are accepted; other keys require
// Command, Control or Option, so text entry and ordinary navigation stay usable.
type Shortcut struct {
	Key       string
	Modifiers ShortcutModifiers
}

func (s Shortcut) Validate() error {
	if s == (Shortcut{}) {
		return nil
	}
	if s.Modifiers & ^(ShortcutControl|ShortcutOption|ShortcutShift|ShortcutCommand) != 0 {
		return fmt.Errorf("ui: unknown shortcut modifier")
	}
	fn := false
	if strings.HasPrefix(s.Key, "F") {
		if n, e := strconv.Atoi(strings.TrimPrefix(s.Key, "F")); e == nil && n >= 1 && n <= 24 && s.Key == fmt.Sprintf("F%d", n) {
			fn = true
		}
	}
	named := map[string]bool{"Space": true, "Return": true, "Tab": true, "Escape": true, "Delete": true, "ForwardDelete": true, "Left": true, "Right": true, "Up": true, "Down": true, "Home": true, "End": true, "PageUp": true, "PageDown": true}
	printable := len(s.Key) == 1 && s.Key[0] >= 33 && s.Key[0] <= 126 && strings.ToUpper(s.Key) == s.Key
	if !fn && !named[s.Key] && !printable {
		return fmt.Errorf("ui: invalid shortcut key %q", s.Key)
	}
	if !fn && s.Modifiers&(ShortcutControl|ShortcutOption|ShortcutCommand) == 0 {
		return fmt.Errorf("ui: shortcut requires Command, Control or Option")
	}
	return nil
}

// Label uses the conventional macOS modifier order and key glyphs.
func (s Shortcut) Label() string {
	if s == (Shortcut{}) {
		return "Record shortcut"
	}
	text := ""
	for _, m := range []struct {
		flag  ShortcutModifiers
		glyph string
	}{{ShortcutControl, "⌃"}, {ShortcutOption, "⌥"}, {ShortcutShift, "⇧"}, {ShortcutCommand, "⌘"}} {
		if s.Modifiers&m.flag != 0 {
			text += m.glyph
		}
	}
	labels := map[string]string{"Space": "Space", "Return": "↩", "Tab": "⇥", "Escape": "⎋", "Delete": "⌫", "ForwardDelete": "⌦", "Left": "←", "Right": "→", "Up": "↑", "Down": "↓", "Home": "↖", "End": "↘", "PageUp": "⇞", "PageDown": "⇟"}
	if k, ok := labels[s.Key]; ok {
		return text + k
	}
	return text + s.Key
}

// Chord returns Cogent/command.Registry's canonical modifier names. It does not
// register or execute the shortcut. Applications still own command registration.
func (s Shortcut) Chord() string {
	if s == (Shortcut{}) {
		return ""
	}
	text := ""
	for _, m := range []struct {
		flag ShortcutModifiers
		name string
	}{{ShortcutControl, "Control+"}, {ShortcutCommand, "Meta+"}, {ShortcutOption, "Alt+"}, {ShortcutShift, "Shift+"}} {
		if s.Modifiers&m.flag != 0 {
			text += m.name
		}
	}
	names := map[string]string{"Space": "Spacebar", "Return": "ReturnEnter", "Escape": "Escape", "Delete": "Backspace", "ForwardDelete": "Delete", "Left": "LeftArrow", "Right": "RightArrow", "Up": "UpArrow", "Down": "DownArrow"}
	if key, ok := names[s.Key]; ok {
		return text + key
	}
	return text + s.Key
}

type ShortcutEvent struct {
	ID    string
	Value Shortcut
}
type shortcutRecorderIdentity struct{ id string }

// ShortcutRecorder is a retained, owner-held shortcut editor. A click or Return
// starts one local recording session. Change proposes an accepted, validated
// chord; the application must assign Value itself. Validate may reject conflicts
// (including clearing); Invalid can display the error. Retain on one owner and
// do not mutate concurrently. Recording never installs a system/global hotkey.
type ShortcutRecorder struct {
	ID, Label string
	Value     Shortcut
	Disabled  bool
	Validate  func(Shortcut) error
	Change    func(ShortcutEvent)
	Invalid   func(error)
	Style     Style
	recording bool
	session   uint64
	identity  *shortcutRecorderIdentity
}

func (s *ShortcutRecorder) available() bool {
	return s != nil && s.ID != "" && !s.Disabled && s.Change != nil
}
func (s *ShortcutRecorder) Begin() bool {
	if !s.available() {
		return false
	}
	s.session++
	if s.session == 0 {
		s.session++
	}
	s.recording = true
	return true
}
func (s *ShortcutRecorder) Recording() bool { return s != nil && s.recording && s.available() }
func (s *ShortcutRecorder) Session() uint64 {
	if s == nil {
		return 0
	}
	return s.session
}

// Cancel affects only the specified session, so stale native input cannot stop
// a later recording. Zero is not a recording session.
func (s *ShortcutRecorder) Cancel(session uint64) {
	if s != nil && session != 0 && s.session == session {
		s.recording = false
	}
}

// Record completes a live session through the same validation path for a chord
// or a cleared zero value. Rejected values leave Value unchanged.
func (s *ShortcutRecorder) Record(session uint64, value Shortcut) error {
	if !s.Recording() || session == 0 || s.session != session {
		return nil
	}
	s.recording = false
	id := s.ID
	err := value.Validate()
	if err == nil && s.Validate != nil {
		err = s.Validate(value)
	}
	if err != nil {
		if s.Invalid != nil {
			s.Invalid(err)
		}
		return err
	}
	if !s.available() || s.ID != id || s.session != session || s.recording {
		return nil
	}
	if value != s.Value {
		s.Change(ShortcutEvent{ID: s.ID, Value: value})
	}
	return nil
}
func (s *ShortcutRecorder) View() (Node, error) {
	if s == nil || s.ID == "" || len(s.ID) > 256 || !utf8.ValidString(s.ID) || strings.ContainsAny(s.ID, "\x00\r\n\t") {
		return Node{}, fmt.Errorf("ui: shortcut recorder requires a bounded ID")
	}
	if err := s.Value.Validate(); err != nil {
		return Node{}, err
	}
	if s.identity == nil || s.identity.id != s.ID {
		s.recording = false
		s.identity = &shortcutRecorderIdentity{id: s.ID}
	}
	if !s.available() {
		s.recording = false
	}
	identity := s.identity
	label := s.Label
	if label == "" {
		label = "Keyboard shortcut"
	}
	text := s.Value.Label()
	if s.Recording() {
		text = "Type shortcut…"
	}
	n := Node{ID: s.ID, Kind: Button, Text: text, Disabled: !s.available(), ShortcutRecorder: s, Style: Style{MinSize: Some(Size{Width: 156, Height: 28}), GrowX: Some(float32(0)), Padding: Some(Insets{Top: 5, Bottom: 5, Left: 12, Right: 12})}.Overlay(s.Style), Accessibility: Accessibility{Name: label, Help: "Current: " + text + ". Press to record. Escape or Tab cancels; Delete clears. Requires Command, Control or Option, or a function key."}, Activate: func() {
		if s.identity == identity && s.ID == identity.id {
			s.Begin()
		}
	}}
	return n, Validate(n)
}

// ValidateShortcutRecorderNode keeps recorder metadata on its matching managed
// button. Drivers can call it alongside ValidateControl.
func ValidateShortcutRecorderNode(n Node) error {
	if n.ShortcutRecorder == nil {
		return nil
	}
	if n.Kind != Button || n.ID != n.ShortcutRecorder.ID || n.Activate == nil || n.PlacementSource != nil {
		return fmt.Errorf("ui: shortcut recorder metadata requires its matching dedicated button and activation")
	}
	if !n.Disabled && !n.ShortcutRecorder.available() {
		return fmt.Errorf("ui: unavailable shortcut recorder must be disabled")
	}
	return n.ShortcutRecorder.Value.Validate()
}
