//go:build offscreen

package cogent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io/fs"
	"math"
	"slices"
	"sort"
	"sync"
	"testing/fstest"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/system"
	"cogentcore.org/core/text/shaped"
	"cogentcore.org/core/text/shaped/shapers/shapedgt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
	"github.com/go-text/typesetting/font"
)

type manualImplementation struct {
	owner                   *app.Application
	config                  ManualConfig
	lifetime                context.Context
	mu                      sync.Mutex // wake may be called by a worker, never acquires adapter dispatch ownership.
	windows                 []*managedManualWindow
	nextWindow              uint64
	fonts                   fontMemoryFS
	workbenchMode           *manualWorkbenchMode
	failure                 error // first failure, protected by mu; inspection never clears it.
	ownerBudget, dispatched int   // dispatch owner only; bounded real Custom drains.
}
type managedManualWindow struct {
	state               *nativeWindow
	core                *core.ManualWindow
	incarnation         uint64
	nextRecord          uint64
	records             map[*record]uint64
	dockHeaders         map[*core.Tab]uint64
	layoutPending       bool
	lastSnapshot        ManualWindowSnapshot
	suppressDomainClose bool
	domain              context.Context
}
type fontMemoryFS struct{ fstest.MapFS }
type fontMemoryFile struct {
	*bytes.Reader
	info fs.FileInfo
}

func (f *fontMemoryFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f *fontMemoryFile) Close() error               { return nil }
func (f fontMemoryFS) Open(name string) (fs.File, error) {
	entry, err := f.MapFS.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := entry.Stat()
	if err != nil {
		entry.Close()
		return nil, err
	}
	if info.IsDir() {
		return entry, nil
	}
	entry.Close()
	return &fontMemoryFile{Reader: bytes.NewReader(f.MapFS[name].Data), info: info}, nil
}

func validatedManualConfig(config ManualConfig) (ManualConfig, fontMemoryFS, error) {
	if config.MaxControls == 0 {
		config.MaxControls = 4096
	}
	if config.MaxControls < 1 || config.MaxControls > 65536 {
		return config, fontMemoryFS{}, ErrManualConfig
	}
	if config.Clock == nil || config.Executor == nil || config.OwnerBatch < 1 || config.OwnerBatch > 1<<20 || config.MaxWindows < 1 || config.MaxWindows > 64 || config.EventCapacity < 4 || config.EventCapacity > 1<<20 || config.LogicalDPI <= 0 || config.LogicalDPI > 1920 || math.IsNaN(float64(config.LogicalDPI)) || math.IsInf(float64(config.LogicalDPI), 0) || len(config.Fonts) < 1 || len(config.Fonts) > 32 {
		return config, fontMemoryFS{}, ErrManualConfig
	}
	clock, err := config.Clock.ReadySnapshot()
	if err != nil || clock.Closed || config.Executor.Snapshot().Closed {
		return config, fontMemoryFS{}, ErrManualConfig
	}
	if err = config.Theme.Validate(); err != nil {
		return config, fontMemoryFS{}, err
	}
	if !config.Theme.Default.Foreground.Set || !config.Theme.Default.Background.Set || !config.Theme.Default.Typography.Set || config.Theme.Default.Typography.Value.Family == "" {
		return config, fontMemoryFS{}, fmt.Errorf("%w: explicit managed colors and font family required", ErrManualConfig)
	}
	fonts := fontMemoryFS{MapFS: fstest.MapFS{}}
	total := 0
	copied := make([]ManualFont, len(config.Fonts))
	for i, f := range config.Fonts {
		if !fs.ValidPath(f.Name) || f.Name == "." || fonts.MapFS[f.Name] != nil || len(f.Data) == 0 || len(f.Data) > 4<<20 {
			return config, fontMemoryFS{}, ErrManualConfig
		}
		total += len(f.Data)
		if total > 16<<20 {
			return config, fontMemoryFS{}, ErrManualConfig
		}
		digest := sha256.Sum256(f.Data)
		if f.SHA256 != hex.EncodeToString(digest[:]) {
			return config, fontMemoryFS{}, fmt.Errorf("%w: font digest mismatch", ErrManualConfig)
		}
		data := slices.Clone(f.Data)
		if _, err = font.ParseTTC(bytes.NewReader(data)); err != nil {
			return config, fontMemoryFS{}, fmt.Errorf("%w: invalid font: %v", ErrManualConfig, err)
		}
		copied[i] = ManualFont{Name: f.Name, SHA256: f.SHA256, Data: data}
		fonts.MapFS[f.Name] = &fstest.MapFile{Data: data, Mode: 0444}
	}
	config.Fonts = copied
	config.Theme = config.Theme.Clone()
	return config, fonts, nil
}

// NewManual is additive and offscreen-only. It constructs registered windows
// without starting an event loop, worker, frame, or virtual-time advance.
var manualOwners = struct {
	sync.Mutex
	owners map[*app.Application]*Manual
	active *Manual
}{owners: map[*app.Application]*Manual{}}

type manualServices struct{}

func (manualServices) ReadClipboard(context.Context) (string, error) { return "", ErrManualUnsupported }
func (manualServices) WriteClipboard(context.Context, string) error  { return ErrManualUnsupported }
func (manualServices) ChooseFile(context.Context, app.FileOptions) (string, error) {
	return "", ErrManualUnsupported
}
func (manualServices) SetMenu(context.Context, []app.MenuGroup) error { return ErrManualUnsupported }

