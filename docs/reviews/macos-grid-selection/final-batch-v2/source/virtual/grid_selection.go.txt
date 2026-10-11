package virtual

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"
)

var ErrGridSelection = errors.New("virtual: row selection is not enabled")

// GridRowIndex supplies indexed stable identities, without cell materialization.
type GridRowIndex interface {
	Identity() string
	Revision() uint64
	RowCount() int
	RowKey(index int) (string, error)
	IndexOf(key string) (index int, found bool, err error)
}

// GridSelectionSource can capture the current view's immutable identity index.
// The returned index must match the GridSource stamp, remain immutable forever,
// and allow concurrent reads on workers. Capture must be bounded/nonblocking;
// prepare loaded-data indexes before installation, never by scanning on capture.
// Returning a mutable source itself does not fulfill this contract.
type GridSelectionSource interface {
	SelectionSnapshot() (GridRowIndex, error)
}

type GridSelectMode uint8

const (
	GridSelectReplace GridSelectMode = iota
	GridSelectExtend
	GridSelectToggle
)

// GridSelectionState counts captured identities, not rows in a later filtered
// view. Scope describes the base captured order; sparse additions may come from
// later same-incarnation views. Hidden/removed identities remain captured until
// a new selection or incarnation. Domain actions must validate live identities.
type GridSelectionState struct {
	Revision     uint64
	CapturedRows int
	AnchorKey    string
	Scope        GridStamp
	Overrides    int
}

// GridSelection is a copyable immutable membership snapshot. Its private sparse
// map is never modified after publication. It borrows one immutable source index,
// which may retain domain/index memory, but has no Grid/application backpointer.
type GridSelection struct {
	state      GridSelectionState
	index      GridRowIndex
	first, end int
	overrides  map[string]bool
	bytes      int
}

func (s GridSelection) State() GridSelectionState { return s.state }
func (s GridSelection) current() error {
	if s.index == nil {
		return nil
	}
	if err := sameStamp(s.index, s.state.Scope); err != nil {
		return fmt.Errorf("%w: immutable selection index: %w", ErrGridSource, err)
	}
	return nil
}
func (s GridSelection) baseContains(key string) (selected, found bool, err error) {
	if s.index == nil {
		return false, false, nil
	}
	i, found, err := checkedGridIndex(s.index, s.state.Scope, key)
	return found && i >= s.first && i < s.end, found, err
}
func (s GridSelection) query(key string) (selected, found bool, err error) {
	selected, _, found, err = s.queryIndex(key)
	return
}
func (s GridSelection) queryIndex(key string) (selected bool, row int, found bool, err error) {
	if key == "" || len(key) > gridMaxIdentityBytes || !utf8.ValidString(key) {
		return false, 0, false, ErrGridRange
	}
	if err := s.current(); err != nil {
		return false, 0, false, err
	}
	if s.index != nil {
		row, found, err = checkedGridIndex(s.index, s.state.Scope, key)
		if err != nil {
			return false, 0, false, err
		}
		selected = found && row >= s.first && row < s.end
	}
	if override, ok := s.overrides[key]; ok {
		selected = override
	}
	if err := s.current(); err != nil {
		return false, 0, false, err
	}
	return selected, row, found, nil
}
func (s GridSelection) selectedRow(key string, row int, stamp GridStamp) (bool, error) {
	selected, capturedRow, found, err := s.queryIndex(key)
	if err != nil {
		return false, err
	}
	if s.index != nil && s.state.Scope == stamp && (!found || capturedRow != row) {
		return false, ErrGridSource
	}
	return selected, nil
}
func (s GridSelection) Contains(key string) (bool, error) {
	selected, _, err := s.query(key)
	return selected, err
}

