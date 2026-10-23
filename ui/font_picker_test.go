package ui

import (
	"math"
	"strings"
	"testing"
)

func testFontPicker() *FontPicker {
	return &FontPicker{ID: "font", Families: []FontFamily{
		{ID: "sans", Name: "Helvetica Neue", Typefaces: []FontTypeface{{ID: "regular", Name: "Regular"}, {ID: "bold", Name: "Bold", Weight: 700}, {ID: "italic", Name: "Italic", Weight: 400, Italic: true}}},
		{ID: "serif", Name: "Times New Roman", Typefaces: []FontTypeface{{ID: "regular", Name: "Regular"}}},
	}, Value: FontSelection{FamilyID: "sans", TypefaceID: "regular", Size: 13}}
}

func TestFontPickerControlledEditsAndPreview(t *testing.T) {
	p := testFontPicker()
	p.Preview = "<b>Plain & literal</b>"
	calls, invalids := 0, 0
	p.Change = func(e FontSelectionEvent) {
		calls++
		if e.ID != p.ID {
			t.Fatal(e)
		}
	} // Deliberately reject: View must preserve the accepted model.
	p.Invalid = func(error) { invalids++ }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	familyMenu := helperNode(t, n, ComboMenuID(FontFamilyID(p.ID)))
	faceMenu := helperNode(t, n, ComboMenuID(FontTypefaceID(p.ID)))
	size := helperNode(t, n, FontSizeID(p.ID))
	familyMenu.Change(ChangeEvent{SelectedID: "serif"})
	faceMenu.Change(ChangeEvent{SelectedID: "bold"})
	size.Change(ChangeEvent{Number: 20})
	if calls != 3 || p.Value != (FontSelection{FamilyID: "sans", TypefaceID: "regular", Size: 13}) {
		t.Fatal("controlled request mutated accepted value", calls, p.Value)
	}
	for _, number := range []float32{0, 145, float32(math.NaN()), float32(math.Inf(1))} {
		size.Change(ChangeEvent{Number: number})
	}
	helperNode(t, n, ComboInputID(FontFamilyID(p.ID))).Change(ChangeEvent{Text: "Unlisted"})
	helperNode(t, n, ComboInputID(FontTypefaceID(p.ID))).Change(ChangeEvent{Text: "Missing"})
	if calls != 3 || invalids != 6 {
		t.Fatal("invalid edits admitted", calls, invalids)
	}
	p.Change = func(e FontSelectionEvent) { p.Value = e.Apply(p.Value) }
	faceMenu.Change(ChangeEvent{SelectedID: "italic"})
	size.Change(ChangeEvent{Number: 24})
	n, err = p.View()
	if err != nil {
		t.Fatal(err)
	}
	preview := helperNode(t, n, FontPreviewID(p.ID))
	if preview.Text != p.Preview || preview.Style.Typography.Value != (Typography{Family: "Helvetica Neue", Size: 24, Weight: 400, Italic: true}) {
		t.Fatal("preview lost plain text or accepted typeface", preview)
	}
	familyMenu.Change(ChangeEvent{SelectedID: "serif"})
	if p.Value != (FontSelection{FamilyID: "serif", TypefaceID: "regular", Size: 24}) {
		t.Fatal("family switch did not choose a valid face and retain current size", p.Value)
	}
	faceMenu.Change(ChangeEvent{SelectedID: "bold"})
	if p.Value.TypefaceID != "regular" {
		t.Fatal("old family face callback admitted")
	}
}

func TestFontSelectionQueuedEdits(t *testing.T) {
	v := FontSelection{FamilyID: "a", TypefaceID: "regular", Size: 12}
	family := FontSelectionEvent{Field: FontFamilyField, Value: FontSelection{FamilyID: "b", TypefaceID: "plain", Size: 12}}
	size := FontSelectionEvent{Field: FontSizeField, Value: FontSelection{FamilyID: "a", TypefaceID: "regular", Size: 25}}
	oldFace := FontSelectionEvent{Field: FontTypefaceField, Value: FontSelection{FamilyID: "a", TypefaceID: "bold", Size: 12}}
	v = oldFace.Apply(size.Apply(family.Apply(v)))
	if v != (FontSelection{FamilyID: "b", TypefaceID: "plain", Size: 25}) {
		t.Fatal("queued edits overwrote independent field or changed wrong family", v)
	}
}