func NewManual(ctx context.Context, a *app.Application, config ManualConfig) (*Manual, error) {
	if ctx == nil || a == nil {
		return nil, ErrManualConfig
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if system.TheApp.Platform() != system.Offscreen {
		return nil, ErrManualUnsupported
	}
	config, fonts, err := validatedManualConfig(config)
	if err != nil {
		return nil, err
	}
	if a.Context().Err() != nil {
		return nil, ErrManualClosed
	}
	if len(a.Windows()) == 0 {
		return nil, ErrManualConfig
	}
	if _, err = a.Services(); !errors.Is(err, app.ErrUnavailable) {
		return nil, fmt.Errorf("%w: native services already installed", ErrManualUnsupported)
	}
	m := &Manual{}
	m.self = m
	m.implementation = &manualImplementation{owner: a, config: config, lifetime: ctx, fonts: fonts, workbenchMode: &manualWorkbenchMode{}}
	m.implementation.workbenchMode.active.Store(true)
	manualOwners.Lock()
	if manualOwners.active != nil {
		manualOwners.Unlock()
		return nil, ErrManualBusy
	}
	manualOwners.owners[a] = m
	manualOwners.active = m
	manualOwners.Unlock()
	constructed := false
	defer func() {
		if !constructed {
			// Do closes domain lifetimes on panic/Goexit. Always attempt every native
			// retirement and release the exclusive owner claim while Go unwinds.
			defer func() {
				manualOwners.Lock()
				delete(manualOwners.owners, a)
				if manualOwners.active == m {
					manualOwners.active = nil
				}
				manualOwners.Unlock()
				for _, spec := range a.Windows() {
					_ = a.Invalidate(spec.ID)
				}
			}()
			m.implementation.rollback(m.implementation.windowList())
			if len(m.implementation.windowList()) == 0 {
				m.implementation.workbenchMode.active.Store(false)
			}
		}
	}()
	if err = m.implementation.syncTopology(ctx); err != nil {
		return nil, err
	}
	a.SetServices(manualServices{})
	a.SetWake(m.implementation.wake)
	constructed = true
	return m, nil
}

// Retire the remainder even if an abnormal private cleanup unwinds this call.
// Ordinary error rollback preserves open domain registration and dirty state;
// panic/Goexit keep Go's own unwinding semantics rather than swallowing failure.
func (p *manualImplementation) rollback(windows []*managedManualWindow) {
	if len(windows) == 0 {
		return
	}
	defer p.rollback(windows[1:])
	windows[0].suppressDomainClose = true
	_ = p.retire(context.Background(), windows[0])
}
func (p *manualImplementation) fail(err error) {
	if err == nil || errors.Is(err, app.ErrClosed) || errors.Is(err, app.ErrWindowClosed) {
		return
	}
	p.mu.Lock()
	if p.failure == nil {
		p.failure = err
	}
	p.mu.Unlock()
}
func (p *manualImplementation) error() error { p.mu.Lock(); defer p.mu.Unlock(); return p.failure }
func (m *Manual) enter(ctx context.Context, cleanup bool) error {
	if m == nil || m.self != m || m.implementation == nil || ctx == nil {
		return ErrManualConfig
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !m.busy.CompareAndSwap(false, true) {
		return ErrManualBusy
	}
	if !cleanup {
		if err := m.implementation.error(); err != nil {
			m.busy.Store(false)
			return err
		}
		if m.closed.Load() {
			m.busy.Store(false)
			return ErrManualClosed
		}
		if err := m.implementation.lifetime.Err(); err != nil {
			m.busy.Store(false)
			return err
		}
		if m.implementation.owner.Context().Err() != nil {
			m.busy.Store(false)
			return ErrManualClosed
		}
	}
	return nil
}
func (p *manualImplementation) windowList() []*managedManualWindow {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.windows)
}
func (p *manualImplementation) wake() {
	for _, w := range p.windowList() {
		n := w.state
		if source := n.source.Load(); source != nil && !n.closing.Load() && n.pending.CompareAndSwap(false, true) {
			source.Custom("vellum.wake")
		}
	}
}
func (p *manualImplementation) refresh(w *managedManualWindow) error {
	n := w.state
	if n.closing.Load() {
		return nil
	}
	var updateErr error
	err := p.owner.Do(func() {
		n.rootMu.Lock()
		defer n.rootMu.Unlock()
		if !p.owner.TakeDirty(n.window.ID) {
			return
		}
		w.layoutPending = true
		view := n.window.View()
		release, err := p.prepareManualView(n.window.ID, view)
		if err != nil {
			updateErr = err
			return
		}
		defer release()
		n.root, updateErr = reconcile(n.host, n.root, view, p.owner, n.window.ID, p.config.Theme)
		if updateErr != nil {
			return
		}
		bindControlWake(n.root, func(e events.Event) { p.custom(n, e) }, p.fail)
		bindGridScopes(n.root)
		bindTreeScopes(n.root)
		flushDialogs(n.root)
		bindWorkbenchInputScopes(n.root)
		if updateErr = applyFocus(n.root, view.Focus, &n.lastFocus); updateErr != nil {
			return
		}
		if updateErr = bindTextWindow(n); updateErr != nil {
			return
		}
		accessibilityBindFocusInvalidation(n.root, p.owner, n.window.ID)
		present := map[*record]uint64{}
		headers := map[*core.Tab]uint64{}
		var recordIDs func(*record)
		recordIDs = func(r *record) {
			if r == nil {
				return
			}
			id, ok := w.records[r]
			if !ok {
				if w.nextRecord == math.MaxUint64 {
					updateErr = ErrManualConfig
					return
				}
				w.nextRecord++
				id = w.nextRecord
			}
			present[r] = id
			if dock, ok := r.node.Custom.(*Workbench); r.node.Kind == ui.Extension && ok {
				for _, panel := range dock.Panels {
					if tab := dock.buttons[panel.ID]; tab != nil {
						token, exists := w.dockHeaders[tab]
						if !exists {
							if w.nextRecord == math.MaxUint64 {
								updateErr = ErrManualConfig
								return
							}
							w.nextRecord++
							token = w.nextRecord
						}
						headers[tab] = token
					}
				}
			}
			retainedWorkbenchRecords(r, func(_ string, child *record) { recordIDs(child) })
			if edit := gridEditingRecord(r); edit != nil {
				recordIDs(edit)
			}
			for _, c := range r.children {
				recordIDs(c)
			}
		}
		recordIDs(n.root)
		w.records = present
		w.dockHeaders = headers
		w.layoutPending = true
	})
	p.fail(errors.Join(err, updateErr))
	return errors.Join(err, updateErr)
}
func (p *manualImplementation) custom(n *nativeWindow, e events.Event) {
	n.pending.Store(false)
	handleTextEvent(n, e)
	if p.ownerBudget > 0 {
		processed, err := p.owner.DrainN(min(p.config.OwnerBatch, p.ownerBudget))
		p.dispatched += processed
		p.ownerBudget -= processed
		p.fail(err)
	}
	if ce, ok := e.(*events.CustomEvent); ok && ce.Data == "vellum.wake" {
		e.SetHandled()
	}
}
func (p *manualImplementation) cleanup(w *managedManualWindow) error {
	n := w.state
	cleanup := func() {
		n.rootMu.Lock()
		defer n.rootMu.Unlock()
		closeTextWindow(n)
		disposeControlTree(n.root)
		n.root = nil
	}
	err := p.owner.Do(cleanup)
	if errors.Is(err, app.ErrClosed) {
		cleanup()
		return nil
	}
	return err
}
func (p *manualImplementation) syncTopology(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	desired := p.owner.Windows()
	if len(desired) > p.config.MaxWindows {
		return ErrManualConfig
	}
	active := map[string]bool{}
	for _, spec := range desired {
		active[spec.ID] = true
	}
	for _, w := range p.windowList() {
		if !active[w.state.window.ID] {
			if err := p.retire(ctx, w); err != nil {
				return err
			}
		}
	}
	for _, spec := range desired {
		var existing *managedManualWindow
		for _, w := range p.windowList() {
			if w.state.window.ID == spec.ID {
				existing = w
				break
			}
		}
		if existing != nil {
			continue
		}
		if p.nextWindow == math.MaxUint64 {
			return ErrManualConfig
		}
		if spec.Width < 1 || spec.Height < 1 || float32(int(spec.Width)) != spec.Width || float32(int(spec.Height)) != spec.Height {
			return ErrManualConfig
		}
		domain, err := p.owner.WindowContext(spec.ID)
		if err != nil || domain.Err() != nil {
			return ErrManualClosed
		}
		w := &managedManualWindow{domain: domain, state: &nativeWindow{window: spec}, records: map[*record]uint64{}, layoutPending: true}
		p.nextWindow++
		w.incarnation = p.nextWindow
		n := w.state
		if err := p.owner.Do(func() {
			n.body = core.NewBody(spec.Title)
			n.body.Scene.GPUSceneRendering = p.config.GPUSceneRendering
			setGridScrollReuse(n.body.Scene, p.config.ExperimentalGridScrollReuse)
			n.host = newAccessibilityHost(n)
			n.body.Styler(func(s *styles.Style) {
				applyStyle(s, ui.Node{Kind: ui.Column}, p.config.Theme)
				// RunManual already fixes the physical scene size. Reusing that size
				// as a body-content minimum ignores scene padding/borders and clips
				// grown roots at the far edge. Grow into the actual allocated content.
				s.Grow.Set(1, 1)
			})
			n.host.Styler(func(s *styles.Style) { s.Grow.Set(1, 1); s.Direction = styles.Column })
		}); err != nil {
			return err
		}
		p.mu.Lock()
		p.windows = append(p.windows, w)
		p.mu.Unlock()
		if err := p.refresh(w); err != nil {
			return err
		}
		n.body.Scene.OnFirst(events.Custom, func(e events.Event) { p.custom(n, e) })
		bindAccessibilityKeyboard(n, p.owner, nil)
		if p.config.Commands != nil {
			bindCommandShortcuts(n.body.Scene, p.owner, spec.ID, p.config.Commands, nil)
		}
		bindCloseRequest(n, p.owner, nil)
		n.body.OnShow(func(events.Event) {
			p.fail(p.owner.Do(func() {
				if n.root != nil {
					flushDialogs(n.root)
					bindWorkbenchInputScopes(n.root)
					p.fail(applyFocus(n.root, n.root.node.Focus, &n.lastFocus))
				}
			}))
		})
		n.body.OnClose(func(events.Event) {
			n.closing.Store(true)
			if !w.suppressDomainClose {
				// Once the owner stops, model notifications are no longer valid.
				// Native/control cleanup below still retires every mounted record.
				if err := p.owner.Do(func() { n.rootMu.Lock(); defer n.rootMu.Unlock(); p.fail(closeWorkbenchExtensions(n.root)) }); err != nil && !errors.Is(err, app.ErrClosed) {
					p.fail(err)
				}
				// A workbench close notification may already have retired this
				// registration. Preserve the original captured lifetime check.
				if w.domain.Err() == nil {
					p.fail(p.owner.CloseWindow(spec.ID))
				}
			}
			p.fail(p.cleanup(w))
		})
		w.core, err = n.body.NewWindow().RunManual(ctx, core.ManualWindowConfig{Size: image.Pt(int(spec.Width), int(spec.Height)), LogicalDPI: p.config.LogicalDPI, EventCapacity: p.config.EventCapacity, Now: p.config.Clock.Now, AfterFunc: func(c context.Context, d time.Duration, f func()) (core.ManualTimer, error) {
			return p.config.Clock.AfterFunc(c, d, f)
		}, NewShaper: func() (shaped.Shaper, error) { return shapedgt.NewShaperWithFonts([]fs.FS{p.fonts}), nil }})
		if err != nil {
			return err
		}
		n.source.Store(n.body.SystemEvents())
		// No Cocoa attachment or native-main readiness worker is started. The
		// existing originating boundary still synchronizes legacy editor changes.
		t, err := newTextWindow(p.owner, &n.source.Load().Deque)
		if err != nil {
			return err
		}
		t.window = n
		t.manual = true
		textWindows.Lock()
		textWindows.windows[n.window.ID] = t
		textWindows.Unlock()
		lifetime, err := p.owner.WindowContext(spec.ID)
		if err != nil {
			return err
		}
		t.lifetime.Store(&textLifetime{ctx: lifetime})
		n.input.Store(t)
		var bindingErr error
		if err = p.owner.Do(func() { bindingErr = bindTextWindow(n) }); err != nil {
			return err
		}
		if bindingErr != nil {
			return bindingErr
		}
	}
	return nil
}
func (p *manualImplementation) retire(ctx context.Context, w *managedManualWindow) error {
	w.state.closing.Store(true)
	var err error
	if w.core != nil {
		err = p.cleanup(w)
		err = errors.Join(err, w.core.Close(ctx))
	} else if w.state.body != nil {
		err = p.cleanup(w)
		w.state.body.Scene.Destroy()
	}
	if err != nil {
		return err
	} // Keep ownership for a later explicit cleanup retry.
	p.mu.Lock()
	for i, entry := range p.windows {
		if entry == w {
			p.windows = append(p.windows[:i], p.windows[i+1:]...)
			break
		}
	}
	p.mu.Unlock()
	return nil
}
func (p *manualImplementation) find(id string) *managedManualWindow {
	for _, w := range p.windowList() {
		if w.state.window.ID == id {
			return w
		}
	}
	return nil
}

// Core may handle these after ordinary widget/command routing. Those fallback
// handlers access OS services, global windows/settings, or unmanaged search UI.
// Refuse before reconciling or admitting an originating event; qualified model
// commands and ordinary focus/navigation/editing remain actual routed input.
func manualUnqualifiedInput(input ManualInput) bool {
	if (input.Type == events.MouseDown || input.Type == events.MouseUp) && (input.Button != events.Left || key.HasAnyModifier(input.Modifiers, key.Control)) {
		return true
	}
	if !input.Type.IsKey() {
		return false
	}
	if (input.Code == key.CodeC || input.Code == key.CodeX || input.Code == key.CodeV) && key.HasAnyModifier(input.Modifiers, key.Control, key.Meta) {
		return true
	}
	switch keymap.Of(key.NewChord(input.Rune, input.Code, input.Modifiers)) {
	case keymap.Copy, keymap.Cut, keymap.Paste, keymap.PasteHist,
		keymap.WinSnapshot, keymap.ZoomIn, keymap.ZoomOut, keymap.Refresh,
		keymap.Find, keymap.Menu, keymap.WinFocusNext:
		return true
	}
	return false
}

func (m *Manual) Input(ctx context.Context, input ManualInput) error {
	if err := m.enter(ctx, false); err != nil {
		return err
	}
	defer m.busy.Store(false)
	if manualUnqualifiedInput(input) {
		return ErrManualUnsupported
	}
	p := m.implementation
	if err := p.syncTopology(ctx); err != nil {
		return err
	}
	for _, w := range p.windowList() {
		if err := p.refresh(w); err != nil {
			return err
		}
	}
	w := p.find(input.Window)
	if w == nil {
		return ErrManualClosed
	}
	if w.domain.Err() != nil {
		return ErrManualClosed
	}
	if w.incarnation != input.WindowIncarnation {
		return ErrManualStale
	}
	position := input.Position
	if input.Target != "" {
		if w.layoutPending {
			return ErrManualLayout
		}
		found := false
		for _, e := range w.lastSnapshot.Elements {
			if e.ID == input.Target {
				found = true
				if e.Incarnation != input.TargetIncarnation {
					return ErrManualStale
				}
				if !e.Enabled || !e.Visible {
					return ErrManualDisabled
				}
				if input.Type == events.MouseDown || input.Type == events.MouseUp || input.Type == events.MouseMove || input.Type == events.Scroll {
					if e.Bounds.Empty() {
						return ErrManualLayout
					}
					position = e.Bounds.Min.Add(e.Bounds.Size().Div(2))
				}
				break
			}
		}
		if !found {
			return ErrManualStale
		}
	}
	return w.core.Input(ctx, core.ManualInput{Type: input.Type, Rune: input.Rune, Code: input.Code, Modifiers: input.Modifiers, Button: input.Button, Position: position, Delta: input.Delta})
}

// Step spends at most maximum combined originating events and owner callbacks.
// Render frames are a subset of originating events, not extra charged units.
func (m *Manual) Step(ctx context.Context, maximum int) (ManualStepResult, error) {
	if maximum < 1 || maximum > 1<<20 {
		return ManualStepResult{}, ErrManualConfig
	}
	if err := m.enter(ctx, false); err != nil {
		return ManualStepResult{}, err
	}
	defer m.busy.Store(false)
	p := m.implementation
	defer func() { p.ownerBudget = 0 }()
	var result ManualStepResult
	p.dispatched = 0
	if err := p.syncTopology(ctx); err != nil {
		return result, err
	}
	remaining := maximum
	for remaining > 0 {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		progress := false
		for _, w := range p.windowList() {
			if remaining == 0 {
				break
			}
			if err := p.refresh(w); err != nil {
				return result, err
			}
			before := p.dispatched
			p.ownerBudget = remaining - 1
			step, err := w.core.Step(ctx, 1)
			p.ownerBudget = 0
			consumed := step.Events + p.dispatched - before
			remaining -= consumed
			result.Events += step.Events
			result.Frames += step.Frames
			result.OwnerCallbacks += p.dispatched - before
			progress = progress || consumed > 0
			if err != nil {
				return result, err
			}
			if err = p.error(); err != nil {
				return result, err
			}
			if step.Frames > 0 {
				if err = p.captureWindow(ctx, w); err != nil {
					return result, err
				}
			}
		}
		if remaining > 0 && p.owner.Context().Err() == nil {
			processed, err := p.owner.DrainN(min(remaining, p.config.OwnerBatch))
			remaining -= processed
			result.OwnerCallbacks += processed
			progress = progress || processed > 0
			if err != nil {
				return result, err
			}
		}
		if err := p.syncTopology(ctx); err != nil {
			return result, err
		}
		if !progress {
			break
		}
	}
	return result, nil
}

// RequestFrame only requests originating paint. It performs no reconciliation,
// callback, owner drain, layout or time advance; subsequent Step spends the work.
func (m *Manual) RequestFrame(ctx context.Context, window string) error {
	if err := m.enter(ctx, false); err != nil {
		return err
	}
	defer m.busy.Store(false)
	w := m.implementation.find(window)
	if w == nil {
		return ErrManualClosed
	}
	if w.domain.Err() != nil {
		return ErrManualClosed
	}
	requester, ok := any(w.core).(interface{ RequestFrame(context.Context) error })
	if !ok {
		return ErrManualUnsupported
	}
	return requester.RequestFrame(ctx)
}

// Frame requests paint through the originating queue and processes one node.
// Earlier Front nodes may win; Frames==0 means no new capture was made.
func (m *Manual) Frame(ctx context.Context, window string) (ManualStepResult, error) {
	if err := m.enter(ctx, false); err != nil {
		return ManualStepResult{}, err
	}
	defer m.busy.Store(false)
	p := m.implementation
	defer func() { p.ownerBudget = 0 }()
	if err := p.syncTopology(ctx); err != nil {
		return ManualStepResult{}, err
	}
	w := p.find(window)
	if w == nil {
		return ManualStepResult{}, ErrManualClosed
	}
	if w.domain.Err() != nil {
		return ManualStepResult{}, ErrManualClosed
	}
	if err := p.refresh(w); err != nil {
		return ManualStepResult{}, err
	}
	p.dispatched = 0
	p.ownerBudget = p.config.OwnerBatch
	step, err := w.core.Frame(ctx)
	p.ownerBudget = 0
	result := ManualStepResult{Events: step.Events, Frames: step.Frames, OwnerCallbacks: p.dispatched}
	if err == nil {
		err = p.error()
	}
	if err == nil && step.Frames > 0 {
		err = p.captureWindow(ctx, w)
	}
	return result, err
}

func manualAccessibilityIndex(nodes []accessibilityNode) map[string]accessibilityNode {
	index := make(map[string]accessibilityNode, len(nodes))
	for _, node := range nodes {
		if _, exists := index[node.ID]; !exists {
			index[node.ID] = node // Full AX identity separates managed nodes from group-local option IDs.
		}
	}
	return index
}

func (p *manualImplementation) captureWindow(ctx context.Context, w *managedManualWindow) error {
	n := w.state
	native, err := w.core.Snapshot(ctx)
	if err != nil {
		return err
	}
	var result ManualWindowSnapshot
	gridPending := false
	var gridCaptureErr error
	err = p.owner.Do(func() {
		result = ManualWindowSnapshot{ID: n.window.ID, Incarnation: w.incarnation, Frames: native.Frames, Events: native.Events, Timers: native.Timers, Blocked: native.Blocked, ReadyPaint: native.ReadyPaint, FuturePaint: native.FuturePaint, Closed: native.Closed, Failure: native.Failure, Now: native.Now, NextPaint: native.NextPaint, Elements: []ManualElement{}}
		bounds := accessibilityBounds(n.root)
		nodes, _ := accessibilitySnapshot(n.root, nil)
		byID := manualAccessibilityIndex(nodes)
		focusTargets := map[string]accessibilityFocusTarget{}
		accessibilityNativeFocusTargets(n.root, "", focusTargets)
		var collect func(*record, string, string)
		collect = func(r *record, parent, prefix string) {
			if r == nil {
				return
			}
			id := prefix + r.node.ID
			axID := prefix + accessibilityManagedID(r.node.ID)
			value := any(r.node.Text)
			selection := [2]int{-1, 0}
			focus := 0
			selected := false
			node, visible := byID[axID]
			if visible {
				value = node.Value
				selected = node.Selected
				if node.TextSelection != nil {
					selection = *node.TextSelection
					focus = node.TextFocus
				}
			}
			// The AX node was captured before this cache update. Keep the live
			// field text and directional selection together, with one checked read.
			if field, ok := r.widget.(*core.TextField); ok {
				state := field.EditingState()
				cacheFieldEditingState(r, state)
				value = state.Text
				offsets := utf16Selection(state.Text, state.Anchor, state.Focus)
				selection, focus = offsets.rangeValue, offsets.focus
			}
			b := bounds[axID]
			rect := image.Rect(int(b[0]), int(b[1]), int(b[0]+b[2]), int(b[1]+b[3]))
			focused := r.widget.AsWidget().StateIs(states.Focused)
			if ui.IsFocusable(r.node.Kind) && gridEditingRecord(r) == nil {
				focused = focusTargets[axID].focused
			}
			if focused {
				result.FocusedID = id
			}
			result.Elements = append(result.Elements, ManualElement{ID: id, Parent: parent, Incarnation: w.records[r], Kind: r.node.Kind, Name: r.node.Accessibility.Name, Value: value, Enabled: !controlDisabled(r), Focused: focused, Visible: visible, Selected: selected, Selection: selection, TextFocus: focus, Bounds: rect})

			if visible && len(r.node.CanvasObjects) > 0 && r.control != nil && r.control.placement != nil {
				binding := r.control.placement
				frame := binding.frame.Load()
				if !canvasCompletedFrame(binding, frame) {
					gridPending = true
				} else {
					inspection := inspectCanvasFrame(binding, frame)
					inspection.ID = prefix + inspection.ID
					for i := range inspection.Objects {
						object := &inspection.Objects[i]
						object.ID = prefix + object.ID
						rect := object.Bounds
						result.Elements = append(result.Elements, ManualElement{ID: object.ID, Parent: id, Incarnation: inspection.Incarnation, Kind: ui.Canvas2D, Name: object.Name, Value: object.ObjectID, Enabled: !controlDisabled(r) && binding.window.active == nil, Visible: true, Selected: object.Selected, Bounds: image.Rect(int(math.Floor(float64(rect[0]))), int(math.Floor(float64(rect[1]))), int(math.Ceil(float64(rect[0]+rect[2]))), int(math.Ceil(float64(rect[1]+rect[3]))))})
					}
					result.Canvases = append(result.Canvases, inspection)
				}
			}
			if menuOpen(r) {
				popup := r.control.menu.popup
				if popup.manualIncarnation == 0 {
					if w.nextRecord == math.MaxUint64 {
						gridCaptureErr = ErrManualConfig
						return
					}
					w.nextRecord++
					popup.manualIncarnation = w.nextRecord
				}
				for _, o := range r.control.menu.options {
					if o.Separator {
						continue
					}
					id := prefix + menuActionID(r.node.ID, o.ID)
					b := bounds[id]
					item := popup.items[o.ID]
					focused := item != nil && item.ContainsFocus()
					if focused {
						result.FocusedID = id
					}
					value := ""
					if o.Checkable {
						value = fmt.Sprint(o.Checked)
					}
					result.Elements = append(result.Elements, ManualElement{ID: id, Parent: prefix + r.node.ID, Incarnation: popup.manualIncarnation, Kind: ui.MenuButton, Name: o.Label, Value: value, Enabled: !controlDisabled(r) && !o.Disabled, Visible: true, Focused: focused, Bounds: image.Rect(int(b[0]), int(b[1]), int(b[0]+b[2]), int(b[1]+b[3]))})
				}
			}
			if visible && r.control != nil && r.control.tabs != nil {
				for _, o := range r.node.Options {
					if !o.Closable {
						continue
					}
					binding := r.control.tabs.bindings[o.ID]
					if binding.manualIncarnation == 0 {
						if w.nextRecord == math.MaxUint64 {
							gridCaptureErr = ErrManualConfig
							return
						}
						w.nextRecord++
						binding.manualIncarnation = w.nextRecord
					}
					id := prefix + tabCloseID(r.node.ID, o.ID)
					b := bounds[id]
					box := image.Rect(int(b[0]), int(b[1]), int(b[0]+b[2]), int(b[1]+b[3]))
					result.Elements = append(result.Elements, ManualElement{ID: id, Parent: prefix + r.node.ID, Incarnation: binding.manualIncarnation, Kind: ui.Button, Name: "Close " + o.Label, Enabled: tabCloseEnabled(r, o.ID, binding), Visible: true, Bounds: box})
				}
			}
			if r.control != nil && r.control.disclosure != nil {
				id := prefix + accessibilityDisclosureID(r.node.ID)
				heading := r.control.disclosure.header
				b := bounds[id]
				box := image.Rect(int(b[0]), int(b[1]), int(b[0]+b[2]), int(b[1]+b[3]))
				result.Elements = append(result.Elements, ManualElement{ID: id, Parent: prefix + r.node.ID, Incarnation: w.records[r], Kind: ui.Disclosure, Name: r.node.Text, Value: r.node.Checked, Enabled: !controlDisabled(r), Focused: heading.ContainsFocus(), Visible: visible && heading.IsVisible(), Bounds: box})
			}
			if r.control != nil && r.control.split != nil {
				id := prefix + accessibilitySplitHandleID(r.node.ID)
				handle := r.control.split.handle
				b := bounds[id]
				box := image.Rect(int(b[0]), int(b[1]), int(b[0]+b[2]), int(b[1]+b[3]))
				result.Elements = append(result.Elements, ManualElement{ID: id, Parent: prefix + r.node.ID, Incarnation: w.records[r], Kind: ui.SplitPane, Name: r.node.Accessibility.Name, Value: r.node.Number, Enabled: !controlDisabled(r), Focused: handle.ContainsFocus(), Visible: visible, Bounds: box})
			}

			if r.control != nil && r.control.outline != nil && visible {
				inspection, treeErr := InspectTree(p.owner, r.node.Tree)
				if treeErr != nil {
					if errors.Is(treeErr, ErrTreeLayout) {
						gridPending = true
					} else {
						gridCaptureErr = errors.Join(gridCaptureErr, treeErr)
					}
				} else {
					qualifyManualTree(&inspection, prefix)
					result.Trees = append(result.Trees, inspection)
					for _, row := range inspection.Rows {
						if row.Bounds.Empty() {
							continue
						}
						result.Elements = append(result.Elements, ManualElement{ID: row.ID, Parent: prefix + r.node.ID, Incarnation: inspection.RowIncarnation, Kind: ui.DataTree, Name: row.Label, Value: row.Label, Enabled: !controlDisabled(r), Visible: true, Selected: row.Selected, Bounds: row.Bounds})
						if row.HasChildren && !row.DisclosureBounds.Empty() {
							result.Elements = append(result.Elements, ManualElement{ID: row.DisclosureID, Parent: row.ID, Incarnation: inspection.RowIncarnation, Kind: ui.Button, Name: "Toggle expansion", Enabled: !controlDisabled(r) && r.node.TreeDisclosure != nil, Visible: true, Bounds: row.DisclosureBounds})
						}
					}
					for _, page := range inspection.Pages {
						result.Elements = append(result.Elements, ManualElement{ID: page.ID, Parent: prefix + r.node.ID, Incarnation: inspection.RowIncarnation, Kind: ui.Button, Name: page.Name, Enabled: !controlDisabled(r) && page.Enabled, Visible: true, Bounds: page.Bounds})
					}
				}
			}
			if r.control != nil && r.control.grid != nil && visible {
				inspection, gridErr := InspectGrid(p.owner, r.node.Grid)
				if gridErr != nil {
					if errors.Is(gridErr, ErrGridLayout) {
						gridPending = true
					} else {
						gridCaptureErr = errors.Join(gridCaptureErr, gridErr)
					}
				} else {
					qualifyManualGrid(&inspection, prefix)
					result.Grids = append(result.Grids, inspection)
					for _, cell := range inspection.Cells {
						if cell.Bounds.Empty() {
							continue
						}
						result.Elements = append(result.Elements, ManualElement{ID: cell.ID, Parent: prefix + r.node.ID, Incarnation: inspection.CellIncarnation, Kind: ui.DataGrid, Name: cell.Text, Value: cell.Text, Enabled: !controlDisabled(r), Visible: visible, Selected: cell.Active, Bounds: cell.Bounds})
					}
				}
			}
			if dock, ok := r.node.Custom.(*Workbench); r.node.Kind == ui.Extension && ok {
				accessibilityWorkbenchLeaves(dock, func(group *workbench.Group) {
					for _, panel := range group.Tabs {
						tab := dock.buttons[panel]
						if tab == nil {
							continue
						}
						target := prefix + accessibilityDockOptionID(r.node.ID, panel)
						b := bounds[target]
						node, shown := byID[target]
						focused := tab.ContainsFocus()
						if focused {
							result.FocusedID = target
						}
						result.Elements = append(result.Elements, ManualElement{ID: target, Parent: id, Incarnation: w.dockHeaders[tab], Kind: ui.Tabs, Name: node.Name, Value: panel, Visible: shown, Enabled: !controlDisabled(r) && !node.Disabled, Focused: focused, Selected: group.Active == panel, Bounds: image.Rect(int(b[0]), int(b[1]), int(b[0]+b[2]), int(b[1]+b[3]))})
					}
				})
				retainedWorkbenchRecords(r, func(panel string, child *record) {
					collect(child, id, prefix+accessibilityWorkbenchNamespace(r.node.ID, panel))
				})
			}
			if edit := gridEditingRecord(r); edit != nil {
				collect(edit, prefix+r.node.ID, prefix)
			}
			for _, c := range r.children {
				collect(c, prefix+r.node.ID, prefix)
			}
		}
		collect(n.root, "", "")
		sort.Slice(result.Elements, func(i, j int) bool { return result.Elements[i].ID < result.Elements[j].ID })
	})
	err = errors.Join(err, gridCaptureErr)
	if err == nil {
		w.lastSnapshot = result
		w.layoutPending = gridPending
	}
	return err
}

type manualSelection struct {
	rangeValue [2]int
	focus      int
}

func utf16Selection(text string, anchor, focus int) manualSelection {
	offsets := []int{0}
	total := 0
	for _, r := range text {
		if r > 0xffff {
			total += 2
		} else {
			total++
		}
		offsets = append(offsets, total)
	}
	if anchor < 0 || focus < 0 || anchor >= len(offsets) || focus >= len(offsets) {
		return manualSelection{rangeValue: [2]int{-1, 0}}
	}
	a, f := offsets[anchor], offsets[focus]
	return manualSelection{rangeValue: [2]int{min(a, f), max(a, f) - min(a, f)}, focus: f}
}
func (m *Manual) Snapshot(ctx context.Context) (ManualSnapshot, error) {
	if err := m.enter(ctx, true); err != nil {
		return ManualSnapshot{}, err
	}
	defer m.busy.Store(false)
	p := m.implementation
	clock, err := p.config.Clock.ReadySnapshot()
	if err != nil {
		return ManualSnapshot{}, err
	}
	s := ManualSnapshot{Application: p.owner.Pending(), Clock: clock, Executor: p.config.Executor.Snapshot(), Closed: m.closed.Load(), Failure: p.error(), Windows: []ManualWindowSnapshot{}}
	for _, w := range p.windowList() {
		native, err := w.core.Snapshot(ctx)
		if err != nil {
			return s, err
		}
		captured := w.lastSnapshot
		captured.Elements = slices.Clone(captured.Elements)
		captured.Grids = cloneManualGrids(captured.Grids)
		captured.Trees = cloneManualTrees(captured.Trees)
		captured.Canvases = cloneManualCanvases(captured.Canvases)
		captured.ID = w.state.window.ID
		captured.Incarnation = w.incarnation
		captured.Events = native.Events
		captured.Timers = native.Timers
		captured.Blocked = native.Blocked
		captured.ReadyPaint = native.ReadyPaint
		captured.FuturePaint = native.FuturePaint
		captured.Frames = native.Frames
		captured.Closed = native.Closed
		captured.DomainClosed = w.domain.Err() != nil
		captured.RetirementPending = captured.DomainClosed && !native.Closed
		captured.Failure = native.Failure
		if s.Failure == nil {
			s.Failure = native.Failure
		}
		captured.LayoutPending = w.layoutPending || s.Application.DirtyWindows > 0 || native.Events > 0 || native.ReadyPaint
		captured.Now = native.Now
		captured.NextPaint = native.NextPaint
		s.Windows = append(s.Windows, captured)
	}
	return s, nil
}
func (m *Manual) Capture(ctx context.Context, window string) (*image.RGBA, error) {
	if err := m.enter(ctx, false); err != nil {
		return nil, err
	}
	defer m.busy.Store(false)
	w := m.implementation.find(window)
	if w == nil {
		return nil, ErrManualClosed
	}
	if w.domain.Err() != nil {
		return nil, ErrManualClosed
	}
	pending := m.implementation.owner.Pending()
	native, err := w.core.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	if w.layoutPending || pending.DirtyWindows > 0 || pending.Ordinary+pending.ReadyCompletions > 0 || native.Events > 0 || native.ReadyPaint {
		return nil, ErrManualLayout
	}
	return w.core.Image(ctx)
}
func (m *Manual) Close(ctx context.Context) error {
	if err := m.enter(ctx, true); err != nil {
		return err
	}
	defer m.busy.Store(false)
	p := m.implementation
	if m.closed.Load() {
		return nil
	}
	p.owner.SetWake(nil)
	p.owner.Close() // explicit force-close cancels every domain lifetime before native destruction.
	var err error
	for _, w := range p.windowList() {
		err = errors.Join(err, p.retire(ctx, w))
	}
	if err == nil {
		m.closed.Store(true)
		p.workbenchMode.active.Store(false)
		p.owner.SetServices(nil)
		manualOwners.Lock()
		delete(manualOwners.owners, p.owner)
		if manualOwners.active == m {
			manualOwners.active = nil
		}
		manualOwners.Unlock()
	}
	return errors.Join(err, p.error())
}

func cloneManualGrids(grids []GridInspection) []GridInspection {
	out := slices.Clone(grids)
	for i := range out {
		out[i].Cells = slices.Clone(out[i].Cells)
		out[i].Headers = slices.Clone(out[i].Headers)
		if out[i].Editor != nil {
			editor := *out[i].Editor
			out[i].Editor = &editor
		}
	}
	return out
}
func cloneManualTrees(trees []TreeInspection) []TreeInspection {
	out := slices.Clone(trees)
	for i := range out {
		out[i].Rows = slices.Clone(out[i].Rows)
		out[i].Pages = slices.Clone(out[i].Pages)
	}
	return out
}

func cloneManualCanvases(canvases []CanvasInspection) []CanvasInspection {
	out := slices.Clone(canvases)
	for i := range out {
		out[i].Objects = slices.Clone(out[i].Objects)
	}
	return out
}
