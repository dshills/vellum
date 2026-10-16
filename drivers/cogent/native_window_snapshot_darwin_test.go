//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestNativeWindowMetadataInertProductionSerializer(t *testing.T) {
	raw := nativeWindowSnapshotFixture()
	var state struct {
		MainThread      bool
		Windows         []nativeServiceWindowSnapshot
		IdentityCases   []nativeServiceWindowSnapshot
		RefusedMetadata []bool
	}
	if err := json.Unmarshal([]byte(raw), &state); err != nil || !state.MainThread || len(state.Windows) != 4 || len(state.IdentityCases) != 4 || len(state.RefusedMetadata) != 6 {
		t.Fatal("production serializer fixture did not decode", raw, err)
	}
	for i, w := range state.Windows {
		if !w.IdentityAvailable || w.WindowClass != "FixtureWindow" || w.ContentViewClass != "FixtureView" || w.WindowNumber != int64(41+i) || w.Title != "fixture" || !w.Visible || w.Miniaturized || !w.Key || w.OcclusionVisible != (i >= 2) || w.Frame != [4]float64{10, 20, 30, 40} {
			t.Fatal("metadata or original state was lost", i, w)
		}
	}
	wantClasses := [][2]string{{"FixtureWindow", ""}, {strings.Repeat("a", 256), strings.Repeat("é", 128)}, {strings.Repeat("é", 128), strings.Repeat("a", 256)}, {"FixtureWindow", ""}}
	wantNumbers := []int64{-1, math.MaxInt64, math.MinInt64, 0}
	for i, w := range state.IdentityCases {
		if !w.IdentityAvailable || w.WindowClass != wantClasses[i][0] || w.ContentViewClass != wantClasses[i][1] || w.WindowNumber != wantNumbers[i] || w.Frame != [4]float64{1, 2, 3, 4} {
			t.Fatal("boundary metadata did not preserve exact values", i, w)
		}
	}
	for i, refused := range state.RefusedMetadata {
		if !refused {
			t.Fatal("invalid or oversized UTF-8 metadata was accepted", i)
		}
	}
}

func TestNativeWindowMetadataLegacyUnavailableAndTypedWire(t *testing.T) {
	var legacy nativeServiceWindowSnapshot
	if err := json.Unmarshal([]byte(`{"Title":"old","Visible":false,"WindowNumber":0}`), &legacy); err != nil || legacy.IdentityAvailable || legacy.WindowClass != "" || legacy.ContentViewClass != "" {
		t.Fatal("old records gained an identity observation", legacy, err)
	}
	for _, raw := range []string{
		`{"IdentityAvailable":1}`, `{"WindowClass":true}`, `{"ContentViewClass":3}`,
		`{"WindowNumber":"42"}`, `{"WindowNumber":1.5}`, `{"WindowNumber":9223372036854775808}`,
	} {
		var sample nativeServiceWindowSnapshot
		if err := json.Unmarshal([]byte(raw), &sample); err == nil {
			t.Fatal("malformed typed metadata accepted", raw, sample)
		}
	}
}