func TestFontPickerStaleAndDisabledCallbacks(t *testing.T) {
	p := testFontPicker()
	calls := 0
	p.Change = func(FontSelectionEvent) { calls++ }
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	size := helperNode(t, n, FontSizeID(p.ID))
	familyInput := helperNode(t, n, ComboInputID(FontFamilyID(p.ID)))
	faceMenu := helperNode(t, n, ComboMenuID(FontTypefaceID(p.ID)))
	propose := func() {
		size.Change(ChangeEvent{Number: 20})
		familyInput.Change(ChangeEvent{Text: "Times New Roman"})
		faceMenu.Change(ChangeEvent{SelectedID: "bold"})
	}
	p.Disabled = true
	propose()
	p.Disabled = false
	p.Families[0].Typefaces[1].Name = "Changed"
	propose()
	p.Families[0].Typefaces[1].Name = "Bold"
	p.ID = "replacement"
	propose()
	p.ID = "font"
	p.Change = nil
	propose()
	if calls != 0 {
		t.Fatal("obsolete/disabled editor emitted", calls)
	}
	n, err = p.View()
	if err != nil || !n.Disabled || !helperNode(t, n, FontSizeID(p.ID)).Disabled {
		t.Fatal("read-only view active", err)
	}
}

func TestFontPickerValidation(t *testing.T) {
	invalids := []func(*FontPicker){
		func(p *FontPicker) { p.ID = "" },
		func(p *FontPicker) { p.Families = nil },
		func(p *FontPicker) { p.Families = append(p.Families, p.Families[0]) },
		func(p *FontPicker) { p.Families[1].Name = p.Families[0].Name },
		func(p *FontPicker) { p.Families[0].Typefaces = nil },
		func(p *FontPicker) { p.Families[0].Typefaces[1].ID = "regular" },
		func(p *FontPicker) { p.Families[0].Typefaces[1].Name = "Regular" },
		func(p *FontPicker) { p.Families[0].Typefaces[0].Weight = 150 },
		func(p *FontPicker) { p.Families[0].Name = "Bad\x00Name" },
		func(p *FontPicker) { p.Value.FamilyID = "absent" },
		func(p *FontPicker) { p.Value.TypefaceID = "absent" },
		func(p *FontPicker) { p.Value.Size = float32(math.NaN()) },
		func(p *FontPicker) { p.Value.Size = 145 },
		func(p *FontPicker) { p.Preview = strings.Repeat("x", 1025) },
		func(p *FontPicker) { p.Preview = "\xff" },
		func(p *FontPicker) { p.FieldStyle.MinSize = Some(Size{Width: -1}) },
	}
	for i, mutate := range invalids {
		p := testFontPicker()
		mutate(p)
		if _, err := p.View(); err == nil {
			t.Fatalf("invalid case %d admitted", i)
		}
	}
	var nilPicker *FontPicker
	if _, err := nilPicker.View(); err == nil {
		t.Fatal("nil picker admitted")
	}
}

func TestFontPickerCompactFieldStyle(t *testing.T) {
	p := testFontPicker()
	p.FieldStyle = MacTheme(true).Kinds[TextField]
	n, err := p.View()
	if err != nil {
		t.Fatal(err)
	}
	field := helperNode(t, n, FontFamilyID(p.ID))
	if field.Style.Padding.Value != (Insets{}) || field.Style.Background != p.FieldStyle.Background || field.Style.Border != p.FieldStyle.Border {
		t.Fatal("input theme lost or doubled outer insets", field.Style)
	}
	size := helperNode(t, n, FontSizeID(p.ID))
	if size.NumberButtons != NumberButtonsStepper || size.Style.MaxSize.Value.Width != 88 {
		t.Fatal("font size lost compact stepper affordance", size)
	}
}
