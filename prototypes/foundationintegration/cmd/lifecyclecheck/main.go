//go:build darwin && cgo

package main

import (
	"cogentcore.org/core/system"
	_ "cogentcore.org/core/system/driver"
	"cogentcore.org/core/system/driver/desktop"
	"encoding/json"
	"errors"
	"example.com/vellum/foundationcontracts/mailbox"
	"example.com/vellum/foundationintegration/inputbridge"
	"example.com/vellum/foundationintegration/macbridge"
	"example.com/vellum/foundationintegration/semantics"
	"example.com/vellum/foundationintegration/transport"
	"flag"
	"fmt"
	"github.com/go-gl/glfw/v3.4/glfw"
	"image"
	"os"
	"runtime"
	"time"
)

type lifetime struct {
	w      *desktop.Window
	in     *inputbridge.Input
	ax     *macbridge.Bridge
	q      *transport.Pump[inputbridge.Update]
	closed *int
	focus  *int
}

func (l *lifetime) Focus(f bool) { *l.focus++; l.ax.Focus(f) }
func (l *lifetime) Close() {
	if l.w.Glw == nil || l.w.Draw == nil {
		panic("late detach")
	}
	l.in.Close()
	l.ax.Close()
	l.q.Close()
	*l.closed++
}
func measurement(c int) map[string]any {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return map[string]any{"cycle": c, "heap_alloc": m.HeapAlloc, "heap_inuse": m.HeapInuse, "resident": macbridge.ResidentBytes(), "goroutines": runtime.NumGoroutine()}
}
func main() {
	n := flag.Int("cycles", 1000, "real Cogent/Metal surface and native adapter creation/teardown cycles")
	flag.Parse()
	if *n < 1 {
		panic("positive cycles required")
	}
	a := desktop.TheApp
	attached, closed, focus, accepted, applied := 0, 0, 0, 0, 0
	var latest *lifetime
	a.NativeAttach = func(w *desktop.Window) (desktop.NativeAttachment, error) {
		if w.Glw.GetAttrib(glfw.Visible) != glfw.False || w.Draw == nil {
			panic("wrong attachment ordering")
		}
		q, e := transport.New[inputbridge.Update](mailbox.Limits{Slots: 32, PayloadBytes: 65536, MessageBytes: 8192}, func() {})
		if e != nil {
			panic(e)
		}
		l := &lifetime{w: w, q: q, closed: &closed, focus: &focus}
		initial := semantics.Snapshot{Focus: 1, Nodes: []semantics.Node{{ID: 1, Role: "window", Label: "Lifecycle"}}}
		l.ax = macbridge.Attach(w.Glw.GetCocoaView(), w.Glw.GetCocoaWindow(), initial, func(semantics.Action) {})
		var pending *transport.Ticket[inputbridge.Update]
		l.in = inputbridge.Attach(w.Glw.GetCocoaView(), func(u inputbridge.Update) bool { var e error; pending, e = q.Preflight(u); return e == nil }, func() {
			if e := pending.Commit(); e != nil {
				panic(e)
			}
			accepted++
		})
		if l.in == nil {
			l.ax.Close()
			return nil, errors.New("input attach")
		}
		attached++
		latest = l
		return l, nil
	}
	contracts := publicationContracts()
	contracts["native_geometry"] = geometryContracts()
	start := time.Now()
	samples := []map[string]any{measurement(0)}
	for cycle := 1; cycle <= *n; cycle++ {
		pool := poolBegin()
		sw, e := a.NewWindow(&system.NewWindowOptions{Title: "Vellum lifetime", Size: image.Pt(400, 250), StartHidden: true})
		if e != nil {
			panic(e)
		}
		w := sw.(*desktop.Window)
		w.SetOnDemand(true)
		latest.in.Update("A👩Z", 1, 1, -1, -1, true, 0, [4]float64{10, 10, 1, 23})
		for _, step := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9} {
			if latest.in.Probe(step) == "" {
				panic("probe failed")
			}
		}
		// Keep one accepted mark pending during detach to exercise shutdown drain.
		latest.in.Probe(1)
		w.Focused(w.Glw, true)
		w.Focused(w.Glw, false)
		w.Close()
		w.Close()
		if w.Glw != nil || w.Draw != nil || w.Compose != nil || len(a.Windows) != 0 {
			panic("window/surface retained")
		}
		if e := latest.q.Drain(32, func(u inputbridge.Update) {
			if _, e := inputbridge.Decode(u); e != nil {
				panic(e)
			}
			applied++
		}); e != nil {
			panic(e)
		}
		if latest.q.Stats().Queue.Outstanding != 0 {
			panic("accepted update lost")
		}
		latest = nil
		glfw.PollEvents()
		poolEnd(pool)
		if cycle%100 == 0 || cycle == *n {
			samples = append(samples, measurement(cycle))
		}
	}
	// Failed creation must close a returned partial attachment without registration.
	failClosed := 0
	a.NativeAttach = func(w *desktop.Window) (desktop.NativeAttachment, error) {
		return &failureLifetime{w, &failClosed}, errors.New("injected attachment failure")
	}
	before := len(a.Windows)
	pool := poolBegin()
	sw, e := a.NewWindow(&system.NewWindowOptions{StartHidden: true, Size: image.Pt(400, 250)})
	poolEnd(pool)
	if e == nil || sw != nil || len(a.Windows) != before || failClosed != 1 {
		panic("failed creation cleanup")
	}
	if accepted != *n*11 || applied != accepted || closed != *n || attached != *n || focus != *n*2 {
		panic(fmt.Sprint("count mismatch ", attached, closed, focus, accepted, applied))
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"publication_contracts": contracts, "cycles": *n, "attachments": attached, "detaches": closed, "focus_protocol_callbacks": focus, "accepted": accepted, "applied_after_close": applied, "failure_cleanup": failClosed, "seconds": time.Since(start).Seconds(), "samples": samples, "scope": "actual Cogent window, Metal surface/drawer and Cocoa/AX adapter lifetime; hidden windows, no submitted frames or font/atlas workload; focus protocol calls are not live OS focus/VoiceOver"})
	glfw.Terminate()
}

