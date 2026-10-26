package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"image/color"
	"strings"
	"unicode/utf8"
)

type CommandPaletteItem struct {
	ID, Label, Group, Shortcut, Keywords string
	Icon                                 Symbol
	IconImage                            *canvas.TiledImage
	Disabled                             bool
	Available                            func() error
}
type CommandPaletteEvent struct{ ID, CommandID string }
type CommandPaletteResult struct {
	Item     CommandPaletteItem
	Disabled bool
	Reason   string
}
type PaletteKeyEvent struct {
	Text   string
	Delta  int
	Accept bool
}

// PaletteInput is immutable native routing metadata emitted only by View.
// Drivers call Change on the application owner; never inspect the source model.
type PaletteInput struct {
	ID     string
	Change func(PaletteKeyEvent)
	source *CommandPalette
}

func (p *PaletteInput) valid(id string) bool {
	return p != nil && p.source != nil && p.ID == id && p.Change != nil
}

type paletteEpoch struct{ marker byte }

// CommandPalette is a retained, owner-only bounded search panel. The application
// opens it by setting Open; it owns query/selection and closes before execution.
// Availability is pure/fast and rechecked for execution. Invoke normally calls
// the application's command.Registry.Invoke so domain/history validation stays
// there. Execution errors close the panel and go to OnError/Message for app
// presentation; they do not automatically reopen or show an in-panel error.
// Dismiss is an optional close notification. No action, I/O, timer or goroutine
// runs from View except availability checks. One live binding is supported.
type CommandPalette struct {
	ID, Label, Query, SelectedID, ReturnFocusID string
	Open, Disabled, Dark                        bool
	Items                                       []CommandPaletteItem
	MaxResults                                  int
	Invoke                                      func(CommandPaletteEvent) error
	Dismiss                                     func()
	OnError                                     func(error)
	Style                                       Style
	epoch                                       *paletteEpoch
	identity                                    string
	message                                     string
	browsing                                    bool
	browseStart                                 int
	sheet                                       Sheet
}

func CommandPaletteInputID(id string) string { return SearchInputID(id + "/palette/query") }
func CommandPaletteItemID(id, item string) string {
	return id + "/palette/item/" + base64.RawURLEncoding.EncodeToString([]byte(item))
}
func (p *CommandPalette) Message() string {
	if p == nil {
		return ""
	}
	return p.message
}
func paletteText(s string, n int, empty bool) bool {
	return len(s) <= n && utf8.ValidString(s) && !strings.ContainsRune(s, 0) && (empty || strings.TrimSpace(s) != "")
}
func (p *CommandPalette) validate() error {
	if p == nil || !paletteText(p.ID, 128, false) || len(p.Items) > 256 || p.MaxResults < 0 || p.MaxResults > 50 || !paletteText(p.Query, 1024, true) {
		return fmt.Errorf("ui: invalid command palette configuration")
	}
	seen := map[string]bool{}
	for _, v := range p.Items {
		if !paletteText(v.ID, 128, false) || seen[v.ID] || !paletteText(v.Label, 256, false) || !paletteText(v.Group, 128, true) || !paletteText(v.Shortcut, 64, true) || !paletteText(v.Keywords, 1024, true) {
			return fmt.Errorf("ui: invalid or duplicate palette item")
		}
		seen[v.ID] = true
	}
	return nil
}

