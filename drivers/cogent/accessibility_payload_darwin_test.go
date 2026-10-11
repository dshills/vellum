//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNativeAccessibilityPayloadCopiesAndReleases(t *testing.T) {
	for _, text := range []string{"A😀é < & >", strings.Repeat("cell · value 😀", 32768)} {
		payload, err := json.Marshal(map[string]any{"text": text, "nested": []any{true, uint64(18446744073709551615), nil}})
		if err != nil {
			t.Fatal(err)
		}
		for _, consume := range []bool{true, false} {
			for range 8 {
				raw := nativeAccessibilityPayloadFixture(string(payload), consume)
				var got struct {
					Alive    bool `json:"alive_while_queued"`
					Copied   bool `json:"copied_bytes"`
					Decoded  bool `json:"decoded"`
					Released bool `json:"released_after_consumer"`
				}
				if err := json.Unmarshal([]byte(raw), &got); err != nil {
					t.Fatal(err, raw)
				}
				if !got.Alive || !got.Released || got.Copied != consume || got.Decoded != consume {
					t.Fatalf("owned deferred payload (bytes=%d, consume=%v): %s", len(payload), consume, raw)
				}
			}
		}
	}
}
