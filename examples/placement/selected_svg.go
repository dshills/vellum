package main

import (
	"errors"
	"path/filepath"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/document"
)

func (m *model) selectedSVGAvailable() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	if len(m.selectedIDs()) == 0 {
		return errors.New("select visible components to export")
	}
	return nil
}

// Capture whole selected units in paint order, independently of the camera and
// editor overlays. Value copies keep a pending file chooser's snapshot frozen.
func (m *model) captureSelectedSVG() ([]svgComponent, error) {
	if err := m.selectedSVGAvailable(); err != nil {
		return nil, err
	}
	ids := m.selectedIDs()
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	image := make([]svgComponent, 0, len(ids))
	for _, o := range m.visibleObjects() {
		if wanted[o.ID] {
			image = append(image, svgComponent{Rect: o, Name: m.layerLabel(o.ID)})
		}
	}
	return image, nil
}

func (m *model) exportSelectedSVG(a *app.Application, ref app.WindowRef, store document.Store) error {
	image, err := m.captureSelectedSVG()
	if err != nil {
		return err
	}
	codec := document.Codec[[]svgComponent]{Capture: func() ([]svgComponent, error) { return image, nil }, Encode: encodeSVG, Decode: func([]byte) ([]svgComponent, error) { return nil, app.ErrUnavailable }, Install: func([]svgComponent) error { return app.ErrUnavailable }}
	name := "Drawing"
	if m.files != nil && m.files.State().Path != "" {
		name = drawingSuggestion(filepath.Base(m.files.State().Path))
	}
	return startDetachedSave(m, a, ref, detachedSaveConfig[[]svgComponent]{title: "Export Selected SVG", filename: name + " Selection", extension: "svg", success: "Exported selected SVG", codec: codec, store: store})
}
