package app

import (
	"context"
	"errors"
	"github.com/dshills/vellum/ui"
	"testing"
)

func sampleWindow(id string) Window {
	return Window{ID: id, Title: id, Width: 100, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Column} }}
}

func windowRef(t *testing.T, a *Application, id string) WindowRef {
	t.Helper()
	ref, err := a.WindowRef(id)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

type windowsDriver struct{ t *testing.T }

func (d windowsDriver) Run(ctx context.Context, a *Application) error {
	t := d.t
	if err := a.Open(sampleWindow("second")); err != nil {
		t.Fatal(err)
	}
	firstRef, secondRef := windowRef(t, a, "first"), windowRef(t, a, "second")
	first, second := firstRef.Context(), secondRef.Context()
	count := 0
	if err := a.PostWindow(firstRef, func() { t.Fatal("closed window delivery") }); err != nil {
		t.Fatal(err)
	}
	if err := a.CloseWindow(windowRef(t, a, "first")); err != nil {
		t.Fatal(err)
	}
	if first.Err() == nil || second.Err() != nil || ctx.Err() != nil {
		t.Fatal("window lifetime isolation failed")
	}
	reopened, err := a.OpenWindow(sampleWindow("first"))
	if err != nil || reopened == firstRef {
		t.Fatal("fresh opening refused or aliased", err)
	}
	if err := a.CloseWindow(reopened); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(a.PostWindow(firstRef, func() {}), ErrWindowClosed) {
		t.Fatal("closed delivery accepted")
	}
	if err := a.PostWindow(secondRef, func() { count++ }); err != nil {
		t.Fatal(err)
	}
	a.Drain()
	if count != 1 || len(a.Windows()) != 1 {
		t.Fatal("surviving window lost")
	}
	if err := a.CloseWindow(secondRef); err != nil {
		t.Fatal(err)
	}
	if ctx.Err() == nil || second.Err() == nil {
		t.Fatal("last close did not stop lifetime")
	}
	return nil
}
func TestDynamicWindowIsolation(t *testing.T) {
	a, _ := New(Config{})
	a.Open(sampleWindow("first"))
	if err := a.Run(context.Background(), windowsDriver{t}); err != nil {
		t.Fatal(err)
	}
}
func TestWindowTopologyWakes(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	calls := 0
	a.SetWake(func() { calls++ })
	a.Open(sampleWindow("first"))
	a.Open(sampleWindow("second"))
	a.CloseWindow(windowRef(t, a, "first"))
	if calls != 3 {
		t.Fatal(calls)
	}
}

type fakeServices struct{ text string }

func (s *fakeServices) ReadClipboard(context.Context) (string, error)    { return s.text, nil }
func (s *fakeServices) WriteClipboard(_ context.Context, v string) error { s.text = v; return nil }
func (s *fakeServices) ChooseFile(context.Context, FileOptions) (string, error) {
	return "test.txt", nil
}
func (s *fakeServices) SetMenu(context.Context, []MenuGroup) error { return nil }
func TestInjectedServicesLifetime(t *testing.T) {
	a, _ := New(Config{})
	if _, err := a.Services(); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	a.SetServices(&fakeServices{})
	s, err := a.Services()
	if err != nil {
		t.Fatal(err)
	}
	if err = s.WriteClipboard(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if v, err := s.ReadClipboard(context.Background()); err != nil || v != "hello" {
		t.Fatal(v, err)
	}
	a.Close()
	if _, err := a.Services(); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
func TestFileOptionValidation(t *testing.T) {
	for _, o := range []FileOptions{{Mode: 9}, {Extensions: []string{".txt"}}, {Extensions: []string{"../txt"}}, {Title: "a\x00b"}} {
		if o.Validate() == nil {
			t.Fatal(o)
		}
	}
	if err := (FileOptions{Extensions: []string{"txt", "go"}}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentLastCloseAndOpen(t *testing.T) {
	for i := 0; i < 100; i++ {
		a, _ := New(Config{})
		a.Open(sampleWindow("first"))
		firstRef := windowRef(t, a, "first")
		start := make(chan struct{})
		opened := make(chan error, 1)
		closed := make(chan error, 1)
		go func() { <-start; opened <- a.Open(sampleWindow("second")) }()
		go func() { <-start; closed <- a.CloseWindow(firstRef) }()
		close(start)
		openErr, closeErr := <-opened, <-closed
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		if openErr == nil {
			if a.Context().Err() != nil || len(a.Windows()) != 1 {
				t.Fatal("accepted window closed by a stale last-close decision")
			}
		} else if !errors.Is(openErr, ErrClosed) {
			t.Fatal(openErr)
		}
		a.Close()
	}
}

func TestNativeWindowAndFileStringsValidateBeforeDispatch(t *testing.T) {
	for _, bad := range []string{"bad\x00name", "\xff"} {
		cases := map[string]FileOptions{
			"title": {Title: bad}, "directory": {Directory: bad}, "filename": {Filename: bad}, "extension": {Extensions: []string{bad}},
		}
		for field, options := range cases {
			t.Run(field, func(t *testing.T) {
				if options.Validate() == nil {
					t.Fatal("invalid native string accepted")
				}
			})
		}
		a, _ := New(Config{})
		w := sampleWindow("window")
		w.Title = bad
		if err := a.Open(w); err == nil {
			t.Fatal("invalid native window title accepted")
		}
		a.Close()
	}
}
