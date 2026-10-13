//go:build offscreen

package cogent

import (
	"bytes"
	"image/png"
	"os"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/ui"
)

func TestExpandedControlsManualFrameAndInput(t *testing.T) {
	enabled, pinned := false, false
	view := func() ui.Node {
		fraction := float32(0)
		if enabled {
			fraction += 0.5
		}
		if pinned {
			fraction += 0.5
		}
		return ui.Node{ID: "root", Kind: ui.Column, Focus: ui.FocusRequest{ID: "switch", Sequence: 1}, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(12)), Gap: ui.Some(float32(10))}, Children: []ui.Node{
			{ID: "title", Kind: ui.Label, Text: "Expanded Vellum controls"},
			{ID: "switch", Kind: ui.Switch, Text: "Enable snapping", Checked: enabled, Change: func(e ui.ChangeEvent) { enabled = e.Checked }},
			{ID: "toggle", Kind: ui.ToggleButton, Text: "Pin panel", Checked: pinned, Change: func(e ui.ChangeEvent) { pinned = e.Checked }},
			{ID: "separator", Kind: ui.Separator},
			{ID: "progress", Kind: ui.ProgressBar, Number: fraction, Accessibility: ui.Accessibility{Name: "Operation progress"}},
		}}
	}
	m, _ := simpleManual(t, manualConfig(t, 7), view)
	unitSettle(t, m)
	before, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w := snapshot.Windows[0]
	if w.FocusedID != "switch" {
		t.Fatal("switch did not accept managed focus", w.FocusedID)
	}
	if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Type: events.KeyChord, Code: key.CodeReturnEnter}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if !enabled {
		t.Fatal("actual offscreen keyboard route did not request switch change")
	}
	snapshot, err = m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	w = snapshot.Windows[0]
	var target ManualElement
	for _, e := range w.Elements {
		if e.ID == "toggle" {
			target = e
		}
	}
	if target.Bounds.Empty() {
		t.Fatal("toggle has no measured geometry")
	}
	for _, typ := range []events.Types{events.MouseDown, events.MouseUp} {
		if err := m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: w.Incarnation, Target: "toggle", TargetIncarnation: target.Incarnation, Type: typ, Button: events.Left}); err != nil {
			t.Fatal(err)
		}
	}
	unitSettle(t, m)
	if !pinned {
		t.Fatal("actual offscreen pointer route did not request toggle change")
	}
	after, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(before.Pix, after.Pix) {
		t.Fatal("completed control change produced no raster update")
	}
	snapshot, err = m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	var progress ManualElement
	for _, e := range snapshot.Windows[0].Elements {
		if e.ID == "progress" {
			progress = e
		}
	}
	if progress.Value != float32(1) || progress.Bounds.Empty() {
		t.Fatal("completed progress missing fraction/geometry", progress)
	}
	changed := 0
	for y := progress.Bounds.Min.Y; y < progress.Bounds.Max.Y; y++ {
		for x := progress.Bounds.Min.X; x < progress.Bounds.Max.X; x++ {
			if before.RGBAAt(x, y) != after.RGBAAt(x, y) {
				changed++
			}
		}
	}
	if changed == 0 {
		t.Fatal("progress fraction did not change rendered bar")
	}
	if path := os.Getenv("VELLUM_CONTROLS_CAPTURE"); path != "" {
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, after)
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
	}
}

func TestExpandedRejectedToggleRestoresVisibleStyle(t *testing.T) {
	requests := 0
	a, _ := newAXTest(t, 8)
	host := core.NewFrame(core.NewBody("rejected toggle"))
	n := ui.Node{ID: "toggle", Kind: ui.ToggleButton, Text: "Rejected toggle", Checked: false, Change: func(ui.ChangeEvent) { requests++ }}
	r, err := reconcile(host, nil, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	w := r.widget.(*core.Switch)
	w.Style()
	before := w.Styles.Background
	w.Send(events.Click)
	if !w.IsChecked() {
		t.Fatal("actual switch click did not toggle native state")
	}
	// Complete the backend's post-event restyle while the owner callback is
	// still queued. This fixture uses the real Switch event/Styler paths.
	w.Style()
	if w.Styles.Background == before {
		t.Fatal("fixture did not expose checked native background")
	}
	a.Drain()
	if requests != 1 {
		t.Fatal("routed toggle request missing", requests)
	}
	r, err = reconcile(host, r, n, a, "ax")
	if err != nil {
		t.Fatal(err)
	}
	if w.IsChecked() || w.Styles.Background != before {
		t.Fatal("rejected toggle retained selected background", w.Styles.Background, before)
	}
}
