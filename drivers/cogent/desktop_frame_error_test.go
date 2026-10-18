package cogent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestDesktopCompositionErrorIdentityAndLegacyText(t *testing.T) {
	for _, unavailable := range []bool{false, true} {
		original := &DesktopCompositionError{Failure: "occluded-like text", Unavailable: unavailable}
		if original.Error() != "cogent: desktop composition: occluded-like text" {
			t.Fatal("existing error text changed", original)
		}
		wrapped := fmt.Errorf("consumer: %w", original)
		var result *DesktopCompositionError
		if !errors.As(wrapped, &result) || result != original || errors.Is(wrapped, ErrDesktopFrameUnavailable) != unavailable || errors.Is(wrapped, ErrDesktopClosed) {
			t.Fatal("typed error lost/guessed identity", wrapped)
		}
	}
	var empty *DesktopCompositionError
	if empty.Is(ErrDesktopFrameUnavailable) || empty.Error() == "" {
		t.Fatal("nil error method contract")
	}
	// The added false scalar preserves existing JSON payload shape.
	wire, err := json.Marshal(FrameSample{})
	if err != nil || strings.Contains(string(wire), "Unavailable") {
		t.Fatal("zero metadata changed legacy wire", string(wire), err)
	}
}
