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
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/document"
	"github.com/dshills/vellum/history"
)

type copiedComponent struct {
	Kind        string  `json:"kind"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	Group       int     `json:"group,omitempty"`
	Name        string  `json:"name,omitempty"`
	Hidden      bool    `json:"hidden,omitempty"`
	Locked      bool    `json:"locked,omitempty"`
	GroupName   string  `json:"group_name,omitempty"`
	GroupHidden bool    `json:"group_hidden,omitempty"`
	GroupLocked bool    `json:"group_locked,omitempty"`
}

func (o copiedComponent) componentMetadata() layerMetadata {
	return layerMetadata{Name: o.Name, Hidden: o.Hidden, Locked: o.Locked}
}
func (o copiedComponent) groupMetadata() layerMetadata {
	return layerMetadata{Name: o.GroupName, Hidden: o.GroupHidden, Locked: o.GroupLocked}
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

// Cut owns the same detached document/selection that was copied. It must not
// delete a later selection or a replacement drawing after clipboard IO returns.
type cutSelection struct {
	scene    *canvas.Scene
	journal  *history.Journal
	revision uint64
	document drawingFile
	ids      []string
	primary  string
}

func (m *model) captureCut() (*cutSelection, error) {
	if err := m.editAvailable(); err != nil {
		return nil, err
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		return nil, err
	}
	return &cutSelection{scene: m.scene, journal: m.journal, revision: m.journal.Snapshot().Revision,
		document: f, ids: slices.Clone(m.selectedIDs()), primary: m.scene.Selected}, nil
}

func (m *model) completeCut(cut *cutSelection) error {
	if err := m.editAvailable(); err != nil {
		return err
	}
	if m.scene != cut.scene || m.journal != cut.journal || m.journal.Snapshot().Revision != cut.revision ||
		m.scene.Selected != cut.primary || !slices.Equal(m.selectedIDs(), cut.ids) {
		return history.ErrStale
	}
	f, err := drawingCodec(m).Capture()
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(f, cut.document) {
		return history.ErrStale
	}
	selected := make(map[string]bool, len(cut.ids))
	for _, id := range cut.ids {
		selected[id] = true
	}
	after := make([]canvas.Rect, 0, len(m.scene.Objects)-len(cut.ids))
	for _, object := range m.scene.Objects {
		if !selected[object.ID] {
			after = append(after, object)
		}
	}
	if err := m.applyObjects("Cut selection", after, ""); err != nil {
		return err
	}
	m.pasteText, m.pasteCount = "", 0
	m.status = fmt.Sprintf("Cut %d selected components", len(cut.ids))
	return nil
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
			var groupInfo layerMetadata
			if group := m.groups[o.ID]; group != "" {
				if tags[group] == 0 {
					tags[group] = len(tags) + 1
				}
				tag = tags[group]
				groupInfo = m.layerInfo[group]
			}
			info := m.layerInfo[o.ID]
			objects = append(objects, copiedComponent{Kind: kind, X: o.X, Y: o.Y, Width: o.Width, Height: o.Height, Group: tag,
				Name: info.Name, Hidden: info.Hidden, Locked: info.Locked,
				GroupName: groupInfo.Name, GroupHidden: groupInfo.Hidden, GroupLocked: groupInfo.Locked})
		}
	}
	return objects, nil
}

func encodeSelection(objects []copiedComponent) (string, error) {
	if _, err := copiedGroupCounts(objects); err != nil {
		return "", err
	}
	version := 1
	for _, o := range objects {
		if o.componentMetadata() != (layerMetadata{}) || o.groupMetadata() != (layerMetadata{}) {
			version = 2
		}
	}
	data, err := json.Marshal(selectionClipboard{"vellum-placement-selection", version, objects})
	if err != nil {
		return "", err
	}
	if len(data) > drawingByteLimit {
		return "", document.ErrTooLarge
	}
	return string(data), nil
}

func decodeSelection(text string) ([]copiedComponent, error) {
	if len(text) > drawingByteLimit {
		return nil, document.ErrTooLarge
	}
	if !utf8.ValidString(text) {
		return nil, errors.New("invalid clipboard text encoding")
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
	if f.Format != "vellum-placement-selection" || (f.Version != 1 && f.Version != 2) || len(f.Objects) == 0 || len(f.Objects) > 512 {
		return nil, errors.New("unsupported Vellum component clipboard")
	}
	for _, o := range f.Objects {
		if _, ok := componentByKind(o.Kind); !ok || !validPosition(o.X, o.Y) || !validSize(o.Width, o.Height) {
			return nil, errors.New("invalid clipboard component")
		}
		if f.Version == 1 && (o.componentMetadata() != (layerMetadata{}) || o.groupMetadata() != (layerMetadata{})) {
			return nil, errors.New("clipboard layer properties require version 2")
		}
	}
	if _, err := copiedGroupCounts(f.Objects); err != nil {
		return nil, err
	}
	return f.Objects, nil
}

func copiedGroupCounts(objects []copiedComponent) (map[int]int, error) {
	counts := make(map[int]int)
	metadata := make(map[int]layerMetadata)
	for _, o := range objects {
		if o.Group < 0 || o.Group > 512 {
			return nil, errors.New("invalid clipboard group tag")
		}
		if !validLayerName(o.Name) || !validLayerName(o.GroupName) {
			return nil, errors.New("invalid clipboard layer name")
		}
		if o.Group == 0 && o.groupMetadata() != (layerMetadata{}) {
			return nil, errors.New("clipboard group properties require group membership")
		}
		if o.Group > 0 {
			if prior, seen := metadata[o.Group]; seen && prior != o.groupMetadata() {
				return nil, errors.New("inconsistent clipboard group properties")
			}
			metadata[o.Group] = o.groupMetadata()
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
	return m.insertCopiesOffset(objects, canvas.Point{X: offset, Y: offset}, label)
}
func (m *model) insertCopiesOffset(objects []copiedComponent, offset canvas.Point, label string) error {
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
	info := maps.Clone(m.layerInfo)
	if info == nil {
		info = make(map[string]layerMetadata)
	}
	after, ids := slices.Clone(before), make([]string, 0, len(objects))
	for i, o := range objects {
		c, ok := componentByKind(o.Kind)
		if !ok || (m.rejectNotes && o.Kind == "note") || !validPosition(o.X+offset.X, o.Y+offset.Y) || !validSize(o.Width, o.Height) {
			return errors.New("copied components rejected or outside drawing bounds")
		}
		id := fmt.Sprintf("%s-%d", o.Kind, m.next+uint64(i)+1)
		after = append(after, canvas.Rect{ID: id, X: o.X + offset.X, Y: o.Y + offset.Y, Width: o.Width, Height: o.Height, Color: c.Color})
		ids = append(ids, id)
		if value := o.componentMetadata(); value != (layerMetadata{}) {
			info[id] = value
		}
		if o.Group > 0 {
			g := groupIDs[o.Group]
			if g == "" {
				g = fmt.Sprintf("group-%d", m.nextGroup+uint64(len(groupIDs))+1)
				groupIDs[o.Group] = g
			}
			groups[id] = g
			if value := o.groupMetadata(); value != (layerMetadata{}) {
				info[g] = value
			}
		}
	}
	// Encode the actual proposed saved drawing, including escaped names and
	// fresh IDs/counters, before admitting any document or history mutation.
	prospective := *m
	scene := *m.scene
	prospective.scene, scene.Objects = &scene, after
	prospective.groups, prospective.layerInfo = groups, info
	prospective.next += uint64(len(objects))
	prospective.nextGroup += uint64(len(groupIDs))
	prospective.setSelection(ids, ids[len(ids)-1])
	f, err := drawingCodec(&prospective).Capture()
	if err != nil {
		return err
	}
	if _, err := drawingCodec(&prospective).Encode(f); err != nil {
		return err
	}
	err = m.applyDocumentMetadata(label, after, groups, info, ids, ids[len(ids)-1])
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
	for _, c := range []struct{ id, label, shortcut string }{{"copy", "Copy", "Meta+C"}, {"cut", "Cut", "Meta+X"}, {"paste", "Paste", "Meta+V"}, {"paste-center", "Paste at View Center", ""}, {"duplicate", "Duplicate", "Meta+D"}} {
		available := m.editAvailable
		if c.id == "paste" || c.id == "paste-center" {
			available = m.clipboardAvailable
		}
		if err := m.commands.Register(command.Command{ID: c.id, Label: c.label, Shortcut: c.shortcut, ShortcutScope: command.AfterFocusShortcuts, Available: available, Execute: func(context.Context) error {
			if c.id == "duplicate" {
				return m.duplicateSelected()
			}
			return m.startClipboardAction(a, ref, c.id)
		}}); err != nil {
			return err
		}
	}
	return nil
}
func (m *model) startClipboard(a *app.Application, ref app.WindowRef, paste bool) error {
	action := "copy"
	if paste {
		action = "paste"
	}
	return m.startClipboardAction(a, ref, action)
}

func (m *model) startClipboardAction(a *app.Application, ref app.WindowRef, action string) error {
	if err := m.clipboardAvailable(); err != nil {
		return err
	}
	paste := action == "paste" || action == "paste-center"
	center := canvas.Point{}
	if action == "paste-center" {
		if m.scene == nil || !finiteResize(m.scene.View.Center.X) || !finiteResize(m.scene.View.Center.Y) {
			return errors.New("invalid paste center")
		}
		center = m.scene.View.Center
	}
	var cut *cutSelection
	if action == "cut" {
		var err error
		cut, err = m.captureCut()
		if err != nil {
			return err
		}
	}
	var text string
	if !paste {
		objects, err := m.copiedSelection()
		if err != nil {
			return err
		}
		payload, err := encodeSelection(objects)
		if err != nil {
			return err
		}
		text = payload
	}
	service, err := a.Services()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ref.Context(), 5*time.Second)
	ticket, err := a.ReserveWindowCompletion(ref, func(error) { cancel() })
	if err != nil {
		cancel()
		return err
	}
	job := &clipboardOperation{cancel, ticket}
	m.clipboardJob = job
	m.status = "Copying components…"
	if paste {
		m.status = "Reading component clipboard…"
	} else if cut != nil {
		m.status = "Cutting components…"
	}
	go func() {
		// Cut keeps its context live through owner admission, so a canceled or
		// timed-out successful write cannot remove objects while queued.
		if cut == nil && action != "paste-center" {
			defer cancel()
		}
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
		publishErr := ticket.Publish(func() {
			defer cancel()
			if m.clipboardJob != job {
				return
			}
			m.clipboardJob = nil
			if err == nil && (cut != nil || action == "paste-center") {
				err = ctx.Err()
			}
			if err == nil {
				if paste {
					if action == "paste-center" {
						err = m.pasteSelectionAt(text, center)
					} else {
						err = m.pasteSelection(text)
					}
				} else if cut != nil {
					err = m.completeCut(cut)
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
		if publishErr != nil {
			cancel()
		}
	}()
	return nil
}

// Center the payload's complete bounds, independent of camera zoom and grid.
// insertion still validates the entire proposed drawing and allocates fresh IDs.
func (m *model) pasteSelectionAt(text string, center canvas.Point) error {
	if err := m.selectionAvailable(); err != nil {
		return err
	}
	objects, err := decodeSelection(text)
	if err != nil {
		return err
	}
	if !finiteResize(center.X) || !finiteResize(center.Y) {
		return errors.New("invalid paste center")
	}
	left, top, right, bottom := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, o := range objects {
		left = math.Min(left, o.X)
		top = math.Min(top, o.Y)
		right = math.Max(right, o.X+o.Width)
		bottom = math.Max(bottom, o.Y+o.Height)
	}
	offset := canvas.Point{X: center.X - (left+right)/2, Y: center.Y - (top+bottom)/2}
	return m.insertCopiesOffset(objects, offset, "Paste at view center")
}
