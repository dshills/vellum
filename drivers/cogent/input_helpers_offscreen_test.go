//go:build offscreen

package cogent

import (
	"bytes"
	"fmt"
	"image/color"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func TestMacInputHelpersActualRoutingAndRetention(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprintf("dark-%v", dark), func(t *testing.T) {
			search := &ui.SearchBox{ID: "search", Label: "Search drawings", Text: "Cube", Placeholder: "Find a drawing"}
			calls := 0
			reject := true
			search.Change = func(e ui.SearchEvent) {
				calls++
				if !reject {
					search.Text = e.Text
				}
			}
			view := func() ui.Node {
				children := []ui.Node{}
				n, err := search.View()
				if err != nil {
					panic(err)
				}
				children = append(children, n)
				field, err := (ui.FormField{ID: "name-group", Label: "Drawing name", Required: true, Help: "Use a unique name", ValidationMessage: "Already used",
					Content: ui.Node{ID: "name", Kind: ui.TextField, Text: "Cube"}}).View()
				if err != nil {
					panic(err)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Children: append(children, field)}
			}
			config := manualConfig(t, 7)
			config.Theme = ui.MacTheme(dark)
			config.Theme.Default.Background = ui.Some(color.NRGBA{R: 246, G: 246, B: 248, A: 255})
			if dark {
				config.Theme.Default.Background = ui.Some(color.NRGBA{R: 29, G: 29, B: 31, A: 255})
			}
			// Deterministic tests use bundled fonts; actual macOS typography is checked live.
			config.Theme.Default.Typography = ui.Some(ui.Typography{Family: "Noto Sans", Size: 13})
			m, a := simpleManual(t, config, view)
			unitSettle(t, m)
			w, clear := imageManualElement(t, m, ui.SearchClearID(search.ID))
			before, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			click := func(target ManualElement) {
				t.Helper()
				for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
					if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Target: target.ID, TargetIncarnation: target.Incarnation, Type: typ, Button: events.Left}); err != nil {
						t.Fatal(err)
					}
				}
				unitSettle(t, m)
			}
			click(clear)
			if calls != 1 || search.Text != "Cube" {
				t.Fatal("rejected clear changed accepted model", calls, search.Text)
			}
			_, retained := imageManualElement(t, m, ui.SearchClearID(search.ID))
			if retained.Incarnation != clear.Incarnation {
				t.Fatal("clear identity changed")
			}
			reject = false
			click(retained)
			if calls != 2 || search.Text != "" {
				t.Fatal("clear not admitted", calls, search.Text)
			}
			fieldRecord := findRecord(m.implementation.find("one").state.root, ui.SearchInputID(search.ID))
			field := fieldRecord.widget.(*core.TextField)
			if field.Text() != "" || field.Placeholder != "Find a drawing" {
				t.Fatal("accepted clear/placeholder not restored", field.Text(), field.Placeholder)
			}
			if _, err = m.Frame(testContext, "one"); err != nil {
				t.Fatal(err)
			}
			after, err := m.Capture(testContext, "one")
			if err != nil || bytes.Equal(before.Pix, after.Pix) {
				t.Fatal("clear did not change completed frame", err)
			}
			sem := findRecord(m.implementation.find("one").state.root, "name").widget.AsTree().Property(SemanticsProperty).(Semantics)
			if sem.Name != "Drawing name" || sem.Help != "Required. Use a unique name. Error: Already used" {
				t.Fatal("field label/error metadata", sem)
			}
			// Real managed focus and keyboard text route, followed by ordinary field commit.
			if err := a.Do(func() { search.Text = ""; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			_, inputTarget := imageManualElement(t, m, ui.SearchInputID(search.ID))
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.MouseDown, Target: ui.SearchInputID(search.ID), TargetIncarnation: inputTarget.Incarnation, Button: events.Left}); err != nil {
				t.Fatal(err)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.MouseUp, Target: ui.SearchInputID(search.ID), TargetIncarnation: inputTarget.Incarnation, Button: events.Left}); err != nil {
				t.Fatal(err)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeB, Rune: 'b'}); err != nil {
				t.Fatal(err)
			}
			if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			if search.Text != "b" {
				t.Fatal("keyboard edit not accepted", search.Text)
			}
		})
	}
}

func TestInputHelpersQueuedNativeRemovalAndDisable(t *testing.T) {
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("input helper retirement"))
	calls := 0
	search := &ui.SearchBox{ID: "search", Text: "accepted", Change: func(ui.SearchEvent) { calls++ }}
	notice := &ui.InlineNotice{ID: "notice", Text: "Try again", ActionLabel: "Retry", Activate: func() { calls++ }}
	view := func() ui.Node {
		s, err := search.View()
		if err != nil {
			panic(err)
		}
		n, err := notice.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{s, n}}
	}
	r, err := reconcile(host, nil, view(), a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	findRecord(r, ui.SearchInputID("search")).widget.(*core.TextField).SetText("queued").SendChange()
	findRecord(r, "notice/notice/action").widget.(*core.Button).Send(events.Click)
	empty := ui.Node{ID: "root", Kind: ui.Column}
	r, err = reconcile(host, r, empty, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if calls != 0 {
		t.Fatal("retired callbacks reached model", calls)
	}
	r, err = reconcile(host, r, view(), a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	findRecord(r, ui.SearchClearID("search")).widget.(*core.Button).Send(events.Click)
	search.Disabled = true
	r, err = reconcile(host, r, view(), a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if calls != 0 {
		t.Fatal("disabled queued clear reached model", calls)
	}
}
