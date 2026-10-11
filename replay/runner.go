package replay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

var (
	ErrBusy       = errors.New("replay: concurrent or reentrant run")
	ErrBudget     = errors.New("replay: step budget exhausted")
	ErrNoProgress = errors.New("replay: backend reports ready work without progress")
	ErrBackend    = errors.New("replay: invalid backend result")
	ErrTime       = errors.New("replay: explicit scheduling time mismatch")
)

// Readiness describes the backend's single explicit scheduling clock. Ready
// counts work that Step can dispatch now. Running and Awaiting keep admitted
// work observable, including a selected worker or blocked origin that cannot
// progress through Step. Future timers alone do not prevent current quiescence.
type Readiness struct {
	Now                      time.Time
	Ready, Running, Awaiting int
	Blocked                  bool
	DueTimers, FutureTimers  int
	NextDeadline             time.Time
}

func (r Readiness) QuiescentNow() bool {
	return r.validate() == nil && r.Ready == 0 && r.Running == 0 && r.Awaiting == 0 && r.DueTimers == 0 && !r.Blocked
}
func (r Readiness) validate() error {
	if r.Now.IsZero() || r.Now.Location() != time.UTC || r.Now.Year() < 1 || r.Now.Year() > 9999 || r.Ready < 0 || r.Running < 0 || r.Awaiting < 0 || r.DueTimers < 0 || r.FutureTimers < 0 || r.Ready > 1<<26 || r.Running > 1<<26 || r.Awaiting > 1<<26 || r.DueTimers > 1<<26 || r.FutureTimers > 1<<26 {
		return ErrBackend
	}
	if !r.NextDeadline.IsZero() && (r.NextDeadline.Location() != time.UTC || r.NextDeadline.Year() < 1 || r.NextDeadline.Year() > 9999) {
		return ErrBackend
	}
	if r.DueTimers > 0 {
		if r.NextDeadline.IsZero() || r.NextDeadline.After(r.Now) {
			return ErrBackend
		}
	} else if r.FutureTimers > 0 {
		if r.NextDeadline.IsZero() || !r.NextDeadline.After(r.Now) {
			return ErrBackend
		}
	} else if !r.NextDeadline.IsZero() {
		return ErrBackend
	}
	return nil
}

// Progress counts actual bounded dispatch units. A paint event can count both
// an event and a frame within one unit; category counts need not sum to Units.
type Progress struct{ Units, Events, Owners, Timers, Frames int }

func (p Progress) validate(budget int) error {
	if p.Units < 0 || p.Units > budget {
		return ErrBackend
	}
	for _, n := range []int{p.Events, p.Owners, p.Timers, p.Frames} {
		if n < 0 || n > p.Units {
			return ErrBackend
		}
	}
	return nil
}

// Observation is detached semantic/model/layout JSON and an optional identity
// of an already completed image. Capture must not request a frame or advance
// time. The concrete backend refuses an unrendered or stale image.
type Observation struct {
	Data        json.RawMessage
	ImageSHA256 string
}

// Backend routes through its actual driver. It must check fixture/font/theme
// identities, target/window lifetimes, expected revisions and operation times.
// Step uses one shared explicit clock and never advances it or selects workers.
// Future-only timers and work awaiting explicit operations are reported rather
// than executed implicitly. Methods must honor ctx and their dispatch budgets.
// The runner borrows the backend; its caller owns backend construction/closure.
type Backend interface {
	CheckManifest(context.Context, Manifest) error
	Apply(context.Context, Operation) error
	Readiness(context.Context) (Readiness, error)
	Step(context.Context, int) (Progress, error)
	Capture(context.Context, Operation) (Observation, error)
}

// RunnerConfig bounds dispatch and retained capture data. MaxCheckpointBytes
// bounds the sum of canonical JSON encodings of retained Checkpoints, including
// each checkpoint's identities. It does not bound an encoding of Result, whose
// wrapper, separators and bounded scalar/readiness metadata add bytes.
type RunnerConfig struct {
	RecordLimits                                     Limits
	MaxSteps, EventBatch                             int
	MaxCaptures, MaxCaptureBytes, MaxCheckpointBytes int
}

