//go:build offscreen

package cogent

import (
	"cogentcore.org/core/core"
	"encoding/json"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/workbench"
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

func TestNativeContentWorkbenchRawGeometryUsesSemanticNamespace(t *testing.T) {
	a, b := newAXTest(t, 8)
	workspace, e := workbench.New(workbench.Layout{Version: workbench.Version, Root: &workbench.Group{ID: "group", Tabs: []string{"pdf"}, Active: "pdf"}}, []string{"pdf"})
	if e != nil {
		t.Fatal(e)
	}
	p := ui.PDFViewer{ID: "document", Label: "Reference", Path: "/tmp/synthetic.pdf"}
	w := &Workbench{Owner: a, Window: "ax", Workspace: workspace, Panels: []Panel{{ID: "pdf", Title: "PDF", View: func() ui.Node {
		n, e := p.View()
		if e != nil {
			panic(e)
		}
		return n
	}}}}
	host := core.NewFrame(core.NewBody("Native content dock"))
	root, e := reconcile(host, nil, ui.Node{ID: "dock", Kind: ui.Extension, Custom: w}, a, "ax")
	if e != nil {
		t.Fatal(e)
	}
	defer root.widget.AsTree().Delete()
	nodes, _ := accessibilitySnapshot(root, b)
	bounds := nativeAccessibilityBounds(root)
	found := false
	for _, n := range nodes {
		if n.NativeContent != nil {
			found = true
			if _, ok := bounds["native-frame:"+n.ID]; !ok {
				t.Fatal("raw geometry does not match namespaced source", n.ID, bounds)
			}
		}
	}
	if !found {
		t.Fatal("native source missing")
	}
}

func TestNativeContentTitleGetsWidthBeforeSourceSelection(t *testing.T) {
	for _, path := range []string{"", "/tmp/reference.pdf"} {
		t.Run(path, func(t *testing.T) {
			p := ui.PDFViewer{ID: "pdf", Label: "PDF document", Path: path}
			m, _ := simpleManual(t, manualConfig(t, 1812), func() ui.Node {
				n, e := p.View()
				if e != nil {
					t.Fatal(e)
				}
				return n
			})
			unitSettle(t, m)
			_, title := imageManualElement(t, m, "pdf/label/text")
			if title.Bounds.Dx() < 100 || title.Bounds.Dy() > 35 {
				t.Fatal("title collapsed into a narrow column", title.Bounds)
			}
		})
	}
}
