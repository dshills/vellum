package ui

import (
	"errors"
	"strings"
	"testing"
)

func TestShortcutCanonicalMacLabelsAndChords(t *testing.T) {
	s := Shortcut{Key: "K", Modifiers: ShortcutCommand | ShortcutShift | ShortcutOption | ShortcutControl}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if s.Label() != "⌃⌥⇧⌘K" || s.Chord() != "Control+Meta+Alt+Shift+K" {
		t.Fatal(s.Label(), s.Chord())
	}
	for _, good := range []Shortcut{{}, {Key: "F1"}, {Key: "F24", Modifiers: ShortcutShift}, {Key: "Left", Modifiers: ShortcutOption}, {Key: "+", Modifiers: ShortcutCommand}} {
		if err := good.Validate(); err != nil {
			t.Error(good, err)
		}
	}
	for _, bad := range []Shortcut{{Key: "A"}, {Key: "A", Modifiers: ShortcutShift}, {Key: "a", Modifiers: ShortcutCommand}, {Key: "F25"}, {Key: "F01"}, {Key: "Å", Modifiers: ShortcutCommand}, {Key: "A B", Modifiers: ShortcutCommand}, {Modifiers: ShortcutCommand}, {Key: "K", Modifiers: 128}} {
		if bad.Validate() == nil {
			t.Error("accepted invalid", bad)
		}
	}
}
func TestShortcutRecorderControlledValidationAndCancellation(t *testing.T) {
	changes, invalid := 0, 0
	original := Shortcut{Key: "D", Modifiers: ShortcutCommand}
	s := &ShortcutRecorder{ID: "duplicate", Value: original, Change: func(e ShortcutEvent) {
		changes++
		if e.ID != "duplicate" {
			t.Error(e)
		}
	}, Invalid: func(error) { invalid++ }}
	n, err := s.View()
	if err != nil {
		t.Fatal(err)
	}
	n.Activate()
	session := s.Session()
	if !s.Recording() {
		t.Fatal("activation did not begin")
	}
	if err = s.Record(session, Shortcut{Key: "K", Modifiers: ShortcutCommand}); err != nil {
		t.Fatal(err)
	}
	if changes != 1 || s.Value != original || s.Recording() {
		t.Fatal("not controlled", changes, s.Value)
	}
	_ = s.Record(session, Shortcut{})
	if changes != 1 {
		t.Fatal("duplicate settlement")
	}
	s.Begin()
	old := s.Session()
	s.Begin()
	current := s.Session()
	s.Cancel(old)
	if !s.Recording() {
		t.Fatal("stale cancellation stopped current session")
	}
	s.Cancel(current)
	if s.Recording() {
		t.Fatal("cancel failed")
	}
	s.Validate = func(v Shortcut) error {
		if v.Key == "S" || v == (Shortcut{}) {
			return errors.New("already assigned")
		}
		return nil
	}
	for _, value := range []Shortcut{{Key: "S", Modifiers: ShortcutCommand}, {}, {Key: "K"}} {
		s.Begin()
		if s.Record(s.Session(), value) == nil {
			t.Error("accepted rejected", value)
		}
	}
	if invalid != 3 || changes != 1 || s.Value != original {
		t.Fatal(invalid, changes, s.Value)
	}
	s.Validate = nil
	s.Begin()
	_ = s.Record(s.Session(), Shortcut{})
	if changes != 2 || s.Value != original {
		t.Fatal("clear not proposed through controlled callback")
	}
}
func TestShortcutRecorderStaleViewsAndReadOnly(t *testing.T) {
	s := &ShortcutRecorder{ID: "one", Change: func(ShortcutEvent) {}}
	old, _ := s.View()
	s.ID = "two"
	_, _ = s.View()
	old.Activate()
	if s.Recording() {
		t.Fatal("retired identity activated")
	}
	s.Begin()
	s.Disabled = true
	n, err := s.View()
	if err != nil || !n.Disabled || s.Recording() {
		t.Fatal(n, err)
	}
	n.Activate()
	if s.Recording() {
		t.Fatal("disabled recorder began")
	}
	s.Disabled = false
	s.Change = nil
	n, _ = s.View()
	if !n.Disabled || s.Begin() {
		t.Fatal("nil callback writable")
	}
	s.Change = func(ShortcutEvent) {}
	n, _ = s.View()
	if n.ShortcutRecorder != s || n.Kind != Button || !strings.Contains(n.Accessibility.Help, "Escape") {
		t.Fatal(n)
	}
	wrong := n
	wrong.Kind = TextField
	if ValidateShortcutRecorderNode(wrong) == nil {
		t.Fatal("metadata allowed on text field")
	}
	wrong = n
	wrong.ID = "other"
	if ValidateShortcutRecorderNode(wrong) == nil {
		t.Fatal("mismatched identity allowed")
	}
}
func TestShortcutRecorderValidationMayRetireOrRestart(t *testing.T) {
	calls := 0
	s := &ShortcutRecorder{ID: "r", Change: func(ShortcutEvent) { calls++ }}
	_, _ = s.View()
	s.Validate = func(Shortcut) error { s.Change = nil; return nil }
	s.Begin()
	_ = s.Record(s.Session(), Shortcut{Key: "K", Modifiers: ShortcutCommand})
	if calls != 0 {
		t.Fatal("retired validation delivered")
	}
	s.Change = func(ShortcutEvent) { calls++ }
	s.Validate = func(Shortcut) error { s.Begin(); return nil }
	s.Begin()
	_ = s.Record(s.Session(), Shortcut{Key: "K", Modifiers: ShortcutCommand})
	if calls != 0 || !s.Recording() {
		t.Fatal("restarted session was settled by old proposal")
	}
}
