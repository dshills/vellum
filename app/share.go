package app

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/ui"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type ShareKind string

const (
	ShareText ShareKind = "text"
	ShareURL  ShareKind = "url"
	ShareFile ShareKind = "file"
)

// ShareItem names explicit content only. Files are local regular-file paths;
// services may read them. The application retains file access/restoration policy.
type ShareItem struct {
	Kind  ShareKind
	Value string
}
type ShareOptions struct {
	AnchorID string
	Items    []ShareItem
}
type ShareResult struct{ Service string }

// Validate checks metadata without file I/O. ShareServices implementations must
// additionally check regular files on their worker before native presentation.
func (o ShareOptions) Validate() error {
	if strings.TrimSpace(o.AnchorID) == "" || len(o.AnchorID) > 256 || !utf8.ValidString(o.AnchorID) || strings.ContainsRune(o.AnchorID, 0) {
		return fmt.Errorf("vellum: share requires a bounded managed button anchor")
	}
	if len(o.Items) < 1 || len(o.Items) > 16 {
		return fmt.Errorf("vellum: share requires 1..16 items")
	}
	total := 0
	for _, item := range o.Items {
		total += len(item.Value)
		if total > 64<<10 {
			return ErrServiceTooLarge
		}
		if strings.TrimSpace(item.Value) == "" || !utf8.ValidString(item.Value) || strings.ContainsRune(item.Value, 0) {
			return fmt.Errorf("vellum: invalid share item")
		}
		switch item.Kind {
		case ShareText:
		case ShareURL:
			if e := ui.ValidateLinkURL(item.Value); e != nil {
				return e
			}
		case ShareFile:
			if !filepath.IsAbs(item.Value) || len(item.Value) > 4096 {
				return fmt.Errorf("vellum: share file requires an absolute local path")
			}
		default:
			return fmt.Errorf("vellum: invalid share kind")
		}
	}
	return nil
}

// ShareServices presents the actual macOS sharing picker at a visible managed
// button in the exact parent window. Call on a worker. The first implementation
// requires macOS 13+. Other hosts return ErrUnavailable. Cancellation and parent
// closure dismiss an uncommitted picker. Success records the user's service
// choice, not delivery; the OS/provider owns any subsequent interaction/transfer.
// Choosing a service commits that OS operation and cannot be rolled back here.
// Implementations must call Validate and reject nonregular file items before
// presentation. Local files can still change after validation; services own reads.
// Filesystem metadata calls are synchronous on that worker and cannot be
// interrupted mid-syscall (including on mounted/network filesystems). Context
// cancellation is checked between calls and before native presentation.
type ShareServices interface {
	Share(context.Context, WindowRef, ShareOptions) (ShareResult, error)
}

func (a *Application) Sharing() (ShareServices, error) {
	s, e := a.Services()
	if e != nil {
		return nil, e
	}
	v, ok := s.(ShareServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return v, nil
}
