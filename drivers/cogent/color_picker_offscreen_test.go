//go:build offscreen

package cogent

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

var colorPickerCapture = flag.String("vellum-color-picker-capture", "", "optional new absolute output PNG for explicit visual verification")

func TestColorPickerManagedKeyboardReconcileAndPixels(t *testing.T) {
	picker := &ui.ColorPicker{ID: "ink", Label: "Ink", Color: color.NRGBA{R: 64, G: 128, B: 192, A: 128}}
	calls := 0
	picker.Change = func(e ui.ColorEdit) { calls++; picker.Color = e.Apply(picker.Color) }
	focus := ui.FocusRequest{ID: ui.ColorFieldID("ink", ui.ColorRed), Sequence: 1}
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "one", Title: "Color picker", Width: 400, Height: 460, View: func() ui.Node { n := picker.View(); n.Focus = focus; return n }}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManual(testContext, a, manualConfig(t, 7))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close(testContext)
	unitSettle(t, m)
	before, red := imageManualElement(t, m, ui.ColorFieldID("ink", ui.ColorRed))
	_, preview := imageManualElement(t, m, "ink/color/preview")
	if before.FocusedID != red.ID || red.Value != float32(64) || red.Bounds.Empty() || preview.Bounds.Size() != image.Pt(32, 32) {
		pr := findRecord(m.implementation.find("one").state.root, "ink/color/preview").widget.AsWidget()
		t.Log("native preview geom/styles", pr.Geom, pr.Styles.Min, pr.Styles.Max)
		for _, id := range []string{"ink", "ink/color/preview-row", "ink/color/value"} {
			w := findRecord(m.implementation.find("one").state.root, id).widget.AsWidget()
			t.Log("ancestor", id, w.Geom, w.Styles.Grow, w.Styles.Align, w.Styles.Justify)
		}
		t.Fatal("component geometry/focus missing", before.FocusedID, red, preview)
	}
	capture, err := m.Capture(testContext, "one")
	if err != nil {
		t.Fatal(err)
	}
	point := preview.Bounds.Min.Add(image.Pt(4, 4))
	if got := capture.RGBAAt(point.X, point.Y); got != (color.RGBA{R: 152, G: 184, B: 216, A: 255}) {
		t.Fatal("completed independent alpha/checker pixel", got)
	}
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: before.Incarnation, Type: events.KeyChord, Code: key.CodeRightArrow}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	if calls != 1 || picker.Color != (color.NRGBA{R: 65, G: 128, B: 192, A: 128}) {
		t.Fatal("actual slider key did not deliver color patch", calls, picker.Color)
	}
	after, redAfter := imageManualElement(t, m, red.ID)
	_, hex := imageManualElement(t, m, ui.ColorFieldID("ink", ui.ColorHex))
	if redAfter.Incarnation != red.Incarnation || redAfter.Value != float32(65) || hex.Value != "#4180C080" {
		t.Fatal("keyed field or canonical hex not retained", redAfter, hex)
	}
	// Native refused/normalized commits still reapply accepted values through
	// the ordinary managed controls; the component owns no hidden input widget.
	picker.Change = func(ui.ColorEdit) { calls++ }
	if err = m.Input(testContext, ManualInput{Window: "one", WindowIncarnation: after.Incarnation, Type: events.KeyChord, Code: key.CodeRightArrow}); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	_, restored := imageManualElement(t, m, red.ID)
	if calls != 2 || picker.Color.R != 65 || restored.Value != float32(65) || findRecord(m.implementation.find("one").state.root, red.ID).widget.(*core.Slider).Value != 65 {
		t.Fatal("rejected edit failed to restore accepted channel", picker.Color, restored)
	}
	if err = a.Do(func() { picker.Disabled = true; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	unitSettle(t, m)
	snapshot, err := m.Snapshot(testContext)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range snapshot.Windows[0].Elements {
		if e.ID == hex.ID && e.Enabled {
			t.Fatal("disabled component retains editable hex")
		}
	}
	if *colorPickerCapture != "" {
		if !filepath.IsAbs(*colorPickerCapture) {
			t.Fatal("capture path must be absolute")
		}
		frame, err := m.Capture(testContext, "one")
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.OpenFile(*colorPickerCapture, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			t.Fatal(err)
		}
		if err = png.Encode(f, frame); err != nil {
			_ = f.Close()
			t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
