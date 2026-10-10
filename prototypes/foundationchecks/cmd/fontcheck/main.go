// fontcheck isolates shaping memory and glyph coverage with the same corpus.
package main

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/fonts"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/shaped/shapers/shapedgt"
	"cogentcore.org/core/text/text"
	"encoding/json"
	"flag"
	"os"
	"runtime"
	"runtime/pprof"
)

var corpus = []string{"Latin: office AV", "Arabic: العربية 123", "Hebrew: שלום 456", "Indic: नमस्ते", "CJK: 日本語 中文", "Combining: é å", "Emoji: 👩‍💻 👨‍👩‍👧‍👦 🇺🇳"}

type memory struct {
	HeapAlloc, HeapInuse, TotalAlloc uint64
	Goroutines                       int
}

func mem() memory {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return memory{m.HeapAlloc, m.HeapInuse, m.TotalAlloc, runtime.NumGoroutine()}
}
func main() {
	flag.Bool("nogui", true, "use toolkit offscreen initialization")
	systemOnly := flag.Bool("system-only", false, "diagnostic control without embedded fonts; changes font policy, not an accepted fix")
	profile := flag.String("profile", "", "write live heap profile")
	cache := flag.String("font-cache", "", "owned writable font index directory; required")
	repeat := flag.Int("repeat", 1, "independent shaper lifetimes")
	flag.Parse()
	if *repeat < 1 {
		panic("repeat must be positive")
	}
	if *cache == "" {
		panic("font-cache is required to verify system font discovery")
	}
	if *systemOnly {
		fonts.Embedded = nil
	}
	var samples []map[string]any
	for cycle := 0; cycle < *repeat; cycle++ {
		before := mem()
		sh := shapedgt.NewShaperWithFonts(fonts.Embedded).(*shapedgt.Shaper)
		if err := sh.FontMap().UseSystemFonts(*cache); err != nil {
			panic(err)
		}
		initialized := mem()
		uc := units.Context{}
		uc.Defaults()
		ts := text.NewStyle()
		ts.ToDots(&uc)
		var coverage []map[string]any
		for _, s := range corpus {
			tx := rich.NewText(rich.NewStyle(), []rune(s))
			runs := sh.Shape(tx, ts)
			zero := 0
			glyphs := 0
			var locations []string
			for _, r := range runs {
				rr := r.(*shapedgt.Run)
				locations = append(locations, sh.FontMap().FontLocation(rr.Face.Font).File)
				for _, g := range rr.Glyphs {
					glyphs++
					if g.GlyphID == 0 {
						zero++
					}
				}
			}
			coverage = append(coverage, map[string]any{"text": s, "glyph_count": glyphs, "zero_glyph_ids": zero, "font_locations": locations})
		}
		// Capture the geometry of a mixed-direction line without assuming it supplies
		// a visual keyboard-navigation policy.
		mixed := "abc אבג 123"
		sty := rich.NewStyle()
		ln := sh.WrapLines(rich.NewText(sty, []rune(mixed)), sty, ts, math32.Vec2(1000, 100))
		var boxes [][4]float32
		for i := range []rune(mixed) {
			b := ln.RuneBounds(i)
			boxes = append(boxes, [4]float32{b.Min.X, b.Min.Y, b.Max.X, b.Max.Y})
		}
		live := mem()
		if cycle == *repeat-1 && *profile != "" {
			f, err := os.Create(*profile)
			if err != nil {
				panic(err)
			}
			if err = pprof.WriteHeapProfile(f); err != nil {
				panic(err)
			}
			f.Close()
		}
		runtime.KeepAlive(sh)
		ln = nil
		sh = nil
		released := mem()
		samples = append(samples, map[string]any{"cycle": cycle + 1, "before": before, "initialized": initialized, "live": live, "after_release": released, "coverage": coverage, "mixed_text": mixed, "mixed_rune_bounds": boxes})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(map[string]any{"system_only": *systemOnly, "scope": "isolated shaper diagnostics; glyph IDs do not prove visual correctness; not full app memory", "samples": samples}); err != nil {
		panic(err)
	}
}
