package main

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/system"
	"image"
	"sync/atomic"
	"testing"
)

type geometryTestApp struct {
	system.App
	window system.Window
	count  int
	refuse bool
	onMain atomic.Bool
}

func (a *geometryTestApp) RunOnMainQueued(f func()) {
	if a.refuse {
		return
	}
	done := make(chan struct{})
	go func() { a.onMain.Store(true); f(); a.onMain.Store(false); close(done) }()
	<-done
}
func (a *geometryTestApp) NWindows() int {
	if !a.onMain.Load() {
		panic("off-main count")
	}
	return a.count
}
func (a *geometryTestApp) Window(int) system.Window {
	if !a.onMain.Load() {
		panic("off-main window")
	}
	return a.window
}

type geometryTestWindow struct {
	system.Window
	a       *geometryTestApp
	screen  system.Screen
	closed  bool
	locked  bool
	name    string
	size    image.Point
	unlocks int
}

func (w *geometryTestWindow) Name() string { return w.name }
func (w *geometryTestWindow) Screen() *system.Screen {
	if !w.a.onMain.Load() {
		panic("off-main screen")
	}
	return &w.screen
}
func (w *geometryTestWindow) LogicalDPI() float32 { return 162 }
func (w *geometryTestWindow) Lock() bool {
	if !w.a.onMain.Load() {
		panic("off-main lock")
	}
	if w.closed {
		return false
	}
	w.locked = true
	return true
}
func (w *geometryTestWindow) Unlock() {
	if !w.locked {
		panic("unbalanced unlock")
	}
	w.locked = false
	w.unlocks++
}
func (w *geometryTestWindow) Size() image.Point {
	if !w.locked {
		panic("unprotected size")
	}
	return w.size
}
func (w *geometryTestWindow) WinSize() image.Point         { return w.Size() }
func (w *geometryTestWindow) RenderGeom() math32.Geom2DInt { return math32.Geom2DInt{Size: w.Size()} }
func geometryFixture() (*geometryTestApp, *geometryTestWindow) {
	a := &geometryTestApp{count: 1}
	w := &geometryTestWindow{a: a, name: perfWindowTitle, size: image.Pt(1519, 1266), screen: system.Screen{Name: "fixture", PixelSize: image.Pt(5120, 1440), Geometry: image.Rect(0, 0, 5120, 1440), LogicalDPI: 162, DevicePixelRatio: 1}}
	a.window = w
	return a, w
}
func TestPerfWindowObservationUsesQueuedMainAndDetachedValues(t *testing.T) {
	a, w := geometryFixture()
	v, err := observePerfWindow(a)
	if err != nil || v.ClientPixels != w.size || v.ClientUnits != w.size || v.WindowLogicalDPI != 162 || w.locked || w.unlocks != 1 {
		t.Fatalf("%+v %v", v, err)
	}
	w.screen.Name = "changed"
	w.screen.PixelSize = image.Pt(1, 1)
	w.size = image.Pt(1, 1)
	if v.ScreenName != "fixture" || v.ClientPixels != image.Pt(1519, 1266) || v.ScreenPixels != image.Pt(5120, 1440) {
		t.Fatal("snapshot retained native state")
	}
}
func TestPerfWindowObservationRefusesStoppedWrongAndClosed(t *testing.T) {
	for _, name := range []string{"stopped", "multiple", "wrong", "closed", "invalid"} {
		t.Run(name, func(t *testing.T) {
			a, w := geometryFixture()
			switch name {
			case "stopped":
				a.refuse = true
			case "multiple":
				a.count = 2
			case "wrong":
				w.name = "other"
			case "closed":
				w.closed = true
			case "invalid":
				w.size = image.Point{}
			}
			v, err := observePerfWindow(a)
			if err == nil || v != (perfWindowGeometry{}) || w.locked {
				t.Fatalf("partial/accepted snapshot: %+v %v", v, err)
			}
		})
	}
}
