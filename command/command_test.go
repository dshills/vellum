package command

import (
	"context"
	"errors"
	"testing"
)

func TestEveryEntryRechecksAvailability(t *testing.T) {
	r := New()
	allowed := true
	n := 0
	err := r.Register(Command{ID: "save", Label: "Save", Shortcut: "Meta+S", Available: func() error {
		if !allowed {
			return errors.New("no document")
		}
		return nil
	}, Execute: func(context.Context) error { n++; return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Invoke(context.Background(), "save"); err != nil {
		t.Fatal(err)
	}
	allowed = false
	if r.Invoke(context.Background(), "save") == nil {
		t.Fatal("disabled action")
	}
	if matched, err := r.Shortcut(context.Background(), "Meta+S"); !matched || err == nil {
		t.Fatal("shortcut bypass")
	}
	if n != 1 {
		t.Fatal(n)
	}
	if len(r.Search("SAVE")) != 1 {
		t.Fatal("search")
	}
	if r.Register(Command{ID: "another", Label: "Other", Shortcut: "meta+s", Execute: func(context.Context) error { return nil }}) == nil {
		t.Fatal("duplicate shortcut")
	}
}

func TestShortcutPhasesRetainValidationAndUniqueness(t *testing.T) {
	r := New()
	calls := 0
	enabled := true
	c := Command{ID: "undo", Label: "Undo", Shortcut: "Meta+Z", ShortcutScope: AfterFocusShortcuts,
		Available: func() error {
			if !enabled {
				return errors.New("empty")
			}
			return nil
		}, Execute: func(context.Context) error { calls++; return nil }}
	if err := r.Register(c); err != nil {
		t.Fatal(err)
	}
	if matched, err := r.ShortcutInScope(context.Background(), "meta+z", WindowShortcuts); matched || err != nil || calls != 0 {
		t.Fatal("fallback ran before focus")
	}
	if matched, err := r.ShortcutInScope(context.Background(), "Meta+Z", AfterFocusShortcuts); !matched || err != nil || calls != 1 {
		t.Fatal(matched, err, calls)
	}
	enabled = false
	if matched, err := r.ShortcutInScope(context.Background(), "Meta+Z", AfterFocusShortcuts); !matched || err == nil || calls != 1 {
		t.Fatal("disabled fallback ran")
	}
	c.ID = "duplicate"
	c.ShortcutScope = WindowShortcuts
	if r.Register(c) == nil {
		t.Fatal("cross-phase duplicate shortcut")
	}
	c.ID = "invalid"
	c.Shortcut = ""
	c.ShortcutScope = 99
	if r.Register(c) == nil {
		t.Fatal("invalid shortcut scope")
	}
	if matched, err := r.ShortcutInScope(context.Background(), "Meta+Z", 99); matched || err == nil {
		t.Fatal("invalid routing scope")
	}
}
