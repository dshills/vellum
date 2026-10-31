//go:build offscreen

package cogent

import (
	"cogentcore.org/core/events"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"image"
	"testing"
)

// Route actual pointer events instead of setting widget focus directly.
func TestPointerFocusMovesBetweenTitleAndDocument(t *testing.T) {
	for _, prose := range []bool{false, true} {
		name := "code"
		if prose {
			name = "prose"
		}
		t.Run(name, func(t *testing.T) {
			model, err := editor.New("Synthetic document body", editor.Config{})
			if err != nil {
				t.Fatal(err)
			}
			m, a := simpleManual(t, manualConfig(t, 7), func() ui.Node {
				return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
					{ID: "title", Kind: ui.TextField, Text: "Synthetic title"},
					{ID: "body", Kind: ui.MultilineText, Editor: model, Prose: prose, SoftWrap: true, Grow: true},
				}}
			})
			unitSettle(t, m)
			w := m.implementation.find("one")
			if w == nil {
				t.Fatal("window one not found")
			}
			for cycle := 0; cycle < 3; cycle++ {
				for _, id := range []string{"body", "title"} {
					b := findRecord(w.state.root, id).widget.AsWidget().Geom.TotalBBox
					p := b.Min.Add(image.Pt(12, 12))
					if !p.In(b) {
						t.Fatalf("unusable %s bounds: %v", id, b)
					}
					for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
						if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.incarnation, Type: kind, Button: events.Left, Position: p}); err != nil {
							t.Fatal(err)
						}
						unitSettle(t, m)
					}
					snap, err := m.Snapshot(testContext)
					if err != nil {
						t.Fatal(err)
					}
					if len(snap.Windows) != 1 {
						t.Fatalf("expected one window; got %d", len(snap.Windows))
					}
					if snap.Windows[0].FocusedID != id {
						t.Fatalf("cycle %d clicked %s; focused %q", cycle, id, snap.Windows[0].FocusedID)
					}
					var invalidateErr error
					if err := a.Do(func() { invalidateErr = a.Invalidate("one") }); err != nil {
						t.Fatal(err)
					}
					if invalidateErr != nil {
						t.Fatal(invalidateErr)
					}
					unitSettle(t, m)
					snap, err = m.Snapshot(testContext)
					if err != nil {
						t.Fatal(err)
					}
					if len(snap.Windows) != 1 {
						t.Fatalf("expected one window; got %d", len(snap.Windows))
					}
					if snap.Windows[0].FocusedID != id {
						t.Fatalf("refresh lost %s focus: %q", id, snap.Windows[0].FocusedID)
					}
				}
			}
		})
	}
}
