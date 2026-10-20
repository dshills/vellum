package main

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/tree"
	"github.com/dshills/vellum/canvas"
)

func testLayerLayout(m *model, scale float64) layerLayout {
	g := layerLayout{clip: canvas.Bounds{Min: canvas.Point{X: 100 * scale, Y: 10 * scale}, Max: canvas.Point{X: 400 * scale, Y: 1000 * scale}}, scale: scale}
	for i, r := range m.layerRows() {
		y := float64(10+i*40) * scale
		g.rows = append(g.rows, layerBox{r.id, canvas.Bounds{Min: canvas.Point{X: 110 * scale, Y: y}, Max: canvas.Point{X: 390 * scale, Y: y + 40*scale}}})
	}
	return g
}
func layerPoint(g layerLayout, id string, front bool) canvas.Point {
	for _, r := range g.rows {
		if r.id == id {
			fraction := .75
			if front {
				fraction = .25
			}
			return canvas.Point{X: (r.bounds.Min.X + r.bounds.Max.X) / 2, Y: r.bounds.Min.Y + (r.bounds.Max.Y-r.bounds.Min.Y)*fraction}
		}
	}
	panic("fixture row missing")
}
func TestLayerInsertionBoundariesAndStableOrder(t *testing.T) {
	before := []canvas.Rect{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"}}
	groups := map[string]string{"a": "g1", "c": "g1", "d": "g2", "e": "g2"}
	for _, tc := range []struct {
		ids    []string
		target string
		front  bool
		want   []string
	}{
		{[]string{"a", "c"}, "b", true, []string{"b", "a", "c", "d", "e"}},
		{[]string{"a", "c"}, "b", false, []string{"a", "c", "b", "d", "e"}},
		{[]string{"b"}, "g1", true, []string{"a", "c", "b", "d", "e"}},
		{[]string{"b"}, "g1", false, []string{"b", "a", "c", "d", "e"}},
		{[]string{"a", "c"}, "g2", true, []string{"b", "d", "e", "a", "c"}},
		{[]string{"a", "c"}, "g2", false, []string{"b", "a", "c", "d", "e"}},
		{[]string{"b"}, "a", true, []string{"a", "b", "c", "d", "e"}},
	} {
		after, err := reorderedLayers(before, groups, tc.ids, tc.target, tc.front)
		if err != nil || !slices.Equal(objectIDs(after), tc.want) {
			t.Fatal("insertion", tc, objectIDs(after), err)
		}
	}
	for _, tc := range []struct {
		ids    []string
		target string
	}{{nil, "a"}, {[]string{"missing"}, "a"}, {[]string{"a"}, ""}, {[]string{"a"}, "missing"}, {[]string{"a", "c"}, "g1"}, {[]string{"a"}, "a"}} {
		if _, err := reorderedLayers(before, groups, tc.ids, tc.target, true); err == nil {
			t.Fatal("invalid insertion accepted", tc)
		}
	}
	if !slices.Equal(objectIDs(before), []string{"a", "b", "c", "d", "e"}) || groups["a"] != "g1" {
		t.Fatal("input mutated")
	}
}
func TestLayerDragGroupCommitHistoryExactlyOnceAndScale(t *testing.T) {
	for _, scale := range []float64{1, 2} {
		m := groupedModel(t)
		m.scene.View.Zoom = 2.75
		m.scene.View.Center = canvas.Point{X: 73, Y: -82}
		before, groups, view, next, ng := slices.Clone(m.scene.Objects), maps.Clone(m.groups), m.scene.View, m.next, m.nextGroup
		count := m.journal.Snapshot().UndoCount
		_ = m.journal.MarkSaved()
		g := testLayerLayout(m, scale)
		start := layerPoint(g, "group-1", true)
		s := m.beginLayerDrag("group-1", start, g)
		if s == nil || len(s.ids) != 2 {
			t.Fatal("group not captured")
		}
		p := layerPoint(g, "connector-2", true)
		if !m.updateLayerDrag(s, p, g) || !s.valid || !s.dragging || s.target != "connector-2" || !s.front || m.journal.Snapshot().UndoCount != count || !slices.Equal(m.scene.Objects, before) {
			t.Fatal("preview committed or wrong target")
		}
		if err := m.finishLayerDrag(s, p, g); err != nil {
			t.Fatal(err)
		}
		after := slices.Clone(m.scene.Objects)
		if !slices.Equal(objectIDs(after), []string{"connector-2", "block-1", "note-3"}) || m.layerDrag != nil || m.journal.Snapshot().UndoCount != count+1 || !m.journal.Snapshot().Dirty || !maps.Equal(groups, m.groups) || m.scene.View != view || m.next != next || m.nextGroup != ng {
			t.Fatal("commit changed unrelated state")
		}
		_ = m.finishLayerDrag(s, p, g)
		if m.journal.Snapshot().UndoCount != count+1 {
			t.Fatal("duplicate release committed")
		}
		if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) || !maps.Equal(m.groups, groups) || m.journal.Snapshot().Dirty {
			t.Fatal("drag Undo", err)
		}
		if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
			t.Fatal("drag Redo", err)
		}
	}
}
func TestLayerDragUnselectedSourceSelectionAndNoop(t *testing.T) {
	m := groupedModel(t)
	g := testLayerLayout(m, 1)
	prior := m.selectedIDs()
	primary := m.scene.Selected
	s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	if s == nil || !slices.Equal(s.ids, []string{"connector-2"}) || !slices.Equal(m.selectedIDs(), prior) {
		t.Fatal("press selected unselected source early")
	}
	count := m.journal.Snapshot().UndoCount
	_ = m.finishLayerDrag(s, layerPoint(g, "group-1", true), g)
	if m.scene.Selected != "connector-2" || len(m.selectedIDs()) != 1 || m.scene.Objects[2].ID != "connector-2" || m.journal.Snapshot().UndoCount != count+1 {
		t.Fatal("unselected row drop")
	}
	_ = m.journal.Undo()
	if !slices.Equal(m.selectedIDs(), prior) || m.scene.Selected != primary {
		t.Fatal("Undo lost original selection")
	}
	// A component already immediately above the target remains a true no-op.
	g = testLayerLayout(m, 1)
	s = m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	_ = m.finishLayerDrag(s, layerPoint(g, "block-1", true), g)
	if m.layerDrag != nil || m.journal.Snapshot().UndoCount != count || m.journal.Snapshot().RedoCount != 1 || !slices.Equal(m.selectedIDs(), prior) {
		t.Fatal("no-op branched history or selected source")
	}
}
func TestLayerDragClicksThresholdCancelOutsideAndSelf(t *testing.T) {
	m := editModel(t)
	m.layerAdd = true
	_ = m.selectLayer("connector-2")
	g := testLayerLayout(m, 2)
	p := layerPoint(g, "connector-2", true)
	s := m.beginLayerDrag("connector-2", p, g)
	if s == nil {
		t.Fatal("begin")
	}
	count := m.journal.Snapshot().UndoCount
	p.X += 5 * 2
	_ = m.finishLayerDrag(s, p, g)
	if !slices.Equal(m.selectedIDs(), []string{"note-3"}) || m.journal.Snapshot().UndoCount != count {
		t.Fatal("short click must toggle once")
	}
	for _, mode := range []string{"cancel", "outside", "self"} {
		g = testLayerLayout(m, 1)
		before, ids := slices.Clone(m.scene.Objects), m.selectedIDs()
		s = m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		p = layerPoint(g, "block-1", true)
		_ = m.updateLayerDrag(s, p, g)
		switch mode {
		case "cancel":
			m.cancelLayerDrag(s)
			_ = m.finishLayerDrag(s, p, g)
		case "outside":
			_ = m.finishLayerDrag(s, canvas.Point{}, g)
		case "self":
			p = layerPoint(g, "connector-2", false)
			_ = m.finishLayerDrag(s, p, g)
		}
		if m.layerDrag != nil || !slices.Equal(m.scene.Objects, before) || !slices.Equal(m.selectedIDs(), ids) || m.journal.Snapshot().UndoCount != count {
			t.Fatal("canceled drag changed document", mode)
		}
	}
}
func TestLayerDragStaleGeometryAndBusyLifecycle(t *testing.T) {
	for _, mode := range []string{"layout", "scale", "clip", "objects", "groups", "selection", "close"} {
		m := groupedModel(t)
		_ = m.journal.MarkSaved()
		g := testLayerLayout(m, 1)
		s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		if s == nil {
			t.Fatal("begin")
		}
		if m.commands.Enabled("undo") == nil || m.commands.Enabled("bring-forward") == nil || m.selectLayer("block-1") == nil {
			t.Fatal("busy guard")
		}
		if _, err := drawingCodec(m).Capture(); err == nil {
			t.Fatal("save during layer drag")
		}
		switch mode {
		case "layout":
			g.rows = slices.Clone(g.rows)
			g.rows[0].bounds.Max.Y++
		case "scale":
			g.scale = 2
		case "clip":
			g.clip.Max.Y--
		case "objects":
			m.scene.Objects = m.scene.Objects[:2]
		case "groups":
			m.groups = nil
		case "selection":
			m.setSelection(nil, "")
		case "close":
			if !m.requestClose(nil) {
				t.Fatal("clean close refused")
			}
		}
		p := layerPoint(g, "group-1", true)
		_ = m.finishLayerDrag(s, p, g)
		if m.layerDrag != nil || m.journal.Snapshot().Dirty || m.journal.Snapshot().UndoCount != 4 {
			t.Fatal("stale or closed drag committed", mode)
		}
	}
}
func TestLayerLayoutClipHalvesAndSelectedSet(t *testing.T) {
	m := groupedModel(t)
	m.setSelection([]string{"block-1", "connector-2"}, "connector-2")
	g := testLayerLayout(m, 1)
	s := m.beginLayerDrag("group-1", layerPoint(g, "group-1", true), g)
	if s == nil || len(s.ids) != 3 || s.primary != "connector-2" {
		t.Fatal("selected set not captured")
	}
	m.cancelLayerDrag(s)
	p := layerPoint(g, "connector-2", true)
	id, front, _ := g.rowAt(p)
	if id != "connector-2" || !front {
		t.Fatal("upper half")
	}
	p = layerPoint(g, "connector-2", false)
	id, front, _ = g.rowAt(p)
	if id != "connector-2" || front {
		t.Fatal("lower half")
	}
	g.clip.Max.Y = p.Y - 1
	if id, _, _ := g.rowAt(p); id != "" {
		t.Fatal("unclipped row target")
	}
}

