// fontcheck isolates shaping memory and glyph coverage with the same corpus.
package main

import (
	"bytes"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/fonts"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/shaped/shapers/shapedgt"
	"cogentcore.org/core/text/text"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-text/typesetting/font"
	_ "golang.org/x/image/tiff"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
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
			var bitmaps []map[string]any
			var outlines []string
			for _, r := range runs {
				rr := r.(*shapedgt.Run)
				locations = append(locations, sh.FontMap().FontLocation(rr.Face.Font).File)
				for _, g := range rr.Glyphs {
					glyphs++
					if outline, ok := rr.Face.GlyphDataOutline(g.GlyphID); ok {
						data, e := json.Marshal(outline.Segments)
						if e != nil {
							panic(e)
						}
						outlines = append(outlines, fmt.Sprintf("%x", sha256.Sum256(data)))
					} else {
						outlines = append(outlines, "unavailable")
					}

					if len(bitmaps) < 16 && g.GlyphID != 0 && strings.HasPrefix(s, "Emoji:") {
						rr.Face.SetPpem(32, 32)
						if bitmap, ok := rr.Face.GlyphDataBitmap(g.GlyphID); ok {
							decoded, format, err := decodeBitmap(bitmap)
							if err != nil {
								panic(fmt.Errorf("bitmap format %v decode: %w", bitmap.Format, err))
							}
							pix := []byte{}
							for y := decoded.Bounds().Min.Y; y < decoded.Bounds().Max.Y; y++ {
								for x := decoded.Bounds().Min.X; x < decoded.Bounds().Max.X; x++ {
									r, g, b, a := decoded.At(x, y).RGBA()
									pix = append(pix, byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8))
								}
							}
							bitmaps = append(bitmaps, map[string]any{"glyph": g.GlyphID, "format": format, "encoded_bytes": len(bitmap.Data), "width": bitmap.Width, "height": bitmap.Height, "pixel_sha256": fmt.Sprintf("%x", sha256.Sum256(pix))})
						}
					}
					if g.GlyphID == 0 {
						zero++
					}
				}
			}
			coverage = append(coverage, map[string]any{"text": s, "glyph_count": glyphs, "zero_glyph_ids": zero, "font_locations": locations, "decoded_bitmaps": bitmaps, "outline_hashes": outlines})
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
		var snapshots map[string]any
		if stats, ok := any(sh.FontMap()).(interface{ SnapshotStats() (int, int64, uint64) }); ok {
			count, bytes, fallbacks := stats.SnapshotStats()
			snapshots = map[string]any{"count": count, "bytes": bytes, "eager_fallbacks": fallbacks}
		}
		if cycle == *repeat-1 && *profile != "" {
			f, err := os.Create(*profile)
			if err != nil {
				panic(err)
			}
			if err = pprof.WriteHeapProfile(f); err != nil {
				panic(err)
			}
			if err := f.Close(); err != nil {
				panic(err)
			}
		}
		runtime.KeepAlive(sh)
		ln = nil
		sh = nil
		released := mem()
		samples = append(samples, map[string]any{"snapshots": snapshots, "cycle": cycle + 1, "before": before, "initialized": initialized, "live": live, "after_release": released, "coverage": coverage, "mixed_text": mixed, "mixed_rune_bounds": boxes})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(map[string]any{"system_only": *systemOnly, "scope": "isolated shaper diagnostics; glyph IDs do not prove visual correctness; not full app memory", "samples": samples}); err != nil {
		panic(err)
	}
}

func decodeBitmap(bitmap font.GlyphBitmap) (image.Image, string, error) {
	if bitmap.Width <= 0 || bitmap.Height <= 0 || bitmap.Width > 4096 || bitmap.Height > 4096 {
		return nil, "", fmt.Errorf("invalid/oversized decoded bitmap %dx%d", bitmap.Width, bitmap.Height)
	}
	switch bitmap.Format {
	case font.PNG, font.JPG, font.TIFF:
		config, _, err := image.DecodeConfig(bytes.NewReader(bitmap.Data))
		if err != nil {
			return nil, "", err
		}
		if config.Width != bitmap.Width || config.Height != bitmap.Height {
			return nil, "", fmt.Errorf("bitmap dimensions disagree")
		}
		return image.Decode(bytes.NewReader(bitmap.Data))
	case font.BlackAndWhite, font.BlackAndWhiteByteAligned:
		out := image.NewGray(image.Rect(0, 0, bitmap.Width, bitmap.Height))
		strideBits := bitmap.Width
		if bitmap.Format == font.BlackAndWhiteByteAligned {
			strideBits = (bitmap.Width + 7) / 8 * 8
		}
		if len(bitmap.Data)*8 < strideBits*(bitmap.Height-1)+bitmap.Width {
			return nil, "", fmt.Errorf("short monochrome bitmap")
		}
		for y := 0; y < bitmap.Height; y++ {
			for x := 0; x < bitmap.Width; x++ {
				bit := y*strideBits + x
				v := byte(0)
				if bitmap.Data[bit/8]&(0x80>>uint(bit%8)) != 0 {
					v = 255
				}
				out.Pix[y*out.Stride+x] = v
			}
		}
		return out, "monochrome", nil
	default:
		return nil, "", fmt.Errorf("unsupported bitmap format %v", bitmap.Format)
	}
}
