package textinput

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExpandDeltaGlobalUnicodeRanges(t *testing.T) {
	before := strings.Repeat("line\n", 300000) + "A👩Z\n"
	start, _ := UTF16Offset(before, 1500001)
	u := Update{Delta: true, ReplaceStart: start, ReplaceEnd: start + 2, BeforeText: "👩", Text: "日本", Kind: "insert", Anchor: start + 2, Focus: start + 2, MarkedStart: -1, MarkedEnd: -1}
	next, err := ExpandDelta(u, before)
	if err != nil {
		t.Fatal(err)
	}
	if next.Text != strings.Repeat("line\n", 300000)+"A日本Z\n" {
		t.Fatal("global replacement lost prefix or suffix")
	}
	d, err := Decode(next)
	if err != nil || d.RuneFocus != 1500003 {
		t.Fatal("global selection lost", d.RuneFocus, err)
	}
	u.ReplaceStart++
	if _, err := ExpandDelta(u, before); err == nil {
		t.Fatal("surrogate interior accepted")
	}
	u.ReplaceStart--
	u.BeforeText = "other"
	if _, err := ExpandDelta(u, before); err == nil {
		t.Fatal("expected range mismatch accepted")
	}
}
func TestDeltaBudgets(t *testing.T) {
	u := Update{Delta: true, ReplaceEnd: MaxDeltaFragmentBytes, BeforeText: strings.Repeat("\x01", MaxDeltaFragmentBytes), Text: strings.Repeat("\x02", MaxDeltaFragmentBytes), MarkedStart: -1, MarkedEnd: -1}
	b, _ := json.Marshal(u)
	if len(b) > 50<<20 {
		t.Fatal("escaped full-budget edit exceeds action credit", len(b))
	}
	if _, err := ExpandDelta(u, u.BeforeText); err != nil {
		t.Fatal(err)
	}
	u.Text += "x"
	if _, err := ExpandDelta(u, u.BeforeText); err == nil {
		t.Fatal("oversized document accepted")
	}
	u.Text = ""
	if next, err := ExpandDelta(u, u.BeforeText); err != nil || next.Text != "" {
		t.Fatal("full-budget Cut refused", err)
	}
	u.ReplaceStart = -1
	if _, err := ExpandDelta(u, u.BeforeText); err == nil {
		t.Fatal("invalid range accepted")
	}
}
