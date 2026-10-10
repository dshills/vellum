// history-smoke checks document and native text Undo/Redo precedence on macOS.
// It uses real native events and AppKit accessibility without changing settings.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

func main() {
	a, err := app.New(app.Config{})
	must(err)
	defer a.Close()
	journal, err := history.New(8)
	must(err)
	value := 0
	text := ""
	commandErrors := 0
	probeCalls := 0
	must(journal.Apply(history.Change{Label: "Set document value", Apply: func() error { value = 42; return nil }, Revert: func() error { value = 0; return nil }}))
	registry := command.New()
	for _, op := range []struct {
		id, chord string
		redo      bool
	}{{"undo", "Meta+Z", false}, {"redo", "Meta+Shift+Z", true}} {
		op := op
		must(registry.Register(command.Command{ID: op.id, Label: op.id, Shortcut: op.chord, ShortcutScope: command.AfterFocusShortcuts, Available: func() error {
			s := journal.Snapshot()
			if (!op.redo && s.UndoCount == 0) || (op.redo && s.RedoCount == 0) {
				return history.ErrEmpty
			}
			return nil
		}, Execute: func(context.Context) error {
			if op.redo {
				return journal.Redo()
			}
			return journal.Undo()
		}}))
	}
	must(registry.Register(command.Command{ID: "window.probe", Label: "Probe window shortcut", Shortcut: "Meta+K", Execute: func(context.Context) error { probeCalls++; return nil }}))
	must(a.Open(app.Window{ID: "history", Title: "Vellum document history smoke", Width: 640, Height: 320, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "text", Sequence: 1}, Children: []ui.Node{
			{ID: "text", Kind: ui.TextField, Text: text, Accessibility: ui.Accessibility{Name: "Text undo fixture"}, Change: func(e ui.ChangeEvent) { text = e.Text }},
			{ID: "document", Kind: ui.Button, Text: fmt.Sprintf("Document value %d", value)},
		}}
	}}))
	must(a.Open(app.Window{ID: "focusless", Title: "Vellum focus-free history smoke", Width: 440, Height: 220, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Label, Text: fmt.Sprintf("Document value %d", value)}
	}}))
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	outcome := make(chan error, 1)
	go func() {
		result := check(ctx, a, journal, &value, &text, &commandErrors, &probeCalls)
		outcome <- result
		a.Close()
	}()
	runErr := a.Run(ctx, cogent.Driver{Commands: registry, OnError: func(error) { commandErrors++ }})
	select {
	case err = <-outcome:
	case <-ctx.Done():
		err = ctx.Err()
	}
	if runErr != nil {
		err = errors.Join(runErr, err)
	}
	must(err)
	fmt.Println("PASS macOS native text Undo/Redo precedence, document fallback, focus-free window routing, dirty state and disabled command refusal")
}

