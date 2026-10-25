package notes

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

// Files is app-local document wiring; it consumes only existing public framework
// APIs. Owner callbacks own all state. A timer only posts bounded owner requests.
type Files struct {
	saved            bool
	owner            *app.Application
	ref              app.WindowRef
	model            *Model
	main             *document.Session[Notebook]
	recovery         *document.Session[Recovery]
	offer            *Recovery
	prompt, intent   string
	clear            bool
	lastRecovery     uint64
	capturedRevision uint64
	capturedClear    bool
	stop             context.CancelFunc
	done             chan struct{}
	stopOnce         sync.Once
}

func AttachFiles(a *app.Application, ref app.WindowRef, m *Model, r *Repository, offer *Recovery) (*Files, error) {
	if a == nil || m == nil || r == nil || r.lock == nil || ref.Context().Err() != nil || m.files != nil {
		return nil, ErrInvalid
	}
	current, err := a.WindowRef(ref.ID())
	if err != nil {
		return nil, err
	}
	if current != ref {
		return nil, app.ErrInvalidWindowRef
	}
	if offer != nil {
		if offer.Version != 1 || offer.Notebook == nil {
			return nil, ErrInvalid
		}
		b := *offer.Notebook
		if err := b.Validate(); err != nil {
			return nil, err
		}
		b.Notes = append([]Note(nil), b.Notes...)
		offer = &Recovery{Version: 1, Notebook: &b}
	}
	f := &Files{saved: r.exists, owner: a, ref: ref, model: m, offer: offer, done: make(chan struct{})}
	f.main, err = document.New(document.Config[Notebook]{Owner: a, Window: ref.ID(), Journal: m.journal, Path: r.Path, Filename: "Notebook.vellumnotes", Extensions: []string{"vellumnotes"}, MaxBytes: MaxNotebookBytes, Codec: document.Codec[Notebook]{Capture: m.CaptureNotebook, Encode: EncodeNotebook, Decode: DecodeNotebook, Install: m.installNotebook}, OnComplete: f.completed})
	if err != nil {
		return nil, err
	}
	recoveryJournal, _ := history.New(1)
	f.recovery, err = document.New(document.Config[Recovery]{Owner: a, Window: ref.ID(), Journal: recoveryJournal, Path: r.RecoveryPath, MaxBytes: MaxNotebookBytes, Codec: document.Codec[Recovery]{Capture: func() (Recovery, error) {
		f.capturedClear = f.clear
		f.capturedRevision = m.History().Revision
		if f.clear {
			return Recovery{Version: 1}, nil
		}
		b, err := m.CaptureNotebook()
		return Recovery{Version: 1, Notebook: &b}, err
	}, Encode: encodeRecovery, Decode: decodeRecovery, Install: func(Recovery) error { return ErrInvalid }}, OnComplete: f.recovered})
	if err != nil {
		return nil, err
	}
	if offer != nil {
		f.prompt = "recovery"
	}
	commands := []command.Command{
		{ID: "notes.save", Label: "Save Notebook", Shortcut: "Meta+S", Execute: func(context.Context) error { return m.Request(func() error { return f.Save(false) }) }},
		{ID: "notes.save-as", Label: "Save Notebook As…", Shortcut: "Meta+Shift+S", Execute: func(context.Context) error { return m.Request(func() error { return f.Save(true) }) }},
		{ID: "notes.open", Label: "Open Notebook…", Shortcut: "Meta+O", Execute: func(context.Context) error { return m.Request(f.Open) }},
	}
	// Preflight the complete batch before mutating the live owner-only registry.
	// No IO, callbacks or concurrent registry mutations occur between these loops.
	validated := command.New()
	for _, c := range m.commands.Search("") {
		if err := validated.Register(c); err != nil {
			return nil, err
		}
	}
	for i := range commands {
		commands[i].Available = m.actionReady
		if err := validated.Register(commands[i]); err != nil {
			return nil, err
		}
	}
	for _, c := range commands {
		if err := m.commands.Register(c); err != nil {
			return nil, err
		}
	}
	m.files = f

	ctx, cancel := context.WithCancel(ref.Context())
	f.stop = cancel
	go func() {
		defer close(f.done)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := a.PostWindow(ref, f.Checkpoint); errors.Is(err, app.ErrClosed) || errors.Is(err, app.ErrWindowClosed) || errors.Is(err, app.ErrInvalidWindowRef) {
					return
				}
			}
		}
	}()
	return f, nil
}
func (f *Files) busy() bool    { return f.main.State().Busy || f.recovery.State().Busy }
func (f *Files) blocked() bool { return f.busy() || f.prompt != "" || f.clear }
func (f *Files) Save(choose bool) error {
	if f.blocked() {
		return document.ErrBusy
	}
	f.model.status = "Saving notebook…"
	return f.main.Save(f.ref.Context(), document.SaveOptions{Choose: choose})
}
func (f *Files) Open() error {
	if f.blocked() {
		return document.ErrBusy
	}
	if err := f.model.flush(); err != nil {
		return err
	}
	if f.model.History().Dirty {
		f.prompt = "open"
		return nil
	}
	return f.main.Open(f.ref.Context(), document.OpenOptions{})
}
func (f *Files) completed(r document.Result) {
	if r.Err != nil {
		f.intent = ""
		f.model.status = r.Err.Error()
		return
	}
	f.saved = true
	if r.Action == document.Save {
		f.model.status = "Saved notebook"
	} else {
		f.model.status = "Opened notebook"
	}
	f.clear = true
	f.startRecovery()
}
func (f *Files) startRecovery() {
	if f.recovery.State().Busy {
		return
	}
	if err := f.recovery.Save(f.ref.Context(), document.SaveOptions{}); err != nil {
		f.model.status = "Recovery: " + err.Error()
		f.clear = false
		f.intent = ""
	}
}
func (f *Files) recovered(r document.Result) {
	if r.Err != nil {
		f.model.status = "Recovery: " + r.Err.Error()
		f.clear = false
		f.intent = ""
		return
	}
	if f.capturedClear {
		f.clear = false
		f.lastRecovery = 0
		intent := f.intent
		f.intent = ""
		if intent == "close" {
			_ = f.owner.CloseWindow(f.ref)
		} else if intent == "open" {
			f.model.accept(f.Open())
		}
	} else {
		f.lastRecovery = f.capturedRevision
		if f.clear {
			f.startRecovery()
		}
	}
}

