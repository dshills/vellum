// macshell is a small public-API multiwindow toolbar/command consumer.
package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/examples/macapps/artwork"
	"github.com/dshills/vellum/ui"
	"image/png"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	type document struct {
		ref      app.WindowRef
		commands *command.Registry
		edits    int
		locked   bool
	}
	theme := ui.MacTheme(false)
	documents := map[app.WindowRef]*document{}
	for _, name := range []string{"A", "B"} {
		name := name
		d := &document{commands: command.New()}
		ref, err := a.OpenWindow(app.Window{ID: "shell-" + name, Title: "Vellum Mac Widgets — " + name, Width: 480, Height: 280, Chrome: app.WindowChrome{Appearance: app.AppearanceLight}, View: func() ui.Node {
			return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Style: ui.Style{Padding: ui.Some(ui.UniformInsets(24)), Background: ui.Some(theme.Palette.Value.Window)}, Children: []ui.Node{
				{ID: "title", Kind: ui.Label, Text: "Independent native toolbar " + name, Style: ui.Style{Typography: ui.Some(ui.MacTypography(20, 600))}},
				{ID: "status", Kind: ui.Label, Text: fmt.Sprintf("Document %s edits: %d", name, d.edits)},
				{ID: "lock", Kind: ui.Checkbox, Text: "Disable editing", Checked: d.locked, Change: func(e ui.ChangeEvent) { d.locked = e.Checked; _ = a.Invalidate(d.ref.ID()) }},
				{ID: "close", Kind: ui.Button, Text: "Close this window", Activate: func() { _ = a.CloseWindow(d.ref) }},
			}}
		}})
		if err != nil {
			return err
		}
		d.ref = ref
		documents[ref] = d
		if err = d.commands.Register(command.Command{ID: "edit", Label: "Edit document " + name, Shortcut: "Meta+E", Available: func() error {
			if d.locked {
				return fmt.Errorf("document locked")
			}
			return d.ref.Context().Err()
		}, Execute: func(context.Context) error { d.edits++; return a.Invalidate(d.ref.ID()) }}); err != nil {
			return err
		}
	}
	image, err := artwork.ToolbarIcon(ui.SymbolDocument)
	if err != nil {
		return err
	}
	var pngData bytes.Buffer
	if err = png.Encode(&pngData, image); err != nil {
		return err
	}
	return a.Run(context.Background(), cogent.Driver{Theme: theme, TitlebarToolbars: func(ref app.WindowRef) (*cogent.TitlebarToolbar, error) {
		d := documents[ref]
		if d == nil {
			return nil, nil
		}
		return &cogent.TitlebarToolbar{Window: ref, Commands: d.commands, Items: []cogent.TitlebarToolbarItem{{ID: "edit", Command: "edit", Kind: ui.Button, PNG: pngData.Bytes()}}}, nil
	}})
}
