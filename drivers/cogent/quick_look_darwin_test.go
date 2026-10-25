//go:build darwin && cgo

package cogent

import "testing"

func TestQuickLookRecipeWithoutPresentation(t *testing.T) {
	valid := `{"Items":[{"ID":"a","Path":"/tmp/a.txt","Title":"A"}],"SelectedID":"a"}`
	if nativeQuickLookRecipeFixture(valid) != "valid" {
		t.Fatal("valid")
	}
	for _, s := range []string{`[]`, `{}`, `{"Items":[{"ID":"a","Path":"https://example.org","Title":"A"}],"SelectedID":"a"}`, `{"Items":[{"ID":"a","Path":"/tmp/a","Title":"A"}],"SelectedID":"missing"}`} {
		if nativeQuickLookRecipeFixture(s) != "" {
			t.Fatal("invalid", s)
		}
	}
}
