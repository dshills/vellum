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
		return cogent.SettleTextField(ref.Context(), a, window, "notes-search", func(err error) {
			if err == nil {
				err = ref.Context().Err()
			}
			if err != nil {
				complete(err)
				return
			}
			// Search is always mounted, even when no note is selected. Capture
			// the current title/body only after its controlled Change is admitted.
			titleID, body := m.CurrentInput()
			if titleID == "" {
				complete(nil)
				return
			}
			if err := cogent.SettleTextField(ref.Context(), a, window, titleID, func(err error) {
				if err != nil {
					complete(err)
					return
				}
				if err := cogent.SettleEditor(ref.Context(), a, window, body, complete); err != nil {
					complete(err)
				}
			}); err != nil {
				complete(err)
			}
		})
	}); err != nil {
		return err
	}
	files, err = notes.AttachFiles(a, ref, m, repository, offer)
	if err != nil {
		return err
	}
	runErr := a.Run(ctx, cogent.Driver{Theme: notes.Theme(), Commands: m.Commands(), OnError: func(err error) { fmt.Fprintln(os.Stderr, err) }})
	stopErr := files.Stop()
	if stopErr != nil {
		releaseLock = false
	}
	return errors.Join(runErr, stopErr)
}
