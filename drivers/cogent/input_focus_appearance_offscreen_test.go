//go:build offscreen

package cogent

import (
	"image/color"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"github.com/dshills/vellum/ui"
)

func TestMacInputFocusOutlineAndFlatInterior(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, kind := range []string{"combo", "secure", "plain"} {
			t.Run(kind+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				cfg := manualConfig(t, 950)
				cfg.Theme = ui.MacTheme(dark)
				cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 246, G: 246, B: 248, A: 255})
				if dark {
					cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 29, G: 29, B: 31, A: 255})
				}
				style := cfg.Theme.Kinds[ui.TextField]
				style.Padding = ui.Some(ui.Insets{})
				target := "plain"
				var render func() (ui.Node, error)
				switch kind {
				case "combo":
					p := &ui.ComboBox{ID: "combo", Label: "Units", Text: "Millimeters", Options: []ui.Option{{ID: "mm", Label: "Millimeters"}}, Style: style, Change: func(ui.ComboEvent) {}}
					render, target = p.View, ui.ComboInputID("combo")
				case "secure":
					p := &ui.SecureField{ID: "secure", Label: "Synthetic password", Text: "synthetic", AllowReveal: true, Style: style, Change: func(ui.ChangeEvent) {}}
					render, target = p.View, ui.SecureInputID("secure")
				default:
					render = func() (ui.Node, error) {
						return ui.Node{ID: "plain", Kind: ui.TextField, Text: "Drawing", Change: func(ui.ChangeEvent) {}}, nil
					}
				}
				view := func() ui.Node {
					n, err := render()
					if err != nil {
						panic(err)
					}
					return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: target, Sequence: 1}, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(12))}, Children: []ui.Node{n, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
				}
				m, _ := simpleManual(t, cfg, view)
				unitSettle(t, m)
				root := m.implementation.find("one").state.root
				outer := findRecord(root, kind).widget.AsWidget()
				field := findRecord(root, target).widget.AsWidget()
				frame, err := m.Capture(testContext, "one")
				if err != nil {
					t.Fatal(err)
				}
				interior := field.Geom.ContentBBox
				got := frame.RGBAAt(interior.Max.X-12, (interior.Min.Y+interior.Max.Y)/2)
				want := color.RGBAModel.Convert(cfg.Theme.Kinds[ui.TextField].Background.Value).(color.RGBA)
				if got != want {
					t.Fatal("focused input has an inner state-layer fill", got, want)
				}
				if outer.Styles.Border.Width.Top.Value < 2 || outer.Styles.Border.Width.Bottom.Value != outer.Styles.Border.Width.Top.Value {
					t.Fatal("focus is not a full outer outline", outer.Styles.Border.Width)
				}
				if kind != "plain" && (field.Styles.Border.Width.Top.Value != 0 || field.Styles.Border.Width.Bottom.Value != 0) {
					t.Fatal("embedded field kept a private focus border")
				}
				bounds := outer.Geom.TotalBBox
				w, _ := imageManualElement(t, m, target)
				tabs := 1
				if kind != "plain" {
					tabs = 2
				}
				for step := range tabs {
					if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeTab}); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
					if kind != "plain" && step == 0 && outer.Styles.Border.Width.Top.Value != 2 {
						t.Fatal("accessory focus lost the shared outline")
					}
				}
				if !findRecord(root, "after").widget.AsWidget().ContainsFocus() {
					t.Fatal("Tab did not leave input")
				}
				if outer.Styles.Border.Width.Top.Value != 1 || outer.Styles.Border.Width.Bottom.Value != 1 {
					t.Fatal("outer outline did not restore after focus departure", outer.Styles.Border.Width)
				}
				if outer.Geom.TotalBBox != bounds {
					t.Fatal("focus changed input allocation", bounds, outer.Geom.TotalBBox)
				}
			})
		}
	}
}

