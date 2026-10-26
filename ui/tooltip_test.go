package ui

import (
	"strings"
	"testing"
)

func TestTooltipPlainHelpAndValidation(t *testing.T) {
	host := Node{ID: "save", Kind: Button, Text: "Save", Activate: func() {}}
	n, e := WithTooltip(host, "Save <all> & close")
	if e != nil || n.Tooltip != "Save <all> & close" || n.Accessibility.Help != n.Tooltip {
		t.Fatal(n, e)
	}
	if len(FocusOrder(n)) != 1 {
		t.Fatal("extra focus stop")
	}
	host.Accessibility.Help = "Long accessible explanation"
	n, e = WithTooltip(host, "Short hint")
	if e != nil || n.Accessibility.Help != host.Accessibility.Help {
		t.Fatal(n, e)
	}
	for _, s := range []string{strings.Repeat("x", 4097), "x\x00y", string([]byte{255})} {
		if _, e = WithTooltip(host, s); e == nil {
			t.Fatal("invalid help admitted")
		}
	}
}
