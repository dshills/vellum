package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"maps"
	"strconv"
	"strings"
	"unicode/utf8"
)

type batchRenameDraft struct {
	snapshot               selectionSnapshot
	prefix, start, message string
	invalid                [2]bool
}

func (m *model) bindBatchRename() error {
	return m.commands.Register(command.Command{ID: "batch-rename", Label: "Batch rename selected…", Available: m.primarySelectionAvailable, Execute: func(context.Context) error { return m.beginBatchRename() }})
}
func (m *model) beginBatchRename() error {
	if err := m.primarySelectionAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	m.batchRename = &batchRenameDraft{snapshot: m.selectionSnapshot(), prefix: "Component", start: "1"}
	m.prompt = "batch-rename"
	m.modalEpoch++
	return nil
}
func (m *model) batchNames(d *batchRenameDraft) (map[string]string, []string, error) {
	if d.invalid[0] || d.invalid[1] {
		return nil, nil, errors.New("input exceeded its byte limit and was shortened; edit that field before applying")
	}
	prefix := strings.TrimSpace(d.prefix)
	if prefix == "" || !validLayerName(prefix) {
		return nil, nil, errors.New("enter a prefix without controls; final names must fit 80 characters/128 bytes")
	}
	start, err := strconv.Atoi(d.start)
	if err != nil || start < 0 || start > 1000000 {
		return nil, nil, errors.New("start number must be an integer from 0 to 1000000")
	}
	units, _ := m.selectionUnits()
	if len(units) == 0 {
		return nil, nil, errNoSelection
	}
	names := map[string]string{}
	preview := []string{}
	for i, u := range units {
		name := fmt.Sprintf("%s %d", prefix, start+i)
		if !validLayerName(name) {
			return nil, nil, errors.New("a numbered name exceeds 80 characters/128 bytes")
		}
		names[u.ID] = name
		if i < 8 {
			preview = append(preview, m.layerLabel(u.ID)+" → "+name)
		}
	}
	if len(units) > 8 {
		preview = append(preview, fmt.Sprintf("… and %d more", len(units)-8))
	}
	return names, preview, nil
}
func (m *model) applyBatchRename(d *batchRenameDraft) error {
	if d == nil || m.batchRename != d || !d.snapshot.current(m) {
		return history.ErrStale
	}
	if err := m.primarySelectionAvailable(); err != nil {
		return err
	}
	names, _, err := m.batchNames(d)
	if err != nil {
		return err
	}
	info := maps.Clone(m.layerInfo)
	if info == nil {
		info = map[string]layerMetadata{}
	}
	for id, name := range names {
		v := info[id]
		v.Name = name
		info[id] = v
	}
	if err = m.applyLayerProperties("Batch rename layers", prunedLayerInfo(info, m.scene.Objects, m.groups)); err != nil {
		return err
	}
	m.status = fmt.Sprintf("Renamed %d selected units", len(names))
	return nil
}
func (m *model) batchRenameDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.batchRename
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.batchRename == d && m.prompt == "batch-rename"
	}
	cancel := func() {
		if active() {
			m.batchRename = nil
			m.prompt = ""
			_ = a.Invalidate("placement")
		}
	}
	prefix, start, message, preview := "", "", "", ""
	valid := false
	if d != nil {
		prefix, start, message = d.prefix, d.start, d.message
		_, lines, err := m.batchNames(d)
		if err != nil {
			message = err.Error()
		} else {
			valid = true
			preview = strings.Join(lines, "\n")
		}
	}
	field := func(id, label, value string, prefixField bool) ui.Node {
		return ui.Node{ID: id, Kind: ui.TextField, Text: value, Accessibility: ui.Accessibility{Name: label}, Change: func(e ui.ChangeEvent) {
			if active() {
				limit := 16
				if prefixField {
					limit = 128
				}
				idx := 1
				if prefixField {
					idx = 0
				}
				value := e.Text
				d.invalid[idx] = len(value) > limit
				if d.invalid[idx] {
					value = value[:limit]
					for !utf8.ValidString(value) && len(value) > 0 {
						value = value[:len(value)-1]
					}
				}
				if prefixField {
					d.prefix = value
				} else {
					d.start = value
				}
				d.message = ""
				_ = a.Invalidate("placement")
			}
		}}
	}
	return ui.Node{ID: "batch-rename-dialog", Kind: ui.Dialog, Text: "Batch rename selected", Open: active(), Dismiss: cancel, Children: []ui.Node{
		{ID: "batch-rename-help", Kind: ui.Label, Text: "Number selected units in back-to-front paint order. Each group receives one group name; member names and IDs stay intact. Preview shows up to eight names. Apply is one Undo step.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 500})}},
		field("batch-prefix", "Name prefix", prefix, true), field("batch-start", "Starting number", start, false),
		{ID: "batch-preview", Kind: ui.Label, Text: preview, Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 500})}}, {ID: "batch-error", Kind: ui.Label, Text: message},
		{ID: "batch-apply", Kind: ui.Button, Text: "Rename selected", Disabled: !valid, Activate: func() {
			if !active() {
				return
			}
			m.prompt = ""
			if err := m.applyBatchRename(d); err != nil {
				m.status = err.Error()
				if errors.Is(err, history.ErrStale) {
					m.batchRename = nil
				} else {
					d.message = err.Error()
					m.prompt = "batch-rename"
				}
			} else {
				m.batchRename = nil
			}
			_ = a.Invalidate("placement")
		}},
		{ID: "batch-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel}}}
}
