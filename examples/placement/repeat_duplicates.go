package main

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strconv"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
)

type repeatDraft struct {
	count, dx, dy, message string
	invalidInput           [3]bool
	scene                  *canvas.Scene
	revision               uint64
	document               drawingFile
	ids                    []string
	primary                string
}

func (m *model) repeatAvailable() error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	return m.editAvailable()
}

func (m *model) bindRepeatDuplicates() error {
	return m.commands.Register(command.Command{ID: "repeat-duplicates", Label: "Repeat duplicates", Available: m.repeatAvailable, Execute: func(context.Context) error { return m.beginRepeatDuplicates() }})
}

func (m *model) beginRepeatDuplicates() error {
	if err := m.repeatAvailable(); err != nil {
		return err
	}
	if m.modalEpoch == ^uint64(0) {
		return errors.New("dialog focus sequence exhausted")
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		return err
	}
	m.repeatDraft = &repeatDraft{count: "3", dx: "20", dy: "20", scene: m.scene, revision: m.journal.Snapshot().Revision, document: f, ids: slices.Clone(m.selectedIDs()), primary: m.scene.Selected}
	m.prompt = "repeat-duplicates"
	m.modalEpoch++
	return nil
}

// Repeat creates a single proposed insertion, so validation, fresh identities,
// group ownership and history all use the same atomic path as ordinary Duplicate.
func (m *model) repeatDuplicates(count int, dx, dy float64) error {
	if err := m.repeatAvailable(); err != nil {
		return err
	}
	if count < 1 || count > 100 || !finiteResize(dx) || !finiteResize(dy) || dx < -20000 || dx > 20000 || dy < -20000 || dy > 20000 {
		return errors.New("use 1–100 copies and finite X/Y offsets from -20000 to 20000")
	}
	objects, err := m.copiedSelection()
	if err != nil {
		return err
	}
	if len(objects) == 0 || len(objects) > (512-len(m.scene.Objects))/count {
		return errors.New("repeated copies exceed the 512-component drawing limit")
	}
	if _, err := copiedGroupCounts(objects); err != nil {
		return err
	}
	all := make([]copiedComponent, 0, len(objects)*count)
	nextTag := 0
	for n := 1; n <= count; n++ {
		tags := make(map[int]int)
		for _, original := range objects {
			o := original
			o.X += float64(n) * dx
			o.Y += float64(n) * dy
			if !validPosition(o.X, o.Y) {
				return errors.New("repeated copies exceed component origin limits")
			}
			if o.Group > 0 {
				if tags[o.Group] == 0 {
					nextTag++
					tags[o.Group] = nextTag
				}
				o.Group = tags[o.Group]
			}
			all = append(all, o)
		}
	}
	return m.insertCopiesOffset(all, canvas.Point{}, "Repeat duplicates")
}

func (m *model) applyRepeatDuplicates(d *repeatDraft) error {
	if d == nil || m.repeatDraft != d || m.scene != d.scene || m.journal.Snapshot().Revision != d.revision || !slices.Equal(m.selectedIDs(), d.ids) || m.scene.Selected != d.primary {
		return history.ErrStale
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(f, d.document) {
		return history.ErrStale
	}
	for _, invalid := range d.invalidInput {
		if invalid {
			return errors.New("numeric fields are limited to 64 characters; edit the invalid field")
		}
	}
	count, err := strconv.Atoi(d.count)
	if err != nil {
		return errors.New("copy count must be an integer from 1 to 100")
	}
	dx, err := strconv.ParseFloat(d.dx, 64)
	if err != nil {
		return errors.New("X offset must be a finite number")
	}
	dy, err := strconv.ParseFloat(d.dy, 64)
	if err != nil {
		return errors.New("Y offset must be a finite number")
	}
	return m.repeatDuplicates(count, dx, dy)
}

func (m *model) repeatDuplicatesDialog(a *app.Application, ref app.WindowRef) ui.Node {
	d := m.repeatDraft
	active := func() bool {
		return ref.Context().Err() == nil && d != nil && m.repeatDraft == d && m.prompt == "repeat-duplicates"
	}
	cancel := func() {
		if active() {
			m.repeatDraft, m.prompt = nil, ""
			_ = a.Invalidate("placement")
		}
	}
	values := []string{"3", "20", "20"}
	message := ""
	if d != nil {
		values, message = []string{d.count, d.dx, d.dy}, d.message
	}
	children := []ui.Node{{ID: "repeat-help", Kind: ui.Label, Text: "Add 1–100 copies of the selection. Copy n uses n × X/Y offset in world units. Each copied group is independent. Zero offsets intentionally overlap. Snapping is ignored; the entire operation is one Undo.", Style: ui.Style{MaxSize: ui.Some(ui.Size{Width: 440})}}}
	for i, field := range []struct{ id, label string }{{"repeat-count", "Copy count"}, {"repeat-dx", "X offset"}, {"repeat-dy", "Y offset"}} {
		children = append(children, ui.Node{ID: field.id, Kind: ui.TextField, Text: values[i], Accessibility: ui.Accessibility{Name: field.label}, Change: func(e ui.ChangeEvent) {
			if !active() {
				return
			}
			value := e.Text
			if len(value) > 64 {
				d.invalidInput[i] = true
				d.message = field.label + ": input over 64 characters was ignored; edit this field before Apply"
				_ = a.Invalidate("placement")
				return
			}
			d.invalidInput[i] = false
			d.message = ""
			for _, invalid := range d.invalidInput {
				if invalid {
					d.message = "numeric fields are limited to 64 characters; edit the invalid field"
				}
			}
			switch i {
			case 0:
				d.count = value
			case 1:
				d.dx = value
			case 2:
				d.dy = value
			}
			_ = a.Invalidate("placement")
		}})
	}
	children = append(children,
		ui.Node{ID: "repeat-error", Kind: ui.Label, Text: message},
		ui.Node{ID: "repeat-apply", Kind: ui.Button, Text: "Create copies", Activate: func() {
			if !active() {
				return
			}
			m.prompt = ""
			if err := m.applyRepeatDuplicates(d); err != nil {
				m.status = err.Error()
				if errors.Is(err, history.ErrStale) {
					m.repeatDraft = nil
				} else {
					d.message, m.prompt = err.Error(), "repeat-duplicates"
				}
			} else {
				m.repeatDraft = nil
			}
			_ = a.Invalidate("placement")
		}},
		ui.Node{ID: "repeat-cancel", Kind: ui.Button, Text: "Cancel", Activate: cancel})
	return ui.Node{ID: "repeat-dialog", Kind: ui.Dialog, Text: "Repeat duplicates", Open: active(), Dismiss: cancel, Children: children}
}