func TestSharedInputOutlineSurvivesReparentingAndStyleRemoval(t *testing.T) {
	cfg := manualConfig(t, 951)
	moved, styled := false, true
	text := "Synthetic"
	calls := 0
	focus := ui.FocusRequest{ID: "field", Sequence: 1}
	clear := ui.Style{Border: ui.Some(ui.Border{}), Background: ui.Some(color.NRGBA{}), MinSize: ui.Some(ui.Size{Height: 26}), MaxSize: ui.Some(ui.Size{}), GrowX: ui.Some(float32(1))}
	outline := ui.Style{Border: ui.Some(ui.Border{Width: 1, Color: color.NRGBA{R: 191, G: 191, B: 195, A: 255}}), Background: ui.Some(color.NRGBA{R: 255, G: 255, B: 255, A: 255}), CornerRadius: ui.Some(float32(5))}
	view := func() ui.Node {
		field := ui.Node{ID: "field", Kind: ui.TextField, Text: text, Change: func(e ui.ChangeEvent) { calls++; text = e.Text }}
		if styled {
			field.Style = clear
		}
		left := ui.Node{ID: "left", Kind: ui.Row, Style: outline}
		right := ui.Node{ID: "right", Kind: ui.Row, Style: outline}
		if moved {
			right.Children = []ui.Node{field}
		} else {
			left.Children = []ui.Node{field}
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: focus, Children: []ui.Node{left, right, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
	}
	m, a := simpleManual(t, cfg, view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	original := findRecord(root, "field").widget.(*core.TextField)
	if err := a.Do(func() { moved = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	root = m.implementation.find("one").state.root
	field := findRecord(root, "field").widget.(*core.TextField)
	if field != original || !field.ContainsFocus() || field.Styles.RenderBox || findRecord(root, "left").widget.AsWidget().Styles.Border.Width.Top.Value != 1 || findRecord(root, "right").widget.AsWidget().Styles.Border.Width.Top.Value != 2 {
		t.Fatal("reparenting lost shared focus ownership")
	}
	if err := a.Do(func() { styled = false; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	root = m.implementation.find("one").state.root
	field = findRecord(root, "field").widget.(*core.TextField)
	if field != original || !field.Styles.RenderBox || field.Styles.Border.Width.Top.Value != 0 || field.Styles.Border.Width.Bottom.Value != 2 || findRecord(root, "right").widget.AsWidget().Styles.Border.Width.Top.Value != 1 {
		t.Fatal("removing explicit editor style did not restore native focus defaults", field.Styles.Border.Width)
	}
	w, _ := imageManualElement(t, m, "field")
	for _, input := range []ManualInput{{Type: events.KeyChord, Code: key.CodeX, Rune: 'x'}, {Type: events.KeyChord, Code: key.CodeReturnEnter}} {
		input.Window = "one"
		input.WindowIncarnation = w.Incarnation
		if err := m.Input(testContext, input); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
	}
	if calls != 1 {
		t.Fatal("style/reparent transition changed normal edit delivery", calls)
	}
}

func TestInputOutlineStyleWaitsForOriginRefresh(t *testing.T) {
	p := &ui.ComboBox{ID: "combo", Style: ui.Style{Border: ui.Some(ui.Border{Width: 1, Color: color.NRGBA{A: 255}})}, Change: func(ui.ComboEvent) {}}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "after", Sequence: 1}, Children: []ui.Node{n, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
	}
	m, a := simpleManual(t, manualConfig(t, 952), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	parent := findRecord(root, "combo").widget.AsWidget()
	field := findRecord(root, ui.ComboInputID("combo"))
	passes := 0
	parent.FinalStyler(func(*styles.Style) { passes++ })
	// Model-owner callbacks can be drained from another window. They must only
	// request refresh; they cannot run this native parent's stylers directly.
	field.widget.AsWidget().SetFocus()
	originPasses := passes
	a.Drain()
	if passes != originPasses {
		t.Fatal("owner queue directly restyled an originating native parent", passes)
	}
	unitSettle(t, m)
	if parent.Styles.Border.Width.Top.Value != 2 || !field.widget.AsWidget().ContainsFocus() {
		t.Fatal("origin refresh did not consume pending outline styling", passes)
	}
}

func TestOrdinaryInputFocusKeepsSingleAccessibilityRefresh(t *testing.T) {
	views := 0
	view := func() ui.Node {
		views++
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "field", Sequence: 1}, Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "Synthetic"}}}
	}
	m, _ := simpleManual(t, manualConfig(t, 953), view)
	unitSettle(t, m)
	baseline := views
	field := findRecord(m.implementation.find("one").state.root, "field")
	field.widget.AsWidget().Send(events.Focus)
	unitSettle(t, m)
	if views != baseline+1 {
		t.Fatal("ordinary focus added work beyond the existing accessibility refresh", baseline, views)
	}
}

func TestInputOutlineQueueRefusalStillRefreshes(t *testing.T) {
	p := &ui.ComboBox{ID: "combo", Style: ui.Style{Border: ui.Some(ui.Border{Width: 1, Color: color.NRGBA{A: 255}})}, Change: func(ui.ComboEvent) {}}
	view := func() ui.Node {
		n, err := p.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "after", Sequence: 1}, Children: []ui.Node{n, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
	}
	m, a := simpleManual(t, manualConfig(t, 954), view)
	unitSettle(t, m)
	field := findRecord(m.implementation.find("one").state.root, ui.ComboInputID("combo"))
	for i := 0; i < 4; i++ {
		if err := a.Post(func() {}); err != nil {
			t.Fatal(err)
		}
	}
	field.widget.AsWidget().SetFocus()
	if field.control.refused.Load() != nil {
		t.Fatal("passive queue refusal lost the wake or became an edit error")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !field.widget.AsWidget().ContainsFocus() || findRecord(m.implementation.find("one").state.root, "combo").widget.AsWidget().Styles.Border.Width.Top.Value != 2 {
		t.Fatal("full owner queue stranded focus styling")
	}
}

func TestMacNumberFieldFocusKeepsThemeFill(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, buttons := range []ui.NumberFieldButtons{"", ui.NumberButtonsStepper} {
			mode := "inline"
			if buttons == ui.NumberButtonsStepper {
				mode = "stepper"
			}
			t.Run(mode+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				cfg := manualConfig(t, 955)
				cfg.Theme = ui.MacTheme(dark)
				cfg.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
				cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 246, G: 246, B: 248, A: 255})
				if dark {
					cfg.Theme.Default.Background = ui.Some(color.NRGBA{R: 29, G: 29, B: 31, A: 255})
				}
				value := float32(1)
				calls := 0
				width := float32(160)
				if buttons == ui.NumberButtonsStepper {
					width = 104
				}
				view := func() ui.Node {
					return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: "number", Sequence: 1}, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(12))}, Children: []ui.Node{{ID: "number", Kind: ui.NumberField, Number: value, NumberButtons: buttons, Numeric: ui.Numeric{Min: 0, Max: 3, Step: 1}, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: width, Height: 28}), MaxSize: ui.Some(ui.Size{Width: width})}, Change: func(e ui.ChangeEvent) { calls++; value = e.Number }}, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
				}
				m, _ := simpleManual(t, cfg, view)
				unitSettle(t, m)
				field := findRecord(m.implementation.find("one").state.root, "number").widget.AsWidget()
				original := field.Geom.TotalBBox
				frame, err := m.Capture(testContext, "one")
				if err != nil {
					t.Fatal(err)
				}
				inside := field.Geom.ContentBBox
				got := frame.RGBAAt((inside.Min.X+inside.Max.X)/2, (inside.Min.Y+inside.Max.Y)/2)
				want := color.RGBAModel.Convert(cfg.Theme.Kinds[ui.NumberField].Background.Value).(color.RGBA)
				if got != want {
					t.Fatal("focused number field changed its theme fill", got, want)
				}
				w, _ := imageManualElement(t, m, "number")
				for _, code := range []key.Codes{key.CodeUpArrow, key.CodeDownArrow} {
					if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: code}); err != nil {
						t.Fatal(err)
					}
					unitSettle(t, m)
				}
				if calls != 2 || value != 1 {
					t.Fatal("focus styling changed numeric stepping", calls, value)
				}
				if field.Geom.TotalBBox != original {
					t.Fatal("focus styling changed number field allocation")
				}
			})
		}
	}
}

