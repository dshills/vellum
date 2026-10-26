package contentexample

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// Each setup owns one admitted chooser call; overlapping requests must refuse.
type chooser struct {
	calls    atomic.Int32
	started  chan app.WindowRef
	release  chan string
	finished chan struct{}
}

func (c *chooser) ReadClipboard(context.Context) (string, error) { return "", app.ErrUnavailable }
func (c *chooser) WriteClipboard(context.Context, string) error  { return app.ErrUnavailable }
func (c *chooser) ChooseFile(context.Context, app.FileOptions) (string, error) {
	return "", app.ErrUnavailable
}
func (c *chooser) SetMenu(context.Context, []app.MenuGroup) error { return app.ErrUnavailable }
func (c *chooser) RevealPath(context.Context, string) error       { return nil }
func (c *chooser) ChoosePath(ctx context.Context, w app.WindowRef, o app.PathOptions) (string, error) {
	if c.calls.Add(1) != 1 {
		return "", errors.New("one-shot chooser reused")
	}
	defer close(c.finished)
	c.started <- w
	select {
	case p := <-c.release:
		return p, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
func setup(t *testing.T) (*app.Application, app.WindowRef, *Session, *chooser) {
	t.Helper()
	a, _ := app.New(app.Config{})
	t.Cleanup(a.Close)
	w, e := a.OpenWindow(app.Window{ID: "reference", Title: "Synthetic", Width: 400, Height: 400, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }})
	if e != nil {
		t.Fatal(e)
	}
	s, e := New(a, w)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(s.Close)
	c := &chooser{started: make(chan app.WindowRef, 1), release: make(chan string, 1), finished: make(chan struct{})}
	a.SetServices(c)
	return a, w, s, c
}
func await(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("bounded worker did not finish")
	}
}
func TestSessionExactWindowAndOwnerDelivery(t *testing.T) {
	a, w, s, c := setup(t)
	p := filepath.Join(t.TempDir(), "reference.pdf")
	os.WriteFile(p, []byte("synthetic descriptor only"), 0600)
	if e := s.Field.Request(ui.FilePickerEvent{ID: s.Field.ID, Kind: s.Field.Kind, Action: ui.FilePickerBrowse}); e != nil {
		t.Fatal(e)
	}
	if got := <-c.started; got != w {
		t.Fatal("wrong window")
	}
	if e := s.Field.Request(ui.FilePickerEvent{ID: s.Field.ID, Kind: s.Field.Kind, Action: ui.FilePickerBrowse}); !errors.Is(e, app.ErrServiceBusy) {
		t.Fatal(e)
	}
	c.release <- p
	await(t, c.finished)
	ticket := s.ticket
	// Wait for Publish without mutating UI; acceptance runs only when owner drains.
	until := time.Now().Add(2 * time.Second)
	for a.Pending().ReadyCompletions == 0 && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
	}
	if s.Field.Path != "" {
		t.Fatal("worker mutated owner state")
	}
	if !s.busy() {
		t.Fatal("published result became nonbusy before owner delivery")
	}
	if e := s.Field.Request(ui.FilePickerEvent{ID: s.Field.ID, Kind: s.Field.Kind, Action: ui.FilePickerBrowse}); !errors.Is(e, app.ErrServiceBusy) {
		t.Fatal("published overlap admitted", e)
	}
	if c.calls.Load() != 1 {
		t.Fatal("overlap started another chooser")
	}

	if e := a.Drain(); e != nil {
		t.Fatal(e)
	}
	await(t, ticket.Done())
	if s.PDF.Path != p || s.Media.Path != "" {
		t.Fatal("source not accepted")
	}
	if _, e := s.View(); e != nil {
		t.Fatal(e)
	}
}
func TestSessionCloseRevokesPendingChoice(t *testing.T) {
	a, w, s, c := setup(t)
	s.Field.Request(ui.FilePickerEvent{ID: s.Field.ID, Kind: s.Field.Kind, Action: ui.FilePickerBrowse})
	<-c.started
	a.CloseWindow(w)
	await(t, c.finished)
	s.Close()
	a.Drain()
	if s.Field.Path != "" || s.busy() {
		t.Fatal("late acceptance")
	}
	if _, e := s.View(); e != nil {
		t.Fatal(e)
	}
}

func TestSessionDropNeedsNoNativeChooserAndRejectsChangedConfiguration(t *testing.T) {
	a, _, s, _ := setup(t)
	a.SetServices(nil)
	path := filepath.Join(t.TempDir(), "reference.pdf")
	if e := os.WriteFile(path, []byte("synthetic"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := s.Field.Request(ui.FilePickerEvent{ID: s.Field.ID, Kind: s.Field.Kind, Action: ui.FilePickerDrop, Path: path}); e != nil {
		t.Fatal(e)
	}
	until := time.Now().Add(2 * time.Second)
	for a.Pending().ReadyCompletions == 0 && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
	}
	a.Drain()
	if s.PDF.Path != path {
		t.Fatal("drop required native service", s.Status)
	}
	s.Field.Request(ui.FilePickerEvent{ID: s.Field.ID, Kind: s.Field.Kind, Action: ui.FilePickerDrop, Path: path})
	s.Field.Extensions = []string{"wav"}
	until = time.Now().Add(2 * time.Second)
	for a.Pending().ReadyCompletions == 0 && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
	}
	a.Drain()
	if s.Status != "Reference configuration changed" {
		t.Fatal("stale configuration accepted", s.Status)
	}
}
