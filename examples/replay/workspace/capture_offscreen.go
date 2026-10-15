//go:build offscreen

package workspace

import (
	"github.com/dshills/vellum/drivers/cogent"
	"image"
)

type Capture struct {
	Model   Snapshot
	Windows []WindowCapture
}
type WindowCapture struct {
	ID, FocusedID string
	Bounds        image.Rectangle
	Elements      []cogent.ManualElement
	Grids         []cogent.GridInspection
	Trees         []cogent.TreeInspection
}
