package app

import (
	"context"
	"errors"
	"fmt"
)

var ErrInvalidWindowRef = errors.New("vellum: invalid or foreign window reference")

// Nonzero size makes separately allocated issuer identities distinct. Lifetime
// records retain this small identity, not the Application and all its windows.
type windowIssuer struct{ marker byte }
type windowLifetime struct {
	issuer *windowIssuer
	id     string
	ctx    context.Context
}

// WindowRef identifies one opening of a named window. Copies refer to the same
// lifetime. Closing and reopening its name never makes an old reference valid.
// References are comparable and cannot be reconstructed from a serialized name.
type WindowRef struct{ lifetime *windowLifetime }

// ID returns the logical name, or an empty string for the zero reference.
// Names may be reused; ID is not a substitute for lifetime identity.
func (r WindowRef) ID() string {
	if r.lifetime == nil {
		return ""
	}
	return r.lifetime.id
}

var invalidWindowContext = func() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}()

// Context returns the originally captured window context. It stays canceled
// after retirement. The zero reference returns an already-canceled context.
func (r WindowRef) Context() context.Context {
	if r.lifetime == nil {
		return invalidWindowContext
	}
	return r.lifetime.ctx
}

// Ref returns the lifetime captured by an Application.Windows snapshot.
// An unregistered Window declaration has a zero reference.
func (w Window) Ref() WindowRef { return w.reference }

// WindowRef captures the current live opening of a logical name. Capture during
// setup and retain the result for later work; do not resolve names in a delayed
// callback intended for an earlier opening.
func (a *Application) WindowRef(id string) (WindowRef, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return WindowRef{}, ErrClosed
	}
	life := a.lifetimes[id]
	if life == nil {
		return WindowRef{}, fmt.Errorf("%w: no registered window %q", ErrWindowClosed, id)
	}
	return WindowRef{lifetime: life}, nil
}

// Called under mu, atomically with admission/retirement. A stale reference's
// context is retained by its holder rather than by an application history map.
func (a *Application) windowRefErrorLocked(ref WindowRef) error {
	if ref.lifetime == nil || a.issuer == nil || ref.lifetime.issuer != a.issuer {
		return ErrInvalidWindowRef
	}
	if a.lifetimes[ref.ID()] != ref.lifetime || ref.Context().Err() != nil {
		return ErrWindowClosed
	}
	return nil
}
