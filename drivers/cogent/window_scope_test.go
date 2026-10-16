package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/tree"
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"github.com/dshills/vellum/workbench"
	"image"
	"testing"
)

func scopeDeclaration(id string) app.Window {
	return app.Window{ID: id, Title: id, Width: 400, Height: 240, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}
}
func TestWindowScopeNeverRecapturesFromRetiredParent(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	if err := a.Open(scopeDeclaration("survivor")); err != nil {
		t.Fatal(err)
	}
	old, err := a.OpenWindow(scopeDeclaration("same"))
	if err != nil {
		t.Fatal(err)
	}
	host := core.NewBody("scope")
	bindWindowScope(host, a, old)
	child := core.NewFrame(host)
	if owner, ref, ok := WindowScope(child); !ok || owner != a || ref != old {
		t.Fatal("scope not inherited")
	}
	if err = a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	replacement, err := a.OpenWindow(scopeDeclaration("same"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = captureMountScope(child, a, "same"); !errors.Is(err, app.ErrWindowClosed) {
		t.Fatal("retired parent rebound", err)
	}
	if ref, err := captureMountScope(core.NewBody("new setup"), a, "same"); err != nil || ref != replacement {
		t.Fatal("direct setup cannot capture current opening", err)
	}
	if _, err = mountScoped(child, ui.Node{ID: "late", Kind: ui.Button}, a, old); !errors.Is(err, app.ErrWindowClosed) {
		t.Fatal("late managed mount rebound", err)
	}
	foreign, _ := app.New(app.Config{})
	defer foreign.Close()
	if _, err = captureMountScope(child, foreign, "same"); !errors.Is(err, app.ErrInvalidWindowRef) {
		t.Fatal("foreign owner accepted", err)
	}
	host.Scene.Destroy()
}
func TestNativeWindowRefScopesCloseCommandsAndAccessibility(t *testing.T) {
	a, oldBridge := newAXTest(t, 16)
	if err := a.Open(scopeDeclaration("survivor")); err != nil {
		t.Fatal(err)
	}
	oldWindow := fixtureWindow(a, "ax")
	old := oldWindow.Ref()
	commandCalls, oldActions, newPolicyCalls := 0, 0, 0
	registry := command.New()
	if err := registry.Register(command.Command{ID: "action", Label: "Action", Shortcut: "Control+K", Execute: func(context.Context) error { commandCalls++; return nil }}); err != nil {
		t.Fatal(err)
	}
	body := core.NewBody("retired source")
	body.Scene.AddChild(body)
	body.Scene.Stage = &core.Stage{}
	native := &nativeWindow{window: oldWindow, body: body}
	bindCommandShortcutsScoped(body.Scene, a, old, registry, nil)
	bindCloseRequest(native, a, nil)
	oldBridge.mu.Lock()
	oldBridge.handlers["button"] = func(accessibilityAction) bool { oldActions++; return true }
	oldBridge.mu.Unlock()
	oldBridge.generation.Store(1)
	if !accessibilityDeliver(oldBridge.token, 1, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("original action not queued")
	}
	if err := a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	spec := scopeDeclaration("ax")
	spec.CloseRequest = func() bool { newPolicyCalls++; return false }
	replacement, err := a.OpenWindow(spec)
	if err != nil {
		t.Fatal(err)
	}
	event := events.NewKey(events.KeyChord, 'k', key.CodeK, key.Modifiers(1<<key.Control))
	body.Scene.Listeners.First.Call(event)
	body.Scene.Close() // forced retirement must bypass the replacement's policy
	if event.IsHandled() || commandCalls != 0 || newPolicyCalls != 0 {
		t.Fatal("retired source reached replacement", commandCalls, newPolicyCalls)
	}
	if _, err = lookupAccessibility("ax"); err == nil {
		t.Fatal("old bridge exposed as replacement")
	}
	if accessibilityDeliver(oldBridge.token, 1, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("old native action admitted after reuse")
	}
	next := newAccessibilityBridge(a)
	defer next.close()
	next.bindScoped(replacement)
	next.mu.Lock()
	next.handlers["button"] = func(accessibilityAction) bool { commandCalls++; return true }
	next.mu.Unlock()
	next.generation.Store(1)
	oldBridge.bindScoped(old)
	oldBridge.bindScoped(replacement)
	oldBridge.close()
	if got, err := lookupAccessibility("ax"); err != nil || got != next {
		t.Fatal("old publication or cleanup displaced replacement", err)
	}
	if !accessibilityDeliver(next.token, 1, "button", accessibilityAction{Kind: "press"}) {
		t.Fatal("replacement action refused")
	}
	if err = a.Drain(); err != nil {
		t.Fatal(err)
	}
	if oldActions != 0 || commandCalls != 1 || replacement.Context().Err() != nil {
		t.Fatal("scope isolation failed", oldActions, commandCalls)
	}
	body.Scene.Destroy()
}
func TestWindowScopeGPUCallbacksKeepTheirOrigin(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(scopeDeclaration("survivor"))
	old, _ := a.OpenWindow(scopeDeclaration("same"))
	// Exercise real owner-event fences without creating a device or submitting GPU work.
	view := &GPUView{Owner: a, Window: "same"}
	vw := &gpuViewWidget{owner: view, application: a, windowID: "same", windowRef: old}
	view.widget = vw
	mesh := &MeshCanvas{Owner: a, Window: "same"}
	mw := &meshWidget{owner: mesh, application: a, windowID: "same", windowRef: old}
	mesh.widget = mw
	canvas := &GPUCanvas{Owner: a, Window: "same"}
	cw := &canvasWidget{owner: canvas, application: a, windowRef: old}
	canvas.widget = cw
	if err := a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	a.Open(scopeDeclaration("same"))
	calls := 0
	vw.ownerEvent(func() { calls++ })
	mw.ownerEvent(func() { calls++ })
	cw.ownerEvent(func() { calls++ })
	if calls != 0 {
		t.Fatal("retired GPU event rebound", calls)
	}
}

func TestWindowScopeEditorHistoryAfterRetirement(t *testing.T) {
	for _, redo := range []bool{false, true} {
		t.Run(map[bool]string{false: "Undo", true: "Redo"}[redo], func(t *testing.T) {
			a, w, _, r, model := editorFixture(t, "before")
			a.Open(scopeDeclaration("survivor"))
			if err := a.Do(func() {
				if err := model.Replace("after", editor.Selection{Anchor: 5, Focus: 5}); err != nil {
					t.Error(err)
				}
				if redo {
					if err := model.Journal().Undo(); err != nil {
						t.Error(err)
					}
				}
				syncEditor(r)
			}); err != nil {
				t.Fatal(err)
			}
			before := model.Snapshot()
			if err := a.CloseWindow(w.window.Ref()); err != nil {
				t.Fatal(err)
			}
			replacement, err := a.OpenWindow(scopeDeclaration(w.window.ID))
			if err != nil {
				t.Fatal(err)
			}
			modifiers := key.Modifiers(1 << key.Meta)
			if redo {
				modifiers |= 1 << key.Shift
			}
			r.widget.AsWidget().Send(events.KeyChord, events.NewKey(events.KeyChord, 'z', key.CodeZ, modifiers))
			after := model.Snapshot()
			if after.Text != before.Text || after.History != before.History || replacement.Context().Err() != nil {
				t.Fatal("retired editor changed shared history", before, after)
			}
		})
	}
}
func TestWindowScopeLegacyVirtualListenersRetainEachMount(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(scopeDeclaration("survivor"))
	old, _ := a.OpenWindow(scopeDeclaration("same"))
	tableModel, _ := virtual.NewTable(largeSource{}, 4)
	nodes := []virtual.TreeNode{}
	for i := 0; i < 12; i++ {
		nodes = append(nodes, virtual.TreeNode{ID: fmt.Sprint(i)})
	}
	treeModel, _ := virtual.NewTree(nodes)
	calls := 0
	table := &VirtualTable{Owner: a, Window: "same", Model: tableModel, Select: func(virtual.Row) { calls++ }}
	outline := &VirtualTree{Owner: a, Window: "same", Model: treeModel, Count: 4, Select: func(string) { calls++ }}
	host := core.NewBody("old virtual scopes")
	bindWindowScope(host, a, old)
	if native := table.Mount(host); native == nil {
		t.Fatal("table mount")
	} else if err := table.Sync(native); err != nil {
		t.Fatal(err)
	}
	if native := outline.Mount(host); native == nil {
		t.Fatal("tree mount")
	} else if err := outline.Sync(native); err != nil {
		t.Fatal(err)
	}
	row, treeRow, tableHost, treeHost := table.rows[0], outline.rows[0], table.host, outline.host
	pager := outline.host.Children[0].(*core.Frame).Children[1].(*core.Button)
	stale := func() {
		row.Send(events.Click)
		treeRow.Send(events.Click)
		pager.Send(events.Click)
		tableHost.Send(events.Scroll, events.NewScroll(image.Point{}, math32.Vector2{Y: 96}, 0))
		treeHost.Send(events.Scroll, events.NewScroll(image.Point{}, math32.Vector2{Y: 96}, 0))
	}
	if err := a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	replacement, err := a.OpenWindow(scopeDeclaration("same"))
	if err != nil {
		t.Fatal(err)
	}
	stale()
	nextHost := core.NewBody("new virtual scopes")
	bindWindowScope(nextHost, a, replacement)
	table.Mount(nextHost)
	outline.Mount(nextHost)
	table.Sync(table.host)
	outline.Sync(outline.host)
	stale() // reusing the public extension object must not retarget old listeners
	if calls != 0 || tableModel.Selected != "" || treeModel.Selected != "" || tableModel.StartRow() != 0 || outline.Start != 0 {
		t.Fatal("old virtual input changed shared state", calls, tableModel.StartRow(), outline.Start)
	}
	table.rows[0].Send(events.Click)
	outline.rows[0].Send(events.Click)
	if calls != 2 || tableModel.Selected != "0" || treeModel.Selected != "0" {
		t.Fatal("new mount not independently active", calls)
	}
	host.Scene.Destroy()
	nextHost.Scene.Destroy()
}
func TestWindowScopeRechecksAfterOwnerBarrier(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(scopeDeclaration("survivor"))
	old, _ := a.OpenWindow(scopeDeclaration("same"))
	held, release, finished := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() { finished <- a.Do(func() { close(held); <-release }) }()
	<-held
	started, result := make(chan struct{}), make(chan error, 1)
	calls := 0
	go func() { close(started); result <- withWindowOwner(a, old, func() { calls++ }) }()
	<-started
	if err := a.CloseWindow(old); err != nil {
		close(release)
		t.Fatal(err)
	}
	replacement, err := a.OpenWindow(scopeDeclaration("same"))
	if err != nil {
		close(release)
		t.Fatal(err)
	}
	close(release)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	if err := <-result; !errors.Is(err, app.ErrWindowClosed) {
		t.Fatal("retired initialization was not reported as skipped", err)
	}
	if calls != 0 || replacement.Context().Err() != nil {
		t.Fatal("old callback crossed owner barrier", calls)
	}
	if err := withWindowOwner(a, replacement, func() { calls++ }); err != nil || calls != 1 {
		t.Fatal("replacement callback refused", calls, err)
	}
}

func TestWindowScopeWorkbenchSyncRefusesBeforeUserCallbacks(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(scopeDeclaration("survivor"))
	old, _ := a.OpenWindow(scopeDeclaration("same"))
	model, err := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "dock", Tabs: []string{"one", "two"}, Active: "one"}}, []string{"one", "two"})
	if err != nil {
		t.Fatal(err)
	}
	panels, reported := 0, 0
	dock := &Workbench{Owner: a, Window: "same", Workspace: model, OnError: func(error) { reported++ }}
	for _, id := range []string{"one", "two"} {
		id := id
		dock.Panels = append(dock.Panels, Panel{ID: id, Title: id, View: func() ui.Node { panels++; return ui.Node{ID: "content", Kind: ui.Label, Text: id} }})
	}
	host := core.NewBody("old direct sync")
	bindWindowScope(host, a, old)
	if native := dock.Mount(host); native == nil {
		t.Fatal("mount")
	} else if err = dock.Sync(native); err != nil {
		t.Fatal(err)
	}
	previous := dock.records["one"]
	children := dock.host.NumChildren()
	signature := dock.signature
	before := panels
	if err = model.Split("two", "dock", "split", workbench.Horizontal, false); err != nil {
		t.Fatal(err)
	}
	pending := errors.New("queued old error")
	dock.pendingError.Store(&pending)
	if err = a.CloseWindow(old); err != nil {
		t.Fatal(err)
	}
	replacement, err := a.OpenWindow(scopeDeclaration("same"))
	if err != nil {
		t.Fatal(err)
	}
	var syncErr error
	if err = a.Do(func() { syncErr = dock.Sync(dock.host) }); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(syncErr, app.ErrWindowClosed) {
		t.Fatal("stale Sync accepted", syncErr)
	}
	if panels != before || reported != 0 || dock.pendingError.Load() != &pending || dock.records["one"] != previous || dock.signature != signature || dock.host.NumChildren() != children || replacement.Context().Err() != nil {
		t.Fatal("stale Sync ran callbacks or rebuilt native records", panels, before, reported)
	}
	dock.disposeControlRecords()
	host.Scene.Destroy()
}

