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
	"path/filepath"
	"time"

	"cogentcore.org/core/core"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/drivers/cogent/internal/nativefixture"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
)

type gallery struct {
	inputHelpers                                                                               *inputHelpersGallery
	notes                                                                                      *editor.Model
	textRequests, editRequests, dialogOpens, dialogDismisses, dialogApplies                    int
	textReject, textDisabled, notesDisabled                                                    bool
	mediaPage                                                                                  string
	mediaTools, mediaLast, imageReplacements, imageFitRequests                                 int
	detailsRequests, splitRequests, structurePageRequests, documentSelectRequests              int
	detailsReject, detailsReadOnly, detailsDisabled, splitReject, splitReadOnly, splitDisabled bool
	structurePage                                                                              string
	selectionRequests, numberRequests, outsideClicks                                           int
	enabledRequests, snappingRequests, pinnedRequests                                          int
	selectionReject, numberReject                                                              bool
	documents                                                                                  []string
	documentSelected                                                                           string
	documentText                                                                               map[string]string
	documentReject                                                                             bool
	documentCloseRequests                                                                      int
	enabled                                                                                    bool
	details                                                                                    bool
	snapping, pinned                                                                           bool
	selection, tab, text                                                                       string
	number                                                                                     float32
	split                                                                                      float32
	dialog                                                                                     bool
	focus                                                                                      ui.FocusRequest
	picture                                                                                    *canvas.TiledImage
	imageFit                                                                                   ui.ImageFit
	menuCount                                                                                  int
	viewMenuRequests                                                                           int
	viewMenuReject                                                                             bool
	menuLast                                                                                   string
	menuDisabled                                                                               bool
	menuChecked                                                                                bool
	colorPicker                                                                                *ui.ColorPicker
	breadcrumb                                                                                 *ui.Breadcrumb
	breadcrumbRequests                                                                         int
	breadcrumbReject                                                                           bool
	colorEdits, colorInvalid                                                                   int
	colorReject, colorNormalize                                                                bool
	listSelection                                                                              string
	listReject, listReadOnly, listDisabled, listReversed                                       bool
	listRequests                                                                               int
}

