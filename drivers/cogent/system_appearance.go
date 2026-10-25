package cogent

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image/color"
	"sync"
	"sync/atomic"
)

// SystemAppearance is a detached macOS preference observation. It never changes
// system settings. Callbacks run on the application owner, not AppKit main.
type SystemAppearance struct {
	Known, Dark, HighContrast, ReduceMotion, ReduceTransparency bool
	Accent                                                      color.NRGBA
}

// MacSystemTheme makes a fresh Mac preset. Explicit node/scoped themes still win.
// ReduceTransparency is satisfied by the preset's opaque semantic surfaces.
func MacSystemTheme(v SystemAppearance) ui.Theme {
	t := ui.MacThemeWithAccent(v.Dark, v.Accent)
	t.ReducedMotion = v.ReduceMotion
	if v.HighContrast {
		p := t.Palette.Value
		if v.Dark {
			p.Ink = color.NRGBA{255, 255, 255, 255}
			p.Secondary = p.Ink
			p.Line = color.NRGBA{200, 200, 205, 255}
		} else {
			p.Ink = color.NRGBA{0, 0, 0, 255}
			p.Secondary = p.Ink
			p.Line = color.NRGBA{65, 65, 70, 255}
		}
		t.Palette = ui.Some(p)
		t.Default.Foreground = ui.Some(p.Ink)
		for k, s := range t.Kinds {
			if s.Foreground.Set {
				if s.Background.Set && s.Background.Value == p.Accent {
					s.Foreground = ui.Some(p.AccentInk)
				} else {
					s.Foreground = ui.Some(p.Ink)
				}
			}
			if s.Border.Set {
				b := s.Border.Value
				b.Color = p.Line
				s.Border = ui.Some(b)
			}
			t.Kinds[k] = s
		}
	}
	return t
}

type systemAppearanceState struct {
	mu        sync.Mutex
	owner     *app.Application
	callback  func(SystemAppearance)
	value     SystemAppearance
	closed    bool
	delivered SystemAppearance
	id        uint64
}

var appearanceSequence atomic.Uint64
var appearanceRegistry struct {
	sync.Mutex
	current *systemAppearanceState
}

func registerSystemAppearance(a *app.Application, cb func(SystemAppearance)) *systemAppearanceState {
	s := &systemAppearanceState{owner: a, callback: cb, id: appearanceSequence.Add(1)}
	appearanceRegistry.Lock()
	appearanceRegistry.current = s
	appearanceRegistry.Unlock()
	return s
}
func (s *systemAppearanceState) snapshot() SystemAppearance {
	if s == nil {
		return SystemAppearance{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}
func (s *systemAppearanceState) publish(v SystemAppearance) {
	s.mu.Lock()
	if s.closed || !v.Known || v.Accent.A != 255 || v == s.value {
		s.mu.Unlock()
		return
	}
	s.value = v
	owner := s.owner
	s.mu.Unlock()
	// Reuse the driver's coalesced dirty-window wake, not ordinary queue credits.
	// This cannot lose a preference change when the application's Post queue is full.
	_ = owner.Invalidate("")
}

// deliver runs only inside an existing application-owner refresh. Native
// notifications merely replace one detached value and request the existing wake.
func (s *systemAppearanceState) deliver() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.closed || !s.value.Known || s.value == s.delivered {
		s.mu.Unlock()
		return
	}
	value, cb := s.value, s.callback
	s.delivered = value
	s.mu.Unlock()
	if cb != nil {
		cb(value)
	}
}
func (s *systemAppearanceState) close() {
	if s == nil {
		return
	}
	appearanceRegistry.Lock()
	if appearanceRegistry.current == s {
		appearanceRegistry.current = nil
	}
	appearanceRegistry.Unlock()
	s.mu.Lock()
	s.closed = true
	s.callback = nil
	s.owner = nil
	s.mu.Unlock()
	nativeStopSystemAppearance(s.id)
}
func receiveSystemAppearance(token uint64, v SystemAppearance) {
	appearanceRegistry.Lock()
	s := appearanceRegistry.current
	appearanceRegistry.Unlock()
	if s != nil && s.id == token {
		s.publish(v)
	}
}
