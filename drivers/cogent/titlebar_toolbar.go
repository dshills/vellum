package cogent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"sync"
	"sync/atomic"
	"unicode"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

// TitlebarToolbar is the supported personal-macOS native toolbar attachment.
// It retains the documented one-standard-window, immutable-catalogue limits.
// Customization requires macOS 15+; original-color icons and autosave are AppKit-owned.
type TitlebarToolbar = TitlebarToolbarPrototype

// TitlebarToolbarPrototype is the compatibility name for the native toolbar, an, immutable macOS attachment for
// one registered window opening. It is deliberately separate from portable
// ui.Toolbar. Configure before Driver.Run; do not mutate during Run. Additional
// windows/reopened lifetimes and ChromeUnified are refused in this prototype.
// AppKit owns placement, overflow, accessibility, and native button appearance.
// A nonempty CustomizationID enables AppKit customization and autosave on macOS
// 15+. An empty ID retains the original noncustomizable, unsaved prototype.
type TitlebarToolbarPrototype struct {
	Window app.WindowRef
	Items  []TitlebarToolbarItem
	// CustomizationID is a stable application/window-kind settings namespace,
	// independent of WindowRef. Use different IDs for independently saved bars.
	// ASCII letters, digits, dot, underscore and hyphen; at most 128 bytes.
	CustomizationID string
}

// TitlebarToolbarItem supports Button, ToolbarSpace and ToolbarFlexibleSpace.
// Buttons require a name, a PNG (1..128 pixels per dimension, at most 64 KiB),
// and an Activate callback. PNG artwork retains its original colors.
// Spaces require only a unique ID and Kind. Callbacks run on the Vellum owner,
// never on AppKit main, and are suppressed during managed modal dialogs.
type TitlebarToolbarItem struct {
	ID       string
	Label    string
	Kind     ui.Kind
	PNG      []byte
	Disabled bool
	Activate func()
}

type titlebarWireItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Kind     string `json:"kind"`
	PNG      []byte `json:"png,omitempty"`
	Disabled bool   `json:"disabled"`
}
type preparedTitlebarToolbar struct {
	ref             app.WindowRef
	customizationID string
	wire            string
	actions         map[string]func()
}

func toolbarPrototypeText(s string) bool {
	if s == "" || len(s) > 128 || !utf8.ValidString(s) {
		return false
	}
	for _, c := range s {
		if unicode.IsControl(c) {
			return false
		}
	}
	return true
}
func prepareTitlebarToolbar(a *app.Application, spec *TitlebarToolbarPrototype) (*preparedTitlebarToolbar, error) {
	if spec == nil {
		return nil, nil
	}
	if _, err := scopedWindowContext(a, spec.Window); err != nil {
		return nil, err
	}
	windows := a.Windows()
	if len(windows) != 1 || windows[0].Ref() != spec.Window || windows[0].Chrome.Style != app.ChromeStandard {
		return nil, fmt.Errorf("cogent: titlebar prototype requires one standard-chrome window opening")
	}
	if len(spec.Items) == 0 || len(spec.Items) > 16 {
		return nil, fmt.Errorf("cogent: titlebar prototype requires 1..16 items")
	}
	if len(spec.CustomizationID) > 128 {
		return nil, fmt.Errorf("cogent: titlebar customization ID is too long")
	}
	for _, c := range spec.CustomizationID {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-') {
			return nil, fmt.Errorf("cogent: invalid titlebar customization ID")
		}
	}
	p := &preparedTitlebarToolbar{ref: spec.Window, customizationID: spec.CustomizationID, actions: map[string]func(){}}
	wire := make([]titlebarWireItem, 0, len(spec.Items))
	seen := map[string]bool{}
	buttons := 0
	for _, item := range spec.Items {
		if !toolbarPrototypeText(item.ID) || seen[item.ID] {
			return nil, fmt.Errorf("cogent: invalid or duplicate titlebar item ID")
		}
		seen[item.ID] = true
		w := titlebarWireItem{ID: item.ID, Label: item.Label, Disabled: item.Disabled}
		switch item.Kind {
		case ui.Button:
			if !toolbarPrototypeText(item.Label) || item.Activate == nil || len(item.PNG) == 0 || len(item.PNG) > 64<<10 {
				return nil, fmt.Errorf("cogent: titlebar button requires label, bounded PNG and action")
			}
			config, err := png.DecodeConfig(bytes.NewReader(item.PNG))
			if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 128 || config.Height > 128 {
				return nil, fmt.Errorf("cogent: titlebar PNG dimensions must be 1..128")
			}
			if _, err = png.Decode(bytes.NewReader(item.PNG)); err != nil {
				return nil, fmt.Errorf("cogent: malformed titlebar PNG: %w", err)
			}
			w.Kind = "button"
			w.PNG = bytes.Clone(item.PNG)
			buttons++
			if !item.Disabled {
				p.actions[item.ID] = item.Activate
			}
		case ui.ToolbarSpace, ui.ToolbarFlexibleSpace:
			if item.Label != "" || len(item.PNG) != 0 || item.Activate != nil || item.Disabled {
				return nil, fmt.Errorf("cogent: titlebar spaces are decoration only")
			}
			w.Kind = "space"
			if item.Kind == ui.ToolbarFlexibleSpace {
				w.Kind = "flexible"
			}
		default:
			return nil, fmt.Errorf("cogent: unsupported titlebar prototype item kind")
		}
		wire = append(wire, w)
	}
	if buttons == 0 {
		return nil, fmt.Errorf("cogent: titlebar prototype requires an action")
	}
	data, err := json.Marshal(wire)
	if err != nil {
		return nil, err
	}
	p.wire = string(data)
	return p, nil
}
func (p *preparedTitlebarToolbar) accepts(ref app.WindowRef) bool { return p == nil || p.ref == ref }

