package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	vdoc "github.com/dshills/vellum/document"
	"github.com/dshills/vellum/viewport"
	"io"
	"reflect"
)

type bodyClipboard struct {
	Format  string
	Version int
	Objects []viewport.Object
}

func (m *workbench) copiedBodies() ([]viewport.Object, error) {
	var out []viewport.Object
	ids := m.selectionSet()
	for _, o := range m.state.Objects {
		if ids[o.ID] {
			out = append(out, o)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("select bodies to copy")
	}
	return out, nil
}
func encodeBodies(objects []viewport.Object) (string, error) {
	b, e := json.Marshal(bodyClipboard{"VellumForgeSelection", 1, objects})
	if e != nil {
		return "", e
	}
	if _, e = decodeBodies(string(b)); e != nil {
		return "", e
	}
	return string(b), nil
}
func decodeBodies(text string) ([]viewport.Object, error) {
	if len(text) > studyLimit {
		return nil, vdoc.ErrTooLarge
	}
	var c bodyClipboard
	d := json.NewDecoder(bytes.NewBufferString(text))
	d.DisallowUnknownFields()
	if e := d.Decode(&c); e != nil {
		return nil, e
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return nil, fmt.Errorf("trailing clipboard data")
	}
	if c.Format != "VellumForgeSelection" || c.Version != 1 || len(c.Objects) == 0 || len(c.Objects) > 48 {
		return nil, fmt.Errorf("not a Forge selection")
	}
	s := viewport.Scene{Camera: viewport.DefaultCamera(), Objects: c.Objects}
	if e := s.Validate(); e != nil {
		return nil, e
	}
	for _, o := range c.Objects {
		if o.Overlay || o.ScreenSize != 0 {
			return nil, fmt.Errorf("clipboard contains overlay")
		}
	}
	return c.Objects, nil
}
func (m *workbench) insertBodies(objects []viewport.Object, delta viewport.Vec3, label string) error {
	if _, e := encodeBodies(objects); e != nil {
		return e
	}
	if len(m.state.Objects)+len(objects) > 48 || len(m.state.Features)+len(objects) > 96 || m.next+uint64(len(objects)) > 1<<53 {
		return fmt.Errorf("study limit reached")
	}
	s := clone(m.state)
	s.Selection = nil
	next := m.next
	for _, o := range objects {
		o.ID = fmt.Sprintf("body-%d", next)
		next++
		o.Bounds.Min = o.Bounds.Min.Add(delta)
		o.Bounds.Max = o.Bounds.Max.Add(delta)
		s.Objects = append(s.Objects, o)
		s.Features = append(s.Features, feature{ID: "create-" + o.ID, Label: o.Name, Object: o.ID})
		s.Selection = append(s.Selection, o.ID)
		s.Selected = o.ID
	}
	s.Marker = len(s.Features)
	old := m.next
	m.next = next
	if e := m.apply(label, s); e != nil {
		m.next = old
		return e
	}
	return nil
}
func (m *workbench) duplicateBodies() error {
	objects, e := m.copiedBodies()
	if e != nil {
		return e
	}
	return m.insertBodies(objects, vec(.25, 0, .25), "Duplicate bodies")
}
func (m *workbench) clipboardReady() error {
	if m.clipboardBusy || m.drag != nil || m.prompt != "" || (m.files != nil && m.files.State().Busy) {
		return vdoc.ErrBusy
	}
	return nil
}
func (m *workbench) bindClipboardCommands() {
	m.register("copy", "Copy", "Meta+C", func() error {
		if e := m.clipboardReady(); e != nil {
			return e
		}
		_, e := m.copiedBodies()
		return e
	}, func() error { return m.beginClipboard(false) })
	m.register("paste", "Paste", "Meta+V", m.clipboardReady, func() error { return m.beginClipboard(true) })
	m.register("duplicate", "Duplicate", "Meta+D", func() error {
		if e := m.clipboardReady(); e != nil {
			return e
		}
		_, e := m.copiedBodies()
		return e
	}, m.duplicateBodies)
}
func (m *workbench) completePaste(before document, revision uint64, text string) error {
	if m.journal.Snapshot().Revision != revision || !reflect.DeepEqual(before, m.state) {
		return fmt.Errorf("study changed while reading clipboard; paste again")
	}
	objects, e := decodeBodies(text)
	if e != nil {
		return e
	}
	key := sha256.Sum256([]byte(text))
	count := 1
	if key == m.pasteKey {
		count = m.pasteCount + 1
	}
	if count > 48 {
		return fmt.Errorf("paste offset limit; copy again")
	}
	e = m.insertBodies(objects, vec(float64(count)*.25, 0, float64(count)*.25), "Paste bodies")
	if e == nil {
		m.pasteKey = key
		m.pasteCount = count
	}
	return e
}
func (m *workbench) beginClipboard(paste bool) error {
	if e := m.clipboardReady(); e != nil {
		return e
	}
	if m.application == nil {
		return app.ErrUnavailable
	}
	services, e := m.application.Services()
	if e != nil {
		return e
	}
	before := clone(m.state)
	revision := m.journal.Snapshot().Revision
	var text string
	if !paste {
		objects, err := m.copiedBodies()
		if err != nil {
			return err
		}
		text, e = encodeBodies(objects)
		if e != nil {
			return e
		}
	}
	ticket, e := m.application.ReserveWindowCompletion(m.ref, nil)
	if e != nil {
		return e
	}
	ctx, cancel := context.WithCancel(m.ref.Context())
	m.clipboardBusy = true
	m.clipboardCancel = cancel
	go func() {
		defer cancel()
		var err error
		if paste {
			text, err = services.ReadClipboard(ctx)
		} else {
			err = services.WriteClipboard(ctx, text)
		}
		publishErr := ticket.Publish(func() {
			m.clipboardBusy = false
			m.clipboardCancel = nil
			if err == nil && paste {
				err = m.completePaste(before, revision, text)
			}
			if err == nil {
				if paste {
					m.status = "Pasted bodies"
				} else {
					m.pasteKey = [32]byte{}
					m.pasteCount = 0
					m.status = "Copied selection"
				}
			}
			m.result(err)
		})
		if publishErr != nil {
			_ = ticket.Abort()
		} // Reservation can be revoked only by window/application retirement; no live owner remains.
	}()
	return nil
}
