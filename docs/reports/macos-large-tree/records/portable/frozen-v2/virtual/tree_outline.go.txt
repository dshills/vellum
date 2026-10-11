package virtual

import (
	"math"
	"strings"
)

type OutlineChangeKind uint8

const (
	OutlineKeep OutlineChangeKind = iota
	OutlineToggle
	OutlineExpand
	OutlineCollapse
	OutlineExpandAll
	OutlineCollapseAll
	OutlineFilter
)

// OutlineChange is one bounded intent against the accepted index. A newer
// Prepare replaces earlier unaccepted intents; it does not accumulate a queue.
type OutlineChange struct {
	Kind       OutlineChangeKind
	ID, Filter string
}

type OutlineRequest struct {
	token    *outlineRequestToken
	base     *treeIndexToken
	previous *TreeIndex
	data     *TreeData
	change   OutlineChange
}

func (r OutlineRequest) Stamp() TreeStamp {
	if r.token == nil {
		return TreeStamp{}
	}
	return r.data.Stamp()
}
func (r OutlineRequest) Change() OutlineChange { return r.change }

type OutlineConfig struct{ RowHeight, MaxRows int }
type OutlineViewport struct{ Width, Height, ScrollY int }
type OutlineRect struct{ X, Y, Width, Height int }
type OutlineState struct {
	Stamp                    TreeStamp
	Generation               uint64
	VisibleRows, TotalHeight int
	Selected                 string
	SelectedVisible          bool
	Filter                   string
	Pending                  bool
	Config                   OutlineConfig
	Viewport                 OutlineViewport
}
type outlineRefToken struct{ marker byte }

// OutlineRef has no model/data backpointer. Generation and private model
// identity make captured input stale after selection/viewport/index changes.
type OutlineRef struct {
	token      *outlineRefToken
	generation uint64
	row        int
	key        string
}
type OutlineRow struct {
	TreeRow
	Index      int
	Selected   bool
	Rect, Clip OutlineRect
	Ref        OutlineRef
}
type OutlineSnapshot struct {
	State OutlineState
	Rows  []OutlineRow
}

// Outline is owner-held state. Construction/admission/navigation/counts never
// scan the full tree. Only sealed worker TreeIndex results can be installed.
// Copies are invalid; use a pointer on one application owner.
type Outline struct {
	self       *Outline
	token      *outlineRefToken
	index      *TreeIndex
	config     OutlineConfig
	viewport   OutlineViewport
	selected   string
	generation uint64
	pending    *outlineRequestToken
}