// Results returns at most MaxResults (default20, maximum50), grouped in first
// group order and then item order. Search terms must all match metadata. Refine
// the query to reach items beyond the bound. Call only on the owner.
func (p *CommandPalette) Results() ([]CommandPaletteResult, error) {
	if e := p.validate(); e != nil {
		return nil, e
	}
	limit := p.MaxResults
	if limit == 0 {
		limit = 20
	}
	groups := []string{}
	buckets := map[string][]CommandPaletteItem{}
	terms := strings.Fields(strings.ToLower(p.Query))
	for _, v := range p.Items {
		hay := strings.ToLower(v.ID + " " + v.Label + " " + v.Group + " " + v.Keywords)
		matches := true
		for _, term := range terms {
			matches = matches && strings.Contains(hay, term)
		}
		if !matches {
			continue
		}
		if _, ok := buckets[v.Group]; !ok {
			groups = append(groups, v.Group)
		}
		buckets[v.Group] = append(buckets[v.Group], v)
	}
	out := []CommandPaletteResult{}
	for _, g := range groups {
		for _, v := range buckets[g] {
			row := CommandPaletteResult{Item: v, Disabled: p.Disabled || v.Disabled || p.Invoke == nil}
			if v.Available != nil {
				if e := v.Available(); e != nil {
					row.Disabled = true
					row.Reason = boundedWidgetMessage(e.Error())
				}
			}
			out = append(out, row)
			if len(out) == limit {
				return out, nil
			}
		}
	}
	return out, nil
}
func (p *CommandPalette) selected(rows []CommandPaletteResult) string {
	if p.browsing {
		start := max(0, min(p.browseStart, len(rows)))
		rows = rows[start:min(len(rows), start+5)]
	}
	for _, r := range rows {
		if r.Item.ID == p.SelectedID && !r.Disabled {
			return r.Item.ID
		}
	}
	for _, r := range rows {
		if !r.Disabled {
			return r.Item.ID
		}
	}
	return ""
}
func (p *CommandPalette) setQuery(q string) {
	if !paletteText(q, 1024, true) {
		p.message = "Search is limited to 1024 UTF-8 bytes"
		return
	}
	if p.Query != q {
		p.Query = q
		p.SelectedID = ""
		p.browsing = false
	}
	p.message = ""
}
func (p *CommandPalette) move(delta int) {
	if delta == 0 {
		return
	}
	direction := 1
	if delta < 0 {
		direction = -1
	}
	rows, err := p.Results()
	if err != nil || len(rows) == 0 {
		return
	}
	current := -1
	selected := p.selected(rows)
	for i, row := range rows {
		if row.Item.ID == selected && !row.Disabled {
			current = i
			break
		}
	}
	if current < 0 {
		if p.browsing {
			current = p.browseStart
			if direction > 0 {
				current = min(len(rows), p.browseStart+5) - 1
			}
		} else if direction < 0 {
			current = 0
		}
	}
	for step := 1; step <= len(rows); step++ {
		index := (current + direction*step) % len(rows)
		if index < 0 {
			index += len(rows)
		}
		if !rows[index].Disabled {
			p.SelectedID = rows[index].Item.ID
			p.browsing = false
			return
		}
	}
	p.SelectedID = ""
}
func (p *CommandPalette) close() {
	p.Open = false
	p.epoch = nil
	if p.Dismiss != nil {
		p.Dismiss()
	}
}
func (p *CommandPalette) run(id string) {
	rows, e := p.Results()
	if e != nil {
		p.message = boundedWidgetMessage(e.Error())
		return
	}
	for _, row := range rows {
		if row.Item.ID != id {
			continue
		}
		if row.Disabled {
			p.message = row.Reason
			if p.message == "" {
				p.message = "Command is unavailable"
			}
			return
		}
		fn := p.Invoke
		if fn == nil {
			return
		}
		event := CommandPaletteEvent{ID: p.ID, CommandID: id}
		p.close()
		if err := fn(event); err != nil {
			p.message = boundedWidgetMessage(err.Error())
			if p.OnError != nil {
				p.OnError(err)
			}
		}
		return
	}
	p.message = "Choose a matching command"
}
func (p *CommandPalette) View() (Node, error) {
	if e := p.validate(); e != nil {
		return Node{}, e
	}
	if p.identity != p.ID || !p.Open || p.Disabled {
		p.epoch = nil
		p.identity = p.ID
	}
	if p.Open && !p.Disabled && p.epoch == nil {
		p.epoch = &paletteEpoch{}
		p.browsing = false
		p.browseStart = 0
	}
	id, epoch := p.ID, p.epoch
	live := func() bool { return p.ID == id && p.Open && !p.Disabled && p.epoch == epoch && epoch != nil }
	rows := []CommandPaletteResult{}
	if p.Open {
		var err error
		rows, err = p.Results()
		if err != nil {
			return Node{}, err
		}
	}
	selected := p.selected(rows)
	search := &SearchBox{ID: id + "/palette/query", Label: "Search commands", Text: p.Query, Placeholder: "Type a command…", Change: func(e SearchEvent) {
		if live() {
			p.setQuery(e.Text)
		}
	}}
	query, e := search.View()
	if e != nil {
		return Node{}, e
	}
	for i := range query.Children {
		if query.Children[i].ID == CommandPaletteInputID(id) {
			query.Children[i].PaletteInput = &PaletteInput{ID: CommandPaletteInputID(id), source: p, Change: func(e PaletteKeyEvent) {
				if !live() {
					return
				}
				p.setQuery(e.Text)
				if p.message != "" {
					return
				}
				if e.Delta == 1 || e.Delta == -1 {
					p.move(e.Delta)
				}
				if e.Accept {
					rs, err := p.Results()
					if err == nil {
						if selected := p.selected(rs); selected != "" {
							p.run(selected)
						}
					}
				}
			}}
		}
	}
	pal := MacPalette(p.Dark, color.NRGBA{})
	children := []Node{query}
	index := 0
	for i, row := range rows {
		if row.Item.ID == selected {
			index = i
		}
	}
	// A bounded moving five-row window keeps the active item visible without
	// stealing typing focus. Previous/Next supports pointer-only browsing too.
	start := max(0, min(index-2, len(rows)-5))
	if p.browsing {
		start = max(0, min(p.browseStart, len(rows)-5))
	}
	end := min(len(rows), start+5)
	group := ""
	haveGroup := false
	for _, row := range rows[start:end] {
		item := row.Item
		if !haveGroup || group != item.Group {
			group = item.Group
			haveGroup = true
			if group != "" {
				children = append(children, Node{ID: id + "/palette/group/" + base64.RawURLEncoding.EncodeToString([]byte(group)), Kind: Label, Text: group, Style: Style{Typography: Some(MacTypography(11, 600)), Foreground: Some(pal.Secondary)}})
			}
		}
		caption := item.Label
		if item.Shortcut != "" {
			caption += "    " + item.Shortcut
		}
		if row.Disabled {
			caption += " · Unavailable"
		}
		style := Style{GrowX: Some(float32(1)), TextAlign: Some(AlignStart), Padding: Some(Insets{Top: 7, Right: 9, Bottom: 7, Left: 9}), Border: Some(Border{}), Background: Some(pal.Popup), Foreground: Some(pal.Ink)}
		if item.ID == selected {
			style.Background = Some(pal.Accent)
			style.Foreground = Some(pal.AccentInk)
		}
		children = append(children, Node{ID: CommandPaletteItemID(id, item.ID), Kind: Button, Text: caption, Icon: item.Icon, IconImage: item.IconImage, Appearance: ButtonTool, Disabled: row.Disabled, Style: style, Accessibility: Accessibility{Name: item.Label, Help: row.Reason}, Activate: func() {
			if live() {
				p.run(item.ID)
			}
		}})
	}
	caption := "No matching commands"
	if len(rows) > 0 {
		caption = fmt.Sprintf("%d matches · showing %d–%d · ↑↓ to choose", len(rows), start+1, end)
	}
	children = append(children, Node{ID: id + "/palette/count", Kind: Label, Text: caption, Style: Style{Foreground: Some(pal.Secondary)}})
	if len(rows) > 5 {
		children = append(children, Node{ID: id + "/palette/pages", Kind: Row, Children: []Node{{ID: id + "/palette/previous", Kind: Button, Text: "Previous", Disabled: start == 0, Activate: func() {
			if live() {
				p.page(-1)
			}
		}}, {ID: id + "/palette/next", Kind: Button, Text: "Next", Disabled: end == len(rows), Activate: func() {
			if live() {
				p.page(1)
			}
		}}}})
	}
	if p.message != "" {
		children = append(children, Node{ID: id + "/palette/message", Kind: Label, Text: p.message})
	}
	title := p.Label
	if title == "" {
		title = "Search Commands"
	}
	p.sheet.ID = id
	p.sheet.Title = title
	p.sheet.Open = p.Open
	p.sheet.Disabled = p.Disabled
	p.sheet.ConfirmLabel = "Run"
	p.sheet.CancelLabel = "Cancel"
	p.sheet.ReturnFocusID = p.ReturnFocusID
	p.sheet.Children = children
	p.sheet.ConfirmDisabled = selected == ""
	p.sheet.Confirm = func() {
		if live() {
			rs, err := p.Results()
			if err == nil {
				if selected := p.selected(rs); selected != "" {
					p.run(selected)
				}
			}
		}
	}
	p.sheet.Cancel = func() {
		if live() {
			p.close()
		}
	}
	p.sheet.Style = Style{Background: Some(pal.Popup), Foreground: Some(pal.Ink), MinSize: Some(Size{Width: 420}), MaxSize: Some(Size{Width: 560}), Typography: Some(MacTypography(13, 400)), CornerRadius: Some(float32(12))}.Overlay(p.Style)
	return p.sheet.View()
}

// Pointer paging clamps by displayed result rows, including unavailable rows.
// It never wraps through enabled-item indices or hides an unavailable-only page.
func (p *CommandPalette) page(direction int) {
	rows, err := p.Results()
	if err != nil || len(rows) == 0 {
		return
	}
	index := 0
	selected := p.selected(rows)
	for i, row := range rows {
		if row.Item.ID == selected {
			index = i
			break
		}
	}
	start := max(0, min(index-2, len(rows)-5))
	if p.browsing {
		start = p.browseStart
	}
	start = max(0, min(start+direction*5, len(rows)-5))
	p.browseStart = start
	p.browsing = true
	p.SelectedID = ""
	for _, row := range rows[start:min(len(rows), start+5)] {
		if !row.Disabled {
			p.SelectedID = row.Item.ID
			break
		}
	}
}
