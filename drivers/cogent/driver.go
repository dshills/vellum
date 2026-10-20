// Package cogent implements Vellum's native driver using pinned Cogent Core.
// It is experimental; one process main loop owns multiple desktop windows.
package cogent

import (
	"context"
	"errors"
	"fmt"
	"html"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

// Extension is implemented by optional adapter widgets (workbench, canvas, etc.).
// Sync runs under the UI owner and preserves the native widget's resources.
type Extension interface {
	Mount(tree.Node) core.Widget
	Sync(core.Widget) error
}

// ThemeExtension receives a private application theme snapshot for managed subtrees.
type ThemeExtension interface{ SetTheme(ui.Theme) }

type Driver struct {
	// DesktopSession opts into bounded native integration routing and frame barriers.
	// It cannot be combined with FrameMetrics. Ordinary applications leave it nil.
	DesktopSession *DesktopSession
	Commands       *command.Registry
	OnError        func(error)
	Theme          ui.Theme
	// ContinuousRendering opts into refresh-rate paint ticks; default uses invalidation.
	ContinuousRendering bool
	// ExperimentalDisplayBackpressure is a Mac-only display credit trial. Zero
	// keeps default rendering. Missing/zero callbacks refuse and close the run.
	ExperimentalDisplayBackpressure time.Duration
	// ExperimentalGridScrollReuse enables checked CPU pixel reuse in main-window
	// grid scenes. Default replay is retained: current table workloads are slower
	// with this experiment. It does not reduce GPU uploads. Copied at Run startup.
	ExperimentalGridScrollReuse bool
	// GPUSceneRendering enables ordered compute paint on main-window scenes.
	// Unsupported paint uses the current CPU image. This candidate is opt-in
	// while representative native performance and lifecycle checks are ongoing.
	GPUSceneRendering bool
	// FrameMetrics opts into bounded diagnostic samples for explicitly tagged
	// routed grid/tree input; ordinary frames do not pay for timing calls.
	FrameMetrics *FrameMetrics
	// PresentationMetrics opts into bounded native display feedback on macOS.
	PresentationMetrics *PresentationMetrics
}
type record struct {
	scope    app.WindowRef
	widget   core.Widget
	node     ui.Node
	children []*record
	theme    ui.Theme
	control  *controlState
}

var processRun atomic.Bool

type nativeWindow struct {
	window           app.Window
	body             *core.Body
	host             *core.Frame
	root             *record
	rootMu           sync.Mutex
	lastFocus        ui.FocusRequest
	axScope          string // originating event loop, under rootMu
	stopLoss         func()
	lossMu           sync.Mutex
	input            atomic.Pointer[textWindow]
	visibility       atomic.Pointer[windowVisibility]
	metrics          *frameWindowMetrics
	presentation     *presentationSink
	desktop          *DesktopSession
	source           atomic.Pointer[events.Source]
	pending, closing atomic.Bool
}

func (d Driver) Run(ctx context.Context, a *app.Application) error {
	if d.DesktopSession != nil && d.FrameMetrics != nil {
		return ErrDesktopSession
	}
	if d.PresentationMetrics != nil && !d.PresentationMetrics.valid() {
		return ErrFrameMetricsConfig
	}
	if d.FrameMetrics != nil && !d.FrameMetrics.valid() {
		return ErrFrameMetricsConfig
	}
	if err := d.Theme.Validate(); err != nil {
		return err
	}
	d.Theme = d.Theme.Clone()
	if len(a.Windows()) == 0 {
		return fmt.Errorf("cogent: at least one registered window required")
	}
	for _, spec := range a.Windows() {
		if err := validateWindowChromeSupport(spec); err != nil {
			return err
		}
	}
	if !processRun.CompareAndSwap(false, true) {
		return fmt.Errorf("cogent: native main loop runs once per process")
	}
	var started, shown atomic.Bool
	defer func() {
		if !started.Load() {
			processRun.Store(false)
		}
	}()
	if d.DesktopSession != nil {
		if err := d.DesktopSession.start(a); err != nil {
			return err
		}
		defer d.DesktopSession.close()
	}
	if err := configureAccessibility(a); err != nil {
		return err
	}
	defer releaseAccessibilityConfiguration()
	var failure atomic.Pointer[error]
	var mu sync.Mutex
	windows := map[string]*nativeWindow{}
	snapshot := func() []*nativeWindow {
		mu.Lock()
		defer mu.Unlock()
		out := make([]*nativeWindow, 0, len(windows))
		for _, w := range windows {
			out = append(out, w)
		}
		return out
	}
	services := &platformServices{owner: a, commands: d.Commands, onError: d.OnError}
	removeServiceCleanup, err := configureServiceCleanup(services)
	if err != nil {
		return err
	}
	defer removeServiceCleanup()
	var quitting atomic.Bool
	quit := func() {
		if !shown.Load() {
			return
		}
		if quitting.CompareAndSwap(false, true) {
			go func() { services.shutdown(); core.TheApp.QuitReq() }()
		}
	}
	defer func() {
		if started.Load() {
			services.shutdown()
		}
	}()
	fail := func(err error) {
		if failure.CompareAndSwap(nil, &err) {
			a.Close()
			quit()
		}
	}
	stopDisplayExperiment, err := configureDisplayBackpressure(d.ExperimentalDisplayBackpressure, fail)
	if err != nil {
		return err
	}
	defer stopDisplayExperiment()
	failWindow := func(w *nativeWindow, err error) {
		if w.window.Ref().Context().Err() != nil && (errors.Is(err, app.ErrWindowClosed) || errors.Is(err, app.ErrClosed)) {
			return
		}
		fail(err)
	}
	wake := func() {
		for _, w := range snapshot() {
			if s := w.source.Load(); s != nil && !w.closing.Load() && w.pending.CompareAndSwap(false, true) {
				s.Custom("vellum.wake")
			}
		}
	}
	var custom func(*nativeWindow, events.Event)
	refresh := func(w *nativeWindow) {
		if w.closing.Load() {
			return
		}
		withWindowOwner(a, w.window.Ref(), func() {
			w.rootMu.Lock()
			defer w.rootMu.Unlock()
			if w.closing.Load() {
				return
			}
			if !a.TakeDirty(w.window.Ref()) {
				return
			}
			d.DesktopSession.invalidate(w)
			desktopTag, desktopStart := d.DesktopSession.reconcileStart(w)
			defer d.DesktopSession.reconcileEnd(w, desktopTag, desktopStart)
			metricTag, metricStart := w.metrics.reconcileStart()
			defer w.metrics.reconcileEnd(metricTag, metricStart)
			n := w.window.View()
			if w.window.Ref().Context().Err() != nil {
				return
			}
			if err := ui.Validate(n); err != nil {
				failWindow(w, err)
				return
			}
			var err error
			w.root, err = reconcileScoped(w.host, w.root, n, a, w.window.Ref(), d.Theme)
			if err != nil {
				failWindow(w, err)
				return
			}
			if w.root == nil {
				fail(fmt.Errorf("cogent: reconciliation returned no root"))
				return
			}
			nativeRefreshServices(w.body)
			bindControlWake(w.root, func(e events.Event) { custom(w, e) }, d.OnError)
			bindGridScopes(w.root)
			bindTreeScopes(w.root)
			flushDialogs(w.root)
			bindWorkbenchInputScopes(w.root)
			if err := applyFocus(w.root, n.Focus, &w.lastFocus); err != nil {
				failWindow(w, err)
			}
			if err := bindTextWindow(w); err != nil {
				failWindow(w, err)
				return
			}
			syncAccessibility(w, a)
		})
	}
	var topologyMu sync.Mutex
	var syncTopology func()
	syncTopology = func() {
		// No other window's native render lock is acquired while holding the owner.
		// Native CloseReq runs asynchronously so Close events can finish independently.
		topologyMu.Lock()
		defer topologyMu.Unlock()
		var launch []*nativeWindow
		// If construction or refresh fails, discard every tree that never entered
		// a native loop. Close callbacks cannot clean up unlaunched bodies.
		defer func() {
			for _, w := range launch {
				w.closing.Store(true)
				cleanup := func() {
					w.rootMu.Lock()
					defer w.rootMu.Unlock()
					stopDeviceLoss(w)
					closeTextWindow(w)
					closeAccessibility(w, "driver-unlaunched-cleanup")
					w.metrics.retire()
					w.presentation.Closed()
					w.body.Scene.SetOriginCustomHandler(nil)
					d.DesktopSession.retire(w)
					nativeUnmountServices(w.body)
					disposeControlTree(w.root)
					// NewBody assigns a Scene parent before adding the body to
					// its children, so Delete cannot find an unlaunched body.
					scene := w.body.Scene
					w.body.Destroy()
					scene.Destroy()
				}
				if err := a.Do(cleanup); err != nil {
					cleanup()
				}
				mu.Lock()
				if windows[w.window.ID] == w {
					delete(windows, w.window.ID)
				}
				mu.Unlock()
			}
		}()
		desired := a.Windows()
		active := map[string]app.WindowRef{}
		for _, w := range desired {
			active[w.ID] = w.Ref()
		}
		for _, w := range snapshot() {
			if active[w.window.ID] != w.window.Ref() && w.closing.CompareAndSwap(false, true) {
				if w.body.Scene.RenderWindow() != nil {
					win := w.body.Scene.RenderWindow().SystemWindow
					go win.CloseReq()
				}
			}
		}
		for _, spec := range desired {
			mu.Lock()
			_, exists := windows[spec.ID]
			mu.Unlock()
			if exists || ctx.Err() != nil || spec.Ref().Context().Err() != nil {
				continue
			}
			if err := validateWindowChromeSupport(spec); err != nil {
				fail(err)
				return
			}
			w := &nativeWindow{window: spec}
			buildErr := a.Do(func() {
				w.body = core.NewBody(spec.Title)
				if d.PresentationMetrics != nil {
					var err error
					w.presentation, err = d.PresentationMetrics.bind(spec.ID)
					if err != nil {
						failWindow(w, err)
						return
					}
				}
				w.body.Scene.GPUSceneRendering = d.GPUSceneRendering
				setGridScrollReuse(w.body.Scene, d.ExperimentalGridScrollReuse)
				if d.FrameMetrics != nil {
					var metricErr error
					w.metrics, metricErr = d.FrameMetrics.bind(a, w)
					if metricErr != nil {
						failWindow(w, metricErr)
						return
					}
				}
				if err := d.DesktopSession.bind(w); err != nil {
					failWindow(w, err)
					return
				}
				bindWindowOriginCustom(w, d.DesktopSession, custom)
				nativeMountServices(w.body, services)
				configureInitialWindow(w.body)
				configureWindowChromeContent(w.body, spec)
				w.host = newAccessibilityHost(w)
				configureRootScrolling(w.body, w.host, spec)
				bindWindowScope(w.host, a, spec.Ref())
				w.host.Styler(func(s *styles.Style) { s.Grow.Set(1, 1); s.Direction = styles.Column; s.Gap.Zero() })
				bindAccessibilityKeyboard(w, a, d.OnError)
			})
			if buildErr != nil {
				return
			}
			mu.Lock()
			windows[spec.ID] = w
			mu.Unlock()
			launch = append(launch, w)
			w.body.OnShow(func(events.Event) {
				rw := w.body.Scene.RenderWindow()
				if rw == nil || rw.SystemWindow == nil {
					if !w.closing.Load() && w.window.Ref().Context().Err() == nil {
						fail(fmt.Errorf("cogent: render window unavailable on show"))
					}
					return
				}
				shown.Store(true)
				if w.closing.Load() {
					go rw.SystemWindow.CloseReq()
					return
				}
				if err := attachPresentation(w, rw.SystemWindow); err != nil {
					failWindow(w, err)
					return
				}
				if err := configureRendering(rw.SystemWindow, d.ContinuousRendering); err != nil {
					failWindow(w, err)
					return
				}
				watchDeviceLoss(w, func(err error) { failWindow(w, err) })
				lifetime := spec.Ref().Context()
				if w.closing.Load() || lifetime.Err() != nil {
					w.closing.Store(true)
					if native := w.body.Scene.RenderWindow(); native != nil {
						go native.SystemWindow.CloseReq()
					}
					return
				}
				w.source.Store(w.body.SystemEvents())
				if err := bindWindowVisibility(w); err != nil {
					failWindow(w, err)
					return
				}
				if err := withWindowOwner(a, w.window.Ref(), func() {
					w.rootMu.Lock()
					defer w.rootMu.Unlock()
					if err := startTextWindow(w, a); err != nil {
						failWindow(w, err)
					}
				}); err != nil {
					failWindow(w, err)
					return
				}
				if err := withWindowOwner(a, w.window.Ref(), func() {
					w.rootMu.Lock()
					defer w.rootMu.Unlock()
					flushDialogs(w.root)
					bindWorkbenchInputScopes(w.root)
					if err := bindTextWindow(w); err != nil {
						failWindow(w, err)
						return
					}
					syncAccessibility(w, a)
				}); err != nil {
					failWindow(w, err)
					return
				}
				if w.window.Ref().Context().Err() != nil {
					return
				}
				if ctx.Err() != nil || failure.Load() != nil {
					quit()
					return
				}
				a.SetServices(services)
				wake()
			})
			w.body.Scene.OnFirst(events.Custom, func(e events.Event) { custom(w, e) })
			if d.Commands != nil {
				bindCommandShortcutsScoped(w.body.Scene, a, spec.Ref(), d.Commands, d.OnError)
			}
			bindCloseRequest(w, a, d.OnError)
			w.body.OnClose(func(events.Event) {
				w.closing.Store(true)
				// CloseWindow changes registration and wakes topology reconciliation;
				// native CloseReq runs outside the application owner. If the owner
				// already stopped, skip model notifications; native/control cleanup
				// below still releases every mounted widget.
				if err := withWindowOwner(a, w.window.Ref(), func() {
					w.rootMu.Lock()
					defer w.rootMu.Unlock()
					if spec.Ref().Context().Err() == nil {
						if err := closeWorkbenchExtensions(w.root); err != nil {
							failWindow(w, err)
						}
					}
				}); err != nil && err != app.ErrClosed {
					failWindow(w, err)
				}
				stopDeviceLoss(w)
				a.CloseWindow(spec.Ref())
				cleanup := func() {
					w.rootMu.Lock()
					defer w.rootMu.Unlock()
					closeTextWindow(w)
					closeAccessibility(w, "driver-window-cleanup")
					w.metrics.retire()
					w.presentation.Closed()
					w.body.Scene.SetOriginCustomHandler(nil)
					d.DesktopSession.retire(w)
					nativeUnmountServices(w.body)
					disposeControlTree(w.root)
				}
				if err := a.Do(cleanup); err != nil {
					cleanup()
				}
				w.source.Store(nil)
				mu.Lock()
				if windows[spec.ID] == w {
					delete(windows, spec.ID)
				}
				mu.Unlock()
				// A new registration may have been accepted just before the last native
				// window closes, before its queued Custom wake could create that window.
				// Launch remaining registrations before returning from Close, keeping
				// Cogent's window wait group alive without requiring a shown event source.
				if ctx.Err() == nil {
					syncTopology()
				}
			})
			refresh(w)
			if failure.Load() != nil {
				return
			}
		}
		for len(launch) > 0 {
			w := launch[0]
			launch = launch[1:]
			started.Store(true)
			initialWindowStage(w.body, w.window).Run()
		}
	}
	custom = func(w *nativeWindow, e events.Event) {
		w.pending.Store(false)
		metricTag, metricStart := w.metrics.ownerStart()
		func() {
			defer w.metrics.ownerEnd(metricTag, metricStart)
			desktopTag, desktopStart := d.DesktopSession.ownerStart(w)
			handleTextEvent(w, e)
			drainErr := a.Drain()
			d.DesktopSession.ownerEnd(w, desktopTag, desktopStart, drainErr)
			syncTopology()
		}()
		refresh(w)
		if d.DesktopSession != nil && d.DesktopSession.handle(w, e) {
			return
		}
		if isWindowWake(e) {
			e.SetHandled()
		}
	}
	a.SetWake(wake)
	defer a.SetWake(nil)
	defer a.SetServices(nil)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	syncTopology()
	if err := failure.Load(); err != nil && !started.Load() {
		return *err
	}
	if ctx.Err() != nil && !started.Load() {
		return ctx.Err()
	}
	if !started.Load() {
		return fmt.Errorf("cogent: no native window started")
	}
	stop := context.AfterFunc(ctx, quit)
	defer stop()
	core.Wait()
	var runErr error
	if err := failure.Load(); err != nil {
		runErr = *err
	}
	if err := services.cleanupErr.Load(); err != nil {
		return errors.Join(runErr, *err)
	}
	return runErr
}
func sameShape(r *record, n ui.Node) bool {
	if r.node.ID != n.ID || r.node.Kind != n.Kind || len(r.children) != len(n.Children) {
		return false
	}
	if n.Kind == ui.Extension {
		if reflect.TypeOf(r.node.Custom) != reflect.TypeOf(n.Custom) {
			return false
		}
		a, b := reflect.ValueOf(r.node.Custom), reflect.ValueOf(n.Custom)
		if a.IsValid() && a.Kind() == reflect.Pointer && a.Pointer() != b.Pointer() {
			return false
		}
	}
	for i, c := range n.Children {
		if !sameShape(r.children[i], c) {
			return false
		}
	}
	return true
}
func mountScoped(parent tree.Node, n ui.Node, a *app.Application, scope app.WindowRef, themes ...ui.Theme) (*record, error) {
	r, err := mountShallowScoped(parent, n, a, scope, themes...)
	if err != nil {
		return nil, err
	}
	for _, c := range n.Children {
		child, err := mountScoped(childHost(r, c.ID), c, a, scope, r.theme)
		if err != nil {
			disposeControlTree(r)
			r.widget.AsTree().Delete()
			return nil, err
		}
		r.children = append(r.children, child)
	}
	return r, nil
}

func mountShallowScoped(parent tree.Node, n ui.Node, a *app.Application, scope app.WindowRef, themes ...ui.Theme) (*record, error) {
	var theme ui.Theme
	if len(themes) > 0 {
		theme = themes[0]
	}
	if _, err := ui.ResolveStyle(n, theme); err != nil {
		return nil, err
	}
	if _, err := scopedWindowContext(a, scope); err != nil {
		return nil, err
	}
	r := &record{node: n, theme: theme, scope: scope}
	if ext, ok := n.Custom.(ThemeExtension); n.Kind == ui.Extension && ok {
		ext.SetTheme(theme.Clone())
	}
	switch n.Kind {
	case ui.Column, ui.Row:
		r.widget = core.NewFrame(parent)
	case ui.Label:
		r.widget = core.NewText(parent)
	case ui.Extension:
		if ext, ok := n.Custom.(Extension); ok {
			r.widget = ext.Mount(parent)
		} else {
			return nil, fmt.Errorf("cogent: unsupported extension %q", n.ID)
		}
	default:
		var ok bool
		var err error
		r.widget, ok, err = mountControl(parent, r, a, scope.ID())
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("cogent: unsupported kind %q", n.Kind)
		}
	}
	if r.widget == nil || (reflect.ValueOf(r.widget).Kind() == reflect.Pointer && reflect.ValueOf(r.widget).IsNil()) {
		if scope.Context().Err() != nil {
			return nil, app.ErrWindowClosed
		}
		return nil, fmt.Errorf("cogent: extension %q returned nil widget", n.ID)
	}
	bindWindowScope(r.widget, a, scope)
	r.widget.AsWidget().SetName(n.ID)
	syncSemantics(r)
	bindFileDrop(r, a, scope.ID())
	r.widget.AsWidget().Styler(func(s *styles.Style) {
		applyStyle(s, r.node, r.theme)
		if r.node.Kind == ui.Label {
			value, err := ui.ResolveStyle(r.node, r.theme)
			r.widget.(*core.Text).AllocationLayout = err == nil && (value.TextAlign.Set || value.GrowX.Set || value.GrowY.Set)
		}
	})
	return r, nil
}
func syncRecord(r *record, n ui.Node) error {
	visible, scopes := tabCloseInheritedContext(r)
	return syncRecordVisible(r, n, visible && accessibilityDialogVisible(r), scopes)
}

