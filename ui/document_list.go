package ui

import (
	"encoding/base64"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type DocumentItem struct {
	ID, Name, Kind   string
	Modified         time.Time
	SizeBytes        int64
	Folder, Disabled bool
	Icon             Symbol
	Image            *canvas.TiledImage
}
type DocumentColumn struct {
	ID, Title string
	Width     float32
}
type DocumentSort struct {
	ColumnID   string
	Descending bool
}
type DocumentSortEvent struct {
	ID   string
	Sort DocumentSort
}
type DocumentResizeEvent struct {
	ID, ColumnID string
	Width        float32
}
type DocumentEvent struct{ ID, ItemID string }

// DocumentListLayout is detached immutable presentation data produced by View.
// Columns, Rows and cell slices must not be mutated after publication.
type DocumentListLayout struct {
	source   *DocumentList
	Columns  []DocumentColumn
	Rows     []DocumentRow
	Activate func(DocumentEvent)
	Preview  func(DocumentEvent)
}
type DocumentRow struct {
	ID    string
	Cells []string
	Icon  Symbol
	Image *canvas.TiledImage
}

// ColumnResizeHandle is immutable input metadata produced by DocumentList.View.
type ColumnResizeHandle struct {
	source *DocumentList
	column string
	Width  float32
	Change func(float32)
}

// DocumentList is a bounded Finder-style list over caller-supplied metadata.
// It never enumerates the filesystem. Retain one instance per owner; accept sort,
// width and selection proposals by updating the corresponding fields.
type DocumentList struct {
	ID, Label   string
	Items       []DocumentItem
	Columns     []DocumentColumn
	Sort        DocumentSort
	SelectedIDs []string
	Disabled    bool
	Change      func(ListSelectionEvent)
	SortChange  func(DocumentSortEvent)
	Resize      func(DocumentResizeEvent)
	Activate    func(DocumentEvent)
	Preview     func(DocumentEvent)
	Style       Style
}

func DefaultDocumentColumns() []DocumentColumn {
	return []DocumentColumn{{"name", "Name", 220}, {"kind", "Kind", 100}, {"modified", "Date Modified", 140}, {"size", "Size", 80}}
}
func DocumentListID(id string) string { return id + "/documents/list" }
func DocumentHeaderID(id, column string) string {
	return id + "/documents/header/" + base64.RawURLEncoding.EncodeToString([]byte(column))
}
func documentColumn(id string) bool {
	return id == "name" || id == "kind" || id == "modified" || id == "size"
}
func (d *DocumentList) columns() []DocumentColumn {
	if len(d.Columns) == 0 {
		return DefaultDocumentColumns()
	}
	return slices.Clone(d.Columns)
}
func (d *DocumentList) validate() error {
	if d == nil || !validTokenText(d.ID) || !validTokenText(d.Label) || len(d.Items) > MaxListBoxOptions {
		return fmt.Errorf("ui: document list needs ID/label and <=256 items")
	}
	cols := d.columns()
	seen := map[string]bool{}
	if len(cols) < 1 || len(cols) > 4 {
		return fmt.Errorf("ui: document list needs 1..4 columns")
	}
	for _, c := range cols {
		if !documentColumn(c.ID) || seen[c.ID] || !validTokenText(c.Title) || !finite(c.Width) || c.Width < 60 || c.Width > 600 {
			return fmt.Errorf("ui: invalid document column")
		}
		seen[c.ID] = true
	}
	if d.Sort.ColumnID != "" && !seen[d.Sort.ColumnID] {
		return fmt.Errorf("ui: absent document sort column")
	}
	seen = map[string]bool{}
	for _, i := range d.Items {
		if !validTokenText(i.ID) || !validTokenText(i.Name) || len(i.Name) > 512 || len(i.Kind) > 256 || !utf8.ValidString(i.Kind) || strings.ContainsAny(i.Name+i.Kind, "\x00\r\n") || i.SizeBytes < 0 || seen[i.ID] {
			return fmt.Errorf("ui: invalid document item")
		}
		if err := validateIcon(Node{Kind: Button, Text: i.Name, Icon: i.Icon, IconImage: i.Image}); err != nil {
			return err
		}
		seen[i.ID] = true
	}
	selected := map[string]bool{}
	for _, id := range d.SelectedIDs {
		if !seen[id] || selected[id] {
			return fmt.Errorf("ui: absent or duplicate document selection")
		}
		selected[id] = true
	}
	return nil
}
func (d *DocumentList) View() (Node, error) {
	if err := d.validate(); err != nil {
		return Node{}, err
	}
	id := d.ID
	cols := d.columns()
	items := slices.Clone(d.Items)
	live := func() bool { return d.ID == id && !d.Disabled }
	eligible := func(target string) bool {
		if !live() {
			return false
		}
		for _, i := range d.Items {
			if i.ID == target {
				return !i.Disabled
			}
		}
		return false
	}
	sort := d.Sort
	if sort.ColumnID != "" {
		slices.SortStableFunc(items, func(a, b DocumentItem) int {
			if a.Folder != b.Folder {
				if a.Folder {
					return -1
				}
				return 1
			}
			result := 0
			switch sort.ColumnID {
			case "name":
				result = strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
			case "kind":
				result = strings.Compare(strings.ToLower(a.Kind), strings.ToLower(b.Kind))
			case "modified":
				result = a.Modified.Compare(b.Modified)
			case "size":
				if a.SizeBytes < b.SizeBytes {
					result = -1
				} else if a.SizeBytes > b.SizeBytes {
					result = 1
				}
			}
			if sort.Descending {
				return -result
			}
			return result
		})
	}
	layout := &DocumentListLayout{source: d, Columns: cols}
	if d.Preview != nil {
		layout.Preview = func(e DocumentEvent) {
			if eligible(e.ItemID) && d.Preview != nil {
				d.Preview(DocumentEvent{ID: id, ItemID: e.ItemID})
			}
		}
	}
	list := Node{ID: DocumentListID(id), Kind: ListBox, Multiple: true, SelectedIDs: slices.Clone(d.SelectedIDs), Disabled: d.Disabled, Accessibility: Accessibility{Name: d.Label, Help: "Document list. Command-click toggles; Shift extends; Return opens."}, Style: Style{Padding: Some(Insets{}), GrowX: Some(float32(0)), MinSize: Some(Size{Height: 160}), MaxSize: Some(Size{Height: 320})}}
	total := float32(0)
	headers := []Node{}
	for _, column := range cols {
		c := column
		total += c.Width
		label := c.Title
		if c.ID == sort.ColumnID {
			if sort.Descending {
				label += " ▾"
			} else {
				label += " ▴"
			}
		}
		header := DocumentHeaderID(id, c.ID)
		resize := func(width float32) {
			if !live() || d.Resize == nil || !finite(width) {
				return
			}
			for _, current := range d.columns() {
				if current.ID == c.ID {
					width = max(60, min(600, width))
					if current.Width != width {
						d.Resize(DocumentResizeEvent{ID: id, ColumnID: c.ID, Width: width})
					}
					return
				}
			}
		}
		headers = append(headers, Node{ID: header + "/slot", Kind: Row, Style: Style{MinSize: Some(Size{Width: c.Width}), MaxSize: Some(Size{Width: c.Width}), Gap: Some(float32(0))}, Children: []Node{
			{ID: header, Kind: Button, Text: label, Appearance: ButtonTool, Grow: true, Disabled: d.SortChange == nil, Style: Style{Border: Some(Border{}), Padding: Some(Insets{Left: 6, Right: 2, Top: 4, Bottom: 4}), TextAlign: Some(AlignStart)}, Activate: func() {
				if !live() || d.SortChange == nil {
					return
				}
				for _, current := range d.columns() {
					if current.ID == c.ID {
						next := DocumentSort{ColumnID: c.ID}
						if d.Sort.ColumnID == c.ID {
							next.Descending = !d.Sort.Descending
						}
						d.SortChange(DocumentSortEvent{ID: id, Sort: next})
						return
					}
				}
			}},
			{ID: header + "/resize", Kind: Button, Text: "│", Appearance: ButtonTool, Disabled: d.Resize == nil, ColumnResize: &ColumnResizeHandle{source: d, column: c.ID, Width: c.Width, Change: resize}, Accessibility: Accessibility{Name: "Resize " + c.Title, Help: "Drag or use Left/Right. Shift changes ten points."}, Style: Style{Border: Some(Border{}), Padding: Some(Insets{})}},
		}})
		if c.ID == sort.ColumnID {
			children := headers[len(headers)-1].Children
			children[0].Text = c.Title
			symbol := SymbolUp
			label := "Sorted ascending"
			if sort.Descending {
				symbol = SymbolDown
				label = "Sorted descending"
			}
			chevron := Node{ID: header + "/sort", Kind: Icon, Icon: symbol, Text: label, Style: Style{MinSize: Some(Size{Width: 10, Height: 10}), MaxSize: Some(Size{Width: 10, Height: 10})}}
			nextChildren := []Node{children[0], chevron}
			children = append(nextChildren, children[1:]...)
			headers[len(headers)-1].Children = children
		}

	}
	for _, i := range items {
		cells := []string{}
		for _, c := range cols {
			value := ""
			switch c.ID {
			case "name":
				value = i.Name
			case "kind":
				value = i.Kind
			case "modified":
				if !i.Modified.IsZero() {
					value = i.Modified.Format("Jan 2, 2006")
				}
			case "size":
				if !i.Folder {
					value = strconv.FormatInt(i.SizeBytes, 10) + " B"
				}
			}
			cells = append(cells, value)
		}
		layout.Rows = append(layout.Rows, DocumentRow{ID: i.ID, Cells: cells, Icon: i.Icon, Image: i.Image})
		list.Options = append(list.Options, Option{ID: i.ID, Label: i.Name, Disabled: i.Disabled})
	}
	list.DocumentList = layout
	list.Style.MinSize = Some(Size{Width: total, Height: 160})
	if d.Change != nil {
		list.SelectionChange = func(e ListSelectionEvent) {
			if !live() || d.Change == nil || d.validate() != nil {
				return
			}
			enabled := make(map[string]bool, len(d.Items))
			for _, item := range d.Items {
				enabled[item.ID] = !item.Disabled
			}
			if !enabled[e.TargetID] {
				return
			}
			for _, target := range e.IDs {
				if !enabled[target] {
					return
				}
			}
			e.ID = id
			e.IDs = slices.Clone(e.IDs)
			d.Change(e)
		}
	}
	if d.Activate != nil {
		layout.Activate = func(e DocumentEvent) {
			if d.Activate != nil && eligible(e.ItemID) {
				d.Activate(DocumentEvent{ID: id, ItemID: e.ItemID})
			}
		}
	}
	n := Node{ID: id, Kind: Scroll, Disabled: d.Disabled, Style: Style{GrowX: Some(float32(1)), Padding: Some(Insets{}), Gap: Some(float32(0)), MaxSize: Some(Size{Height: 400})}.Overlay(d.Style), Children: []Node{{ID: id + "/documents/headers", Kind: Row, Style: Style{Gap: Some(float32(0)), MinSize: Some(Size{Width: total})}, Children: headers}, list}}
	if err := Validate(n); err != nil {
		return Node{}, err
	}
	return n, nil
}
func ValidateDocumentListNode(n Node) error {
	l := n.DocumentList
	if l == nil {
		return nil
	}
	if l.source == nil || n.Kind != ListBox || !n.Multiple || n.Collection != nil || len(l.Rows) != len(n.Options) || len(l.Columns) < 1 || len(l.Columns) > 4 {
		return fmt.Errorf("ui: invalid document list presentation")
	}

	seen := map[string]bool{}
	for _, column := range l.Columns {
		if !documentColumn(column.ID) || seen[column.ID] || !validTokenText(column.Title) || !finite(column.Width) || column.Width < 60 || column.Width > 600 {
			return fmt.Errorf("ui: invalid detached document column")
		}
		seen[column.ID] = true
	}
	for i, row := range l.Rows {
		if err := validateIcon(Node{Kind: Button, Text: n.Options[i].Label, Icon: row.Icon, IconImage: row.Image}); err != nil {
			return err
		}
		for _, cell := range row.Cells {
			if len(cell) > 1024 || !utf8.ValidString(cell) || strings.ContainsAny(cell, "\x00\r\n") {
				return fmt.Errorf("ui: invalid detached document cell")
			}
		}
		if row.ID != n.Options[i].ID || len(row.Cells) != len(l.Columns) {
			return fmt.Errorf("ui: document rows/options disagree")
		}
	}
	return nil
}

func (l *DocumentListLayout) SameSource(other *DocumentListLayout) bool {
	if l == nil || other == nil {
		return l == other
	}
	return l.source != nil && l.source == other.source
}
func (h *ColumnResizeHandle) SameSource(other *ColumnResizeHandle) bool {
	if h == nil || other == nil {
		return h == other
	}
	return h.source != nil && h.source == other.source && h.column == other.column
}
