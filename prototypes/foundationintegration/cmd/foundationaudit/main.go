// foundationaudit checks unresolved contracts without interpreting unavailable
// native platforms or assistive-technology observations as successful checks.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/fonts"
	"cogentcore.org/core/text/lines"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/shaped/shapers/shapedgt"
	"cogentcore.org/core/text/text"
	"example.com/vellum/foundationcontracts/mailbox"
	"example.com/vellum/foundationintegration/transport"
	"github.com/go-text/typesetting/fontscan"
)

type record struct{ N int }
type orderingResult struct {
	AppliedAtShortcut       int    `json:"applied_at_shortcut"`
	RequiredBeforeShortcut  int    `json:"required_before_shortcut"`
	AppliedFinal            int    `json:"applied_final"`
	FIFO                    bool   `json:"fifo"`
	OrderedWithLegacyKeys   bool   `json:"ordered_with_legacy_keys"`
	LegacyKeyBetweenBatches bool   `json:"legacy_key_ran_between_batches"`
	MaxAppliedPerWake       int    `json:"max_applied_per_wake"`
	Wakes                   uint64 `json:"wakes"`
}

// This uses the actual toolkit Source/Deque and the fixture's wake placement.
// A queued legacy shortcut must observe all earlier admitted native updates.
func checkOrdering() (orderingResult, error) {
	var source events.Source
	p, e := transport.New[record](mailbox.Limits{Slots: 64, PayloadBytes: 4096, MessageBytes: 128}, func() {
		ev := &events.CustomEvent{}
		ev.Typ = events.Custom
		ev.Init()
		ev.SetUnique()
		source.Deque.Send(ev)
	})
	if e != nil {
		return orderingResult{}, e
	}
	for i := 1; i <= 32; i++ {
		if e = p.Send(record{i}); e != nil {
			return orderingResult{}, e
		}
	}
	source.Key(events.KeyDown, 'z', key.CodeZ, key.Meta)
	r := orderingResult{AppliedAtShortcut: -1, RequiredBeforeShortcut: 32, FIFO: true}
	for turns := 0; turns < 100; turns++ {
		ev, ok := source.Deque.PollEvent()
		if !ok {
			break
		}
		switch ev.Type() {
		case events.Custom:
			before := r.AppliedFinal
			if e = p.Drain(16, func(v record) {
				if v.N != r.AppliedFinal+1 {
					r.FIFO = false
				}
				r.AppliedFinal++
			}); e != nil {
				return r, e
			}
			r.MaxAppliedPerWake = max(r.MaxAppliedPerWake, r.AppliedFinal-before)
		case events.KeyChord:
			r.AppliedAtShortcut = r.AppliedFinal
		}
	}
	r.Wakes = p.Stats().Wakes
	r.OrderedWithLegacyKeys = r.AppliedAtShortcut == r.RequiredBeforeShortcut
	r.LegacyKeyBetweenBatches = r.AppliedAtShortcut > 0 && r.AppliedAtShortcut < 32
	if p.Stats().Queue.Outstanding != 0 || r.AppliedAtShortcut < 0 {
		return r, errors.New("ordering probe did not complete")
	}
	return r, nil
}

func heap() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

func checkLongText() map[string]any {
	source := []byte(strings.Repeat("abc אבג 日本語 é 👩‍💻\n", 10000) + "last\n\n")
	start := time.Now()
	ls := lines.NewLinesExact().SetText(source)
	exact := bytes.Equal(source, ls.Text())
	end := ls.EndPos()
	previous := ls.MoveBackwardGrapheme(end, 1)
	ls.DeleteText(previous, end)
	deletion := bytes.Equal(ls.Text(), source[:len(source)-1])
	ls.Close()
	// Repeat boundary queries on a long single line, without a layout or renderer.
	line := strings.Repeat("é👩‍💻", 14000)
	ls = lines.NewLinesExact().SetText([]byte(line))
	pos := ls.EndPos()
	walk := time.Now()
	for i := 0; i < 200; i++ {
		pos = ls.MoveBackwardGrapheme(pos, 1)
	}
	elapsed := time.Since(walk)
	want := len([]rune(line)) - 200*5/2
	movement := pos.Line == 0 && pos.Char == want
	ls.Close()
	return map[string]any{"scope": "exact buffer and logical grapheme primitives; no layout, native editing or visual bidi", "bytes": len(source), "lines": 10003, "exact_roundtrip": exact, "trailing_newline_deletion": deletion, "long_line_runes": len([]rune(line)), "backward_steps": 200, "observed_rune": pos.Char, "expected_rune": want, "grapheme_movement": movement, "boundary_walk_ms": float64(elapsed.Microseconds()) / 1000, "total_ms": float64(time.Since(start).Microseconds()) / 1000}
}

