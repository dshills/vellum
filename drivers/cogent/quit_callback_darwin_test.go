//go:build darwin && cgo

package cogent

import "testing"

func TestNativeQuitDelegateInterceptionAndRestoration(t *testing.T) {
	if !testNativeQuitDelegate() {
		t.Fatal("native Quit selector not intercepted/restored")
	}
}
