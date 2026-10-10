// Package virtual projects large data models into bounded viewport records.
package virtual

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const MaxVisible = 256

type Row struct {
	ID    string
	Cells []string
}

// Source owns domain data. Row materializes only requested visible records.
type Source interface {
	Len() int
	Row(int) Row
}
type Table struct {
	generation            uint64
	source                Source
	sourceLength          int
	sourceRevision        uint64
	projectionInvalidated bool
	order                 []int
	start, count          int
	Selected              string
}

func (t *Table) StartRow() int     { t.syncSource(); return t.start }
func (t *Table) VisibleCount() int { return t.count }

// VersionedSource enables automatic invalidation after same-length edits.
type VersionedSource interface{ Revision() uint64 }

func NewTable(source Source, count int) (*Table, error) {
	if source == nil || source.Len() < 0 || count < 1 || count > MaxVisible {
		return nil, fmt.Errorf("virtual: source and 1..256 visible rows required")
	}
	table := &Table{generation: 1, source: source, sourceLength: source.Len(), count: count}
	if v, ok := source.(VersionedSource); ok {
		table.sourceRevision = v.Revision()
	}
	return table, nil
}
func (t *Table) Len() int {
	t.syncSource()
	if t.order != nil {
		return len(t.order)
	}
	return t.source.Len()
}
func (t *Table) SetViewport(start, count int) error {
	if count < 1 || count > MaxVisible || start < 0 {
		return fmt.Errorf("virtual: invalid viewport")
	}
	max := t.Len() - count
	if max < 0 {
		max = 0
	}
	if start > max {
		start = max
	}
	t.start = start
	t.count = count
	return nil
}
func (t *Table) View() []Row {
	length := t.Len()
	end := t.start + t.count
	if end > length {
		end = length
	}
	if t.start > end {
		return nil
	}
	out := make([]Row, 0, end-t.start)
	for i := t.start; i < end; i++ {
		index := i
		if t.order != nil {
			index = t.order[i]
		}
		out = append(out, t.source.Row(index))
	}
	return out
}

// Project scans domain rows on demand. Use a worker for expensive projections;
// apply only against the same immutable source revision. Indices cost O(N).
type Comparator interface{ Less(int, int, int) bool }
type Matcher interface{ Matches(int, string) bool }

type Projection struct {
	Indices   []int
	length    int
	revision  uint64
	versioned bool
}

