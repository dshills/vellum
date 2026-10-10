//go:build (windows || (linux && !android)) && cgo

package cogent

import (
	"context"
	"image"
	"sync"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
)

type desktopMenuMount struct {
	service    *platformServices
	bar        *core.Frame
	generation uint64
}

var desktopMenus = struct {
	sync.Mutex
	groups     []nativeMenuGroup
	generation uint64
	mounts     map[*core.Body]*desktopMenuMount
}{mounts: make(map[*core.Body]*desktopMenuMount)}

// Platform menus are inside each native window on Windows/Linux. No process
// global OS menu is claimed; activation still rechecks registry availability.
func nativeSetMenu(ctx context.Context, groups []nativeMenuGroup, s *platformServices) error {
	if err := nativeServiceContextErr(ctx); err != nil {
		return err
	}
	err := s.owner.Do(func() {
		if nativeServiceContextErr(ctx) != nil {
			return
		}
		desktopMenus.Lock()
		desktopMenus.groups = cloneDesktopMenu(groups)
		desktopMenus.generation++
		desktopMenus.Unlock()
		s.owner.Invalidate("")
	})
	if err != nil {
		return err
	}
	return nativeServiceContextErr(ctx)
}
func cloneDesktopMenu(groups []nativeMenuGroup) []nativeMenuGroup {
	copyGroups := make([]nativeMenuGroup, len(groups))
	for i, g := range groups {
		copyGroups[i] = g
		copyGroups[i].Items = append([]nativeMenuItem(nil), g.Items...)
	}
	return copyGroups
}

// Mount, refresh and unmount run under the owner and this window's render lock.
func nativeMountServices(body *core.Body, s *platformServices) {
	bar := core.NewFrame(body)
	bar.SetName("vellum-platform-menu")
	bar.Styler(func(style *styles.Style) { style.Direction = styles.Row; style.Grow.Set(1, 0) })
	desktopMenus.Lock()
	desktopMenus.mounts[body] = &desktopMenuMount{service: s, bar: bar}
	desktopMenus.Unlock()
	nativeRefreshServices(body)
}
func nativeRefreshServices(body *core.Body) {
	desktopMenus.Lock()
	mount := desktopMenus.mounts[body]
	if mount == nil || mount.generation == desktopMenus.generation {
		desktopMenus.Unlock()
		return
	}
	groups := cloneDesktopMenu(desktopMenus.groups)
	mount.generation = desktopMenus.generation
	desktopMenus.Unlock()
	mount.bar.DeleteChildren()
	for _, g := range groups {
		group := g
		button := core.NewButton(mount.bar).SetText(group.Label).SetType(core.ButtonText)
		button.SetMenu(func(menu *core.Scene, _ image.Point) {
			for _, v := range group.Items {
				item := v
				option := core.NewButton(menu).SetText(item.Label)
				option.Styler(func(style *styles.Style) { style.SetState(item.Disabled, states.Disabled) })
				option.OnClick(func(events.Event) {
					if !mount.service.stopped.Load() {
						mount.service.invokeMenu(item.Command)
					}
				})
			}
		})
	}
	mount.bar.NeedsLayout()
}
func nativeUnmountServices(body *core.Body) {
	desktopMenus.Lock()
	delete(desktopMenus.mounts, body)
	desktopMenus.Unlock()
}
func clearDesktopMenus() {
	desktopMenus.Lock()
	desktopMenus.groups = nil
	desktopMenus.generation++
	desktopMenus.mounts = make(map[*core.Body]*desktopMenuMount)
	desktopMenus.Unlock()
}
