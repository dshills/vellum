// Package host contains only shared executable wiring, not app-specific widgets.
package host

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"os"
	"os/signal"
)

type Demo struct {
	ID, Title                       string
	Width, Height                   float32
	View                            func() ui.Node
	Theme                           ui.Theme
	Commands                        *command.Registry
	Chrome                          app.WindowChrome
	FixedSize, DisableRootScrolling bool
}

func Run(d Demo) {
	if err := run(d); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(d Demo) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: d.ID, Title: d.Title, Width: d.Width, Height: d.Height, View: d.View, Chrome: d.Chrome, FixedSize: d.FixedSize, DisableRootScrolling: d.DisableRootScrolling}); err != nil {
		return err
	}
	if err = a.Run(ctx, cogent.Driver{Theme: d.Theme, Commands: d.Commands, OnError: func(err error) { fmt.Fprintln(os.Stderr, err) }}); err != nil {
		return err
	}
	return nil
}