// VisitContext enumerates captured identities on a worker, in captured row
// order followed by lexically sorted sparse additions. It is intentionally
// O(selected interval + overrides), not a UI operation. Keys are detached before
// callbacks. Error/cancellation may follow partial callback delivery; accumulate
// privately and publish a result only after a nil return. Callback work cannot
// be forcibly cancelled. A zero snapshot visits nothing.
func (s GridSelection) VisitContext(ctx context.Context, visit func(string) error) error {
	if ctx == nil || visit == nil {
		return ErrGridInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.current(); err != nil {
		return err
	}
	for i := s.first; i < s.end; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		key, err := checkedGridRow(s.index, s.state.Scope, i)
		if err != nil {
			return err
		}
		if value, exists := s.overrides[key]; exists && !value {
			continue
		}
		if err = visit(strings.Clone(key)); err != nil {
			return err
		}
	}
	keys := make([]string, 0, len(s.overrides))
	for key, selected := range s.overrides {
		if selected {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		base, _, err := s.baseContains(key)
		if err != nil {
			return err
		}
		if !base {
			if err := visit(key); err != nil {
				return err
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.current()
}

// Selection captures membership without copying or enumerating domain rows.
// All future Grid transitions replace, rather than mutate, this snapshot.
func (g *Grid) Selection() (GridSelection, error) {
	if err := g.current(); err != nil {
		return GridSelection{}, err
	}
	if !g.config.RowSelection {
		return GridSelection{}, ErrGridSelection
	}
	return g.selection, nil
}

func (g *Grid) selectionIndex() (GridRowIndex, error) {
	source, ok := g.source.(GridSelectionSource)
	if !ok {
		return nil, ErrGridSource
	}
	index, err := source.SelectionSnapshot()
	if err != nil {
		return nil, fmt.Errorf("%w: selection snapshot: %w", ErrGridSource, err)
	}
	stamp, err := sourceStamp(index)
	if err != nil {
		return nil, err
	}
	if stamp != g.stamp {
		return nil, ErrGridSource
	}
	if err := checkedGridRows(g.source, g.stamp, g.rows); err != nil {
		return nil, err
	}
	return index, nil
}

func (g *Grid) rangeSelection(first, end int, anchor string) (GridSelection, error) {
	index, err := g.selectionIndex()
	if err != nil {
		return GridSelection{}, err
	}
	if first < 0 || end < first || end > g.stamp.Rows {
		return GridSelection{}, ErrGridRange
	}
	// Check accessed range endpoints against the live view, without enumerating.
	for _, row := range []int{first, end - 1} {
		if first == end {
			break
		}
		key, err := checkedGridRow(g.source, g.stamp, row)
		if err != nil {
			return GridSelection{}, err
		}
		other, err := checkedGridRow(index, g.stamp, row)
		if err != nil || key != other {
			if err != nil {
				return GridSelection{}, err
			}
			return GridSelection{}, ErrGridSource
		}
	}
	s := GridSelection{index: index, first: first, end: end, state: GridSelectionState{Scope: g.stamp, CapturedRows: end - first, AnchorKey: strings.Clone(anchor)}}
	s.state.Scope.Identity = strings.Clone(s.state.Scope.Identity)
	s.bytes = len(s.state.Scope.Identity) + len(s.state.AnchorKey)
	if s.bytes > g.config.MaxTextBytes {
		return GridSelection{}, ErrGridCapacity
	}
	return s, s.current()
}

func (g *Grid) prepareSelection(key string, row int, mode GridSelectMode) (GridSelection, error) {
	if !g.config.RowSelection {
		if mode != GridSelectReplace {
			return GridSelection{}, ErrGridSelection
		}
		return g.selection, nil
	}
	switch mode {
	case GridSelectReplace:
		return g.rangeSelection(row, row+1, key)
	case GridSelectExtend:
		anchor := g.selection.state.AnchorKey
		if anchor == "" {
			anchor = g.active.RowKey
		}
		origin := row
		if anchor != "" {
			i, found, err := checkedGridIndex(g.source, g.stamp, anchor)
			if err != nil {
				return GridSelection{}, err
			}
			if found {
				origin = i
			} else {
				anchor = key
			}
		} else {
			anchor = key
		}
		return g.rangeSelection(min(origin, row), max(origin, row)+1, anchor)
	case GridSelectToggle:
		s := g.selection
		if s.index == nil {
			return g.rangeSelection(row, row+1, key)
		}
		selected, err := s.Contains(key)
		if err != nil {
			return GridSelection{}, err
		}
		base, _, err := s.baseContains(key)
		if err != nil {
			return GridSelection{}, err
		}
		if !selected && s.state.CapturedRows == math.MaxInt {
			return GridSelection{}, ErrGridCapacity
		}
		count, bytes := len(s.overrides), s.bytes
		_, exists := s.overrides[key]
		keep := !selected != base
		if exists && !keep {
			count--
			bytes -= len(key)
		} else if !exists && keep {
			count++
			if len(key) > g.config.MaxTextBytes-bytes {
				return GridSelection{}, ErrGridCapacity
			}
			bytes += len(key)
		}
		if count > g.config.MaxSelectionKeys {
			return GridSelection{}, ErrGridCapacity
		}
		// Never modify a map shared by an already captured public snapshot.
		s.overrides = make(map[string]bool, count)
		for old, value := range g.selection.overrides {
			if old != key {
				s.overrides[old] = value
			}
		}
		if keep {
			s.overrides[strings.Clone(key)] = !selected
		}
		if selected {
			s.state.CapturedRows--
		} else {
			s.state.CapturedRows++
		}
		s.state.Overrides, s.bytes = count, bytes
		return s, s.current()
	default:
		return GridSelection{}, ErrGridInvalid
	}
}

func sameGridSelection(a, b GridSelection) bool {
	x, y := a.state, b.state
	x.Revision, y.Revision = 0, 0
	if x != y || a.first != b.first || a.end != b.end || len(a.overrides) != len(b.overrides) {
		return false
	}
	for key, value := range a.overrides {
		if other, exists := b.overrides[key]; !exists || value != other {
			return false
		}
	}
	return true
}

func (g *Grid) validateSelectionView(s GridSelection, source GridSource, stamp GridStamp, rows gridRows, v GridViewport) error {
	if !g.config.RowSelection || s.index == nil {
		return nil
	}
	layout, err := g.layout(v, rows)
	if err != nil {
		return err
	}
	for row := layout.firstRow; row < layout.endRow; row++ {
		key, err := checkedGridRow(source, stamp, row)
		if err != nil {
			return err
		}
		_, err = s.selectedRow(key, row, stamp)
		if err != nil {
			return err
		}
	}
	return s.current()
}

func (g *Grid) commitSelection(s GridSelection) error {
	if sameGridSelection(s, g.selection) {
		return g.current()
	}
	if err := g.nextGeneration(); err != nil {
		return err
	}
	if g.selection.state.Revision == math.MaxUint64 {
		return ErrGridExhausted
	}
	if err := g.validateSelectionView(s, g.source, g.stamp, g.rows, g.viewport); err != nil {
		return err
	}
	if err := checkedGridRows(g.source, g.stamp, g.rows); err != nil {
		return err
	}
	if err := s.current(); err != nil {
		return err
	}
	if err := sameStamp(g.source, g.stamp); err != nil {
		return err
	}
	s.state.Revision = g.selection.state.Revision + 1
	g.selection = s
	g.generation++
	return nil
}

func (g *Grid) SelectAll() error {
	if err := g.current(); err != nil {
		return err
	}
	if !g.config.RowSelection {
		return ErrGridSelection
	}
	anchor := g.active.RowKey
	if anchor == "" && g.stamp.Rows > 0 {
		row, err := g.rows.at(g.viewport.ScrollY)
		if err != nil {
			return err
		}
		anchor, err = checkedGridRow(g.source, g.stamp, row)
		if err != nil {
			return err
		}
	}
	s, err := g.rangeSelection(0, g.stamp.Rows, anchor)
	if err != nil {
		return err
	}
	return g.commitSelection(s)
}

// ClearSelection can recover from a stale/invalid captured selection index.
// It reads no source metadata; stale live Grid geometry still requires Refresh.
func (g *Grid) ClearSelection() error {
	if err := g.valid(); err != nil {
		return err
	}
	if !g.config.RowSelection {
		return ErrGridSelection
	}
	s := GridSelection{}
	if sameGridSelection(s, g.selection) {
		return nil
	}
	if err := g.nextGeneration(); err != nil {
		return err
	}
	if g.selection.state.Revision == math.MaxUint64 {
		return ErrGridExhausted
	}
	s.state.Revision = g.selection.state.Revision + 1
	g.selection = s
	g.generation++
	return nil
}
