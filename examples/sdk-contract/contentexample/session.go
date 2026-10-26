// Package contentexample is a small public-API consumer, not framework internals.
// Keep Session on one application owner; it performs no persistence or file writes.
package contentexample

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

// Session owns one bounded file operation and the accepted viewer source.
// Close on the owner when retiring the consumer, after native dispatch stops.
// File metadata calls cannot be interrupted mid-syscall; the native services
// handle their own exact-window panel cancellation and decoder lifetimes.
type Session struct {
	Field  ui.FilePickerField
	PDF    ui.PDFViewer
	Media  ui.MediaPlayer
	Status string
	Dark   bool
	a      *app.Application
	window app.WindowRef
	ticket *app.Completion
	cancel context.CancelFunc
	closed bool
}

func New(a *app.Application, w app.WindowRef) (*Session, error) {
	if a == nil {
		return nil, fmt.Errorf("content example: application required")
	}
	actual, e := a.WindowRef(w.ID())
	if e != nil {
		return nil, e
	}
	if actual != w {
		return nil, app.ErrInvalidWindowRef
	}
	s := &Session{a: a, window: w, PDF: ui.PDFViewer{ID: "reference-pdf", Label: "PDF reference", ShowThumbnails: true}, Media: ui.MediaPlayer{ID: "reference-media", Label: "Media reference"}}
	s.Field = ui.FilePickerField{ID: "reference-path", Label: "Local reference", Kind: ui.FilePickerFile, Extensions: []string{"pdf", "wav", "mp4", "mov", "m4a"}, Request: s.request}
	return s, nil
}
func (s *Session) busy() bool { return s.ticket != nil && s.ticket.Err() == nil }
func (s *Session) request(e ui.FilePickerEvent) error {
	if s.closed {
		return app.ErrClosed
	}
	if s.busy() {
		return app.ErrServiceBusy
	}
	if e.ID != s.Field.ID || e.Kind != s.Field.Kind {
		return fmt.Errorf("content example: stale field identity")
	}
	if e.Action == ui.FilePickerClear {
		s.Field.Path = ""
		s.PDF.Path = ""
		s.Media.Path = ""
		s.Status = "Reference cleared"
		return nil
	}
	if e.Action != ui.FilePickerBrowse && e.Action != ui.FilePickerDrop && e.Action != ui.FilePickerReveal {
		return fmt.Errorf("content example: unsupported action")
	}
	extensions := slices.Clone(s.Field.Extensions)
	id, kind := s.Field.ID, s.Field.Kind
	ticket, err := s.a.ReserveWindowCompletion(s.window, nil)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(s.window.Context())
	s.ticket = ticket
	s.cancel = cancel
	s.Status = "Opening reference…"
	// Capture request/configuration values on owner. The worker never reads Field,
	// PDF, Media, Status or other mutable consumer state.
	a, w := s.a, s.window
	go func() {
		defer cancel()
		path := e.Path
		var err error
		if e.Action != ui.FilePickerDrop {
			var services app.PathServices
			services, err = a.Paths()
			if err == nil {
				switch e.Action {
				case ui.FilePickerBrowse:
					path, err = services.ChoosePath(ctx, w, app.PathOptions{Kind: kind, Title: "Choose a local reference", InitialPath: e.Path, Extensions: extensions})
				case ui.FilePickerReveal:
					err = services.RevealPath(ctx, path)
				}
			}
		}
		if err == nil && e.Action != ui.FilePickerReveal {
			err = ui.ValidatePickerPath(path, kind, extensions)
			if err == nil {
				var info os.FileInfo
				info, err = os.Stat(path)
				if err == nil && !info.Mode().IsRegular() {
					err = fmt.Errorf("content example: regular file required")
				}
			}
		}
		if err == nil {
			err = ctx.Err()
		}
		_ = ticket.Publish(func() {
			if s.closed || s.ticket != ticket {
				return
			}
			s.ticket = nil
			s.cancel = nil
			if s.Field.ID != id || s.Field.Kind != kind || !slices.Equal(s.Field.Extensions, extensions) {
				s.Status = "Reference configuration changed"
				_ = a.Invalidate(w.ID())
				return
			}
			switch {
			case errors.Is(err, app.ErrCanceled) || errors.Is(err, context.Canceled):
				s.Status = "Open canceled"
			case err != nil:
				s.Status = err.Error()
			case e.Action == ui.FilePickerReveal:
				s.Status = "Reference revealed"
			default:
				s.Field.Path = path
				s.PDF.Path = ""
				s.Media.Path = ""
				if strings.EqualFold(filepath.Ext(path), ".pdf") {
					s.PDF.Path = path
				} else {
					s.Media.Path = path
				}
				s.Status = "Reference accepted"
			}
			_ = a.Invalidate(w.ID())
		})
	}()
	return nil
}
func (s *Session) View() (ui.Node, error) {
	s.Field.Busy = s.busy()
	s.Field.Disabled = s.closed
	s.PDF.Dark = s.Dark
	s.Media.Dark = s.Dark
	field, e := s.Field.View()
	if e != nil {
		return ui.Node{}, e
	}
	children := []ui.Node{field, {ID: "reference-status", Kind: ui.Label, Text: s.Status}}
	if s.PDF.Path != "" {
		n, e := s.PDF.View()
		if e != nil {
			return ui.Node{}, e
		}
		children = append(children, n)
	}
	if s.Media.Path != "" {
		n, e := s.Media.View()
		if e != nil {
			return ui.Node{}, e
		}
		children = append(children, n)
	}
	n := ui.Node{ID: "reference-session", Kind: ui.Column, Grow: true, Children: children}
	return n, ui.Validate(n)
}
func (s *Session) Close() {
	if s.closed {
		return
	}
	s.closed = true
	if s.cancel != nil {
		s.cancel()
	}
	if s.ticket != nil {
		s.ticket.Abort()
	}
	s.ticket = nil
	s.cancel = nil
	s.Field.Disabled = true
	s.PDF.Path = ""
	s.Media.Path = ""
}
