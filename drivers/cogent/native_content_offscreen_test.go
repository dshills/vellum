//go:build offscreen

package cogent

import (
	"encoding/json"
	"github.com/dshills/vellum/ui"
	"strings"
	"testing"
)

func TestNativeContentProjectionAndGeometryMembership(t *testing.T) {
	p := &ui.PDFViewer{ID: "pdf", Label: "Reference", Path: "/tmp/synthetic.pdf"}
	m, a := simpleManual(t, manualConfig(t, 1810), func() ui.Node {
		n, e := p.View()
		if e != nil {
			t.Fatal(e)
		}
		return n
	})
	unitSettle(t, m)
	_, target := imageManualElement(t, m, "pdf/content")
	if target.Kind != ui.NativeContent || target.Bounds.Empty() {
		t.Fatal(target)
	}
	// Native surfaces reuse semantic membership, including its separate raw frame.
	b := &accessibilityBridge{}
	n, _ := p.View()
	b.updateGeometryMembership([]accessibilityNode{{ID: "node:pdf/content", NativeContent: n.Children[1].NativeContent}})
	if _, ok := b.geometryMembers["native-frame:node:pdf/content"]; !ok {
		t.Fatal("missing raw frame")
	}
	version := b.geometryVersion
	b.updateGeometryMembership([]accessibilityNode{{ID: "node:pdf/content"}})
	if b.geometryVersion == version {
		t.Fatal("native metadata retirement kept raw membership")
	}
	if _, ok := b.geometryMembers["native-frame:node:pdf/content"]; ok {
		t.Fatal("retired raw frame")
	}
	a.Do(func() { p.Path = ""; a.Invalidate("one") })
	unitSettle(t, m)
	_, target = imageManualElement(t, m, "pdf/content")
	if target.Kind != ui.Label {
		t.Fatal("source removal retained native leaf")
	}
}

func TestNativeContentSemanticDescriptorIsPublishedAndDetached(t *testing.T) {
	p := ui.PDFViewer{ID: "pdf", Label: "Reference", Path: "/tmp/reference.pdf"}
	v, _ := p.View()
	n := accessibilityNode{ID: "pdf", NativeContent: v.Children[1].NativeContent}
	wire, e := accessibilitySemantic([]accessibilityNode{n})
	want, _ := json.Marshal([]accessibilityNode{n})
	if e != nil || wire != string(want) || !strings.Contains(wire, `"native_content"`) {
		t.Fatal("native descriptor omitted", wire, e)
	}
	out := cloneAccessibilityNodes([]accessibilityNode{n})
	n.NativeContent.Path = "/tmp/changed.pdf"
	if out[0].NativeContent.Path == n.NativeContent.Path || accessibilitySameWireNode(out[0], n) {
		t.Fatal("native source aliased or equality ignored it")
	}
}
