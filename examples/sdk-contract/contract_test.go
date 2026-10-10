// These external-consumer scenarios exercise public application contracts.
package sdkcontract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dshills/vellum"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func TestControlledViewAndCommandContract(t *testing.T) {
	if vellum.Version != "v0.1.0-alpha.1" {
		t.Fatal(vellum.Version)
	}
	checked := false
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Style: ui.Style{Gap: ui.Some(float32(0)), Padding: ui.Some(ui.UniformInsets(8))}, Children: []ui.Node{{ID: "enabled", Kind: ui.Checkbox, Checked: checked, Accessibility: ui.Accessibility{Name: "Enable preview"}, Change: func(e ui.ChangeEvent) { checked = e.Checked }}}}
	}
	theme := ui.Theme{Default: ui.Style{Gap: ui.Some(float32(20))}}
	if err := ui.Validate(view()); err != nil {
		t.Fatal(err)
	}
	style, err := ui.ResolveStyle(view(), theme)
	if err != nil || style.Gap.Value != 0 {
		t.Fatalf("explicit zero contract: %+v %v", style, err)
	}
	view().Children[0].Change(ui.ChangeEvent{ID: "enabled", Kind: ui.Checkbox, Checked: true})
	if !view().Children[0].Checked {
		t.Fatal("accepted model edit not reflected")
	}
	registry := command.New()
	enabled := true
	calls := 0
	registry.Register(command.Command{ID: "save", Label: "Save", Shortcut: "Control+S", Available: func() error {
		if !enabled {
			return errors.New("busy")
		}
		return nil
	}, Execute: func(context.Context) error { calls++; return nil }})
	enabled = false
	if err := registry.Invoke(context.Background(), "save"); err == nil || calls != 0 {
		t.Fatal("disabled command executed")
	}
	enabled = true
	matched, err := registry.Shortcut(context.Background(), "control+s")
	if !matched || err != nil || calls != 1 {
		t.Fatalf("shared shortcut contract %v %v %d", matched, err, calls)
	}
}
func TestWindowScopedTaskCommitContract(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "Consumer", Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label, Text: "Consumer"} }}); err != nil {
		t.Fatal(err)
	}
	lifetime, err := a.WindowContext("main")
	if err != nil {
		t.Fatal(err)
	}
	manager, err := tasks.New(a, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	var task *tasks.Task
	var startErr error
	commits := 0
	err = a.Do(func() {
		task, startErr = tasks.Start(manager, lifetime, func(context.Context, func(float64, string)) (int, error) { return 42, nil }, func(v int, err error) {
			if v != 42 || err != nil {
				t.Errorf("result %d %v", v, err)
			}
			commits++
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if startErr != nil {
		t.Fatal(startErr)
	}
	select {
	case <-task.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("worker failed to deliver")
	}
	if commits != 0 {
		t.Fatal("commit ran without owner Drain")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if commits != 1 {
		t.Fatal(commits)
	}
	if err := a.CloseWindow("main"); err != nil {
		t.Fatal(err)
	}
	if lifetime.Err() == nil || a.Context().Err() == nil {
		t.Fatal("final window lifetime not closed")
	}
}
func TestWorkspacesRoundTripFloatingLayoutsWithoutAliasing(t *testing.T) {
	session, err := workbench.NewSession(workbench.SessionLayout{Version: workbench.SessionVersion, Focus: workbench.Focus{Window: "main", Panel: "document", Serial: 1}, Windows: []workbench.WindowLayout{{ID: "main", Title: "Main", Bounds: workbench.Bounds{Width: 800, Height: 600}, Root: &workbench.Group{ID: "main-tabs", Tabs: []string{"document", "inspector"}, Active: "document"}}}}, []string{"document", "inspector"})
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Float("document", workbench.WindowLayout{ID: "floating", Title: "Document", Bounds: workbench.Bounds{Width: 640, Height: 480}, Root: &workbench.Group{ID: "floating-tabs"}}); err != nil {
		t.Fatal(err)
	}
	if err := session.Dock("document", "main", "main-tabs", 0); err != nil {
		t.Fatal(err)
	}
	encoded, err := session.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Decode(encoded); err != nil {
		t.Fatal(err)
	}
	layout := session.Layout()
	if len(layout.Windows) != 1 || layout.Windows[0].Root.Active != "document" {
		t.Fatalf("docking contract: %+v", layout)
	}
	layout.Windows[0].Root.Tabs[0] = "external mutation"
	if session.Layout().Windows[0].Root.Tabs[0] != "document" {
		t.Fatal("returned layout aliases retained state")
	}
}

func TestDocumentHistoryAndSaveRevisionThroughPublicCommands(t *testing.T) {
	journal, err := history.New(2)
	if err != nil {
		t.Fatal(err)
	}
	value := 0
	change := func(next int) history.Change {
		previous := value
		return history.Change{Label: "Property", Apply: func() error { value = next; return nil }, Revert: func() error { value = previous; return nil }}
	}
	if err = journal.Apply(change(42)); err != nil {
		t.Fatal(err)
	}
	captured := journal.Snapshot().Revision
	if err = journal.Apply(change(84)); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(journal.MarkSavedRevision(captured), history.ErrStale) || !journal.Snapshot().Dirty {
		t.Fatal("stale save cleaned document")
	}
	registry := command.New()
	if err = registry.Register(command.Command{ID: "undo", Label: "Undo", Shortcut: "Meta+Z", ShortcutScope: command.AfterFocusShortcuts, Available: func() error {
		if journal.Snapshot().UndoCount == 0 {
			return history.ErrEmpty
		}
		return nil
	}, Execute: func(context.Context) error { return journal.Undo() }}); err != nil {
		t.Fatal(err)
	}
	if matched, err := registry.ShortcutInScope(context.Background(), "Meta+Z", command.WindowShortcuts); matched || err != nil || value != 84 {
		t.Fatal("document undo preempted focused widget")
	}
	if matched, err := registry.ShortcutInScope(context.Background(), "Meta+Z", command.AfterFocusShortcuts); !matched || err != nil || value != 42 {
		t.Fatal("document fallback failed")
	}
	if err = journal.MarkSavedRevision(captured); err != nil || journal.Snapshot().Dirty {
		t.Fatal("matching save rejected")
	}
	if err = registry.Invoke(context.Background(), "undo"); err != nil || value != 0 || !journal.Snapshot().Dirty {
		t.Fatal("menu undo failed")
	}
	if err = registry.Invoke(context.Background(), "undo"); !errors.Is(err, history.ErrEmpty) {
		t.Fatal("empty history allowed")
	}
	if err = journal.Redo(); err != nil || value != 42 || journal.Snapshot().Dirty {
		t.Fatal("saved revision identity not restored")
	}
}
