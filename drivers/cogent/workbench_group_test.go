package cogent

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/system"
	"example.com/vellum/app"
	"example.com/vellum/ui"
	"example.com/vellum/workbench"
)

func mustEncodeSession(t *testing.T, s *workbench.Session) []byte {
	t.Helper()
	data, err := s.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func groupFixture(t *testing.T) (*app.Application, *WorkbenchGroup) {
	t.Helper()
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	s, err := workbench.NewSession(workbench.SessionLayout{Version: workbench.SessionVersion, Windows: []workbench.WindowLayout{{ID: "main", Title: "Main", Bounds: workbench.Bounds{Width: 900, Height: 640}, Root: &workbench.Group{ID: "documents", Tabs: []string{"one", "two", "three"}, Active: "one"}}}, Focus: workbench.Focus{Window: "main", Panel: "one"}}, []string{"one", "two", "three"})
	if err != nil {
		t.Fatal(err)
	}
	panels := []Panel{}
	for _, id := range s.Panels() {
		id := id
		panels = append(panels, Panel{ID: id, Title: id, View: func() ui.Node { return ui.Node{ID: id, Kind: ui.Label, Text: id} }})
	}
	g, err := NewWorkbenchGroup(a, s, panels)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Open(); err != nil {
		t.Fatal(err)
	}
	return a, g
}
func mountGroupWindow(t *testing.T, w *Workbench) {
	t.Helper()
	native := w.Mount(core.NewBody("headless"))
	if err := w.Sync(native); err != nil {
		t.Fatal(err)
	}
}
func TestWorkbenchGroupNativeRecordsStayInTheirWindow(t *testing.T) {
	a, g := groupFixture(t)
	main := g.windows["main"]
	mountGroupWindow(t, main)
	before := main.records["two"].widget
	if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal(err)
	}
	var floating string
	for _, win := range g.Session.Layout().Windows {
		if win.ID != "main" {
			floating = win.ID
		}
	}
	float := g.windows[floating]
	mountGroupWindow(t, float)
	if err := main.Sync(main.host); err != nil {
		t.Fatal(err)
	}
	if float.records["two"].widget == before || main.records["two"] != nil {
		t.Fatal("native panel crossed windows")
	}
	nativeID := float.Window
	if err := g.Apply(func(s *workbench.Session) error { return s.Dock("two", "main", "documents", 0) }); err != nil {
		t.Fatal(err)
	}
	if err := main.Sync(main.host); err != nil {
		t.Fatal(err)
	}
	ctx, _ := a.WindowContext(nativeID)
	if ctx.Err() == nil || len(a.Windows()) != 1 || len(g.Session.Layout().Windows) != 1 {
		t.Fatal("empty floating native registration survived")
	}
	if main.records["two"].widget == float.records["two"].widget {
		t.Fatal("redock reparented foreign native widget")
	}
}
func TestWorkbenchGroupOpenCloseAndPersistenceFailuresCompensate(t *testing.T) {
	a, g := groupFixture(t)
	before := mustEncodeSession(t, g.Session)
	g.open = func(app.Window) error { return errors.New("injected Open failure") }
	if g.Float("two", workbench.Bounds{Width: 500, Height: 400}) == nil {
		t.Fatal("Open failure ignored")
	}
	after := mustEncodeSession(t, g.Session)
	if !bytes.Equal(before, after) || len(a.Windows()) != 1 || len(g.windows) != 1 || a.Context().Err() != nil {
		t.Fatal("Open rollback diverged")
	}
	g.open = a.Open
	if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal(err)
	}
	var floating string
	for id := range g.windows {
		if id != "main" {
			floating = id
		}
	}
	before = mustEncodeSession(t, g.Session)
	g.close = func(string) error { return errors.New("injected Close failure") }
	if g.Close(floating) == nil {
		t.Fatal("Close failure ignored")
	}
	after = mustEncodeSession(t, g.Session)
	if !bytes.Equal(before, after) || len(a.Windows()) != 2 || a.Context().Err() != nil {
		t.Fatal("Close rollback diverged")
	}
	g.close = a.CloseWindow
	oldID := g.windows[floating].Window
	g.PersistencePath = filepath.Join(t.TempDir(), "missing", "session.json")
	if g.Close(floating) == nil {
		t.Fatal("save failure ignored")
	}
	after = mustEncodeSession(t, g.Session)
	if !bytes.Equal(before, after) || len(a.Windows()) != 2 || g.windows[floating].Window == oldID || a.Context().Err() != nil {
		t.Fatal("save compensation did not reopen old session with fresh registration")
	}
	old, _ := a.WindowContext(oldID)
	if old.Err() == nil {
		t.Fatal("retired identity reused")
	}
}
func TestWorkbenchGroupRollbackFailureClosesApplication(t *testing.T) {
	a, g := groupFixture(t)
	if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal(err)
	}
	var floating string
	for id := range g.windows {
		if id != "main" {
			floating = id
		}
	}
	before := mustEncodeSession(t, g.Session)
	g.PersistencePath = filepath.Join(t.TempDir(), "missing", "session.json")
	g.open = func(app.Window) error { return errors.New("cannot restore native window") }
	if g.Close(floating) == nil || a.Context().Err() == nil {
		t.Fatal("failed compensation left incoherent running application")
	}
	after := mustEncodeSession(t, g.Session)
	if !bytes.Equal(before, after) {
		t.Fatal("failed compensation mutated domain session")
	}
}
func TestWorkbenchNativeCloseTransfersFocusedPanel(t *testing.T) {
	a, g := groupFixture(t)
	if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal(err)
	}
	var floating *Workbench
	for id, w := range g.windows {
		if id != "main" {
			floating = w
		}
	}
	if err := floating.WindowClosed(); err != nil {
		t.Fatal(err)
	}
	if len(a.Windows()) != 1 || g.Session.Layout().Focus.Window != "main" || g.Session.Layout().Focus.Panel != "two" {
		t.Fatal("native close did not redock")
	}
	if err := floating.WindowClosed(); err != nil {
		t.Fatal("retired close was not idempotent", err)
	}
}
func TestWorkbenchPointerReorderRunsThroughOwnerMailbox(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	w.tabs["documents"].Geom.TotalBBox = image.Rect(0, 0, 600, 400)
	for i, id := range []string{"one", "two", "three"} {
		w.buttons[id].Geom.TotalBBox = image.Rect(i*100, 0, (i+1)*100, 30)
	}
	event := events.NewMouse(events.SlideStop, events.Left, image.Pt(299, 15), 0)
	w.buttons["one"].Send(events.SlideStop, event)
	if g.Session.Layout().Windows[0].Root.Tabs[0] != "one" {
		t.Fatal("native gesture mutated domain before Drain")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if got := g.Session.Layout().Windows[0].Root.Tabs; got[0] != "two" || got[2] != "one" {
		t.Fatal("pointer reorder", got)
	}
	if err := w.Sync(w.host); err != nil {
		t.Fatal(err)
	}
	w.tabs["documents"].Geom.TotalBBox = image.Rect(0, 0, 600, 400)
	for i, id := range []string{"two", "three", "one"} {
		w.buttons[id].Geom.TotalBBox = image.Rect(i*100, 0, (i+1)*100, 30)
	}
	w.buttons["three"].Send(events.SlideStop, events.NewMouse(events.SlideStop, events.Left, image.Pt(590, 180), 0))
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	root := g.Session.Layout().Windows[0].Root
	if root.Axis != workbench.Horizontal || root.Second.Active != "three" {
		t.Fatal("pointer edge docking", root)
	}
}
func TestWorkbenchPointerCancelsInternalGapsAndFloatsOutsideScene(t *testing.T) {
	a, g := groupFixture(t)
	if err := g.Apply(func(s *workbench.Session) error {
		return s.Split("three", "main", "documents", "right", workbench.Horizontal, false)
	}); err != nil {
		t.Fatal(err)
	}
	w := g.windows["main"]
	mountGroupWindow(t, w)
	// SceneGeom is in render-window coordinates, while tab rectangles are in
	// the scene. The main window's scene starts at (0, 0).
	w.host.Scene.SceneGeom.SetRect(image.Rect(0, 0, 900, 700))
	w.tabs["documents"].Geom.TotalBBox = image.Rect(0, 0, 440, 640)
	w.tabs["right.tab"].Geom.TotalBBox = image.Rect(460, 0, 900, 640)
	before := mustEncodeSession(t, g.Session)
	for _, pos := range []image.Point{image.Pt(450, 250), image.Pt(300, 670)} {
		w.buttons["one"].Send(events.SlideStop, events.NewMouse(events.SlideStop, events.Left, pos, 0))
		if err := a.Drain(); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, mustEncodeSession(t, g.Session)) || len(a.Windows()) != 1 {
			t.Fatalf("internal splitter/gap at %v floated or changed the session", pos)
		}
	}
	// A captured release beyond the source scene still creates a floating
	// window, and the event may only publish that change through the mailbox.
	w.buttons["one"].Send(events.SlideStop, events.NewMouse(events.SlideStop, events.Left, image.Pt(920, 250), 0))
	if !bytes.Equal(before, mustEncodeSession(t, g.Session)) {
		t.Fatal("outside drop mutated domain before Drain")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	layout := g.Session.Layout()
	if len(layout.Windows) != 2 || len(a.Windows()) != 2 || layout.Focus.Window == "main" || layout.Focus.Panel != "one" {
		t.Fatal("outside drop did not float with focus transfer", layout)
	}
}

func TestWorkbenchPointerUsesLocalSnapshotAndRejectsStaleOrFullQueue(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	w.tabs["documents"].Geom.TotalBBox = image.Rect(0, 0, 600, 400)
	w.buttons["one"].Geom.TotalBBox = image.Rect(0, 0, 100, 30)
	hit, ok := w.hitDock("one", image.Pt(300, 100))
	if !ok || hit.group != "documents" {
		t.Fatal("snapshot target")
	}
	ran := false
	w.queue(func() error { ran = true; return nil })
	w.generation.Add(1)
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if ran {
		t.Fatal("retired native event delivered")
	}
	for i := 0; i < 256; i++ {
		if err := a.Post(func() {}); err != nil {
			t.Fatal(err)
		}
	}
	var reported error
	w.OnError = func(err error) { reported = err }
	w.queue(func() error { t.Fatal("refused callback delivered"); return nil })
	if w.pendingError.Load() == nil {
		t.Fatal("queue refusal lost")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := w.Sync(w.host); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(reported, app.ErrQueueFull) {
		t.Fatal("refusal report", reported)
	}
}

func TestWorkbenchPerWindowViewFactoriesAndGroupIdentities(t *testing.T) {
	a, g := groupFixture(t)
	created := map[string]int{}
	g.Panels[1].NewView = func(window string) func() ui.Node {
		created[window]++
		return func() ui.Node { return ui.Node{ID: "two", Kind: ui.Label, Text: window} }
	}
	main := g.windows["main"] // Existing window factories are frozen at registration.
	// Create a new floating registration with this per-window factory.
	if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal(err)
	}
	var floating *Workbench
	for id, w := range g.windows {
		if id != "main" {
			floating = w
		}
	}
	mountGroupWindow(t, floating)
	if err := floating.Sync(floating.host); err != nil {
		t.Fatal(err)
	}
	if created[floating.Window] != 1 {
		t.Fatal("factory not cached per native window", created)
	}
	other, err := NewWorkbenchGroup(a, g.Session, g.Panels)
	if err != nil {
		t.Fatal(err)
	}
	if err := other.Open(); err != nil {
		t.Fatal("distinct group IDs collided", err)
	}
	if other.windows["main"].Window == main.Window {
		t.Fatal("native lifetime IDs not group scoped")
	}
}
func TestWorkbenchFloatRollbackRestoresSoleSourceBeforeClosingAddedWindow(t *testing.T) {
	a, g := groupFixture(t)
	if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal(err)
	}
	before := mustEncodeSession(t, g.Session)
	var original string
	for id := range g.windows {
		if id != "main" {
			original = id
		}
	}
	old := g.windows[original].Window
	g.PersistencePath = filepath.Join(t.TempDir(), "missing", "session.json")
	if g.Float("two", workbench.Bounds{Width: 600, Height: 480}) == nil {
		t.Fatal("save failure ignored")
	}
	after := mustEncodeSession(t, g.Session)
	if !bytes.Equal(before, after) || a.Context().Err() != nil || len(a.Windows()) != 2 || g.windows[original].Window == old {
		t.Fatal("window replacement rollback incoherent")
	}
}
func TestWorkbenchLoadRebasesPersistedFocusRequest(t *testing.T) {
	_, g := groupFixture(t)
	path := filepath.Join(t.TempDir(), "session.json")
	if err := g.Save(path); err != nil {
		t.Fatal(err)
	}
	if err := g.Apply(func(s *workbench.Session) error { return s.Activate("main", "documents", "two") }); err != nil {
		t.Fatal(err)
	}
	before := g.Session.Layout().Focus.Serial
	if err := g.Load(path); err != nil {
		t.Fatal(err)
	}
	if f := g.Session.Layout().Focus; f.Panel != "one" || f.Serial != before+1 {
		t.Fatal("Load reused an old focus serial", f)
	}
}
func TestWorkbenchPendingNativeGeometryCannotOverwriteNewRestore(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	w.queueWindowBounds(workbench.Bounds{X: 40, Y: 40, Width: 800, Height: 500, Positioned: true})
	w.flushWindowBounds()
	desired := workbench.Bounds{X: 160, Y: 120, Width: 1000, Height: 700, Positioned: true}
	if err := g.Apply(func(s *workbench.Session) error { return s.SetBounds("main", desired) }); err != nil {
		t.Fatal(err)
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	win, _ := g.Session.Window("main")
	if win.Bounds != desired {
		t.Fatal("stale native geometry overwrote restored bounds", win.Bounds)
	}
}

func TestWorkbenchControllerCloseDoesNotReenterNativeCloseHook(t *testing.T) {
	for _, failClose := range []bool{false, true} {
		t.Run(fmt.Sprintf("failure=%t", failClose), func(t *testing.T) {
			a, g := groupFixture(t)
			if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
				t.Fatal(err)
			}
			var logical string
			for id := range g.windows {
				if id != "main" {
					logical = id
				}
			}
			floating := g.windows[logical]
			before := mustEncodeSession(t, g.Session)
			calls := 0
			closeErr := errors.New("injected synchronous close failure")
			g.close = func(id string) error {
				calls++
				if calls > 1 {
					t.Fatal("close hook reentered the controller")
				}
				if err := floating.WindowClosed(); err != nil {
					return err
				}
				if failClose {
					return closeErr
				}
				return a.CloseWindow(id)
			}
			err := g.Close(logical)
			if failClose {
				if !errors.Is(err, closeErr) || g.windows[logical] != floating || len(a.Windows()) != 2 || !bytes.Equal(before, mustEncodeSession(t, g.Session)) {
					t.Fatal("failed synchronous close did not restore registration", err)
				}
			} else if err != nil || len(a.Windows()) != 1 || len(g.Session.Layout().Windows) != 1 || g.Session.Layout().Focus.Panel != "two" {
				t.Fatal("synchronous controller close failed", err)
			}
			if calls != 1 {
				t.Fatalf("close calls: %d", calls)
			}
		})
	}
}

