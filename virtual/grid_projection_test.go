package virtual

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unsafe"
)

type projectionTestSource struct {
	rows     [][]string
	revision uint64
	hook     func(string, int) error
	metadata func(string)
}

func (s *projectionTestSource) Identity() string {
	if s.metadata != nil {
		s.metadata("identity")
	}
	return "loaded-domain"
}
func (s *projectionTestSource) Revision() uint64 {
	r := s.revision
	if s.metadata != nil {
		s.metadata("revision")
	}
	return r
}
func (s *projectionTestSource) RowCount() int {
	n := len(s.rows)
	if s.metadata != nil {
		s.metadata("rows")
	}
	return n
}
func (s *projectionTestSource) RowKey(i int) (string, error) {
	if s.hook != nil {
		if err := s.hook("key", i); err != nil {
			return "", err
		}
	}
	if i < 0 || i >= len(s.rows) {
		return "", ErrGridRange
	}
	return fmt.Sprintf("r:%d", i), nil
}
func (s *projectionTestSource) IndexOf(k string) (int, bool, error) {
	i, e := strconv.Atoi(strings.TrimPrefix(k, "r:"))
	if s.hook != nil {
		if err := s.hook("index", i); err != nil {
			return 0, false, err
		}
	}
	return i, e == nil && i >= 0 && i < len(s.rows) && k == fmt.Sprintf("r:%d", i), nil
}
func (s *projectionTestSource) Cell(k, c string) (string, error) {
	i, found, e := s.IndexOf(k)
	if e != nil {
		return "", e
	}
	if !found {
		return "", ErrGridRange
	}
	col, e := strconv.Atoi(strings.TrimPrefix(c, "c:"))
	if e != nil || col < 0 || col >= len(s.rows[i]) {
		return "", ErrGridRange
	}
	if s.hook != nil {
		if e := s.hook("cell", i); e != nil {
			return "", e
		}
	}
	return s.rows[i][col], nil
}
func projectionSource(n int) *projectionTestSource {
	s := &projectionTestSource{revision: 7, rows: make([][]string, n)}
	for i := range s.rows {
		s.rows[i] = []string{[]string{"é", "Z", "a", "日本", "A"}[i%5], strconv.Itoa((i*17)%23 - 11), []string{"keep", "Keep", "skip"}[i%3]}
	}
	return s
}
func projectionKeys(t *testing.T, p GridProjection) []int {
	t.Helper()
	out := make([]int, p.RowCount())
	for i := range out {
		k, e := p.RowKey(i)
		if e != nil {
			t.Fatal(e)
		}
		out[i], e = strconv.Atoi(strings.TrimPrefix(k, "r:"))
		if e != nil {
			t.Fatal(e)
		}
		at, found, e := p.IndexOf(k)
		if e != nil || !found || at != i {
			t.Fatal("not inverse", k, at, found, e)
		}
	}
	return out
}
func projectionOracle(s *projectionTestSource, q GridQuery) []int {
	order := []int{}
	for row, cells := range s.rows {
		include := true
		for _, f := range q.Filters {
			col, _ := strconv.Atoi(strings.TrimPrefix(f.ColumnID, "c:"))
			if f.Kind == GridFilterEquals {
				include = cells[col] == f.Text
			} else {
				include = strings.Contains(cells[col], f.Text)
			}
			if !include {
				break
			}
		}
		if include {
			order = append(order, row)
		}
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := s.rows[order[i]], s.rows[order[j]]
		for _, c := range q.Sort {
			col, _ := strconv.Atoi(strings.TrimPrefix(c.ColumnID, "c:"))
			cmp := strings.Compare(a[col], b[col])
			if c.Kind == GridSortInteger {
				x, _ := strconv.ParseInt(a[col], 10, 64)
				y, _ := strconv.ParseInt(b[col], 10, 64)
				cmp = 0
				if x < y {
					cmp = -1
				} else if x > y {
					cmp = 1
				}
			}
			if cmp != 0 {
				if c.Descending {
					return cmp > 0
				}
				return cmp < 0
			}
		}
		return false
	})
	return order
}
func TestGridProjectionBruteOracleStableTuplesAndInverse(t *testing.T) {
	s := projectionSource(257)
	queries := []GridQuery{{}, {Sort: []GridSortColumn{{ColumnID: "c:0"}}}, {Sort: []GridSortColumn{{ColumnID: "c:0", Descending: true}}}, {Sort: []GridSortColumn{{ColumnID: "c:1", Kind: GridSortInteger, Descending: true}, {ColumnID: "c:0"}}, Filters: []GridFilter{{ColumnID: "c:2", Kind: GridFilterEquals, Text: "keep"}}}, {Filters: []GridFilter{{ColumnID: "c:2", Text: "eep"}, {ColumnID: "c:0", Kind: GridFilterEquals, Text: "日本"}}}, {Filters: []GridFilter{{ColumnID: "c:2", Text: "absent"}}}}
	for i, q := range queries {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			p, e := BuildGridProjection(context.Background(), s, 3, q, GridProjectionLimits{})
			if e != nil {
				t.Fatal(e)
			}
			if got, want := projectionKeys(t, p), projectionOracle(s, q); !reflect.DeepEqual(got, want) {
				t.Fatal(got, want)
			}
			if p.InputStamp() != (GridStamp{Identity: s.Identity(), Revision: 7, Rows: 257}) || p.Revision() != 3 {
				t.Fatal("domain/view revisions mixed", p.InputStamp(), p.Revision())
			}
			for row := range s.rows {
				key := fmt.Sprintf("r:%d", row)
				_, found, e := p.IndexOf(key)
				if e != nil {
					t.Fatal(e)
				}
				present := false
				for _, n := range projectionOracle(s, q) {
					if n == row {
						present = true
					}
				}
				if found != present {
					t.Fatal("hidden index mismatch", row, found)
				}
			}
			snap, e := p.SelectionSnapshot()
			if e != nil || snap.RowCount() != p.RowCount() {
				t.Fatal("constant-time view selection", snap, e)
			}
		})
	}
}

