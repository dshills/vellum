// Package virtual projects indexed domain data into bounded viewport records.
package virtual

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"
)

var (
	ErrGridInvalid   = errors.New("virtual: invalid grid argument")
	ErrGridSource    = errors.New("virtual: invalid grid source")
	ErrGridStale     = errors.New("virtual: stale grid identity")
	ErrGridCapacity  = errors.New("virtual: grid capacity exceeded")
	ErrGridRange     = errors.New("virtual: grid identity not found")
	ErrGridExhausted = errors.New("virtual: grid generation exhausted")
)

// GridSource is a borrowed, indexed snapshot. Identity identifies an incarnation;
// Revision increases whenever content, membership, or ordering changes. Within
// one revision RowKey and IndexOf must be inverse, and all reads must agree.
// Metadata and indexed lookups must be bounded, nonblocking in-memory operations.
// The owner serializes Grid and source changes. Sources may prepare their indices
// on workers; Grid never builds or validates a whole-row index on the owner.
type GridSource interface {
	Identity() string
	Revision() uint64
	RowCount() int
	RowKey(index int) (string, error)
	IndexOf(key string) (index int, found bool, err error)
	Cell(rowKey, columnID string) (string, error)
}

type GridColumn struct {
	ID, Title string
	Width     int
}
type GridStamp struct {
	Identity string
	Revision uint64
	Rows     int
}
type GridViewport struct{ Width, Height, ScrollX, ScrollY int }
type GridRect struct{ X, Y, Width, Height int }

func (r GridRect) Empty() bool { return r.Width <= 0 || r.Height <= 0 }

// GridConfig bounds retained viewport records, including overscan. Coordinates
// are integer logical pixels. Zero capacity fields select documented defaults.
type GridConfig struct {
	RowHeight, HeaderHeight, FrozenColumns      int
	OverscanRows, OverscanColumns               int
	MaxRows, MaxColumns, MaxCells, MaxTextBytes int
}

type GridActive struct{ RowKey, ColumnID string }
type GridState struct {
	Stamp                                GridStamp
	Generation                           uint64
	Viewport                             GridViewport
	Active                               GridActive
	Columns                              int
	TotalWidth, TotalHeight, FrozenWidth int
	Config                               GridConfig
}

// GridCellRef is an opaque, copyable press identity. A source, viewport, active
// cell, or geometry transition retires old refs, even when a pooled slot is reused.
type GridCellRef struct {
	token         *gridRefToken
	generation    uint64
	row, column   int
	key, columnID string
}

// Nonzero size makes independently allocated tokens distinct. No backpointer
// lets detached records retain Grid's borrowed domain source or owner.
type gridRefToken struct{ marker byte }

func (r GridCellRef) RowKey() string   { return r.key }
func (r GridCellRef) ColumnID() string { return r.columnID }
func (r GridCellRef) RowIndex() int    { return r.row }
func (r GridCellRef) ColumnIndex() int { return r.column }

type GridCell struct {
	RowIndex, ColumnIndex  int
	RowKey, ColumnID, Text string
	Rect, Clip             GridRect
	Frozen, Active         bool
	Ref                    GridCellRef
}
type GridHeader struct {
	ColumnIndex     int
	ColumnID, Title string
	Rect, Clip      GridRect
	Frozen          bool
}
type GridSnapshot struct {
	State   GridState
	Headers []GridHeader
	Cells   []GridCell // row-major; columns remain in source definition order
}

// Grid is owner-confined, like Table. Copying a Grid value is refused. Snapshot
// records are detached; strings are immutable and no source arrays are retained.
type Grid struct {
	self         *Grid
	token        *gridRefToken
	source       GridSource
	stamp        GridStamp
	columns      []GridColumn
	prefix       []int
	config       GridConfig
	viewport     GridViewport
	active       GridActive
	generation   uint64
	anchorKey    string
	anchorOffset int
}

const gridMaxDefinitions = 4096
const gridMaxIdentityBytes = 1024