// Sources used by workers must be immutable snapshots. For unversioned sources,
// callers must discard results after edits or replacement.
func Project(source Source, query string, column int, descending bool) (Projection, error) {
	return ProjectContext(context.Background(), source, query, column, descending)
}
func ProjectContext(ctx context.Context, source Source, query string, column int, descending bool) (Projection, error) {
	if ctx == nil || source == nil || source.Len() < 0 || column < 0 {
		return Projection{}, fmt.Errorf("virtual: context, source and column required")
	}
	if err := ctx.Err(); err != nil {
		return Projection{}, err
	}
	if source.Len() > 0 && column >= len(source.Row(0).Cells) {
		return Projection{}, fmt.Errorf("virtual: column absent")
	}
	result := Projection{length: source.Len()}
	if v, ok := source.(VersionedSource); ok {
		result.versioned = true
		result.revision = v.Revision()
	}
	q := strings.ToLower(query)
	out := make([]int, 0, source.Len())
	matcher, _ := source.(Matcher)
	less, _ := source.(Comparator)
	for i := 0; i < source.Len(); i++ {
		if i%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return Projection{}, err
			}
		}
		match := q == ""
		if !match {
			if matcher != nil {
				match = matcher.Matches(i, q)
			} else {
				r := source.Row(i)
				match = strings.Contains(strings.ToLower(strings.Join(r.Cells, " ")), q)
			}
		}
		if match {
			out = append(out, i)
		}
	}
	comparisons := 0
	canceled := false
	check := func() bool {
		comparisons++
		if comparisons%1024 == 0 {
			canceled = ctx.Err() != nil
		}
		return canceled
	}
	if less == nil {
		type item struct {
			Index int
			Key   string
		}
		items := make([]item, len(out))
		for i, index := range out {
			if i%1024 == 0 {
				if err := ctx.Err(); err != nil {
					return Projection{}, err
				}
			}
			row := source.Row(index)
			if column >= len(row.Cells) {
				return Projection{}, fmt.Errorf("virtual: column absent")
			}
			items[i] = item{index, row.Cells[column]}
		}
		sort.SliceStable(items, func(i, j int) bool {
			if check() {
				return false
			}
			if descending {
				return items[i].Key > items[j].Key
			}
			return items[i].Key < items[j].Key
		})
		for i, item := range items {
			out[i] = item.Index
		}
	} else {
		sort.SliceStable(out, func(i, j int) bool {
			if check() {
				return false
			}
			a, b := out[i], out[j]
			if descending {
				a, b = b, a
			}
			return less.Less(a, b, column)
		})
	}
	if err := ctx.Err(); err != nil {
		return Projection{}, err
	}
	result.Indices = out
	return result, nil
}
func (t *Table) ApplyProjection(projection Projection, token uint64) error {
	t.syncSource()
	if token != t.generation || projection.length != t.sourceLength || projection.versioned != isVersioned(t.source) || (projection.versioned && projection.revision != t.sourceRevision) {
		return fmt.Errorf("virtual: stale projection")
	}
	indices := projection.Indices
	var seen []uint64
	var sparse map[int]bool
	if t.source.Len() <= 16000000 {
		seen = make([]uint64, (t.source.Len()+63)/64)
	} else {
		sparse = make(map[int]bool, len(indices))
	}
	for _, i := range indices {
		if i < 0 || i >= t.source.Len() {
			return fmt.Errorf("virtual: invalid projection index")
		}
		if sparse != nil {
			if sparse[i] {
				return fmt.Errorf("virtual: duplicate projection index")
			}
			sparse[i] = true
		} else {
			word, bit := i/64, uint(i%64)
			if seen[word]&(1<<bit) != 0 {
				return fmt.Errorf("virtual: duplicate projection index")
			}
			seen[word] |= 1 << bit
		}
	}
	t.order = append([]int{}, indices...)
	t.projectionInvalidated = false
	return t.SetViewport(t.start, t.count)
}
func (t *Table) Reset() { t.order = nil; t.start = 0 }
func (t *Table) SelectVisible(index int) (Row, error) {
	rows := t.View()
	if index < 0 || index >= len(rows) {
		return Row{}, fmt.Errorf("virtual: row outside viewport")
	}
	t.Selected = rows[index].ID
	return rows[index], nil
}

// Source is immutable in identity; its contents change only on the UI owner.
func (t *Table) Source() Source { return t.source }
func (t *Table) syncSource() {
	length := t.source.Len()
	if length < 0 {
		length = 0
	}
	revision := uint64(0)
	if v, ok := t.source.(VersionedSource); ok {
		revision = v.Revision()
	}
	if length != t.sourceLength || revision != t.sourceRevision {
		t.generation++
		t.sourceLength = length
		t.sourceRevision = revision
		t.projectionInvalidated = t.projectionInvalidated || t.order != nil
		t.order = nil
		maximum := length - t.count
		if maximum < 0 {
			maximum = 0
		}
		if t.start > maximum {
			t.start = maximum
		}
	}
}

// ProjectionInvalidated reports a cleared order after source data changed.
func (t *Table) ProjectionInvalidated() bool { t.syncSource(); return t.projectionInvalidated }
func (t *Table) ReplaceSource(source Source) error {
	if source == nil || source.Len() < 0 {
		return fmt.Errorf("virtual: source required")
	}
	t.generation++
	t.source = source
	t.sourceRevision = 0
	if v, ok := source.(VersionedSource); ok {
		t.sourceRevision = v.Revision()
	}
	t.projectionInvalidated = false
	t.sourceLength = source.Len()
	t.order = nil
	t.start = 0
	return nil
}

func isVersioned(s Source) bool { _, ok := s.(VersionedSource); return ok }

// ProjectionToken must be captured on the UI owner before creating a worker snapshot.
func (t *Table) ProjectionToken() uint64 { t.syncSource(); return t.generation }
