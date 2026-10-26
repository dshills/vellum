package cogent

import (
	"cogentcore.org/core/core"
	"errors"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestDropZoneNativeRouteConsumesRejectionWithoutAncestorAdmission(t *testing.T) {
	a := dropApp(t)
	ancestor := 0
	z := &ui.FileDropZone{ID: "zone", Change: func(e ui.FileDropZoneEvent) error { return errors.New("Rejected synthetic file") }}
	zone, err := z.View()
	if err != nil {
		t.Fatal(err)
	}
	host := core.NewFrame(core.NewBody("Synthetic drop zone"))
	r, err := reconcile(host, nil, ui.Node{ID: "root", Kind: ui.Column, FileDrop: func(ui.FileDropEvent) bool { ancestor++; return true }, Children: []ui.Node{zone}}, a, "main")
	if err != nil {
		t.Fatal(err)
	}
	defer disposeControlTree(r)
	dropGeometry(r.widget)
	dropGeometry(r.children[0].widget)
	e := testFileDrop(t)
	r.children[0].widget.AsWidget().HandleEvent(e)
	if !e.IsHandled() || ancestor != 0 || z.Feedback() != "Rejected synthetic file" {
		t.Fatal("drop rejection routing", ancestor, z.Feedback())
	}
}