func TestLayerGroupMarkerPreviewUsesBackmostMember(t *testing.T) {
	for _, visible := range []bool{true, false} {
		m := groupedModel(t)
		g := testLayerLayout(m, 1)
		if !visible {
			g.clip.Max.Y = 150
		}
		s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		p := layerPoint(g, "group-1", false)
		_ = m.updateLayerDrag(s, p, g)
		if visible {
			if !s.valid || s.line != 170 {
				t.Fatal("group preview must mark last member boundary", s.line)
			}
		} else if !s.valid || s.line != g.clip.Max.Y {
			t.Fatal("offscreen group boundary must mark the viewport edge")
		}
		m.cancelLayerDrag(s)
	}
}

func shiftedLayerLayout(g layerLayout, delta float64) layerLayout {
	g.rows = slices.Clone(g.rows)
	for i := range g.rows {
		g.rows[i].bounds.Min.Y -= delta
		g.rows[i].bounds.Max.Y -= delta
	}
	return g
}

func TestLayerDragScrollRetargetsStationaryPointerAndHistory(t *testing.T) {
	for _, scale := range []float64{1, 2} {
		m := groupedModel(t)
		g := testLayerLayout(m, scale)
		s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		p := layerPoint(g, "block-1", false)
		if !m.updateLayerDrag(s, p, g) || !s.valid || s.target != "block-1" {
			t.Fatal("initial target")
		}
		count, before, view := m.journal.Snapshot().UndoCount, slices.Clone(m.scene.Objects), m.scene.View
		// The list moves under the held pointer. The source may leave the
		// viewport; identity and original document snapshots stay intact.
		next := shiftedLayerLayout(g, 40*scale)
		if !m.scrollLayerDrag(s, g, next, 40*scale) || s.pointer != p || s.target != "" || m.journal.Snapshot().UndoCount != count {
			t.Fatal("scroll committed or stale target", s.target)
		}
		if !m.scrollLayerDrag(s, next, g, -40*scale) || s.target != "block-1" || !s.valid {
			t.Fatal("reverse scroll")
		}
		if err := m.finishLayerDrag(s, p, g); err != nil {
			t.Fatal(err)
		}
		after := slices.Clone(m.scene.Objects)
		if m.journal.Snapshot().UndoCount != count+1 || m.scene.View != view || slices.Equal(before, after) {
			t.Fatal("one drop after scrolling")
		}
		if err := m.journal.Undo(); err != nil || !slices.Equal(m.scene.Objects, before) {
			t.Fatal("undo", err)
		}
		if err := m.journal.Redo(); err != nil || !slices.Equal(m.scene.Objects, after) {
			t.Fatal("redo", err)
		}
	}
}

