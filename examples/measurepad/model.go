package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/quantity"
	"github.com/dshills/vellum/ui"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type measurement struct {
	ID          uint64
	Name        string
	Millimeters float64
}
type notebook struct {
	Version int
	Next    uint64
	Entries []measurement
}
type model struct {
	book                                 notebook
	journal                              *history.Journal
	name, amount, unit, selected, status string
	changed                              func()
}

func newModel() *model {
	j, _ := history.New(100)
	return &model{book: notebook{Version: 1, Next: 1}, journal: j, name: "Part", amount: "25", unit: "mm", changed: func() {}}
}
func validate(b notebook) error {
	if b.Version != 1 || b.Next < 1 || b.Next > 1<<53 || len(b.Entries) > 100 {
		return fmt.Errorf("invalid notebook limits")
	}
	seen := map[uint64]bool{}
	for _, v := range b.Entries {
		if v.ID < 1 || v.ID >= b.Next || seen[v.ID] || strings.TrimSpace(v.Name) == "" || len(v.Name) > 128 || !utf8.ValidString(v.Name) || math.IsNaN(v.Millimeters) || math.IsInf(v.Millimeters, 0) || v.Millimeters <= 0 || v.Millimeters > 1e6 {
			return fmt.Errorf("invalid measurement")
		}
		for _, r := range v.Name {
			if unicode.IsControl(r) {
				return fmt.Errorf("control character in name")
			}
		}
		seen[v.ID] = true
	}
	return nil
}
func clone(b notebook) notebook { b.Entries = append([]measurement(nil), b.Entries...); return b }
func encode(b notebook) ([]byte, error) {
	if e := validate(b); e != nil {
		return nil, e
	}
	return json.MarshalIndent(b, "", "  ")
}
func decode(data []byte) (notebook, error) {
	var b notebook
	if len(data) > 65536 {
		return b, fmt.Errorf("notebook too large")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if e := d.Decode(&b); e != nil {
		return b, e
	}
	if d.Decode(new(any)) != io.EOF {
		return b, fmt.Errorf("trailing data")
	}
	return b, validate(b)
}
func (m *model) install(b notebook) error {
	if e := validate(b); e != nil {
		return e
	}
	m.book = clone(b)
	m.selected = ""
	m.changed()
	return nil
}
func (m *model) apply(label string, b notebook) error {
	if e := validate(b); e != nil {
		return e
	}
	before, after := clone(m.book), clone(b)
	return m.journal.Apply(history.Change{Label: label, Apply: func() error { return m.install(after) }, Revert: func() error { return m.install(before) }})
}
func (m *model) add() error {
	if len(m.amount) > 100 {
		return fmt.Errorf("expression too long")
	}
	v, e := quantity.Length(m.amount, m.unit, nil)
	if e != nil {
		return e
	}
	b := clone(m.book)
	b.Entries = append(b.Entries, measurement{b.Next, strings.TrimSpace(m.name), v})
	b.Next++
	return m.apply("Add measurement", b)
}
func (m *model) remove() error {
	b := clone(m.book)
	for i, v := range b.Entries {
		if strconv.FormatUint(v.ID, 10) == m.selected {
			b.Entries = append(b.Entries[:i], b.Entries[i+1:]...)
			return m.apply("Delete measurement", b)
		}
	}
	return fmt.Errorf("select a measurement")
}
func (m *model) result(e error) {
	if e != nil {
		m.status = e.Error()
	} else {
		m.status = "Ready"
	}
	m.changed()
}
func (m *model) view(actions ...ui.Node) ui.Node {
	options := []ui.Option{}
	for _, v := range m.book.Entries {
		options = append(options, ui.Option{ID: strconv.FormatUint(v.ID, 10), Label: fmt.Sprintf("%s   •   %.3f mm   /   %.4f in", v.Name, v.Millimeters, v.Millimeters/25.4)})
	}
	children := []ui.Node{
		{ID: "heading", Kind: ui.Label, Text: "Measure Pad", Style: ui.Style{Typography: ui.Some(ui.MacTypography(22, 600))}},
		{ID: "help", Kind: ui.Label, Text: "A local notebook for dimensions. Enter 2 in or arithmetic length expressions."},
		{ID: "name", Kind: ui.TextField, Text: m.name, Accessibility: ui.Accessibility{Name: "Measurement name"}, Change: func(e ui.ChangeEvent) { m.name = e.Text }},
		{ID: "amount", Kind: ui.TextField, Text: m.amount, Accessibility: ui.Accessibility{Name: "Length expression"}, Change: func(e ui.ChangeEvent) { m.amount = e.Text }},
		{ID: "unit", Kind: ui.Selector, SelectedID: m.unit, Options: []ui.Option{{ID: "mm", Label: "Millimeters"}, {ID: "cm", Label: "Centimeters"}, {ID: "in", Label: "Inches"}}, Accessibility: ui.Accessibility{Name: "Default units"}, Change: func(e ui.ChangeEvent) { m.unit = e.SelectedID }},
	}
	children = append(children, actions...)
	children = append(children, ui.Node{ID: "measurements", Kind: ui.ListBox, Options: options, SelectedID: m.selected, Accessibility: ui.Accessibility{Name: "Measurements"}, Change: func(e ui.ChangeEvent) { m.selected = e.SelectedID; m.changed() }}, ui.Node{ID: "status", Kind: ui.Label, Text: m.status})
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(20)), Gap: ui.Some(float32(10))}, Children: children}
}
