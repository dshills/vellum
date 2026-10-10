// Package command routes UI entry points through one validated action registry.
package command

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type Command struct {
	ID, Label, Shortcut string
	// Available returns a disabled reason. Both presentation and execution check it.
	Available func() error
	Execute   func(context.Context) error
}

// Registry is owned by the application UI owner, including availability callbacks.
type Registry struct{ entries map[string]Command }

func New() *Registry { return &Registry{entries: map[string]Command{}} }
func (r *Registry) Register(c Command) error {
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
	for _, c := range r.entries {
		if c.Shortcut != "" && strings.EqualFold(chord, c.Shortcut) {
			return true, r.Invoke(ctx, c.ID)
		}
	}
	return false, nil
}
