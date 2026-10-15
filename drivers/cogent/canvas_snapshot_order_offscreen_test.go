//go:build offscreen

package cogent

import (
	"testing"
)

// This draft uses the checkpoint's actual completed offscreen fixture. It does
// not launch a native app or replace the selection callback with a test route.
func TestReviewerCanvasSnapshotDetachesBothLevels(t *testing.T) {
	f := newCanvasAX(t, 7, 160)
	first, err := f.manual.Snapshot(testContext)
	if err != nil || len(first.Windows) != 1 || len(first.Windows[0].Canvases) != 1 || len(first.Windows[0].Canvases[0].Objects) == 0 {
		t.Fatal("missing completed canvas", err)
	}
	originalID := first.Windows[0].Canvases[0].ID
	originalObject := first.Windows[0].Canvases[0].Objects[0]
	first.Windows[0].Canvases[0].ID = "caller-mutated-host"
	first.Windows[0].Canvases[0].Objects[0].Name = "caller-mutated-object"
	first.Windows[0].Canvases[0].Objects[0].Bounds = [4]float32{}
	second, err := f.manual.Snapshot(testContext)
	if err != nil || len(second.Windows[0].Canvases) != 1 {
		t.Fatal("second snapshot", err)
	}
	if got := second.Windows[0].Canvases[0].ID; got != originalID {
		t.Errorf("returned canvas header aliases retained snapshot: %q", got)
	}
	if got := second.Windows[0].Canvases[0].Objects[0]; got != originalObject {
		t.Errorf("returned nested object aliases retained snapshot: got %+v want %+v", got, originalObject)
	}
}

func TestReviewerCanvasAXChildrenPreservePaintOrder(t *testing.T) {
	f := newCanvasAX(t, 7, 160)
	// Names are looked up by ID; their descriptor order must not change the
	// z-order used by Cocoa's reverse child hit test.
	if err := f.owner.Do(func() {
		f.metadata[0], f.metadata[1] = f.metadata[1], f.metadata[0]
	}); err != nil {
		t.Fatal(err)
	}
	f.refresh(t)
	nodes, _ := f.snapshot(t)
	host := accessibilityManagedID("drawing")
	var got []string
	for _, n := range nodes {
		if n.Parent == host {
			got = append(got, n.DomainID)
		}
	}
	if len(got) != 3 || got[0] != "rear:/β" || got[1] != "front" || got[2] != "edge" {
		t.Fatalf("semantic label order changed native paint priority: %v", got)
	}
}
