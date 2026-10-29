// documents exercises two independent documents and active native menu routing.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/examples/macapps/notes"
	"os"
	"path/filepath"
)

type documentWindow struct {
	ref        app.WindowRef
	model      *notes.Model
	files      *notes.Files
	repository *notes.Repository
}

// workspace is confined to Application owner callbacks, including activation and menus.
type workspace struct {
	owner     *app.Application
	documents map[app.WindowRef]*documentWindow
	active    app.WindowRef
	menus     *command.Registry
}

func (w *workspace) activate(ref app.WindowRef) {
	if w.documents[ref] != nil && ref.Context().Err() == nil {
		w.active = ref
		_ = w.owner.Invalidate("")
	}
}
func (w *workspace) current() (*documentWindow, error) {
	d := w.documents[w.active]
	if d == nil || w.active.Context().Err() != nil {
		return nil, app.ErrWindowClosed
	}
	return d, nil
}
func (w *workspace) bindMenus() error {
	w.menus = command.New()
	for _, id := range []string{"notes.new", "notes.open", "notes.save", "notes.save-as", "notes.undo", "notes.redo", "notes.delete", "notes.pin"} {
		id := id
		c := command.Command{ID: id, Label: map[string]string{"notes.new": "New Note", "notes.open": "Open…", "notes.save": "Save", "notes.save-as": "Save As…", "notes.undo": "Undo", "notes.redo": "Redo", "notes.delete": "Delete", "notes.pin": "Pin"}[id], Available: func() error {
			d, e := w.current()
			if e != nil {
				return e
			}
			return d.model.Commands().Enabled(id)
		}, Execute: func(ctx context.Context) error {
			d, e := w.current()
			if e != nil {
				return e
			}
			e = d.model.Commands().Invoke(ctx, id)
			_ = w.owner.Invalidate("")
			return e
		}}
		if e := w.menus.Register(c); e != nil {
			return e
		}
	}
	return nil
}
func openDocument(a *app.Application, id, dir string) (result *documentWindow, err error) {
	r, e := notes.OpenRepository(dir)
	if e != nil {
		return nil, e
	}
	d := &documentWindow{repository: r, model: notes.New()}
	defer func() {
		if err != nil {
			if d.ref.ID() != "" {
				_ = a.CloseWindow(d.ref)
			}
			_ = r.Close()
		}
	}()
	d.model.SetNativeToolbar(true)
	offer, e := r.Load(d.model)
	if e != nil {
		return nil, e
	}
	d.ref, e = a.OpenWindow(app.Window{ID: id, Title: "Vellum Documents — " + id, Width: 1000, Height: 720, Chrome: app.WindowChrome{Style: app.ChromeStandard, Appearance: app.AppearanceLight}, View: d.model.View, CloseRequest: func() bool { return d.files != nil && d.files.RequestClose() }})
	if e != nil {
		return nil, e
	}
	e = d.model.SetSettlement(func(done func(error)) error {
		finish := func(e error) {
			if e == nil {
				e = d.ref.Context().Err()
			}
			done(e)
			_ = a.Invalidate(id)
		}
		return cogent.SettleTextField(d.ref.Context(), a, id, "notes-search", func(e error) {
			if e != nil {
				finish(e)
				return
			}
			title, body := d.model.CurrentInput()
			if title == "" {
				finish(nil)
				return
			}
			if e := cogent.SettleTextField(d.ref.Context(), a, id, title, func(e error) {
				if e != nil {
					finish(e)
					return
				}
				if e := cogent.SettleEditor(d.ref.Context(), a, id, body, finish); e != nil {
					finish(e)
				}
			}); e != nil {
				finish(e)
			}
		})
	})
	if e != nil {
		return nil, e
	}
	d.files, e = notes.AttachFiles(a, d.ref, d.model, r, offer)
	if e != nil {
		return nil, e
	}
	return d, nil
}
func run() (runErr error) {
	dir := flag.String("data-dir", "", "required absolute private workspace directory")
	flag.Parse()
	if !filepath.IsAbs(*dir) {
		return fmt.Errorf("supply -data-dir /absolute/private/workspace")
	}
	a, e := app.New(app.Config{})
	if e != nil {
		return e
	}
	w := &workspace{owner: a, documents: map[app.WindowRef]*documentWindow{}}
	defer func() {
		a.Close()
		for _, d := range w.documents {
			if d.files != nil {
				if e := d.files.Stop(); e != nil {
					runErr = errors.Join(runErr, e)
					continue
				}
			}
			_ = d.repository.Close()
		}
	}()
	var files []*notes.Files
	for _, id := range []string{"A", "B"} {
		d, e := openDocument(a, id, filepath.Join(*dir, id))
		if e != nil {
			return e
		}
		w.documents[d.ref] = d
		files = append(files, d.files)
	}
	if e = w.bindMenus(); e != nil {
		return e
	}
	quit, e := notes.NewQuitGroup(a, files...)
	if e != nil {
		return e
	}
	return a.Run(context.Background(), cogent.Driver{OnQuitRequested: quit.Request, Theme: notes.Theme(), Commands: w.menus, CommandMenus: nativeNotesMenus(), OnWindowActivated: w.activate, TitlebarToolbars: func(ref app.WindowRef) (*cogent.TitlebarToolbar, error) {
		d := w.documents[ref]
		if d == nil {
			return nil, app.ErrWindowClosed
		}
		t, e := nativeNotesToolbar(ref, d.model)
		if t != nil {
			t.CustomizationID = "vellum.documents." + ref.ID() + ".toolbar.v1"
		}
		return t, e
	}, OnError: func(e error) {
		if !errors.Is(e, app.ErrWindowClosed) {
			fmt.Fprintln(os.Stderr, e)
		}
	}})
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
