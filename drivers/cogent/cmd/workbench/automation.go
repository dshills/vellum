package main

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/automation"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
)

func componentIndex(source *partSource, id string) (int, error) {
	if !strings.HasPrefix(id, "component:") {
		return 0, errors.New("invalid component identity")
	}
	n, err := strconv.Atoi(strings.TrimPrefix(id, "component:"))
	if err != nil || n < 1 || n > source.Len() || id != "component:"+strconv.Itoa(n) {
		return 0, errors.New("invalid component identity")
	}
	return n - 1, nil
}

// Explicit application DTOs, rather than native frame/AX state.
func workbenchAutomation(owner *app.Application, commands *command.Registry, source *partSource, refs *fileReferences, selected func() (int, uint64), selectID func(string) error) (*automation.Registry, error) {
	type stamp struct {
		source, selection, references uint64
		enabled                       uint8
	}
	var previous stamp
	generation := uint64(1)
	ids := []string{"component.rotate", "component.raise-value", "edit.undo", "edit.redo"}
	current := func() stamp {
		_, selection := selected()
		s := stamp{source: source.Revision(), selection: selection, references: refs.revision}
		for i, id := range ids {
			if commands.Enabled(id) == nil {
				s.enabled |= 1 << i
			}
		}
		return s
	}
	previous = current()
	r := automation.NewRegistry(func() (uint64, error) {
		if generation == math.MaxUint64 {
			return 0, history.ErrRevisionExhausted
		}
		if next := current(); next != previous {
			generation++
			previous = next
			if generation == math.MaxUint64 {
				return 0, history.ErrRevisionExhausted
			}
		}
		return generation, nil
	})
	type componentQuery struct {
		ID string `json:"id"`
	}
	if err := automation.RegisterQuery(r, "component", func(ctx context.Context, q componentQuery) (any, error) {
		index, _ := selected()
		var err error
		if q.ID != "" {
			index, err = componentIndex(source, q.ID)
			if err != nil {
				return nil, err
			}
		}
		if index < 0 || index >= source.Len() {
			return nil, errors.New("no selected component")
		}
		row := source.Row(index)
		return struct {
			ID        string `json:"id"`
			Reference string `json:"reference"`
			Value     uint32 `json:"value"`
			Package   string `json:"package"`
			Rotation  uint16 `json:"rotation"`
		}{row.ID, row.Cells[0], source.parts[index].Value, row.Cells[2], source.parts[index].Rotation}, nil
	}); err != nil {
		return nil, err
	}
	if err := automation.RegisterQuery(r, "workbench", func(context.Context, struct{}) (any, error) {
		index, _ := selected()
		type availability struct {
			ID      string `json:"id"`
			Enabled bool   `json:"enabled"`
		}
		states := make([]availability, len(ids))
		for i, id := range ids {
			states[i] = availability{id, commands.Enabled(id) == nil}
		}
		return struct {
			Selected     string         `json:"selected"`
			Components   int            `json:"components"`
			DocumentPath string         `json:"document_path"`
			References   int            `json:"references"`
			Commands     []availability `json:"commands"`
		}{source.Row(index).ID, source.Len(), "[redacted]", len(refs.paths), states}, nil
	}); err != nil {
		return nil, err
	}
	if err := automation.RegisterAction(r, "component.select", func(ctx context.Context, q componentQuery) (any, error) {
		if q.ID == "" {
			return nil, errors.New("component identity required")
		}
		if err := selectID(q.ID); err != nil {
			return nil, err
		}
		return struct{}{}, owner.Invalidate("")
	}); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if err := automation.RegisterAction(r, id, func(ctx context.Context, _ struct{}) (any, error) {
			if err := commands.Invoke(ctx, id); err != nil {
				return nil, err
			}
			return struct{}{}, owner.Invalidate("")
		}); err != nil {
			return nil, err
		}
	}
	return r, nil
}
