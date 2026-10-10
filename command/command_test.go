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
