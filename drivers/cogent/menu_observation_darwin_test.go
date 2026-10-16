//go:build darwin && cgo

package cogent

import (
	"encoding/json"
	"testing"
)

func TestNativeMenuObservationDistinguishesUncapturedOriginal(t *testing.T) {
	// Pointers detect missing fields as well as false values. Decoding into bool
	// also refuses the integer NSNumber wire format used by boxed C comparisons.
	var records []struct {
		Owned, OriginalCaptured, OriginalRestored, TargetRetired *bool
		Items                                                    []json.RawMessage
	}
	raw := nativeServiceMenuObservationFixture()
	if err := json.Unmarshal([]byte(raw), &records); err != nil || len(records) != 8 {
		t.Fatalf("production menu serializer: records=%d err=%v wire=%s", len(records), err, raw)
	}
	for flags, record := range records {
		if record.Owned == nil || record.OriginalCaptured == nil || record.OriginalRestored == nil || record.TargetRetired == nil || record.Items == nil {
			t.Fatalf("missing observation field in case %d: %s", flags, raw)
		}
		owned, captured, matches := flags&4 != 0, flags&2 != 0, flags&1 != 0
		if *record.Owned != owned || *record.OriginalCaptured != captured || *record.OriginalRestored != (captured && matches) || *record.TargetRetired == owned || len(record.Items) != 0 {
			t.Fatalf("wrong menu observation in case %d: %s", flags, raw)
		}
	}
}