func normalizeGridConfig(c GridConfig) (GridConfig, error) {
	if c.MaxRows == 0 {
		c.MaxRows = 256
	}
	if c.MaxColumns == 0 {
		c.MaxColumns = 64
	}
	if c.MaxCells == 0 {
		c.MaxCells = 4096
	}
	if c.MaxTextBytes == 0 {
		c.MaxTextBytes = 8 << 20
	}
	if c.RowHeight <= 0 || c.HeaderHeight < 0 || c.FrozenColumns < 0 || c.OverscanRows < 0 || c.OverscanColumns < 0 ||
		c.MaxRows < 1 || c.MaxRows > 4096 || c.MaxColumns < 1 || c.MaxColumns > gridMaxDefinitions ||
		c.MaxCells < 1 || c.MaxCells > 1<<20 || c.MaxTextBytes < 1 || c.MaxTextBytes > 64<<20 {
		return c, ErrGridInvalid
	}
	return c, nil
}

func sourceStamp(s GridSource) (GridStamp, error) {
	if s == nil {
		return GridStamp{}, ErrGridSource
	}
	v := reflect.ValueOf(s)
	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface || v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Func) && v.IsNil() {
		return GridStamp{}, ErrGridSource
	}
	stamp := GridStamp{s.Identity(), s.Revision(), s.RowCount()}
	if stamp.Identity == "" || len(stamp.Identity) > gridMaxIdentityBytes || !utf8.ValidString(stamp.Identity) || stamp.Revision == 0 || stamp.Rows < 0 {
		return GridStamp{}, ErrGridSource
	}
	return stamp, nil
}

// NewGrid copies bounded column definitions; it never enumerates source rows.
func NewGrid(source GridSource, columns []GridColumn, config GridConfig, viewport GridViewport) (*Grid, error) {
	c, err := normalizeGridConfig(config)
	if err != nil {
		return nil, err
	}
	stamp, err := sourceStamp(source)
	if err != nil {
		return nil, err
	}
	if len(columns) < 1 || len(columns) > gridMaxDefinitions || c.FrozenColumns > len(columns) {
		return nil, ErrGridInvalid
	}
	prefix := make([]int, len(columns)+1)
	ids := make(map[string]bool, len(columns))
	textBytes := 0
	for i, col := range columns {
		if col.ID == "" || len(col.ID) > gridMaxIdentityBytes || !utf8.ValidString(col.ID) || ids[col.ID] || col.Width < 1 || col.Width > math.MaxInt-prefix[i] {
			return nil, ErrGridInvalid
		}
		if len(col.ID) > c.MaxTextBytes-textBytes || len(col.Title) > c.MaxTextBytes-textBytes-len(col.ID) {
			return nil, ErrGridCapacity
		}
		if !utf8.ValidString(col.Title) {
			return nil, ErrGridInvalid
		}
		textBytes += len(col.ID) + len(col.Title)
		ids[col.ID] = true
		prefix[i+1] = prefix[i] + col.Width
	}
	g := &Grid{token: new(gridRefToken), source: source, stamp: stamp, columns: append([]GridColumn(nil), columns...), prefix: prefix, config: c, generation: 1}
	g.self = g
	view, err := g.clampViewport(viewport, stamp)
	if err != nil {
		return nil, err
	}
	if _, err = g.layout(view, stamp); err != nil {
		return nil, err
	}
	key, offset, err := g.anchor(source, stamp, view)
	if err != nil {
		return nil, err
	}
	if err = sameStamp(source, stamp); err != nil {
		return nil, err
	}
	// Copying a string header would retain any large caller backing allocation.
	// Detach only values admitted to persistent Grid state, after validation.
	g.stamp.Identity = strings.Clone(stamp.Identity)
	for i := range g.columns {
		g.columns[i].ID = strings.Clone(g.columns[i].ID)
		g.columns[i].Title = strings.Clone(g.columns[i].Title)
	}
	g.viewport, g.anchorKey, g.anchorOffset = view, strings.Clone(key), offset
	return g, nil
}

