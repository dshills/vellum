package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestNativeClosePolicyBeforeWidgetCleanup(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	accepted := false
	calls := 0
	spec := app.Window{ID: "main", Title: "Document", Width: 200, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }, CloseRequest: func() bool { calls++; return accepted }}
	if err := a.Open(spec); err != nil {
		t.Fatal(err)
	}
	w := &nativeWindow{window: fixtureWindow(a, spec.ID), body: core.NewBody("Document")}
	w.body.Scene.AddChild(w.body)
	w.body.Scene.Stage = &core.Stage{}
	bindCloseRequest(w, a, func(err error) { t.Error(err) })
	cleanup := 0
	w.body.OnClose(func(events.Event) { cleanup++ })
	child := core.NewButton(w.body)
	child.OnClose(func(events.Event) { cleanup++ })
	// Scene.Close walks the actual widget handlers even without a native stage.
	// With refusal, handled propagation prevents Body and child cleanup.
	w.body.Scene.Close()
	if calls != 1 || cleanup != 0 || len(a.Windows()) != 1 {
		t.Fatal("refused close reached cleanup", calls, cleanup)
	}
	accepted = true
	w.body.Scene.Close()
	if calls != 2 || cleanup != 2 || len(a.Windows()) != 1 {
		t.Fatal("permission gate committed teardown early", calls, cleanup)
	}
}
func TestNativeForcedCloseBypassesPolicy(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	calls := 0
	spec := app.Window{ID: "main", Title: "Document", Width: 200, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }, CloseRequest: func() bool { calls++; return false }}
	a.Open(spec)
	w := &nativeWindow{window: fixtureWindow(a, spec.ID), body: core.NewBody("Document")}
	w.body.Scene.AddChild(w.body)
	w.body.Scene.Stage = &core.Stage{}
	bindCloseRequest(w, a, nil)
	a.Close()
	cleanup := 0
	w.body.OnClose(func(events.Event) { cleanup++ })
	w.body.Scene.Close()
	if calls != 0 || cleanup != 1 {
		t.Fatal(calls, cleanup)
	}
}
