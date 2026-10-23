package ui

import (
	"fmt"
	"image/color"
	"strings"
)

// ColorWell is an owner-held, controlled color swatch with an anchored editor.
// Retain one instance on one application owner. Change receives ColorEdit
// proposals; apply them to the current accepted Color to merge queued channel
// edits. The editor reuses ColorPicker's straight-alpha sRGB bytes, opacity
// slider, hex validation and checkerboard preview. The compact button composites
// alpha over a neutral white backing; the editor shows the actual checkerboard.
// Recent is caller-owned accepted history (at most 16 colors); View never adds
// rejected proposals to it. This is a same-window panel, not NSColorPanel.
type ColorWell struct {
	ID, Label           string
	Color               color.NRGBA
	Recent              []color.NRGBA
	Disabled            bool
	Change              func(ColorEdit)
	Invalid             func(error)
	ValidationMessage   string
	Style, PopoverStyle Style
	picker              ColorPicker
	open                bool
	session             uint64
}

func ColorWellButtonID(root string) string { return root + "/color-well/button" }
func ColorWellPickerID(root string) string { return root + "/color-well/editor" }

func (p *ColorWell) close() {
	p.open = false
	p.session++
}

func (p *ColorWell) View() (Node, error) {
	if p == nil || strings.TrimSpace(p.ID) == "" || len(p.Recent) > 16 {
		return Node{}, fmt.Errorf("ui: color well requires an ID and at most 16 recent colors")
	}
	id := p.ID
	label := p.Label
	if strings.TrimSpace(label) == "" {
		label = "Color"
	}
	available := func() bool { return p.ID == id && !p.Disabled && p.Change != nil }
	session := p.session
	editing := func() bool { return available() && p.open && p.session == session }
	// Keep the editor object so unchanged immutable preview pixels can be reused.
	p.picker.ID = ColorWellPickerID(id)
	p.picker.Label = label
	p.picker.Color = p.Color
	p.picker.Disabled = !available()
	p.picker.ValidationMessage = p.ValidationMessage
	p.picker.Change = func(edit ColorEdit) {
		if editing() {
			p.Change(edit)
		}
	}
	p.picker.Invalid = func(err error) {
		if editing() && p.Invalid != nil {
			p.Invalid(err)
		}
	}
	// Snapshot callback wrappers: an old description must not accidentally call
	// a new session's wrapper through the retained picker after it is rerendered.
	picker := p.picker
	content := []Node{picker.View()}
	p.picker.preview, p.picker.previewColor, p.picker.previewValid = picker.preview, picker.previewColor, picker.previewValid
	if len(p.Recent) > 0 {
		rows := []Node{}
		seen := map[color.NRGBA]bool{}
		for _, c := range p.Recent {
			if seen[c] {
				continue
			}
			seen[c] = true
			candidate := c
			rows = append(rows, Node{ID: id + "/color-well/recent/" + ColorHexString(c), Kind: Button, Text: " ", Style: colorWellSwatchStyle(c), Accessibility: Accessibility{Name: label + " recent " + ColorHexString(c)}, Activate: func() {
				if !editing() {
					return
				}
				for _, current := range p.Recent {
					if current == candidate {
						p.Change(ColorEdit{Channel: ColorHex, RGBA: candidate})
						return
					}
				}
			}})
		}
		content = append(content, Node{ID: id + "/color-well/recent-label", Kind: Label, Text: "Recent colors"})
		// Eight compact swatches per row keep the editor's width bounded.
		for start := 0; start < len(rows); start += 8 {
			end := min(start+8, len(rows))
			content = append(content, Node{ID: fmt.Sprintf("%s/color-well/recent-row/%d", id, start/8), Kind: Row, Style: Style{Gap: Some(float32(4))}, Children: rows[start:end]})
		}
	}
	content = append(content, Node{ID: id + "/color-well/done", Kind: Button, Text: "Done", Activate: func() {
		if editing() {
			p.close()
		}
	}})
	panel, err := (Popover{ID: id + "/color-well/popup", AnchorID: ColorWellButtonID(id), Label: label, Open: p.open, Disabled: !available(), Style: p.PopoverStyle, Children: content, Dismiss: func() {
		if p.ID == id && p.session == session {
			p.close()
		}
	}}).View()
	if err != nil {
		return Node{}, err
	}
	button := Node{ID: ColorWellButtonID(id), Kind: Button, Text: " ", Disabled: !available(), Style: colorWellSwatchStyle(p.Color).Overlay(p.Style), Accessibility: Accessibility{Name: label + " " + ColorHexString(p.Color), Help: "Open color editor; includes opacity and recent colors"}, Activate: func() {
		if available() && p.session == session {
			p.open = !p.open
			p.session++
		}
	}}
	root := Node{ID: id, Kind: Column, Style: Style{GrowX: Some(float32(0)), Gap: Some(float32(0))}, Children: []Node{button, panel}}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}

func colorWellSwatchStyle(c color.NRGBA) Style {
	blend := func(v uint8) uint8 { return uint8((uint16(v)*uint16(c.A) + 255*(255-uint16(c.A)) + 127) / 255) }
	return Style{
		MinSize: Some(Size{Width: 36, Height: 24}), MaxSize: Some(Size{Width: 36, Height: 24}),
		Grow: Some(float32(0)), Padding: Some(UniformInsets(3)), CornerRadius: Some(float32(4)),
		Border:     Some(Border{Width: 1, Color: color.NRGBA{R: 142, G: 142, B: 147, A: 255}}),
		Background: Some(color.NRGBA{R: blend(c.R), G: blend(c.G), B: blend(c.B), A: 255}),
	}
}
