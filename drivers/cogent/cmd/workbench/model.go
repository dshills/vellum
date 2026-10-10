package main

import (
	"bytes"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type part struct {
	Value             uint32
	Rotation, Package uint16
}
type partSource struct {
	parts    []part
	revision uint64
}

func newParts(count int) *partSource {
	s := &partSource{parts: make([]part, count), revision: 1}
	for i := range s.parts {
		s.parts[i] = part{Value: uint32(100 + i%997), Rotation: uint16(i % 4 * 90), Package: uint16(i % 2)}
	}
	return s
}
func (s *partSource) Len() int { return len(s.parts) }
func (s *partSource) Row(i int) virtual.Row {
	p := s.parts[i]
	pkg := "0603"
	if p.Package != 0 {
		pkg = "0805"
	}
	return virtual.Row{ID: "component:" + strconv.Itoa(i+1), Cells: []string{fmt.Sprintf("R%07d", i+1), strconv.Itoa(int(p.Value)) + " Ω", pkg, strconv.Itoa(int(p.Rotation)) + "°"}}
}
func (s *partSource) Less(i, j, column int) bool {
	a, b := s.parts[i], s.parts[j]
	switch column {
	case 1:
		return a.Value < b.Value
	case 2:
		return a.Package < b.Package
	case 3:
		return a.Rotation < b.Rotation
	default:
		return i < j
	}
}
func (s *partSource) Matches(i int, q string) bool {
	var storage [128]byte
	dst := storage[:0]
	dst = append(dst, 'r')
	n := i + 1
	digits := 1
	for x := n; x >= 10; x /= 10 {
		digits++
	}
	for j := digits; j < 7; j++ {
		dst = append(dst, '0')
	}
	dst = strconv.AppendInt(dst, int64(n), 10)
	dst = append(dst, ' ')
	p := s.parts[i]
	dst = strconv.AppendInt(dst, int64(p.Value), 10)
	dst = append(dst, []byte(" ω ")...)
	if p.Package == 0 {
		dst = append(dst, []byte("0603 ")...)
	} else {
		dst = append(dst, []byte("0805 ")...)
	}
	dst = strconv.AppendInt(dst, int64(p.Rotation), 10)
	dst = append(dst, []byte("°")...)
	return bytes.Contains(dst, []byte(q))
}
func (s *partSource) Snapshot() *partSource {
	return &partSource{parts: append([]part(nil), s.parts...), revision: s.revision}
}

func (s *partSource) Revision() uint64 { return s.revision }

// partEditor records sparse deltas rather than copying the million-row document.
// All mutations, including replay, advance the source revision so workers and
// table projections cannot install results from an older accepted model.
type partEditor struct {
	source  *partSource
	table   *virtual.Table
	board   *canvas.Scene
	journal *history.Journal
}

func (e *partEditor) change(index int, label string, next part) error {
	if index < 0 || index >= e.source.Len() {
		return fmt.Errorf("component index out of range")
	}
	before := e.source.parts[index]
	if before == next {
		return nil
	}
	var oldRect, newRect canvas.Rect
	hasRect := e.board != nil && index < len(e.board.Objects)
	if hasRect {
		oldRect = e.board.Objects[index]
		newRect = oldRect
		if (before.Rotation/90)%2 != (next.Rotation/90)%2 {
			newRect.X += (newRect.Width - newRect.Height) / 2
			newRect.Y += (newRect.Height - newRect.Width) / 2
			newRect.Width, newRect.Height = newRect.Height, newRect.Width
		}
	}
	assign := func(p part, rect canvas.Rect) error {
		e.source.parts[index] = p
		if hasRect {
			e.board.Objects[index] = rect
		}
		e.source.revision++
		e.table.Reset()
		return nil
	}
	return e.journal.Apply(history.Change{Label: label,
		Apply:  func() error { return assign(next, newRect) },
		Revert: func() error { return assign(before, oldRect) }})
}

// File references share document history; importing file contents is a separate
// application policy. Journal captures and retained values never alias callers.
type fileReferences struct{ paths []string }

func (r *fileReferences) replace(journal *history.Journal, paths []string) error {
	after, err := ui.CopyFileDropPaths(paths)
	if err != nil {
		return err
	}
	if slices.Equal(r.paths, after) {
		return nil
	}
	before := slices.Clone(r.paths)
	assign := func(paths []string) error { r.paths = slices.Clone(paths); return nil }
	return journal.Apply(history.Change{Label: "Reference files", Apply: func() error { return assign(after) }, Revert: func() error { return assign(before) }})
}

// Keep the chrome small even for a maximum-size drop. The complete paths stay
// in the document, independently of this abbreviated display.
func (r *fileReferences) summary() string {
	if len(r.paths) == 0 {
		return "Drop files on Board A to reference them"
	}
	names := make([]string, 0, min(3, len(r.paths)))
	for _, path := range r.paths[:min(3, len(r.paths))] {
		name := filepath.Base(path)
		count := 0
		for index := range name {
			if count == 64 {
				name = name[:index] + "…"
				break
			}
			count++
		}
		names = append(names, name)
	}
	if len(r.paths) > len(names) {
		names = append(names, "…")
	}
	return fmt.Sprintf("Referenced %d files: %s", len(r.paths), strings.Join(names, " · "))
}
