// The gallery exercises the standard Vellum controls without the workbench's
// large dataset. Run from drivers/cogent with go run ./cmd/gallery.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"os"
	"time"

	"cogentcore.org/core/core"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type gallery struct {
	enabled              bool
	snapping, pinned     bool
	selection, tab, text string
	number               float32
	dialog               bool
	focus                ui.FocusRequest
}

func newGallery() *gallery {
	return &gallery{enabled: true, selection: "basic", tab: "controls", text: "Vellum", number: 25}
}
func (g *gallery) view() ui.Node {
	numeric := ui.Numeric{Min: 0, Max: 100, Step: 1}
	changeNumber := func(e ui.ChangeEvent) { g.number = e.Number }
	controls := ui.Node{ID: "controls", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		{ID: "enable", Kind: ui.Checkbox, Text: "Enable editing controls", Checked: g.enabled, Change: func(e ui.ChangeEvent) { g.enabled = e.Checked }, Accessibility: ui.Accessibility{Help: "Toggle the disabled state of the controls below"}},
		{ID: "fields", Kind: ui.Column, Disabled: !g.enabled, Style: ui.Style{Gap: ui.Some(float32(10))}, Children: []ui.Node{
			{ID: "selection-label", Kind: ui.Label, Text: "Preset"},
			{ID: "selection", Kind: ui.Selector, SelectedID: g.selection, Options: []ui.Option{{ID: "basic", Label: "Basic"}, {ID: "advanced", Label: "Advanced"}, {ID: "unavailable", Label: "Unavailable", Disabled: true}}, Change: func(e ui.ChangeEvent) { g.selection = e.SelectedID }, Accessibility: ui.Accessibility{Name: "Preset"}},
			{ID: "radio-preset", Kind: ui.RadioGroup, SelectedID: g.selection, Options: []ui.Option{{ID: "basic", Label: "Basic"}, {ID: "advanced", Label: "Advanced"}, {ID: "unavailable", Label: "Unavailable", Disabled: true}}, Change: func(e ui.ChangeEvent) { g.selection = e.SelectedID }, Accessibility: ui.Accessibility{Name: "Preset radio choices"}},
			{ID: "segmented-preset", Kind: ui.SegmentedControl, SelectedID: g.selection, Options: []ui.Option{{ID: "basic", Label: "Basic"}, {ID: "advanced", Label: "Advanced"}, {ID: "unavailable", Label: "Unavailable", Disabled: true}}, Change: func(e ui.ChangeEvent) { g.selection = e.SelectedID }, Accessibility: ui.Accessibility{Name: "Preset segments"}},
			{ID: "value-label", Kind: ui.Label, Text: fmt.Sprintf("Shared slider / numeric value: %.0f", g.number)},
			{ID: "slider", Kind: ui.Slider, Number: g.number, Numeric: numeric, Change: changeNumber, MinWidth: 300, Accessibility: ui.Accessibility{Name: "Level", Help: "Drag or use arrow keys"}},
			{ID: "number", Kind: ui.NumberField, Number: g.number, Numeric: numeric, Change: changeNumber, Accessibility: ui.Accessibility{Name: "Numeric level"}},
			{ID: "text-label", Kind: ui.Label, Text: "Project name"},
			{ID: "project-name", Kind: ui.TextField, Text: g.text, Change: func(e ui.ChangeEvent) { g.text = e.Text }, Accessibility: ui.Accessibility{Name: "Project name", Help: "Single-line native text field"}},
			{ID: "focus-name", Kind: ui.Button, Text: "Focus project name", Activate: func() { g.focus.ID = "project-name"; g.focus.Sequence++ }},
		}},
		{ID: "more-controls", Kind: ui.Row, Children: []ui.Node{
			{ID: "snapping", Kind: ui.Switch, Text: "Snapping", Checked: g.snapping, Change: func(e ui.ChangeEvent) { g.snapping = e.Checked }},
			{ID: "pinned", Kind: ui.ToggleButton, Text: "Pin panel", Checked: g.pinned, Change: func(e ui.ChangeEvent) { g.pinned = e.Checked }},
		}},
		{ID: "controls-separator", Kind: ui.Separator},
		{ID: "progress", Kind: ui.ProgressBar, Number: g.number / 100, Accessibility: ui.Accessibility{Name: "Operation progress"}},
		{ID: "show-dialog", Kind: ui.Button, Text: "Open preferences dialog", Activate: func() { g.dialog = true }},
		{ID: "state", Kind: ui.Label, Text: fmt.Sprintf("Preset %s · project %s", g.selection, g.text)},
	}}
	return ui.Node{ID: "gallery", Kind: ui.Column, Grow: true, Focus: g.focus, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(12))}, Children: []ui.Node{
		{ID: "title", Kind: ui.Label, Text: "Vellum controls gallery", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 24, Weight: 600})}},
		{ID: "pages", Kind: ui.Tabs, Grow: true, SelectedID: g.tab, Options: []ui.Option{{ID: "controls", Label: "Controls"}, {ID: "about", Label: "About"}}, Change: func(e ui.ChangeEvent) { g.tab = e.SelectedID }, Children: []ui.Node{controls, {ID: "about", Kind: ui.Column, Children: []ui.Node{{ID: "about-text", Kind: ui.Label, Text: "Ordinary Go nodes, retained native controls, local theme values."}}}}},
		{ID: "preferences", Kind: ui.Dialog, Text: "Preferences", Open: g.dialog, Dismiss: func() { g.dialog = false }, Children: []ui.Node{
			{ID: "dialog-description", Kind: ui.Label, Text: "Native modal dialog with retained content. Close with Escape or the Close button."},
			{ID: "dialog-name", Kind: ui.TextField, Text: g.text, Change: func(e ui.ChangeEvent) { g.text = e.Text }, Accessibility: ui.Accessibility{Name: "Project name in preferences"}},
			{ID: "apply", Kind: ui.Button, Text: "Apply and close", Activate: func() { g.dialog = false }},
		}},
	}}
}
func main() {
	exit := flag.Duration("exit-after", 0, "close after a duration for smoke checks")
	smoke := flag.Bool("smoke", false, "open and close a native modal during the smoke check")
	flag.Parse()
	a, err := app.New(app.Config{})
	must(err)
	defer a.Close()
	g := newGallery()
	must(a.Open(app.Window{ID: "gallery", Title: "Vellum Controls", Width: 720, Height: 650, View: g.view}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if *exit > 0 {
		time.AfterFunc(*exit, cancel)
	}
	if *smoke {
		core.DebugSettings.WindowRenderTrace = true
		time.AfterFunc(1200*time.Millisecond, func() { _ = a.Post(func() { g.dialog = true; a.Invalidate("gallery") }) })
		time.AfterFunc(2200*time.Millisecond, func() { _ = a.Post(func() { g.dialog = false; a.Invalidate("gallery") }) })
	}
	theme := ui.Theme{Default: ui.Style{Typography: ui.Some(ui.Typography{Size: 16}), Foreground: ui.Some(color.NRGBA{R: 32, G: 36, B: 42, A: 255})}, Kinds: map[ui.Kind]ui.Style{ui.Column: {Gap: ui.Some(float32(8))}}}
	err = a.Run(ctx, cogent.Driver{Theme: theme})
	if err != nil && !errors.Is(err, context.Canceled) {
		must(err)
	}
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