func TestWorkbenchQueuedErrorWithoutHandlerReachesSync(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	failure := errors.New("injected queued transaction failure")
	before := mustEncodeSession(t, g.Session)
	retained := w.records["one"]
	children := len(w.host.Children)
	w.queue(func() error { return failure })
	w.queue(func() error { return errors.New("later queued failure") })
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	// A subsequent delivery refusal must not replace the first accepted
	// callback's failure either. Keep the pending slot bounded across bursts.
	for i := 0; i < 256; i++ {
		if err := a.Post(func() {}); err != nil {
			t.Fatal(err)
		}
	}
	w.queue(func() error { t.Fatal("refused callback delivered"); return nil })
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if w.pendingError.Load() == nil {
		t.Fatal("unhandled queued failure dropped")
	}
	if err := w.Sync(w.host); !errors.Is(err, failure) {
		t.Fatal("unhandled failure not surfaced", err)
	}
	if !bytes.Equal(before, mustEncodeSession(t, g.Session)) {
		t.Fatal("failed callback changed layout")
	}
	if w.records["one"] != retained || len(w.host.Children) != children {
		t.Fatal("queued failure destroyed retained native resources")
	}
	if w.pendingError.Load() != nil {
		t.Fatal("surfaced failure retained")
	}
}

func TestWorkbenchGroupHandlerAssignedAfterOpenReceivesQueuedFailure(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	failure := errors.New("injected queued transaction failure")
	var reported error
	g.OnError = func(err error) { reported = err }
	w.queue(func() error { return failure })
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(reported, failure) || w.pendingError.Load() != nil {
		t.Fatal("late group handler missed error", reported)
	}
}

