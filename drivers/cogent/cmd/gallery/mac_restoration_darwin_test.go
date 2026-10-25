//go:build darwin && cgo

package main

import "testing"

func TestMacWidgetRestorationIsProcessLocal(t *testing.T) {
	if !galleryRestorationFixture() {
		t.Fatal("fresh Cocoa session must preserve argument values without writing persistent preferences")
	}
}