func (c RunnerConfig) normalized() (RunnerConfig, error) {
	var err error
	c.RecordLimits, err = c.RecordLimits.normalized()
	if err != nil {
		return c, err
	}
	if c.MaxSteps == 0 {
		c.MaxSteps = 10000
	}
	if c.EventBatch == 0 {
		c.EventBatch = 64
	}
	if c.MaxCaptures == 0 {
		c.MaxCaptures = 128
	}
	if c.MaxCaptureBytes == 0 {
		c.MaxCaptureBytes = 1 << 20
	}
	if c.MaxCheckpointBytes == 0 {
		c.MaxCheckpointBytes = 8 << 20
	}
	if c.MaxSteps < 1 || c.MaxSteps > 1000000 || c.EventBatch < 1 || c.EventBatch > 4096 || c.MaxCaptures < 1 || c.MaxCaptures > 10000 || c.MaxCaptureBytes < 1 || c.MaxCaptureBytes > 8<<20 || c.MaxCheckpointBytes < 1 || c.MaxCheckpointBytes > 64<<20 || c.MaxCaptureBytes > c.MaxCheckpointBytes {
		return c, ErrInvalid
	}
	return c, nil
}

type Checkpoint struct {
	Sequence      uint64
	Label, Window string
	Data          json.RawMessage
	ImageSHA256   string
}

// Residual is the last successfully inspected readiness, including on error.
// It is not a new observation after a canceled or failed backend call.
type Result struct {
	Applied      int
	Progress     Progress
	Residual     Readiness
	QuiescentNow bool
	Checkpoints  []Checkpoint
}
type IdleResult struct {
	Progress     Progress
	Residual     Readiness
	QuiescentNow bool
}

// Runner admits one run at a time; it has no background loop or hidden sleeps.
// A failed run can have applied a prefix. No model/time rollback is implied.
// The caller supplies context deadlines; this runner cannot preempt a callback
// that ignores its context or make an arbitrary backend operation reversible.
// A run completes its operation list even if it reports explicit pending work.
// Do not copy Runner after construction.
type Runner struct {
	config  RunnerConfig
	backend Backend
	self    *Runner
	busy    atomic.Bool
}

