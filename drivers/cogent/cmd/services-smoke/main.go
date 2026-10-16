//go:build darwin && cgo && !offscreen && vellum_native_checks

// services-smoke checks installed AppKit menu/status objects and native teardown.
// Its actions use native selectors, not physical pointer/keyboard interaction.
// It never reads or writes the clipboard or changes OS preferences.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"slices"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type model struct {
	one, two, denied int
	enabled          bool
	failure          error
}
type counts struct{ One, Two, Denied int }

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	cycles := flag.Int("cycles", 3, "additional native tray create/clear cycles, 1..100")
	flag.Parse()
	if *cycles < 1 || *cycles > 100 {
		return errors.New("cycles must be 1..100")
	}
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	removeObserver, err := cogent.ObserveNativeServiceCleanup()
	if err != nil {
		return err
	}
	defer removeObserver()
	m := &model{enabled: true}
	disabled := errors.New("native service fixture disabled")
	registry := command.New()
	for _, id := range []string{"one", "two"} {
		if err := registry.Register(command.Command{ID: id, Label: "Native check " + id, Available: func() error {
			if !m.enabled {
				return disabled
			}
			return nil
		}, Execute: func(context.Context) error {
			if id == "one" {
				m.one++
			} else {
				m.two++
			}
			return nil
		}}); err != nil {
			return err
		}
	}
	for _, id := range []string{"first", "second"} {
		if err := a.Open(app.Window{ID: id, Title: "Vellum native services " + id, Width: 400, Height: 200, View: func() ui.Node {
			return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Native service fixture: " + id}}}
		}}); err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	outcome := make(chan error, 1)
	go func() {
		outcome <- check(ctx, a, m, *cycles)
		a.Close()
	}()
	runErr := a.Run(ctx, cogent.Driver{Commands: registry, OnError: func(err error) {
		if errors.Is(err, disabled) {
			m.denied++
		} else {
			m.failure = err
		}
	}})
	cancel()
	select {
	case err = <-outcome:
	case <-time.After(5 * time.Second):
		err = errors.New("native service fixture worker did not stop")
	}
	cleanup, observed := cogent.NativeServiceCleanupObservation()
	if !observed || !cleanup.MainThread || !cleanup.CleanupObserved || !cleanup.Menu.OriginalCaptured || !cleanup.Menu.OriginalRestored || !cleanup.Menu.TargetRetired || cleanup.Menu.Owned || cleanup.Tray.Present || !cleanup.Tray.TargetRetired || !cleanup.CallbacksStopped || cleanup.PendingCallbacks != 0 || cleanup.MenuCallbacks != 0 || cleanup.TrayCallbacks != 0 {
		err = errors.Join(err, fmt.Errorf("terminal native service ownership: %+v", cleanup))
	}
	if err := errors.Join(runErr, err); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"result": "PASS_NATIVE_MENU_TRAY", "cleanup": cleanup, "extra_create_clear_cycles": *cycles, "physical_click_tested": false, "clipboard_access": false, "preferences_changed": false})
}
func waitWindow(ctx context.Context, id string) error {
	for {
		s, err := cogent.InspectAccessibility(ctx, id)
		if err == nil && s.Reachable && len(s.Elements) > 0 && s.Elements[0].Frame[2] > 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("window %s not ready: %w", id, ctx.Err())
		case <-time.After(15 * time.Millisecond):
		}
	}
}
func snapshot(ctx context.Context, a *app.Application, m *model) (counts, error) {
	// This owner queue barrier follows the native callback's Post. Direct Do
	// could overtake queued work and cannot prove a negative stale-action check.
	result := make(chan counts, 1)
	if err := a.Post(func() { result <- counts{m.one, m.two, m.denied} }); err != nil {
		return counts{}, err
	}
	select {
	case state := <-result:
		return state, nil
	case <-ctx.Done():
		return counts{}, ctx.Err()
	}
}
func expectCounts(ctx context.Context, a *app.Application, m *model, expected counts) error {
	state, err := snapshot(ctx, a, m)
	if err != nil {
		return err
	}
	if state != expected {
		return fmt.Errorf("command counts %+v, want %+v", state, expected)
	}
	return nil
}
func probe(ctx context.Context, op, index int) (cogent.NativeServiceCheckState, error) {
	s, err := cogent.CheckNativeServices(ctx, op, index)
	if err == nil {
		err = json.NewEncoder(os.Stdout).Encode(map[string]any{"operation": op, "state": s})
	}
	return s, err
}
func iconPNG() ([]byte, error) {
	icon := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	icon.SetNRGBA(0, 0, color.NRGBA{255, 255, 255, 255})
	icon.SetNRGBA(1, 0, color.NRGBA{255, 255, 255, 128})
	var encoded bytes.Buffer
	err := png.Encode(&encoded, icon)
	return encoded.Bytes(), err
}
func check(ctx context.Context, a *app.Application, m *model, cycles int) error {
	firstRef, err := a.WindowRef("first")
	if err != nil {
		return err
	}
	for _, id := range []string{"first", "second"} {
		if err := waitWindow(ctx, id); err != nil {
			return err
		}
	}
	service, err := a.Services()
	if err != nil {
		return err
	}
	tray, err := a.Tray()
	if err != nil {
		return err
	}
	initial, err := probe(ctx, cogent.NativeCheckCaptureOriginalMenu, 0)
	if err != nil {
		return err
	}
	if !initial.Menu.OriginalCaptured || !initial.Menu.OriginalRestored {
		return fmt.Errorf("original menu capture was not observed: %+v", initial.Menu)
	}
	if !slices.Contains(initial.WindowTitles, "Vellum native services first") || !slices.Contains(initial.WindowTitles, "Vellum native services second") {
		return fmt.Errorf("native fixture windows missing: %v", initial.WindowTitles)
	}
	menu := []app.MenuGroup{{Label: "Native checks", Commands: []string{"one", "two"}}}
	if err = service.SetMenu(ctx, menu); err != nil {
		return err
	}
	s, err := probe(ctx, cogent.NativeCheckMenuAction, 0)
	if err != nil || !s.Menu.Owned || len(s.Menu.Items) != 2 || s.Menu.Items[0].Label != "Native check one" || !s.Menu.Items[0].Enabled {
		return errors.Join(err, fmt.Errorf("installed menu: %+v", s.Menu))
	}
	if err = expectCounts(ctx, a, m, counts{1, 0, 0}); err != nil {
		return err
	}
	if err = a.Do(func() { m.enabled = false }); err != nil {
		return err
	}
	if _, err = probe(ctx, cogent.NativeCheckMenuAction, 0); err != nil {
		return err
	}
	if err = expectCounts(ctx, a, m, counts{1, 0, 1}); err != nil {
		return err
	}
	if err = service.SetMenu(ctx, menu); err != nil {
		return err
	}
	s, err = probe(ctx, cogent.NativeCheckSnapshot, 0)
	if err != nil || len(s.Menu.Items) != 2 || s.Menu.Items[0].Enabled || s.Menu.Items[1].Enabled {
		return errors.Join(err, errors.New("disabled native menu presentation"))
	}
	if err = a.Do(func() { m.enabled = true }); err != nil {
		return err
	}
	if _, err = probe(ctx, cogent.NativeCheckCaptureStaleMenu, 0); err != nil {
		return err
	}
	if err = service.SetMenu(ctx, []app.MenuGroup{{Label: "Native replacement", Commands: []string{"two", "one"}}}); err != nil {
		return err
	}
	if _, err = probe(ctx, cogent.NativeCheckStaleMenuAction, 0); err != nil {
		return err
	}
	if err = expectCounts(ctx, a, m, counts{1, 0, 1}); err != nil {
		return err
	}
	if _, err = probe(ctx, cogent.NativeCheckMenuAction, 0); err != nil {
		return err
	}
	if err = expectCounts(ctx, a, m, counts{1, 1, 1}); err != nil {
		return err
	}
	icon, err := iconPNG()
	if err != nil {
		return err
	}
	o := app.TrayOptions{Title: "Vellum check", Tooltip: "Vellum native check", AccessibilityLabel: "Vellum native tray", IconPNG: icon, Template: true, Menu: []app.TrayMenuItem{{Command: "one"}, {Separator: true}, {Command: "two"}}}
	if err = tray.SetTray(ctx, o); err != nil {
		return err
	}
	s, err = probe(ctx, cogent.NativeCheckTrayAction, 0)
	if err != nil || !s.Tray.Present || !s.Tray.Visible || !s.Tray.Template || s.Tray.Identity == 0 || s.Tray.ImageWidth != 18 || s.Tray.ImageHeight != 9 || s.Tray.PixelWidth != 4 || s.Tray.PixelHeight != 2 || s.Tray.AccessibilityLabel != o.AccessibilityLabel || len(s.Tray.Items) != 3 || !s.Tray.Items[1].Separator {
		return errors.Join(err, fmt.Errorf("native status item: %+v", s.Tray))
	}
	identity := s.Tray.Identity
	if err = expectCounts(ctx, a, m, counts{2, 1, 1}); err != nil {
		return err
	}
	if _, err = probe(ctx, cogent.NativeCheckCaptureStaleTray, 0); err != nil {
		return err
	}
	o.Title = "Vellum updated"
	o.Menu = []app.TrayMenuItem{{Command: "two"}, {Command: "one"}}
	if err = tray.SetTray(ctx, o); err != nil {
		return err
	}
	s, err = probe(ctx, cogent.NativeCheckStaleTrayAction, 0)
	if err != nil || s.Tray.Identity != identity || s.Tray.Title != o.Title || len(s.Tray.Items) != 2 || s.Tray.Items[0].Label != "Native check two" {
		return errors.Join(err, errors.New("native tray replacement lost identity/presentation"))
	}
	if err = expectCounts(ctx, a, m, counts{2, 1, 1}); err != nil {
		return err
	}
	if _, err = probe(ctx, cogent.NativeCheckTrayAction, 0); err != nil {
		return err
	}
	if err = expectCounts(ctx, a, m, counts{2, 2, 1}); err != nil {
		return err
	}
	if _, err = probe(ctx, cogent.NativeCheckCaptureStaleTray, 0); err != nil {
		return err
	}
	if err = tray.ClearTray(ctx); err != nil {
		return err
	}
	s, err = probe(ctx, cogent.NativeCheckStaleTrayAction, 0)
	if err != nil || s.Tray.Present || !s.Tray.TargetRetired {
		return errors.Join(err, errors.New("cleared status item retained production ownership"))
	}
	if err = expectCounts(ctx, a, m, counts{2, 2, 1}); err != nil {
		return err
	}
	for i := 0; i < cycles; i++ {
		if err = tray.SetTray(ctx, app.TrayOptions{Title: "Vellum cycle"}); err != nil {
			return err
		}
		s, err = probe(ctx, cogent.NativeCheckSnapshot, 0)
		if err != nil || !s.Tray.Present || s.Tray.Identity <= identity {
			return errors.Join(err, errors.New("new status item identity not observed"))
		}
		identity = s.Tray.Identity
		if err = tray.ClearTray(ctx); err != nil {
			return err
		}
		s, err = probe(ctx, cogent.NativeCheckSnapshot, 0)
		if err != nil || s.Tray.Present {
			return errors.Join(err, errors.New("status item survived clear"))
		}
	}
	o.Title = "Vellum final"
	o.Menu = []app.TrayMenuItem{{Command: "one"}}
	if err = tray.SetTray(ctx, o); err != nil {
		return err
	}
	if err = a.CloseWindow(firstRef); err != nil {
		return err
	}
	if err = waitWindow(ctx, "second"); err != nil {
		return err
	}
	for {
		s, err = cogent.CheckNativeServices(ctx, cogent.NativeCheckSnapshot, 0)
		if err != nil {
			return err
		}
		if !slices.Contains(s.WindowTitles, "Vellum native services first") && slices.Contains(s.WindowTitles, "Vellum native services second") {
			break
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("native window closure not observed: %v: %w", s.WindowTitles, ctx.Err())
		case <-time.After(15 * time.Millisecond):
		}
	}
	if _, err = probe(ctx, cogent.NativeCheckTrayAction, 0); err != nil {
		return err
	}
	if err = expectCounts(ctx, a, m, counts{3, 2, 1}); err != nil {
		return err
	}
	var modelErr error
	if err = a.Do(func() { modelErr = m.failure }); err != nil {
		return err
	}
	return modelErr // Leave the final tray/menu installed for terminal cleanup.
}
