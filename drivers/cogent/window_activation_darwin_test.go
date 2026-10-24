//go:build darwin && cgo

package cogent

import "testing"

func TestMacAttachmentForwardsWindowFocus(t *testing.T) {
	b := newAccessibilityBridge(nil)
	defer b.closeFrom("activation-attachment-test")
	attachment := &accessibilityAttachment{bridge: b}
	attachment.Focus(false)
	if !b.activation.known || b.activation.focused {
		t.Fatal("lost native inactive sample")
	}
	attachment.Focus(true)
	if !b.activation.known || !b.activation.focused {
		t.Fatal("lost native active sample")
	}
}
