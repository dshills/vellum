package cogent

import (
	"errors"
	"image"
	"sync/atomic"
	"time"

	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

var (
	ErrManualConfig      = errors.New("cogent: invalid manual configuration or copied adapter")
	ErrManualUnsupported = errors.New("cogent: manual routing requires the offscreen driver")
	ErrManualBusy        = errors.New("cogent: concurrent or reentrant manual operation")
	ErrManualClosed      = errors.New("cogent: manual adapter or window closed")
	ErrManualStale       = errors.New("cogent: stale manual window or control incarnation")
	ErrManualLayout      = errors.New("cogent: target requires a current explicit frame")
	ErrManualDisabled    = errors.New("cogent: target is disabled or hidden by a modal")
)

// ManualFont is checked, detached font data. Name is a slash-separated local
// font filename; the digest is required. No ambient system font lookup occurs.
type ManualFont struct {
	Name, SHA256 string
	Data         []byte
}
type ManualConfig struct {
	Clock                                 *execution.ManualClock
	Executor                              *execution.Manual
	Fonts                                 []ManualFont
	Theme                                 ui.Theme
	Commands                              *command.Registry
	LogicalDPI                            float32
	EventCapacity, OwnerBatch, MaxWindows int
	MaxControls                           int // zero uses 4096; each window descriptor is checked before allocation.
}

// ManualInput either supplies physical coordinates or resolves a current target
// to its measured center. Keys still route through actual focus; targets never
// invoke a descriptor callback. Window incarnation is always required.
type ManualInput struct {
	Window            string
	WindowIncarnation uint64
	Target            string
	TargetIncarnation uint64
	Type              events.Types
	Rune              rune
	Code              key.Codes
	Modifiers         key.Modifiers
	Button            events.Buttons
	Position          image.Point
	Delta             math32.Vector2
}

type ManualElement struct {
	ID, Parent                          string
	Incarnation                         uint64
	Kind                                ui.Kind
	Name                                string
	Value                               any
	Enabled, Focused, Visible, Selected bool
	Selection                           [2]int // UTF-16 range; focus direction is preserved separately.
	TextFocus                           int
	Bounds                              image.Rectangle // actual scene pixels, including dialog scene origin.
}
type ManualWindowSnapshot struct {
	ID                                                      string
	Incarnation                                             uint64
	Elements                                                []ManualElement
	FocusedID                                               string
	Frames                                                  uint64
	Now, NextPaint                                          time.Time
	Events, Timers                                          int
	Blocked, ReadyPaint, FuturePaint, Closed, LayoutPending bool
	Failure                                                 error
	DomainClosed, RetirementPending                         bool
}
type ManualSnapshot struct {
	Application app.PendingState
	Clock       execution.ClockReadiness
	Executor    execution.Snapshot
	Windows     []ManualWindowSnapshot // immutable application registration order.
	Closed      bool
	Failure     error
}
type ManualStepResult struct{ Events, OwnerCallbacks, Frames int }

// Manual owns actual offscreen managed windows. Methods have one dispatch
// owner and refuse concurrency/reentrancy; callbacks may call Application APIs,
// but must not call back into this adapter. Clock and executor remain caller
// owned and are never implicitly advanced/stepped/closed.
type Manual struct {
	self           *Manual
	busy           atomic.Bool
	closed         atomic.Bool
	implementation *manualImplementation
}