func TestLayerDragScrollRejectsStructuralChangesAndCancels(t *testing.T) {
	for _, mutate := range []func(*layerLayout){
		func(g *layerLayout) { g.clip.Max.Y-- },
		func(g *layerLayout) { g.scale *= 2 },
		func(g *layerLayout) { g.rows[0].id = "removed" },
		func(g *layerLayout) { g.rows[0].bounds.Max.Y++ },
		func(g *layerLayout) { g.rows[0].bounds.Min.X++ },
		func(g *layerLayout) { g.rows = g.rows[1:] },
	} {
		m := groupedModel(t)
		g := testLayerLayout(m, 1)
		s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		m.updateLayerDrag(s, layerPoint(g, "block-1", true), g)
		after := shiftedLayerLayout(g, 20)
		mutate(&after)
		before, count := slices.Clone(m.scene.Objects), m.journal.Snapshot().UndoCount
		if m.scrollLayerDrag(s, g, after, 20) || m.layerDrag != nil {
			t.Fatal("invalid scroll retained")
		}
		_ = m.finishLayerDrag(s, layerPoint(g, "block-1", true), g)
		if !slices.Equal(before, m.scene.Objects) || m.journal.Snapshot().UndoCount != count {
			t.Fatal("canceled scroll edited")
		}
	}
}

