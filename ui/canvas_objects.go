package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// CanvasObject supplies semantics for an existing Canvas2D scene object. ID is
// its stable scene ID; Name and Help are literal text, independent of paint.
// Geometry and selection always come from the scene, never these descriptions.
type CanvasObject struct{ ID, Name, Help string }

const MaxCanvasObjects = 10000
const MaxCanvasObjectMetadataBytes = 1 << 20

func validateCanvasObjects(n Node) error {
	if len(n.CanvasObjects) == 0 {
		return nil
	}
	if n.Kind != Canvas2D || n.Canvas == nil || len(n.CanvasObjects) > MaxCanvasObjects || len(n.Canvas.Objects) > MaxCanvasObjects {
		return fmt.Errorf("ui: canvas object semantics require Canvas2D and a bounded object set")
	}
	ids := make(map[string]bool, len(n.Canvas.Objects))
	for _, o := range n.Canvas.Objects {
		ids[o.ID] = true
	}
	seen := make(map[string]bool, len(n.CanvasObjects))
	total := 0
	for _, o := range n.CanvasObjects {
		if !ids[o.ID] || seen[o.ID] || o.ID == "" || len(o.ID) > 256 || strings.TrimSpace(o.Name) == "" || len(o.Name) > 1024 || len(o.Help) > 4096 {
			return fmt.Errorf("ui: invalid or duplicate canvas object semantics")
		}
		for _, s := range []string{o.ID, o.Name, o.Help} {
			if !utf8.ValidString(s) || strings.ContainsRune(s, 0) {
				return fmt.Errorf("ui: canvas object metadata requires valid UTF-8 without NUL")
			}
			total += len(s)
		}
		if total > MaxCanvasObjectMetadataBytes {
			return fmt.Errorf("ui: canvas object metadata budget exceeded")
		}
		seen[o.ID] = true
	}
	return nil
}
