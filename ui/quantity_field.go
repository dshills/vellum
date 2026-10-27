package ui

import (
	"fmt"
	"github.com/dshills/vellum/quantity"
	"math"
)

// QuantityField is a controlled length-expression editor. Editing proposes Text;
// Apply parses current Text and proposes canonical millimeters once. Invalid
// input never calls Commit. Owners retain draft text and validation messages.
type QuantityField struct {
	ID, Label, Text, Unit, ValidationMessage string
	Variables                                map[string]float64
	Min, Max                                 float64
	Disabled                                 bool
	Change                                   func(string)
	Commit                                   func(float64)
	Invalid                                  func(error)
	Style                                    Style
}

func (p *QuantityField) parse() (float64, error) {
	v, e := quantity.Length(p.Text, p.Unit, p.Variables)
	if e == nil && (v < p.Min || v > p.Max) {
		e = fmt.Errorf("length must be between %g and %g mm", p.Min, p.Max)
	}
	return v, e
}
func (p *QuantityField) View() (Node, error) {
	if p == nil || !validTokenText(p.ID) || !validTokenText(p.Label) || math.IsNaN(p.Min) || math.IsNaN(p.Max) || math.IsInf(p.Min, 0) || math.IsInf(p.Max, 0) || p.Min > p.Max {
		return Node{}, fmt.Errorf("ui: invalid quantity field")
	}
	if _, ok := quantity.Unit(p.Unit); !ok {
		return Node{}, fmt.Errorf("ui: invalid quantity unit")
	}
	id := p.ID
	input := Node{ID: id + "/quantity/input", Kind: TextField, Text: p.Text, Disabled: p.Change == nil, Accessibility: Accessibility{Name: p.Label}, Style: Style{GrowX: Some(float32(1)), MinSize: Some(Size{Width: 64, Height: 28})}, Change: func(e ChangeEvent) {
		if p.ID == id && !p.Disabled && p.Change != nil {
			p.Change(e.Text)
		}
	}}
	apply := Node{ID: id + "/quantity/apply", Kind: Button, Text: "Apply", Disabled: p.Commit == nil, Accessibility: Accessibility{Name: "Apply " + p.Label}, Activate: func() {
		if p.ID != id || p.Disabled || p.Commit == nil {
			return
		}
		v, e := p.parse()
		if e != nil {
			if p.Invalid != nil {
				p.Invalid(e)
			}
			return
		}
		p.Commit(v)
	}}
	n := Node{ID: id, Kind: Column, Disabled: p.Disabled, Style: Style{Gap: Some(float32(6)), GrowX: Some(float32(1))}.Overlay(p.Style), Children: []Node{{ID: id + "/quantity/label", Kind: Label, Text: p.Label, Style: Style{Typography: Some(MacTypography(11, 500))}}, {ID: id + "/quantity/row", Kind: Row, Style: Style{Gap: Some(float32(6)), GrowX: Some(float32(1))}, Children: []Node{input, apply}}}}
	if p.ValidationMessage != "" {
		n.Children = append(n.Children, Node{ID: id + "/quantity/error", Kind: Label, Text: p.ValidationMessage, Style: Style{Typography: Some(MacTypography(11, 400))}})
	}
	return n, Validate(n)
}
