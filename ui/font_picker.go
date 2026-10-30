package ui

import (
	"fmt"
	"image/color"
	"reflect"
	"strings"
	"unicode/utf8"
)

const (
	MinFontPickerSize float32 = 6
	MaxFontPickerSize float32 = 144
)

// FontTypeface describes an available style within a caller-provided family.
// Weight is 0 (driver default) or 100..900 in increments of 100.
type FontTypeface struct {
	ID, Name string
	Weight   int
	Italic   bool
}

// FontFamily names an installed family and its caller-approved styles. Vellum
// does not enumerate fonts or guarantee availability; the driver may substitute.
// At most MaxMenuOptions families and typefaces per family are supported.
type FontFamily struct {
	ID, Name  string
	Typefaces []FontTypeface
}

// FontSelection uses stable catalog identities and a size in logical pixels.
type FontSelection struct {
	FamilyID, TypefaceID string
	Size                 float32
}

type FontField string

const (
	FontFamilyField   FontField = "family"
	FontTypefaceField FontField = "typeface"
	FontSizeField     FontField = "size"
)

// FontSelectionEvent proposes one field edit. Apply merges independent queued
// edits into the current accepted selection. A family edit includes its valid
// typeface; a typeface edit is ignored if the family has since changed.
type FontSelectionEvent struct {
	ID    string
	Field FontField
	Value FontSelection
}

func (e FontSelectionEvent) Apply(current FontSelection) FontSelection {
	switch e.Field {
	case FontFamilyField:
		current.FamilyID, current.TypefaceID = e.Value.FamilyID, e.Value.TypefaceID
	case FontTypefaceField:
		if current.FamilyID == e.Value.FamilyID {
			current.TypefaceID = e.Value.TypefaceID
		}
	case FontSizeField:
		current.Size = e.Value.Size
	}
	return current
}

// FontPicker is an owner-held, controlled family/typeface/size editor with a
// live plain-text preview. Accept proposals in Change, then call View again.
// Typed names must exactly match a catalog choice; arbitrary fonts are
// not accepted. Nil Change makes the editor read-only. Preview is bounded to
// 1024 UTF-8 bytes. Empty Preview uses a standard sample. The root reserves its
// root+"/font/" namespace. Do not mutate or render concurrently across owners.
type FontPicker struct {
	ID, Label, Preview string
	Value              FontSelection
	Families           []FontFamily
	Disabled           bool
	Change             func(FontSelectionEvent)
	Invalid            func(error)
	Style              Style
	FieldStyle         Style // optional family/typeface field overrides; internal content owns padding
}

func FontFamilyID(root string) string   { return root + "/font/family" }
func FontTypefaceID(root string) string { return root + "/font/typeface" }
func FontSizeID(root string) string     { return root + "/font/size" }
func FontPreviewID(root string) string  { return root + "/font/preview" }

func fontPickerName(s string) bool {
	return len(s) > 0 && len(s) <= 128 && strings.TrimSpace(s) == s && utf8.ValidString(s) && !strings.ContainsAny(s, "\x00\r\n\t")
}

func (p *FontPicker) validate() error {
	if p == nil || !fontPickerName(p.ID) || len(p.Families) == 0 || len(p.Families) > MaxMenuOptions || len(p.Preview) > 1024 || !utf8.ValidString(p.Preview) || strings.ContainsRune(p.Preview, 0) {
		return fmt.Errorf("ui: font picker requires an ID, a bounded catalog and a UTF-8 preview of at most 1024 bytes")
	}
	families, names := map[string]bool{}, map[string]bool{}
	selectedFamily, selectedFace := false, false
	for _, f := range p.Families {
		if !fontPickerName(f.ID) || !fontPickerName(f.Name) || families[f.ID] || names[f.Name] || len(f.Typefaces) == 0 || len(f.Typefaces) > MaxMenuOptions {
			return fmt.Errorf("ui: font families require unique identities/names and bounded typefaces")
		}
		families[f.ID], names[f.Name] = true, true
		faces, faceNames := map[string]bool{}, map[string]bool{}
		for _, face := range f.Typefaces {
			if !fontPickerName(face.ID) || !fontPickerName(face.Name) || faces[face.ID] || faceNames[face.Name] || (face.Weight != 0 && (face.Weight < 100 || face.Weight > 900 || face.Weight%100 != 0)) {
				return fmt.Errorf("ui: font typefaces require unique identities/names and weight 0 or 100..900")
			}
			faces[face.ID], faceNames[face.Name] = true, true
			if f.ID == p.Value.FamilyID && face.ID == p.Value.TypefaceID {
				selectedFace = true
			}
		}
		selectedFamily = selectedFamily || f.ID == p.Value.FamilyID
	}
	if !selectedFamily || !selectedFace || !finite(p.Value.Size) || p.Value.Size < MinFontPickerSize || p.Value.Size > MaxFontPickerSize {
		return fmt.Errorf("ui: font selection must exist in the catalog and have size 6..144")
	}
	return nil
}

