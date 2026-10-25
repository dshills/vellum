package ui

import (
	"github.com/dshills/vellum/editor"
	"testing"
)

func TestProseRequiresSoftWrappedEditor(t *testing.T) {
	model, err := editor.New("Draft", editor.Config{})
	if err != nil {
		t.Fatal(err)
	}
	node := Node{ID: "body", Kind: MultilineText, Editor: model, SoftWrap: true, Prose: true}
	if err := Validate(node); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []Node{{ID: "label", Kind: Label, Prose: true, SoftWrap: true}, {ID: "nowrap", Kind: MultilineText, Editor: model, Prose: true}} {
		if Validate(invalid) == nil {
			t.Fatal("invalid prose presentation admitted", invalid.ID)
		}
	}
}
