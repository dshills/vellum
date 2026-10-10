package textbridge

import (
	"reflect"
	"strings"
	"testing"
)

func TestNativeEncodingBoundaries(t *testing.T) {
	r, err := New("Aé👩‍💻🇺🇳Z")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.RuneBounds, []int{0, 1, 3, 6, 8, 9}) || !reflect.DeepEqual(r.UTF16Bounds, []int{0, 1, 3, 8, 12, 13}) || !reflect.DeepEqual(r.ByteLengths, []uint8{1, 3, 11, 8, 1}) {
		t.Fatalf("%+v", r)
	}
	for i, boundary := range r.RuneBounds {
		index, ok := r.CharacterAtRune(boundary)
		if !ok || index != i {
			t.Fatal("rune boundary")
		}
		runeOffset, ok := r.RuneAtCharacter(i)
		if !ok || runeOffset != boundary {
			t.Fatal("character boundary")
		}
		index, ok = r.CharacterAtUTF16(r.UTF16Bounds[i])
		if !ok || index != i {
			t.Fatal("UTF16 boundary")
		}
	}
	if _, ok := r.CharacterAtRune(2); ok {
		t.Fatal("combining interior accepted")
	}
	if _, ok := r.CharacterAtUTF16(4); ok {
		t.Fatal("surrogate interior accepted")
	}
	if _, ok := r.RuneAtCharacter(6); ok {
		t.Fatal("out-of-range accepted")
	}
}
func TestRejectInvalidOrOversizedCluster(t *testing.T) {
	for _, text := range []string{string([]byte{0xff}), "a\x00b", "e" + strings.Repeat("\u0301", 128)} {
		if _, err := New(text); err == nil {
			t.Fatal("unrepresentable run accepted")
		}
	}
	r, err := New("")
	if err != nil || len(r.ByteLengths) != 0 || len(r.RuneBounds) != 1 {
		t.Fatal("empty run")
	}
}
