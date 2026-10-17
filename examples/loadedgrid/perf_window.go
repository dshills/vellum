package main

import (
	"errors"
	"image"
	"math"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/system"
)

const perfWindowTitle = "Vellum stored-grid performance"
const perfWindowWidthDP, perfWindowHeightDP = 1500, 1250

// Endpoint observations are outside recorded frames. They do not prove there
// was no transient resize between endpoints or physical display delivery.
type perfWindowGeometry struct {
	RequestedDP      [2]float32       `json:"requested_dp"`
	ClientPixels     image.Point      `json:"client_pixels"`
	ClientUnits      image.Point      `json:"client_units"`
	Render           math32.Geom2DInt `json:"render"`
	WindowLogicalDPI float32          `json:"window_logical_dpi"`
	ScreenNumber     int              `json:"screen_number"`
	ScreenName       string           `json:"screen_name"`
	ScreenPixels     image.Point      `json:"screen_pixels"`
	ScreenGeometry   image.Rectangle  `json:"screen_geometry"`
	ScreenLogicalDPI float32          `json:"screen_logical_dpi"`
	DPR              float32          `json:"dpr"`
}

func observePerfWindow(application system.App) (value perfWindowGeometry, err error) {
	queue, ok := application.(interface{ RunOnMainQueued(func()) })
	if !ok {
		return value, errors.New("loadedgrid perf: native queued geometry observation unsupported")
	}
	executed := false
	// A worker must wait for queue publication; the inline startup fallback is
	// unsafe for native queries. Never invoke while holding the UI owner.
	queue.RunOnMainQueued(func() {
		executed = true
		if application.NWindows() != 1 {
			err = errors.New("loadedgrid perf: geometry requires exactly one window")
			return
		}
		window := application.Window(0)
		if window == nil || window.Name() != perfWindowTitle {
			err = errors.New("loadedgrid perf: wrong native window")
			return
		}
		screen := window.Screen()
		if screen == nil {
			err = errors.New("loadedgrid perf: no native screen")
			return
		}
		dpi := window.LogicalDPI()
		if !window.Lock() {
			err = errors.New("loadedgrid perf: native window closed")
			return
		}
		defer window.Unlock()
		value = perfWindowGeometry{RequestedDP: [2]float32{perfWindowWidthDP, perfWindowHeightDP}, ClientPixels: window.Size(), ClientUnits: window.WinSize(), Render: window.RenderGeom(), WindowLogicalDPI: dpi, ScreenNumber: screen.ScreenNumber, ScreenName: screen.Name, ScreenPixels: screen.PixelSize, ScreenGeometry: screen.Geometry, ScreenLogicalDPI: screen.LogicalDPI, DPR: screen.DevicePixelRatio}
	})
	if !executed {
		return perfWindowGeometry{}, errors.New("loadedgrid perf: native geometry queue stopped")
	}
	if err != nil {
		return perfWindowGeometry{}, err
	}
	if err = checkPerfWindow(value); err != nil {
		return perfWindowGeometry{}, err
	}
	return value, nil
}
func checkPerfWindow(value perfWindowGeometry) error {
	finitePositive := func(v float32) bool { return v > 0 && !math.IsInf(float64(v), 0) }
	if value.RequestedDP != [2]float32{perfWindowWidthDP, perfWindowHeightDP} || value.ClientPixels.X <= 0 || value.ClientPixels.Y <= 0 || value.ClientUnits.X <= 0 || value.ClientUnits.Y <= 0 || value.Render.Size.X <= 0 || value.Render.Size.Y <= 0 || value.ScreenPixels.X <= 0 || value.ScreenPixels.Y <= 0 || value.ScreenGeometry.Empty() || value.ScreenName == "" || value.ScreenNumber < 0 || !finitePositive(value.WindowLogicalDPI) || !finitePositive(value.ScreenLogicalDPI) || !finitePositive(value.DPR) {
		return errors.New("loadedgrid perf: invalid native window geometry")
	}
	render := image.Rectangle{Min: value.Render.Pos, Max: value.Render.Pos.Add(value.Render.Size)}
	if !render.In(image.Rectangle{Max: value.ClientPixels}) {
		return errors.New("loadedgrid perf: render area outside native client")
	}
	return nil
}