type projectionMetricSource struct {
	*projectionTestSource
	heights *RowHeights
}

func (s *projectionMetricSource) RowOffset(i int) (int, error) {
	if s.hook != nil {
		if e := s.hook("offset", i); e != nil {
			return 0, e
		}
	}
	return s.heights.RowOffset(i)
}
func (s *projectionMetricSource) RowAt(i int) (int, error) {
	if s.hook != nil {
		if e := s.hook("at", i); e != nil {
			return 0, e
		}
	}
	return s.heights.RowAt(i)
}

type projectionEditSource struct{ *projectionTestSource }

func (s *projectionEditSource) EditCell(k, c string) (GridEditValue, error) {
	text, e := s.Cell(k, c)
	return GridEditValue{Text: text, Version: 9, Editable: true}, e
}

type projectionMetricEditSource struct{ *projectionMetricSource }

func (s *projectionMetricEditSource) EditCell(k, c string) (GridEditValue, error) {
	text, e := s.Cell(k, c)
	return GridEditValue{Text: text, Version: 9, Editable: true}, e
}
func TestGridProjectionCapabilityVariantsAndProjectedHeights(t *testing.T) {
	base := projectionSource(9)
	h, e := NewRowHeights([]int{1, 7, 3, 11, 5, 13, 2, 17, 4})
	if e != nil {
		t.Fatal(e)
	}
	metrics := &projectionMetricSource{base, h}
	inputs := []GridSource{base, metrics, &projectionEditSource{base}, &projectionMetricEditSource{metrics}}
	q := GridQuery{Sort: []GridSortColumn{{ColumnID: "c:1", Kind: GridSortInteger, Descending: true}}, Filters: []GridFilter{{ColumnID: "c:2", Text: "eep"}}}
	for i, s := range inputs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			p, e := BuildGridProjection(context.Background(), s, 8, q, GridProjectionLimits{})
			if e != nil {
				t.Fatal(e)
			}
			_, hasMetrics := p.(GridRowMetrics)
			_, hasEdit := p.(GridEditSource)
			if hasMetrics != (i == 1 || i == 3) || hasEdit != (i >= 2) {
				t.Fatal("optional capability changed", i, hasMetrics, hasEdit)
			}
			private, exact := unwrapGridProjection(p)
			if !exact {
				t.Fatal("own result refused")
			}
			integer := int(unsafe.Sizeof(int(0)))
			index := (cap(private.order) + cap(private.inverse) + cap(private.prefix)) * integer
			if p.Stats().IndexBytes != index {
				t.Fatal("index charge differs actual capacity", p.Stats(), index)
			}
			if m, ok := p.(GridRowMetrics); ok {
				offset := 0
				for row, original := range projectionKeys(t, p) {
					start, e := m.RowOffset(row)
					if e != nil || start != offset {
						t.Fatal(start, offset, e)
					}
					begin, _ := h.RowOffset(original)
					end, _ := h.RowOffset(original + 1)
					height := end - begin
					for pixel := offset; pixel < offset+height; pixel++ {
						got, e := m.RowAt(pixel)
						if e != nil || got != row {
							t.Fatal("projected pixel inverse", pixel, got, row, e)
						}
					}
					offset += height
				}
				if got, e := m.RowOffset(p.RowCount()); e != nil || got != offset {
					t.Fatal(got, offset, e)
				}
				if _, e := m.RowAt(offset); !errors.Is(e, ErrGridRange) {
					t.Fatal("extent is not a pixel", e)
				}
			}
			if edit, ok := p.(GridEditSource); ok {
				key, _ := p.RowKey(0)
				v, e := edit.EditCell(key, "c:1")
				if e != nil || v.Version != 9 || !v.Editable {
					t.Fatal(v, e)
				}
				if _, e := edit.EditCell("r:2", "c:1"); !errors.Is(e, ErrGridRange) {
					t.Fatal("hidden edit allowed", e)
				}
			}
		})
	}
	empty := projectionSource(0)
	emptyH, _ := NewRowHeights(nil)
	p, e := BuildGridProjection(context.Background(), &projectionMetricSource{empty, emptyH}, 1, GridQuery{}, GridProjectionLimits{})
	if e != nil {
		t.Fatal(e)
	}
	m := p.(GridRowMetrics)
	if total, e := m.RowOffset(0); total != 0 || e != nil {
		t.Fatal(total, e)
	}
	if _, e := m.RowAt(0); !errors.Is(e, ErrGridRange) {
		t.Fatal(e)
	}
}
func TestGridProjectionValidationBudgetAndIntegerGrammar(t *testing.T) {
	s := projectionSource(2)
	for _, q := range []GridQuery{{Sort: make([]GridSortColumn, 5)}, {Sort: []GridSortColumn{{ColumnID: "", Kind: GridSortText}}}, {Sort: []GridSortColumn{{ColumnID: "c:0", Kind: 9}}}, {Filters: []GridFilter{{ColumnID: "c:0", Text: string([]byte{255})}}}, {Filters: []GridFilter{{ColumnID: "c:0", Kind: 9}}}} {
		if p, e := BuildGridProjection(context.Background(), s, 1, q, GridProjectionLimits{}); p != nil || !errors.Is(e, ErrGridInvalid) {
			t.Fatal("bad query admitted", q, p, e)
		}
	}
	for _, l := range []GridProjectionLimits{{MaxRows: -1}, {MaxRows: 4<<20 + 1}, {MaxIndexBytes: 512<<20 + 1}, {MaxWorkingBytes: 1<<30 + 1}, {MaxCellBytes: 1<<20 + 1}, {MaxQueryBytes: 64<<10 + 1}} {
		if _, e := BuildGridProjection(context.Background(), s, 1, GridQuery{}, l); !errors.Is(e, ErrGridInvalid) {
			t.Fatal(l, e)
		}
	}
	q := GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0"}}}
	structure, index, _ := projectionFootprint(2, 1, 0, false)
	cost := structure + len("c:0") + len(s.Identity())
	for _, l := range []GridProjectionLimits{{MaxRows: 1}, {MaxIndexBytes: index - 1}, {MaxWorkingBytes: cost - 1}, {MaxQueryBytes: 2}, {MaxCellBytes: 1}, {MaxWorkingBytes: cost}} {
		if p, e := BuildGridProjection(context.Background(), s, 1, q, l); p != nil || !errors.Is(e, ErrGridCapacity) {
			t.Fatal("budget admitted", l, p, e)
		}
	}
	// A filter cannot make the worker bypass structural input-sized admission.
	if _, e := BuildGridProjection(context.Background(), s, 1, GridQuery{Filters: []GridFilter{{ColumnID: "c:0", Text: "never"}}}, GridProjectionLimits{MaxIndexBytes: index - 1}); !errors.Is(e, ErrGridCapacity) {
		t.Fatal(e)
	}
	for _, bad := range []string{"", "+", "-", " 1", "1 ", "1_0", "0x10", "1e2", "١", "9223372036854775808", "-9223372036854775809"} {
		s.rows[0][1] = bad
		if _, e := BuildGridProjection(context.Background(), s, 1, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:1", Kind: GridSortInteger}}}, GridProjectionLimits{}); !errors.Is(e, ErrGridSource) {
			t.Fatal("malformed numeric cell admitted", bad, e)
		}
	}
	for _, good := range []string{"+0001", "-0", "9223372036854775807", "-9223372036854775808"} {
		s.rows[0][1] = good
		if _, e := BuildGridProjection(context.Background(), s, 1, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:1", Kind: GridSortInteger}}}, GridProjectionLimits{}); e != nil {
			t.Fatal("strict legal int refused", good, e)
		}
	}
	s.rows[0][0] = string([]byte{255})
	if _, e := BuildGridProjection(context.Background(), s, 1, q, GridProjectionLimits{}); !errors.Is(e, ErrGridSource) {
		t.Fatal("invalid cell UTF8", e)
	}
	var nilSource *projectionTestSource
	for _, source := range []GridSource{nil, nilSource} {
		if _, e := BuildGridProjection(context.Background(), source, 1, GridQuery{}, GridProjectionLimits{}); !errors.Is(e, ErrGridSource) {
			t.Fatal("nil provider", e)
		}
	}
	if _, e := BuildGridProjection(nil, s, 1, GridQuery{}, GridProjectionLimits{}); !errors.Is(e, ErrGridInvalid) {
		t.Fatal(e)
	}
	if _, e := BuildGridProjection(context.Background(), s, 0, GridQuery{}, GridProjectionLimits{}); !errors.Is(e, ErrGridInvalid) {
		t.Fatal(e)
	}
}
func TestGridProjectionCallbacksCancellationErrorsAndDrift(t *testing.T) {
	cause := errors.New("provider failure")
	for _, operation := range []string{"key", "index", "cell", "offset", "at"} {
		t.Run(operation, func(t *testing.T) {
			s := projectionSource(512)
			h, _ := NewRowHeights(makeHeights(512, 3))
			source := &projectionMetricSource{s, h}
			s.hook = func(op string, i int) error {
				if op == operation {
					return cause
				}
				return nil
			}
			p, e := BuildGridProjection(context.Background(), source, 1, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0"}}}, GridProjectionLimits{})
			if p != nil || !errors.Is(e, ErrGridSource) || !errors.Is(e, cause) {
				t.Fatal("cause discarded", p, e)
			}
			ctx, cancel := context.WithCancel(context.Background())
			s.hook = func(op string, i int) error {
				if op == operation {
					cancel()
				}
				return nil
			}
			p, e = BuildGridProjection(ctx, source, 1, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0"}}}, GridProjectionLimits{})
			if p != nil || !errors.Is(e, context.Canceled) {
				t.Fatal("callback cancellation ignored", p, e)
			}
			s.hook = func(op string, i int) error {
				if op == operation {
					s.revision++
				}
				return nil
			}
			p, e = BuildGridProjection(context.Background(), source, 1, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0"}}}, GridProjectionLimits{})
			if p != nil || !errors.Is(e, ErrGridSource) || !errors.Is(e, ErrGridStale) {
				t.Fatal("callback drift admitted", p, e)
			}
		})
	}
	s := projectionSource(8)
	p, e := BuildGridProjection(context.Background(), s, 1, GridQuery{}, GridProjectionLimits{})
	if e != nil {
		t.Fatal(e)
	}
	s.revision++
	if _, e := p.RowKey(0); !errors.Is(e, ErrGridStale) {
		t.Fatal(e)
	}
	if _, _, e := p.IndexOf("r:0"); !errors.Is(e, ErrGridStale) {
		t.Fatal(e)
	}
	if _, e := p.Cell("r:0", "c:0"); !errors.Is(e, ErrGridStale) {
		t.Fatal(e)
	}
	if _, e := p.SelectionSnapshot(); !errors.Is(e, ErrGridStale) {
		t.Fatal(e)
	}
}
func makeHeights(n, height int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = height
	}
	return out
}

type projectionCancelComparisons struct {
	context.Context
	cancel context.CancelFunc
	armed  func() bool
	checks int
}

func (c *projectionCancelComparisons) Err() error {
	if c.armed() {
		c.checks++
		if c.checks == 30 {
			c.cancel()
		}
	}
	return c.Context.Err()
}
func TestGridProjectionSortCancellationAndQueryIsolation(t *testing.T) {
	s := projectionSource(1024)
	q := GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0"}}, Filters: []GridFilter{{ColumnID: "c:2", Text: "eep"}}}
	original := GridQuery{Sort: append([]GridSortColumn(nil), q.Sort...), Filters: append([]GridFilter(nil), q.Filters...)}
	once := false
	s.hook = func(op string, i int) error {
		if op == "cell" && !once {
			once = true
			q.Sort[0].Descending = true
			q.Filters[0].Text = "absent"
		}
		return nil
	}
	p, e := BuildGridProjection(context.Background(), s, 1, q, GridProjectionLimits{})
	if e != nil {
		t.Fatal(e)
	}
	if got, want := projectionKeys(t, p), projectionOracle(s, original); !reflect.DeepEqual(got, want) {
		t.Fatal("query borrowed caller arrays")
	}
	s = projectionSource(1024)
	reads := 0
	s.hook = func(op string, i int) error {
		if op == "cell" {
			reads++
		}
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checked := &projectionCancelComparisons{Context: ctx, cancel: cancel, armed: func() bool { return reads == len(s.rows) }}
	p, e = BuildGridProjection(checked, s, 1, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0"}}}, GridProjectionLimits{})
	if p != nil || !errors.Is(e, context.Canceled) || checked.checks != 30 {
		t.Fatal("sort kept comparing after cancellation", p, e, checked.checks)
	}
}

type projectionMalformedIndex struct{ *projectionTestSource }

func (s *projectionMalformedIndex) IndexOf(k string) (int, bool, error) {
	i, ok, e := s.projectionTestSource.IndexOf(k)
	return i + 1, ok, e
}

type projectionMalformedMetrics struct{ *projectionMetricSource }

func (s *projectionMalformedMetrics) RowAt(i int) (int, error) {
	at, e := s.projectionMetricSource.RowAt(i)
	return at + 1, e
}
func TestGridProjectionMalformedIndicesAndMetadataCallbackDrift(t *testing.T) {
	s := projectionSource(8)
	if _, e := BuildGridProjection(context.Background(), &projectionMalformedIndex{s}, 1, GridQuery{}, GridProjectionLimits{}); !errors.Is(e, ErrGridSource) {
		t.Fatal("inverse index malformed", e)
	}
	h, _ := NewRowHeights(makeHeights(8, 3))
	if _, e := BuildGridProjection(context.Background(), &projectionMalformedMetrics{&projectionMetricSource{s, h}}, 1, GridQuery{}, GridProjectionLimits{}); !errors.Is(e, ErrGridSource) {
		t.Fatal("metric inverse malformed", e)
	}
	s = projectionSource(2)
	p, e := BuildGridProjection(context.Background(), s, 1, GridQuery{}, GridProjectionLimits{})
	if e != nil {
		t.Fatal(e)
	}
	once := false
	s.metadata = func(op string) {
		if op == "rows" && !once {
			once = true
			s.revision++
		}
	}
	if _, e := p.SelectionSnapshot(); !errors.Is(e, ErrGridStale) {
		t.Fatal("metadata callback changed stamp after captured return", e)
	}
}

type projectionExternalLikeWrapper struct{ GridProjection }

func (p projectionExternalLikeWrapper) Revision() uint64 { return math.MaxUint64 }
func TestGridProjectionUnwrapConcurrentImmutableAndChargedStats(t *testing.T) {
	s := projectionSource(100)
	q := GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0"}, {ColumnID: "c:1", Kind: GridSortInteger}}, Filters: []GridFilter{{ColumnID: "c:2", Text: "eep"}}}
	p, e := BuildGridProjection(context.Background(), s, 9, q, GridProjectionLimits{})
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := unwrapGridProjection(projectionExternalLikeWrapper{p}); ok {
		t.Fatal("inherited marker wrapper admitted")
	}
	var plain *gridProjectionBase
	var metric *gridProjectionMetrics
	var edit *gridProjectionEdit
	var both *gridProjectionMetricsEdit
	for _, value := range []GridProjection{nil, plain, metric, edit, both, &gridProjectionMetrics{}, &gridProjectionEdit{}, &gridProjectionMetricsEdit{}} {
		if _, ok := unwrapGridProjection(value); ok {
			t.Fatal("typed-nil/invalid private variant admitted")
		}
	}
	private, _ := unwrapGridProjection(p)
	structure, index, _ := projectionFootprint(100, 2, 1, false)
	queryBytes, _ := projectionQuerySize(q, 16<<10)
	payload := 0
	for _, i := range private.order {
		payload += len(s.rows[i][0]) + len(s.rows[i][1])
	}
	want := structure + queryBytes + len(s.Identity()) + payload
	if stats := p.Stats(); stats.IndexBytes != index || stats.WorkingBytes != want || stats.CellReads != uint64(100+len(private.order)*2) || stats.SortComparisons == 0 {
		t.Fatal("inaccurate logical statistics", stats, index, want)
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repeat := 0; repeat < 20; repeat++ {
				for row := 0; row < p.RowCount(); row++ {
					k, e := p.RowKey(row)
					if e != nil {
						t.Error(e)
						return
					}
					at, found, e := p.IndexOf(k)
					if e != nil || !found || at != row {
						t.Error(at, found, e)
						return
					}
					if _, e := p.Cell(k, "c:0"); e != nil {
						t.Error(e)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

type projectionHugeSource struct {
	*projectionTestSource
	count int
}

func (s *projectionHugeSource) RowCount() int { return s.count }

type projectionMetricBoundaryFault struct {
	*projectionMetricSource
	fault string
}

func (s *projectionMetricBoundaryFault) RowOffset(i int) (int, error) {
	if s.fault == "zero" && i == 1 {
		return 0, nil
	}
	if s.fault == "negative" && i == 1 {
		return -1, nil
	}
	if s.fault == "origin" && i == 0 {
		return 1, nil
	}
	return s.projectionMetricSource.RowOffset(i)
}
func TestGridProjectionPreallocationAdmissionAndProviderBoundaries(t *testing.T) {
	base := projectionSource(1)
	calls := 0
	base.hook = func(string, int) error { calls++; return nil }
	for _, source := range []GridSource{&projectionHugeSource{base, math.MaxInt}, &projectionHugeSource{base, 1_000_000}} {
		limits := GridProjectionLimits{}
		if source.RowCount() == 1_000_000 {
			limits.MaxIndexBytes = 1
		}
		if _, err := BuildGridProjection(context.Background(), source, 1, GridQuery{}, limits); !errors.Is(err, ErrGridCapacity) {
			t.Fatal(err)
		}
	}
	if calls != 0 {
		t.Fatal("scanned rows before structural admission", calls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := BuildGridProjection(ctx, base, 1, GridQuery{}, GridProjectionLimits{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("called provider after prior cancellation")
	}
	base = projectionSource(2)
	h, _ := NewRowHeights([]int{2, 3})
	metrics := &projectionMetricSource{base, h}
	for _, fault := range []string{"zero", "negative", "origin"} {
		if _, err := BuildGridProjection(context.Background(), &projectionMetricBoundaryFault{metrics, fault}, 1, GridQuery{}, GridProjectionLimits{}); !errors.Is(err, ErrGridSource) {
			t.Fatal(fault, err)
		}
	}
	source := &projectionEditSource{projectionSource(3)}
	p, err := BuildGridProjection(context.Background(), source, 1, GridQuery{}, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{"", strings.Repeat("c", 1025), string([]byte{255})} {
		if _, err := p.Cell("r:0", column); !errors.Is(err, ErrGridRange) {
			t.Fatal(column, err)
		}
		if _, err := p.(GridEditSource).EditCell("r:0", column); !errors.Is(err, ErrGridRange) {
			t.Fatal(column, err)
		}
	}
	source.rows[0][0] = strings.Repeat("x", 64<<10+1)
	if _, err := p.Cell("r:0", "c:0"); !errors.Is(err, ErrGridCapacity) {
		t.Fatal("runtime cell bound", err)
	}
	source.rows[0][0] = "bad\x00text"
	if _, err := p.(GridEditSource).EditCell("r:0", "c:0"); !errors.Is(err, ErrGridSource) {
		t.Fatal("NUL edit baseline", err)
	}
}
func TestGridProjectionDescendingNumericTiesAndConstantSelectionCapture(t *testing.T) {
	source := &projectionTestSource{revision: 1, rows: [][]string{{"x", "+0001"}, {"x", "1"}, {"x", "-0"}, {"x", "0"}, {"x", "0001"}}}
	p, err := BuildGridProjection(context.Background(), source, 10, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:1", Kind: GridSortInteger, Descending: true}}}, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got := projectionKeys(t, p); !reflect.DeepEqual(got, []int{0, 1, 4, 2, 3}) {
		t.Fatal("descending reversed ties", got)
	}
	indexCalls := 0
	metadataCalls := 0
	source.hook = func(string, int) error { indexCalls++; return nil }
	source.metadata = func(string) { metadataCalls++ }
	snapshot, err := p.SelectionSnapshot()
	if err != nil || snapshot == nil {
		t.Fatal(err)
	}
	if indexCalls != 0 || metadataCalls != 6 {
		t.Fatal("selection capture enumerated index", indexCalls, metadataCalls)
	}
}

type projectionEditFault struct {
	*projectionTestSource
	value GridEditValue
	err   error
	hook  func()
}

func (s *projectionEditFault) EditCell(string, string) (GridEditValue, error) {
	if s.hook != nil {
		s.hook()
	}
	return s.value, s.err
}

type projectionIdentitySource struct {
	*projectionTestSource
	identity string
}

func (s *projectionIdentitySource) Identity() string { return s.identity }

func TestGridProjectionMetadataCancellationAndOptionalProviderFailures(t *testing.T) {
	for _, operation := range []string{"identity", "revision", "rows"} {
		t.Run(operation, func(t *testing.T) {
			source := projectionSource(3)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var calls []string
			source.metadata = func(op string) {
				calls = append(calls, op)
				if op == operation {
					cancel()
				}
			}
			if p, err := BuildGridProjection(ctx, source, 1, GridQuery{}, GridProjectionLimits{}); p != nil || !errors.Is(err, context.Canceled) {
				t.Fatal("metadata cancellation ignored", p, err)
			}
			if calls[len(calls)-1] != operation {
				t.Fatal("called metadata provider after cancellation", calls)
			}
		})
	}
	base := projectionSource(2)
	cause := errors.New("edit callback failure")
	source := &projectionEditFault{projectionTestSource: base, value: GridEditValue{Text: base.rows[0][0], Version: 1, Editable: true}}
	p, err := BuildGridProjection(context.Background(), source, 1, GridQuery{}, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	edit := p.(GridEditSource)
	source.err = cause
	if _, err := edit.EditCell("r:0", "c:0"); !errors.Is(err, ErrGridSource) || !errors.Is(err, cause) {
		t.Fatal("edit cause discarded", err)
	}
	source.err = nil
	for _, bad := range []GridEditValue{{Text: base.rows[0][0]}, {Text: "different", Version: 1}, {Text: string([]byte{255}), Version: 1}} {
		source.value = bad
		if _, err := edit.EditCell("r:0", "c:0"); !errors.Is(err, ErrGridSource) {
			t.Fatal("invalid edit provider admitted", bad, err)
		}
	}
	source.value = GridEditValue{Text: base.rows[0][0], Version: 1}
	source.hook = func() { base.revision++ }
	if _, err := edit.EditCell("r:0", "c:0"); !errors.Is(err, ErrGridSource) || !errors.Is(err, ErrGridStale) {
		t.Fatal("edit callback drift admitted", err)
	}
	base = projectionSource(2)
	heights, err := NewRowHeights([]int{3, 5})
	if err != nil {
		t.Fatal(err)
	}
	p, err = BuildGridProjection(context.Background(), &projectionMetricSource{base, heights}, 1, GridQuery{}, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	base.revision++
	metrics := p.(GridRowMetrics)
	if _, err := metrics.RowOffset(0); !errors.Is(err, ErrGridStale) {
		t.Fatal("metric boundary ignored drift", err)
	}
	if _, err := metrics.RowAt(0); !errors.Is(err, ErrGridStale) {
		t.Fatal("metric inverse ignored drift", err)
	}
}

func TestGridProjectionLengthAdmissionAndDetachedMetadata(t *testing.T) {
	source := projectionSource(2)
	oversizedMalformed := strings.Repeat("x", 100) + string([]byte{255})
	query := GridQuery{Filters: []GridFilter{{ColumnID: "c:0", Text: oversizedMalformed}}}
	if _, err := BuildGridProjection(context.Background(), source, 1, query, GridProjectionLimits{MaxQueryBytes: 10}); !errors.Is(err, ErrGridCapacity) {
		t.Fatal("validated malformed payload before length admission", err)
	}
	source.rows[0][0] = oversizedMalformed
	if _, err := BuildGridProjection(context.Background(), source, 1, GridQuery{Sort: []GridSortColumn{{ColumnID: "c:0"}}}, GridProjectionLimits{MaxCellBytes: 10}); !errors.Is(err, ErrGridCapacity) {
		t.Fatal("validated malformed cell before length admission", err)
	}
	// The result must not retain a tiny identity substring's large backing store.
	backing := strings.Repeat("x", 1<<20)
	identity := backing[100:110]
	s := &projectionIdentitySource{projectionSource(2), identity}
	p, err := BuildGridProjection(context.Background(), s, 1, GridQuery{}, GridProjectionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if p.Identity() != identity || unsafe.StringData(p.Identity()) == unsafe.StringData(identity) {
		t.Fatal("identity was not detached")
	}
}
