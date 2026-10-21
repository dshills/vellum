package cogent

import (
	"image"
	"math"
	"path/filepath"
	"reflect"
	"testing"

	"cogentcore.org/core/base/fileinfo/mimedata"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
)

func dropGeometry(widget core.Widget) {
	w := widget.AsWidget()
	w.Geom.Pos.Total = math32.Vec2(10.25, -20.5) // scrolled origin outside clipped bounds
	w.Geom.TotalBBox = image.Rect(20, 0, 200, 200)
	w.Styles.UnitContext.DPI = units.DpPerInch * 2
}
func testFileDrop(t *testing.T) *events.DragDrop {
	t.Helper()
	e := events.NewExternalDrop(events.Drop, events.Left, image.Pt(30, 40), 0, mimedata.Mimes{mimedata.NewTextData(filepath.Join(t.TempDir(), "space Ω\nfile"))})
	e.Init()
	return e
}
func dropApp(t *testing.T) *app.Application {
	t.Helper()
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Close() })
	for _, id := range []string{"main", "other"} {
		if err := a.Open(app.Window{ID: id, Title: "headless", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}); err != nil {
			t.Fatal(err)
		}
	}
	return a
}
func TestFileDropBubblesWithPrivatePathsAndRetainedCallbacks(t *testing.T) {
	a := dropApp(t)
	host := core.NewFrame(core.NewBody("drop headless"))
	calls := []string{}
	var payload ui.FileDropEvent
	node := ui.Node{ID: "root", Kind: ui.Column, FileDrop: func(e ui.FileDropEvent) bool { calls = append(calls, e.ID); payload = e; return true }, Children: []ui.Node{{ID: "child", Kind: ui.Label, FileDrop: func(e ui.FileDropEvent) bool { calls = append(calls, e.ID); e.Paths[0] = "mutation"; return false }}}}
	r, err := reconcile(host, nil, node, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	child := r.children[0]
	dropGeometry(r.widget)
	dropGeometry(child.widget)
	e := testFileDrop(t)
	child.widget.AsWidget().HandleEvent(e)
	if e.IsHandled() {
		t.Fatal("declining child handled")
	}
	r.widget.AsWidget().HandleEvent(e)
	if !reflect.DeepEqual(calls, []string{"child", "root"}) || payload.Paths[0] == "mutation" || payload.ID != "root" || payload.WindowID != "main" || payload.Position != (ui.Point{X: 9.875, Y: 30.25}) || !e.IsHandled() || e.Target != r.widget || e.DropMod != events.DropCopy {
		t.Fatalf("drop routing: %v %+v %+v", calls, payload, e)
	}
	// Add a callback to the retained widget, then replace/remove it.
	retained := child.widget
	for _, callback := range []func(ui.FileDropEvent) bool{nil, func(e ui.FileDropEvent) bool { calls = append(calls, "updated"); return true }, nil} {
		node.Children[0].FileDrop = callback
		r, err = reconcile(host, r, node, a, "main")
		if err != nil {
			t.Fatal(err)
		}
		if r.children[0].widget != retained {
			t.Fatal("callback update remounted")
		}
		before := len(calls)
		event := testFileDrop(t)
		retained.AsWidget().HandleEvent(event)
		if callback == nil && (len(calls) != before || event.IsHandled()) {
			t.Fatal("removed callback ran")
		}
		if callback != nil && (len(calls) != before+1 || calls[len(calls)-1] != "updated" || !event.IsHandled()) {
			t.Fatal("updated callback missing")
		}
	}
	node.Children[0].FileDrop = func(ui.FileDropEvent) bool { calls = append(calls, "forbidden"); return true }
	r, err = reconcile(host, r, node, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	before := len(calls)
	for _, edit := range []func(*events.DragDrop){func(e *events.DragDrop) { e.Source = retained }, func(e *events.DragDrop) { e.Data = mimedata.Mimes{mimedata.NewTextData("relative")} }, func(e *events.DragDrop) { e.Data = mimedata.Mimes{nil} }, func(e *events.DragDrop) {
		e.Data = mimedata.Mimes{&mimedata.Data{Type: "text/uri-list", Data: []byte("file:///tmp/name")}}
	}, func(e *events.DragDrop) { e.Where = image.Pt(19, 40) }} {
		event := testFileDrop(t)
		edit(event)
		event.SetLocalOff(event.LocalOff())
		retained.AsWidget().HandleEvent(event)
		if event.IsHandled() || len(calls) != before {
			t.Fatal("invalid or clipped drop admitted")
		}
	}
	node.Disabled = true
	r, err = reconcile(host, r, node, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	retained.AsWidget().HandleEvent(testFileDrop(t))
	if len(calls) != before {
		t.Fatal("disabled ancestor admitted drop")
	}
	node.Disabled = false
	r, err = reconcile(host, r, node, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	for _, scale := range []float32{-1, float32(math.NaN()), float32(math.Inf(1))} {
		retained.AsWidget().Styles.UnitContext.DPI = scale
		retained.AsWidget().HandleEvent(testFileDrop(t))
		if len(calls) != before {
			t.Fatal("invalid scale admitted drop")
		}
	}
	dropGeometry(retained)
	if err := a.CloseWindow(fixtureWindowRef(a, "main")); err != nil {
		t.Fatal(err)
	}
	retained.AsWidget().HandleEvent(testFileDrop(t))
	if len(calls) != before {
		t.Fatal("closed window admitted drop")
	}
}
func TestFileDropInactiveTabsAndModalHost(t *testing.T) {
	a := dropApp(t)
	host := core.NewFrame(core.NewBody("drop pages"))
	calls := 0
	callback := func(ui.FileDropEvent) bool { calls++; return true }
	n := ui.Node{ID: "tabs", Kind: ui.Tabs, SelectedID: "one", Options: []ui.Option{{ID: "one", Label: "One"}, {ID: "two", Label: "Two"}}, Children: []ui.Node{{ID: "one", Kind: ui.Column, FileDrop: callback}, {ID: "two", Kind: ui.Column, FileDrop: callback}}}
	r, err := reconcile(host, nil, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	for _, child := range r.children {
		dropGeometry(child.widget)
	}
	r.children[1].widget.AsWidget().HandleEvent(testFileDrop(t))
	if calls != 0 {
		t.Fatal("inactive tab admitted")
	}
	r.children[0].widget.AsWidget().HandleEvent(testFileDrop(t))
	if calls != 1 {
		t.Fatal("active tab refused")
	}
	// Native selection is optimistic; a drop may precede the custom wake that
	// admits or refuses the controlled tab Change. The model still selects one.
	r.widget.(*core.Tabs).SelectTabByName("two")
	r.children[1].widget.AsWidget().HandleEvent(testFileDrop(t))
	if calls != 1 {
		t.Fatal("unaccepted native tab switch admitted a drop")
	}
	r.widget.(*core.Tabs).SelectTabByName("one")
	n.Options[0].Disabled = true
	r, err = reconcile(host, r, n, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	r.children[0].widget.AsWidget().HandleEvent(testFileDrop(t))
	if calls != 1 {
		t.Fatal("disabled selected page admitted")
	}
	modal := ui.Node{ID: "dialog", Kind: ui.Dialog, Open: true, FileDrop: callback, Children: []ui.Node{{ID: "modal-child", Kind: ui.Label, FileDrop: func(ui.FileDropEvent) bool { return false }}}}
	d, err := reconcile(host, nil, modal, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	body := prepareDialogBody(d)
	body.NewDialog(d.widget)
	d.control.dialog.active = true
	dropGeometry(d.control.dialog.host)
	dropGeometry(d.children[0].widget)
	event := testFileDrop(t)
	d.children[0].widget.AsWidget().HandleEvent(event)
	d.control.dialog.host.HandleEvent(event)
	if calls != 2 || !event.IsHandled() || event.Target != d.control.dialog.host {
		t.Fatal("visible modal host did not receive fallback")
	}
	d.node.Open = false
	d.control.dialog.host.HandleEvent(testFileDrop(t))
	if calls != 2 {
		t.Fatal("closed modal admitted")
	}
	d.node.Open = true
	d.control.dialog.dismissed = true
	d.control.dialog.host.HandleEvent(testFileDrop(t))
	if calls != 2 {
		t.Fatal("dismissed modal admitted")
	}
	disposeControlTree(d)
}
func TestFileDropInactiveWorkbenchPanelAndWindowIdentity(t *testing.T) {
	a := dropApp(t)
	model, err := workbench.New(workbench.Layout{Version: 1, Root: &workbench.Group{ID: "group", Tabs: []string{"one", "two"}, Active: "one"}}, []string{"one", "two"})
	if err != nil {
		t.Fatal(err)
	}
	calls := []string{}
	w := &Workbench{Owner: a, Window: "other", Workspace: model}
	for _, id := range []string{"one", "two"} {
		w.Panels = append(w.Panels, Panel{ID: id, Title: id, View: func() ui.Node {
			return ui.Node{ID: id, Kind: ui.Label, FileDrop: func(e ui.FileDropEvent) bool { calls = append(calls, e.ID+":"+e.WindowID); return true }}
		}})
	}
	widget := w.Mount(core.NewBody("drop dock"))
	if err := w.Sync(widget); err != nil {
		t.Fatal(err)
	}
	for _, r := range w.records {
		dropGeometry(r.widget)
	}
	w.records["two"].widget.AsWidget().HandleEvent(testFileDrop(t))
	if len(calls) != 0 {
		t.Fatal("inactive dock panel admitted")
	}
	w.records["one"].widget.AsWidget().HandleEvent(testFileDrop(t))
	if !reflect.DeepEqual(calls, []string{"one:other"}) {
		t.Fatal(calls)
	}
}
