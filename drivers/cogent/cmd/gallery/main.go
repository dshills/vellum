// The gallery exercises the standard Vellum controls without the workbench's
// large dataset. Run from drivers/cogent with go run ./cmd/gallery.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"cogentcore.org/core/core"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type gallery struct {
	enabled                     bool
	details                     bool
	snapping, pinned            bool
	selection, tab, text        string
	number                      float32
	split                       float32
	dialog                      bool
	focus                       ui.FocusRequest
	picture                     *canvas.TiledImage
	imageFit                    ui.ImageFit
	menuCount                   int
	menuLast                    string
	menuDisabled                bool
	menuChecked                 bool
	colorPicker                 *ui.ColorPicker
	colorEdits, colorInvalid    int
	colorReject, colorNormalize bool
}

func newGallery() *gallery {
	g := &gallery{enabled: true, details: true, selection: "basic", tab: "controls", text: "Vellum", number: 25, split: 0.35, picture: galleryImage(), imageFit: ui.ImageContain}
	g.colorPicker = &ui.ColorPicker{ID: "ink", Label: "Ink color", Color: color.NRGBA{R: 32, G: 96, B: 224, A: 255}, Style: ui.Style{Gap: ui.Some(float32(8))}}
	g.colorPicker.Change = func(e ui.ColorEdit) {
		g.colorEdits++
		if g.colorReject {
			return
		}
		g.colorPicker.Color = e.Apply(g.colorPicker.Color)
		if g.colorNormalize {
			g.colorPicker.Color.A = 255
		}
		g.colorPicker.ValidationMessage = ""
	}
	g.colorPicker.Invalid = func(err error) { g.colorInvalid++; g.colorPicker.ValidationMessage = err.Error() }
	return g
}