func TestWindowScopeOldCanvasMouseDownCannotFocusRemount(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(scopeDeclaration("survivor"))
	original, _ := a.OpenWindow(scopeDeclaration("same"))
	model, _ := canvas.New(nil)
	model.View.SetSize(100, 100)
	g := &GPUCanvas{Owner: a, Window: "same", Model: model}
	host := core.NewBody("old canvas")
	bindWindowScope(host, a, original)
	if g.Mount(host) == nil {
		t.Fatal("old mount")
	}
	old := g.widget
	if err := a.CloseWindow(original); err != nil {
		t.Fatal(err)
	}
	replacement, err := a.OpenWindow(scopeDeclaration("same"))
	if err != nil {
		t.Fatal(err)
	}
	// Simulate the retired-widget marker while preserving a listener already
	// retained by an originating native dispatch. No GPU/device is created.
	old.retired.Store(true)
	next := core.NewBody("new canvas")
	bindWindowScope(next, a, replacement)
	if g.Mount(next) == nil {
		t.Fatal("new mount")
	}
	focused := 0
	g.widget.OnFocus(func(events.Event) { focused++ })
	old.Send(events.MouseDown, events.NewMouse(events.MouseDown, events.Left, image.Pt(5, 5), 0))
	if focused != 0 {
		t.Fatal("retired native listener focused replacement", focused)
	}
	g.widget.Send(events.MouseDown, events.NewMouse(events.MouseDown, events.Left, image.Pt(5, 5), 0))
	if focused != 1 {
		t.Fatal("current canvas focus refused", focused)
	}
	host.Scene.Destroy()
	next.Scene.Destroy()
}