func TestBorderlessOrdinaryFieldRetainsNativeFocusIndicator(t *testing.T) {
	for _, fill := range []string{"theme-fill", "explicit-fill", "transparent"} {
		t.Run(fill, func(t *testing.T) {
			cfg := manualConfig(t, 957)
			style := ui.Style{Border: ui.Some(ui.Border{}), MinSize: ui.Some(ui.Size{Width: 220, Height: 28})}
			if fill == "explicit-fill" {
				style.Background = ui.Some(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			} else if fill == "transparent" {
				style.Background = ui.Some(color.NRGBA{})
			}
			view := func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "field", Sequence: 1}, Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "Synthetic", Style: style}, {ID: "after", Kind: ui.Button, Text: "After", Activate: func() {}}}}
			}
			m, _ := simpleManual(t, cfg, view)
			unitSettle(t, m)
			field := findRecord(m.implementation.find("one").state.root, "field").widget.AsWidget()
			if field.Styles.Border.Width.Bottom.Value != 2 || field.Styles.MaxBorder.Width.Bottom.Value != 2 || !field.Styles.RenderBox {
				t.Fatal("zero resting border removed ordinary field focus cue", field.Styles.Border.Width, field.Styles.MaxBorder.Width, field.Styles.RenderBox)
			}
			bounds := field.Geom.TotalBBox
			w, _ := imageManualElement(t, m, "field")
			for _, mods := range []key.Modifiers{0, 1 << key.Shift} {
				if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeTab, Modifiers: mods}); err != nil {
					t.Fatal(err)
				}
				unitSettle(t, m)
				focused := mods != 0
				if field.ContainsFocus() != focused || field.Geom.TotalBBox != bounds {
					t.Fatal("focus/allocation changed unexpectedly")
				}
				want := float32(0)
				if focused {
					want = 2
				}
				if field.Styles.Border.Width.Bottom.Value != want {
					t.Fatal("focus cue did not follow Tab", field.Styles.Border.Width)
				}
			}
		})
	}
}

