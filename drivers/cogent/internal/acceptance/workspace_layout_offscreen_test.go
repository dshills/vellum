//go:build offscreen

package acceptance

import (
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/text/fonts"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"io/fs"
	"strings"
	"testing"
	"time"
)

func TestWorkspaceCompletedDensityGeometry(t *testing.T) {
	ctx := context.Background()
	clock, _ := execution.NewManualClock(time.Unix(0, 0), 256, 256)
	defer clock.Close()
	executor, _ := execution.NewManual(8)
	defer executor.Close()
	config := cogent.ManualConfig{Clock: clock, Executor: executor, LogicalDPI: 96, EventCapacity: 128, OwnerBatch: 16, MaxWindows: 1, Theme: ui.Theme{Default: ui.Style{Foreground: ui.Some(color.NRGBA{A: 255}), Background: ui.Some(color.NRGBA{R: 248, G: 248, B: 248, A: 255}), Typography: ui.Some(ui.Typography{Family: "Noto Sans", Size: 16})}}}
	if err := fs.WalkDir(fonts.Default, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(name, ".ttf") && !strings.HasSuffix(name, ".otf") {
			return nil
		}
		data, e := fs.ReadFile(fonts.Default, name)
		if e != nil {
			return e
		}
		sum := sha256.Sum256(data)
		config.Fonts = append(config.Fonts, cogent.ManualFont{Name: name, Data: data, SHA256: hex.EncodeToString(sum[:])})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	a, _ := app.New(app.Config{})
	defer a.Close()
	workspace, err := New(a, "one", struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Open(app.Window{ID: "one", Title: "Fixture", Width: 1920, Height: 1080, View: func() ui.Node {
		n := workspace.View()
		// Manual deliberately refuses native extensions. Keep their allocation in
		// this control/label layout test; native composition is checked separately.
		var project func(*ui.Node)
		project = func(n *ui.Node) {
			if n.Kind == ui.Extension {
				n.Kind = ui.Column
				n.Custom = nil
			}
			for i := range n.Children {
				project(&n.Children[i])
			}
		}
		project(&n)
		if workspace.tab == "selections" {
			n.Focus = ui.FocusRequest{ID: SelectionListID, Sequence: 1}
		}
		return n
	}}); err != nil {
		t.Fatal(err)
	}
	manual, err := cogent.NewManual(ctx, a, config)
	if err != nil {
		t.Fatal(err)
	}
	defer manual.Close(ctx)
	for range 32 {
		if _, err = manual.Step(ctx, 16); err != nil {
			t.Fatal(err)
		}
		if _, err = manual.Frame(ctx, "one"); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := manual.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defined, visible := 0, 0
	var clip, root image.Rectangle
	for _, e := range snapshot.Windows[0].Elements {
		if e.ID == "density-scroll" {
			clip = e.Bounds
		}
		if e.ID == "root" {
			root = e.Bounds
		}
	}
	for _, e := range snapshot.Windows[0].Elements {
		if len(e.ID) == 10 && strings.HasPrefix(e.ID, "dense-") {
			defined++
			if e.Visible && !e.Bounds.Empty() && e.Bounds.In(clip) {
				visible++
			}
		}
	}
	if defined != DenseLabels || visible != DenseLabels {
		for _, e := range snapshot.Windows[0].Elements {
			if e.ID == "root" || e.ID == "density-scroll" || e.ID == "dense" || e.ID == "dense-row-00" || e.ID == "dense-0000" || e.ID == "controls" {
				t.Log(e.ID, e.Bounds, e.Visible)
			}
		}
		t.Fatalf("completed density defined=%d visible=%d clip=%v", defined, visible, clip)
	}
	for _, item := range []string{"home", "models", "part"} {
		found := false
		for _, e := range snapshot.Windows[0].Elements {
			if e.ID == ui.BreadcrumbItemID("workspace-path", item) {
				found = e.Visible && e.Enabled && !e.Bounds.Empty() && e.Bounds.In(root)
			}
		}
		if !found {
			t.Fatalf("composed breadcrumb item lacks visible geometry: %s", item)
		}
	}
	// A selected scroll page must allocate real field geometry. This catches
	// fixture regressions that catalog validation and hidden-page tests cannot.
	if err = a.Do(func() { workspace.tab = "colors"; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	for range 32 {
		if _, err = manual.Step(ctx, 16); err != nil {
			t.Fatal(err)
		}
		if _, err = manual.Frame(ctx, "one"); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err = manual.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []ui.ColorChannel{ui.ColorRed, ui.ColorGreen, ui.ColorBlue, ui.ColorAlpha, ui.ColorHex} {
		found := false
		for _, e := range snapshot.Windows[0].Elements {
			if e.ID == ui.ColorFieldID("color", c) {
				found = e.Visible && e.Enabled && !e.Bounds.Empty()
				if !found {
					t.Log("color field geometry", e)
				}
			}
		}
		if !found {
			t.Fatalf("selected color field has no visible geometry: %s", c)
		}
	}
	settle := func() {
		t.Helper()
		for range 32 {
			if _, err := manual.Step(ctx, 16); err != nil {
				t.Fatal(err)
			}
			if _, err := manual.Frame(ctx, "one"); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err = a.Do(func() { workspace.tab = "selections"; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	settle()
	snapshot, err = manual.Snapshot(ctx)
	if err != nil || snapshot.Windows[0].FocusedID != SelectionListID {
		t.Fatal("selected list page did not allocate focus", snapshot, err)
	}
	window := snapshot.Windows[0].Incarnation
	for i, code := range []key.Codes{key.CodeEnd, key.CodeHome, key.CodeEnd} {
		if err = manual.Input(ctx, cogent.ManualInput{Window: "one", WindowIncarnation: window, Type: events.KeyChord, Code: code}); err != nil {
			t.Fatal(err)
		}
		settle()
		if workspace.State().ListSelection != "last" || workspace.State().Changes[SelectionListID] != i+1 {
			t.Fatal("actual list accept/reject/repeat route", i, workspace.State())
		}
	}
	for i, id := range []string{SelectionLeftID, SelectionRightID} {
		snapshot, err = manual.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var target [4]float32
		for _, drawing := range snapshot.Windows[0].Canvases {
			if drawing.ID == SelectionCanvasID {
				for _, object := range drawing.Objects {
					if object.ObjectID == id {
						target = object.Bounds
					}
				}
			}
		}
		if target[2] <= 0 || target[3] <= 0 {
			t.Fatal("selected canvas object lacks completed measured geometry", id, snapshot.Windows[0].Canvases)
		}
		point := image.Pt(int(target[0]+target[2]/2), int(target[1]+target[3]/2))
		for _, kind := range []events.Types{events.MouseDown, events.MouseUp} {
			if err = manual.Input(ctx, cogent.ManualInput{Window: "one", WindowIncarnation: window, Type: kind, Button: events.Left, Position: point}); err != nil {
				t.Fatal(err)
			}
		}
		settle()
		if workspace.State().DrawingSelection != id || workspace.State().Changes[SelectionCanvasID] != i+1 {
			t.Fatal("actual measured canvas pointer selection", i, workspace.State())
		}
	}
	if err = a.Do(func() { workspace.tab = "live"; _ = a.Invalidate("one") }); err != nil {
		t.Fatal(err)
	}
	settle()
	snapshot, err = manual.Snapshot(ctx)
	if err != nil || len(snapshot.Windows[0].Canvases) != 0 || workspace.State().ListSelection != "last" || workspace.State().DrawingSelection != SelectionRightID {
		t.Fatal("hidden selection page leaked geometry or lost model state", snapshot, err)
	}

}
