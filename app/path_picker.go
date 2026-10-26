package app

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/ui"
	"strings"
	"unicode/utf8"
)

type PathOptions struct {
	Kind               ui.FilePickerKind
	Title, InitialPath string
	Extensions         []string
}

func (o PathOptions) Validate() error {
	if e := ui.ValidatePickerFilters(o.Kind, o.Extensions); e != nil {
		return e
	}
	if len(o.Title) > 256 || !utf8.ValidString(o.Title) || strings.ContainsRune(o.Title, 0) {
		return fmt.Errorf("vellum: invalid picker title")
	}
	if o.InitialPath != "" {
		if len(o.InitialPath) > 4096 {
			return ErrServiceTooLarge
		}
		if _, e := ui.CopyFileDropPaths([]string{o.InitialPath}); e != nil {
			return e
		}
	}
	return nil
}

// PathServices is optional. Worker-side ChoosePath is exact-parent/cancelable;
// it checks regular files/directories. RevealPath requests Finder selection,
// never edits files. Metadata syscalls cannot be interrupted mid-call; a Finder
// request cannot be rolled back after dispatch. Unsupported hosts refuse.
type PathServices interface {
	ChoosePath(context.Context, WindowRef, PathOptions) (string, error)
	RevealPath(context.Context, string) error
}

func (a *Application) Paths() (PathServices, error) {
	s, e := a.Services()
	if e != nil {
		return nil, e
	}
	p, ok := s.(PathServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return p, nil
}
