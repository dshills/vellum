// Package document connects bounded asynchronous persistence to UI-owner history.
package document

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/execution"
	"path/filepath"
	"slices"
	"sync"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/history"
)

var (
	ErrBusy    = errors.New("document: an operation is pending")
	ErrUnsaved = errors.New("document: unsaved changes require an explicit decision")
)

const DefaultMaxBytes int64 = 64 << 20
const MaximumMaxBytes int64 = 256 << 20

type Action string

const (
	Save Action = "save"
	Open Action = "open"
)

// Result describes the completed operation. Written means the file replacement
// occurred, even if cancellation, stale revision, or directory sync then failed.
// Err == nil is the only acknowledged successful save/load.
type Result struct {
	Action  Action
	Path    string
	Written bool
	Err     error
}
type State struct {
	Path                        string
	Busy, Dirty, Closed         bool
	WorkerPending, OwnerPending bool
	Action                      Action
	Last                        Result
}

// Codec transfers a detached immutable snapshot to Encode on a worker. Decode
// runs on that worker; Install consumes its validated result on the UI owner.
// Install must leave the entire model unchanged when returning an error. Its
// panic invalidates the journal. Callbacks must not enter Do or this session.
type Codec[T any] struct {
	Capture func() (T, error)
	Encode  func(T) ([]byte, error)
	Decode  func([]byte) (T, error)
	Install func(T) error
}

type Config[T any] struct {
	Owner      *app.Application
	Window     string
	Journal    *history.Journal
	Codec      Codec[T]
	Path       string
	Filename   string
	Extensions []string
	MaxBytes   int64
	Store      Store
	Executor   execution.Executor
	// Choose overrides the native chooser for independent/headless consumers.
	// It runs on the operation's worker and must honor cancellation.
	Choose func(context.Context, app.FileOptions) (string, error)
	// OnComplete runs on the UI owner after state/history update. It may start
	// another operation or force CloseWindow; it must not nest Do or block.
	OnComplete func(Result)
}

// Session belongs to one UI owner and window lifetime. Save/Open/Cancel/State
// are owner-only. One operation remains busy through queued result admission.
// Window closure cancels work and prevents late model/history/user callbacks.
type Session[T any] struct {
	config    Config[T]
	lifetime  context.Context
	window    app.WindowRef
	path      string
	busy      bool
	action    Action
	cancel    context.CancelFunc
	last      Result
	operation *operationState
}

type SaveOptions struct {
	Path   string
	Choose bool
}
type OpenOptions struct {
	Path           string
	DiscardChanges bool
}

func New[T any](c Config[T]) (*Session[T], error) {
	if c.Owner == nil || c.Journal == nil || c.Codec.Capture == nil || c.Codec.Encode == nil || c.Codec.Decode == nil || c.Codec.Install == nil {
		return nil, fmt.Errorf("document: owner, journal and complete codec required")
	}
	window, err := c.Owner.WindowRef(c.Window)
	if err != nil {
		return nil, err
	}
	if window.Context().Err() != nil {
		return nil, app.ErrWindowClosed
	}
	if c.MaxBytes == 0 {
		c.MaxBytes = DefaultMaxBytes
	}
	if err := validateLimit(c.MaxBytes); err != nil {
		return nil, err
	}
	if c.Path != "" {
		if err := validatePath(c.Path); err != nil {
			return nil, err
		}
	}
	if c.Filename == "" {
		c.Filename = "Untitled"
	}
	c.Extensions = slices.Clone(c.Extensions)
	if err := (app.FileOptions{Filename: c.Filename, Extensions: c.Extensions}).Validate(); err != nil {
		return nil, err
	}
	if c.Executor == nil {
		c.Executor = &execution.Goroutine{}
	}
	if c.Store == nil {
		c.Store = FileStore{}
	}
	return &Session[T]{config: c, lifetime: window.Context(), window: window, path: c.Path}, nil
}

func (s *Session[T]) State() State {
	closed := s.lifetime.Err() != nil
	workerPending, ownerPending := false, false
	if s.operation != nil {
		s.operation.mu.Lock()
		workerPending = !s.operation.workerDone
		ownerPending = !s.operation.ownerDone
		s.operation.mu.Unlock()
	}
	return State{WorkerPending: workerPending, OwnerPending: ownerPending, Path: s.path, Busy: s.busy && !closed, Dirty: s.config.Journal.Snapshot().Dirty, Closed: closed, Action: s.action, Last: s.last}
}
func (s *Session[T]) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Session[T]) ready(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("document: context required")
	}
	if s.lifetime.Err() != nil {
		return app.ErrWindowClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.busy {
		return ErrBusy
	}
	if s.config.Journal.Snapshot().Busy {
		return history.ErrBusy
	}
	return nil
}

type operationState struct {
	mu                    sync.Mutex
	workerDone, ownerDone bool
}

