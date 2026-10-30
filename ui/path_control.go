package ui

import (
	"encoding/base64"
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"
)

// MaxPathItems bounds a PathControl's complete domain path, including hidden ancestors.
const MaxPathItems = 64

type PathItemKind string

const (
	PathFolder PathItemKind = "folder"
	PathFile   PathItemKind = "file"
)

// PathItem is an application-provided location, not a filesystem URL. Empty Kind
// means folder. Only the final location may be a file. Labels are literal text.
type PathItem struct {
	ID, Label string
	Kind      PathItemKind
	Disabled  bool
}

// PathEvent proposes navigation. The owner accepts by replacing Items; the
// component never reads the filesystem or modifies its accepted path.
type PathEvent struct{ ID, ItemID string }

// PathControl presents compact Finder-like segments with ancestor menus. Retain
// on one owner. Nil Navigate is read-only. No file access or drag/drop is implied.
// MaxVisible is 0 (default 4) or 3..8 visible path segments, excluding the overflow
// button. Longer paths retain the root and the last MaxVisible-1 segments, so the
// parent and current location remain visible. This bound is explicit, rather
// than dependent on the window width. MaxLabelRunes is 0 (default 24) or 8..128;
// visible labels truncate centrally while accessibility and menus keep full text.
// The root owns root+"/path/". Supply a MacTheme to inherit macOS typography/colors.
type PathControl struct {
	ID, Label                 string
	Items                     []PathItem
	Disabled                  bool
	Navigate                  func(PathEvent)
	MaxVisible, MaxLabelRunes int
	Style                     Style
}

func PathSegmentID(root, item string) string {
	return root + "/path/item/" + base64.RawURLEncoding.EncodeToString([]byte(item))
}
func PathOverflowID(root string) string { return root + "/path/overflow" }

func (p *PathControl) validate() error {
	if p == nil || p.ID == "" || len(p.Items) == 0 || len(p.Items) > MaxPathItems {
		return fmt.Errorf("ui: path control requires an ID and 1..%d items", MaxPathItems)
	}
	if p.MaxVisible != 0 && (p.MaxVisible < 3 || p.MaxVisible > 8) {
		return fmt.Errorf("ui: path visible segment bound must be 0 or 3..8")
	}
	if p.MaxLabelRunes != 0 && (p.MaxLabelRunes < 8 || p.MaxLabelRunes > 128) {
		return fmt.Errorf("ui: path label bound must be 0 or 8..128")
	}
	seen := map[string]bool{}
	for i, item := range p.Items {
		if item.ID == "" || len(item.ID) > 1024 || seen[item.ID] || strings.TrimSpace(item.Label) == "" || len(item.Label) > 4096 || !utf8.ValidString(item.ID) || !utf8.ValidString(item.Label) || strings.ContainsRune(item.ID, 0) || strings.ContainsRune(item.Label, 0) {
			return fmt.Errorf("ui: path items require unique bounded UTF-8 IDs and nonblank labels without NUL")
		}
		if item.Kind != "" && item.Kind != PathFolder && item.Kind != PathFile || item.Kind == PathFile && i != len(p.Items)-1 {
			return fmt.Errorf("ui: only the current path item can be a file")
		}
		seen[item.ID] = true
	}
	return p.Style.Validate()
}

func pathShortLabel(label string, limit int) string {
	r := []rune(label)
	if len(r) <= limit {
		return label
	}
	left := (limit - 1 + 1) / 2
	return string(r[:left]) + "…" + string(r[len(r)-(limit-1-left):])
}

