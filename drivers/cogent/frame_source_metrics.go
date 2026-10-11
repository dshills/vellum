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
}

// FrameSourceDetails is available only through the explicit diagnostic
// constructor. Source order is local to its Composition ordinal. Records retain
// no source/widget/image pointers or user text. Upload failures are unavailable
// because the existing backend image API logs errors without returning them.
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
		}
	}
	return out
}
