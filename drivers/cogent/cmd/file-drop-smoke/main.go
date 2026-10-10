// file-drop-smoke tests the actual Cocoa file-URL receive path and native scene
// routing. -interactive waits for a separate physical external file drag.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

const mainTitle = "Vellum file drop smoke"
const secondTitle = "Vellum second file drop smoke"

type delivery struct {
	ui.FileDropEvent
	Route string
}

func main() {
	interactive := flag.Bool("interactive", false, "wait for a physical drag rather than the Cocoa fixture")
	duration := flag.Duration("timeout", 30*time.Second, "bounded check duration")
	flag.Parse()
	a, err := app.New(app.Config{})
	must(err)
	defer a.Close()
	deliveries := make(chan delivery, 16)
	handler := func(route string, accept bool) func(ui.FileDropEvent) bool {
		return func(e ui.FileDropEvent) bool {
			captured := e
			captured.Paths = append([]string(nil), e.Paths...)
			deliveries <- delivery{captured, route}
			if !accept {
				e.Paths[0] = "declining child mutation"
			}
			return accept
		}
	}
	version := 0
	modal := false
	must(a.Open(app.Window{ID: "drop", Title: mainTitle, Width: 640, Height: 400, View: func() ui.Node {
		var childDrop func(ui.FileDropEvent) bool
		if version < 2 {
			childDrop = handler(fmt.Sprintf("child-%d", version), true)
		}
		return ui.Node{ID: "root", Kind: ui.Column, Grow: true, FileDrop: handler("root", true), Children: []ui.Node{
			{ID: "help", Kind: ui.Label, Text: "Drop files on the large target below"},
			{ID: "child", Kind: ui.Column, Grow: true, MinHeight: 100, FileDrop: childDrop, Accessibility: ui.Accessibility{Name: fmt.Sprintf("File target %d", version)}, Children: []ui.Node{{ID: "child-label", Kind: ui.Label, Text: fmt.Sprintf("File target %d", version)}}},
			{ID: "decline", Kind: ui.Label, Text: "Declines to the root", FileDrop: handler("decline", false)},
			{ID: "disabled", Kind: ui.Label, Text: "Disabled target", Disabled: true, FileDrop: handler("disabled", true)},
			{ID: "modal", Kind: ui.Dialog, Text: "Drop modal", Open: modal, Dismiss: func() { modal = false }, FileDrop: handler("modal", true), Children: []ui.Node{{ID: "modal-child", Kind: ui.Label, Text: "Modal child declines to host", FileDrop: handler("modal-child", false)}}},
		}}
	}}))
	must(a.Open(app.Window{ID: "second", Title: secondTitle, Width: 420, Height: 200, View: func() ui.Node {
		return ui.Node{ID: "second-root", Kind: ui.Column, Grow: true, FileDrop: handler("second", true), Children: []ui.Node{{ID: "second-child", Kind: ui.Label, Text: "Second window target"}}}
	}}))
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()
	outcome := make(chan error, 1)
	go func() {
		var result error
		if *interactive {
			_, result = waitElement(ctx, "drop", "child", "")
			if result == nil {
				fmt.Println("READY physical drag target: " + mainTitle)
				select {
				case d := <-deliveries:
					fmt.Printf("PHYSICAL delivery window=%s target=%s local_dp=%+v paths=%q\n", d.WindowID, d.ID, d.Position, d.Paths)
					if d.Route != "child-0" || (len(flag.Args()) > 0 && !reflect.DeepEqual(d.Paths, flag.Args())) {
						result = fmt.Errorf("unexpected physical target: %s", d.Route)
					}
				case <-ctx.Done():
					result = ctx.Err()
				}
			}
		} else {
			result = check(ctx, a, deliveries, &version, &modal)
		}
		outcome <- result
		a.Close()
	}()
	runErr := a.Run(ctx, cogent.Driver{})
	select {
	case err = <-outcome:
	case <-ctx.Done():
		err = ctx.Err()
	}
	must(errors.Join(runErr, err))
	if *interactive {
		fmt.Println("PASS physical external file drop")
	} else {
		fmt.Println("PASS Cocoa file-URL receive, exact names, child admission/refusal, parent copies, callback replacement/removal, disabled fallback, independent windows, modal host and background blocking")
	}
}
func waitElement(ctx context.Context, window, id, name string) (cogent.AccessibilityElement, error) {
	for {
		s, err := cogent.InspectAccessibility(ctx, window)
		if err == nil && s.Reachable {
			for _, e := range s.Elements {
				if e.DomainID == id && (name == "" || e.Name == name) && e.Frame[2] > 0 && e.Frame[3] > 0 {
					return e, nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return cogent.AccessibilityElement{}, fmt.Errorf("waiting for %s/%s %s: %w", window, id, name, ctx.Err())
		case <-time.After(15 * time.Millisecond):
		}
	}
}
func fixture(ctx context.Context, window, id, title string, paths []string) error {
	target, err := waitElement(ctx, window, id, "")
	if err != nil {
		return err
	}
	return receiveFileURLs(ctx, title, paths, target.Frame[0]+target.Frame[2]/2, target.Frame[1]+target.Frame[3]/2)
}
func expect(ctx context.Context, deliveries <-chan delivery, route, window, target string, paths []string) error {
	select {
	case d := <-deliveries:
		if d.Route != route || d.WindowID != window || d.ID != target || !reflect.DeepEqual(d.Paths, paths) || d.Position.X < 0 || d.Position.Y < 0 {
			return fmt.Errorf("delivery mismatch: route=%s payload=%+v expected=%s/%s/%s %q", d.Route, d.FileDropEvent, route, window, target, paths)
		}
		fmt.Printf("delivery route=%s window=%s target=%s local_dp=%+v paths=%q\n", route, d.WindowID, d.ID, d.Position, d.Paths)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func check(ctx context.Context, a *app.Application, deliveries <-chan delivery, version *int, modal *bool) error {
	dir, err := os.MkdirTemp("", "vellum-drop-smoke-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	paths := []string{filepath.Join(dir, "space Ω.txt"), filepath.Join(dir, "line\nfile.txt")}
	for _, path := range paths {
		if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
			return err
		}
	}
	if err := fixture(ctx, "drop", "child", mainTitle, paths); err != nil {
		return err
	}
	if err := expect(ctx, deliveries, "child-0", "drop", "child", paths); err != nil {
		return err
	}
	if err := fixture(ctx, "drop", "decline", mainTitle, paths); err != nil {
		return err
	}
	if err := expect(ctx, deliveries, "decline", "drop", "decline", paths); err != nil {
		return err
	}
	if err := expect(ctx, deliveries, "root", "drop", "root", paths); err != nil {
		return err
	}
	if err := fixture(ctx, "drop", "disabled", mainTitle, paths); err != nil {
		return err
	}
	if err := expect(ctx, deliveries, "root", "drop", "root", paths); err != nil {
		return err
	}
	for v := 1; v <= 2; v++ {
		if err := a.Do(func() { *version = v; a.Invalidate("drop") }); err != nil {
			return err
		}
		if _, err := waitElement(ctx, "drop", "child", fmt.Sprintf("File target %d", v)); err != nil {
			return err
		}
		if err := fixture(ctx, "drop", "child", mainTitle, paths); err != nil {
			return err
		}
		route, target := "root", "root"
		if v == 1 {
			route, target = "child-1", "child"
		}
		if err := expect(ctx, deliveries, route, "drop", target, paths); err != nil {
			return err
		}
	}
	if err := fixture(ctx, "second", "second-child", secondTitle, paths); err != nil {
		return err
	}
	if err := expect(ctx, deliveries, "second", "second", "second-root", paths); err != nil {
		return err
	}
	// Save a background point before the modal replaces the AX projection.
	background, err := waitElement(ctx, "drop", "help", "")
	if err != nil {
		return err
	}
	if err := a.Do(func() { *modal = true; a.Invalidate("drop") }); err != nil {
		return err
	}
	if err := fixture(ctx, "drop", "modal-child", mainTitle, paths); err != nil {
		return err
	}
	if err := expect(ctx, deliveries, "modal-child", "drop", "modal-child", paths); err != nil {
		return err
	}
	if err := expect(ctx, deliveries, "modal", "drop", "modal", paths); err != nil {
		return err
	}
	if err := receiveFileURLs(ctx, mainTitle, paths, background.Frame[0]+background.Frame[2]/2, background.Frame[1]+background.Frame[3]/2); err != nil {
		return err
	}
	// An ordered native key event is the barrier after the blocked background
	// drop. Escape dismisses the modal and the following refresh restores AX.
	native := core.TheApp.WindowByName(mainTitle)
	if native == nil {
		return errors.New("main native window missing")
	}
	event := events.NewKey(events.KeyChord, 0, key.CodeEscape, 0)
	event.Init()
	native.Events().Deque.Send(event)
	if _, err := waitElement(ctx, "drop", "help", ""); err != nil {
		return err
	}
	select {
	case d := <-deliveries:
		return fmt.Errorf("modal background drop leaked: %+v", d)
	default:
	}
	return nil
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
