package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
)

type copiedComponent struct {
	Kind   string  `json:"kind"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Group  int     `json:"group,omitempty"`
}
type selectionClipboard struct {
	Format  string            `json:"format"`
	Version int               `json:"version"`
	Objects []copiedComponent `json:"objects"`
}
type clipboardOperation struct {
	cancel context.CancelFunc
	ticket *app.Completion
}

func (m *model) clipboardAvailable() error {
	if m.fileBusy() || m.move != nil || m.prompt != "" || m.palette.State().Active {
		return document.ErrBusy
	}
	return nil
}
func (m *model) copiedSelection() ([]copiedComponent, error) {
	if err := m.editAvailable(); err != nil {
		return nil, err
	}
	ids := m.selectedIDs()
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	objects := make([]copiedComponent, 0, len(ids))
	tags := make(map[string]int)
	for _, o := range m.scene.Objects {
		if wanted[o.ID] {
			kind, _, _ := strings.Cut(o.ID, "-")
			tag := 0
			if group := m.groups[o.ID]; group != "" {
				if tags[group] == 0 {
					tags[group] = len(tags) + 1
				}
				tag = tags[group]
			}
			objects = append(objects, copiedComponent{Kind: kind, X: o.X, Y: o.Y, Width: o.Width, Height: o.Height, Group: tag})
		}
	}
	return objects, nil
}
func decodeSelection(text string) ([]copiedComponent, error) {
	if len(text) > drawingByteLimit {
		return nil, document.ErrTooLarge
	}
	var f selectionClipboard
	d := json.NewDecoder(bytes.NewBufferString(text))
	d.DisallowUnknownFields()
	if err := d.Decode(&f); err != nil {
		return nil, errors.New("clipboard does not contain Vellum components")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, errors.New("clipboard contains trailing data")
	}
	if f.Format != "vellum-placement-selection" || f.Version != 1 || len(f.Objects) == 0 || len(f.Objects) > 512 {
		return nil, errors.New("unsupported Vellum component clipboard")
	}
	for _, o := range f.Objects {
		if _, ok := componentByKind(o.Kind); !ok || !validPosition(o.X, o.Y) || !validSize(o.Width, o.Height) {
			return nil, errors.New("invalid clipboard component")
		}
	}
	if _, err := copiedGroupCounts(f.Objects); err != nil {
		return nil, err
	}
	return f.Objects, nil
}

func copiedGroupCounts(objects []copiedComponent) (map[int]int, error) {
	counts := make(map[int]int)
	for _, o := range objects {
		if o.Group < 0 || o.Group > 512 {
			return nil, errors.New("invalid clipboard group tag")
		}
		if o.Group > 0 {
			counts[o.Group]++
		}
	}
	for _, count := range counts {
		if count < 2 {
			return nil, errors.New("incomplete clipboard group")
		}
	}
	return counts, nil
}

func (m *model) insertCopies(objects []copiedComponent, offset float64, label string) error {
	if err := m.clipboardAvailable(); err != nil {
		return err
	}
	if len(objects) == 0 || len(objects) > 512-len(m.scene.Objects) || uint64(len(objects)) > math.MaxUint64-m.next || m.locked {
		return errors.New("cannot insert components: capacity, IDs or placement rejection")
	}
	counts, err := copiedGroupCounts(objects)
	if err != nil {
		return err
	}
	if uint64(len(counts)) > math.MaxUint64-m.nextGroup {
		return errors.New("group IDs exhausted")
	}
	before := slices.Clone(m.scene.Objects)
	groups := maps.Clone(m.groups)
	if groups == nil {
		groups = make(map[string]string)
	}
	groupIDs := make(map[int]string, len(counts))
	after, ids := slices.Clone(before), make([]string, 0, len(objects))
	for i, o := range objects {
		c, ok := componentByKind(o.Kind)
		if !ok || (m.rejectNotes && o.Kind == "note") || !validPosition(o.X+offset, o.Y+offset) || !validSize(o.Width, o.Height) {
			return errors.New("copied components rejected or outside drawing bounds")
		}
		id := fmt.Sprintf("%s-%d", o.Kind, m.next+uint64(i)+1)
		after = append(after, canvas.Rect{ID: id, X: o.X + offset, Y: o.Y + offset, Width: o.Width, Height: o.Height, Color: c.Color})
		ids = append(ids, id)
		if o.Group > 0 {
			g := groupIDs[o.Group]
			if g == "" {
				g = fmt.Sprintf("group-%d", m.nextGroup+uint64(len(groupIDs))+1)
				groupIDs[o.Group] = g
			}
			groups[id] = g
		}
	}
	err = m.applyDocument(label, after, groups, ids, ids[len(ids)-1])
	if err == nil {
		m.next += uint64(len(objects))
		m.nextGroup += uint64(len(groupIDs))
		m.status = fmt.Sprintf("%s: %d components", label, len(objects))
	}
	return err
}
func (m *model) duplicateSelected() error {
	objects, err := m.copiedSelection()
	if err != nil {
		return err
	}
	return m.insertCopies(objects, 20, "Duplicate selection")
}
func (m *model) pasteSelection(text string) error {
	objects, err := decodeSelection(text)
	if err != nil {
		return err
	}
	count := uint64(1)
	if text == m.pasteText {
		count = m.pasteCount + 1
	}
	if count > 1000 {
		return errors.New("paste offset exhausted; copy again")
	}
	if err = m.insertCopies(objects, float64(count)*20, "Paste selection"); err != nil {
		return err
	}
	m.pasteText, m.pasteCount = text, count
	return nil
}

func (m *model) bindClipboard(a *app.Application, ref app.WindowRef) error {
	for _, c := range []struct{ id, label, shortcut string }{{"copy", "Copy", "Meta+C"}, {"paste", "Paste", "Meta+V"}, {"duplicate", "Duplicate", "Meta+D"}} {
		available := m.editAvailable
		if c.id == "paste" {
			available = m.clipboardAvailable
		}
		if err := m.commands.Register(command.Command{ID: c.id, Label: c.label, Shortcut: c.shortcut, ShortcutScope: command.AfterFocusShortcuts, Available: available, Execute: func(context.Context) error {
			if c.id == "duplicate" {
				return m.duplicateSelected()
			}
			return m.startClipboard(a, ref, c.id == "paste")
		}}); err != nil {
			return err
		}
	}
	return nil
}
func (m *model) startClipboard(a *app.Application, ref app.WindowRef, paste bool) error {
	if err := m.clipboardAvailable(); err != nil {
		return err
	}
	var text string
	if !paste {
		objects, err := m.copiedSelection()
		if err != nil {
			return err
		}
		data, err := json.Marshal(selectionClipboard{"vellum-placement-selection", 1, objects})
		if err != nil {
			return err
		}
		if len(data) > drawingByteLimit {
			return document.ErrTooLarge
		}
		text = string(data)
	}
	service, err := a.Services()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ref.Context(), 5*time.Second)
	ticket, err := a.ReserveWindowCompletion(ref, nil)
	if err != nil {
		cancel()
		return err
	}
	job := &clipboardOperation{cancel, ticket}
	m.clipboardJob = job
	m.status = "Copying components…"
	if paste {
		m.status = "Reading component clipboard…"
	}
	go func() {
		defer cancel()
		var err error
		if paste {
			text, err = service.ReadClipboard(ctx)
		} else {
			err = service.WriteClipboard(ctx, text)
		}
		if paste && len(text) > drawingByteLimit {
			err = document.ErrTooLarge
		}
		if err == nil {
			err = ctx.Err()
		}
		if err != nil {
			text = ""
		}
		_ = ticket.Publish(func() {
			if m.clipboardJob != job {
				return
			}
			m.clipboardJob = nil
			if err == nil {
				if paste {
					err = m.pasteSelection(text)
				} else {
					m.pasteText, m.pasteCount = "", 0
					m.status = "Copied selected components"
				}
			}
			if err != nil {
				m.status = "Clipboard: " + err.Error()
			}
			_ = a.Invalidate("placement")
		})
	}()
	return nil
}