func galleryImage() *canvas.TiledImage {
	pixels := image.NewNRGBA(image.Rect(0, 0, 96, 32))
	palette := []color.NRGBA{{R: 240, G: 64, B: 48, A: 255}, {R: 48, G: 192, B: 96, A: 255}, {R: 48, G: 96, B: 240, A: 128}}
	for y := range 32 {
		for x := range 96 {
			pixels.SetNRGBA(x, y, palette[x/32])
		}
	}
	source, err := canvas.NewTiledImage(pixels)
	if err != nil {
		panic(err)
	} // The fixed, private source satisfies admission.
	return source
}
func (g *gallery) view() ui.Node {
	numeric := ui.Numeric{Min: 0, Max: 100, Step: 1}
	changeNumber := func(e ui.ChangeEvent) { g.number = e.Number }
	controls := ui.Node{ID: "controls", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		{ID: "enable", Kind: ui.Checkbox, Text: "Enable editing controls", Checked: g.enabled, Change: func(e ui.ChangeEvent) { g.enabled = e.Checked }, Accessibility: ui.Accessibility{Help: "Toggle the disabled state of the controls below"}},
		{ID: "details", Kind: ui.Disclosure, Text: "Editing controls", Checked: g.details, Change: func(e ui.ChangeEvent) { g.details = e.Checked }, Children: []ui.Node{
			{ID: "fields", Kind: ui.FormGroup, Text: "Editing", Disabled: !g.enabled, Style: ui.Style{Gap: ui.Some(float32(10))}, Children: []ui.Node{
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
		}},
		{ID: "more-controls", Kind: ui.Toolbar, Accessibility: ui.Accessibility{Name: "Control tools"}, Children: []ui.Node{
			{ID: "snapping", Kind: ui.Switch, Text: "Snapping", Checked: g.snapping, Change: func(e ui.ChangeEvent) { g.snapping = e.Checked }},
			{ID: "pinned", Kind: ui.ToggleButton, Icon: ui.SymbolCheck, Text: "Pin panel", Checked: g.pinned, Change: func(e ui.ChangeEvent) { g.pinned = e.Checked }},
		}},
		{ID: "controls-separator", Kind: ui.Separator},
		{ID: "progress", Kind: ui.ProgressBar, Number: g.number / 100, Accessibility: ui.Accessibility{Name: "Operation progress"}},
		{ID: "show-dialog", Kind: ui.Button, Text: "Open preferences dialog", Activate: func() { g.dialog = true }},
		{ID: "state", Kind: ui.StatusBar, Accessibility: ui.Accessibility{Name: "Application status"}, Text: fmt.Sprintf("Preset %s · project %s", g.selection, g.text)},
	}}
	splitDemo := ui.Node{ID: "split-demo", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "split-instructions", Kind: ui.Label, Text: "Drag the divider, or focus it and use Left/Right, Home and End."},
		{ID: "panes", Kind: ui.SplitPane, Number: g.split, Grow: true, Change: func(e ui.ChangeEvent) { g.split = e.Number }, Accessibility: ui.Accessibility{Name: "Inspector width", Help: "Resize the inspector and workspace"}, Children: []ui.Node{
			{ID: "split-inspector", Kind: ui.FormGroup, Text: "Inspector", Children: []ui.Node{{ID: "split-name", Kind: ui.TextField, Text: g.text, Change: func(e ui.ChangeEvent) { g.text = e.Text }, Accessibility: ui.Accessibility{Name: "Project name"}}}},
			{ID: "split-workspace", Kind: ui.Scroll, Children: []ui.Node{{ID: "split-content", Kind: ui.Label, Text: "Workspace content retains its native controls as the panes resize."}, {ID: "split-reset", Kind: ui.Button, Text: "Reset divider", Activate: func() { g.split = 0.35 }}}},
		}},
	}}
	symbols := ui.Node{ID: "symbols", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		{ID: "menu-actions", Kind: ui.MenuButton, Text: "Document actions", Icon: ui.SymbolSettings, Disabled: g.menuDisabled,
			Options: []ui.Option{{ID: "blocked", Label: "Unavailable", Disabled: true}, {ID: "first", Label: "Run action"}, {ID: "second", Label: "Run action"}},
			Change:  func(e ui.ChangeEvent) { g.menuCount++; g.menuLast = e.SelectedID }},
		{ID: "menu-status", Kind: ui.StatusBar, Text: fmt.Sprintf("Action %s · requests %d", g.menuLast, g.menuCount)},
		{ID: "view-menu", Kind: ui.MenuButton, Text: "View options", Options: []ui.Option{
			{ID: "grid", Label: "Show grid", Checkable: true, Checked: g.menuChecked},
			{ID: "group-divider", Separator: true},
			{ID: "reset", Label: "Reset view"},
		}, Change: func(e ui.ChangeEvent) {
			if e.SelectedID == "grid" {
				g.menuChecked = e.Checked
			} else if e.SelectedID == "reset" {
				g.menuChecked = false
			}
		}},
		{ID: "view-state", Kind: ui.Label, Text: fmt.Sprintf("Show grid: %t", g.menuChecked)},
	}}
	for _, symbol := range ui.Symbols() {
		symbols.Children = append(symbols.Children, ui.Node{ID: "symbol-" + string(symbol), Kind: ui.Row, Style: ui.Style{Gap: ui.Some(float32(12))}, Children: []ui.Node{
			{ID: "art-" + string(symbol), Kind: ui.Icon, Icon: symbol},
			{ID: "label-" + string(symbol), Kind: ui.Label, Text: string(symbol)},
		}})
	}
	symbols.Children = append([]ui.Node{{ID: "action-appearances", Kind: ui.Toolbar, Accessibility: ui.Accessibility{Name: "Action appearances"}, Children: []ui.Node{
		{ID: "primary-action", Kind: ui.Button, Text: "Primary", Icon: ui.SymbolCheck, Appearance: ui.ButtonPrimary},
		{ID: "secondary-action", Kind: ui.Button, Text: "Secondary", Appearance: ui.ButtonSecondary},
		{ID: "quiet-action", Kind: ui.Button, Text: "Quiet", Appearance: ui.ButtonQuiet},
		{ID: "tool-action", Kind: ui.Button, Icon: ui.SymbolSearch, Appearance: ui.ButtonTool, Accessibility: ui.Accessibility{Name: "Search", Help: "Compact icon-only tool"}, Activate: func() { g.focus.ID = "project-name"; g.focus.Sequence++; g.tab = "controls" }},
	}}}, symbols.Children...)
	images := ui.Node{ID: "images", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		{ID: "image-instructions", Kind: ui.Label, Text: "Immutable shared pixels. Contain shows the full image; cover crops centrally; stretch fills both axes."},
		{ID: "image-fit", Kind: ui.SegmentedControl, SelectedID: string(g.imageFit), Accessibility: ui.Accessibility{Name: "Image fit"}, Options: []ui.Option{{ID: string(ui.ImageContain), Label: "Contain"}, {ID: string(ui.ImageCover), Label: "Cover"}, {ID: string(ui.ImageStretch), Label: "Stretch"}}, Change: func(e ui.ChangeEvent) { g.imageFit = ui.ImageFit(e.SelectedID) }},
		{ID: "image-preview", Kind: ui.Image, Image: g.picture, Fit: g.imageFit, Text: "Three color strips, with transparency in the blue strip", Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 320, Height: 160})}},
		{ID: "image-replace", Kind: ui.Button, Text: "Replace first strip", Activate: func() {
			patch := image.NewNRGBA(image.Rect(0, 0, 32, 32))
			for y := range 32 {
				for x := range 32 {
					patch.SetNRGBA(x, y, color.NRGBA{R: 240, G: 192, A: 255})
				}
			}
			source, err := g.picture.WithPatch(patch.Bounds(), patch)
			if err != nil {
				panic(err)
			}
			g.picture = source
		}},
	}}
	colorsPage := ui.Node{ID: "colors", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		g.colorPicker.View(),
		{ID: "color-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "color-reset", Kind: ui.Button, Text: "Reset color", Activate: func() {
				g.colorPicker.Color = color.NRGBA{R: 32, G: 96, B: 224, A: 255}
				g.colorPicker.ValidationMessage = ""
			}},
			{ID: "color-disabled", Kind: ui.Checkbox, Text: "Disable color editing", Checked: g.colorPicker.Disabled, Change: func(e ui.ChangeEvent) { g.colorPicker.Disabled = e.Checked }},
			{ID: "color-focus", Kind: ui.Button, Text: "Focus hex", Activate: func() { g.focus.ID = ui.ColorFieldID("ink", ui.ColorHex); g.focus.Sequence++ }},
		}},
	}}

	return ui.Node{ID: "gallery", Kind: ui.Column, Grow: true, Focus: g.focus, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(12))}, Children: []ui.Node{
		{ID: "title", Kind: ui.Label, Text: "Vellum controls gallery", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 24, Weight: 600})}},
		{ID: "pages", Kind: ui.Tabs, Grow: true, SelectedID: g.tab, Options: []ui.Option{{ID: "controls", Label: "Controls"}, {ID: "split-demo", Label: "Split panes"}, {ID: "symbols", Label: "Icons and actions"}, {ID: "images", Label: "Images"}, {ID: "colors", Label: "Colors"}, {ID: "about", Label: "About"}}, Change: func(e ui.ChangeEvent) { g.tab = e.SelectedID }, Children: []ui.Node{controls, splitDemo, symbols, images, colorsPage, {ID: "about", Kind: ui.Column, Children: []ui.Node{{ID: "about-text", Kind: ui.Label, Text: "Ordinary Go nodes, retained native controls, local theme values."}}}}},
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
	disclosureSmoke := flag.Bool("disclosure-smoke", false, "check actual Cocoa disclosure publication and actions in a bounded fixture")
	menuSmoke := flag.Bool("menu-smoke", false, "check actual Cocoa menu routing and lifetime in a bounded fixture")
	colorSmoke := flag.Bool("color-smoke", false, "verify bounded native color field acceptance and lifetime")
	flag.Parse()
	if *colorSmoke && (*smoke || *menuSmoke || *disclosureSmoke || *exit > 0) {
		must(errors.New("color-smoke requires its own bounded process"))
	}
	if *menuSmoke && (*smoke || *disclosureSmoke || *exit > 0) {
		must(errors.New("menu-smoke requires its own bounded process"))
	}
	if *disclosureSmoke && (*smoke || *exit > 0) {
		must(errors.New("disclosure-smoke requires its own bounded process"))
	}
	if *disclosureSmoke || *menuSmoke || *colorSmoke {
		core.DebugSettings.DisableWindowGeometrySaver = true
	}
	a, err := app.New(app.Config{})
	must(err)
	defer a.Close()
	g := newGallery()
	if *menuSmoke {
		g.tab = "symbols"
	}
	if *colorSmoke {
		g.tab = "colors"
	}
	must(a.Open(app.Window{ID: "gallery", Title: "Vellum Controls", Width: 720, Height: 650, View: g.view}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var colorDone chan error
	var removeColorCleanup func()
	if *colorSmoke {
		var observeErr error
		removeColorCleanup, observeErr = beginColorCleanup()
		must(observeErr)
		defer removeColorCleanup()
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, 60*time.Second)
		defer stop()
		colorDone = make(chan error, 1)
		go func() { colorDone <- checkColor(ctx, a, g); cancel() }()
	}
	var menuDone chan error
	if *menuSmoke {
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, 60*time.Second)
		defer stop()
		menuDone = make(chan error, 1)
		go func() { menuDone <- checkMenu(ctx, a, g); cancel() }()
	}
	var disclosureDone chan error
	if *disclosureSmoke {
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, 45*time.Second)
		defer stop()
		disclosureDone = make(chan error, 1)
		go func() { disclosureDone <- checkDisclosure(ctx); cancel() }()
	}
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
	if colorDone != nil {
		cancel()
		var checkErr error
		select {
		case checkErr = <-colorDone:
		case <-time.After(6 * time.Second):
			checkErr = errors.New("color worker did not join after native exit")
		}
		checkErr = errors.Join(checkErr, finishColorCleanup())
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		must(errors.Join(err, checkErr))
		return
	}
	if menuDone != nil {
		cancel()
		var checkErr error
		select {
		case checkErr = <-menuDone:
		case <-time.After(6 * time.Second):
			checkErr = errors.New("menu worker did not join after native exit")
		}
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		must(errors.Join(err, checkErr))
		return
	}
	if disclosureDone != nil {
		cancel()
		select {
		case checkErr := <-disclosureDone:
			must(checkErr)
		case <-time.After(6 * time.Second):
			must(errors.New("disclosure worker did not join after native exit"))
		}
	}
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
