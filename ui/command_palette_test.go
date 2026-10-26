package ui

import (
	"errors"
	"fmt"
	"testing"
)

func paletteNode(n Node, id string) *Node {
	if n.ID == id {
		return &n
	}
	for _, c := range n.Children {
		if found := paletteNode(c, id); found != nil {
			return found
		}
	}
	return nil
}
func TestCommandPaletteSearchNavigationAndLiveExecution(t *testing.T) {
	allowed, calls := true, 0
	p := &CommandPalette{ID: "commands", Open: true, Items: []CommandPaletteItem{{ID: "new", Label: "New drawing", Group: "File", Keywords: "create"}, {ID: "save", Label: "Save drawing", Group: "File", Available: func() error {
		if !allowed {
			return errors.New("Unavailable now")
		}
		return nil
	}}, {ID: "blocked", Label: "Blocked", Disabled: true}}, Invoke: func(e CommandPaletteEvent) error {
		calls++
		if e.CommandID != "save" {
			t.Fatal(e)
		}
		return nil
	}}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	p.move(1)
	if p.SelectedID != "save" {
		t.Fatal(p.SelectedID)
	}
	old := paletteNode(n, CommandPaletteItemID(p.ID, "save"))
	allowed = false
	old.Activate()
	if calls != 0 || !p.Open {
		t.Fatal("stale availability executed")
	}
	allowed = true
	old.Activate()
	if calls != 1 || p.Open {
		t.Fatal("execute/close", calls, p.Open)
	}
	old.Activate()
	if calls != 1 {
		t.Fatal("closed callback executed")
	}
	p.View()
	p.Open = true
	p.View()
	old.Activate()
	if calls != 1 {
		t.Fatal("prior open epoch executed")
	}
	p.Query = "create drawing"
	rows, e := p.Results()
	if e != nil || len(rows) != 1 || rows[0].Item.ID != "new" {
		t.Fatal(rows, e)
	}
	p.Query = "missing"
	rows, e = p.Results()
	if e != nil || len(rows) != 0 {
		t.Fatal(rows, e)
	}
}
func TestCommandPaletteBoundsAndMovingVisibleWindow(t *testing.T) {
	p := &CommandPalette{ID: "many", Open: true, Invoke: func(CommandPaletteEvent) error { return nil }}
	for i := range 80 {
		p.Items = append(p.Items, CommandPaletteItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Action %d", i), Group: fmt.Sprint(i % 2)})
	}
	rows, e := p.Results()
	if e != nil || len(rows) != 20 {
		t.Fatal(len(rows), e)
	}
	p.SelectedID = rows[12].Item.ID
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	if paletteNode(n, CommandPaletteItemID(p.ID, p.SelectedID)) == nil {
		t.Fatal("selected action not projected")
	}
	p.Items = append(p.Items, p.Items[0])
	if _, e = p.View(); e == nil {
		t.Fatal("duplicate identity admitted")
	}
}

func TestCommandPalettePagingClampsAndPreservesSameQueryChoice(t *testing.T) {
	invoked := ""
	p := &CommandPalette{ID: "pages", Open: true, Query: "action", Invoke: func(e CommandPaletteEvent) error { invoked = e.CommandID; return nil }}
	for i := range 12 {
		p.Items = append(p.Items, CommandPaletteItem{ID: fmt.Sprint(i), Label: fmt.Sprint("Action ", i), Disabled: i >= 5 && i < 10})
	}
	if _, e := p.View(); e != nil {
		t.Fatal(e)
	}
	p.page(1)
	if p.browseStart != 5 || p.SelectedID != "" {
		t.Fatal("unavailable page was skipped", p.browseStart, p.SelectedID)
	}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	if paletteNode(n, CommandPaletteItemID(p.ID, "5")) == nil {
		t.Fatal("unavailable page not rendered")
	}
	p.page(1)
	if p.browseStart != 7 || p.SelectedID != "10" {
		t.Fatal("last page did not clamp", p.browseStart, p.SelectedID)
	}
	p.page(1)
	if p.browseStart != 7 {
		t.Fatal("next wrapped")
	}
	p.page(-1)
	p.page(-1)
	if p.browseStart != 0 {
		t.Fatal("previous wrapped")
	}
	p.browsing = false
	p.SelectedID = "2"
	n, e = p.View()
	if e != nil {
		t.Fatal(e)
	}
	paletteNode(n, CommandPaletteInputID(p.ID)).PaletteInput.Change(PaletteKeyEvent{Text: "action", Accept: true})
	if invoked != "2" {
		t.Fatal("unchanged query discarded selected command", invoked)
	}
	p.View()
	p.Open = true
	p.Query = "no matches"
	n, e = p.View()
	if e != nil {
		t.Fatal(e)
	}
	paletteNode(n, CommandPaletteInputID(p.ID)).PaletteInput.Change(PaletteKeyEvent{Text: p.Query, Accept: true})
	if !p.Open || p.Message() != "" {
		t.Fatal("empty Return should stay inert", p.Message())
	}
}

func TestCommandPaletteArrowsFromUnavailablePageAndExecutionError(t *testing.T) {
	p := &CommandPalette{ID: "arrows", Open: true, Invoke: func(CommandPaletteEvent) error { return nil }}
	for i := range 12 {
		p.Items = append(p.Items, CommandPaletteItem{ID: fmt.Sprint(i), Label: fmt.Sprint(i), Disabled: i >= 5 && i < 10})
	}
	p.View()
	p.page(1)
	p.move(1)
	if p.SelectedID != "10" {
		t.Fatal("Down ignored unavailable page", p.SelectedID)
	}
	p.browseStart = 5
	p.browsing = true
	p.SelectedID = ""
	p.move(-1)
	if p.SelectedID != "4" {
		t.Fatal("Up ignored unavailable page", p.SelectedID)
	}
	notified := 0
	p.Invoke = func(CommandPaletteEvent) error { return errors.New("Synthetic execution error") }
	p.OnError = func(error) {
		notified++
		if p.Open {
			t.Fatal("execution error unexpectedly reopened panel")
		}
	}
	p.run("4")
	if notified != 1 || p.Open || p.Message() != "Synthetic execution error" {
		t.Fatal("closed-panel error reporting", notified, p.Message())
	}
}