func syncRecordVisible(r *record, n ui.Node, visible bool, scopes []*tabCloseScope) error {
	for _, scope := range scopes {
		if scope == nil || scope.closed.Load() {
			visible = false
			break
		}
	}
	if _, err := ui.ResolveStyle(n, r.theme); err != nil {
		return err
	}
	if err := ui.ValidateControl(n); err != nil {
		return err
	}
	previous := r.node
	r.node = n
	disabledChanged := r.widget.AsWidget().IsDisabled() != controlDisabled(r)
	if previous.Style != n.Style || previous.Grow != n.Grow || previous.MinWidth != n.MinWidth || previous.MinHeight != n.MinHeight {
		if r.control != nil && r.control.grid != nil {
			r.control.grid.layoutPending.Store(true)
			r.control.grid.layout.valid = false
		}
		r.widget.AsWidget().Update()
	}
	switch n.Kind {
	case ui.Label:
		if previous.Text != n.Text || r.widget.(*core.Text).Text != html.EscapeString(n.Text) {
			r.widget.(*core.Text).SetText(html.EscapeString(n.Text)).Update()
		}
	case ui.Extension:
		r.widget.AsTree().SetProperty(tabCloseContextProperty, &tabCloseContext{visible: visible, scopes: append([]*tabCloseScope(nil), scopes...)})
		if dock, ok := n.Custom.(*Workbench); ok && dock != nil {
			dock.setInputScope(visible, controlDisabled(r), scopes)
		}
		ext, ok := n.Custom.(Extension)
		if !ok {
			return fmt.Errorf("cogent: unsupported extension %q", n.ID)
		}
		if err := ext.Sync(r.widget); err != nil {
			return err
		}
	}
	syncSemantics(r)
	if err := syncControl(r, previous); err != nil {
		return err
	}
	// An ancestor update may already have styled this widget using its old
	// inherited Disabled flag. Refresh after its own state and control sync,
	// even when its descriptor is unchanged, before syncing descendants.
	if disabledChanged {
		r.widget.AsWidget().Update()
	}
	if n.Kind == ui.Dialog && r.control != nil && r.control.dialog != nil {
		d := r.control.dialog
		scope := d.tabScope.Load()
		visible = visible && n.Open && !d.dismissed && scope != nil && !scope.closed.Load()
		scopes = append(append([]*tabCloseScope(nil), scopes...), scope)
	}
	syncTabVisibility(r, visible, scopes)
	for i, c := range n.Children {
		childVisible := visible
		switch n.Kind {
		case ui.Tabs:
			childVisible = childVisible && c.ID == n.SelectedID
		case ui.Disclosure:
			childVisible = childVisible && n.Checked
		case ui.Dialog:
			childVisible = childVisible && n.Open
		}
		if err := syncRecordVisible(r.children[i], c, childVisible, scopes); err != nil {
			return err
		}
	}
	return nil
}

func drainCustom(a *app.Application, pending *atomic.Bool, refresh func(), e events.Event) {
	pending.Store(false)
	a.Drain()
	refresh()
	if isWindowWake(e) {
		e.SetHandled()
	}
}

// Initial dimensions belong to creation options, not persistent content minima.
func configureInitialWindow(body *core.Body) {
	body.Styler(func(s *styles.Style) { s.Grow.Set(1, 1) })
}
func initialWindowStage(body *core.Body, spec app.Window) *core.Stage {
	return initialWindowChrome(body.NewWindow().SetUseMinSize(false).SetResizable(!spec.FixedSize).SetInitialSizeDP(spec.Width, spec.Height), spec)
}
