// window-lifetime-smoke runs one real native lifetime regression per process.
// Run each case separately: the native main loop runs once per process.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

const invalidKind ui.Kind = "window-smoke-invalid"

type closingExtension struct {
	a        *app.Application
	closed   atomic.Bool
	mountErr error // accessed only on the application owner
}

func (x *closingExtension) Mount(parent tree.Node) core.Widget {
	widget := core.NewFrame(parent)
	scene := widget.Scene
	if scene == nil || scene.Body == nil {
		x.mountErr = errors.New("closing extension has no native body")
		return widget
	}
	scene.Body.OnClose(func(events.Event) { x.closed.Store(true) })
	owner, ref, ok := cogent.WindowScope(parent)
	if !ok || owner != x.a {
		x.mountErr = app.ErrInvalidWindowRef
	} else {
		x.mountErr = x.a.CloseWindow(ref)
	}
	return widget
}
func (x *closingExtension) Sync(core.Widget) error { return x.mountErr }

func waitUntil(ctx context.Context, check func() (bool, error)) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		ready, err := check()
		if err != nil || ready {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func run(which string) error {
	if which != "invalid-view" && which != "close-before-show" && which != "initial-invalid-view" {
		return fmt.Errorf("unknown case %q; use invalid-view, close-before-show, or initial-invalid-view", which)
	}
	// A hard watchdog makes a broken native shutdown fail the process instead of
	// hanging the regression runner. The cooperative deadline allows cleanup first.
	watchdog := time.AfterFunc(8*time.Second, func() {
		fmt.Fprintf(os.Stderr, "FAIL %s: native loop did not finish within eight seconds\n", which)
		os.Exit(1)
	})
	defer watchdog.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	if which == "initial-invalid-view" {
		return runInitialInvalid(ctx, a)
	}
	invalid := false // owner-only model
	var delivered atomic.Bool
	var workerErr atomic.Pointer[error]
	failWorker := func(err error) {
		if err != nil {
			workerErr.CompareAndSwap(nil, &err)
			cancel()
		}
	}
	mainRef, err := a.OpenWindow(app.Window{ID: "main", Title: "Vellum lifetime smoke: " + which, Width: 360, Height: 180, View: func() ui.Node {
		if invalid {
			return ui.Node{ID: "root", Kind: invalidKind}
		}
		return ui.Node{ID: "root", Kind: ui.Label, Text: "Window lifetime regression"}
	}})
	if err != nil {
		return err
	}
	var closingRef app.WindowRef
	closing := &closingExtension{a: a}
	if which == "close-before-show" {
		closingRef, err = a.OpenWindow(app.Window{ID: "closing", Title: "Closing before show", Width: 280, Height: 140, View: func() ui.Node {
			return ui.Node{ID: "closing-root", Kind: ui.Extension, Custom: closing}
		}})
		if err != nil {
			return err
		}
	}
	go func() {
		err := waitUntil(ctx, func() (bool, error) {
			_, err := a.Services()
			if errors.Is(err, app.ErrUnavailable) {
				return false, nil
			}
			return err == nil, err
		})
		if err != nil {
			failWorker(fmt.Errorf("services never became ready: %w", err))
			return
		}
		if which == "close-before-show" {
			if err := waitUntil(ctx, func() (bool, error) { return closing.closed.Load(), nil }); err != nil {
				failWorker(fmt.Errorf("closing native window did not close: %w", err))
				return
			}
		}
		err = a.PostWindow(mainRef, func() {
			if which == "invalid-view" {
				invalid = true
				delivered.Store(true)
				a.Invalidate("main")
				return
			}
			mainLife, err := a.WindowContext("main")
			if err != nil || mainLife.Err() != nil {
				failWorker(fmt.Errorf("main window stopped before delivery: %v", err))
				return
			}
			closingLife := closingRef.Context()
			if closingLife.Err() == nil || !closing.closed.Load() {
				failWorker(errors.New("closing lifetime/native close was not completed"))
				return
			}
			delivered.Store(true)
			a.Close()
		})
		if err != nil {
			failWorker(fmt.Errorf("main window dispatch refused: %w", err))
		}
	}()
	err = a.Run(ctx, cogent.Driver{})
	if failed := workerErr.Load(); failed != nil {
		return *failed
	}
	if ctx.Err() != nil {
		return fmt.Errorf("%s deadline/shutdown failure: %w", which, ctx.Err())
	}
	if !delivered.Load() {
		return fmt.Errorf("%s returned without main dispatch", which)
	}
	if which == "invalid-view" {
		if err == nil || !strings.Contains(err.Error(), "unknown kind") || !strings.Contains(err.Error(), string(invalidKind)) {
			return fmt.Errorf("invalid View did not return its validation error: %v", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("close-before-show Run failed: %w", err)
	}
	if !closing.closed.Load() {
		return errors.New("closing native window remained open")
	}
	return nil
}

// pendingExtension observes construction and destruction without launching a
// native window. WidgetBase embeds NodeBase, whose This becomes nil on Destroy.
type pendingExtension struct {
	mounted core.Widget
	scene   *core.Scene
	shown   atomic.Bool
}

func (x *pendingExtension) Mount(parent tree.Node) core.Widget {
	x.mounted = core.NewFrame(parent)
	x.scene = x.mounted.AsWidget().Scene
	if x.scene != nil && x.scene.Body != nil {
		x.scene.Body.OnShow(func(events.Event) { x.shown.Store(true) })
	}
	return x.mounted
}
func (x *pendingExtension) Sync(core.Widget) error { return nil }

func runInitialInvalid(ctx context.Context, a *app.Application) error {
	pending := &pendingExtension{}
	if err := a.Open(app.Window{ID: "first", Title: "Unlaunched valid window", Width: 320, Height: 180, View: func() ui.Node {
		return ui.Node{ID: "first-root", Kind: ui.Extension, Custom: pending}
	}}); err != nil {
		return err
	}
	if err := a.Open(app.Window{ID: "invalid", Title: "Invalid second window", Width: 320, Height: 180, View: func() ui.Node {
		return ui.Node{ID: "invalid-root", Kind: invalidKind}
	}}); err != nil {
		return err
	}
	started := time.Now()
	err := a.Run(ctx, cogent.Driver{})
	if err == nil || !strings.Contains(err.Error(), "unknown kind") || !strings.Contains(err.Error(), string(invalidKind)) {
		return fmt.Errorf("initial invalid View did not return its validation error: %v", err)
	}
	if ctx.Err() != nil || time.Since(started) >= 3*time.Second {
		return fmt.Errorf("initial validation failed to return promptly (elapsed %s): %v", time.Since(started), ctx.Err())
	}
	if pending.mounted == nil {
		return errors.New("first valid window was not constructed before invalid second View")
	}
	if pending.scene == nil {
		return errors.New("first valid window had no construction scene")
	}
	if pending.shown.Load() || pending.scene.RenderWindow() != nil {
		return errors.New("initial validation failure started a native window")
	}
	if pending.mounted.AsTree().This != nil || !tree.IsNil(pending.mounted) {
		return errors.New("abandoned first-window widget was not destroyed")
	}
	return nil
}

func main() {
	which := flag.String("case", "", "invalid-view, close-before-show, or initial-invalid-view (one case per process)")
	flag.Parse()
	if err := run(*which); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("PASS:", *which)
}