func TestLayerClippedRowHitUsesFullMidpoint(t *testing.T) {
	m := groupedModel(t)
	g := testLayerLayout(m, 1)
	// Only the bottom of the first row is visible. It is still its lower
	// half; clipping must not create a new midpoint or change identity.
	g.clip.Min.Y = g.rows[0].bounds.Max.Y - 8
	p := canvas.Point{X: 200, Y: g.clip.Min.Y + 1}
	id, front, _ := g.rowAt(p)
	if id != g.rows[0].id || front {
		t.Fatal("clipped target midpoint")
	}
	if m.beginLayerDrag(id, p, g) == nil {
		t.Fatal("clipped source rejected")
	}
}

func TestLayerFirstNativeScrollAllowsUniformSubpixelAlignment(t *testing.T) {
	for _, rounding := range []float64{-.5, .5, 1.5} {
		m := groupedModel(t)
		g := testLayerLayout(m, 1)
		s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
		m.updateLayerDrag(s, layerPoint(g, "block-1", false), g)
		next := shiftedLayerLayout(g, 12.15+rounding)
		if got := m.scrollLayerDrag(s, g, next, 12.15); got != (math.Abs(rounding) <= 1) {
			t.Fatal("initial layout rounding", rounding, got)
		}
	}
}

func TestLayerFractionalScrollAtMaximumRowsAndHighScale(t *testing.T) {
	m, err := newModel()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.palette.Close)
	m.groups = make(map[string]string)
	for i := 0; i < 512; i++ {
		id := fmt.Sprintf("block-%d", i+1)
		m.scene.Objects = append(m.scene.Objects, canvas.Rect{ID: id, Width: 80, Height: 50})
		m.groups[id] = fmt.Sprintf("group-%d", i/2+1)
	}
	g := testLayerLayout(m, 4)
	g.clip.Max.Y = 800
	s := m.beginLayerDrag("group-256", layerPoint(g, "group-256", true), g)
	if s == nil {
		t.Fatal("begin")
	}
	m.updateLayerDrag(s, canvas.Point{X: 600, Y: 600}, g)
	for _, delta := range []float64{0.33333334, 0.7, -0.1875} {
		next := g
		next.rows = slices.Clone(g.rows)
		for i := range next.rows {
			next.rows[i].bounds.Min.Y = float64(float32(float32(g.rows[i].bounds.Min.Y) - float32(delta)))
			next.rows[i].bounds.Max.Y = float64(float32(float32(g.rows[i].bounds.Max.Y) - float32(delta)))
		}
		if !m.scrollLayerDrag(s, g, next, delta) || m.layerDrag != s {
			t.Fatal("valid fractional native scroll canceled", delta)
		}
		g = next
	}
	m.cancelLayerDrag(s)
}

