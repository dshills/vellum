package cogent

import "cogentcore.org/core/system/composer"

// MaxFrameSourceRecords bounds source observations across every intermediate
// composition and the final armed composition of one traced interaction.
const MaxFrameSourceRecords = 32

// MaxSourceFrameMetricsCapacity bounds diagnostic retention independently of
// the larger capacity allowed by ordinary frame metrics.
const MaxSourceFrameMetricsCapacity = 4096

const (
	FrameSourceOther uint8 = iota
	FrameSourceScene
	FrameSourceSprites
	FrameSourceScrim
	FrameSourceInsets
)

// FrameSceneHostSample contains bounded scalar diagnostics from one GPU scene
// source draw. Capture/Encode/Lease are disjoint subsets of DrawNS; Resources
// is inside Lease and Input inside Resources. These elapsed host intervals
// never add to TotalUIWork and do not measure GPU execution or presentation.
// Input includes buffer retirement/admission/allocation and checked input
// queue writes, excluding seed-image writes. LogicalInputBytes counts attempted
// API input bytes, including errors, not physical or successfully transferred
// bytes. InputWriteErrors covers checked input writes only; it does not change
// FrameSourceDetails.UploadFailuresAvailable for the legacy image API.
// DrawAccepted means the scene lease admitted the composition; a later lease
// refusal can coexist with QueueSubmitAttempts and CPU fallback.
type FrameSceneHostSample struct {
	Available                                               bool
	CaptureAttempted, EncodeAttempted, LeaseAttempted       bool
	ResourcesEntered, InputAttempted                        bool
	CaptureNS, EncodeNS, LeaseNS, ResourcesNS, InputNS      int64
	InputWriteAttempts, InputWriteErrors, LogicalInputBytes uint64
	QueueSubmitAttempts                                     uint64
	DrawAccepted, Reused, Fallback                          bool
}

// FrameSourceSample is a scalar observation of one composition source. Nested
// intervals are subsets of DrawNS and RasterUploadNS, never additional UI work.
// SetFromGoImageNS includes conversion/allocation/queue API/sampler CPU work.
// LogicalUploadRGBABytes counts attempted image dimensions at four bytes/pixel,
// not successful texture writes, physical transfer bytes or GPU execution.
type FrameSourceSample struct {
	Composition, Source                    uint32
	Kind                                   uint8
	DrawNS, RasterNS, CopyNS               int64
	RasterAvailable, CopyAvailable         bool
	ImageWidth, ImageHeight                int
	ImageChanged                           bool
	ImageUploadTimingAvailable             bool
	SetFromGoImageNS                       int64
	SetFromGoImageAttempts, ImageCacheHits uint64
	LogicalUploadRGBABytes                 uint64
	SceneHost                              FrameSceneHostSample
}

// FrameSourceDetails is available only through the explicit diagnostic
// constructor. Source order is local to its Composition ordinal. Records retain
// no source/widget/image pointers or user text. Legacy image-upload failures are
// unavailable because that image API logs errors without returning them. Checked
// GPU scene input-write errors are recorded separately in SceneHost.
type FrameSourceDetails struct {
	Compositions            uint32
	UploadFailuresAvailable bool
	Records                 []FrameSourceSample
}

func cloneFrameSourceDetails(details *FrameSourceDetails) *FrameSourceDetails {
	if details == nil {
		return nil
	}
	copy := *details
	copy.Records = append([]FrameSourceSample(nil), details.Records...)
	return &copy
}

func frameSourceDetails(details *composer.SourceDetails) *FrameSourceDetails {
	if details == nil {
		return nil
	}
	out := &FrameSourceDetails{Compositions: details.Compositions, Records: make([]FrameSourceSample, len(details.Records))}
	for i, record := range details.Records {
		out.Records[i] = FrameSourceSample{
			Composition: record.Composition, Source: record.Source, Kind: record.Kind,
			DrawNS: record.DrawNS, RasterNS: record.RasterNS, CopyNS: record.CopyNS,
			RasterAvailable: record.RasterAvailable, CopyAvailable: record.CopyAvailable,
			ImageWidth: record.ImageWidth, ImageHeight: record.ImageHeight, ImageChanged: record.ImageChanged,
			ImageUploadTimingAvailable: record.ImageUploadTimingAvailable, SetFromGoImageNS: record.SetFromGoImageNS,
			SetFromGoImageAttempts: record.SetFromGoImageAttempts, ImageCacheHits: record.ImageCacheHits,
			LogicalUploadRGBABytes: record.LogicalUploadRGBABytes,
			SceneHost: FrameSceneHostSample{
				Available:        record.SceneHost.Available,
				CaptureAttempted: record.SceneHost.CaptureAttempted, EncodeAttempted: record.SceneHost.EncodeAttempted,
				LeaseAttempted: record.SceneHost.LeaseAttempted, ResourcesEntered: record.SceneHost.ResourcesEntered,
				InputAttempted: record.SceneHost.InputAttempted,
				CaptureNS:      record.SceneHost.CaptureNS, EncodeNS: record.SceneHost.EncodeNS,
				LeaseNS: record.SceneHost.LeaseNS, ResourcesNS: record.SceneHost.ResourcesNS, InputNS: record.SceneHost.InputNS,
				InputWriteAttempts: record.SceneHost.InputWriteAttempts, InputWriteErrors: record.SceneHost.InputWriteErrors,
				LogicalInputBytes: record.SceneHost.LogicalInputBytes, QueueSubmitAttempts: record.SceneHost.QueueSubmitAttempts,
				DrawAccepted: record.SceneHost.DrawAccepted, Reused: record.SceneHost.Reused, Fallback: record.SceneHost.Fallback,
			},
		}
	}
	return out
}
