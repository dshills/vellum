package virtual

import (
	"context"
	"strings"
	"sync/atomic"
)

type TreeIndexOptions struct {
	ExpandAll bool
	Filter    string
}

type treeIndexToken struct{ marker byte }
type outlineRequestToken struct{ marker byte }

type TreeIndexStats struct {
	Nodes, VisibleRows int
	StorageBytes       int64
	Operations         uint64
}

// TreeIndex owns a completed immutable visible order and expansion state. It
// retains its sealed TreeData, never an Outline or a prior index/result chain.
// It is concurrent-readable. Struct copies are invalid.
type TreeIndex struct {
	self      *TreeIndex
	token     *treeIndexToken
	data      *TreeData
	expanded  []bool
	visible   []int
	positions []int // one-based visible position; zero denotes hidden
	filter    string
	stats     TreeIndexStats
	request   *outlineRequestToken
	base      *treeIndexToken
}

func (i *TreeIndex) valid() error {
	if i == nil || i.self != i || i.token == nil || i.data.valid() != nil {
		return ErrTreeInvalid
	}
	return nil
}
func (i *TreeIndex) Stamp() TreeStamp {
	if i.valid() != nil {
		return TreeStamp{}
	}
	return i.data.stamp
}
func (i *TreeIndex) Len() int {
	if i.valid() != nil {
		return 0
	}
	return len(i.visible)
}
func (i *TreeIndex) Stats() TreeIndexStats {
	if i.valid() != nil {
		return TreeIndexStats{}
	}
	return i.stats
}

type treeIndexFrame struct{ next int }
type TreeBuildProgress struct {
	Operations  uint64
	VisibleRows int
	Done        bool
}

// TreeIndexBuilder is single-worker bounded-step execution. Every node copy,
// filter test, ancestor propagation, traversal node and stack pop consumes one
// operation; no fan-out loop or ancestor-chain scan hides inside an operation.
// Constructors allocate bounded arrays on the worker; never call them on UI.
// Struct copies, concurrent/reentrant calls and nonpositive step budgets refuse.
type TreeIndexBuilder struct {
	self       *TreeIndexBuilder
	busy       atomic.Bool
	ctx        context.Context
	data       *TreeData
	previous   *TreeIndex
	change     OutlineChange
	options    TreeIndexOptions
	request    *outlineRequestToken
	base       *treeIndexToken
	expanded   []bool
	included   []bool
	positions  []int
	visible    []int
	stack      []treeIndexFrame
	phase      uint8
	cursor     int
	operations uint64
	storage    int64
	filter     string
	err        error
	result     *TreeIndex
}