// Checkpoint captures committed model data; it never settles or commits native
// marked text. The next interval retries a busy/composing/failed capture.
func (f *Files) Checkpoint() {
	if f.blocked() || f.model.pending || !f.model.History().Dirty || f.lastRecovery == f.model.History().Revision {
		return
	}
	if err := f.model.ready(); err != nil {
		return
	}
	f.startRecovery()
}

// RequestClose is a native owner callback. Settlement must finish before deciding
// whether the current revision is saved. Force CloseWindow is used only afterward.
func (f *Files) RequestClose() bool {
	if f.prompt == "recovery" && !f.busy() && !f.model.pending {
		_ = f.owner.CloseWindow(f.ref)
		return false
	}
	if f.blocked() || f.model.pending {
		f.model.status = "Finish the current input, file operation or dialog, then close again"
		_ = f.owner.Invalidate(f.ref.ID())
		return false
	}
	_ = f.model.Request(func() error {
		if err := f.model.flush(); err != nil {
			return err
		}
		if f.model.History().Dirty {
			f.prompt = "close"
			return nil
		}
		f.intent = "close"
		f.clear = true
		f.startRecovery()
		return nil
	})
	_ = f.owner.Invalidate(f.ref.ID())
	return false
}
func (f *Files) settleIntent(intent string, save bool) {
	if f.prompt != intent || f.busy() {
		return
	}
	f.prompt = ""
	if save {
		f.intent = intent
		if err := f.Save(false); err != nil {
			f.intent = ""
			f.model.accept(err)
		}
		return
	}
	if intent == "close" {
		f.intent = "close"
		f.clear = true
		f.startRecovery()
	} else {
		f.model.accept(f.main.Open(f.ref.Context(), document.OpenOptions{DiscardChanges: true}))
	}
}
func (f *Files) status() string {
	if f.prompt == "recovery" {
		return "An interrupted-session recovery copy is available"
	}
	if f.busy() {
		return "Finishing notebook file operation…"
	}
	if f.model.History().Dirty {
		return "Edited · Command-S to save"
	}
	if !f.saved {
		return "Sample notebook · Command-S to keep locally"
	}
	return "Saved · " + f.main.State().Path
}
func (f *Files) promptView() ui.Node {
	prompt := f.prompt
	current := func() bool {
		return f.ref.Context().Err() == nil && f.prompt == prompt && !f.busy() && !f.model.pending
	}
	cancel := func() {
		if current() {
			f.prompt = ""
		}
	}
	text := "Save changes before opening another notebook?"
	if prompt == "close" {
		text = "Save changes before closing Notes?"
	}
	children := []ui.Node{{ID: "notes-file-prompt-text", Kind: ui.Label, Text: text},
		{ID: "notes-decision-save", Kind: ui.Button, Text: "Save", Activate: func() {
			if current() {
				f.settleIntent(prompt, true)
			}
		}},
		{ID: "notes-decision-discard", Kind: ui.Button, Text: "Discard Changes", Activate: func() {
			if current() {
				f.settleIntent(prompt, false)
			}
		}},
		{ID: "notes-decision-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel}}
	if prompt == "recovery" {
		cancel = func() {
			if current() {
				_ = f.owner.CloseWindow(f.ref)
			}
		}
		children = []ui.Node{{ID: "notes-recovery-text", Kind: ui.Label, Text: "Restore the last completed interrupted-session copy? It becomes an edited notebook; Save writes it to this app's local notebook. History is session-local."},
			{ID: "notes-recovery-restore", Kind: ui.Button, Text: "Restore Recovery", Activate: func() {
				if !current() || f.offer == nil || f.offer.Notebook == nil {
					return
				}
				b := *f.offer.Notebook
				if err := f.model.journal.Replace(false, func() error { return f.model.installNotebook(b) }); err != nil {
					f.model.accept(err)
					return
				}
				f.offer = nil
				f.prompt = ""
				f.model.status = "Recovered notebook · Save to keep it"
			}},
			{ID: "notes-recovery-discard", Kind: ui.Button, Text: "Discard Recovery", Activate: func() {
				if !current() {
					return
				}
				f.offer = nil
				f.prompt = ""
				f.clear = true
				f.startRecovery()
			}},
			{ID: "notes-recovery-close", Kind: ui.Button, Text: "Close without changing files", Activate: cancel}}
	}
	return ui.Node{ID: "notes-file-prompt", Kind: ui.Dialog, Text: "Vellum Notes", Open: prompt != "", Dismiss: cancel, Children: children}
}

// Stop runs after Application.Run returns. Do not release the repository lock if
// worker retirement fails; process exit will release it after terminating IO.
func (f *Files) Stop() error {
	f.stopOnce.Do(func() { f.stop() })
	deadline := time.Now().Add(3 * time.Second)
	select {
	case <-f.done:
	case <-time.After(time.Until(deadline)):
		return errors.New("notes: recovery scheduler did not retire; repository lock retained until process exit")
	}
	for f.main.State().WorkerPending || f.recovery.State().WorkerPending {
		if time.Now().After(deadline) {
			return errors.New("notes: file worker did not retire; repository lock retained until process exit")
		}
		time.Sleep(5 * time.Millisecond)
	}
	return nil
}