func (g *Grid) valid() error {
	if g == nil || g.self != g {
		return ErrGridInvalid
	}
	return nil
}
func sameStamp(source GridSource, want GridStamp) error {
	got, err := sourceStamp(source)
	if err != nil {
		return err
	}
	if got != want {
		return ErrGridStale
	}
	return nil
}
func (g *Grid) current() error {
	if err := g.valid(); err != nil {
		return err
	}
	return sameStamp(g.source, g.stamp)
}
func (g *Grid) nextGeneration() error {
	if g.generation == math.MaxUint64 {
		return ErrGridExhausted
	}
	return nil
}
func (g *Grid) state() GridState {
	return GridState{Stamp: g.stamp, Generation: g.generation, Viewport: g.viewport, Active: g.active, Columns: len(g.columns), TotalWidth: g.prefix[len(g.columns)], TotalHeight: g.stamp.Rows * g.config.RowHeight, FrozenWidth: g.prefix[g.config.FrozenColumns], Config: g.config}
}
func (g *Grid) State() (GridState, error) {
	if err := g.current(); err != nil {
		return GridState{}, err
	}
	return g.state(), nil
}

func (g *Grid) clampViewport(v GridViewport, stamp GridStamp) (GridViewport, error) {
	if v.Width < 1 || v.Height < 1 || v.ScrollX < 0 || v.ScrollY < 0 || stamp.Rows > (math.MaxInt-g.config.HeaderHeight)/g.config.RowHeight {
		return v, ErrGridInvalid
	}
	frozen := g.prefix[g.config.FrozenColumns]
	panewidth := max(0, v.Width-frozen)
	maxx := 0
	if panewidth > 0 {
		maxx = max(0, g.prefix[len(g.columns)]-frozen-panewidth)
	}
	bodyheight := max(0, v.Height-g.config.HeaderHeight)
	maxy := 0
	if bodyheight > 0 {
		maxy = max(0, stamp.Rows*g.config.RowHeight-bodyheight)
	}
	v.ScrollX = min(v.ScrollX, maxx)
	v.ScrollY = min(v.ScrollY, maxy)
	return v, nil
}

func checkedGridRow(s GridSource, stamp GridStamp, index int) (string, error) {
	if index < 0 || index >= stamp.Rows {
		return "", ErrGridRange
	}
	key, err := s.RowKey(index)
	if err != nil {
		return "", fmt.Errorf("%w: row key: %w", ErrGridSource, err)
	}
	if key == "" || len(key) > gridMaxIdentityBytes || !utf8.ValidString(key) {
		return "", ErrGridSource
	}
	i, found, err := s.IndexOf(key)
	if err != nil {
		return "", fmt.Errorf("%w: index: %w", ErrGridSource, err)
	}
	if !found || i != index {
		return "", ErrGridSource
	}
	return key, nil
}
func checkedGridIndex(s GridSource, stamp GridStamp, key string) (int, bool, error) {
	if key == "" || len(key) > gridMaxIdentityBytes || !utf8.ValidString(key) {
		return 0, false, ErrGridRange
	}
	i, found, err := s.IndexOf(key)
	if err != nil {
		return 0, false, fmt.Errorf("%w: index: %w", ErrGridSource, err)
	}
	if !found {
		return 0, false, nil
	}
	if i < 0 || i >= stamp.Rows {
		return 0, false, ErrGridSource
	}
	actual, err := s.RowKey(i)
	if err != nil {
		return 0, false, fmt.Errorf("%w: row key: %w", ErrGridSource, err)
	}
	if actual != key {
		return 0, false, ErrGridSource
	}
	return i, true, nil
}
func (g *Grid) anchor(s GridSource, stamp GridStamp, v GridViewport) (string, int, error) {
	if stamp.Rows == 0 || v.Height <= g.config.HeaderHeight {
		return "", 0, nil
	}
	i := min(v.ScrollY/g.config.RowHeight, stamp.Rows-1)
	key, err := checkedGridRow(s, stamp, i)
	return key, v.ScrollY % g.config.RowHeight, err
}

