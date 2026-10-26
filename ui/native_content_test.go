package ui

import (
	"strings"
	"testing"
)

func TestPDFViewerAdmission(t *testing.T) {
	p := PDFViewer{ID: "pdf", Label: "Reference", Path: "/tmp/document.pdf", ShowThumbnails: true}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	c := n.Children[1]
	if c.Kind != NativeContent || !c.NativeContent.Thumbnails {
		t.Fatal(c)
	}
	// Every View detaches its driver metadata from caller mutation.
	p.Path = "/tmp/other.pdf"
	if c.NativeContent.Path == p.Path {
		t.Fatal("aliased")
	}
	for _, path := range []string{"relative.pdf", "https://example.com/a.pdf", "/tmp/a\x00.pdf", "/tmp/" + strings.Repeat("a", 4097)} {
		p.Path = path
		if _, e = p.View(); e == nil {
			t.Fatal(path)
		}
	}
	p.Path = ""
	if n, e = p.View(); e != nil || n.Children[1].Kind != Label {
		t.Fatal(e)
	}
	if Validate(Node{ID: "forged", Kind: NativeContent, NativeContent: &NativeContentInput{Type: "pdf", Path: "/tmp/a.pdf"}}) == nil {
		t.Fatal("forged metadata")
	}
	if Validate(Node{ID: "wrong", Kind: Label, NativeContent: c.NativeContent}) == nil {
		t.Fatal("wrong kind")
	}
}
