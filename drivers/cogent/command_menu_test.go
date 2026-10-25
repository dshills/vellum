package cogent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
)

func TestCommandMenusCopyValidateAndCoalesceAvailability(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	enabled := false
	registry := command.New()
	_ = registry.Register(command.Command{ID: "save", Label: "Save", Shortcut: "Meta+S", Available: func() error {
		if !enabled {
			return errors.New("disabled")
		}
		return nil
	}, Execute: func(context.Context) error { return nil }})
	input := []app.MenuGroup{{Label: "File", Commands: []string{"save"}}}
	groups, err := prepareCommandMenus(input, registry)
	if err != nil {
		t.Fatal(err)
	}
	input[0].Commands[0] = "changed"
	if groups[0].Commands[0] != "save" {
		t.Fatal("caller catalogue retained")
	}
	service := &platformServices{owner: a, commands: registry}
	p := &commandMenuPump{service: service, groups: groups, ctx: a.Context(), wake: make(chan struct{}, 1)}
	p.ready.Store(true)
	_ = a.Do(p.capture)
	if len(p.wake) != 1 || !p.latest.Load().groups[0].Items[0].Disabled {
		t.Fatal("initial menu snapshot")
	}
	enabled = true
	_ = a.Do(p.capture)
	if len(p.wake) != 1 || p.latest.Load().groups[0].Items[0].Disabled {
		t.Fatal("menu updates did not coalesce")
	}
	before := p.latest.Load()
	_ = a.Do(p.capture)
	if p.latest.Load() != before {
		t.Fatal("unchanged menus republished")
	}
	if _, err := prepareCommandMenus([]app.MenuGroup{{Label: "File", Commands: []string{"missing"}}}, registry); err == nil {
		t.Fatal("unknown command accepted")
	}
}

func TestCommandMenuWorkerStopsBeforeDesktopAndKeepsFocusPriority(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	registry := command.New()
	_ = registry.Register(command.Command{ID: "undo", Label: "Undo", Shortcut: "Meta+Z", ShortcutScope: command.AfterFocusShortcuts, Execute: func(context.Context) error { return nil }})
	groups, err := prepareCommandMenus([]app.MenuGroup{{Label: "Edit", Commands: []string{"undo"}}}, registry)
	if err != nil {
		t.Fatal(err)
	}
	p := newCommandMenuPump(&platformServices{owner: a, commands: registry}, groups)
	p.close()
	p.close()
	select {
	case <-p.done:
	default:
		t.Fatal("menu worker survived close")
	}
	fake := &commandMenuPump{service: &platformServices{owner: a, commands: registry}, groups: groups, ctx: a.Context(), wake: make(chan struct{}, 1)}
	fake.ready.Store(true)
	_ = a.Do(fake.capture)
	if fake.latest.Load().groups[0].Items[0].Shortcut != "" {
		t.Fatal("native menu stole focus-first text shortcut")
	}
}

func TestCommandMenuFailedInstallAllowsOneBoundedRetry(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	registry := command.New()
	_ = registry.Register(command.Command{ID: "save", Label: "Save", Execute: func(context.Context) error { return nil }})
	groups, _ := prepareCommandMenus([]app.MenuGroup{{Label: "File", Commands: []string{"save"}}}, registry)
	p := &commandMenuPump{service: &platformServices{owner: a, commands: registry}, groups: groups, ctx: a.Context(), wake: make(chan struct{}, 1)}
	p.ready.Store(true)
	_ = a.Do(p.capture)
	key := p.last
	first := p.latest.Load()
	p.failure.Store(&commandMenuFailure{err: errors.New("temporary"), key: key})
	_ = a.Do(p.reportFailure)
	_ = a.Do(p.capture)
	if p.latest.Load() == first || p.last != key || p.retries != 1 {
		t.Fatal("failed state suppressed retry")
	}
	second := p.latest.Load()
	p.failure.Store(&commandMenuFailure{err: errors.New("still unavailable"), key: key})
	_ = a.Do(p.reportFailure)
	_ = a.Do(p.capture)
	if p.latest.Load() != second {
		t.Fatal("persistent failure entered retry loop")
	}
}
func TestCommandMenuCloseCancelsInFlightInstall(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	ctx, cancel := context.WithCancel(a.Context())
	entered := make(chan struct{})
	p := &commandMenuPump{service: &platformServices{owner: a}, ctx: ctx, cancel: cancel, done: make(chan struct{}), wake: make(chan struct{}, 1), install: func(ctx context.Context, _ []nativeMenuGroup, _ *platformServices) error {
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	}}
	p.latest.Store(&commandMenuSnapshot{key: "request"})
	go p.run()
	p.wake <- struct{}{}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("install did not start")
	}
	p.close()
	select {
	case <-p.done:
	default:
		t.Fatal("in-flight native operation stranded worker")
	}
}