// View returns detached controls. Ancestor menus list the nearest location first;
// an ancestor's own menu also includes itself. The current item is never a
// navigation destination. Actions use the latest callback and reject changes to
// the rendered path (including labels, item kinds and disabled flags).
func (p *PathControl) View() (Node, error) {
	if err := p.validate(); err != nil {
		return Node{}, err
	}
	id := p.ID
	shown := append([]PathItem(nil), p.Items...)
	label := p.Label
	if strings.TrimSpace(label) == "" {
		label = "Location"
	}
	visible, limit := p.MaxVisible, p.MaxLabelRunes
	if visible == 0 {
		visible = 4
	}
	if limit == 0 {
		limit = 24
	}
	style := Style{Background: Some(color.NRGBA{}), Gap: Some(float32(2)), Align: Some(AlignCenter), GrowX: Some(float32(1)), GrowY: Some(float32(0))}.Overlay(p.Style)
	root := Node{ID: id, Kind: Toolbar, Disabled: p.Disabled, Style: style, Accessibility: Accessibility{Name: label}}
	propose := func(target string) {
		if p.validate() != nil || p.ID != id || p.Disabled || p.Navigate == nil || len(p.Items) != len(shown) {
			return
		}
		for i, item := range shown {
			if p.Items[i] != item {
				return
			}
		}
		for _, item := range shown[:len(shown)-1] {
			if item.ID == target && !item.Disabled {
				p.Navigate(PathEvent{ID: id, ItemID: target})
				return
			}
		}
	}
	segmentStyle := Style{GrowX: Some(float32(0)), GrowY: Some(float32(0)), Border: Some(Border{}), Background: Some(color.NRGBA{}), Padding: Some(Insets{Top: 1, Bottom: 1, Left: 5, Right: 5}), MinSize: Some(Size{Height: 16})}
	appendSegment := func(n Node) {
		if len(root.Children) > 0 {
			root.Children = append(root.Children, Node{ID: n.ID + "/separator", Kind: Icon, Icon: SymbolRight})
		}
		root.Children = append(root.Children, n)
	}
	// Each menu carries only a bounded rendered snapshot; it cannot address an
	// item that was not offered by this particular segment or overflow menu.
	menu := func(n Node, indices []int) Node {
		n.Kind = MenuButton
		for _, i := range indices {
			item := shown[i]
			n.Options = append(n.Options, Option{ID: item.ID, Label: item.Label, Disabled: item.Disabled})
		}
		options := append([]Option(nil), n.Options...)
		n.Change = func(e ChangeEvent) {
			for _, o := range options {
				if o.ID == e.SelectedID && !o.Disabled {
					propose(o.ID)
					return
				}
			}
		}
		return n
	}
	for i, item := range shown {
		if len(shown) > visible && i > 0 && i < len(shown)-(visible-1) {
			if i == 1 {
				var hidden []int
				for j := len(shown) - visible; j >= 1; j-- {
					hidden = append(hidden, j)
				}
				n := Node{ID: PathOverflowID(id), Kind: Label, Text: "…", Appearance: ButtonQuiet, Style: segmentStyle, Accessibility: Accessibility{Name: "Hidden ancestors", Help: "Choose an omitted ancestor"}}
				if p.Navigate != nil {
					n = menu(n, hidden)
				} else {
					n.Appearance = ""
					names := []string{}
					for _, j := range hidden {
						names = append(names, shown[j].Label)
					}
					n.Accessibility.Name = "Hidden ancestors: " + strings.Join(names, " / ")
					n.Accessibility.Help = ""
				}
				appendSegment(n)
			}
			continue
		}
		icon := SymbolFolder
		if item.Kind == PathFile {
			icon = SymbolDocument
		}
		n := Node{ID: PathSegmentID(id, item.ID), Kind: Label, Text: pathShortLabel(item.Label, limit), Disabled: item.Disabled, Style: segmentStyle, Accessibility: Accessibility{Name: item.Label}}
		if i == len(shown)-1 {
			n.Accessibility.Help = "Current location"
		} else {
			n.Accessibility.Help = "Navigate to this location"
		}
		if p.Navigate != nil && len(shown) > 1 {
			n.Icon = icon
			n.Appearance = ButtonQuiet
			if i == 0 {
				n.Kind = Button
				target := item.ID
				n.Activate = func() { propose(target) }
			} else {
				var ancestors []int
				last := i
				if i == len(shown)-1 {
					last--
				}
				for j := last; j >= 0; j-- {
					ancestors = append(ancestors, j)
				}
				n = menu(n, ancestors)
			}
		} else {
			// Labels cannot carry bundled artwork; a row keeps read-only semantics and
			// the same icon/text layout without introducing a focus stop.
			n = Node{ID: n.ID, Kind: Row, Disabled: n.Disabled, Style: Style{Gap: Some(float32(4)), Align: Some(AlignCenter), GrowX: Some(float32(0))}, Accessibility: n.Accessibility, Children: []Node{{ID: n.ID + "/icon", Kind: Icon, Icon: icon}, {ID: n.ID + "/label", Kind: Label, Text: n.Text, Style: segmentStyle}}}
		}
		appendSegment(n)
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
