package main

import (
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"testing"
)

func TestCatalogueSelectAllUsesActualCommandBitAndMacBinding(t *testing.T) {
	mods := catalogueSelectAllModifiers()
	chord := key.NewChord('a', key.CodeA, mods)
	if !mods.HasFlag(key.Meta) || mods.HasFlag(key.Control) || mods.HasFlag(key.Alt) || mods.HasFlag(key.Shift) || chord != "Meta+A" {
		t.Fatalf("wrong native Select All modifiers/chord: %v %q", mods, chord)
	}
	mac, _, ok := keymap.AvailableMaps.MapByName("MacStandard")
	if !ok || (*mac)[chord] != keymap.SelectAll {
		t.Fatal("Command+A does not map to actual standard Mac Select All")
	}
	// Enum constants identify bit positions. The old literal is a different
	// actual Mac binding; this control catches that representation mistake.
	old := key.NewChord('a', key.CodeA, key.Meta)
	if old != "Control+A" || (*mac)[old] != keymap.Home {
		t.Fatal("counterfactual no longer represents original native guard failure", old)
	}
}
