//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/icons"
	"github.com/dshills/vellum/ui"
	"image/color"
	"testing"
)

func TestMacSourceListFlatRowsAndFocusAppearance(t *testing.T) {
	for _, dark := range []bool{false, true} {
		name := "light"
		if dark {
			name = "dark"
		}
		t.Run(name, func(t *testing.T) {
			selected := []string{"a", "b"}
			view := func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "list", Sequence: 1}, Children: []ui.Node{{ID: "group", Kind: ui.FormGroup, Text: "Objects", Children: []ui.Node{{ID: "list", Kind: ui.ListBox, Multiple: true, SelectedIDs: selected, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}, {ID: "c", Label: "Charlie"}}, SelectionChange: func(e ui.ListSelectionEvent) { selected = e.Apply(selected) }}}}, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
			}
			cfg := manualConfig(t, 936)
			cfg.Theme = ui.MacTheme(dark)
			cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			bg := color.NRGBA{R: 246, G: 246, B: 248, A: 255}
			if dark {
				bg = color.NRGBA{R: 29, G: 29, B: 31, A: 255}
			}
			cfg.Theme.Default.Background = ui.Some(bg)
			m, _ := simpleManual(t, cfg, view)
			unitSettle(t, m)
			root := m.implementation.find("one").state.root
			list := findRecord(root, "list")
			a, b, c := list.control.choice.buttons["a"], list.control.choice.buttons["b"], list.control.choice.buttons["c"]
			for _, row := range list.control.choice.buttons {
				if row.IconOn != icons.None || row.IconOff != icons.None {
					t.Fatal("source row retains checkbox/segment indicator")
				}
			}
			if b.Styles.Border.Width.Top.Dots != 0 || c.Styles.Border.Width.Top.Dots != 0 {
				t.Fatal("unfocused source row has push-button outline")
			}
			if a.Styles.Border.Width.Top.Dots == 0 {
				t.Fatal("focused row lost keyboard focus outline")
			}
			if list.widget.AsWidget().Geom.TotalBBox.Dx() < findRecord(root, "group").control.chrome.host.Geom.ContentBBox.Dx()-2 {
				t.Fatal("Mac list does not fill its allocated form width")
			}
			frame, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			for _, id := range []string{"a", "b"} {
				box := list.control.choice.buttons[id].Geom.ContentBBox
				got := frame.RGBAAt(box.Max.X-12, (box.Min.Y+box.Max.Y)/2)
				if got != (color.RGBA{G: 112, B: 235, A: 255}) {
					t.Fatal("active selection raster", id, got)
				}
			}
			snapshot, err := m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: snapshot.Windows[0].Incarnation, Type: events.KeyChord, Code: key.CodeTab}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			for _, id := range []string{"a", "b"} {
				row := list.control.choice.buttons[id]
				want := color.NRGBA{R: 160, G: 160, B: 165, A: 100}
				got := color.NRGBAModel.Convert(row.Styles.Background.At(0, 0)).(color.NRGBA)
				if got != want {
					t.Fatal("selected row did not become inactive on focus departure", id, got)
				}
			}
			frame, err = m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			field := cfg.Theme.Kinds[ui.ListBox].Background.Value
			blend := func(v, under uint8) uint8 { return uint8((int(v)*100 + int(under)*155 + 127) / 255) }
			expected := color.RGBA{R: blend(160, field.R), G: blend(160, field.G), B: blend(165, field.B), A: 255}
			for _, id := range []string{"a", "b"} {
				box := list.control.choice.buttons[id].Geom.ContentBBox
				got := frame.RGBAAt(box.Max.X-12, (box.Min.Y+box.Max.Y)/2)
				close := func(a, b uint8) bool { return int(a)-int(b) >= -1 && int(a)-int(b) <= 1 }
				if !close(got.R, expected.R) || !close(got.G, expected.G) || !close(got.B, expected.B) || got.A != 255 {
					t.Fatal("inactive selection raster", id, got, expected)
				}
			}
			if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: snapshot.Windows[0].Incarnation, Type: events.KeyChord, Code: key.CodeTab, Modifiers: 1 << key.Shift}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			for _, id := range []string{"a", "b"} {
				row := list.control.choice.buttons[id]
				want := color.NRGBA{G: 112, B: 235, A: 255}
				got := color.NRGBAModel.Convert(row.Styles.Background.At(0, 0)).(color.NRGBA)
				if got != want {
					t.Fatal("selected row did not regain active selection", id, got)
				}
			}
		})
	}
}

func TestSourceListModeChangesRestoreSingleChoiceAppearance(t *testing.T) {
	multiple := true
	view := func() ui.Node {
		n := ui.Node{ID: "list", Kind: ui.ListBox, Options: []ui.Option{{ID: "a", Label: "Alpha"}, {ID: "b", Label: "Beta"}}, Focus: ui.FocusRequest{ID: "list", Sequence: 1}}
		if multiple {
			n.Multiple = true
			n.SelectedIDs = []string{"a"}
			n.SelectionChange = func(ui.ListSelectionEvent) {}
		} else {
			n.SelectedID = "a"
			n.Change = func(ui.ChangeEvent) {}
		}
		return n
	}
	m, a := simpleManual(t, manualConfig(t, 937), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	original := root.control.choice.buttons["b"]
	if err := a.Do(func() { multiple = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	row := m.implementation.find("one").state.root.control.choice.buttons["b"]
	if row != original || row.IconOn != icons.Check || row.Styles.Border.Width.Top.Dots == 0 {
		t.Fatal("single choice lost original native presentation")
	}
	if m.implementation.find("one").state.root.control.choice.buttons["a"].Styles.IconSize.X.Value == 0 {
		t.Fatal("selected single-choice row lost check indicator")
	}
	if err := a.Do(func() { multiple = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	row = m.implementation.find("one").state.root.control.choice.buttons["b"]
	if row != original || row.IconOn != icons.None || row.Styles.Border.Width.Top.Dots != 0 {
		t.Fatal("multiple choice did not regain flat source presentation")
	}
}