// SetViewport clamps offsets, but refuses a viewport whose visible/overscan
// records exceed configured capacities. Refusal leaves all state and refs intact.
func (g *Grid) SetViewport(v GridViewport) error {
	if err := g.current(); err != nil {
		return err
	}
	view, err := g.clampViewport(v, g.stamp)
	if err != nil {
		return err
	}
	if view == g.viewport {
		return nil
	}
	if _, err = g.layout(view, g.stamp); err != nil {
		return err
	}
	if err = g.nextGeneration(); err != nil {
		return err
	}
	key, offset, err := g.anchor(g.source, g.stamp, view)
	if err != nil {
		return err
	}
	if err = sameStamp(g.source, g.stamp); err != nil {
		return err
	}
	g.viewport, g.anchorKey, g.anchorOffset = view, strings.Clone(key), offset
	g.generation++
	return nil
}
func addGridOffset(value, delta int) int {
	if delta > 0 && value > math.MaxInt-delta {
		return math.MaxInt
	}
	if delta < 0 && delta < -value {
		return 0
	}
	return value + delta
}
func (g *Grid) Scroll(dx, dy int) error {
	if err := g.current(); err != nil {
		return err
	}
	v := g.viewport
	v.ScrollX = addGridOffset(v.ScrollX, dx)
	v.ScrollY = addGridOffset(v.ScrollY, dy)
	return g.SetViewport(v)
}

// Refresh acknowledges a new source revision without enumerating rows. It keeps
// active/anchor keys within the same incarnation when they still exist. A new
// incarnation clears both. Removed active keys clear active state.
func (g *Grid) Refresh() error {
	if err := g.valid(); err != nil {
		return err
	}
	stamp, err := sourceStamp(g.source)
	if err != nil {
		return err
	}
	if stamp == g.stamp {
		return nil
	}
	return g.ReplaceSource(g.source)
}
func (g *Grid) ReplaceSource(source GridSource) error {
	if err := g.valid(); err != nil {
		return err
	}
	stamp, err := sourceStamp(source)
	if err != nil {
		return err
	}
	if stamp.Identity == g.stamp.Identity {
		if stamp.Revision < g.stamp.Revision {
			return ErrGridStale
		}
		if stamp.Revision == g.stamp.Revision && stamp.Rows != g.stamp.Rows {
			return ErrGridSource
		}
	}
	if err = g.nextGeneration(); err != nil {
		return err
	}
	active := g.active
	v := g.viewport
	if stamp.Identity != g.stamp.Identity {
		active = GridActive{}
		v.ScrollX, v.ScrollY = 0, 0
	} else {
		if active.RowKey != "" {
			_, found, e := checkedGridIndex(source, stamp, active.RowKey)
			if e != nil {
				return e
			}
			if !found {
				active = GridActive{}
			}
		}
		if g.anchorKey != "" {
			i, found, e := checkedGridIndex(source, stamp, g.anchorKey)
			if e != nil {
				return e
			}
			if found {
				if i > (math.MaxInt-g.anchorOffset)/g.config.RowHeight {
					return ErrGridInvalid
				}
				v.ScrollY = i*g.config.RowHeight + g.anchorOffset
			}
		}
	}
	v, err = g.clampViewport(v, stamp)
	if err != nil {
		return err
	}
	if _, err = g.layout(v, stamp); err != nil {
		return err
	}
	key, offset, err := g.anchor(source, stamp, v)
	if err != nil {
		return err
	}
	if err = sameStamp(source, stamp); err != nil {
		return err
	}
	stamp.Identity = strings.Clone(stamp.Identity)
	g.source, g.stamp, g.active, g.viewport, g.anchorKey, g.anchorOffset = source, stamp, active, v, strings.Clone(key), offset
	g.generation++
	return nil
}