func checkBidi(cache string) (map[string]any, error) {
	body := core.NewBody("offscreen foundation audit")
	tf := core.NewTextField(body).SetText("abc אבג 123")
	if !tf.SetEditingSelection(0, 0) {
		return nil, errors.New("initial selection refused")
	}
	var cursor []int
	for i := 0; i < 7; i++ {
		ev := events.NewKey(events.KeyChord, 0, key.CodeRightArrow, 0)
		ev.Init()
		tf.HandleEvent(ev)
		if !ev.IsHandled() {
			return nil, errors.New("right-arrow handler not reached")
		}
		cursor = append(cursor, tf.EditingState().Cursor)
	}
	sh := shapedgt.NewShaperWithFonts(fonts.Embedded).(*shapedgt.Shaper)
	if e := sh.FontMap().UseSystemFonts(cache); e != nil {
		return nil, e
	}
	uc := units.Context{}
	uc.Defaults()
	ts := text.NewStyle()
	ts.ToDots(&uc)
	sty := rich.NewStyle()
	ln := sh.WrapLines(rich.NewText(sty, []rune("abc אבג 123")), sty, ts, math32.Vec2(1000, 100))
	var boxes [][4]float32
	for i := range []rune("abc אבג 123") {
		b := ln.RuneBounds(i)
		boxes = append(boxes, [4]float32{b.Min.X, b.Min.Y, b.Max.X, b.Max.Y})
	}
	logical := true
	for i, v := range cursor {
		if v != i+1 {
			logical = false
		}
	}
	descending := boxes[4][0] > boxes[5][0] && boxes[5][0] > boxes[6][0]
	runtime.KeepAlive(sh)
	return map[string]any{"scope": "actual field right-arrow handler plus shaped rune boxes; no rendered caret or native reference editor", "text": "abc אבג 123", "right_arrow_rune_positions": cursor, "rune_boxes": boxes, "logical_navigation": logical, "rtl_box_order_descends": descending, "visual_navigation_established": false, "reason": "the handler advances logical graphemes; shaped boxes are not a bidi caret/selection policy"}, nil
}

func checkFontScale(cache string) (map[string]any, error) {
	before := heap()
	fm := fontscan.NewFontMap(nil)
	if e := fm.UseSystemFonts(cache); e != nil {
		return nil, e
	}
	fps, e := fontscan.SystemFonts(nil, cache)
	if e != nil {
		return nil, e
	}
	families := map[string]bool{}
	for _, f := range fps {
		families[f.Family] = true
	}
	names := make([]string, 0, len(families))
	for n := range families {
		names = append(names, n)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, errors.New("font scale probe found no system families")
	}
	if len(names) > 64 {
		names = names[:64]
	}
	resolved := map[fontscan.Location]bool{}
	var samples []map[string]any
	for i, n := range names {
		fm.SetQuery(fontscan.Query{Families: []string{n}})
		face := fm.ResolveFace('A')
		if face == nil {
			return nil, fmt.Errorf("no fallback for family %q", n)
		}
		resolved[fm.FontLocation(face.Font)] = true
		if (i+1)%8 == 0 || i == len(names)-1 {
			count, size, fallbacks := fm.SnapshotStats()
			samples = append(samples, map[string]any{"queries": i + 1, "unique_locations": len(resolved), "heap_alloc": heap(), "snapshots": count, "snapshot_bytes": size, "snapshot_eager_fallbacks": fallbacks})
		}
	}
	runtime.KeepAlive(fm)
	return map[string]any{"scope": "one font map querying up to 64 sorted system families for A; wider than the seven-script corpus; no atlas or rendering", "heap_before": before, "families": names, "samples": samples, "global_face_budget_implemented": false}, nil
}

func run() (bool, error) {
	flag.Bool("nogui", true, "use the offscreen toolkit driver")
	cache := flag.String("font-cache", "", "owned writable font cache directory")
	flag.Parse()
	if *cache == "" {
		return false, errors.New("font-cache required")
	}
	ordering, e := checkOrdering()
	if e != nil {
		return false, e
	}
	long := checkLongText()
	bidi, e := checkBidi(*cache)
	if e != nil {
		return false, e
	}
	scale, e := checkFontScale(*cache)
	if e != nil {
		return false, e
	}
	scale["heap_after_map_release"] = heap()
	runtime.Gosched()
	scale["heap_after_second_gc"] = heap()
	count, size, fallbacks := fontscan.NewFontMap(nil).SnapshotStats()
	scale["snapshots_after_gc"] = map[string]any{"count": count, "bytes": size, "fallbacks": fallbacks}
	result := map[string]any{"schema": 1, "ordering": ordering, "long_text": long, "bidi": bidi, "font_scale": scale, "G1_passed": false, "unavailable_checks": []string{"full live IME candidate panel and screen-reader workflow", "Windows/Linux native runtimes", "trusted release on a clean machine"}}
	if e = json.NewEncoder(os.Stdout).Encode(result); e != nil {
		return false, e
	}
	if !ordering.OrderedWithLegacyKeys || !ordering.FIFO || ordering.AppliedFinal != 32 || !long["exact_roundtrip"].(bool) || !long["trailing_newline_deletion"].(bool) || !long["grapheme_movement"].(bool) {
		return false, nil
	}
	return true, nil
}
func main() {
	passed, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if !passed {
		fmt.Fprintln(os.Stderr, "foundation acceptance checks failed; see JSON observations")
		os.Exit(1)
	}
}
