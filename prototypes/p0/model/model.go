// Package model is the identical, toolkit-independent P0 fixture.
// Session methods belong to the UI owner. Workers receive copied input.
package model

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

type Row struct {
	ID        int `table:"-"`
	Reference string
	Value     string
	Package   string
	Rotation  int
	Net       string
	X         float32
	Y         float32
}

type Document struct {
	Name       string
	Rows       []Row
	Revision   uint64
	SelectedID int
}

type Command struct{ ID, Label string }

var Commands = []Command{
	{"component.rotate", "Rotate selected 90°"},
	{"component.set_value", "Apply inspector value"},
	{"view.sort_reference", "Sort by reference"},
	{"view.sort_value", "Sort by value"},
	{"analysis.start", "Calculate board checksum"},
	{"analysis.cancel", "Cancel calculation"},
}

type Task struct {
	progress atomic.Int64
	cancel   context.CancelFunc
	done     chan Result
}

type Result struct {
	Document int
	Revision uint64
	Checksum uint64
	Err      error
}

type Session struct {
	Documents    []Document
	Active       int
	View         []Row // A projection: toolkit sorting must not mutate domain rows.
	Filter       string
	SortBy       string
	Task         *Task
	Status       string
	LastChecksum uint64
	WorkDelay    time.Duration
}

func New(count int) *Session {
	s := &Session{SortBy: "reference", Status: "Ready", WorkDelay: 5 * time.Millisecond}
	for d := 0; d < 2; d++ {
		doc := Document{Name: fmt.Sprintf("Board %c", 'A'+d), SelectedID: 17}
		doc.Rows = make([]Row, count)
		values := []string{"1k", "4.7k", "10k", "100n", "1u"}
		for i := range doc.Rows {
			id := i + 1
			doc.Rows[i] = Row{ID: id, Reference: fmt.Sprintf("R%d", id), Value: values[(i+d)%len(values)], Package: "0603", Rotation: (i % 4) * 90, Net: fmt.Sprintf("NET_%02d", i%32), X: float32(i % 100), Y: float32((i / 100) % 100)}
		}
		s.Documents = append(s.Documents, doc)
	}
	s.RebuildView()
	return s
}

func (s *Session) Doc() *Document { return &s.Documents[s.Active] }
func (s *Session) Selected() Row  { return s.Doc().Rows[s.Doc().SelectedID-1] }
func (s *Session) Select(id int) error {
	if id < 1 || id > len(s.Doc().Rows) {
		return errors.New("unknown component")
	}
	s.Doc().SelectedID = id
	return nil
}
func (s *Session) Switch(index int) {
	if index >= 0 && index < len(s.Documents) {
		s.Active = index
		s.RebuildView()
	}
}
func (s *Session) RebuildView() {
	q := strings.ToLower(strings.TrimSpace(s.Filter))
	s.View = s.View[:0]
	for _, r := range s.Doc().Rows {
		if q == "" || strings.Contains(strings.ToLower(r.Reference+" "+r.Value+" "+r.Net), q) {
			s.View = append(s.View, r)
		}
	}
	sort.Slice(s.View, func(i, j int) bool {
		a, b := s.View[i], s.View[j]
		if s.SortBy == "value" && a.Value != b.Value {
			return a.Value < b.Value
		}
		return a.ID < b.ID
	})
}
func (s *Session) SetFilter(q string) { s.Filter = q; s.RebuildView() }
func (s *Session) Enabled(id string) bool {
	switch id {
	case "analysis.start":
		return s.Task == nil
	case "analysis.cancel":
		return s.Task != nil
	case "component.rotate", "component.set_value", "view.sort_reference", "view.sort_value":
		return true
	}
	return false
}
func (s *Session) Execute(id, value string, wake func()) error {
	if !s.Enabled(id) {
		return fmt.Errorf("command unavailable: %s", id)
	}
	switch id {
	case "component.rotate":
		r := &s.Doc().Rows[s.Doc().SelectedID-1]
		r.Rotation = (r.Rotation + 90) % 360
		s.Doc().Revision++
		s.RebuildView()
	case "component.set_value":
		value = strings.TrimSpace(value)
		if value == "" {
			return errors.New("value must not be empty")
		}
		s.Doc().Rows[s.Doc().SelectedID-1].Value = value
		s.Doc().Revision++
		s.RebuildView()
	case "view.sort_reference":
		s.SortBy = "reference"
		s.RebuildView()
	case "view.sort_value":
		s.SortBy = "value"
		s.RebuildView()
	case "analysis.start":
		s.start(wake)
		return nil
	case "analysis.cancel":
		s.Task.cancel()
		s.Status = "Cancelling…"
		return nil
	}
	s.Status = "Executed " + id
	return nil
}
func (s *Session) start(wake func()) {
	ctx, cancel := context.WithCancel(context.Background())
	t := &Task{cancel: cancel, done: make(chan Result, 1)}
	input := append([]Row(nil), s.Doc().Rows...)
	doc, rev, delay := s.Active, s.Doc().Revision, s.WorkDelay
	s.Task = t
	s.Status = "Calculating…"
	go func() {
		result := Result{Document: doc, Revision: rev}
		defer func() {
			if p := recover(); p != nil {
				result.Err = fmt.Errorf("worker panic: %v", p)
			}
			t.done <- result
			if wake != nil {
				wake()
			}
		}()
		for i, r := range input {
			if i%256 == 0 {
				if err := ctx.Err(); err != nil {
					result.Err = err
					return
				}
				t.progress.Store(int64(i * 100 / len(input)))
				if wake != nil {
					wake()
				}
				if delay > 0 {
					timer := time.NewTimer(delay)
					select {
					case <-ctx.Done():
						timer.Stop()
						result.Err = ctx.Err()
						return
					case <-timer.C:
					}
				}
			}
			result.Checksum = result.Checksum*31 + uint64(r.ID+r.Rotation) + uint64(len(r.Value))
		}
		t.progress.Store(100)
	}()
}
func (s *Session) Poll() bool {
	if s.Task == nil {
		return false
	}
	select {
	case r := <-s.Task.done:
		s.Task.cancel()
		s.Task = nil
		if r.Err != nil {
			s.Status = "Calculation: " + r.Err.Error()
		} else if s.Documents[r.Document].Revision != r.Revision {
			s.Status = "Discarded stale calculation"
		} else {
			s.LastChecksum = r.Checksum
			s.Status = fmt.Sprintf("Checksum %016x (%s)", r.Checksum, s.Documents[r.Document].Name)
		}
		return true
	default:
		return false
	}
}
func (s *Session) Progress() int {
	if s.Task == nil {
		return 0
	}
	return int(s.Task.progress.Load())
}
func (s *Session) Close() {
	if s.Task != nil {
		s.Task.cancel()
	}
}

// TreeRows is deliberately a bounded navigation tree, not a million-node UI.
func (s *Session) TreeRows() []Row {
	end := min(64, len(s.View))
	return s.View[:end]
}

// CanvasRows returns a fixed 12x8 board view; selection is stable across sorting.
func (s *Session) CanvasRows() []Row { return s.Doc().Rows[:min(96, len(s.Doc().Rows))] }