func TestLayerEdgeBandAndScrollLimits(t *testing.T) {
	g := layerLayout{clip: canvas.Bounds{Min: canvas.Point{X: 100, Y: 10}, Max: canvas.Point{X: 400, Y: 110}}, scale: 2}
	s := &layerDrag{layout: g, dragging: true}
	for _, tc := range []struct {
		p    canvas.Point
		want float64
	}{
		{canvas.Point{X: 200, Y: 11}, -24}, {canvas.Point{X: 200, Y: 109}, 24},
		{canvas.Point{X: 200, Y: 60}, 0}, {canvas.Point{X: 99, Y: 109}, 0}, {canvas.Point{X: 200, Y: 111}, 0},
	} {
		s.pointer = tc.p
		if got := layerEdgeDelta(s); got != tc.want {
			t.Fatal(tc, got)
		}
	}
	s.dragging = false
	if layerEdgeDelta(s) != 0 {
		t.Fatal("pre-threshold auto scroll")
	}
	fr := tree.New[core.Frame]()
	fr.HasScroll[math32.Y] = true
	fr.Geom.Size.Internal.Y = 1000
	fr.Geom.Size.Alloc.Content.Y = 100
	sb := tree.New[core.Slider]()
	fr.Scrolls[math32.Y] = sb
	for _, tc := range []struct {
		current float32
		delta   float64
		want    float32
		changed bool
	}{
		{0, -10, 0, false}, {0, 10, 10, true}, {895, 20, 900, true}, {900, 20, 900, false}, {100, 1e9, 200, true}, {100, math.NaN(), 0, false},
	} {
		sb.Value = tc.current
		got, changed := layerScrollDestination(fr, tc.delta)
		if got != tc.want || changed != tc.changed {
			t.Fatal(tc, got, changed)
		}
	}
}

func TestLayerPulseRetiresQueuedWorkOnCancelAndDestroy(t *testing.T) {
	for _, destroy := range []bool{false, true} {
		f := fileModel(t, false)
		f.place(t, "block", canvas.Point{})
		f.place(t, "note", canvas.Point{X: 200})
		f.owner(t, func() {
			g := testLayerLayout(f.m, 1)
			s := f.m.beginLayerDrag("block-1", layerPoint(g, "block-1", true), g)
			w := tree.New[layerMount]()
			w.editor = &layerEditor{a: f.a, ref: f.ref, model: f.m}
			w.active = s
			var err error
			w.ticket, err = f.a.ReserveWindowCompletion(f.ref, nil)
			if err != nil {
				t.Fatal(err)
			}
			pulse, err := f.a.ReserveWindowCompletion(f.ref, nil)
			if err != nil {
				t.Fatal(err)
			}
			w.pulse = &layerPulse{ticket: pulse}
			if err = pulse.Publish(func() { t.Error("retired autoscroll ran") }); err != nil {
				t.Fatal(err)
			}
			w.pulse.timer = time.AfterFunc(time.Hour, func() { t.Error("retired timer fired") })
			if destroy {
				w.Destroy()
			} else {
				f.m.cancelLayerDrag(s)
				w.endCapture()
			}
			if w.pulse != nil || w.active != nil {
				t.Fatal("pulse references retained")
			}
		})
		if err := f.a.Drain(); err != nil {
			t.Fatal(err)
		}
		f.owner(t, func() {
			if f.m.layerDrag != nil {
				t.Fatal("domain capture retained")
			}
		})
	}
}

