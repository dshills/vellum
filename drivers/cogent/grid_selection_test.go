package cogent

import (
	"testing"

	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/virtual"
)

func gridSelectionModifiers(flags ...key.Modifiers) key.Modifiers {
	var out key.Modifiers
	for _, flag := range flags {
		out.SetFlag(true, flag)
	}
	return out
}

func TestManagedGridSelectionPointerModifiers(t *testing.T) {
	cases := []struct {
		name     string
		mods     key.Modifiers
		mode     virtual.GridSelectMode
		accepted bool
	}{
		{"replace", 0, virtual.GridSelectReplace, true},
		{"shift", gridSelectionModifiers(key.Shift), virtual.GridSelectExtend, true},
		{"command", gridSelectionModifiers(key.Meta), virtual.GridSelectToggle, true},
		{"lock-flags", gridSelectionModifiers(key.CapsLock, key.NumLock, key.Shift), virtual.GridSelectExtend, true},
		{"control", gridSelectionModifiers(key.Control), virtual.GridSelectReplace, false},
		{"option", gridSelectionModifiers(key.Alt), virtual.GridSelectReplace, false},
		{"mixed", gridSelectionModifiers(key.Shift, key.Meta), virtual.GridSelectReplace, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mode, accepted := gridPointerSelectionMode(c.mods)
			if mode != c.mode || accepted != c.accepted {
				t.Fatal("selection modifier admission", mode, accepted, c)
			}
		})
	}
}
