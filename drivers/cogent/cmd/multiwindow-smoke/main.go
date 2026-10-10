// multiwindow-smoke opens two temporary native windows, closes the first, and
// verifies that the surviving window processes another update before shutdown.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func main() {
	a, newErr := app.New(app.Config{})
	if newErr != nil {
		fmt.Fprintln(os.Stderr, newErr)
		os.Exit(1)
	}
	defer a.Close()
	var firstViews, secondViews atomic.Int32
	window := func(id string, views *atomic.Int32) app.Window {
		return app.Window{ID: id, Title: "Vellum smoke " + id, Width: 360, Height: 180, View: func() ui.Node {
			views.Add(1)
			return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Temporary multiwindow lifecycle check"}}}
		}}
	}
	if err := a.Open(window("first", &firstViews)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	registry := command.New()
	registry.Register(command.Command{ID: "smoke", Label: "Smoke test", Execute: func(context.Context) error { return nil }})
	timeout, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		wait := func(test func() bool) error {
			for !test() {
				select {
				case <-timeout.Done():
					return timeout.Err()
				case <-time.After(30 * time.Millisecond):
				}
			}
			return nil
		}
		finish := func(err error) { result <- err; a.Close() }
		if err := wait(func() bool { _, err := a.Services(); return err == nil }); err != nil {
			finish(err)
			return
		}
		service, _ := a.Services()
		if err := service.SetMenu(timeout, []app.MenuGroup{{Label: "Empty menu"}}); err != nil {
			finish(fmt.Errorf("empty native menu: %w", err))
			return
		}
		if err := service.SetMenu(timeout, []app.MenuGroup{{Label: "Smoke", Commands: []string{"smoke"}}}); err != nil {
			finish(fmt.Errorf("native menu: %w", err))
			return
		}
		if _, err := service.ReadClipboard(timeout); err != nil {
			finish(fmt.Errorf("clipboard read: %w", err))
			return
		}
		panelCtx, panelCancel := context.WithTimeout(timeout, 250*time.Millisecond)
		_, panelErr := service.ChooseFile(panelCtx, app.FileOptions{Title: "Vellum cancellation check", Extensions: []string{"txt"}})
		panelCancel()
		if !errors.Is(panelErr, context.DeadlineExceeded) {
			finish(fmt.Errorf("panel cancel: %v", panelErr))
			return
		}
		if err := a.Open(window("second", &secondViews)); err != nil {
			finish(err)
			return
		}
		if err := wait(func() bool { return secondViews.Load() > 0 }); err != nil {
			finish(err)
			return
		}
		// Let the second window finish native creation before closing the first.
		time.Sleep(250 * time.Millisecond)
		first, err := a.WindowContext("first")
		if err != nil {
			finish(err)
			return
		}
		second, err := a.WindowContext("second")
		if err != nil {
			finish(err)
			return
		}
		if err := a.CloseWindow("first"); err != nil {
			finish(err)
			return
		}
		if first.Err() == nil || second.Err() != nil || a.Context().Err() != nil {
			finish(fmt.Errorf("lifetime isolation"))
			return
		}
		previous := secondViews.Load()
		a.Invalidate("second")
		if err := wait(func() bool { return secondViews.Load() > previous }); err != nil {
			finish(err)
			return
		}
		pending := make(chan error, 1)
		go func() {
			_, err := service.ChooseFile(context.Background(), app.FileOptions{Title: "Vellum shutdown cancellation check"})
			pending <- err
		}()
		time.Sleep(150 * time.Millisecond)
		if err := a.CloseWindow("second"); err != nil {
			finish(err)
			return
		}
		select {
		case err := <-pending:
			if !errors.Is(err, context.Canceled) {
				finish(fmt.Errorf("pending service survived shutdown: %v", err))
				return
			}
		case <-timeout.Done():
			finish(timeout.Err())
			return
		}
		result <- nil
	}()
	err := a.Run(timeout, cogent.Driver{Commands: registry, OnError: func(err error) { fmt.Fprintln(os.Stderr, err) }})
	select {
	case check := <-result:
		if err == nil {
			err = check
		}
	case <-timeout.Done():
		if err == nil {
			err = timeout.Err()
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("PASS native two-window isolation; first=%d second=%d; native menu, clipboard read, file-panel cancellation and pending-service shutdown\n", firstViews.Load(), secondViews.Load())
}