func newGallery() *gallery {
	g := &gallery{enabled: true, details: true, selection: "basic", tab: "controls", text: "Vellum", number: 25, split: 0.35, picture: galleryImage(), imageFit: ui.ImageContain}
	var err error
	g.notes, err = editor.New("Draft\nsecond line\n", editor.Config{})
	if err != nil {
		panic(err)
	}
	g.structurePage = "sections"
	g.mediaPage = "chrome"
	g.documents = []string{"sketch", "notes"}
	g.documentSelected = "sketch"
	g.documentText = map[string]string{"sketch": "Untitled sketch", "notes": "Working notes"}
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
	g.breadcrumb = &ui.Breadcrumb{ID: "project-path", Label: "Project location", Items: galleryBreadcrumbItems()}
	g.breadcrumb.Navigate = func(e ui.BreadcrumbEvent) {
		g.breadcrumbRequests++
		if g.breadcrumbReject {
			return
		}
		for i, item := range g.breadcrumb.Items {
			if item.ID == e.ItemID {
				g.breadcrumb.Items = append([]ui.BreadcrumbItem(nil), g.breadcrumb.Items[:i+1]...)
				return
			}
		}
	}

	g.inputHelpers = newInputHelpersGallery()
	g.listSelection = "layer-00"
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
func galleryBreadcrumbItems() []ui.BreadcrumbItem {
	return []ui.BreadcrumbItem{{ID: "projects", Label: "Projects"}, {ID: "workspace", Label: "Workspace"}, {ID: "document", Label: "Document"}}
}
func (g *gallery) view() ui.Node {
	path, err := g.breadcrumb.View()
	if err != nil {
		panic(err)
	}

	controls := ui.Node{ID: "controls", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		g.enableControl(),
		g.detailsControl(
			ui.Node{ID: "fields", Kind: ui.FormGroup, Text: "Editing", Disabled: !g.enabled, Style: ui.Style{Gap: ui.Some(float32(10))}, Children: []ui.Node{
				{ID: "selection-label", Kind: ui.Label, Text: "Preset"},
				g.selectionControl(),
				g.radioControl(),
				g.segmentedControl(),
				{ID: "value-label", Kind: ui.Label, Text: fmt.Sprintf("Shared slider / numeric value: %.0f", g.number)},
				g.sliderControl(),
				g.numberControl(),
				{ID: "text-label", Kind: ui.Label, Text: "Project name"},
				g.projectTextField("project-name", "Project name"),
				{ID: "focus-name", Kind: ui.Button, Text: "Focus project name", Activate: func() { g.focus.ID = "project-name"; g.focus.Sequence++ }},
			}},
		),
		{ID: "more-controls", Kind: ui.Toolbar, Accessibility: ui.Accessibility{Name: "Control tools"}, Children: []ui.Node{
			g.snappingControl(),
			g.pinnedControl(),
		}},

		path,
		{ID: "path-tools", Kind: ui.Toolbar, Children: []ui.Node{
			{ID: "path-reset", Kind: ui.Button, Text: "Reset location", Activate: func() { g.breadcrumb.Items = galleryBreadcrumbItems() }},
			{ID: "path-reject", Kind: ui.Checkbox, Text: "Reject navigation", Checked: g.breadcrumbReject, Change: func(e ui.ChangeEvent) { g.breadcrumbReject = e.Checked }},
			{ID: "path-disabled", Kind: ui.Checkbox, Text: "Disable navigation", Checked: g.breadcrumb.Disabled, Change: func(e ui.ChangeEvent) { g.breadcrumb.Disabled = e.Checked }},
		}},
		{ID: "path-status", Kind: ui.Label, Text: fmt.Sprintf("Navigation requests: %d", g.breadcrumbRequests)},
		{ID: "controls-separator", Kind: ui.Separator},
		g.progressBarControl(),
		g.progressShapesControl(),
		{ID: "show-dialog", Kind: ui.Button, Text: "Open preferences dialog", Activate: g.openPreferences},
		{ID: "state", Kind: ui.StatusBar, Accessibility: ui.Accessibility{Name: "Application status"}, Text: fmt.Sprintf("Preset %s · project %s", g.selection, g.text)},
	}}
	splitDemo := ui.Node{ID: "split-demo", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "split-instructions", Kind: ui.Label, Text: "Drag the divider, or focus it and use Left/Right, Home and End."},
		g.splitControl(),
	}}
	symbols := ui.Node{ID: "symbols", Kind: ui.Scroll, Grow: true, Children: []ui.Node{
		g.actionMenuControl(),
		{ID: "menu-status", Kind: ui.StatusBar, Text: fmt.Sprintf("Action %s · requests %d", g.menuLast, g.menuCount)},
		g.viewMenuControl(),
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
	images := g.imagesView()
	colorsPage := g.colorsView()

	return ui.Node{ID: "gallery", Kind: ui.Column, Grow: true, Focus: g.focus, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16)), Gap: ui.Some(float32(12))}, Children: []ui.Node{
		{ID: "title", Kind: ui.Label, Text: "Vellum controls gallery", Style: ui.Style{Typography: ui.Some(ui.Typography{Size: 24, Weight: 600})}},
		{ID: "pages", Kind: ui.Tabs, Grow: true, SelectedID: g.tab, Options: []ui.Option{{ID: "controls", Label: "Controls"}, {ID: "split-demo", Label: "Split panes"}, {ID: "symbols", Label: "Icons and actions"}, {ID: "images", Label: "Images"}, {ID: "colors", Label: "Colors"}, {ID: "documents", Label: "Document tabs"}, {ID: "lists", Label: "Lists"}, {ID: "chrome", Label: "Chrome"}, {ID: "text-page", Label: "Text and dialogs"}, {ID: "inputs", Label: "Input helpers"}, {ID: "about", Label: "About"}}, Change: func(e ui.ChangeEvent) { g.tab = e.SelectedID }, Children: []ui.Node{controls, splitDemo, symbols, images, colorsPage, g.documentsView(), g.listBoxView(), g.chromeView(), g.textPage(), g.inputHelpers.view(), {ID: "about", Kind: ui.Column, Children: []ui.Node{{ID: "about-text", Kind: ui.Label, Text: "Ordinary Go nodes, retained native controls, local theme values."}}}}},
		g.preferencesDialog(),
	}}
}
func main() {
	exit := flag.Duration("exit-after", 0, "close after a duration for smoke checks")
	smoke := flag.Bool("smoke", false, "open and close a native modal during the smoke check")
	disclosureSmoke := flag.Bool("disclosure-smoke", false, "check actual Cocoa disclosure publication and actions in a bounded fixture")
	menuSmoke := flag.Bool("menu-smoke", false, "check actual Cocoa menu routing and lifetime in a bounded fixture")
	colorSmoke := flag.Bool("color-smoke", false, "verify bounded native color field acceptance and lifetime")
	progressSmoke := flag.Bool("progress-smoke", false, "verify native progress shape values, geometry and lifetime")
	catalogue := flag.Bool("control-native-check", false, "run bounded ordinary selector/number-field native FIFO acceptance")
	choices := flag.Bool("choice-native-check", false, "run bounded ordinary choices/toggles/slider native FIFO acceptance")
	navigation := flag.Bool("navigation-native-check", false, "run bounded list and action-menu native FIFO acceptance")
	structure := flag.Bool("structure-native-check", false, "run bounded tabs/disclosure/split/scroll native FIFO acceptance")
	media := flag.Bool("media-native-check", false, "run bounded image/chrome/composite native FIFO acceptance")
	textModal := flag.Bool("text-native-check", false, "run bounded text/editor/dialog native FIFO acceptance")
	widgetDefault, darkDefault := macWidgetExecutableDefaults(os.Args[0])
	macWidgets := flag.Bool("mac-widgets", widgetDefault, "open the focused Mac input widget gallery")
	macDark := flag.Bool("mac-dark", darkDefault, "use the dark Mac appearance with mac-widgets")
	nativeTitlebar := flag.Bool("native-titlebar", widgetDefault, "attach the single-window native AppKit toolbar prototype (mac-widgets only)")
	flag.Parse()
	if *nativeTitlebar && !*macWidgets {
		must(errors.New("native-titlebar requires mac-widgets"))
	}
	if *macDark && !*macWidgets {
		must(errors.New("mac-dark requires mac-widgets"))
	}
	if *macWidgets && (*textModal || *media || *structure || *navigation || *choices || *catalogue || *smoke || *disclosureSmoke || *menuSmoke || *colorSmoke || *progressSmoke) {
		must(errors.New("mac-widgets runs separately from native acceptance fixtures"))
	}
	if *textModal && (*media || *structure || *navigation || *choices || *catalogue || *smoke || *disclosureSmoke || *menuSmoke || *colorSmoke || *progressSmoke || *exit > 0) {
		must(errors.New("text-native-check requires its own bounded process"))
	}
	if *media && (*structure || *navigation || *choices || *catalogue || *smoke || *disclosureSmoke || *menuSmoke || *colorSmoke || *progressSmoke || *exit > 0) {
		must(errors.New("media-native-check requires its own bounded process"))
	}
	if *structure && (*navigation || *choices || *catalogue || *smoke || *disclosureSmoke || *menuSmoke || *colorSmoke || *progressSmoke || *exit > 0) {
		must(errors.New("structure-native-check requires its own bounded process"))
	}
	if *navigation && (*choices || *catalogue || *smoke || *disclosureSmoke || *menuSmoke || *colorSmoke || *progressSmoke || *exit > 0) {
		must(errors.New("navigation-native-check requires its own bounded process"))
	}
	if *choices && (*catalogue || *smoke || *disclosureSmoke || *menuSmoke || *colorSmoke || *progressSmoke || *exit > 0) {
		must(errors.New("choice-native-check requires its own bounded process"))
	}
	if *catalogue && (*smoke || *disclosureSmoke || *menuSmoke || *colorSmoke || *progressSmoke || *exit > 0) {
		must(errors.New("control-native-check requires its own bounded process"))
	}
	if *progressSmoke && (*colorSmoke || *menuSmoke || *disclosureSmoke || *smoke || *exit > 0) {
		must(errors.New("progress-smoke requires its own bounded process"))
	}
	if *colorSmoke && (*smoke || *menuSmoke || *disclosureSmoke || *exit > 0) {
		must(errors.New("color-smoke requires its own bounded process"))
	}
	if *menuSmoke && (*smoke || *disclosureSmoke || *exit > 0) {
		must(errors.New("menu-smoke requires its own bounded process"))
	}
	if *disclosureSmoke && (*smoke || *exit > 0) {
		must(errors.New("disclosure-smoke requires its own bounded process"))
	}
	if *textModal || *media || *disclosureSmoke || *menuSmoke || *colorSmoke || *progressSmoke || *catalogue || *choices || *navigation || *structure {
		restoreGeometry, err := nativefixture.IsolateGeometry()
		must(err)
		defer restoreGeometry()
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
	view := g.view
	var macGallery *macWidgetsGallery
	if *macWidgets {
		g.inputHelpers.setAppearance(*macDark)
		macGallery = newMacWidgetsGallery(g.inputHelpers, *macDark)
		macGallery.toolbars.enablePreferences()
		view = macGallery.view
	}
	if *catalogue {
		view = g.controlCatalogueView
	}
	if *choices {
		view = g.choiceCatalogueView
	}
	if *navigation {
		view = g.navigationCatalogueView
	}
	if *structure {
		view = g.structureCatalogueView
	}
	if *media {
		view = g.mediaCatalogueView
	}
	if *textModal {
		view = g.textCatalogueView
	}
	if *progressSmoke {
		view = g.progressView
	}
	width, height := float32(720), float32(650)
	if *textModal || *media || *choices || *navigation || *structure {
		width, height = 900, 700
	}
	window := app.Window{ID: "gallery", Title: "Vellum Controls", Width: width, Height: height, View: view}
	if *macWidgets {
		window.Title = "Vellum Mac Widgets"
		window.Width = 660
		window.Height = 740
		// Keep native title-bar appearance and dragging under the OS default policy.
		window.Chrome = app.WindowChrome{}
	}
	if *nativeTitlebar {
		window.Chrome.Appearance = app.AppearanceLight
		if *macDark {
			window.Chrome.Appearance = app.AppearanceDark
		}
		macGallery.page = "mac-components"
	}
	must(a.Open(window))
	var titlebar *cogent.TitlebarToolbarPrototype
	if *nativeTitlebar {
		ref, refErr := a.WindowRef(window.ID)
		must(refErr)
		titlebar, err = macGallery.titlebarPrototype(ref)
		must(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var progressDone chan error
	if *progressSmoke {
		removeCleanup, observeErr := beginProgressCleanup()
		must(observeErr)
		defer removeCleanup()
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, 45*time.Second)
		defer stop()
		progressDone = make(chan error, 1)
		go func() { progressDone <- checkProgress(ctx, a, g); cancel() }()
	}
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
	if *macWidgets {
		theme = ui.MacTheme(*macDark)
	}
	if *textModal {
		must(runTextCatalogue(ctx, a, g, theme))
		return
	}
	if *media {
		must(runMediaCatalogue(ctx, a, g, theme))
		return
	}
	if *structure {
		must(runStructureCatalogue(ctx, a, g, theme))
		return
	}
	if *navigation {
		must(runNavigationCatalogue(ctx, a, g, theme))
		return
	}
	if *choices {
		must(runChoiceCatalogue(ctx, a, g, theme))
		return
	}
	if *catalogue {
		must(runControlCatalogue(ctx, a, g, theme))
		return
	}
	err = a.Run(ctx, cogent.Driver{Theme: theme, TitlebarToolbarPrototype: titlebar})
	if progressDone != nil {
		cancel()
		var checkErr error
		select {
		case checkErr = <-progressDone:
		case <-time.After(6 * time.Second):
			checkErr = errors.New("progress worker did not join after native exit")
		}
		checkErr = errors.Join(checkErr, finishProgressCleanup())
		must(errors.Join(progressRunError(err), checkErr))
		return
	}
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

// A development .app can use the real Go executable directly, without a shell
// launcher. Ordinary gallery binaries still start in the full gallery mode.
func macWidgetExecutableDefaults(executable string) (bool, bool) {
	switch filepath.Base(executable) {
	case "Vellum Mac Widgets":
		return true, false
	case "Vellum Mac Widgets Dark":
		return true, true
	default:
		return false, false
	}
}
