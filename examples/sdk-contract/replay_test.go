package sdkcontract_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dshills/vellum/replay"
)

func TestPublicReplayRecordSurvivesSDKRelocation(t *testing.T) {
	m := replay.Manifest{Version: replay.Version, FixtureSHA256: strings.Repeat("a", 64), ThemeSHA256: strings.Repeat("b", 64), LayoutVersion: "sdk-1", RendererVersion: "sdk-1", Start: time.Unix(0, 0).UTC(), Windows: []replay.Window{{ID: "main", Incarnation: 1, Width: 640, Height: 480, DPI: 160}}, Fonts: []replay.Font{{Family: "fixture", SHA256: strings.Repeat("c", 64)}}}
	r, err := replay.NewRecorder(m, replay.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Append(replay.Operation{Kind: replay.Frame, Window: "main", WindowIncarnation: 1}); err != nil {
		t.Fatal(err)
	}
	record, err := r.Snapshot()
	if err != nil || len(record.Operations) != 1 || record.Validate(replay.Limits{}) != nil {
		t.Fatal(record, err)
	}
	if _, err := replay.NewRunner(replay.RunnerConfig{}, nil); !errors.Is(err, replay.ErrInvalid) {
		t.Fatal(err)
	}
}
