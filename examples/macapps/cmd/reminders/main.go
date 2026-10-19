// This example consumes Vellum through public Go APIs.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/examples/macapps/reminders"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	m := reminders.New()
	defer m.Close()
	window, err := a.OpenWindow(app.Window{ID: "reminders", Title: "Vellum Reminders", Width: 1000, Height: 740, View: m.View, Chrome: app.WindowChrome{Style: app.ChromeUnified, Appearance: app.AppearanceLight}})
	if err != nil {
		return err
	}
	if err = m.SetInputBoundary(window.Context(), func(ctx context.Context, id string, done func(error)) error {
		return cogent.SettleTextField(ctx, a, "reminders", id, func(err error) { done(err); _ = a.Invalidate("reminders") })
	}); err != nil {
		return err
	}
	return a.Run(ctx, cogent.Driver{Theme: reminders.Theme(), Commands: m.Commands(), OnError: func(err error) { fmt.Fprintln(os.Stderr, err) }})
}
