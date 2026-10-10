package cogent

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"example.com/vellum/app"
	"example.com/vellum/command"
)

type nativeMenuItem struct {
	Label, Command string
	Disabled       bool
}
type nativeMenuGroup struct {
	Label string
	Items []nativeMenuItem
}
type platformServices struct {
	owner    *app.Application
	commands *command.Registry
	onError  func(error)
	stopped  atomic.Bool
}

func (s *platformServices) check(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("cogent: service context required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.stopped.Load() || s.owner.Context().Err() != nil {
		return app.ErrClosed
	}
	return nil
}

type serviceLifetimeKey struct{}

// The caller context retains the normal Context contract. Native dispatch also
// checks the app lifetime directly so app shutdown prevents a pending OS action
// even before its AfterFunc has propagated cancellation to the caller scope.
func nativeServiceContextErr(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if lifetime, ok := ctx.Value(serviceLifetimeKey{}).(context.Context); ok {
		return lifetime.Err()
	}
	return nil
}
func (s *platformServices) serviceContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	scope, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(s.owner.Context(), cancel)
	return context.WithValue(scope, serviceLifetimeKey{}, s.owner.Context()), func() { stop(); cancel() }, nil
}
func (s *platformServices) ReadClipboard(ctx context.Context) (string, error) {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return "", err
	}
	defer cancel()
	return nativeReadClipboard(scope)
}
func (s *platformServices) WriteClipboard(ctx context.Context, value string) error {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if !utf8.ValidString(value) {
		return fmt.Errorf("cogent: invalid UTF-8 in clipboard text")
	}
	if strings.ContainsRune(value, 0) {
		return fmt.Errorf("cogent: NUL in clipboard text")
	}
	return nativeWriteClipboard(scope, value)
}
func (s *platformServices) ChooseFile(ctx context.Context, o app.FileOptions) (string, error) {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return "", err
	}
	defer cancel()
	if err := o.Validate(); err != nil {
		return "", err
	}
	return nativeChooseFile(scope, o)
}
func (s *platformServices) SetMenu(ctx context.Context, groups []app.MenuGroup) error {
	if err := s.check(ctx); err != nil {
		return err
	}
	var menu []nativeMenuGroup
	err := s.owner.Do(func() {
		for _, group := range groups {
			if strings.TrimSpace(group.Label) == "" || !utf8.ValidString(group.Label) || strings.ContainsRune(group.Label, 0) {
				menu = nil
				return
			}
			g := nativeMenuGroup{Label: group.Label}
			for _, id := range group.Commands {
				if s.commands == nil {
					menu = nil
					return
				}
				found := false
				for _, cmd := range s.commands.Search("") {
					if cmd.ID == id {
						if !utf8.ValidString(cmd.Label) || strings.ContainsRune(cmd.Label, 0) {
							menu = nil
							return
						}
						g.Items = append(g.Items, nativeMenuItem{Label: cmd.Label, Command: id, Disabled: s.commands.Enabled(id) != nil})
						found = true
						break
					}
				}
				if !found {
					menu = nil
					return
				}
			}
			menu = append(menu, g)
		}
	})
	if err != nil {
		return err
	}
	if len(groups) > 0 && menu == nil {
		return fmt.Errorf("cogent: menu label or command invalid")
	}
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return nativeSetMenu(scope, menu, s)
}

// Native application menus are process-wide. Activation uses the application
// lifetime and rechecks command availability under the current UI owner.
func (s *platformServices) invokeMenu(id string) {
	err := s.owner.Post(func() {
		if s.commands == nil {
			return
		}
		if err := s.commands.Invoke(s.owner.Context(), id); err != nil && s.onError != nil {
			s.onError(err)
		}
		s.owner.Invalidate("")
	})
	if err != nil && s.onError != nil {
		go func() { s.owner.Do(func() { s.onError(err) }) }()
	}
}
func (s *platformServices) shutdown() {
	if s.stopped.CompareAndSwap(false, true) {
		nativeShutdown()
	}
}