func NewRunner(config RunnerConfig, backend Backend) (*Runner, error) {
	c, err := config.normalized()
	if err != nil {
		return nil, err
	}
	if backend == nil {
		return nil, ErrInvalid
	}
	value := reflect.ValueOf(backend)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return nil, ErrInvalid
		}
	}
	r := &Runner{config: c, backend: backend}
	r.self = r
	return r, nil
}
func (r *Runner) enter(ctx context.Context) error {
	if r == nil || r.self != r || ctx == nil {
		return ErrInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !r.busy.CompareAndSwap(false, true) {
		return ErrBusy
	}
	return nil
}
func (r *Runner) readiness(ctx context.Context) (Readiness, error) {
	if err := ctx.Err(); err != nil {
		return Readiness{}, err
	}
	value, err := r.backend.Readiness(ctx)
	if err != nil {
		return value, err
	}
	return value, value.validate()
}
func addProgress(total *Progress, p Progress) {
	total.Units += p.Units
	total.Events += p.Events
	total.Owners += p.Owners
	total.Timers += p.Timers
	total.Frames += p.Frames
}

func (r *Runner) drain(ctx context.Context, budget int) (IdleResult, error) {
	var out IdleResult
	for {
		before, err := r.readiness(ctx)
		if err != nil {
			return out, err
		}
		out.Residual = before
		out.QuiescentNow = before.QuiescentNow()
		if before.Ready == 0 {
			return out, nil
		}
		if out.Progress.Units >= budget {
			return out, ErrBudget
		}
		quantum := min(r.config.EventBatch, budget-out.Progress.Units)
		p, stepErr := r.backend.Step(ctx, quantum)
		if err := p.validate(quantum); err != nil {
			return out, err
		}
		addProgress(&out.Progress, p)
		// Even a partial failed dispatch is represented in the returned progress.
		if stepErr != nil {
			return out, stepErr
		}
		after, err := r.readiness(ctx)
		if err != nil {
			return out, err
		}
		out.Residual = after
		out.QuiescentNow = after.QuiescentNow()
		if !after.Now.Equal(before.Now) {
			return out, ErrTime
		}
		if p.Units == 0 {
			if after.Ready == 0 || after.Blocked || after.Running > 0 || after.Awaiting > 0 {
				return out, nil
			}
			return out, ErrNoProgress
		}
	}
}

// RunUntilIdle stops at current-instant quiescence or admitted work that needs
// another explicit operation/external completion. Inspect QuiescentNow and the
// residual counts; a future deadline is never advanced to make it true.
func (r *Runner) RunUntilIdle(ctx context.Context) (IdleResult, error) {
	if err := r.enter(ctx); err != nil {
		return IdleResult{}, err
	}
	defer r.busy.Store(false)
	return r.drain(ctx, r.config.MaxSteps)
}

// Run validates and detaches the whole record before the backend sees it. Each
// operation consumes one budget unit, followed by bounded currently ready work.
// A worker/clock operation remains explicit; pending work does not prevent later
// recorded cancellation or worker selection from releasing it.
func (r *Runner) Run(ctx context.Context, record Record) (Result, error) {
	var out Result
	if err := r.enter(ctx); err != nil {
		return out, err
	}
	defer r.busy.Store(false)
	if err := record.Validate(r.config.RecordLimits); err != nil {
		return out, err
	}
	frozen := Record{Manifest: cloneManifest(record.Manifest), Operations: make([]Operation, len(record.Operations))}
	for i, o := range record.Operations {
		frozen.Operations[i] = cloneOperation(o)
	}
	initial, err := r.readiness(ctx)
	if err != nil {
		return out, err
	}
	out.Residual, out.QuiescentNow = initial, initial.QuiescentNow()
	if err := r.backend.CheckManifest(ctx, cloneManifest(frozen.Manifest)); err != nil {
		return out, err
	}
	bound, err := r.readiness(ctx)
	if err != nil {
		return out, err
	}
	out.Residual, out.QuiescentNow = bound, bound.QuiescentNow()
	if !bound.Now.Equal(initial.Now) {
		return out, ErrTime
	}
	checkpointBytes := 0
	for _, operation := range frozen.Operations {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		if out.Applied+out.Progress.Units >= r.config.MaxSteps {
			return out, ErrBudget
		}
		before, err := r.readiness(ctx)
		if err != nil {
			return out, err
		}
		out.Residual = before
		out.QuiescentNow = before.QuiescentNow()
		if !before.Now.Equal(frozen.Manifest.Start.Add(operation.At)) {
			return out, fmt.Errorf("%w: operation %d timestamp", ErrTime, operation.Sequence)
		}
		if operation.Kind == Capture {
			if len(out.Checkpoints) >= r.config.MaxCaptures {
				return out, ErrCapacity
			}
			capture, err := r.backend.Capture(ctx, cloneOperation(operation))
			if err != nil {
				return out, err
			}
			if len(capture.Data) > r.config.MaxCaptureBytes {
				return out, ErrCapacity
			}
			if !utf8.Valid(capture.Data) || !json.Valid(capture.Data) || capture.ImageSHA256 != "" && !hash(capture.ImageSHA256) {
				return out, ErrBackend
			}
			// Store the same compact JSON representation used for charging. Keeping
			// raw whitespace while charging its compact encoding could retain far
			// more data than MaxCheckpointBytes admits.
			var compact bytes.Buffer
			if err := json.Compact(&compact, capture.Data); err != nil {
				return out, fmt.Errorf("%w: capture encoding", ErrBackend)
			}
			// Compact may reserve len(raw) capacity. Retain a tight slice so even
			// the backing allocation follows the charged compact representation.
			owned := make(json.RawMessage, compact.Len())
			copy(owned, compact.Bytes())
			point := Checkpoint{Sequence: operation.Sequence, Label: strings.Clone(operation.Label), Window: strings.Clone(operation.Window), Data: owned, ImageSHA256: strings.Clone(capture.ImageSHA256)}
			// Bound canonical checkpoint bytes, including identity/label overhead.
			encoded, err := json.Marshal(point)
			if err != nil {
				return out, fmt.Errorf("%w: capture encoding", ErrBackend)
			}
			if len(encoded) > r.config.MaxCheckpointBytes-checkpointBytes {
				return out, ErrCapacity
			}
			checkpointBytes += len(encoded)
			out.Checkpoints = append(out.Checkpoints, point)
		} else {
			if err := r.backend.Apply(ctx, cloneOperation(operation)); err != nil {
				return out, err
			}
		}
		out.Applied++
		after, err := r.readiness(ctx)
		if err != nil {
			return out, err
		}
		out.Residual = after
		out.QuiescentNow = after.QuiescentNow()
		expected := before.Now
		if operation.Kind == Advance {
			expected = expected.Add(operation.Delta)
		}
		if !after.Now.Equal(expected) {
			return out, ErrTime
		}
		idle, err := r.drain(ctx, r.config.MaxSteps-out.Applied-out.Progress.Units)
		addProgress(&out.Progress, idle.Progress)
		if !idle.Residual.Now.IsZero() {
			out.Residual = idle.Residual
			out.QuiescentNow = idle.QuiescentNow
		}
		if err != nil {
			return out, err
		}
	}
	if len(frozen.Operations) == 0 {
		idle, err := r.drain(ctx, r.config.MaxSteps)
		out.Progress = idle.Progress
		if !idle.Residual.Now.IsZero() {
			out.Residual = idle.Residual
			out.QuiescentNow = idle.QuiescentNow
		}
		return out, err
	}
	return out, nil
}