func (g *Grid) columnIndex(id string) (int, error) {
	for i, c := range g.columns {
		if c.ID == id {
			return i, nil
		}
	}
	return 0, ErrGridRange
}
func (g *Grid) reveal(row, col int) GridViewport {
	v := g.viewport
	body := max(0, v.Height-g.config.HeaderHeight)
	if body > 0 {
		start := row * g.config.RowHeight
		if start < v.ScrollY || g.config.RowHeight >= body {
			v.ScrollY = start
		} else if start+g.config.RowHeight-v.ScrollY > body {
			v.ScrollY = start + g.config.RowHeight - body
		}
	}
	if col >= g.config.FrozenColumns {
		pane := max(0, v.Width-g.prefix[g.config.FrozenColumns])
		if pane > 0 {
			start := g.prefix[col] - g.prefix[g.config.FrozenColumns]
			width := g.columns[col].Width
			if start < v.ScrollX || width >= pane {
				v.ScrollX = start
			} else if start+width-v.ScrollX > pane {
				v.ScrollX = start + width - pane
			}
		}
	}
	v, _ = g.clampViewport(v, g.stamp)
	return v
}
func (g *Grid) activateKey(key string, row, col int, reveal bool) error {
	v := g.viewport
	if reveal {
		v = g.reveal(row, col)
	}
	active := GridActive{key, g.columns[col].ID}
	if active == g.active && v == g.viewport {
		return sameStamp(g.source, g.stamp)
	}
	if _, err := g.layout(v, g.stamp); err != nil {
		return err
	}
	if err := g.nextGeneration(); err != nil {
		return err
	}
	anchor, offset, err := g.anchor(g.source, g.stamp, v)
	if err != nil {
		return err
	}
	if err = sameStamp(g.source, g.stamp); err != nil {
		return err
	}
	active.RowKey = strings.Clone(active.RowKey)
	g.active, g.viewport, g.anchorKey, g.anchorOffset = active, v, strings.Clone(anchor), offset
	g.generation++
	return nil
}
func (g *Grid) JumpToKey(key, columnID string) error {
	if err := g.current(); err != nil {
		return err
	}
	col, err := g.columnIndex(columnID)
	if err != nil {
		return err
	}
	row, found, err := checkedGridIndex(g.source, g.stamp, key)
	if err != nil {
		return err
	}
	if !found {
		return ErrGridRange
	}
	return g.activateKey(key, row, col, true)
}

// Move clamps directional row/column deltas. With no active cell it starts at
// the viewport's first row and first column. Page moves a whole-body row count.
func (g *Grid) Move(rows, columns int) error {
	if err := g.current(); err != nil {
		return err
	}
	if g.stamp.Rows == 0 {
		return ErrGridRange
	}
	row, col := min(g.viewport.ScrollY/g.config.RowHeight, g.stamp.Rows-1), 0
	if g.active.RowKey != "" {
		i, found, err := checkedGridIndex(g.source, g.stamp, g.active.RowKey)
		if err != nil {
			return err
		}
		if !found {
			return ErrGridStale
		}
		row = i
		col, _ = g.columnIndex(g.active.ColumnID)
	}
	row = min(addGridOffset(row, rows), g.stamp.Rows-1)
	col = min(addGridOffset(col, columns), len(g.columns)-1)
	key, err := checkedGridRow(g.source, g.stamp, row)
	if err != nil {
		return err
	}
	return g.activateKey(key, row, col, true)
}
func (g *Grid) Page(direction int) error {
	if err := g.current(); err != nil {
		return err
	}
	if direction != -1 && direction != 1 {
		return ErrGridInvalid
	}
	return g.Move(direction*max(1, (g.viewport.Height-g.config.HeaderHeight)/g.config.RowHeight), 0)
}

// Activate accepts only a currently visible cell from this exact Grid generation.
// An overscan record with an empty clip cannot be activated.
func (g *Grid) Activate(ref GridCellRef) error {
	if err := g.current(); err != nil {
		return err
	}
	if ref.token != g.token || ref.generation != g.generation || ref.column < 0 || ref.column >= len(g.columns) || g.columns[ref.column].ID != ref.columnID {
		return ErrGridStale
	}
	key, err := checkedGridRow(g.source, g.stamp, ref.row)
	if err != nil {
		return err
	}
	if key != ref.key {
		return ErrGridStale
	}
	if g.cellClip(g.viewport, ref.row, ref.column).Empty() {
		return ErrGridStale
	}
	return g.activateKey(key, ref.row, ref.column, false)
}
