package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"image/color"
	"os"
	"strconv"
	"time"
)

type rows struct{}

func (rows) Identity() string { return "showcase-parts" }
func (rows) Revision() uint64 { return 1 }
func (rows) RowCount() int    { return 1000000 }
func (rows) RowKey(i int) (string, error) {
	if i < 0 || i >= 1000000 {
		return "", virtual.ErrGridRange
	}
	return strconv.Itoa(i), nil
}
func (rows) IndexOf(k string) (int, bool, error) {
	i, e := strconv.Atoi(k)
	return i, e == nil && i >= 0 && i < 1000000, nil
}
func (s rows) Cell(k, c string) (string, error) {
	i, ok, _ := s.IndexOf(k)
	if !ok {
		return "", virtual.ErrGridRange
	}
	switch c {
	case "part":
		return fmt.Sprintf("R%06d", i+1), nil
	case "value":
		return []string{"10k Ω", "4.7k Ω", "100nF", "1µF"}[i%4], nil
	case "package":
		return []string{"0603", "0805", "0402"}[i%3], nil
	case "status":
		return []string{"Ready", "In review", "Approved"}[i%3], nil
	}
	return "", virtual.ErrGridRange
}

type demo struct {
	latest            *latestFeatures
	preset            string
	pinned            bool
	owner             *app.Application
	renderer          *cogent.GPUCanvas
	scene             *canvas.Scene
	document          *editor.Model
	grid              *virtual.Grid
	tab, name, status string
	count             int
	enabled           bool
	level             float32
	busy, modal       bool
	cancelService     context.CancelFunc
	serviceResult     string
}

