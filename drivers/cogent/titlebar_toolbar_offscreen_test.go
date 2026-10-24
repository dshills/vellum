//go:build offscreen

package cogent

import (
	"context"
	"strings"
	"testing"
)

func TestTitlebarPrototypeOffscreenRefusesBeforeRun(t *testing.T) {
	a, spec, _ := prototypeFixture(t)
	before := processRun.Load()
	err := (Driver{TitlebarToolbarPrototype: spec}).Run(context.Background(), a)
	if err == nil || !strings.Contains(err.Error(), "macOS desktop") {
		t.Fatal("missing explicit backend refusal", err)
	}
	if processRun.Load() != before {
		t.Fatal("refusal consumed process run")
	}
}
