// These external-consumer scenarios exercise public application contracts.
package sdkcontract_test

import (
	"context"
	"errors"
	"image"
	"image/color"
	"path/filepath"
	"testing"
	"time"

	"github.com/dshills/vellum"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
	"strconv"
)

func TestControlledViewAndCommandContract(t *testing.T) {
	if vellum.Version != "v0.1.0-alpha.3" {
		t.Fatal(vellum.Version)
	}
	checked := false
	view := func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Style: ui.Style{Gap: ui.Some(float32(0)), Padding: ui.Some(ui.UniformInsets(8))}, Children: []ui.Node{{ID: "enabled", Kind: ui.Checkbox, Checked: checked, Accessibility: ui.Accessibility{Name: "Enable preview"}, Change: func(e ui.ChangeEvent) { checked = e.Checked }}}}
	}
	theme := ui.Theme{Default: ui.Style{Gap: ui.Some(float32(20))}}
	if err := ui.Validate(view()); err != nil {
		t.Fatal(err)
	}
	style, err := ui.ResolveStyle(view(), theme)
	if err != nil || style.Gap.Value != 0 {
		t.Fatalf("explicit zero contract: %+v %v", style, err)
	}
	view().Children[0].Change(ui.ChangeEvent{ID: "enabled", Kind: ui.Checkbox, Checked: true})
	if !view().Children[0].Checked {
		t.Fatal("accepted model edit not reflected")
	}
	registry := command.New()
	enabled := true
	calls := 0
	registry.Register(command.Command{ID: "save", Label: "Save", Shortcut: "Control+S", Available: func() error {
		if !enabled {
			return errors.New("busy")
		}
		return nil
	}, Execute: func(context.Context) error { calls++; return nil }})
	enabled = false
	if err := registry.Invoke(context.Background(), "save"); err == nil || calls != 0 {
		t.Fatal("disabled command executed")
	}
	enabled = true
	matched, err := registry.Shortcut(context.Background(), "control+s")
	if !matched || err != nil || calls != 1 {
		t.Fatalf("shared shortcut contract %v %v %d", matched, err, calls)
	}
}
func TestWindowScopedTaskCommitContract(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ref, err := a.OpenWindow(app.Window{ID: "main", Title: "Consumer", Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label, Text: "Consumer"} }})
	if err != nil {
		t.Fatal(err)
	}
	lifetime := ref.Context()
	manager, err := tasks.New(a, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	var task *tasks.Task
	var startErr error
	commits := 0
	err = a.Do(func() {
		task, startErr = tasks.Start(manager, lifetime, func(context.Context, func(float64, string)) (int, error) { return 42, nil }, func(v int, err error) {
			if v != 42 || err != nil {
				t.Errorf("result %d %v", v, err)
			}
			commits++
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if startErr != nil {
		t.Fatal(startErr)
	}
	select {
	case <-task.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("worker failed to deliver")
	}
	if commits != 0 {
		t.Fatal("commit ran without owner Drain")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if commits != 1 {
		t.Fatal(commits)
	}
	if err := a.CloseWindow(ref); err != nil {
		t.Fatal(err)
	}
	if lifetime.Err() == nil || a.Context().Err() == nil {
		t.Fatal("final window lifetime not closed")
	}
}
func TestWorkspacesRoundTripFloatingLayoutsWithoutAliasing(t *testing.T) {
	session, err := workbench.NewSession(workbench.SessionLayout{Version: workbench.SessionVersion, Focus: workbench.Focus{Window: "main", Panel: "document", Serial: 1}, Windows: []workbench.WindowLayout{{ID: "main", Title: "Main", Bounds: workbench.Bounds{Width: 800, Height: 600}, Root: &workbench.Group{ID: "main-tabs", Tabs: []string{"document", "inspector"}, Active: "document"}}}}, []string{"document", "inspector"})
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Float("document", workbench.WindowLayout{ID: "floating", Title: "Document", Bounds: workbench.Bounds{Width: 640, Height: 480}, Root: &workbench.Group{ID: "floating-tabs"}}); err != nil {
		t.Fatal(err)
	}
	if err := session.Dock("document", "main", "main-tabs", 0); err != nil {
		t.Fatal(err)
	}
	encoded, err := session.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Decode(encoded); err != nil {
		t.Fatal(err)
	}
	layout := session.Layout()
	if len(layout.Windows) != 1 || layout.Windows[0].Root.Active != "document" {
		t.Fatalf("docking contract: %+v", layout)
	}
	layout.Windows[0].Root.Tabs[0] = "external mutation"
	if session.Layout().Windows[0].Root.Tabs[0] != "document" {
		t.Fatal("returned layout aliases retained state")
	}
}

func TestDocumentHistoryAndSaveRevisionThroughPublicCommands(t *testing.T) {
	journal, err := history.New(2)
	if err != nil {
		t.Fatal(err)
	}
	value := 0
	change := func(next int) history.Change {
		previous := value
		return history.Change{Label: "Property", Apply: func() error { value = next; return nil }, Revert: func() error { value = previous; return nil }}
	}
	if err = journal.Apply(change(42)); err != nil {
		t.Fatal(err)
	}
	captured := journal.Snapshot().Revision
	if err = journal.Apply(change(84)); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(journal.MarkSavedRevision(captured), history.ErrStale) || !journal.Snapshot().Dirty {
		t.Fatal("stale save cleaned document")
	}
	registry := command.New()
	if err = registry.Register(command.Command{ID: "undo", Label: "Undo", Shortcut: "Meta+Z", ShortcutScope: command.AfterFocusShortcuts, Available: func() error {
		if journal.Snapshot().UndoCount == 0 {
			return history.ErrEmpty
		}
		return nil
	}, Execute: func(context.Context) error { return journal.Undo() }}); err != nil {
		t.Fatal(err)
	}
	if matched, err := registry.ShortcutInScope(context.Background(), "Meta+Z", command.WindowShortcuts); matched || err != nil || value != 84 {
		t.Fatal("document undo preempted focused widget")
	}
	if matched, err := registry.ShortcutInScope(context.Background(), "Meta+Z", command.AfterFocusShortcuts); !matched || err != nil || value != 42 {
		t.Fatal("document fallback failed")
	}
	if err = journal.MarkSavedRevision(captured); err != nil || journal.Snapshot().Dirty {
		t.Fatal("matching save rejected")
	}
	if err = registry.Invoke(context.Background(), "undo"); err != nil || value != 0 || !journal.Snapshot().Dirty {
		t.Fatal("menu undo failed")
	}
	if err = registry.Invoke(context.Background(), "undo"); !errors.Is(err, history.ErrEmpty) {
		t.Fatal("empty history allowed")
	}
	if err = journal.Redo(); err != nil || value != 42 || journal.Snapshot().Dirty {
		t.Fatal("saved revision identity not restored")
	}
}

func TestFileDropPublicContract(t *testing.T) {
	paths, err := ui.CopyFileDropPaths([]string{filepath.Join(t.TempDir(), "file Ω")})
	if err != nil {
		t.Fatal(err)
	}
	var captured ui.FileDropEvent
	node := ui.Node{ID: "drop", Kind: ui.Column, FileDrop: func(e ui.FileDropEvent) bool { captured = e; return true }}
	if err := ui.Validate(node); err != nil {
		t.Fatal(err)
	}
	if !node.FileDrop(ui.FileDropEvent{ID: node.ID, WindowID: "consumer", Paths: paths, Position: ui.Point{X: 5, Y: 10}}) || captured.Position.Y != 10 || captured.WindowID != "consumer" {
		t.Fatal("public payload contract")
	}
	if _, err := ui.CopyFileDropPaths([]string{"relative"}); err == nil {
		t.Fatal("invalid path accepted")
	}
}

func TestPublicDocumentSessionSaveAndClosePolicy(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	j, _ := history.New(4)
	value := 42
	requests := 0
	confirmed := false
	ref, err := a.OpenWindow(app.Window{ID: "main", Title: "SDK", Width: 200, Height: 200, CloseRequest: func() bool { requests++; return confirmed }, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "SDK.vellum")
	completed := false
	session, err := document.New(document.Config[int]{Owner: a, Window: "main", Journal: j, Codec: document.Codec[int]{Capture: func() (int, error) { return value, nil }, Encode: func(v int) ([]byte, error) { return []byte(strconv.Itoa(v)), nil }, Decode: func(b []byte) (int, error) { return strconv.Atoi(string(b)) }, Install: func(v int) error { value = v; return nil }}, OnComplete: func(r document.Result) {
		if r.Err != nil {
			t.Error(r.Err)
		}
		completed = true
	}})
	if err != nil {
		t.Fatal(err)
	}
	a.Do(func() {
		if !errors.Is(a.RequestCloseWindow(ref), app.ErrCloseRefused) {
			t.Fatal("public close policy ignored")
		}
		if err := session.Save(context.Background(), document.SaveOptions{Path: path}); err != nil {
			t.Fatal(err)
		}
	})
	deadline := time.Now().Add(3 * time.Second)
	for !completed && time.Now().Before(deadline) {
		a.Drain()
		time.Sleep(time.Millisecond)
	}
	if !completed || session.State().Path != path || session.State().Busy {
		t.Fatal("SDK save incomplete")
	}
	confirmed = true
	a.Do(func() {
		if err := a.RequestCloseWindow(ref); err != nil {
			t.Fatal(err)
		}
	})
	if a.Context().Err() == nil || requests != 2 {
		t.Fatal("SDK confirmed close failed")
	}
}

func TestPublicDrawingWorkloads(t *testing.T) {
	stroke, err := canvas.StrokePolyline([]canvas.Point{{X: 1, Y: 2}, {X: 10, Y: 12}}, 2, color.NRGBA{R: 255, A: 128})
	if err != nil {
		t.Fatal(err)
	}
	if v, ix := stroke.Counts(); v != 4 || ix != 6 {
		t.Fatal("public stroke mesh")
	}
	samples := make([]float32, 1000000)
	samples[12345] = 1
	wave, err := canvas.NewWaveform(samples)
	if err != nil {
		t.Fatal(err)
	}
	buckets, err := wave.Envelope(12000, 13000, 10)
	if err != nil {
		t.Fatal(err)
	}
	peak := false
	for _, b := range buckets {
		peak = peak || b.Max == 1
	}
	if !peak {
		t.Fatal("public peak reduction")
	}
	index, err := canvas.NewRectIndex([]canvas.Rect{{ID: "visible", Width: 10, Height: 10}, {ID: "outside", X: 1000, Width: 10, Height: 10}})
	if err != nil {
		t.Fatal(err)
	}
	mesh, stats, err := index.Mesh(canvas.Viewport{Width: 100, Height: 100, Zoom: 1})
	if err != nil || stats.VisibleObjects != 1 {
		t.Fatal(stats, err)
	}
	if v, ix := mesh.Counts(); v != 4 || ix != 6 {
		t.Fatal("public culled mesh")
	}
}

func TestPublicRetainedImageContract(t *testing.T) {
	source := image.NewNRGBA(image.Rect(-10, 20, 250, 150))
	source.SetNRGBA(-10, 20, color.NRGBA{R: 240, A: 128})
	original, err := canvas.NewTiledImage(source)
	if err != nil {
		t.Fatal(err)
	}
	if original.Size() != image.Pt(260, 130) || original.TileCount() != 6 {
		t.Fatal("normalized image dimensions")
	}
	var publicImage image.Image = original
	if publicImage.At(0, 0) != (color.NRGBA{R: 240, A: 128}) {
		t.Fatal("public image.Image pixels")
	}
	source.SetNRGBA(-10, 20, color.NRGBA{})
	patch := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	patch.SetNRGBA(0, 0, color.NRGBA{G: 200, A: 255})
	next, err := original.WithPatch(image.Rect(129, 1, 130, 2), patch)
	if err != nil {
		t.Fatal(err)
	}
	tiles := next.ChangedTiles(original)
	if len(tiles) != 1 || tiles[0].Bounds() != image.Rect(128, 0, 256, 128) || len(tiles[0].Pixels()) != canvas.ImageTileSize*canvas.ImageTileSize*4 {
		t.Fatal("public changed-tile contract")
	}
	tiles[0].Pixels()[(tiles[0].Bounds().Dx()+1)*4+1] = 0
	patch.Pix[1] = 0
	if next.NRGBAAt(129, 1).G != 200 || original.NRGBAAt(129, 1).G != 0 || original.NRGBAAt(0, 0).R != 240 {
		t.Fatal("public source/version/accessor isolation")
	}
	layer := canvas.ImageLayer{Image: next, Bounds: canvas.Bounds{Min: canvas.Point{X: -10, Y: -10}, Max: canvas.Point{X: 250, Y: 120}}, Filter: canvas.ImageLinear}
	if err := layer.Validate(); err != nil {
		t.Fatal(err)
	}
}
