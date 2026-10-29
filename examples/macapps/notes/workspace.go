package notes

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const workspaceLimit int64 = 16 << 10

// WindowSession is app-owned metadata, separate from notebook and recovery data.
type WindowSession struct {
	Version   int    `json:"version"`
	Path      string `json:"path"`
	Activated int64  `json:"activated"`
}

func (s WindowSession) validate() error {
	if s.Version != 1 || !filepath.IsAbs(s.Path) || len(s.Path) > 4096 || !validText(s.Path, 4096) || !strings.EqualFold(filepath.Ext(s.Path), ".vellumnotes") || s.Activated < 0 {
		return ErrInvalid
	}
	return nil
}
func decodeWindowSession(data []byte) (WindowSession, error) {
	var s WindowSession
	if int64(len(data)) > workspaceLimit {
		return s, ErrInvalid
	}
	if e := decodeDocument(data, &s); e != nil {
		return s, e
	}
	return s, s.validate()
}
func readWindowSession(directory string) (WindowSession, error) {
	data, e := (document.FileStore{}).Read(context.Background(), filepath.Join(directory, "WindowSession.json"), workspaceLimit)
	if e != nil {
		return WindowSession{}, e
	}
	return decodeWindowSession(data)
}

// SessionActivation is startup-only metadata for ordering native window creation.
// A corrupt/missing record gets no preference; it is never modified here.
func SessionActivation(directory string) int64 {
	s, e := readWindowSession(directory)
	if e != nil {
		return 0
	}
	return s.Activated
}

// ConfigureRestoration runs during startup, before Run. IO reads stay outside the
// owner. The optional reopen itself uses the existing worker/owner session path.
func (f *Files) ConfigureRestoration(r *Repository) error {
	if r == nil || r.lock == nil || f.workspace != nil || f.ref.Context().Err() != nil {
		return ErrInvalid
	}
	state, e := readWindowSession(r.Directory)
	if e != nil && !errors.Is(e, os.ErrNotExist) {
		f.workspaceProtected = true
		f.model.status = "Window session preserved: " + e.Error()
	}
	if e == nil {
		f.activated = state.Activated
		f.restorePath = state.Path
	}
	j, _ := history.New(1)
	f.workspace, e = document.New(document.Config[WindowSession]{Owner: f.owner, Window: f.ref.ID(), Journal: j, Path: filepath.Join(r.Directory, "WindowSession.json"), MaxBytes: workspaceLimit,
		Codec: document.Codec[WindowSession]{Capture: func() (WindowSession, error) {
			s := WindowSession{1, f.main.State().Path, f.activated}
			return s, s.validate()
		}, Encode: func(s WindowSession) ([]byte, error) {
			if e := s.validate(); e != nil {
				return nil, e
			}
			return json.Marshal(s)
		}, Decode: decodeWindowSession, Install: func(WindowSession) error { return ErrInvalid }},
		OnComplete: func(result document.Result) {
			if result.Err != nil {
				f.model.status = "Window session: " + result.Err.Error()
				f.workspaceQueued = false
				f.resumeOpen = false
				if f.quitAwaitWorkspace {
					f.quitAwaitWorkspace = false
					f.finishQuit(false)
				}
				return
			}
			if f.workspaceQueued {
				f.workspaceQueued = false
				f.saveWorkspace()
			}
			if f.resumeOpen && !f.workspace.State().Busy {
				f.resumeOpen = false
				f.continueOpen()
			}
			f.resumeExternal()
			if f.quitAwaitWorkspace && !f.workspace.State().Busy {
				f.quitAwaitWorkspace = false
				f.finishQuit(true)
			}
		}})
	if e != nil {
		return e
	}
	if f.restorePath != "" && f.restorePath != r.Path && f.offer == nil && !f.workspaceProtected {
		return f.owner.Do(func() {
			f.restoring = true
			if err := f.main.Open(f.ref.Context(), document.OpenOptions{Path: f.restorePath}); err != nil {
				f.restoring = false
				f.workspaceProtected = true
				f.model.accept(err)
			}
		})
	}
	return nil
}
func (f *Files) saveWorkspace() {
	if f.workspace == nil || f.workspaceProtected || f.ref.Context().Err() != nil {
		return
	}
	if f.workspace.State().Busy {
		f.workspaceQueued = true
		return
	}
	if err := f.workspace.Save(f.ref.Context(), document.SaveOptions{}); err != nil {
		f.model.status = "Window session: " + err.Error()
	}
}

// Activated is owner-only and coalesces metadata writes while preserving user history.
func (f *Files) Activated() {
	f.activated = time.Now().UnixNano()
	if !f.restoring && f.offer == nil {
		f.saveWorkspace()
	}
}
func (f *Files) replaceMainPath(path string) error {
	if f.main.State().Busy {
		return document.ErrBusy
	}
	next, e := document.New(document.Config[Notebook]{Owner: f.owner, Window: f.ref.ID(), Journal: f.model.journal, Path: path, Filename: "Notebook.vellumnotes", Extensions: []string{"vellumnotes"}, MaxBytes: MaxNotebookBytes, Codec: document.Codec[Notebook]{Capture: f.model.CaptureNotebook, Encode: EncodeNotebook, Decode: DecodeNotebook, Install: f.model.installNotebook}, OnComplete: f.completed})
	if e == nil {
		f.main = next
	}
	return e
}