func checkedOutlineConfig(c OutlineConfig) (OutlineConfig, error) {
	if c.RowHeight == 0 {
		c.RowHeight = 24
	}
	if c.MaxRows == 0 {
		c.MaxRows = 256
	}
	if c.RowHeight < 1 || c.RowHeight > 1<<20 || c.MaxRows < 1 || c.MaxRows > MaxTreeViewRows {
		return OutlineConfig{}, ErrTreeCapacity
	}
	return c, nil
}
func outlineViewport(v OutlineViewport, c OutlineConfig, index *TreeIndex) (OutlineViewport, error) {
	if v.Width < 1 || v.Height < 1 || v.ScrollY < 0 {
		return OutlineViewport{}, ErrTreeRange
	}
	if len(index.visible) > math.MaxInt/c.RowHeight {
		return OutlineViewport{}, ErrTreeCapacity
	}
	total := len(index.visible) * c.RowHeight
	v.ScrollY = min(v.ScrollY, max(0, total-v.Height))
	start := v.ScrollY / c.RowHeight
	count := v.Height / c.RowHeight
	tail := v.Height%c.RowHeight + v.ScrollY%c.RowHeight
	count += tail / c.RowHeight
	if tail%c.RowHeight != 0 {
		count++
	}
	if min(count, len(index.visible)-start) > c.MaxRows {
		return OutlineViewport{}, ErrTreeCapacity
	}
	return v, nil
}
func NewOutline(index *TreeIndex, c OutlineConfig, v OutlineViewport) (*Outline, error) {
	if index.valid() != nil {
		return nil, ErrTreeInvalid
	}
	c, err := checkedOutlineConfig(c)
	if err != nil {
		return nil, err
	}
	v, err = outlineViewport(v, c, index)
	if err != nil {
		return nil, err
	}
	o := &Outline{token: &outlineRefToken{}, index: index, config: c, viewport: v, generation: 1}
	o.self = o
	return o, nil
}
func (o *Outline) valid() error {
	if o == nil || o.self != o || o.token == nil || o.index.valid() != nil {
		return ErrTreeInvalid
	}
	return nil
}
func (o *Outline) VisibleLen() int {
	if o.valid() != nil {
		return 0
	}
	return len(o.index.visible)
}
func (o *Outline) State() (OutlineState, error) {
	if err := o.valid(); err != nil {
		return OutlineState{}, err
	}
	visible := false
	if n, ok := o.index.data.byID[o.selected]; ok {
		visible = o.index.positions[n] > 0
	}
	return OutlineState{Stamp: o.index.data.stamp, Generation: o.generation, VisibleRows: len(o.index.visible), TotalHeight: len(o.index.visible) * o.config.RowHeight, Selected: o.selected, SelectedVisible: visible, Filter: o.index.filter, Pending: o.pending != nil, Config: o.config, Viewport: o.viewport}, nil
}
func (o *Outline) row(position int) OutlineRow {
	n := o.index.visible[position]
	node := o.index.data.nodes[n]
	y := position*o.config.RowHeight - o.viewport.ScrollY
	clipTop := max(0, y)
	clipBottom := min(o.viewport.Height, y+o.config.RowHeight)
	clip := OutlineRect{Width: o.viewport.Width}
	if clipBottom > clipTop {
		clip.Y = clipTop
		clip.Height = clipBottom - clipTop
	}
	return OutlineRow{TreeRow: TreeRow{TreeNode: node.TreeNode, Depth: node.depth, Expanded: o.index.expanded[n], HasChildren: node.first >= 0}, Index: position, Selected: node.ID == o.selected, Rect: OutlineRect{Y: y, Width: o.viewport.Width, Height: o.config.RowHeight}, Clip: clip, Ref: OutlineRef{o.token, o.generation, position, node.ID}}
}
func (o *Outline) Snapshot() (OutlineSnapshot, error) {
	state, err := o.State()
	if err != nil {
		return OutlineSnapshot{}, err
	}
	start := o.viewport.ScrollY / o.config.RowHeight
	count := o.viewport.Height / o.config.RowHeight
	tail := o.viewport.Height%o.config.RowHeight + o.viewport.ScrollY%o.config.RowHeight
	count += tail / o.config.RowHeight
	if tail%o.config.RowHeight != 0 {
		count++
	}
	count = min(count, len(o.index.visible)-start)
	rows := make([]OutlineRow, count)
	for i := range rows {
		rows[i] = o.row(start + i)
	}
	return OutlineSnapshot{state, rows}, nil
}

