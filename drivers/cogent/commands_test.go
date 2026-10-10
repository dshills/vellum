package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"context"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
	"testing"
	"time"
)

func TestMenuConstructionDoesNotReenterOwner(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	registry := command.New()
	if err := registry.Register(command.Command{ID: "test", Label: "<literal>", Execute: func(context.Context) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- a.Do(func() {
			host := core.NewBody("headless")
			bar := &CommandBar{Owner: a, Registry: registry}
			widget := bar.Mount(host)
			widget.AsWidget().Update()
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("menu builder nested owner")
	}
}

func TestDocumentFallbackLetsNativeTextUndoConsumeChord(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "main", Title: "Keys", Width: 400, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	body := core.NewBody("Keys")
	field := core.NewTextField(body)
	registry := command.New()
	calls := 0
	if err := registry.Register(command.Command{ID: "undo", Label: "Undo", Shortcut: "Control+Z", ShortcutScope: command.AfterFocusShortcuts, Execute: func(context.Context) error { calls++; return nil }}); err != nil {
		t.Fatal(err)
	}
	bindCommandShortcuts(body.Scene, a, "main", registry, nil)
	event := events.NewKey(events.KeyChord, 'z', key.CodeZ, key.Modifiers(1<<key.Control))
	body.Scene.Listeners.First.Call(event)
	if calls != 0 || event.IsHandled() {
		t.Fatal("document intercepted native editing")
	}
	field.HandleEvent(event)
	body.Scene.Listeners.Final.Call(event)
	if !event.IsHandled() || calls != 0 {
		t.Fatal("native Undo fell through", calls)
	}
	event = events.NewKey(events.KeyChord, 'z', key.CodeZ, key.Modifiers(1<<key.Control))
	body.Scene.Listeners.First.Call(event)
	body.Scene.Listeners.Final.Call(event)
	if !event.IsHandled() || calls != 1 {
		t.Fatal("document fallback not invoked", calls)
	}
	if err := a.CloseWindow("main"); err != nil {
		t.Fatal(err)
	}
	event = events.NewKey(events.KeyChord, 'z', key.CodeZ, key.Modifiers(1<<key.Control))
	body.Scene.Listeners.Final.Call(event)
	if calls != 1 {
		t.Fatal("closed window invoked document")
	}
}
