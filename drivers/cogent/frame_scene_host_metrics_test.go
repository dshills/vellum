package cogent

import (
	"cogentcore.org/core/system/composer"
	"encoding/json"
	"testing"
)

func TestFrameSceneHostMappingSnapshotAndJSONRemainDetached(t *testing.T) {
	upstream := &composer.SourceDetails{Compositions: 1, Records: []composer.SourceTiming{{Composition: 1, Kind: composer.SourceScene, SceneHost: composer.SceneHostTiming{
		Available: true, CaptureAttempted: true, EncodeAttempted: true, LeaseAttempted: true, ResourcesEntered: true, InputAttempted: true,
		CaptureNS: 11, EncodeNS: 13, LeaseNS: 101, ResourcesNS: 83, InputNS: 47,
		InputWriteAttempts: 5, InputWriteErrors: 1, LogicalInputBytes: 1024, QueueSubmitAttempts: 1,
		DrawAccepted: false, Reused: false, Fallback: true,
	}}}}
	want := FrameSceneHostSample{
		Available: true, CaptureAttempted: true, EncodeAttempted: true, LeaseAttempted: true, ResourcesEntered: true, InputAttempted: true,
		CaptureNS: 11, EncodeNS: 13, LeaseNS: 101, ResourcesNS: 83, InputNS: 47,
		InputWriteAttempts: 5, InputWriteErrors: 1, LogicalInputBytes: 1024, QueueSubmitAttempts: 1,
		DrawAccepted: false, Reused: false, Fallback: true,
	}
	m, err := NewFrameMetricsWithSourceDetails(1)
	if err != nil {
		t.Fatal(err)
	}
	m.samples = append(m.samples, FrameSample{Tag: 1, SourceDetails: frameSourceDetails(upstream)})
	upstream.Records[0].SceneHost = composer.SceneHostTiming{}
	first := m.Snapshot()
	if first.Samples[0].SourceDetails.Records[0].SceneHost != want {
		t.Fatal("GPU host fields were lost or aliased during mapping")
	}
	first.Samples[0].SourceDetails.Records[0].SceneHost.LogicalInputBytes = 0
	second := m.Snapshot()
	got := second.Samples[0].SourceDetails.Records[0].SceneHost
	if got != want {
		t.Fatal("GPU host snapshot aliases a caller")
	}
	raw, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	var decoded FrameMetricsSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Samples[0].SourceDetails.Records[0].SceneHost != want {
		t.Fatal("GPU host JSON lost diagnostics")
	}
}
