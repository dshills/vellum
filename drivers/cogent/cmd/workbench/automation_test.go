package main

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/automation"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

func TestWorkbenchSemanticDomainAndVirtualQueries(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Mac endpoint")
	}
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	source := newParts(100000)
	table, err := virtual.NewTable(source, 20)
	if err != nil {
		t.Fatal(err)
	}
	journal, _ := history.New(16)
	edit := &partEditor{source: source, table: table, journal: journal}
	refs := &fileReferences{paths: []string{"/private/secret/project.vellum"}}
	selected := 0
	selectionRevision := uint64(1)
	selectID := func(id string) error {
		index, err := componentIndex(source, id)
		if err != nil {
			return err
		}
		if index != selected {
			selectionRevision++
			selected = index
		}
		return nil
	}
	commands := command.New()
	commands.Register(command.Command{ID: "component.rotate", Label: "Rotate", Execute: func(context.Context) error {
		next := source.parts.at(selected)
		next.Rotation = (next.Rotation + 90) % 360
		return edit.change(selected, "Rotate", next)
	}})
	commands.Register(command.Command{ID: "component.raise-value", Label: "Raise", Execute: func(context.Context) error {
		next := source.parts.at(selected)
		if next.Value > math.MaxUint32-100 {
			return errors.New("overflow")
		}
		next.Value += 100
		return edit.change(selected, "Raise", next)
	}})
	commands.Register(command.Command{ID: "edit.undo", Label: "Undo", Available: func() error {
		if journal.Snapshot().UndoCount == 0 {
			return history.ErrEmpty
		}
		return nil
	}, Execute: func(context.Context) error { return journal.Undo() }})
	commands.Register(command.Command{ID: "edit.redo", Label: "Redo", Available: func() error {
		if journal.Snapshot().RedoCount == 0 {
			return history.ErrEmpty
		}
		return nil
	}, Execute: func(context.Context) error { return journal.Redo() }})
	r, err := workbenchAutomation(a, commands, source, refs, func() (int, uint64) { return selected, selectionRevision }, selectID)
	if err != nil {
		t.Fatal(err)
	}
	s, err := automation.Start(automation.Config{Owner: a, Window: "main", Registry: r})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	wake := make(chan struct{}, 1)
	a.SetWake(func() {
		select {
		case wake <- struct{}{}:
		default:
		}
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-a.Context().Done():
				return
			case <-wake:
				a.Drain()
			}
		}
	}()
	defer func() {
		a.Close()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("owner teardown")
		}
		select {
		case <-s.Done():
		case <-time.After(5 * time.Second):
			t.Error("endpoint teardown")
		}
	}()
	call := func(op, target, params string, rev *uint64) automation.Response {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		v, err := automation.Call(ctx, s.Endpoint(), automation.Request{Version: 1, ID: "domain", Operation: op, Target: target, ExpectedRevision: rev, Params: json.RawMessage(params)})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	code := func(v automation.Response, want string) {
		t.Helper()
		if v.Error == nil || v.Error.Code != want {
			t.Fatalf("want %s: %+v", want, v)
		}
	}
	first := call("inspect", "workbench", `{}`, nil)
	if first.Error != nil || strings.Contains(string(first.Result), "secret") || !strings.Contains(string(first.Result), `"document_path":"[redacted]"`) {
		t.Fatal(first)
	}
	code(call("invoke", "edit.undo", `{}`, first.Revision), "rejected")
	virtual := call("inspect", "component", `{"id":"component:100000"}`, nil)
	if virtual.Error != nil || !strings.Contains(string(virtual.Result), `"reference":"R0100000"`) {
		t.Fatal(virtual)
	}
	code(call("inspect", "component", `{"id":"component:0100000"}`, nil), "rejected")
	selectedResponse := call("invoke", "component.select", `{"id":"component:100000"}`, first.Revision)
	if selectedResponse.Error != nil {
		t.Fatal(selectedResponse)
	}
	rotated := call("invoke", "component.rotate", `{}`, selectedResponse.Revision)
	if rotated.Error != nil {
		t.Fatal(rotated)
	}
	code(call("invoke", "component.raise-value", `{}`, selectedResponse.Revision), "stale_revision")
	undone := call("invoke", "edit.undo", `{}`, rotated.Revision)
	if undone.Error != nil {
		t.Fatal(undone)
	}
	redone := call("invoke", "edit.redo", `{}`, undone.Revision)
	if redone.Error != nil {
		t.Fatal(redone)
	}
	code(call("invoke", "component.rotate", `{}`, rotated.Revision), "stale_revision") // Domain ABA.
	a.Do(func() { selectID("component:1"); selectID("component:100000") })
	code(call("invoke", "component.rotate", `{}`, redone.Revision), "stale_revision") // Selection ABA.
	beforeRef := call("inspect", "workbench", `{}`, nil)
	a.Do(func() {
		if err := refs.replace(journal, []string{"/other/private"}); err != nil {
			t.Error(err)
		}
		if err := journal.Undo(); err != nil {
			t.Error(err)
		}
	})
	code(call("invoke", "component.rotate", `{}`, beforeRef.Revision), "stale_revision") // Reference ABA.
	beforeUI := call("inspect", "component", `{}`, nil)
	a.Do(func() {
		if err := commands.Invoke(a.Context(), "component.raise-value"); err != nil {
			t.Error(err)
		}
	})
	code(call("invoke", "component.rotate", `{}`, beforeUI.Revision), "stale_revision")
	code(call("invoke", "file.save", `{}`, call("inspect", "workbench", `{}`, nil).Revision), "unknown_target")
}

func TestWorkbenchEpochExhaustionPreservesDomain(t *testing.T) {
	source := newParts(3)
	table, _ := virtual.NewTable(source, 2)
	journal, _ := history.New(4)
	e := &partEditor{source: source, table: table, journal: journal}
	source.revision = math.MaxUint64
	before := source.parts.at(0)
	next := before
	next.Value++
	if err := e.change(0, "overflow", next); !errors.Is(err, history.ErrRevisionExhausted) {
		t.Fatal(err)
	}
	if source.parts.at(0) != before || journal.Snapshot().UndoCount != 0 {
		t.Fatal("mutation before epoch refusal")
	}
	original := filepath.Join(t.TempDir(), "original")
	refs := &fileReferences{paths: []string{original}, revision: math.MaxUint64}
	if err := refs.replace(journal, []string{filepath.Join(t.TempDir(), "replacement")}); !errors.Is(err, history.ErrRevisionExhausted) {
		t.Fatal(err)
	}
	if refs.paths[0] != original || journal.Snapshot().UndoCount != 0 {
		t.Fatal("reference mutation before refusal")
	}
}