type failureLifetime struct {
	w *desktop.Window
	n *int
}

func (f *failureLifetime) Focus(bool) {}
func (f *failureLifetime) Close() {
	if f.w.Glw == nil || f.w.Draw == nil {
		panic("late failure cleanup")
	}
	*f.n++
}

func publicationContracts() map[string]any {
	pool := poolBegin()
	defer poolEnd(pool)
	view := bareViewNew()
	defer bareViewFree(view)
	commits := 0
	var last inputbridge.Update
	in := inputbridge.Attach(view, func(u inputbridge.Update) bool {
		last = u
		if _, e := inputbridge.Decode(u); e != nil {
			panic(e)
		}
		return true
	}, func() { commits++ })
	if in == nil {
		panic("bare NSView attachment")
	}
	defer in.Close()
	in.Update("A👩Z", 1, 1, -1, -1, false, 0, [4]float64{})
	if in.ProbeInactive() != "inactive_optional_super_queries_passed" {
		panic("optional superclass fallback")
	}
	in.Update("A👩Z\nאבג", 1, 1, -1, -1, true, 0, [4]float64{10, 20, 1, 23})
	if in.QueryProbe(-1, 1, [2]float64{}) != "" || in.QueryProbe(0, 1<<31, [2]float64{}) != "" {
		panic("invalid query range admitted")
	}
	var queries []map[string]any
	for _, request := range [][2]int{{0, 2147483647}, {2, 1}, {4, 3}, {20, 1}} {
		var observed struct {
			Substring      string `json:"substring"`
			SubstringRange [2]int `json:"substring_range"`
			RectRange      [2]int `json:"rect_range"`
			PointIndex     int    `json:"point_index"`
		}
		raw := in.QueryProbe(request[0], request[1], [2]float64{12, 24})
		if err := json.Unmarshal([]byte(raw), &observed); err != nil {
			panic(err)
		}
		queries = append(queries, map[string]any{"requested": request, "observed": json.RawMessage(raw)})
		if observed.RectRange != [2]int{1, 0} || observed.PointIndex != -1 {
			panic("advertised geometry capability changed unexpectedly")
		}
		if request[0] == 0 && (observed.Substring != "A👩Z\nאבג" || observed.SubstringRange != [2]int{0, 8}) {
			panic("clipped context query")
		}
		if request[0] == 2 && (observed.Substring != "👩" || observed.SubstringRange != [2]int{1, 2}) {
			panic("surrogate context query")
		}
		if request[0] == 4 && (observed.Substring != "\nאב" || observed.SubstringRange != [2]int{4, 3}) {
			panic("multiline context query")
		}
		if request[0] == 20 && observed.SubstringRange != [2]int{-1, 0} {
			panic("outside context query")
		}
	}

	in.Update("A👩Z", 1, 1, 1, 3, true, 0, [4]float64{})
	in.Probe(7) // externally published marked text can unmark without a cancellation snapshot
	if commits != 1 {
		panic("external unmark missing")
	}
	before := in.Probe(0)
	for _, mark := range [][2]int{{3, 1}, {-1, 0}, {1, -1}, {2, 3}, {1, 10}} {
		in.Update("different", 0, 0, mark[0], mark[1], true, 1, [4]float64{})
		// (2,3) is a valid ASCII range, but its selection lies outside it.
		if in.Probe(0) != before {
			panic("invalid publication was not atomic")
		}
	}
	in.Update("A👩Z", 2, 2, -1, -1, true, 1, [4]float64{})
	if in.Probe(0) != before {
		panic("surrogate interior accepted")
	}
	in.Update("A👩Z", 1, 1, 1, 3, true, 1, [4]float64{})
	before = in.Probe(0)
	in.Probe(5)
	if in.Probe(0) != before || commits != 1 {
		panic("cancel without restoration snapshot silently changed state")
	}
	in.Update("A👩Z", 3, 1, -1, -1, true, 1, [4]float64{10, 20, 1, 23})
	in.Probe(1)
	in.Probe(5)
	if last.Kind != "cancel" || last.Anchor != 3 || last.Focus != 1 {
		panic("backward cancellation direction changed")
	}
	before = in.Probe(0)
	in.Update("A👩Z", 1, 1, -1, -1, false, 2, [4]float64{999, 999, 1, 23})
	if in.Probe(0) != before {
		panic("stale text/geometry/activation applied")
	}
	in.Probe(1)
	in.Update("EXにZ", 3, 3, 2, 3, true, 4, [4]float64{10, 20, 1, 23})
	before = in.Probe(0)
	in.Probe(5)
	if in.Probe(0) != before {
		panic("obsolete baseline erased external edit")
	}
	in.Probe(7)
	return map[string]any{"native_queries": queries, "geometry_scope": "legacy insertion-only API remains supported; native_geometry contains the new range/point contract", "backward_cancel_direction": true, "stale_geometry_rejected": true, "obsolete_baseline_invalidated": true, "bare_view_optional_methods": true, "external_unmark": true, "atomic_invalid_publications": 6, "snapshotless_cancel_refused": true}
}
