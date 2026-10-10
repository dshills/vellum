package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles/states"
	"example.com/vellum/foundationintegration/inputbridge"
	"fmt"
	"image"
	"strings"
	"unicode/utf8"
)

// One native window client is rebound only after its prior preedit is drained.
// Payloads carry a stable control ID; widget focus never decides their recipient.
func (f *fixture) nativeClient() (uint64, nativeTextClient, bool) {
	if f.editor.StateIs(states.Focused) {
		return 40, f.editor, f.editor.NativeEditingEnabled()
	}
	if f.field.StateIs(states.Focused) {
		return 8, f.field, true
	}
	if f.nativeClientID == 40 {
		return 40, f.editor, false
	}
	return 8, f.field, false
}

type nativeTextClient interface {
	EditingState() core.TextFieldEditingState
	NativeCompositionRange() (int, int, bool)
	NativeCaretBounds() image.Rectangle
	NativeTextGeometry() core.TextFieldGeometry
	IsReadOnly() bool
}

func (f *fixture) applyNativeSnapshot(u inputbridge.Update, decoded inputbridge.RuneUpdate) bool {
	switch u.ClientID {
	case 8:
		if f.field.EditingState().Text == u.BeforeText {
			return f.field.ApplyNativeEdit(u.Text, decoded.RuneAnchor, decoded.RuneFocus, decoded.RuneMarkedStart, decoded.RuneMarkedEnd, u.Kind)
		}
	case 40:
		return f.editor.ApplyNativeEditFrom(u.BeforeText, u.Text, decoded.RuneAnchor, decoded.RuneFocus, decoded.RuneMarkedStart, decoded.RuneMarkedEnd, u.Kind)
	}
	return false
}

// Unsupported documents relinquish native activation through the same client
// handoff. Keeping the previous active context would route keystrokes incorrectly.
func nativePublicationSupported(text string) bool {
	return utf8.ValidString(text) && !strings.ContainsRune(text, 0) && len(text) <= inputbridge.MaxDocumentBytes
}

// A shadow is published only after Cocoa accepts its matching acknowledgement.
// Accepted queued edits advance it in order. Atomic expected-text comparison
// rejects shared/external mutations anywhere in the document, including outside
// the small delta range.
func (f *fixture) expandNativeUpdate(u inputbridge.Update) (inputbridge.Update, error) {
	if !u.Delta {
		return u, nil
	}
	shadow, ok := f.nativeShadows[u.ClientID]
	if !ok {
		return u, fmt.Errorf("missing native document shadow")
	}
	return inputbridge.ExpandDelta(u, shadow)
}