func waitFocus(ctx context.Context, id string) error {
	for {
		inspection, err := cogent.InspectAccessibility(ctx, "history")
		if err == nil && inspection.Reachable {
			for _, element := range inspection.Elements {
				if element.DomainID == id && inspection.FocusedID == element.ID && element.Frame[2] > 0 {
					return nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for native focus %s: %w", id, ctx.Err())
		case <-time.After(15 * time.Millisecond):
		}
	}
}
func focus(ctx context.Context, id string) error {
	ok, err := cogent.PerformAccessibilityAction(ctx, "history", id, "focus", "")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("focus refused %s", id)
	}
	return waitFocus(ctx, id)
}
func check(ctx context.Context, a *app.Application, journal *history.Journal, value *int, text *string, commandErrors *int, probeCalls *int) error {
	if err := waitFocus(ctx, "text"); err != nil {
		return err
	}
	native := core.TheApp.WindowByName("Vellum document history smoke")
	if native == nil {
		return errors.New("native window missing")
	}
	// Tab completion is the ordered-event barrier; focus loss commits the buffer.
	native.Events().KeyChord('x', key.CodeX, 0)
	native.Events().KeyChord('z', key.CodeZ, key.Modifiers(1<<key.Meta))
	native.Events().KeyChord('\t', key.CodeTab, 0)
	if err := waitFocus(ctx, "document"); err != nil {
		return err
	}
	var result error
	if err := a.Do(func() {
		if *value != 42 || *text != "" || journal.Snapshot().UndoCount != 1 {
			result = fmt.Errorf("text Undo stole document history or failed: value=%d text=%q state=%+v", *value, *text, journal.Snapshot())
		}
	}); err != nil {
		return err
	}
	if result != nil {
		return result
	}
	if err := focus(ctx, "text"); err != nil {
		return err
	}
	native.Events().KeyChord('x', key.CodeX, 0)
	native.Events().KeyChord('z', key.CodeZ, key.Modifiers(1<<key.Meta))
	native.Events().KeyChord('Z', key.CodeZ, key.Modifiers(1<<key.Meta|1<<key.Shift))
	native.Events().KeyChord('\t', key.CodeTab, 0)
	if err := waitFocus(ctx, "document"); err != nil {
		return err
	}
	if err := a.Do(func() {
		if *value != 42 || *text != "x" || journal.Snapshot().UndoCount != 1 {
			result = fmt.Errorf("text Redo stole document history or failed: value=%d text=%q", *value, *text)
		}
	}); err != nil {
		return err
	}
	if result != nil {
		return result
	}
	// Document actions now run as the focused button declines these chords.
	native.Events().KeyChord('z', key.CodeZ, key.Modifiers(1<<key.Meta))
	if err := waitState(ctx, a, func() bool { return *value == 0 && !journal.Snapshot().Dirty }); err != nil {
		return fmt.Errorf("document Undo: %w", err)
	}
	native.Events().KeyChord('Z', key.CodeZ, key.Modifiers(1<<key.Meta|1<<key.Shift))
	if err := waitState(ctx, a, func() bool { return *value == 42 && journal.Snapshot().Dirty }); err != nil {
		return fmt.Errorf("document Redo: %w", err)
	}
	native.Events().KeyChord('Z', key.CodeZ, key.Modifiers(1<<key.Meta|1<<key.Shift))
	if err := waitState(ctx, a, func() bool { return *commandErrors == 1 && *value == 42 && journal.Snapshot().UndoCount == 1 }); err != nil {
		return fmt.Errorf("disabled Redo refusal: %w", err)
	}
	// A second window contains a label only, with no managed focusable control.
	for {
		inspection, err := cogent.InspectAccessibility(ctx, "focusless")
		ready := false
		if err == nil && inspection.Reachable && inspection.FocusedID == "" {
			for _, element := range inspection.Elements {
				if element.DomainID == "root" && element.Frame[2] > 0 {
					ready = true
				}
			}
		}
		if ready {
			break
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("focus-free native window not ready: %w", ctx.Err())
		case <-time.After(15 * time.Millisecond):
		}
	}
	second := core.TheApp.WindowByName("Vellum focus-free history smoke")
	if second == nil {
		return errors.New("focus-free native window missing")
	}
	second.Events().KeyChord('k', key.CodeK, key.Modifiers(1<<key.Meta))
	if err := waitState(ctx, a, func() bool { return *probeCalls == 1 }); err != nil {
		return fmt.Errorf("focus-free window shortcut: %w", err)
	}
	second.Events().KeyChord('z', key.CodeZ, key.Modifiers(1<<key.Meta))
	if err := waitState(ctx, a, func() bool { return *value == 0 && !journal.Snapshot().Dirty }); err != nil {
		return fmt.Errorf("focus-free document Undo: %w", err)
	}
	second.Events().KeyChord('Z', key.CodeZ, key.Modifiers(1<<key.Meta|1<<key.Shift))
	if err := waitState(ctx, a, func() bool { return *value == 42 && journal.Snapshot().Dirty && *commandErrors == 1 }); err != nil {
		return fmt.Errorf("focus-free document Redo: %w", err)
	}
	return nil
}
func waitState(ctx context.Context, a *app.Application, predicate func() bool) error {
	for {
		ok := false
		if err := a.Do(func() { ok = predicate() }); err != nil {
			return err
		}
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(15 * time.Millisecond):
		}
	}
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
