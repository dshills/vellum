//go:build offscreen

package cogent

import (
	"bytes"
	"fmt"
	"html"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

func TestBreadcrumbManagedInputIdentityAndFrames(t *testing.T) {
	for _, dpi := range []float32{96, 120, 144, 192} {
		t.Run(fmt.Sprintf("dpi%.0f", dpi), func(t *testing.T) {
			b := &ui.Breadcrumb{ID: "path", Label: "Project location", Items: []ui.BreadcrumbItem{{ID: "home", Label: "Home"}, {ID: "folder/a", Label: "<b>A & B</b>"}, {ID: "current", Label: "Current"}}}
			calls := 0
			reject := true
			b.Navigate = func(e ui.BreadcrumbEvent) {
				calls++
				if !reject {
					for i, item := range b.Items {
						if item.ID == e.ItemID {
							b.Items = append([]ui.BreadcrumbItem(nil), b.Items[:i+1]...)
							break
						}
					}
				}
			}
			id := ui.BreadcrumbItemID(b.ID, "folder/a")
			view := func() ui.Node {
				n, err := b.View()
				if err != nil {
					panic(err)
				}
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: id, Sequence: 1}, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(12))}, Children: []ui.Node{n}}
			}
			config := manualConfig(t, 7)
			config.LogicalDPI = dpi
			a, err := app.New(app.Config{QueueCapacity: 8})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			if err = a.Open(app.Window{ID: "one", Title: "Breadcrumb four-DPI", Width: 720, Height: 240, View: view}); err != nil {
				t.Fatal(err)
			}
			m, err := NewManual(testContext, a, config)
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close(testContext)
			unitSettle(t, m)
			before, item := imageManualElement(t, m, id)
			_, current := imageManualElement(t, m, ui.BreadcrumbItemID(b.ID, "current"))
			if before.FocusedID != id || item.Bounds.Empty() || current.Bounds.Empty() || current.Kind != ui.Label {
				t.Fatal("path geometry/focus", before.FocusedID, item, current)
			}
			widget := findRecord(m.implementation.find("one").state.root, id).widget.(*core.Button)
			if widget.Text != html.EscapeString(b.Items[1].Label) {
				t.Fatal("markup interpreted", widget.Text)
			}
			frame, err := m.Capture(testContext, "one")
			if err != nil {
				t.Fatal(err)
			}
			if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			_, same := imageManualElement(t, m, id)
			if calls != 1 || len(b.Items) != 3 || same.Incarnation != item.Incarnation {
				t.Fatal("rejected navigation changed identity/path", calls, b.Items, same)
			}
			reject = false
			for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
				if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Target: id, TargetIncarnation: same.Incarnation, Type: typ, Button: events.Left}); err != nil {
					t.Fatal(err)
				}
			}
			unitSettle(t, m)
			_, now := imageManualElement(t, m, id)
			if calls != 2 || len(b.Items) != 2 || now.Kind != ui.Label || now.Incarnation == same.Incarnation {
				t.Fatal("accepted location not passive/current", calls, b.Items, now)
			}
			updated, err := m.Capture(testContext, "one")
			if err != nil || bytes.Equal(frame.Pix, updated.Pix) {
				t.Fatal("accepted navigation did not change completed frame", err)
			}
			if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Target: id, TargetIncarnation: same.Incarnation, Type: events.MouseDown, Button: events.Left}); err == nil {
				t.Fatal("retired button identity admitted")
			}
			if err = a.Do(func() { b.Navigate = nil; _ = a.Invalidate("one") }); err != nil {
				t.Fatal(err)
			}
			unitSettle(t, m)
			snap, err := m.Snapshot(testContext)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range snap.Windows[0].Elements {
				if e.Kind == ui.Button {
					t.Fatal("readonly path has managed button", e)
				}
			}
		})
	}
}

func TestBreadcrumbNarrowOverflowRevealsEnabledAncestors(t *testing.T) {
	b := &ui.Breadcrumb{ID: "path", Label: "Location", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 220})}, Navigate: func(ui.BreadcrumbEvent) {}}
	for i := range 6 {
		b.Items = append(b.Items, ui.BreadcrumbItem{ID: fmt.Sprintf("level-%d", i), Label: fmt.Sprintf("Location %d", i), Disabled: i == 4})
	}
	first := ui.BreadcrumbItemID(b.ID, "level-0")
	last := ui.BreadcrumbItemID(b.ID, "level-3")
	view := func() ui.Node {
		n, err := b.View()
		if err != nil {
			panic(err)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Focus: ui.FocusRequest{ID: first, Sequence: 1}, Children: []ui.Node{n}}
	}
	m, _ := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	before, _ := imageManualElement(t, m, first)
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Type: events.KeyChord, Code: key.CodeEnd}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	after, element := imageManualElement(t, m, last)
	toolbar := findRecord(m.implementation.find("one").state.root, b.ID).control.chrome.host
	if after.FocusedID != last || element.Bounds.Empty() || toolbar.Scrolls[math32.X] == nil || toolbar.Scrolls[math32.X].Value <= 0 || !element.Bounds.In(toolbar.Geom.TotalBBox) {
		t.Fatal("last enabled ancestor not revealed", after.FocusedID, element, toolbar.Geom.TotalBBox)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: after.Incarnation, Type: events.KeyChord, Code: key.CodeHome}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	final, _ := imageManualElement(t, m, first)
	if final.FocusedID != first || toolbar.Scrolls[math32.X].Value != 0 {
		t.Fatal("first ancestor not revealed", final.FocusedID)
	}
}