func (p *FontPicker) View() (Node, error) {
	if err := p.validate(); err != nil {
		return Node{}, err
	}
	id := p.ID
	catalog := make([]FontFamily, len(p.Families))
	for i, f := range p.Families {
		catalog[i] = f
		catalog[i].Typefaces = append([]FontTypeface(nil), f.Typefaces...)
	}
	available := func() bool {
		return p.ID == id && !p.Disabled && p.Change != nil && reflect.DeepEqual(catalog, p.Families) && p.validate() == nil
	}
	invalid := func() {
		if p.Invalid != nil {
			p.Invalid(fmt.Errorf("ui: choose a listed font name or a finite size from 6 to 144"))
		}
	}
	label := p.Label
	if strings.TrimSpace(label) == "" {
		label = "Font"
	}
	var family FontFamily
	var face FontTypeface
	familyOptions := make([]Option, 0, len(catalog))
	for _, f := range catalog {
		familyOptions = append(familyOptions, Option{ID: f.ID, Label: f.Name})
		if f.ID == p.Value.FamilyID {
			family = f
		}
	}
	faceOptions := make([]Option, 0, len(family.Typefaces))
	for _, f := range family.Typefaces {
		faceOptions = append(faceOptions, Option{ID: f.ID, Label: f.Name})
		if f.ID == p.Value.TypefaceID {
			face = f
		}
	}
	comboStyle := Style{MinSize: Some(Size{Width: 160, Height: 16}), GrowX: Some(float32(1)), GrowY: Some(float32(0)), Padding: Some(Insets{}), CornerRadius: Some(float32(5)), Border: Some(Border{Width: 1, Color: color.NRGBA{R: 128, G: 128, B: 132, A: 255}})}.Overlay(p.FieldStyle)
	// ComboBox children already supply text and arrow insets. Reusing an input
	// theme must not add a second outer layer of field padding.
	comboStyle.Padding = Some(Insets{})
	familyCombo := &ComboBox{ID: FontFamilyID(id), Label: label + " family", Text: family.Name, Options: familyOptions, Disabled: !available(), Style: comboStyle}
	familyCombo.Change = func(e ComboEvent) {
		if !available() {
			return
		}
		for _, f := range catalog {
			if (e.OptionID != "" && e.OptionID == f.ID) || (e.OptionID == "" && e.Text == f.Name) {
				v := p.Value
				v.FamilyID, v.TypefaceID = f.ID, f.Typefaces[0].ID
				for _, tf := range f.Typefaces {
					if tf.ID == p.Value.TypefaceID {
						v.TypefaceID = tf.ID
					}
				}
				p.Change(FontSelectionEvent{ID: id, Field: FontFamilyField, Value: v})
				return
			}
		}
		invalid()
	}
	faceCombo := &ComboBox{ID: FontTypefaceID(id), Label: label + " typeface", Text: face.Name, Options: faceOptions, Disabled: !available(), Style: comboStyle}
	// An event from the prior family is obsolete, not invalid user input.
	// Owner dispatch reconciles View after each accepted edit.
	faceCombo.Change = func(e ComboEvent) {
		if !available() || p.Value.FamilyID != family.ID {
			return
		}
		for _, tf := range family.Typefaces {
			if (e.OptionID != "" && e.OptionID == tf.ID) || (e.OptionID == "" && e.Text == tf.Name) {
				v := p.Value
				v.TypefaceID = tf.ID
				p.Change(FontSelectionEvent{ID: id, Field: FontTypefaceField, Value: v})
				return
			}
		}
		invalid()
	}
	familyNode, err := familyCombo.View()
	if err != nil {
		return Node{}, err
	}
	faceNode, err := faceCombo.View()
	if err != nil {
		return Node{}, err
	}
	preview := p.Preview
	if preview == "" {
		preview = "The quick brown fox jumps over the lazy dog."
	}
	root := Node{ID: id, Kind: Column, Disabled: !available(), Style: Style{Gap: Some(float32(8)), GrowX: Some(float32(1)), GrowY: Some(float32(0))}.Overlay(p.Style), Accessibility: Accessibility{Name: label}}
	root.Children = []Node{
		{ID: id + "/font/family-label", Kind: Label, Text: "Family"}, familyNode,
		{ID: id + "/font/style-row", Kind: Row, Style: Style{Gap: Some(float32(8)), Align: Some(AlignCenter)}, Children: []Node{
			faceNode,
			{ID: FontSizeID(id), Kind: NumberField, Number: p.Value.Size, Numeric: Numeric{Min: MinFontPickerSize, Max: MaxFontPickerSize, Step: 1}, NumberButtons: NumberButtonsStepper, Disabled: !available(), Style: Style{MinSize: Some(Size{Width: 88, Height: 16}), MaxSize: Some(Size{Width: 88}), GrowX: Some(float32(0))}, Accessibility: Accessibility{Name: label + " size", Help: "Font size from 6 to 144 logical pixels"}, Change: func(e ChangeEvent) {
				if !available() {
					return
				}
				if !finite(e.Number) || e.Number < MinFontPickerSize || e.Number > MaxFontPickerSize {
					invalid()
					return
				}
				v := p.Value
				v.Size = e.Number
				p.Change(FontSelectionEvent{ID: id, Field: FontSizeField, Value: v})
			}},
		}},
		{ID: FontPreviewID(id), Kind: Label, Text: preview, Accessibility: Accessibility{Name: label + " preview: " + preview}, Style: Style{Padding: Some(UniformInsets(10)), Typography: Some(Typography{Family: family.Name, Size: p.Value.Size, Weight: face.Weight, Italic: face.Italic}), GrowX: Some(float32(1)), GrowY: Some(float32(0))}},
	}
	if err := Validate(root); err != nil {
		return Node{}, err
	}
	return root, nil
}
