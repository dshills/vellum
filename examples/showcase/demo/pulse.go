package demo

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

type Service struct {
	ID, Name, Environment, Status string
	Latency                       float64
	Requests                      int
	Errors                        float64
}
type serviceSource struct {
	revision uint64
	rows     []Service
	indices  map[string]int
}

func (s *serviceSource) Identity() string { return "pulse-fixture" }
func (s *serviceSource) Revision() uint64 { return s.revision }
func (s *serviceSource) RowCount() int    { return len(s.rows) }
func (s *serviceSource) RowKey(i int) (string, error) {
	if i < 0 || i >= len(s.rows) {
		return "", virtual.ErrGridRange
	}
	return s.rows[i].ID, nil
}
func (s *serviceSource) IndexOf(k string) (int, bool, error) {
	i, ok := s.indices[k]
	return i, ok, nil
}
func (s *serviceSource) SelectionSnapshot() (virtual.GridRowIndex, error) { return s, nil }
func (s *serviceSource) Cell(k, c string) (string, error) {
	i, ok := s.indices[k]
	if !ok {
		return "", virtual.ErrGridRange
	}
	r := s.rows[i]
	switch c {
	case "service":
		return r.Name, nil
	case "environment":
		return r.Environment, nil
	case "status":
		return r.Status, nil
	case "latency":
		return fmt.Sprintf("%.1f ms", r.Latency), nil
	case "requests":
		return fmt.Sprintf("%s /m", thousands(r.Requests)), nil
	case "errors":
		return fmt.Sprintf("%.2f%%", r.Errors), nil
	}
	return "", virtual.ErrGridRange
}
func thousands(i int) string {
	if i >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(i)/1000000)
	}
	if i >= 1000 {
		return fmt.Sprintf("%.1fk", float64(i)/1000)
	}
	return fmt.Sprint(i)
}

type Pulse struct {
	owner
	source                                 *serviceSource
	grid                                   *virtual.Grid
	selected, query, filter, period, order string
	sample, focus                          int
	surface                                *cogent.MeshCanvas
	view                                   canvas.Viewport
}

