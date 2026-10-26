//go:build darwin && cgo

package cogent

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestNativeURLRecipeWithoutOpening(t *testing.T) {
	cases := []struct {
		value   string
		allowed bool
	}{
		{"https://example.com/docs?q=a%20b", true},
		{"https://[::1]:8080/docs#here", true},
		{"https://example.com/%E2%9C%93", true},
		{"http://localhost:8080/path", true},
		{"file:///tmp/a", false}, {"javascript:alert(1)", false}, {"https://user:pass@example.com", false}, {"https://", false}, {"https://example.com/a b", false}, {"https://example.com/\n", false},
	}
	for _, tc := range cases {
		portable := ui.ValidateLinkURL(tc.value) == nil
		native := nativeURLRecipeFixture(tc.value)
		if portable != tc.allowed || native != tc.allowed {
			t.Fatalf("%q: portable=%v native=%v want=%v", tc.value, portable, native, tc.allowed)
		}
	}
}
