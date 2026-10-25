package cogent

import (
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"image/color"
	"sync"
	"testing"
)

func TestSystemAppearanceCoalescesWithoutOrdinaryQueueCredit(t *testing.T) {
	a, e := app.New(app.Config{QueueCapacity: 1})
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	ref, e := a.OpenWindow(app.Window{ID: "one", Title: "Appearance", Width: 400, Height: 300, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }})
	if e != nil {
		t.Fatal(e)
	}
	a.TakeDirty(ref)
	calls := 0
	var seen SystemAppearance
	s := registerSystemAppearance(a, func(v SystemAppearance) { calls++; seen = v })
	defer s.close()
	if e = a.Post(func() {}); e != nil {
		t.Fatal(e)
	}
	v := SystemAppearance{Known: true, Accent: color.NRGBA{R: 100, A: 255}}
	for i := 0; i < 100; i++ {
		v.Dark = i%2 == 0
		s.publish(v)
	}
	if calls != 0 {
		t.Fatal("native callback entered owner")
	}
	if !a.TakeDirty(ref) {
		t.Fatal("full ordinary queue lost redraw")
	}
	if e = a.Do(s.deliver); e != nil {
		t.Fatal(e)
	}
	if calls != 1 || seen != v {
		t.Fatal("did not coalesce latest state", calls, seen)
	}
	a.Do(s.deliver)
	if calls != 1 {
		t.Fatal("duplicate owner delivery")
	}
	s.publish(v)
	if a.TakeDirty(ref) {
		t.Fatal("unchanged sample caused redraw")
	}
	s.close()
	s.publish(SystemAppearance{Known: true, Dark: true, Accent: color.NRGBA{A: 255}})
	a.Do(s.deliver)
	if calls != 1 {
		t.Fatal("retired observation revived callback")
	}
}
func TestSystemAppearanceConcurrentSamplesAndRetirement(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	s := registerSystemAppearance(a, func(SystemAppearance) { t.Error("unrequested owner delivery") })
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.publish(SystemAppearance{Known: true, Dark: j%2 == 0, Accent: color.NRGBA{R: uint8(i), A: 255}})
				_ = s.snapshot()
			}
		}(i)
	}
	s.close()
	wg.Wait()
	// A delayed old native token cannot publish into another observation lifetime.
	newer := registerSystemAppearance(a, nil)
	defer newer.close()
	receiveSystemAppearance(s.id, SystemAppearance{Known: true, Accent: color.NRGBA{A: 255}})
	if newer.snapshot().Known {
		t.Fatal("old token entered new observer")
	}
}
func TestMacSystemThemeContrastMotionAndExplicitSnapshot(t *testing.T) {
	for _, dark := range []bool{false, true} {
		v := SystemAppearance{Known: true, Dark: dark, HighContrast: true, ReduceMotion: true, ReduceTransparency: true, Accent: color.NRGBA{R: 240, G: 170, B: 60, A: 255}}
		theme := MacSystemTheme(v)
		if e := theme.Validate(); e != nil {
			t.Fatal(e)
		}
		if !theme.Clone().ReducedMotion || theme.Palette.Value.Accent != v.Accent {
			t.Fatal("lost preference or accent")
		}
		p := theme.Palette.Value
		if p.Window.A != 255 || p.Popup.A != 255 || p.Line == ui.MacTheme(dark).Palette.Value.Line {
			t.Fatal("contrast/transparency not applied")
		}
		clone := theme.Clone()
		clone.Kinds[ui.Button] = ui.Style{}
		if theme.Kinds[ui.Button] == (ui.Style{}) {
			t.Fatal("theme snapshot aliased")
		}
	}
}
func TestSupportedTitlebarAliasPreservesOriginalContract(t *testing.T) {
	a, old, _ := prototypeFixture(t)
	var supported *TitlebarToolbar = old
	if _, err := prepareTitlebarToolbar(a, supported); err != nil {
		t.Fatal(err)
	}
}

func TestSystemAppearanceRetainsSampleAcrossNoWindowRefresh(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	calls := 0
	s := registerSystemAppearance(a, func(SystemAppearance) { calls++ })
	defer s.close()
	v := SystemAppearance{Known: true, Dark: true, Accent: color.NRGBA{B: 200, A: 255}}
	s.publish(v) // Invalidate reports no registered windows, but the value persists.
	if s.snapshot() != v || calls != 0 {
		t.Fatal("no-window sample was dropped or delivered natively")
	}
	if e := a.Open(app.Window{ID: "later", Title: "Later", Width: 300, Height: 200, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); e != nil {
		t.Fatal(e)
	}
	a.Do(s.deliver)
	if calls != 1 {
		t.Fatal("later owner refresh did not deliver retained value")
	}
}