func TestBorderlessUnthemedFieldRetainsNativeFocusIndicator(t *testing.T) {
	a, _ := newAXTest(t, 958)
	host := core.NewFrame(core.NewBody("Borderless focus"))
	r, err := reconcile(host, nil, ui.Node{ID: "field", Kind: ui.TextField, Text: "Synthetic", Style: ui.Style{Border: ui.Some(ui.Border{})}}, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	field := r.widget.AsWidget()
	field.SetState(true, states.Focused)
	field.Style()
	if field.Styles.Border.Width.Bottom.Value != 2 || field.Styles.MaxBorder.Width.Bottom.Value != 2 || !field.Styles.RenderBox {
		t.Fatal("unthemed focus cue lost", field.Styles.Border.Width)
	}
	field.SetState(false, states.Focused)
	field.Style()
	if field.Styles.Border.Width.Bottom.Value != 0 {
		t.Fatal("unthemed resting border did not clear")
	}
}

func TestRemovingSharedInputBorderRestoresEditorFocusCue(t *testing.T) {
	bordered := true
	view := func() ui.Node {
		style := ui.Style{}
		if bordered {
			style.Border = ui.Some(ui.Border{Width: 1, Color: color.NRGBA{A: 255}})
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "field", Sequence: 1}, Children: []ui.Node{{ID: "composite", Kind: ui.Row, Style: style, Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "Synthetic", Style: ui.Style{Border: ui.Some(ui.Border{}), Background: ui.Some(color.NRGBA{})}}}}}}
	}
	m, a := simpleManual(t, manualConfig(t, 959), view)
	unitSettle(t, m)
	root := m.implementation.find("one").state.root
	original := findRecord(root, "field").widget
	for _, next := range []bool{false, true} {
		if err := a.Do(func() { bordered = next; _ = a.Invalidate("one") }); err != nil {
			t.Fatal(err)
		}
		unitSettle(t, m)
		root = m.implementation.find("one").state.root
		field := findRecord(root, "field").widget.AsWidget()
		outer := findRecord(root, "composite").widget.AsWidget()
		if field.This != original || !field.ContainsFocus() {
			t.Fatal("border change replaced editor or lost focus")
		}
		if next {
			if field.Styles.RenderBox || field.Styles.Border.Width.Bottom.Value != 0 || outer.Styles.Border.Width.Top.Value != 2 {
				t.Fatal("restored composite did not own exactly one focus outline")
			}
		} else if !field.Styles.RenderBox || field.Styles.Border.Width.Bottom.Value != 2 {
			t.Fatal("removing composite border removed all keyboard focus cues")
		}
	}
}