// View is a detached bounded row window without changing the viewport. Rect and
// Clip remain relative to the current accepted pixel viewport.
func (o *Outline) View(start, count int) ([]OutlineRow, error) {
	if err := o.valid(); err != nil {
		return nil, err
	}
	if start < 0 || count < 1 || count > o.config.MaxRows {
		return nil, ErrTreeRange
	}
	if start >= len(o.index.visible) {
		return nil, nil
	}
	count = min(count, len(o.index.visible)-start)
	rows := make([]OutlineRow, count)
	for i := range rows {
		rows[i] = o.row(start + i)
	}
	return rows, nil
}
func (o *Outline) Row(ref OutlineRef) (OutlineRow, error) {
	if err := o.valid(); err != nil {
		return OutlineRow{}, err
	}
	if ref.token != o.token || ref.generation != o.generation || ref.row < 0 || ref.row >= len(o.index.visible) || o.index.data.nodes[o.index.visible[ref.row]].ID != ref.key {
		return OutlineRow{}, ErrTreeStale
	}
	return o.row(ref.row), nil
}
func (o *Outline) Current() (OutlineRow, error) {
	if err := o.valid(); err != nil {
		return OutlineRow{}, err
	}
	n, ok := o.index.data.byID[o.selected]
	if !ok {
		return OutlineRow{}, ErrTreeRange
	}
	position := o.index.positions[n]
	if position == 0 {
		return OutlineRow{}, ErrTreeHidden
	}
	return o.row(position - 1), nil
}
func (o *Outline) publishSelection(key string, v OutlineViewport) error {
	if key == o.selected && v == o.viewport {
		return nil
	}
	if o.generation == math.MaxUint64 {
		return ErrTreeExhausted
	}
	o.selected = strings.Clone(key)
	o.viewport = v
	o.generation++
	return nil
}
func (o *Outline) Select(id string) error {
	if err := o.valid(); err != nil {
		return err
	}
	if err := checkedTreeKey(id, o.index.data.limits.MaxNodeTextBytes); err != nil {
		return err
	}
	if _, ok := o.index.data.byID[id]; !ok {
		return ErrTreeRange
	}
	return o.publishSelection(id, o.viewport)
}
func (o *Outline) SelectRef(ref OutlineRef) error {
	row, err := o.Row(ref)
	if err != nil {
		return err
	}
	return o.publishSelection(row.ID, o.viewport)
}
func (o *Outline) SetViewport(v OutlineViewport) error {
	if err := o.valid(); err != nil {
		return err
	}
	v, err := outlineViewport(v, o.config, o.index)
	if err != nil {
		return err
	}
	return o.publishSelection(o.selected, v)
}
func (o *Outline) revealPosition(position int) error {
	v := o.viewport
	top := position * o.config.RowHeight
	bottom := top + o.config.RowHeight
	if top < v.ScrollY {
		v.ScrollY = top
	} else if bottom > v.ScrollY+v.Height {
		v.ScrollY = max(0, bottom-v.Height)
	}
	v, err := outlineViewport(v, o.config, o.index)
	if err != nil {
		return err
	}
	return o.publishSelection(o.index.data.nodes[o.index.visible[position]].ID, v)
}
func (o *Outline) Reveal(id string) error {
	if err := o.valid(); err != nil {
		return err
	}
	if err := checkedTreeKey(id, o.index.data.limits.MaxNodeTextBytes); err != nil {
		return err
	}
	n, ok := o.index.data.byID[id]
	if !ok {
		return ErrTreeRange
	}
	position := o.index.positions[n]
	if position == 0 {
		return ErrTreeHidden
	}
	return o.revealPosition(position - 1)
}
func (o *Outline) Move(delta int) error {
	if err := o.valid(); err != nil {
		return err
	}
	if delta == 0 {
		return nil
	}
	count := len(o.index.visible)
	if count == 0 {
		return ErrTreeRange
	}
	position := -1
	if n, ok := o.index.data.byID[o.selected]; ok {
		position = o.index.positions[n] - 1
	}
	if position < 0 {
		if delta > 0 {
			position = 0
		} else {
			position = count - 1
		}
	} else if delta > 0 {
		position += min(delta, count-1-position)
	} else {
		if delta == math.MinInt {
			position = 0
		} else {
			position -= min(-delta, position)
		}
	}
	return o.revealPosition(position)
}
func (o *Outline) Page(pages int) error {
	if err := o.valid(); err != nil {
		return err
	}
	step := max(1, o.viewport.Height/o.config.RowHeight)
	if pages > math.MaxInt/step {
		return o.Move(math.MaxInt)
	}
	if pages < math.MinInt/step {
		return o.Move(math.MinInt)
	}
	return o.Move(pages * step)
}
func (o *Outline) Home() error {
	if err := o.valid(); err != nil {
		return err
	}
	if len(o.index.visible) == 0 {
		return ErrTreeRange
	}
	return o.revealPosition(0)
}
func (o *Outline) End() error {
	if err := o.valid(); err != nil {
		return err
	}
	if len(o.index.visible) == 0 {
		return ErrTreeRange
	}
	return o.revealPosition(len(o.index.visible) - 1)
}
func (o *Outline) Parent() error {
	if err := o.valid(); err != nil {
		return err
	}
	n, ok := o.index.data.byID[o.selected]
	if !ok || o.index.data.nodes[n].parent < 0 {
		return ErrTreeRange
	}
	p := o.index.data.nodes[n].parent
	position := o.index.positions[p]
	if position == 0 {
		return ErrTreeHidden
	}
	return o.revealPosition(position - 1)
}
func (o *Outline) FirstChild() error {
	current, err := o.Current()
	if err != nil {
		return err
	}
	next := current.Index + 1
	if next >= len(o.index.visible) || o.index.data.nodes[o.index.visible[next]].depth != current.Depth+1 {
		return ErrTreeHidden
	}
	return o.revealPosition(next)
}
func (o *Outline) Prepare(change OutlineChange) (OutlineRequest, error) {
	if err := o.valid(); err != nil {
		return OutlineRequest{}, err
	}
	return o.PrepareData(o.index.data, change)
}
func (o *Outline) PrepareData(data *TreeData, change OutlineChange) (OutlineRequest, error) {
	if err := o.valid(); err != nil {
		return OutlineRequest{}, err
	}
	if data.valid() != nil {
		return OutlineRequest{}, ErrTreeInvalid
	}
	if o.generation == math.MaxUint64 {
		return OutlineRequest{}, ErrTreeExhausted
	}
	if data.stamp.Identity == o.index.data.stamp.Identity && data != o.index.data && data.stamp.Revision <= o.index.data.stamp.Revision {
		return OutlineRequest{}, ErrTreeStale
	}
	if change.Kind > OutlineFilter || len(change.ID) > data.limits.MaxNodeTextBytes || len(change.Filter) > data.limits.MaxQueryBytes {
		return OutlineRequest{}, ErrTreeCapacity
	}
	if !validTreeText(change.ID, data.limits.MaxNodeTextBytes, true) || !validTreeText(change.Filter, data.limits.MaxQueryBytes, true) {
		return OutlineRequest{}, ErrTreeSource
	}
	switch change.Kind {
	case OutlineToggle, OutlineExpand, OutlineCollapse:
		if _, ok := data.byID[change.ID]; !ok {
			return OutlineRequest{}, ErrTreeRange
		}
	}
	change.ID = strings.Clone(change.ID)
	change.Filter = strings.Clone(change.Filter)
	r := OutlineRequest{token: &outlineRequestToken{}, base: o.index.token, previous: o.index, data: data, change: change}
	o.pending = r.token
	return r, nil
}
func (o *Outline) Cancel(request OutlineRequest) error {
	if err := o.valid(); err != nil {
		return err
	}
	if request.token == nil || request.token != o.pending {
		return ErrTreeStale
	}
	o.pending = nil
	return nil
}
func (o *Outline) Complete(request OutlineRequest, index *TreeIndex) error {
	if err := o.valid(); err != nil {
		return err
	}
	if request.token == nil || request.token != o.pending || request.base != o.index.token {
		return ErrTreeStale
	}
	if index.valid() != nil {
		return ErrTreeInvalid
	}
	if index.request != request.token || index.base != request.base || index.data != request.data {
		return ErrTreeStale
	}
	if o.generation == math.MaxUint64 {
		return ErrTreeExhausted
	}
	// Check candidate extent before calculating a retained anchor. This also
	// keeps admission atomic on platforms with narrower integer coordinates.
	if len(index.visible) > math.MaxInt/o.config.RowHeight {
		return ErrTreeCapacity
	}
	v := o.viewport
	// Retain the top accepted stable row when it remains visible; otherwise
	// clamp the existing pixel offset. This is one lookup, not a hidden scan.
	if len(o.index.visible) > 0 && index.data.stamp.Identity == o.index.data.stamp.Identity {
		n := o.index.visible[o.viewport.ScrollY/o.config.RowHeight]
		id := o.index.data.nodes[n].ID
		if next, ok := index.data.byID[id]; ok && index.positions[next] > 0 {
			v.ScrollY = (index.positions[next]-1)*o.config.RowHeight + o.viewport.ScrollY%o.config.RowHeight
		}
	}
	v, err := outlineViewport(v, o.config, index)
	if err != nil {
		return err
	}
	selected := o.selected
	if index.data.stamp.Identity != o.index.data.stamp.Identity {
		selected = ""
	} else if _, ok := index.data.byID[selected]; !ok {
		selected = ""
	}
	o.index = index
	o.viewport = v
	o.selected = selected
	o.pending = nil
	o.generation++
	return nil
}
