package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"strconv"
)

// Property requests use the same sparse document history as commands and replay.
func inspectorView(parts *partSource, edit *partEditor, expanded bool, expansion func(bool), selected func() int, status func(string)) ui.Node {
	index := selected()
	row := parts.Row(index)
	value := parts.parts.at(index)
	change := func(label string, assign func(*part)) {
		i := selected()
		next := parts.parts.at(i)
		assign(&next)
		if err := edit.change(i, label, next); err != nil {
			status(err.Error())
			return
		}
		status(label)
	}
	return ui.Node{ID: "inspector.root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "inspector.reference", Kind: ui.Label, Text: row.Cells[0]},
		{ID: "inspector.properties", Kind: ui.Disclosure, Text: "Component properties", Checked: expanded, Change: func(e ui.ChangeEvent) { expansion(e.Checked) }, Children: []ui.Node{
			{ID: "inspector.fields", Kind: ui.FormGroup, Text: "Properties", Children: []ui.Node{
				{ID: "inspector.value", Kind: ui.TextField, Text: strconv.FormatUint(uint64(value.Value), 10), Accessibility: ui.Accessibility{Name: "Component value in ohms"}, Change: func(e ui.ChangeEvent) {
					v, err := strconv.ParseUint(e.Text, 10, 32)
					if err != nil {
						status("Value requires an integer from 0 to 4294967295")
						return
					}
					change("Edit component value", func(p *part) { p.Value = uint32(v) })
				}},
				{ID: "inspector.package", Kind: ui.RadioGroup, SelectedID: strconv.Itoa(int(value.Package)), Options: []ui.Option{{ID: "0", Label: "0603"}, {ID: "1", Label: "0805"}}, Accessibility: ui.Accessibility{Name: "Component package"}, Change: func(e ui.ChangeEvent) {
					if e.SelectedID != "0" && e.SelectedID != "1" {
						return
					}
					change("Edit component package", func(p *part) { p.Package = uint16(e.SelectedID[0] - '0') })
				}},
				{ID: "inspector.rotation", Kind: ui.SegmentedControl, SelectedID: strconv.Itoa(int(value.Rotation)), Options: []ui.Option{{ID: "0", Label: "0°"}, {ID: "90", Label: "90°"}, {ID: "180", Label: "180°"}, {ID: "270", Label: "270°"}}, Accessibility: ui.Accessibility{Name: "Component rotation"}, Change: func(e ui.ChangeEvent) {
					v, err := strconv.Atoi(e.SelectedID)
					if err != nil || v < 0 || v > 270 || v%90 != 0 {
						return
					}
					change("Edit component rotation", func(p *part) { p.Rotation = uint16(v) })
				}},
			}},
		}},
		{ID: "inspector.summary", Kind: ui.StatusBar, Text: fmt.Sprintf("%s · %s · %s", row.Cells[1], row.Cells[2], row.Cells[3])},
	}}
}