func label(id, text string) ui.Node { return ui.Node{ID: id, Kind: ui.Label, Text: text} }
func (d *demo) refresh()            { _ = d.owner.Invalidate("showcase") }
func (d *demo) view() ui.Node {
	intro := ui.Node{ID: "overview", Kind: ui.Column, Grow: true, Children: []ui.Node{
		label("intro", "Real native controls and retained GPU content. Drag the canvas to pan; scroll to zoom; click a shape to select."),
		{ID: "controls", Kind: ui.Row, Children: []ui.Node{
			{ID: "name", Kind: ui.TextField, Text: d.name, MinWidth: 220, Accessibility: ui.Accessibility{Name: "Your name"}, Change: func(e ui.ChangeEvent) { d.name = e.Text; d.refresh() }},
			{ID: "increment", Kind: ui.Button, Text: "Increment counter", Disabled: !d.enabled, Activate: func() { d.count++; d.refresh() }},
			{ID: "enabled", Kind: ui.Checkbox, Text: "Enable counter", Checked: d.enabled, Change: func(e ui.ChangeEvent) { d.enabled = e.Checked; d.refresh() }},
		}},
		{ID: "latest-controls", Kind: ui.Row, Children: []ui.Node{
			{ID: "latest-switch", Kind: ui.Switch, Text: "Enable counter", Checked: d.enabled, Change: func(e ui.ChangeEvent) { d.enabled = e.Checked; d.refresh() }},
			{ID: "latest-toggle", Kind: ui.ToggleButton, Text: "Pin panel", Checked: d.pinned, Change: func(e ui.ChangeEvent) { d.pinned = e.Checked; d.refresh() }},
			{ID: "latest-number", Kind: ui.NumberField, Number: d.level, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: func(e ui.ChangeEvent) { d.level = e.Number; d.refresh() }},
		}},
		{ID: "latest-radio", Kind: ui.RadioGroup, Orientation: ui.Horizontal, SelectedID: d.preset, Options: []ui.Option{{ID: "basic", Label: "Basic"}, {ID: "advanced", Label: "Advanced"}, {ID: "future", Label: "Unavailable", Disabled: true}}, Change: func(e ui.ChangeEvent) { d.preset = e.SelectedID; d.refresh() }},
		{ID: "latest-segments", Kind: ui.SegmentedControl, SelectedID: d.preset, Options: []ui.Option{{ID: "basic", Label: "Basic"}, {ID: "advanced", Label: "Advanced"}, {ID: "future", Label: "Unavailable", Disabled: true}}, Change: func(e ui.ChangeEvent) { d.preset = e.SelectedID; d.refresh() }},
		{ID: "latest-separator", Kind: ui.Separator},
		{ID: "latest-progress", Kind: ui.ProgressBar, Number: d.level / 100, Accessibility: ui.Accessibility{Name: "Shared level progress"}},
		label("greeting", fmt.Sprintf("Hello, %s · Counter: %d", d.name, d.count)),
		{ID: "level", Kind: ui.Slider, Number: d.level, Numeric: ui.Numeric{Min: 0, Max: 100, Step: 1}, Change: func(e ui.ChangeEvent) { d.level = e.Number; d.refresh() }},
		label("level-value", fmt.Sprintf("Controlled slider: %.0f%%", d.level)),
		{ID: "fit", Kind: ui.Button, Text: "Fit canvas", Activate: func() { d.scene.Fit(); d.refresh() }},
		{ID: "gpu", Kind: ui.Extension, Grow: true, MinHeight: 150, Custom: d.renderer},
	}}
	editing := ui.Node{ID: "editing", Kind: ui.Column, Grow: true, Children: []ui.Node{
		label("editor-hint", "Edit this real document. Native selection, composition, and keyboard undo/redo use the framework editor."),
		{ID: "document", Kind: ui.MultilineText, Editor: d.document, Grow: true, SoftWrap: true, TabInserts: true, Accessibility: ui.Accessibility{Name: "Demo document"}, Change: func(ui.ChangeEvent) { d.status = "Document edited — changes stay in this demo session"; d.refresh() }},
	}}
	data := ui.Node{ID: "data", Kind: ui.Column, Grow: true, Children: []ui.Node{
		label("grid-hint", "1,000,000 procedural records · bounded visible cells · frozen first column. Scroll and use arrow keys to navigate."),
		{ID: "grid-tools", Kind: ui.Row, Children: []ui.Node{
			{ID: "grid-diagonal", Kind: ui.Button, Text: "Scroll rows + columns", Activate: func() {
				if err := d.grid.Scroll(120, 300); err != nil {
					d.status = err.Error()
				}
				d.refresh()
			}},
			{ID: "grid-origin", Kind: ui.Button, Text: "Back to start", Activate: func() {
				state, err := d.grid.State()
				if err == nil {
					v := state.Viewport
					v.ScrollX = 0
					v.ScrollY = 0
					err = d.grid.SetViewport(v)
				}
				if err != nil {
					d.status = err.Error()
				}
				d.refresh()
			}},
		}},
		{ID: "parts", Kind: ui.DataGrid, Grid: d.grid, Grow: true, Accessibility: ui.Accessibility{Name: "Million row parts table"}, GridChange: func(e ui.GridEvent) { d.status = "Table selection changed"; d.refresh() }},
	}}
	// Native editor/grid bindings mount only while their page is visible.
	if d.tab != "overview" {
		intro.Children = []ui.Node{label("overview-idle", "Controls and GPU canvas")}
	}
	if d.tab != "editing" {
		editing.Children = []ui.Node{label("editing-idle", "Document editor")}
	}
	if d.tab != "data" {
		data.Children = []ui.Node{label("data-idle", "Million-row table")}
	}
	latest := ui.Node{ID: "latest", Kind: ui.Column}
	if d.tab == "latest" && d.latest != nil {
		latest = d.latest.view()
	}
	suite := d.suiteView()
	if d.tab != "suite" {
		suite.Children = []ui.Node{label("suite-idle", "Feature demos")}
	}
	services := d.servicesView()
	if d.tab != "services" {
		services.Children = []ui.Node{label("services-idle", "macOS services")}
	}
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(10))}, Children: []ui.Node{
		{ID: "title", Kind: ui.Label, Text: "Vellum " + demoVersion() + " / Demo Suite", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 26, Weight: 700})}},
		label("subtitle", "Experimental alpha · Go-owned application state · macOS native driver"),
		{ID: "tabs", Kind: ui.Tabs, Grow: true, SelectedID: d.tab, Options: []ui.Option{{ID: "latest", Label: "What’s new"}, {ID: "overview", Label: "Controls + GPU"}, {ID: "editing", Label: "Document editor"}, {ID: "data", Label: "Million-row table"}, {ID: "services", Label: "macOS Services"}, {ID: "suite", Label: "All feature demos"}}, Children: []ui.Node{latest, intro, editing, data, services, suite}, Change: func(e ui.ChangeEvent) { d.tab = e.SelectedID; d.status = "Ready"; d.refresh() }},
		label("status", d.status),
		{ID: "modal", Kind: ui.Dialog, Open: d.modal, Dismiss: func() { d.modal = false; d.refresh() }, Children: []ui.Node{label("modal-message", "Modal focus and accessibility scope — close to restore the background controls."), {ID: "modal-close", Kind: ui.Button, Text: "Close dialog", Activate: func() { d.modal = false; d.refresh() }}}},
	}}
}
func run() error {
	gpuScene := flag.Bool("gpu-scene", false, "use the experimental GPU scene renderer")
	smoke := flag.Bool("smoke", false, "verify UI descriptions and a native GPU frame, then close")
	flag.Parse()
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	scene, err := canvas.New([]canvas.Rect{{ID: "teal", X: 0, Y: 0, Width: 160, Height: 100, Color: color.RGBA{45, 190, 160, 255}}, {ID: "blue", X: 190, Y: 25, Width: 120, Height: 150, Color: color.RGBA{65, 125, 235, 255}}, {ID: "gold", X: 55, Y: 135, Width: 100, Height: 75, Color: color.RGBA{245, 180, 65, 255}}})
	if err != nil {
		return err
	}
	scene.Fit()
	doc, err := editor.New("Welcome to Vellum\n\nThis document uses the framework's multiline editor.\nTry editing, selecting, and keyboard undo/redo.\n\nUnicode: café · Ελληνικά · שלום · 日本語 · 👋🏽\n\nThe other tabs demonstrate controlled widgets, a retained GPU canvas,\nand a bounded viewport over one million procedural table records.\n", editor.Config{})
	if err != nil {
		return err
	}
	grid, err := virtual.NewGrid(rows{}, []virtual.GridColumn{{ID: "part", Title: "Part", Width: 180}, {ID: "value", Title: "Value", Width: 180}, {ID: "package", Title: "Package", Width: 160}, {ID: "status", Title: "Status", Width: 220}}, virtual.GridConfig{RowHeight: 30, HeaderHeight: 34, FrozenColumns: 1}, virtual.GridViewport{Width: 900, Height: 480})
	if err != nil {
		return err
	}
	d := &demo{preset: "basic", owner: a, scene: scene, document: doc, grid: grid, tab: "overview", name: "Go developer", enabled: true, level: 65, status: "Ready — explore What’s new and the rebuilt feature demos"}
	d.latest = newLatestFeatures(d)
	d.renderer = &cogent.GPUCanvas{Owner: a, Window: "showcase", Model: scene}
	for _, tab := range []string{"latest", "overview", "editing", "data", "services", "suite"} {
		d.tab = tab
		if err := ui.Validate(d.view()); err != nil {
			return err
		}
	}
	d.tab = "latest"
	if *smoke {
		d.tab = "overview"
	}
	if err = a.Open(app.Window{ID: "showcase", Title: "Vellum Demo Suite " + demoVersion(), Width: 1000, Height: 900, View: d.view}); err != nil {
		return err
	}
	result := make(chan error, 1)
	if *smoke {
		go func() {
			ctx, cancel := context.WithTimeout(a.Context(), 40*time.Second)
			defer cancel()
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					result <- ctx.Err()
					a.Close()
					return
				case <-ticker.C:
				}
				var check error
				err := a.PostWait(ctx, func() {
					img, e := d.renderer.Snapshot()
					check = e
					if e == nil {
						p := scene.View.Screen(canvas.Point{X: 80, Y: 50})
						if int(p.X) < 0 || int(p.Y) < 0 || int(p.X) >= img.Bounds().Dx() || int(p.Y) >= img.Bounds().Dy() {
							check = fmt.Errorf("canvas test point outside frame")
						} else {
							pixel := img.NRGBAAt(int(p.X), int(p.Y))
							if pixel.G < 140 || pixel.R > 100 {
								check = fmt.Errorf("unexpected canvas pixel: %v", pixel)
							}
						}
					}
				})
				if err != nil {
					result <- err
					a.Close()
					return
				}
				if check == cogent.ErrFrameNotReady {
					continue
				}
				result <- check
				a.Close()
				return
			}
		}()
	}
	registry, err := d.commands()
	if err != nil {
		return err
	}
	err = a.Run(context.Background(), cogent.Driver{Commands: registry, GPUSceneRendering: *gpuScene})
	if err != nil {
		return err
	}
	if *smoke {
		return <-result
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
