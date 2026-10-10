package cogent

import (
	"context"
	"errors"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/tree"
	"example.com/vellum/app"
	"example.com/vellum/ui"
)

func TestOtherWindowDrainDefersNativeControlRestoration(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	for _, id := range []string{"origin", "draining"} {
		if err := a.Open(app.Window{ID: id, Title: id, Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
			t.Fatal(err)
		}
	}
	host := core.NewFrame(core.NewBody("origin native owner"))
	changed := 0
	n := ui.Node{ID: "root", Kind: ui.Column, Disabled: true, Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "accepted", Change: func(ui.ChangeEvent) { changed++ }}}}
	r, err := reconcile(host, nil, n, a, "origin")
	if err != nil {
		t.Fatal(err)
	}
	field := findRecord(r, "field")
	native := field.widget.(*core.TextField)
	native.SetText("unaccepted").SendChange()
	// A different window (or worker) owns this Drain; it must not touch origin
	// widget state, traverse its native ancestors, or acquire its native lock.
	drained := make(chan error, 1)
	go func() { drained <- a.Drain() }()
	if err := <-drained; err != nil {
		t.Fatal(err)
	}
	if changed != 0 || native.EditingState().Text != "unaccepted" || !field.control.restore {
		t.Fatal("queued callback mutated another window's native control")
	}
	// Origin's next render-owner refresh performs the controlled-value restore.
	r, err = reconcile(host, r, n, a, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if native.EditingState().Text != "accepted" || field.control.restore {
		t.Fatal("native owner failed to apply deferred restoration")
	}
}

func TestRetiredControlCallbackUsesLogicalLifetime(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(app.Window{ID: "main", Title: "main", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }})
	host := core.NewFrame(core.NewBody("headless"))
	changes := 0
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "one", Change: func(ui.ChangeEvent) { changes++ }}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	field := findRecord(r, "field")
	field.widget.(*core.TextField).SetText("queued").SendChange()
	n.Children = nil
	if err := a.Do(func() { r, err = reconcile(host, r, n, a, "main") }); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if !field.control.retired || !tree.IsNil(field.widget) {
		t.Fatal("removed control lacks logical retirement")
	}
	drained := make(chan error, 1)
	go func() { drained <- a.Drain() }()
	if err := <-drained; err != nil {
		t.Fatal(err)
	}
	if changes != 0 {
		t.Fatal("retired callback reached model")
	}
}

func TestNativeControlCleanupAfterApplicationOwnerBarrier(t *testing.T) {
	a, _ := app.New(app.Config{})
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "dialog", Kind: ui.Dialog, Open: true, Children: []ui.Node{{ID: "field", Kind: ui.TextField, Text: "retained"}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	body := prepareDialogBody(r)
	body.NewDialog(r.widget)
	r.control.dialog.active = true
	a.Close()
	// Do acquires/releases the owner even when rejecting closed model work. The
	// native close handler can then retire/clean its tree despite canceled lifetime.
	if err := a.Do(func() { t.Fatal("closed callback ran") }); err != app.ErrClosed {
		t.Fatal(err)
	}
	disposeControlTree(r)
	if !r.control.retired || r.control.dialog.active || r.control.dialog.body != nil || r.control.dialog.host.Parent != r.widget {
		t.Fatal("final-close cleanup skipped native dialog resources")
	}
	if a.Context().Err() != context.Canceled {
		t.Fatal("lifetime not ended")
	}
}

func TestDisabledSnapshotAfterReparentAndAncestorUpdate(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	changes := 0
	field := ui.Node{ID: "field", Kind: ui.TextField, Text: "accepted", Change: func(ui.ChangeEvent) { changes++ }}
	n := ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "enabled", Kind: ui.Column, Children: []ui.Node{field}}, {ID: "disabled", Kind: ui.Column, Disabled: true}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	native := findRecord(r, "field").widget.(*core.TextField)
	native.SetText("queued before move").SendChange()
	n.Children[0].Children = nil
	n.Children[1].Children = []ui.Node{field}
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if !findRecord(r, "field").control.disabled || findRecord(r, "field").widget != native {
		t.Fatal("snapshot ignored new disabled ancestor")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if changes != 0 {
		t.Fatal("queued edit accepted after move into disabled ancestor")
	}
	n.Children[1].Disabled = false
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	if findRecord(r, "field").control.disabled {
		t.Fatal("unchanged child retained ancestor disabled snapshot")
	}
	native.SetText("enabled edit").SendChange()
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if changes != 1 {
		t.Fatal("enabled ancestor transition refused edit")
	}
}

func TestRefusalDoesNotReadConcurrentOwnerState(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.Open(app.Window{ID: "main", Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("headless"))
	n := ui.Node{ID: "field", Kind: ui.TextField, Text: "accepted"}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	reported := 0
	bindControlWake(r, func(events.Event) {}, func(err error) {
		if !errors.Is(err, app.ErrQueueFull) {
			t.Errorf("refusal error: %v", err)
		}
		reported++
	})
	if err := a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- a.Do(func() {
			for i := 0; i < 2000; i++ {
				r.node.Text = "owner update"
			}
			r.node.Text = "accepted"
		})
	}()
	for i := 0; i < 2000; i++ {
		postControl(r, func() { t.Error("refused callback delivered") })
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if reported != 0 || r.control.refused.Load() == nil {
		t.Fatal("off-owner refusal callback or missing handoff")
	}
	if err := a.Do(func() { r, err = reconcile(host, r, n, a, "main") }); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if reported != 1 || r.control.refused.Load() != nil || r.widget.(*core.TextField).EditingState().Text != "accepted" {
		t.Fatal("owner failed to consume coalesced refusal")
	}
}
