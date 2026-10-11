package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"runtime"
	"strconv"

	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/virtual"
)

const perfPeriod = 64
const perfMinCells = 400

type perfConfig struct {
	Rows            int
	Variable        bool
	Samples, Warmup int
	Novel           bool
	CPUProfile      string
}

func (c perfConfig) validate() error {
	if !validPerfProfilePath(c.CPUProfile) {
		return errors.New("loadedgrid perf: invalid CPU profile path")
	}
	if c.Rows != 10_000 && c.Rows != 1_000_000 || c.Samples < 120 || c.Samples > 2048 || c.Warmup < perfPeriod || c.Warmup > 2048 || c.Warmup%perfPeriod != 0 {
		return errors.New("loadedgrid perf: rows must be10000 or1000000, samples120..2048 and warmup64..2048")
	}
	return nil
}

// The benchmark wrapper changes only indexed stable-key heights. Stored values,
// domain identity, projection machinery and diagnostic source counters are the
// same immutable fullSource used by the ordinary consumer.
type perfVariableSource struct{ *fullSource }

var perfHeights = [4]int{9, 11, 13, 15}

func perfOffset(boundary int) int {
	cycles, remainder := boundary/4, boundary%4
	total := cycles * 48
	for i := 0; i < remainder; i++ {
		total += perfHeights[i]
	}
	return total
}
func (s *perfVariableSource) RowOffset(boundary int) (int, error) {
	if boundary < 0 || boundary > s.rows {
		return 0, virtual.ErrGridRange
	}
	return perfOffset(boundary), nil
}
func (s *perfVariableSource) RowAt(pixel int) (int, error) {
	if pixel < 0 || pixel >= perfOffset(s.rows) {
		return 0, virtual.ErrGridRange
	}
	row, remainder := pixel/48*4, pixel%48
	for _, h := range perfHeights {
		if remainder < h {
			return row, nil
		}
		row++
		remainder -= h
	}
	return 0, virtual.ErrGridSource
}
func perfInput(s *fullSource, variable bool) virtual.GridSource {
	if variable {
		return &perfVariableSource{s}
	}
	return s
}
func perfColumns() []virtual.GridColumn {
	result := gridColumns()
	for i := range result {
		result[i].Width = 96
	}
	return result
}

// A zero-sum closed trajectory warms and reuses a bounded set of rows and
// columns. The separately labelled novel case moves into an untouched range.
func perfDelta(index int) (x, y int) {
	i := index % perfPeriod
	if i < 32 {
		y = 12
	} else {
		y = -12
	}
	switch i {
	case 7, 15, 23, 31:
		x = 96
	case 39, 47, 55, 63:
		x = -96
	}
	return
}

type perfGeometry struct {
	MaterializedCells int                  `json:"materialized_cells"`
	RetainedCells     int                  `json:"retained_cells"`
	RetainedHeaders   int                  `json:"retained_headers"`
	Cells             int                  `json:"cells"`
	Rows              int                  `json:"rows"`
	Columns           int                  `json:"columns"`
	Viewport          virtual.GridViewport `json:"viewport"`
	Bounds            image.Rectangle      `json:"bounds"`
	Scale             float32              `json:"scale"`
	LayoutSHA256      string               `json:"layout_sha256"`
}

// Geometry is measured after the real native frame. Overscan records with empty
// clips do not count toward the fixture; distinct intersecting cells do.
func checkPerfGeometry(g cogent.GridInspection) (perfGeometry, error) {
	out := perfGeometry{Viewport: g.Viewport, Bounds: g.Bounds, Scale: g.Scale, MaterializedCells: len(g.Cells), RetainedCells: g.RetainedCells, RetainedHeaders: g.RetainedHeaders}
	rows, cols := map[int]bool{}, map[int]bool{}
	seen := map[string]bool{}
	type rectangle struct {
		Row, Column int
		Bounds      image.Rectangle
	}
	rectangles := make([]rectangle, 0, len(g.Cells))
	for _, c := range g.Cells {
		if c.Bounds.Intersect(g.Bounds).Empty() {
			continue
		}
		if c.Bounds != c.Bounds.Intersect(g.Bounds) {
			return out, errors.New("loadedgrid perf: measured cell escaped grid clip")
		}
		if seen[c.ID] {
			return out, errors.New("loadedgrid perf: duplicate rendered cell identity")
		}
		seen[c.ID] = true
		rows[c.Row] = true
		cols[c.Column] = true
		out.Cells++
		rectangles = append(rectangles, rectangle{c.Row, c.Column, c.Bounds})
	}
	out.Rows, out.Columns = len(rows), len(cols)
	if out.Cells < perfMinCells || out.Rows < 50 || out.Columns < 8 || g.Frames == 0 || g.Generation == 0 || g.Bounds.Empty() {
		return out, fmt.Errorf("loadedgrid perf: rendered workload too small: cells%d rows%d columns%d", out.Cells, out.Rows, out.Columns)
	}
	// Hash geometry relative to the grid, excluding source row ordinals. Paired
	// rowcounts at the same trajectory phase must have identical measured clips.
	normalized := make([]image.Rectangle, len(rectangles))
	for i, r := range rectangles {
		normalized[i] = r.Bounds.Sub(g.Bounds.Min)
	}
	b, e := json.Marshal(normalized)
	if e != nil {
		return out, e
	}
	sum := sha256.Sum256(b)
	out.LayoutSHA256 = hex.EncodeToString(sum[:])
	return out, nil
}

// Verify only the bounded completed viewport against the retained immutable
// stored array. This observer neither calls source.Cell nor scans full rows.
func verifyPerfCells(source *fullSource, inspection cogent.GridInspection) error {
	for _, c := range inspection.Cells {
		if c.Bounds.Intersect(inspection.Bounds).Empty() {
			continue
		}
		value, _, err := source.value(c.RowKey, c.ColumnID)
		if err != nil {
			return err
		}
		if c.Text != strconv.FormatInt(value, 10) {
			return errors.New("loadedgrid perf: rendered text differs from stored value")
		}
	}
	return nil
}

type perfMemory struct {
	HeapAlloc, HeapInuse, HeapSys, TotalAlloc, Mallocs uint64
	NumGC                                              uint32
	GCCPUFraction                                      float64
	Goroutines                                         int
}

func readPerfMemory() perfMemory {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return perfMemory{m.HeapAlloc, m.HeapInuse, m.HeapSys, m.TotalAlloc, m.Mallocs, m.NumGC, m.GCCPUFraction, runtime.NumGoroutine()}
}
