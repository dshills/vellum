package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/dshills/vellum/canvas"
)

type drawingGroup struct {
	ID      string   `json:"id"`
	Members []string `json:"members"`
}

func (f drawingFile) groupMap(objects []canvas.Rect) (map[string]string, error) {
	if len(f.Groups) > 256 || (f.Version < 3 && (len(f.Groups) != 0 || f.NextGroup != 0)) {
		return nil, errors.New("invalid drawing group metadata")
	}
	valid, seen := make(map[string]bool, len(objects)), make(map[string]bool, len(f.Groups))
	for _, o := range objects {
		valid[o.ID] = true
	}
	groups := make(map[string]string)
	for _, g := range f.Groups {
		prefix, suffix, _ := strings.Cut(g.ID, "-")
		n, err := strconv.ParseUint(suffix, 10, 64)
		if prefix != "group" || err != nil || n == 0 || n > f.NextGroup || strconv.FormatUint(n, 10) != suffix || seen[g.ID] || len(g.Members) < 2 || len(g.Members) > 512 {
			return nil, errors.New("invalid drawing group")
		}
		seen[g.ID] = true
		for _, id := range g.Members {
			if !valid[id] || groups[id] != "" {
				return nil, errors.New("invalid or overlapping drawing group members")
			}
			groups[id] = g.ID
		}
	}
	return groups, nil
}

func (m *model) savedGroups() []drawingGroup {
	indices := make(map[string]int)
	groups := make([]drawingGroup, 0)
	for _, o := range m.scene.Objects {
		if group := m.groups[o.ID]; group != "" {
			index, ok := indices[group]
			if !ok {
				index = len(groups)
				indices[group] = index
				groups = append(groups, drawingGroup{ID: group})
			}
			groups[index].Members = append(groups[index].Members, o.ID)
		}
	}
	return groups
}
