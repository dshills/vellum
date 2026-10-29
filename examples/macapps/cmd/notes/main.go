// This example consumes Vellum through public Go APIs.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/examples/macapps/notes"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	directory := flag.String("data-dir", "", "private notebook directory (defaults to Application Support/Vellum Notes)")
	flag.Parse()
	if *directory == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		*directory = filepath.Join(base, "Vellum Notes")
	}
	repository, err := notes.OpenRepository(*directory)
	if err != nil {
		return err
	}
	releaseLock := true
	defer func() {
		if releaseLock {
			_ = repository.Close()
		}
	}()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	m := notes.New()
	m.SetNativeToolbar(true)
	offer, err := repository.Load(m)
	if err != nil {
		return fmt.Errorf("notes files preserved; startup refused: %w", err)
	}
	var files *notes.Files
	const window = "notes"
	ref, err := a.OpenWindow(app.Window{ID: window, Title: "Vellum Notes", Width: 1100, Height: 740, View: m.View, Chrome: app.WindowChrome{Style: app.ChromeStandard, Appearance: app.AppearanceLight}, CloseRequest: func() bool {
		if files == nil {
			return false
		}
		return files.RequestClose()
	}})
	if err != nil {
		return err
	}
	if err := m.SetSettlement(func(callback func(error)) error {
		// This function and all completion callbacks execute on the UI owner.
		// They never block Cocoa main or wait for their own FIFO acknowledgments.
		_ = a.Invalidate(window)
		complete := func(err error) {
			if err == nil {
				err = ref.Context().Err()
			}
			callback(err)
			_ = a.Invalidate(window)
		}
		return m.SettleInputs(func(field string, callback func(error)) error {
			return cogent.SettleTextField(ref.Context(), a, window, field, callback)
		}, func(body *editor.Model, callback func(error)) error {
			return cogent.SettleEditor(ref.Context(), a, window, body, callback)
		}, complete)
	}); err != nil {
		return err
	}
	files, err = notes.AttachFiles(a, ref, m, repository, offer)
	if err != nil {
		return err
	}
	if err := files.ConfigureRestoration(repository); err != nil {
		return err
	}
	quit, err := notes.NewQuitGroup(a, files)
	if err != nil {
		return err
	}
	runErr := a.Run(ctx, cogent.Driver{OnOpenFiles: func(e cogent.OpenFilesEvent) { files.HandleOpenFiles(e.Paths, e.Initial) }, OnQuitRequested: quit.Request, OnWindowActivated: func(app.WindowRef) { files.Activated() }, Theme: notes.Theme(), Commands: m.Commands(), CommandMenus: nativeNotesMenus(), TitlebarToolbars: func(opening app.WindowRef) (*cogent.TitlebarToolbar, error) { return nativeNotesToolbar(opening, m) }, OnError: func(err error) { fmt.Fprintln(os.Stderr, err) }})
	stopErr := files.Stop()
	if stopErr != nil {
		releaseLock = false
	}
	return errors.Join(runErr, stopErr)
}
