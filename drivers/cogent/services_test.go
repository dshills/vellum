package cogent

import (
	"context"
	"errors"
	"example.com/vellum/app"
	"example.com/vellum/command"
	"testing"
)

func TestMenuActivationRechecksAvailability(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	r := command.New()
	enabled := true
	executed, reported := 0, 0
	r.Register(command.Command{ID: "save", Label: "Save", Available: func() error {
		if !enabled {
			return errors.New("disabled")
		}
		return nil
	}, Execute: func(context.Context) error { executed++; return nil }})
	s := &platformServices{owner: a, commands: r, onError: func(error) { reported++ }}
	s.invokeMenu("save")
	// Availability changes after the menu click is queued but before dispatch.
	a.Do(func() { enabled = false })
	a.Drain()
	if executed != 0 || reported != 1 {
		t.Fatal(executed, reported)
	}
	a.Do(func() { enabled = true })
	s.invokeMenu("save")
	a.Drain()
	if executed != 1 {
		t.Fatal(executed)
	}
	s.invokeMenu("save")
	a.Close()
	a.Drain()
	if executed != 1 {
		t.Fatal("menu delivery survived app close")
	}
}
func TestServicesRejectClosedOrCanceledBeforeNativeDispatch(t *testing.T) {
	a, _ := app.New(app.Config{})
	s := &platformServices{owner: a}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.ReadClipboard(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.WriteClipboard(ctx, "text"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := s.ChooseFile(ctx, app.FileOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.SetMenu(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := s.ChooseFile(context.Background(), app.FileOptions{Mode: 99}); err == nil {
		t.Fatal("invalid mode")
	}
	if err := s.WriteClipboard(context.Background(), "\xff"); err == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
	if err := s.WriteClipboard(context.Background(), "a\x00b"); err == nil {
		t.Fatal("embedded NUL")
	}
	if err := s.SetMenu(context.Background(), []app.MenuGroup{{Label: "File", Commands: []string{"unknown"}}}); err == nil {
		t.Fatal("unknown menu command")
	}
	a.Close()
	if _, err := s.ReadClipboard(context.Background()); !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
	if err := s.SetMenu(context.Background(), nil); !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
}
func TestPendingServiceContextEndsWithApp(t *testing.T) {
	a, _ := app.New(app.Config{})
	s := &platformServices{owner: a}
	scope, cancel, err := s.serviceContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	a.Close()
	if nativeServiceContextErr(scope) == nil {
		t.Fatal("shutdown native dispatch check not immediate")
	}
	<-scope.Done()
	if scope.Err() == nil {
		t.Fatal("shutdown caller not canceled")
	}
}

func TestNativeMenuStringValidation(t *testing.T) {
	for _, bad := range []string{"bad\x00name", "\xff"} {
		for _, field := range []string{"group", "command"} {
			t.Run(field, func(t *testing.T) {
				a, _ := app.New(app.Config{})
				defer a.Close()
				r := command.New()
				label, group := "Save", "File"
				if field == "group" {
					group = bad
				} else {
					label = bad
				}
				r.Register(command.Command{ID: "save", Label: label, Execute: func(context.Context) error { return nil }})
				s := &platformServices{owner: a, commands: r}
				if err := s.SetMenu(context.Background(), []app.MenuGroup{{Label: group, Commands: []string{"save"}}}); err == nil {
					t.Fatal("invalid native menu string accepted")
				}
			})
		}
	}
}