type retiringNilScopeExtension struct {
	owner *app.Application
	ref   app.WindowRef
}

func (e *retiringNilScopeExtension) Mount(tree.Node) core.Widget {
	_ = e.owner.CloseWindow(e.ref)
	return nil
}
func (e *retiringNilScopeExtension) Sync(core.Widget) error { return nil }
func TestWindowScopeRetirementDuringMountKeepsClosedClassification(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	a.Open(scopeDeclaration("survivor"))
	original, _ := a.OpenWindow(scopeDeclaration("same"))
	host := core.NewBody("retiring mount")
	bindWindowScope(host, a, original)
	var mountErr error
	if err := a.Do(func() {
		_, mountErr = mountShallowScoped(host, ui.Node{ID: "ext", Kind: ui.Extension, Custom: &retiringNilScopeExtension{a, original}}, a, original)
	}); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(mountErr, app.ErrWindowClosed) || a.Context().Err() != nil {
		t.Fatal("expected retirement became generic driver failure", mountErr)
	}
	replacement, err := a.OpenWindow(scopeDeclaration("same"))
	if err != nil {
		t.Fatal(err)
	}
	next := core.NewBody("live invalid mount")
	bindWindowScope(next, a, replacement)
	if err = a.Do(func() {
		_, mountErr = mountShallowScoped(next, ui.Node{ID: "ext", Kind: ui.Extension, Custom: nilExtension{}}, a, replacement)
	}); err != nil {
		t.Fatal(err)
	}
	if mountErr == nil || errors.Is(mountErr, app.ErrWindowClosed) || replacement.Context().Err() != nil {
		t.Fatal("live invalid mount error hidden", mountErr)
	}
	host.Scene.Destroy()
	next.Scene.Destroy()
}