func NewTreeIndexBuilder(ctx context.Context, data *TreeData, options TreeIndexOptions) (*TreeIndexBuilder, error) {
	return newTreeBuilder(ctx, data, nil, OutlineChange{}, options, nil, nil)
}
func NewOutlineBuilder(ctx context.Context, request OutlineRequest) (*TreeIndexBuilder, error) {
	if request.token == nil || request.base == nil || request.previous.valid() != nil || request.data.valid() != nil || request.previous.token != request.base {
		return nil, ErrTreeInvalid
	}
	return newTreeBuilder(ctx, request.data, request.previous, request.change, TreeIndexOptions{}, request.token, request.base)
}
func newTreeBuilder(ctx context.Context, data *TreeData, previous *TreeIndex, change OutlineChange, options TreeIndexOptions, request *outlineRequestToken, base *treeIndexToken) (*TreeIndexBuilder, error) {
	if ctx == nil || data.valid() != nil {
		return nil, ErrTreeInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	filter := options.Filter
	if previous != nil {
		filter = previous.filter
		if change.Kind == OutlineFilter {
			filter = change.Filter
		}
	}
	if len(filter) > data.limits.MaxQueryBytes {
		return nil, ErrTreeCapacity
	}
	if !validTreeText(filter, data.limits.MaxQueryBytes, true) {
		return nil, ErrTreeSource
	}
	n := len(data.nodes)
	maxVisible := min(n, data.limits.MaxVisibleNodes)
	// Expanded+included booleans, position slots, visible slots, DFS frames and
	// a bounded detached query. Prior indexes/data and allocation overhead are excluded.
	storage := int64(n)*10 + int64(maxVisible)*8 + int64(min(n+1, data.limits.MaxDepth+2))*8 + int64(len(filter))
	if storage > data.limits.MaxStorageBytes {
		return nil, ErrTreeCapacity
	}
	b := &TreeIndexBuilder{ctx: ctx, data: data, previous: previous, change: change, options: options, request: request, base: base, expanded: make([]bool, n), included: make([]bool, n), positions: make([]int, n), visible: make([]int, 0, maxVisible), stack: make([]treeIndexFrame, 0, min(n+1, data.limits.MaxDepth+2)), filter: strings.Clone(filter), storage: storage}
	b.self = b
	return b, nil
}
func (b *TreeIndexBuilder) enter() error {
	if b == nil || b.self != b {
		return ErrTreeInvalid
	}
	if !b.busy.CompareAndSwap(false, true) {
		return ErrTreeBusy
	}
	return nil
}
func (b *TreeIndexBuilder) Step(maxOperations int) (bool, error) {
	if err := b.enter(); err != nil {
		return false, err
	}
	defer b.busy.Store(false)
	if maxOperations < 1 || maxOperations > MaxTreeStepOperations {
		return false, ErrTreeCapacity
	}
	if b.err != nil {
		return false, b.err
	}
	if err := b.ctx.Err(); err != nil {
		b.err = err
		return false, err
	}
	if b.result != nil {
		return true, nil
	}
	for used := 0; used < maxOperations; used++ {
		if err := b.ctx.Err(); err != nil {
			b.err = err
			return false, err
		}
		if b.operations >= b.data.limits.MaxWorkOperations {
			b.err = ErrTreeCapacity
			return false, b.err
		}
		b.operations++
		switch b.phase {
		case 0: // One expansion-state entry, including stable-key remap on new data.
			if b.cursor == len(b.data.nodes) {
				b.phase = 1
				b.cursor = 0
				continue
			}
			n := b.cursor
			b.cursor++
			value := b.options.ExpandAll
			if b.previous != nil && b.previous.data.stamp.Identity == b.data.stamp.Identity {
				if old, ok := b.previous.data.byID[b.data.nodes[n].ID]; ok {
					value = b.previous.expanded[old]
				}
			}
			if b.previous != nil {
				switch b.change.Kind {
				case OutlineExpandAll:
					value = true
				case OutlineCollapseAll:
					value = false
				case OutlineToggle, OutlineExpand, OutlineCollapse:
					if b.data.nodes[n].ID == b.change.ID {
						switch b.change.Kind {
						case OutlineToggle:
							value = !value
						case OutlineExpand:
							value = true
						case OutlineCollapse:
							value = false
						}
					}
				}
			}
			b.expanded[n] = value
		case 1: // Matching every label once, independent of collapse state.
			if b.cursor == len(b.data.nodes) {
				b.phase = 2
				b.cursor = len(b.data.preorder) - 1
				continue
			}
			n := b.cursor
			b.cursor++
			node := b.data.nodes[n]
			b.included[n] = b.filter == "" || strings.Contains(node.ID, b.filter) || strings.Contains(node.Label, b.filter)
		case 2: // Reverse preorder marks ancestors in O(N), never climbs per match.
			if b.cursor < 0 {
				b.phase = 3
				b.stack = append(b.stack, treeIndexFrame{b.data.first})
				continue
			}
			n := b.data.preorder[b.cursor]
			b.cursor--
			if b.included[n] && b.data.nodes[n].parent >= 0 {
				b.included[b.data.nodes[n].parent] = true
			}
		case 3:
			if len(b.stack) == 0 {
				i := &TreeIndex{token: &treeIndexToken{}, data: b.data, expanded: b.expanded, visible: b.visible, positions: b.positions, filter: b.filter, request: b.request, base: b.base, stats: TreeIndexStats{len(b.data.nodes), len(b.visible), b.storage, b.operations}}
				i.self = i
				b.result = i
				// Completion retains no previous index and no traversal/match storage.
				b.previous = nil
				b.included = nil
				b.stack = nil
				return true, nil
			}
			top := &b.stack[len(b.stack)-1]
			if top.next < 0 {
				b.stack = b.stack[:len(b.stack)-1]
				continue
			}
			n := top.next
			top.next = b.data.nodes[n].next
			if !b.included[n] {
				continue
			}
			if len(b.visible) == b.data.limits.MaxVisibleNodes {
				b.err = ErrTreeCapacity
				return false, b.err
			}
			b.visible = append(b.visible, n)
			b.positions[n] = len(b.visible)
			if b.data.nodes[n].first >= 0 && (b.filter != "" || b.expanded[n]) {
				b.stack = append(b.stack, treeIndexFrame{b.data.nodes[n].first})
			}
		}
	}
	return false, nil
}
func (b *TreeIndexBuilder) Result() (*TreeIndex, error) {
	if err := b.enter(); err != nil {
		return nil, err
	}
	defer b.busy.Store(false)
	if b.err != nil {
		return nil, b.err
	}
	if err := b.ctx.Err(); err != nil {
		return nil, err
	}
	if b.result == nil {
		return nil, ErrTreeNotReady
	}
	return b.result, nil
}
func (b *TreeIndexBuilder) Progress() (TreeBuildProgress, error) {
	if err := b.enter(); err != nil {
		return TreeBuildProgress{}, err
	}
	defer b.busy.Store(false)
	return TreeBuildProgress{b.operations, len(b.visible), b.result != nil}, nil
}
func BuildTreeIndex(ctx context.Context, data *TreeData, options TreeIndexOptions) (*TreeIndex, error) {
	b, err := NewTreeIndexBuilder(ctx, data, options)
	if err != nil {
		return nil, err
	}
	for {
		done, err := b.Step(256)
		if err != nil {
			return nil, err
		}
		if done {
			return b.Result()
		}
	}
}
