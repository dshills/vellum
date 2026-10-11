package textinput

import "testing"

func TestCompositionEncoding(t *testing.T) {
	u := Update{Kind: "mark", Text: "A👩‍💻éZ", Anchor: 7, Focus: 8, MarkedStart: 6, MarkedEnd: 9}
	r, err := Decode(u)
	if err != nil {
		t.Fatal(err)
	}
	if r.RuneAnchor != 5 || r.RuneFocus != 6 || r.RuneMarkedStart != 4 || r.RuneMarkedEnd != 7 {
		t.Fatalf("%+v", r)
	}
	// Allow scalar positions inside combining/ZWJ graphemes during preedit.
	for i := 0; i <= len([]rune(u.Text)); i++ {
		v, ok := UTF16Offset(u.Text, i)
		if !ok {
			t.Fatal(i)
		}
		d, err := Decode(Update{Text: u.Text, Anchor: v, Focus: v, MarkedStart: -1, MarkedEnd: -1})
		if err != nil || d.RuneFocus != i {
			t.Fatal(i, err)
		}
	}
}
func TestRejectNativeRanges(t *testing.T) {
	base := Update{Text: "A👩Z", Anchor: 1, Focus: 3, MarkedStart: -1, MarkedEnd: -1}
	for _, u := range []Update{
		{Text: base.Text, Anchor: 2, Focus: 3, MarkedStart: -1, MarkedEnd: -1},
		{Text: base.Text, Anchor: 1, Focus: 9, MarkedStart: -1, MarkedEnd: -1},
		{Text: base.Text, Anchor: 0, Focus: 0, MarkedStart: 1, MarkedEnd: 3},
		{Text: base.Text, Anchor: 1, Focus: 1, MarkedStart: -1, MarkedEnd: 1},
		{Text: "a\x00b", MarkedStart: -1, MarkedEnd: -1},
		{Text: string([]byte{0xff}), MarkedStart: -1, MarkedEnd: -1},
	} {
		if _, err := Decode(u); err == nil {
			t.Fatalf("accepted %+v", u)
		}
	}
	if _, ok := UTF16Offset(base.Text, 4); ok {
		t.Fatal("out of range")
	}
}

func TestNativeClientShadowEncoding(t *testing.T) {
	for _, before := range []string{"a\x00b", string([]byte{0xff})} {
		if _, err := Decode(Update{ClientID: 40, BeforeText: before, Text: "A\n", MarkedStart: -1, MarkedEnd: -1}); err == nil {
			t.Fatal("invalid shadow decoded")
		}
	}
	u := Update{Sequence: 9, ClientID: 40, BeforeText: "👩\n", Text: "日本👩\n", Anchor: 2, Focus: 2, MarkedStart: 0, MarkedEnd: 2}
	r, err := Decode(u)
	if err != nil || r.ClientID != 40 || r.BeforeText != u.BeforeText {
		t.Fatal("client identity/shadow lost", err, r)
	}
}
