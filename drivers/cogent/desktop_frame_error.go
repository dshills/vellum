package cogent

import "errors"

// ErrDesktopFrameUnavailable identifies a checked surface-acquisition yield.
// It is not a completed frame, a drawable-ready notification or a retry promise.
var ErrDesktopFrameUnavailable = errors.New("cogent: desktop drawable unavailable")

// DesktopCompositionError is a detached failed composition result. Failure keeps
// the existing backend diagnostic text; native error identities already reduced
// to FrameTrace strings are not restored. Unavailable is checked metadata, never
// inferred from Failure text. A failed request publishes no frame capability.
type DesktopCompositionError struct {
	Failure     string
	Unavailable bool `json:",omitempty"`
}

func (e *DesktopCompositionError) Error() string {
	if e == nil {
		return "cogent: desktop composition failed"
	}
	return "cogent: desktop composition: " + e.Failure
}

func (e *DesktopCompositionError) Is(target error) bool {
	return e != nil && e.Unavailable && target == ErrDesktopFrameUnavailable
}
