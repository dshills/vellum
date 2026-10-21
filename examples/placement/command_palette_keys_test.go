//go:build offscreen

package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/canvas"
	"strings"
	"testing"
)

func TestCommandPaletteNativeInputAndReturnAdmission(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() {
		if err := f.m.beginCommandPalette(); err != nil {
			t.Fatal(err)
		}
	})
	body := core.NewBody()
	field := core.NewTextField(body)
	field.SetName("command-query")
	core.NewFrame(body).SetName("command-results")
	e := &commandPaletteKeys{a: f.a, ref: f.ref, m: f.m}
	widget := e.Mount(body)
	if err := e.Sync(widget); err != nil {
		t.Fatal(err)
	}
	if err := e.Sync(widget); err != nil {
		t.Fatal(err)
	}
	field.SetText("nudge-1-right")
	field.Send(events.Input)
	if f.m.commandDraft.selected != "nudge-1-right" {
		t.Fatal("live input not routed")
	}
	field.SetText(strings.Repeat("x", 129))
	field.HandleEvent(events.NewKey(events.KeyChord, 0, key.CodeReturnEnter, 0))
	if f.m.scene.Objects[0].X != 0 || f.m.commandDraft == nil {
		t.Fatal("invalid search ran prior command")
	}
	field.SetText("nudge-1-right")
	field.HandleEvent(events.NewKey(events.KeyChord, 0, key.CodeReturnEnter, 0))
	if f.m.scene.Objects[0].X != 1 || f.m.commandDraft != nil || f.m.journal.Snapshot().UndoCount != 2 {
		t.Fatal("Return did not run exactly once")
	}
}

func TestCommandPaletteOldWidgetDraftCannotRunNewPalette(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.owner(t, func() { _ = f.m.beginCommandPalette() })
	body := core.NewBody()
	field := core.NewTextField(body)
	field.SetName("command-query")
	e := &commandPaletteKeys{a: f.a, ref: f.ref, m: f.m}
	widget := e.Mount(body)
	_ = e.Sync(widget)
	f.owner(t, func() {
		f.m.cancelCommandPalette(f.m.commandDraft)
		_ = f.m.beginCommandPalette()
		f.m.updateCommandQuery(f.m.commandDraft, "nudge-1-right")
	})
	// The retained old widget has not been synchronized into the new opening.
	field.SetText("nudge-1-right")
	field.HandleEvent(events.NewKey(events.KeyChord, 0, key.CodeReturnEnter, 0))
	if f.m.scene.Objects[0].X != 0 || f.m.commandDraft == nil {
		t.Fatal("old widget ran new draft")
	}
}

func TestCommandPaletteNativeReturnIgnoresLockState(t *testing.T) {
	for _, mods := range []key.Modifiers{1 << key.NumLock, 1 << key.CapsLock, 1<<key.NumLock | 1<<key.CapsLock} {
		t.Run(mods.String(), func(t *testing.T) {
			f := fileModel(t, false)
			f.place(t, "block", canvas.Point{})
			f.owner(t, func() { _ = f.m.beginCommandPalette() })
			body := core.NewBody()
			field := core.NewTextField(body)
			field.SetName("command-query")
			e := &commandPaletteKeys{a: f.a, ref: f.ref, m: f.m}
			_ = e.Sync(e.Mount(body))
			field.SetText("nudge-1-right")
			field.HandleEvent(events.NewKey(events.KeyChord, 0, key.CodeReturnEnter, mods|1<<key.Control))
			if f.m.commandDraft == nil || f.m.scene.Objects[0].X != 0 {
				t.Fatal("modified Return ran command")
			}
			field.HandleEvent(events.NewKey(events.KeyChord, 0, key.CodeReturnEnter, mods))
			if f.m.commandDraft != nil || f.m.scene.Objects[0].X != 1 || f.m.journal.Snapshot().UndoCount != 2 {
				t.Fatal("lock state blocked Return")
			}
		})
	}
}
