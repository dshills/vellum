// Package command routes UI entry points through one validated action registry.
package command

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// ShortcutScope controls native keyboard precedence. WindowShortcuts run before
// focused widgets; AfterFocusShortcuts run only when focused widgets decline.
// Explicit Invoke and Shortcut calls always use normal availability validation.
type ShortcutScope uint8

const (
	WindowShortcuts ShortcutScope = iota
	AfterFocusShortcuts
)

type Command struct {
	ID, Label, Shortcut string
	ShortcutScope       ShortcutScope
	// Available returns a disabled reason. Both presentation and execution check it.
	Available func() error
	Execute   func(context.Context) error
}

// Registry is owned by the application UI owner, including availability callbacks.
type Registry struct{ entries map[string]Command }

func New() *Registry { return &Registry{entries: map[string]Command{}} }
func (r *Registry) Register(c Command) error {
	if c.ShortcutScope != WindowShortcuts && c.ShortcutScope != AfterFocusShortcuts {
		return fmt.Errorf("command: invalid shortcut scope")
	}
	if c.ID == "" || c.Label == "" || c.Execute == nil {
		return fmt.Errorf("command: ID, label and action required")
	}
	if _, ok := r.entries[c.ID]; ok {
		return fmt.Errorf("command: duplicate %q", c.ID)
	}
	for _, old := range r.entries {
		if c.Shortcut != "" && strings.EqualFold(old.Shortcut, c.Shortcut) {
			return fmt.Errorf("command: duplicate shortcut %q", c.Shortcut)
		}
	}
	r.entries[c.ID] = c
	return nil
}
func (r *Registry) Enabled(id string) error {
	c, ok := r.entries[id]
	if !ok {
		return fmt.Errorf("command: unknown %q", id)
	}
	if c.Available != nil {
		return c.Available()
	}
	return nil
}
func (r *Registry) Invoke(ctx context.Context, id string) error {
	if ctx == nil {
		return fmt.Errorf("command: context required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.Enabled(id); err != nil {
		return err
	}
	return r.entries[id].Execute(ctx)
}
func (r *Registry) Search(query string) []Command {
	var out []Command
	q := strings.ToLower(strings.TrimSpace(query))
	for _, c := range r.entries {
		if strings.Contains(strings.ToLower(c.ID+" "+c.Label), q) {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (r *Registry) Shortcut(ctx context.Context, chord string) (bool, error) {
	return r.shortcut(ctx, chord, nil)
}

// ShortcutInScope routes one native keyboard phase; it never invokes a command
// registered in the other phase. Shortcut uniqueness spans both scopes.
func (r *Registry) ShortcutInScope(ctx context.Context, chord string, scope ShortcutScope) (bool, error) {
	if scope != WindowShortcuts && scope != AfterFocusShortcuts {
		return false, fmt.Errorf("command: invalid shortcut scope")
	}
	return r.shortcut(ctx, chord, &scope)
}

func (r *Registry) shortcut(ctx context.Context, chord string, scope *ShortcutScope) (bool, error) {
	for _, c := range r.entries {
		if (scope == nil || c.ShortcutScope == *scope) && c.Shortcut != "" && strings.EqualFold(chord, c.Shortcut) {
			return true, r.Invoke(ctx, c.ID)
		}
	}
	return false, nil
}
