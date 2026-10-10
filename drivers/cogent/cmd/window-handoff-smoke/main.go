// window-handoff-smoke deterministically registers a replacement during the old
// window's native Close event, while its render lock prevents queued wake dispatch.
package main

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type handoff struct {
	owner    *app.Application
	next     app.Window
	opened   chan error
	attached atomic.Bool
}

func (h *handoff) Mount(parent tree.Node) core.Widget {
	f := core.NewFrame(parent)
	if h.attached.CompareAndSwap(false, true) {
		f.Scene.Body.OnFirst(events.Close, func(events.Event) { h.opened <- h.owner.Open(h.next) })
	}
	core.NewText(f).SetText("Temporary close-to-open handoff check")
	return f
}
func (*handoff) Sync(core.Widget) error { return nil }
func main() {
	a, newErr := app.New(app.Config{})
	if newErr != nil {
		fmt.Fprintln(os.Stderr, newErr)
		os.Exit(1)
	}
	defer a.Close()
	var secondViews atomic.Int32
	second := app.Window{ID: "second", Title: "Vellum handoff second", Width: 360, Height: 180, View: func() ui.Node {
		secondViews.Add(1)
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Replacement native window"}}}
	}}
	h := &handoff{owner: a, next: second, opened: make(chan error, 1)}
	a.Open(app.Window{ID: "first", Title: "Vellum handoff first", Width: 360, Height: 180, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Extension, Custom: h} }})
	timeout, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		finish := func(err error) { result <- err; a.Close() }
		for {
			if _, err := a.Services(); err == nil {
				break
			}
			select {
			case <-timeout.Done():
				finish(timeout.Err())
				return
			case <-time.After(20 * time.Millisecond):
			}
		}
		native := core.TheApp.WindowByName("Vellum handoff first")
		if native == nil {
			finish(fmt.Errorf("native first not found"))
			return
		}
		// Its OnFirst(Close) opens second under the old native render lock. The
		// normal Close handler must create it without depending on a Custom wake.
		native.CloseReq()
		select {
		case err := <-h.opened:
			if err != nil {
				finish(err)
				return
			}
		case <-timeout.Done():
			finish(timeout.Err())
			return
		}
		dispatched := make(chan struct{}, 1)
		if err := a.PostWindow("second", func() { a.Invalidate("second"); dispatched <- struct{}{} }); err != nil {
			finish(err)
			return
		}
		select {
		case <-dispatched:
		case <-timeout.Done():
			finish(timeout.Err())
			return
		}
		secondCtx, err := a.WindowContext("second")
		if err != nil || secondCtx.Err() != nil || a.Context().Err() != nil || secondViews.Load() == 0 {
			finish(fmt.Errorf("accepted handoff window lost: %v", err))
			return
		}
		result <- nil
		a.CloseWindow("second")
	}()
	err := a.Run(timeout, cogent.Driver{})
	select {
	case outcome := <-result:
		if err == nil {
			err = outcome
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
	fmt.Println("PASS accepted dynamic Open during last native Close survives and dispatches")
}