func TestLayerPulseSceneHandlerDoesNotAccumulateOrRetainMounts(t *testing.T) {
	scene := tree.New[core.Scene]()
	before := len(scene.Listeners.First[events.Custom])
	old := tree.New[layerMount]()
	old.bindPulse(scene)
	current := tree.New[layerMount]()
	current.bindPulse(scene)
	if got := len(scene.Listeners.First[events.Custom]); got != before+1 {
		t.Fatal("handlers accumulated", got, before)
	}
	old.Destroy()
	if scene.Property(layerPulseOwner) != current {
		t.Fatal("old destroy detached current mount")
	}
	current.Destroy()
	if scene.Property(layerPulseOwner) != nil || current.pulseScene != nil {
		t.Fatal("destroy retained mount")
	}
}
func TestLayerMountDestroyTransfersCancellationWithoutOwnerReentry(t *testing.T) {
	f := fileModel(t, false)
	f.place(t, "block", canvas.Point{})
	f.place(t, "note", canvas.Point{X: 200})
	f.owner(t, func() {
		g := testLayerLayout(f.m, 1)
		s := f.m.beginLayerDrag("block-1", layerPoint(g, "block-1", true), g)
		if s == nil {
			t.Fatal("begin")
		}
		ticket, err := f.a.ReserveWindowCompletion(f.ref, nil)
		if err != nil {
			t.Fatal(err)
		}
		w := tree.New[layerMount]()
		w.editor = &layerEditor{a: f.a, ref: f.ref, model: f.m}
		w.active = s
		w.ticket = ticket
		w.Destroy() // This call already holds the non-reentrant app owner.
		if w.ticket != nil || w.active != nil || w.sprite != nil || w.owner != nil {
			t.Fatal("native references retained")
		}
	})
	if err := f.a.Drain(); err != nil {
		t.Fatal(err)
	}
	f.owner(t, func() {
		if f.m.layerDrag != nil {
			t.Fatal("destroy canceled completion before dispatch")
		}
	})
}

func TestManagedCanvasSelectionCancelsLayerDragImmediately(t *testing.T) {
	m := groupedModel(t)
	g := testLayerLayout(m, 1)
	s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	if s == nil {
		t.Fatal("begin")
	}
	count := m.journal.Snapshot().UndoCount
	m.scene.Selected = "block-1" // Managed canvas selects exact ID before its callback.
	m.managedSelection("block-1")
	if m.layerDrag != nil || m.scene.Selected != "block-1" || len(m.selectedIDs()) != 2 || m.journal.Snapshot().UndoCount != count {
		t.Fatal("AX selection left a busy layer gesture")
	}
	_ = m.finishLayerDrag(s, layerPoint(g, "group-1", true), g)
	if m.journal.Snapshot().UndoCount != count {
		t.Fatal("release after semantic cancellation committed")
	}
}

func TestLayerCapturePreservesPresentationButGuardsExecution(t *testing.T) {
	m := groupedModel(t)
	g := testLayerLayout(m, 1)
	before := make(map[string]bool)
	for _, c := range m.commands.Search("") {
		before[c.ID] = m.commandDisabled(c.ID)
	}
	s := m.beginLayerDrag("connector-2", layerPoint(g, "connector-2", true), g)
	if s == nil {
		t.Fatal("begin")
	}
	for id, want := range before {
		if m.commandDisabled(id) != want {
			t.Fatal("capture restyled command", id)
		}
	}
	if m.commands.Enabled("undo") == nil || m.selectLayer("block-1") == nil {
		t.Fatal("stable presentation bypassed semantic guard")
	}
	m.cancelLayerDrag(s)
	if m.commandDisabled("undo") {
		t.Fatal("availability did not recover")
	}
}