func NewPulse() (*Pulse, error) {
	m := &Pulse{selected: "svc-0", filter: "all", period: "24h", order: "name", focus: 39, view: canvas.Viewport{Width: 760, Height: 200, Zoom: 1}}
	m.source = m.project(1)
	g, e := virtual.NewGrid(m.source, []virtual.GridColumn{{ID: "service", Title: "Service", Width: 128}, {ID: "environment", Title: "Environment", Width: 70}, {ID: "status", Title: "Status", Width: 65}, {ID: "latency", Title: "P95 latency", Width: 72}, {ID: "requests", Title: "Requests", Width: 76}, {ID: "errors", Title: "Errors", Width: 59}}, virtual.GridConfig{RowHeight: 22, HeaderHeight: 22, FrozenColumns: 1, MaxRows: 80, MaxCells: 1024, RowSelection: true, OverscanRows: 2}, virtual.GridViewport{Width: 770, Height: 260})
	if e != nil {
		return nil, e
	}
	m.grid = g
	if e = g.JumpToKey(m.selected, "service"); e != nil {
		return nil, e
	}
	m.surface = &cogent.MeshCanvas{View: &m.view, Background: rgb(29, 33, 43), Prepare: m.prepareChart, Pick: m.pickChart}
	m.register("pulse.next", "Next sample", "Meta+R", nil, func() error { return m.NextSample() })
	m.register("pulse.reset", "Reset filters", "Meta+0", nil, func() error { m.query = ""; m.filter = "all"; m.order = "name"; m.period = "24h"; return m.replace() })
	return m, nil
}
func (m *Pulse) Bind(a *app.Application, r app.WindowRef) {
	m.bind(a, r)
	m.surface.Owner = a
	m.surface.Window = r.ID()
}
func (m *Pulse) project(revision uint64) *serviceSource {
	s := &serviceSource{revision: revision, indices: map[string]int{}}
	names := []string{"api-gateway", "identity", "asset-pipeline", "search-index", "event-stream", "edge-cache", "billing", "notebook-sync"}
	for i := range 1024 {
		name := names[i%len(names)]
		if i >= len(names) {
			name = fmt.Sprintf("worker-%03d / %s", i/len(names), name)
		}
		status := "Healthy"
		if i%11 == 4 {
			status = "Degraded"
		}
		environment := "Production"
		if i%7 == 6 {
			environment = "Staging"
		}
		r := Service{ID: fmt.Sprintf("svc-%d", i), Name: name, Environment: environment, Status: status, Latency: 22 + float64((i*17+m.sample*7)%84)/2.2, Requests: 900 + (i*713+m.sample*113)%9000, Errors: float64(i%11) / 100}
		if status == "Degraded" {
			r.Latency *= 2.4
			r.Errors += .8
		}
		if m.filter == "degraded" && status != "Degraded" {
			continue
		}
		if m.filter == "production" && environment != "Production" {
			continue
		}
		if !strings.Contains(strings.ToLower(name), strings.ToLower(m.query)) {
			continue
		}
		s.rows = append(s.rows, r)
	}
	sort.SliceStable(s.rows, func(i, j int) bool {
		if m.order == "latency" {
			return s.rows[i].Latency > s.rows[j].Latency
		}
		return s.rows[i].Name < s.rows[j].Name
	})
	for i, r := range s.rows {
		s.indices[r.ID] = i
	}
	return s
}
func (m *Pulse) replace() error {
	candidate := m.project(m.source.revision + 1)
	if e := m.grid.ReplaceSource(candidate); e != nil {
		return e
	}
	m.source = candidate
	if _, ok := candidate.indices[m.selected]; !ok {
		m.selected = ""
		if len(candidate.rows) > 0 {
			m.selected = candidate.rows[0].ID
		}
	}
	if m.selected != "" {
		if e := m.grid.JumpToKey(m.selected, "service"); e != nil {
			return e
		}
	}
	m.surface.Revision++
	m.changed()
	return nil
}
func (m *Pulse) SetSearch(q string) error {
	if len(q) > 128 || !utf8.ValidString(q) || strings.ContainsRune(q, 0) {
		return errors.New("search exceeds128 bytes")
	}
	before := m.query
	m.query = q
	if e := m.replace(); e != nil {
		m.query = before
		return e
	}
	return nil
}
func (m *Pulse) SetFilter(f string) error {
	if f != "all" && f != "degraded" && f != "production" {
		return errors.New("unknown service filter")
	}
	before := m.filter
	m.filter = f
	if e := m.replace(); e != nil {
		m.filter = before
		return e
	}
	return nil
}
func (m *Pulse) SetPeriod(p string) error {
	if p != "1h" && p != "24h" && p != "7d" {
		return errors.New("unknown period")
	}
	m.period = p
	m.surface.Revision++
	m.changed()
	return nil
}
func (m *Pulse) NextSample() error {
	before := m.sample
	m.sample = (m.sample + 1) % 512
	if e := m.replace(); e != nil {
		m.sample = before
		return e
	}
	return nil
}
func (m *Pulse) service() Service {
	if i, ok := m.source.indices[m.selected]; ok {
		return m.source.rows[i]
	}
	return Service{Name: "No matching service", Status: "No data"}
}
func (m *Pulse) latency(i int, comparison bool) float64 {
	r := m.service()
	phase := float64(m.sample) * .31
	period := map[string]float64{"1h": .23, "24h": .14, "7d": .08}[m.period]
	v := r.Latency * (.82 + .16*math.Sin(float64(i)*period+phase) + .08*math.Sin(float64(i)*.57))
	v += r.Latency * .65 * math.Exp(-math.Pow((float64(i)-43)/4.5, 2))
	if comparison {
		v = r.Latency * (.65 + .14*math.Sin(float64(i)*period+1.5))
	}
	return v
}
func (m *Pulse) pickChart(p canvas.Point) {
	left := m.view.Center.X - m.view.Width/(2*m.view.Zoom)
	m.focus = max(0, min(63, int(math.Round((p.X-left)/(m.view.Width/m.view.Zoom)*63))))
	m.surface.Revision++
	m.changed()
}
func (m *Pulse) prepareChart(v canvas.Viewport) (*canvas.Mesh, error) {
	b := &meshBuilder{}
	world := func(x, y float64) canvas.Point { return v.World(canvas.Point{X: x, Y: y}) }
	w, h := v.Width, v.Height
	for j := range 5 {
		y := float64(j)*(h-16)/4 + 8
		b.stroke([]canvas.Point{world(0, y), world(w, y)}, 1/v.Zoom, rgb(51, 57, 72))
	}
	for k := range 2 {
		points := []canvas.Point{}
		for i := range 64 {
			x := w * float64(i) / 63
			y := h - 16 - m.latency(i, k == 1)/math.Max(1, m.service().Latency*1.8)*(h-28)
			points = append(points, world(x, y))
		}
		c := rgb(80, 210, 218)
		if k == 1 {
			c = rgb(145, 125, 234)
		}
		if k == 0 {
			for i := 0; i < len(points)-1; i++ {
				j := uint32(len(b.v))
				low := world(w*float64(i)/63, h)
				high := world(w*float64(i+1)/63, h)
				b.v = append(b.v, canvas.Vertex{Position: points[i], Color: colorAlpha(c, 64)}, canvas.Vertex{Position: points[i+1], Color: colorAlpha(c, 64)}, canvas.Vertex{Position: high, Color: colorAlpha(c, 2)}, canvas.Vertex{Position: low, Color: colorAlpha(c, 2)})
				b.ix = append(b.ix, j, j+1, j+2, j, j+2, j+3)
			}
		}
		b.stroke(points, 2.5/v.Zoom, c)
	}
	x := w * float64(m.focus) / 63
	b.stroke([]canvas.Point{world(x, 0), world(x, h)}, 1/v.Zoom, rgb(114, 129, 151))
	p := world(x, h-16-m.latency(m.focus, false)/math.Max(1, m.service().Latency*1.8)*(h-28))
	b.circle(p.X, p.Y, 5/v.Zoom, rgb(80, 210, 218))
	return b.mesh()
}
func colorAlpha(c color.NRGBA, a uint8) color.NRGBA { c.A = a; return c }
func (m *Pulse) Toolbar() []cogent.TitlebarToolbarItem {
	return []cogent.TitlebarToolbarItem{{ID: "sample", Label: "Next sample", Kind: ui.Button, PNG: toolbarIcon(rgb(66, 191, 189), "add"), Command: "pulse.next"}, {ID: "space", Kind: ui.ToolbarFlexibleSpace}, {ID: "reset", Label: "Reset filters", Kind: ui.Button, PNG: toolbarIcon(rgb(143, 117, 229), "reset"), Command: "pulse.reset"}}
}
func (m *Pulse) View() ui.Node {
	primary := rgb(234, 238, 246)
	secondary := rgb(138, 148, 170)
	border := rgb(49, 55, 72)
	bg := rgb(22, 25, 34)
	panel := rgb(29, 33, 43)
	teal := rgb(80, 210, 218)
	darkLabel := func(id, s string, size float32, weight int) ui.Node { return label(id, s, size, weight, primary) }
	navItem := func(id, title, filter string, c color.NRGBA) ui.Node {
		n := nav(id, title, m.filter == filter, c, func() { _ = m.SetFilter(filter) })
		n.Style.Foreground = ui.Some(secondary)
		if m.filter == filter {
			n.Style.Foreground = ui.Some(primary)
			n.Style.Background = ui.Some(rgb(49, 57, 78))
		}
		return n
	}
	left := pad(col("pulse-sidebar", darkLabel("pulse-brand", "Pulse", 26, 700), label("pulse-tagline", "A clearer signal", 12, 400, secondary), section("pulse-workspace", "WORKSPACE", navItem("pulse-all", "All services", "all", teal), navItem("pulse-production", "Production", "production", rgb(125, 134, 236)), navItem("pulse-degraded", "Needs attention", "degraded", rgb(238, 162, 83))), spacer("pulse-side-space"), pill("pulse-simulation", "SAMPLE DATASET", teal, rgb(37, 62, 66)), label("pulse-offline", "Offline · no system monitoring", 10, 400, secondary)), 16)
	left = width(left, 196)
	left.Style.Background = ui.Some(rgb(27, 30, 40))
	left.Style.GrowY = ui.Some(float32(1))
	left.Style.Gap = ui.Some(float32(25))
	headings := row("pulse-heading", col("pulse-titles", label("pulse-kicker", "SYSTEM OVERVIEW", 10, 600, secondary), darkLabel("pulse-title", "Good signals. Clear decisions.", 23, 600)), spacer("pulse-title-space"), ui.Node{ID: "pulse-period", Kind: ui.SegmentedControl, SelectedID: m.period, Options: []ui.Option{{ID: "1h", Label: "1h"}, {ID: "24h", Label: "24h"}, {ID: "7d", Label: "7d"}}, Accessibility: ui.Accessibility{Name: "Chart time range"}, Change: func(e ui.ChangeEvent) { _ = m.SetPeriod(e.SelectedID) }})
	count := len(m.source.rows)
	healthy, total, requests := 0, 0., 0
	for _, r := range m.source.rows {
		if r.Status == "Healthy" {
			healthy++
		}
		total += r.Latency
		requests += r.Requests
	}
	average := 0.
	if count > 0 {
		average = total / float64(count)
	}
	stats := row("pulse-metrics")
	for _, metric := range []struct {
		id, title, value, sub string
		c                     color.NRGBA
	}{{"latency", "AVG P95 LATENCY", fmt.Sprintf("%.1f ms", average), "Across visible services", teal}, {"health", "HEALTHY", fmt.Sprintf("%d", healthy), fmt.Sprintf("of %d filtered services", count), rgb(137, 220, 170)}, {"traffic", "REQUEST VOLUME", thousands(requests), "Synthetic requests / min", rgb(163, 145, 235)}} {
		card := pad(col("metric-"+metric.id, label("metric-label-"+metric.id, metric.title, 10, 600, secondary), label("metric-value-"+metric.id, metric.value, 30, 400, metric.c), label("metric-sub-"+metric.id, metric.sub, 11, 400, secondary)), 16)
		card.Style.Gap = ui.Some(float32(7))
		card.Style.Background = ui.Some(panel)
		card.Style.CornerRadius = ui.Some(float32(9))
		stats.Children = append(stats.Children, card)
	}
	chart := pad(col("pulse-chart", row("pulse-chart-heading", darkLabel("pulse-chart-title", "Latency over time", 14, 600), spacer("pulse-chart-space"), pill("pulse-chart-series", "P95", teal, rgb(34, 59, 64)), pill("pulse-chart-reference", "Baseline", rgb(169, 151, 239), rgb(49, 43, 72))), height(ui.Node{ID: "pulse-chart-surface", Kind: ui.Extension, Grow: true, Custom: m.surface, Accessibility: ui.Accessibility{Name: "Synthetic latency chart", Help: "Click a sample to inspect its value."}}, 172), row("pulse-chart-axis", label("pulse-chart-start", "00:00", 10, 400, secondary), spacer("pulse-axis-space"), label("pulse-chart-end", m.period+" window", 10, 400, secondary))), 14)
	chart.Style.Background = ui.Some(panel)
	chart.Style.CornerRadius = ui.Some(float32(9))
	chart.Style.Gap = ui.Some(float32(8))
	search := ui.Node{ID: "pulse-search", Kind: ui.TextField, Text: m.query, Placeholder: "Filter services", Accessibility: ui.Accessibility{Name: "Filter services"}, Change: func(e ui.ChangeEvent) { _ = m.SetSearch(e.Text) }, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 170}), MaxSize: ui.Some(ui.Size{Width: 220}), GrowX: ui.Some(float32(0))}}
	tableHead := row("pulse-table-heading", darkLabel("pulse-table-title", "Services", 15, 600), label("pulse-table-count", fmt.Sprintf("%d", count), 12, 400, secondary), spacer("pulse-table-space"), search, ui.Node{ID: "pulse-sort", Kind: ui.Selector, SelectedID: m.order, Options: []ui.Option{{ID: "name", Label: "Name"}, {ID: "latency", Label: "Latency"}}, Accessibility: ui.Accessibility{Name: "Sort services"}, Change: func(e ui.ChangeEvent) {
		if e.SelectedID == "name" || e.SelectedID == "latency" {
			m.order = e.SelectedID
			_ = m.replace()
		}
	}})
	grid := ui.Node{ID: "pulse-grid", Kind: ui.DataGrid, Grid: m.grid, Grow: true, Accessibility: ui.Accessibility{Name: "Virtualized sample services"}, GridChange: func(e ui.GridEvent) {
		if e.Reason == ui.GridSelection && e.State.Active.RowKey != "" {
			m.selected = e.State.Active.RowKey
			m.surface.Revision++
			m.changed()
		}
	}, Style: ui.Style{MinSize: ui.Some(ui.Size{Height: 170}), GrowY: ui.Some(float32(1)), Background: ui.Some(panel), Foreground: ui.Some(primary)}}
	main := pad(col("pulse-main", headings, stats, chart, tableHead, grid, row("pulse-footer", label("pulse-source-count", "1,024 indexed fixture records", 11, 400, secondary), spacer("pulse-footer-space"), label("pulse-sample-number", fmt.Sprintf("Sample %03d · ⌘R advances", m.sample+1), 11, 400, secondary))), 20)
	main.Style.Background = ui.Some(bg)
	main.Style.GrowY = ui.Some(float32(1))
	main.Style.Gap = ui.Some(float32(16))
	service := m.service()
	statusColor := rgb(137, 220, 170)
	if service.Status == "Degraded" {
		statusColor = rgb(242, 174, 103)
	}
	detail := section("pulse-service", "SELECTED SERVICE", label("pulse-service-name", service.Name, 18, 600, primary), pill("pulse-service-health", service.Status, statusColor, rgb(39, 51, 49)), label("pulse-service-environment", service.Environment, 12, 400, secondary))
	selectedSample := section("pulse-sample", "INSPECTED SAMPLE", label("pulse-sample-value", fmt.Sprintf("%.1f ms", m.latency(m.focus, false)), 31, 400, teal), label("pulse-sample-index", fmt.Sprintf("Point %d of 64", m.focus+1), 12, 400, secondary), label("pulse-sample-tip", "Click the chart to inspect.", 11, 400, secondary))
	note := pad(col("pulse-note", label("pulse-note-title", "Built for the big picture", 12, 600, primary), label("pulse-note-body", "Only visible rows are projected.", 11, 400, secondary), label("pulse-note-body2", "Stable keys keep selection coherent.", 11, 400, secondary)), 12)
	note.Style.Background = ui.Some(rgb(36, 40, 54))
	note.Style.CornerRadius = ui.Some(float32(8))
	inspector := pad(col("pulse-inspector", darkLabel("pulse-details-title", "Service details", 15, 600), detail, selectedSample, note, spacer("pulse-inspector-space"), button("pulse-advance", "Next sample", func() { _ = m.NextSample() }), label("pulse-replay-note", "Deterministic, in-memory fixture.", 10, 400, secondary)), 18)
	inspector = width(inspector, 236)
	inspector.Style.Background = ui.Some(panel)
	inspector.Style.GrowY = ui.Some(float32(1))
	inspector.Style.Gap = ui.Some(float32(25))
	return ui.Node{ID: "pulse-root", Kind: ui.Row, Grow: true, Children: []ui.Node{left, main, inspector}, Style: ui.Style{Gap: ui.Some(float32(1)), Background: ui.Some(border)}}

}
