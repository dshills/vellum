package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// QuickLookItem names an explicit local file. The service never enumerates folders.
type QuickLookItem struct{ ID, Path, Title string }
type QuickLookOptions struct {
	Items      []QuickLookItem
	SelectedID string
}

func (o QuickLookOptions) Validate() error {
	if len(o.Items) < 1 || len(o.Items) > 16 {
		return fmt.Errorf("vellum: Quick Look requires 1..16 files")
	}
	seen := map[string]bool{}
	found := false
	valid := func(s string, n int) bool {
		return strings.TrimSpace(s) != "" && len(s) <= n && utf8.ValidString(s) && !strings.ContainsRune(s, 0)
	}
	for _, v := range o.Items {
		if !valid(v.ID, 256) || seen[v.ID] || !valid(v.Path, 4096) || !filepath.IsAbs(v.Path) || !valid(v.Title, 256) {
			return fmt.Errorf("vellum: invalid Quick Look item")
		}
		seen[v.ID] = true
		found = found || v.ID == o.SelectedID
	}
	if !found {
		return fmt.Errorf("vellum: selected Quick Look item absent")
	}
	return nil
}

// QuickLookServices presents the actual platform preview. Call on a worker;
// it waits until closed. Close/Space/Escape returns nil; context/parent shutdown
// cancels. Files are read by the OS; no editing or file operations are offered by
// this API. Applications retain their own file-access and restoration policy.
type QuickLookServices interface {
	PreviewFiles(context.Context, WindowRef, QuickLookOptions) error
}

func (a *Application) QuickLook() (QuickLookServices, error) {
	s, err := a.Services()
	if err != nil {
		return nil, err
	}
	p, ok := s.(QuickLookServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return p, nil
}
