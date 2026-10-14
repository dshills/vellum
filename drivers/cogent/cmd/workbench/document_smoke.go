package main

import (
	"cogentcore.org/core/core"
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/history"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"time"
)

// This opt-in native acceptance fixture uses only a temporary document and
// a deterministic chooser. It does not certify actual file-panel selection.
type documentSmokeStore struct{ fail atomic.Bool }

func (s *documentSmokeStore) Read(ctx context.Context, path string, max int64) ([]byte, error) {
	return (document.FileStore{}).Read(ctx, path, max)
}
func (s *documentSmokeStore) Write(ctx context.Context, path string, b []byte, max int64) (bool, error) {
	if s.fail.Load() {
		return false, errors.New("injected document write failure")
	}
	return (document.FileStore{}).Write(ctx, path, b, max)
}
func documentOwner(a *app.Application, fn func() error) error {
	var result error
	err := a.Do(func() { result = fn(); a.Invalidate("main") })
	if err != nil {
		return err
	}
	return result
}
func documentWait(ctx context.Context, a *app.Application, fn func() bool) error {
	for {
		ready := false
		if err := a.Do(func() { ready = fn() }); err != nil {
			return err
		}
		if ready {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}
func documentControl(ctx context.Context, id string) error {
	for {
		inspection, err := cogent.InspectAccessibility(ctx, "main")
		if err == nil && inspection.Reachable {
			for _, e := range inspection.Elements {
				if e.DomainID == id && e.Frame[2] > 0 && e.Enabled {
					return nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("native control %s: %w", id, ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
}
func documentPress(ctx context.Context, id string) error {
	if err := documentControl(ctx, id); err != nil {
		return err
	}
	ok, err := cogent.PerformAccessibilityAction(ctx, "main", id, "press", "")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("native action refused %s", id)
	}
	return nil
}
func runDocumentSmoke(ctx context.Context, a *app.Application, session *document.Session[documentImage], flow *documentWorkflow, j *history.Journal, registry *command.Registry, parts *partSource, board *canvas.Scene, refs *fileReferences, path string, store *documentSmokeStore) error {
	if err := documentControl(ctx, "inspector.value"); err != nil {
		return err
	}
	if err := documentOwner(a, func() error {
		if err := registry.Invoke(ctx, "component.rotate"); err != nil {
			return err
		}
		if err := registry.Invoke(ctx, "component.raise-value"); err != nil {
			return err
		}
		return refs.replace(j, []string{filepath.Join(filepath.Dir(path), "reference Ω"), filepath.Join(filepath.Dir(path), "other")})
	}); err != nil {
		return err
	}
	native := core.TheApp.WindowByName("Vellum Workbench")
	if native == nil {
		return errors.New("workbench native window absent")
	}
	native.CloseReq()
	if err := documentPress(ctx, "document.cancel"); err != nil {
		return err
	}
	if err := documentWait(ctx, a, func() bool { return flow.pending == "" && j.Snapshot().Dirty && len(a.Windows()) == 1 }); err != nil {
		return err
	}
	fmt.Println("PASS native CloseReq refusal, modal Cancel and intact workbench")
	var expected documentImage
	if err := documentOwner(a, func() error {
		expected = captureDocument(parts, board, refs)
		return session.Save(ctx, document.SaveOptions{Path: path})
	}); err != nil {
		return err
	}
	if err := documentWait(ctx, a, func() bool { return !session.State().Busy && session.State().Last.Action == document.Save }); err != nil {
		return err
	}
	if err := documentOwner(a, func() error {
		if session.State().Last.Err != nil || j.Snapshot().Dirty {
			return fmt.Errorf("save acknowledgment: %+v", session.State())
		}
		if err := registry.Invoke(ctx, "component.rotate"); err != nil {
			return err
		}
		return flow.open(ctx)
	}); err != nil {
		return err
	}
	if err := documentPress(ctx, "document.discard"); err != nil {
		return err
	}
	if err := documentWait(ctx, a, func() bool {
		return !session.State().Busy && session.State().Last.Action == document.Open && flow.pending == ""
	}); err != nil {
		return err
	}
	var actual documentImage
	if err := documentOwner(a, func() error {
		actual = captureDocument(parts, board, refs)
		if session.State().Last.Err != nil || j.Snapshot().Dirty {
			return errors.New("load didn't acknowledge clean document")
		}
		return nil
	}); err != nil {
		return err
	}
	// Full diagnostic comparison runs on this fixture worker, not the UI owner.
	if !reflect.DeepEqual(materializeDocument(actual), materializeDocument(expected)) {
		return errors.New("load didn't restore complete component/geometry/reference document")
	}
	if err := documentOwner(a, func() error { return registry.Invoke(ctx, "component.raise-value") }); err != nil {
		return err
	}
	fmt.Printf("PASS asynchronous %d-component save/load with exact geometry and ordered references\n", expected.captured.Len())
	native.CloseReq()
	store.fail.Store(true)
	if err := documentPress(ctx, "document.save"); err != nil {
		return err
	}
	if err := documentWait(ctx, a, func() bool {
		return !session.State().Busy && session.State().Last.Err != nil && flow.pending == "close" && j.Snapshot().Dirty && len(a.Windows()) == 1
	}); err != nil {
		return err
	}
	fmt.Println("PASS failed Save keeps unsaved native window open")
	store.fail.Store(false)
	if err := documentPress(ctx, "document.save"); err != nil {
		return err
	}
	select {
	case <-a.Context().Done():
	case <-ctx.Done():
		return ctx.Err()
	}
	data, err := (document.FileStore{}).Read(ctx, path, document.DefaultMaxBytes)
	if err != nil {
		return err
	}
	saved, err := decodeDocument(data, 100)
	if err != nil {
		return err
	}
	if saved.Parts[0].Value != expected.captured.parts.at(0).Value+100 {
		return errors.New("save-close did not persist final edit")
	}
	fmt.Println("PASS native Save confirmation closes only after successful current-revision acknowledgment")
	return nil
}