// Embedding the native interface makes unexpected geometry operations panic;
// these overrides model only the synchronous bounds requests under test.
type workbenchGeometryWindow struct {
	system.Window
	size, pos             image.Point
	sizeWrites, posWrites int
}

func (w *workbenchGeometryWindow) WinSize() image.Point                { return w.size }
func (w *workbenchGeometryWindow) Position(*system.Screen) image.Point { return w.pos }
func (w *workbenchGeometryWindow) SetWinSize(size image.Point)         { w.size = size; w.sizeWrites++ }
func (w *workbenchGeometryWindow) SetPos(pos image.Point, _ *system.Screen) {
	w.pos = pos
	w.posWrites++
}
func (w *workbenchGeometryWindow) ConstrainFrame(bool) sides.Sides[int] { return sides.Sides[int]{} }

func TestWorkbenchRefreshPreservesNativeResizeAwaitingOwnerCapture(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	native := &workbenchGeometryWindow{}
	w.syncWindowBounds(native)
	if native.size != image.Pt(900, 640) || native.sizeWrites != 1 {
		t.Fatal("initial bounds not applied", native)
	}

	// A native resize/move is visible before its queued owner model update.
	captured := workbench.Bounds{X: 80, Y: 60, Width: 1100, Height: 700, Positioned: true}
	native.size, native.pos = image.Pt(captured.Width, captured.Height), image.Pt(captured.X, captured.Y)
	w.queueWindowBounds(captured)
	w.flushWindowBounds()
	if err := w.Sync(w.host); err != nil {
		t.Fatal(err)
	}
	w.syncWindowBounds(native)
	if native.sizeWrites != 1 || native.posWrites != 0 || native.size != image.Pt(1100, 700) || native.pos != image.Pt(80, 60) {
		t.Fatal("unrelated refresh reverted in-progress native geometry", native)
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := w.Sync(w.host); err != nil {
		t.Fatal(err)
	}
	w.syncWindowBounds(native)
	if native.sizeWrites != 1 || native.posWrites != 0 {
		t.Fatal("accepted native capture caused a geometry request", native)
	}
	window, ok := g.Session.Window("main")
	if !ok || window.Bounds != captured {
		t.Fatal("native capture not accepted", window.Bounds)
	}

	// Explicit application restore still pushes changed model bounds.
	restored := workbench.Bounds{X: 120, Y: 100, Width: 800, Height: 500, Positioned: true}
	if err := g.Apply(func(s *workbench.Session) error { return s.SetBounds("main", restored) }); err != nil {
		t.Fatal(err)
	}
	if err := w.Sync(w.host); err != nil {
		t.Fatal(err)
	}
	w.syncWindowBounds(native)
	if native.size != image.Pt(800, 500) || native.pos != image.Pt(120, 100) || native.sizeWrites != 2 || native.posWrites != 1 {
		t.Fatal("explicit restore not applied", native)
	}
}
func TestWorkbenchNativeGeometryBurstAcceptsLatestSnapshot(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	first := workbench.Bounds{X: 20, Y: 20, Width: 950, Height: 650, Positioned: true}
	latest := workbench.Bounds{X: 80, Y: 60, Width: 1100, Height: 700, Positioned: true}
	w.queueWindowBounds(first)
	w.queueWindowBounds(latest)
	w.flushWindowBounds()
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	window, ok := g.Session.Window("main")
	if !ok || window.Bounds != latest {
		t.Fatal("burst lost final native geometry", window.Bounds)
	}
}

func TestWorkbenchRefreshBetweenNativeCapturesDoesNotPushOlderSnapshot(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	native := &workbenchGeometryWindow{}
	w.syncWindowBounds(native)
	first := workbench.Bounds{X: 20, Y: 20, Width: 950, Height: 650, Positioned: true}
	native.size, native.pos = image.Pt(first.Width, first.Height), image.Pt(first.X, first.Y)
	w.queueWindowBounds(first)
	w.flushWindowBounds()
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	latest := workbench.Bounds{X: 80, Y: 60, Width: 1100, Height: 700, Positioned: true}
	native.size, native.pos = image.Pt(latest.Width, latest.Height), image.Pt(latest.X, latest.Y)
	w.queueWindowBounds(latest)
	w.flushWindowBounds()
	if err := w.Sync(w.host); err != nil {
		t.Fatal(err)
	}
	w.syncWindowBounds(native)
	if native.size != image.Pt(1100, 700) || native.pos != image.Pt(80, 60) || native.sizeWrites != 1 || native.posWrites != 0 {
		t.Fatal("refresh pushed older native capture", native)
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkbenchLoadUnchangedBoundsOverridesPendingNativeCapture(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	native := &workbenchGeometryWindow{}
	w.syncWindowBounds(native)
	path := filepath.Join(t.TempDir(), "session.json")
	if err := g.Save(path); err != nil {
		t.Fatal(err)
	}
	captured := workbench.Bounds{X: 80, Y: 60, Width: 1100, Height: 700, Positioned: true}
	native.size, native.pos = image.Pt(captured.Width, captured.Height), image.Pt(captured.X, captured.Y)
	w.queueWindowBounds(captured)
	w.flushWindowBounds()
	if err := g.Load(path); err != nil {
		t.Fatal(err)
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := w.Sync(w.host); err != nil {
		t.Fatal(err)
	}
	w.syncWindowBounds(native)
	if native.size != image.Pt(900, 640) || native.sizeWrites != 2 {
		t.Fatal("explicit unchanged-bounds Load lost to native capture", native)
	}
}

func TestWorkbenchNativeGeometryFiltersTransientAndRetiredCaptures(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	before := mustEncodeSession(t, g.Session)
	for _, b := range []workbench.Bounds{{Width: 0, Height: 0}, {Width: 100, Height: 80}, {X: -32000, Y: -32000, Width: 900, Height: 640, Positioned: true}, {X: 1000001, Width: 900, Height: 640}} {
		w.queueWindowBounds(b)
		if err := w.captureWindowBounds(b); err != nil {
			t.Fatal(err)
		}
	}
	if w.pendingBounds.Load() != nil || !bytes.Equal(before, mustEncodeSession(t, g.Session)) {
		t.Fatal("transient native geometry changed model")
	}
	if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal(err)
	}
	var floating *Workbench
	for id, candidate := range g.windows {
		if id != "main" {
			floating = candidate
		}
	}
	floating.queueWindowBounds(workbench.Bounds{Width: 600, Height: 480})
	if err := g.Close(floating.SessionWindow); err != nil {
		t.Fatal(err)
	}
	floating.flushWindowBounds()
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if floating.pendingError.Load() != nil || len(g.Session.Layout().Windows) != 1 {
		t.Fatal("retired timer changed session or reported cancellation")
	}
}
func TestWorkbenchGeometryBurstUsesOneLatestPendingSnapshot(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	for i := 0; i < 80; i++ {
		w.queueWindowBounds(workbench.Bounds{Width: 900 + i, Height: 640})
	}
	if w.pendingBounds.Load().bounds.Width != 979 {
		t.Fatal("pending snapshot not latest")
	}
	if g.Session.Layout().Windows[0].Bounds.Width != 900 {
		t.Fatal("native producer mutated model")
	}
	w.flushWindowBounds()
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if g.Session.Layout().Windows[0].Bounds.Width != 979 || w.pendingBounds.Load() != nil {
		t.Fatal("coalesced latest bounds not accepted")
	}
}
func TestWorkbenchCommittedTransactionSurvivesRefusedWake(t *testing.T) {
	a, g := groupFixture(t)
	g.open = func(window app.Window) error {
		if err := a.Open(window); err != nil {
			return err
		}
		a.Close()
		return nil
	}
	if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
		t.Fatal("committed transaction reported wake failure", err)
	}
	if len(g.Session.Layout().Windows) != 2 {
		t.Fatal("accepted layout not committed")
	}
}

func TestWorkbenchNativeCaptureAdmissionRejectsOutOfOrderCallbacks(t *testing.T) {
	a, g := groupFixture(t)
	w := g.windows["main"]
	mountGroupWindow(t, w)
	older := workbench.Bounds{Width: 950, Height: 650}
	latest := workbench.Bounds{Width: 1100, Height: 700}
	w.queueWindowBounds(older)
	w.boundsQueued.Store(false)
	first := w.pendingBounds.Swap(nil)
	w.queueWindowBounds(latest)
	w.boundsQueued.Store(false)
	second := w.pendingBounds.Swap(nil)
	if first == nil || second == nil {
		t.Fatal("capture missing")
	}
	// Simulate timer one pausing after take, while timer two posts first.
	w.postWindowBounds(second)
	w.postWindowBounds(first)
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	window, ok := g.Session.Window("main")
	if !ok || window.Bounds != latest {
		t.Fatal("older timer callback overwrote latest bounds", window.Bounds)
	}
}

type workbenchCloseProbe struct {
	calls   *int
	failure error
}

func (p *workbenchCloseProbe) WindowClosed() error { *p.calls++; return p.failure }
func TestWorkbenchCloseTraversalNotifiesEveryExtensionAndJoinsErrors(t *testing.T) {
	firstErr, secondErr := errors.New("first close failure"), errors.New("nested close failure")
	calls := [4]int{}
	probe := func(index int, failure error) *record {
		return &record{node: ui.Node{Kind: ui.Extension, Custom: &workbenchCloseProbe{calls: &calls[index], failure: failure}}}
	}
	root := probe(0, firstErr)
	firstChild := probe(1, nil)
	firstChild.children = []*record{probe(2, secondErr)}
	root.children = []*record{firstChild, probe(3, nil)}
	err := closeWorkbenchExtensions(root)
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatal("close failures not joined", err)
	}
	if calls != [4]int{1, 1, 1, 1} {
		t.Fatal("not all close hooks visited", calls)
	}
}

func TestWorkbenchQueuedCallbackRefusalAfterRetirementIsHarmless(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		t.Run(fmt.Sprintf("shutdown=%t", shutdown), func(t *testing.T) {
			a, g := groupFixture(t)
			w := g.windows["main"]
			if !shutdown {
				if err := g.Float("two", workbench.Bounds{Width: 500, Height: 400}); err != nil {
					t.Fatal(err)
				}
				for id, candidate := range g.windows {
					if id != "main" {
						w = candidate
					}
				}
			}
			mountGroupWindow(t, w)
			if shutdown {
				a.Close()
			} else if err := g.Close(w.SessionWindow); err != nil {
				t.Fatal(err)
			}
			ran := false
			w.queue(func() error { ran = true; return nil })
			if w.pendingError.Load() != nil {
				t.Fatal("expected retirement refusal became fatal", *w.pendingError.Load())
			}
			if !shutdown {
				if err := a.Drain(); err != nil {
					t.Fatal(err)
				}
			}
			if ran {
				t.Fatal("retired callback delivered")
			}
		})
	}
}