// A new immutable permission identity revokes queued input across modal round
// trips. No UI model is read by the native-main callback.
type titlebarPermission struct{ allowed bool }
type titlebarToolbarBinding struct {
	owner                       *app.Application
	scopeMu                     sync.Mutex
	managedAllowed, customizing bool
	config                      *preparedTitlebarToolbar
	permission                  atomic.Pointer[titlebarPermission]
	closed                      atomic.Bool
}

func newTitlebarToolbarBinding(a *app.Application, p *preparedTitlebarToolbar) *titlebarToolbarBinding {
	b := &titlebarToolbarBinding{owner: a, config: p}
	b.permission.Store(&titlebarPermission{})
	return b
}
func (b *titlebarToolbarBinding) setAllowed(allowed bool) bool {
	if b == nil {
		return false
	}
	b.scopeMu.Lock()
	defer b.scopeMu.Unlock()
	if b.closed.Load() || b.managedAllowed == allowed {
		return false
	}
	b.managedAllowed = allowed
	b.publishPermissionLocked()
	return true
}
func (b *titlebarToolbarBinding) publishPermissionLocked() {
	b.permission.Store(&titlebarPermission{allowed: b.managedAllowed && !b.customizing && !b.closed.Load()})
}

// setCustomizing returns admission, including an already-satisfied request;
// duplicate AppKit/KVO samples must not revoke a current permission again.
func (b *titlebarToolbarBinding) setCustomizing(customizing bool) bool {
	if b == nil {
		return false
	}
	b.scopeMu.Lock()
	defer b.scopeMu.Unlock()
	if b.closed.Load() || (customizing && !b.managedAllowed) {
		return false
	}
	if b.customizing != customizing {
		b.customizing = customizing
		b.publishPermissionLocked()
	}
	return true
}
func (b *titlebarToolbarBinding) itemsChanged() {
	if b == nil {
		return
	}
	b.scopeMu.Lock()
	defer b.scopeMu.Unlock()
	if !b.closed.Load() {
		b.publishPermissionLocked()
	}
}
func (b *titlebarToolbarBinding) close() {
	if b == nil {
		return
	}
	b.scopeMu.Lock()
	defer b.scopeMu.Unlock()
	b.closed.Store(true)
	b.publishPermissionLocked()
}
func (b *titlebarToolbarBinding) deliver(id string) bool {
	if b == nil || b.closed.Load() {
		return false
	}
	permission := b.permission.Load()
	if permission == nil || !permission.allowed || b.config.actions[id] == nil {
		return false
	}
	return b.owner.PostWindow(b.config.ref, func() {
		if b.closed.Load() || b.permission.Load() != permission {
			return
		}
		b.config.actions[id]()
		_ = b.owner.Invalidate(b.config.ref.ID())
	}) == nil
}

// Resolve a fresh registry capability for both native and overflow activation.
func deliverTitlebarToolbarAction(token uint64, id string) bool {
	accessibilityRegistry.Lock()
	b := accessibilityRegistry.bridges[token]
	accessibilityRegistry.Unlock()
	return b != nil && !b.closed.Load() && b.titlebar.Load().deliver(id)
}

// CustomizeTitlebarToolbar asynchronously asks AppKit to show the toolbar and
// its native customization palette. This also recovers a toolbar hidden through
// AppKit's context menu. Call from an owner callback; nil means the request was
// accepted, not that the palette has appeared. Managed/native modal scope and
// lifetime are checked again on native main. No native-main wait is performed.
func CustomizeTitlebarToolbar(ref app.WindowRef) error {
	if ref.Context().Err() != nil {
		return app.ErrWindowClosed
	}
	bridge, err := lookupAccessibility(ref.ID())
	if err != nil {
		return err
	}
	bridge.mu.Lock()
	same := bridge.windowRef == ref
	bridge.mu.Unlock()
	if !same {
		return app.ErrInvalidWindowRef
	}
	binding := bridge.titlebar.Load()
	if binding == nil || binding.closed.Load() || binding.config.customizationID == "" {
		return errors.New("cogent: native titlebar customization unavailable")
	}
	nativeTitlebarCustomize(bridge.token)
	return nil
}
func titlebarBindingForToken(token uint64) *titlebarToolbarBinding {
	accessibilityRegistry.Lock()
	b := accessibilityRegistry.bridges[token]
	accessibilityRegistry.Unlock()
	if b == nil || b.closed.Load() {
		return nil
	}
	return b.titlebar.Load()
}
