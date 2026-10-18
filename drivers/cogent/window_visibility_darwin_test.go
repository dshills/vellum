//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"
)

// Foundation notification/weak-reference mechanics only; this creates neither
// NSApplication nor NSWindow and supplies no real visibility/drawable evidence.
func TestWindowVisibilityNativeObserverOwnership(t *testing.T) {
	var got struct {
		Initial, Unrelated, Matching, Sampled, Closed                 int
		Released, Refused, Installed, HandlerCleared, CapturedHandler bool
	}
	wire := nativeVisibilityObserverFixture()
	if err := json.Unmarshal([]byte(wire), &got); err != nil {
		t.Fatal(wire, err)
	}
	if got.Initial != 1 || got.Unrelated != 1 || got.Matching != 2 || got.Sampled != 3 || got.Closed != 3 || !got.Released || !got.Refused || !got.Installed || !got.HandlerCleared || !got.CapturedHandler {
		t.Fatal(wire)
	}
}
