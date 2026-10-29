package notes

import (
	"errors"
	"github.com/dshills/vellum/document"
	"path/filepath"
	"strings"
)

type externalOpen struct {
	path    string
	initial bool
}

// HandleOpenFiles is owner-only host wiring. This single-notebook example
// explicitly refuses multi-file batches instead of silently replacing each file.
func (f *Files) HandleOpenFiles(paths []string, initial bool) {
	if len(paths) != 1 {
		f.model.accept(errors.New("Open one notebook at a time"))
		_ = f.owner.Invalidate("")
		return
	}
	f.model.accept(f.OpenPath(paths[0], initial))
	_ = f.owner.Invalidate("")
}
func (f *Files) OpenPath(path string, initial bool) error {
	if !filepath.IsAbs(path) || !strings.EqualFold(filepath.Ext(path), ".vellumnotes") || !validText(path, 4096) {
		return ErrInvalid
	}
	if f.quitHeld || f.prompt != "" {
		return document.ErrBusy
	}
	if f.restoring {
		if f.pendingExternal != nil {
			return document.ErrBusy
		}
		f.pendingExternal = &externalOpen{path, initial}
		f.main.Cancel()
		return nil
	}
	if f.workspace != nil && f.workspace.State().Busy && !f.main.State().Busy && !f.recovery.State().Busy {
		if f.pendingExternal != nil {
			return document.ErrBusy
		}
		f.pendingExternal = &externalOpen{path, initial}
		return nil
	}
	if initial && !f.model.History().Dirty && !f.model.pending {
		return f.openNotebook(path)
	}
	return f.model.Request(func() error { return f.openNotebook(path) })
}
func (f *Files) openNotebook(path string) error {
	if f.blocked() {
		return document.ErrBusy
	}
	if err := f.model.flush(); err != nil {
		return err
	}
	f.openDestination = path
	if f.model.History().Dirty {
		f.prompt = "open"
		return nil
	}
	return f.main.Open(f.ref.Context(), document.OpenOptions{Path: path})
}
func (f *Files) continueOpen() {
	if f.workspace != nil && f.workspace.State().Busy {
		f.resumeOpen = true
		return
	}
	f.model.accept(f.openNotebook(f.openDestination))
}

func (f *Files) resumeExternal() {
	if f.pendingExternal == nil || f.restoring || f.main.State().Busy || f.recovery.State().Busy || f.workspace != nil && f.workspace.State().Busy {
		return
	}
	request := f.pendingExternal
	f.pendingExternal = nil
	f.model.accept(f.OpenPath(request.path, request.initial))
}