func (s *Session[T]) start(ctx context.Context, action Action, work func(context.Context) (Result, func() error)) error {
	op, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(s.lifetime, cancel)
	state := &operationState{}
	ticket, err := s.config.Owner.ReserveWindowCompletion(s.window, func(error) {
		cancel()
		stop()
		state.mu.Lock()
		state.ownerDone = true
		state.mu.Unlock()
	})
	if err != nil {
		cancel()
		stop()
		return err
	}
	previousAction, previousOperation := s.action, s.operation
	s.busy, s.action, s.cancel, s.operation = true, action, cancel, state
	s.config.Owner.Invalidate(s.config.Window)
	err = s.config.Executor.Submit(func() {
		result := Result{Action: action}
		var commit func() error
		returned := false
		defer func() {
			if p := recover(); p != nil {
				result.Err = &execution.PanicError{Value: p}
				commit = nil
			} else if !returned {
				result.Err = execution.ErrGoexit
				commit = nil
			}
			if op.Err() != nil {
				result.Err = errors.Join(result.Err, op.Err())
				commit = nil
			}
			delivery := ticket.Publish(func() {
				s.busy = false
				s.cancel = nil
				if op.Err() != nil {
					result.Err = errors.Join(result.Err, op.Err())
					commit = nil
				}
				if result.Err == nil && commit != nil {
					result.Err = commit()
				}
				if result.Err == nil {
					s.path = result.Path
				}
				s.last = result
				s.config.Owner.Invalidate(s.config.Window)
				if s.config.OnComplete != nil {
					s.config.OnComplete(result)
				}
			})
			if delivery != nil {
				cancel()
				stop()
			}
			state.mu.Lock()
			state.workerDone = true
			state.mu.Unlock()
		}()
		result, commit = work(op)
		returned = true
	})
	if err != nil {
		ticket.Abort()
		state.mu.Lock()
		state.workerDone = true
		state.mu.Unlock()
		s.busy = false
		s.action = previousAction
		s.cancel = nil
		s.operation = previousOperation
		s.config.Owner.Invalidate(s.config.Window)
		return err
	}
	return nil
}

// Save captures on the owner, encodes/writes on a worker and acknowledges only
// its captured journal revision. Choose always asks for a destination (Save As);
// otherwise Path overrides the current path, with a chooser for an untitled file.
func (s *Session[T]) Save(ctx context.Context, o SaveOptions) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	if o.Choose && o.Path != "" {
		return fmt.Errorf("document: choose and explicit save path are exclusive")
	}
	path := o.Path
	if path != "" {
		if err := validatePath(path); err != nil {
			return err
		}
	}
	if path == "" && !o.Choose {
		path = s.path
	}
	state := s.config.Journal.Snapshot()
	if state.Invalid {
		return history.ErrInvalid
	}
	image, err := s.config.Codec.Capture()
	if err != nil {
		return err
	}
	// Copy all owner state needed by the worker, including current chooser path.
	choosePath := s.path
	choose := s.chooser(choosePath)
	return s.start(ctx, Save, func(op context.Context) (Result, func() error) {
		result := Result{Action: Save, Path: path}
		if result.Path == "" {
			result.Path, result.Err = choose(op, app.SaveFile)
		}
		if result.Err == nil {
			result.Err = validatePath(result.Path)
		}
		if result.Err != nil {
			return result, nil
		}
		data, err := s.config.Codec.Encode(image)
		if err == nil && int64(len(data)) > s.config.MaxBytes {
			err = ErrTooLarge
		}
		if err == nil {
			result.Written, err = s.config.Store.Write(op, result.Path, data, s.config.MaxBytes)
		}
		result.Err = err
		return result, func() error { return s.config.Journal.MarkSavedRevision(state.Revision) }
	})
}

// chooser captures the current path before starting a worker. It never reads
// mutable owner fields such as path/busy/journal from the worker.
func (s *Session[T]) chooser(path string) func(context.Context, app.FileMode) (string, error) {
	c := s.config
	return func(ctx context.Context, mode app.FileMode) (string, error) {
		o := app.FileOptions{Mode: mode, Title: "Open document", Filename: c.Filename, Extensions: slices.Clone(c.Extensions)}
		if mode == app.SaveFile {
			o.Title = "Save document"
		}
		if path != "" {
			o.Directory = filepath.Dir(path)
			o.Filename = filepath.Base(path)
		}
		if c.Choose != nil {
			return c.Choose(ctx, o)
		}
		services, err := c.Owner.Services()
		if err != nil {
			return "", err
		}
		return services.ChooseFile(ctx, o)
	}
}

// Open refuses unsaved changes unless the caller explicitly authorized discard.
// It decodes on a worker and replaces the model/history only if the captured
// revision and invalid flag still match. Failed reads/decode/install preserve it.
func (s *Session[T]) Open(ctx context.Context, o OpenOptions) error {
	if err := s.ready(ctx); err != nil {
		return err
	}
	if o.Path != "" {
		if err := validatePath(o.Path); err != nil {
			return err
		}
	}
	state := s.config.Journal.Snapshot()
	if state.Dirty && !o.DiscardChanges {
		return ErrUnsaved
	}
	choose := s.chooser(s.path)
	return s.start(ctx, Open, func(op context.Context) (Result, func() error) {
		result := Result{Action: Open, Path: o.Path}
		if result.Path == "" {
			result.Path, result.Err = choose(op, app.OpenFile)
		}
		if result.Err == nil {
			result.Err = validatePath(result.Path)
		}
		if result.Err != nil {
			return result, nil
		}
		data, err := s.config.Store.Read(op, result.Path, s.config.MaxBytes)
		if err == nil && int64(len(data)) > s.config.MaxBytes {
			err = ErrTooLarge
		}
		var image T
		if err == nil {
			image, err = s.config.Codec.Decode(data)
		}
		result.Err = err
		return result, func() error {
			current := s.config.Journal.Snapshot()
			if current.Revision != state.Revision || current.Invalid != state.Invalid {
				return history.ErrStale
			}
			return s.config.Journal.Replace(true, func() error { return s.config.Codec.Install(image) })
		}
	})
}
