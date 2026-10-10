//go:build (windows || (linux && !android)) && cgo

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"image"
	"testing"
)

func TestMenuOwnedSnapshotAndWindowRefresh(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	s := &platformServices{owner: a}
	first, second := core.NewBody("first"), core.NewBody("second")
	nativeMountServices(first, s)
	nativeMountServices(second, s)
	groups := []nativeMenuGroup{{Label: "File", Items: []nativeMenuItem{{Label: "Open", Command: "open"}}}}
	if err := nativeSetMenu(context.Background(), groups, s); err != nil {
		t.Fatal(err)
	}
	groups[0].Label = "changed"
	groups[0].Items[0].Label = "changed"
	nativeRefreshServices(first)
	nativeRefreshServices(second)
	desktopMenus.Lock()
	f := desktopMenus.mounts[first]
	b := desktopMenus.mounts[second]
	desktopMenus.Unlock()
	if f.bar.Child(0).(*core.Button).Text != "File" || b.bar.Child(0).(*core.Button).Text != "File" {
		t.Fatal("snapshot not independent")
	}
	if f.bar.Child(0) == b.bar.Child(0) {
		t.Fatal("window widgets shared")
	}
	nativeUnmountServices(first)
	desktopMenus.Lock()
	_, exists := desktopMenus.mounts[first]
	desktopMenus.Unlock()
	if exists {
		t.Fatal("mount leaked")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := nativeSetMenu(ctx, nil, s); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	clearDesktopMenus()
	desktopMenus.Lock()
	defer desktopMenus.Unlock()
	if len(desktopMenus.mounts) != 0 || len(desktopMenus.groups) != 0 {
		t.Fatal("shutdown retained menu")
	}
}

func TestDesktopMenuActivationRechecksOwnerAvailability(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	registry := command.New()
	enabled := true
	calls := 0
	reported := 0
	registry.Register(command.Command{ID: "action", Label: "Action", Available: func() error {
		if !enabled {
			return errors.New("disabled")
		}
		return nil
	}, Execute: func(context.Context) error { calls++; return nil }})
	s := &platformServices{owner: a, commands: registry, onError: func(error) { reported++ }}
	body := core.NewBody("activation")
	nativeMountServices(body, s)
	defer nativeUnmountServices(body)
	if err := s.SetMenu(context.Background(), []app.MenuGroup{{Label: "Menu", Commands: []string{"action"}}}); err != nil {
		t.Fatal(err)
	}
	nativeRefreshServices(body)
	desktopMenus.Lock()
	mount := desktopMenus.mounts[body]
	desktopMenus.Unlock()
	menu := core.NewBody("popup").Scene
	mount.bar.Child(0).(*core.Button).Menu(menu, image.Point{})
	enabled = false
	menu.Child(0).(*core.Button).Send(events.Click)
	a.Drain()
	if calls != 0 || reported != 1 {
		t.Fatalf("disabled activation calls=%d reported=%d", calls, reported)
	}
	enabled = true
	menu.Child(0).(*core.Button).Send(events.Click)
	a.Drain()
	if calls != 1 {
		t.Fatalf("command failed to execute on owner: %d", calls)
	}
}
