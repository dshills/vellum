// Command native-content-smoke opens only synthetic content in an owned window.
// It inspects actual AppKit embedding/lifetime; it does not inject physical input.
package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/drivers/cogent/internal/nativefixture"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func pdf() []byte {
	text := "BT /F1 18 Tf 30 300 Td (Vellum synthetic PDF) Tj ET\n"
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R] /Count 1 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 400] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>", "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>", fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(text), text)}
	var b strings.Builder
	b.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, o := range objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, o := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<< /Root 1 0 R /Size %d >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return []byte(b.String())
}
func run() error {
	width := flag.Int("width", 760, "bounded native viewport width (340 or 760)")
	flag.Parse()
	if *width != 340 && *width != 760 {
		return fmt.Errorf("width must be 340 or 760")
	}
	restore, e := nativefixture.IsolateGeometry()
	if e != nil {
		return e
	}
	defer restore()
	dir, e := os.MkdirTemp("", "vellum-native-content-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "synthetic.pdf")
	if e = os.WriteFile(path, pdf(), 0600); e != nil {
		return e
	}
	a, e := app.New(app.Config{})
	if e != nil {
		return e
	}
	defer a.Close()
	p := &ui.PDFViewer{ID: "pdf", Label: "Synthetic PDF", Path: path, ShowThumbnails: true}
	visible := true
	media := false
	wavPath := filepath.Join(dir, "silent.wav")
	data := make([]byte, 44+16000)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	copy(data[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(data[16:], 16)
	binary.LittleEndian.PutUint16(data[20:], 1)
	binary.LittleEndian.PutUint16(data[22:], 1)
	binary.LittleEndian.PutUint32(data[24:], 8000)
	binary.LittleEndian.PutUint32(data[28:], 16000)
	binary.LittleEndian.PutUint16(data[32:], 2)
	binary.LittleEndian.PutUint16(data[34:], 16)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], 16000)
	if e = os.WriteFile(wavPath, data, 0600); e != nil {
		return e
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	pixels.SetNRGBA(0, 0, color.NRGBA{R: 220, G: 50, B: 140, A: 255})
	pixels.SetNRGBA(1, 0, color.NRGBA{G: 160, B: 230, A: 255})
	poster, _ := canvas.NewTiledImage(pixels)
	player := ui.MediaPlayer{PosterImage: poster, ID: "media", Label: "Synthetic silent audio", Path: wavPath, InitialMuted: true}
	ref, e := a.OpenWindow(app.Window{ID: "native-content-smoke", Title: "Vellum synthetic content check", Width: float32(*width), Height: 600, View: func() ui.Node {
		n := ui.Node{ID: "root", Kind: ui.Column, Grow: true}
		if visible {
			var v ui.Node
			var e error
			if media {
				v, e = player.View()
			} else {
				v, e = p.View()
			}
			if e != nil {
				panic(e)
			}
			n.Children = append(n.Children, v)
		}
		return n
	}})
	if e != nil {
		return e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		await := func(count int, kind string) error {
			var last string
			for ctx.Err() == nil {
				wire, e := cogent.InspectNativeContent(ctx, a, ref)
				last = wire
				var v struct {
					Hosts []struct {
						Rate             float64
						Muted            bool
						Poster           bool
						Class            string `json:"class"`
						Loaded, Attached bool
						Pages            int
						SearchWidth      float64 `json:"search_width"`
						Compact          bool    `json:"compact_toolbar"`
						Opaque           bool    `json:"opaque_background"`
					}
				}
				if e == nil && json.Unmarshal([]byte(wire), &v) == nil && len(v.Hosts) == count && (count == 0 || v.Hosts[0].Loaded && v.Hosts[0].Attached && v.Hosts[0].Opaque && v.Hosts[0].Class == kind && (kind != "PDFView" || v.Hosts[0].Pages == 1 && v.Hosts[0].SearchWidth >= 120 && v.Hosts[0].Compact == (*width == 340)) && (kind != "AVPlayerView" || v.Hosts[0].Rate == 0 && v.Hosts[0].Muted && v.Hosts[0].Poster)) {
					fmt.Println(wire)
					return nil
				}
				select {
				case <-ctx.Done():
				case <-time.After(40 * time.Millisecond):
				}
			}
			return fmt.Errorf("native hosts=%d: %w; last=%s", count, ctx.Err(), last)
		}
		e := await(1, "PDFView")
		if e == nil {
			e = a.PostWindow(ref, func() { visible = false; a.Invalidate(ref.ID()) })
		}
		if e == nil {
			e = await(0, "")
		}
		if e == nil {
			e = a.PostWindow(ref, func() { visible = true; a.Invalidate(ref.ID()) })
		}
		if e == nil {
			e = await(1, "PDFView")
		}
		if e == nil {
			e = a.PostWindow(ref, func() { visible = false; a.Invalidate(ref.ID()) })
		}
		if e == nil {
			e = await(0, "")
		}
		if e == nil {
			e = a.PostWindow(ref, func() { media = true; visible = true; a.Invalidate(ref.ID()) })
		}
		if e == nil {
			e = await(1, "AVPlayerView")
		}
		if e == nil {
			e = a.PostWindow(ref, func() { visible = false; a.Invalidate(ref.ID()) })
		}
		if e == nil {
			e = await(0, "")
		}
		done <- e
		cancel()
	}()
	native := a.Run(ctx, cogent.Driver{Theme: ui.MacTheme(false)})
	e = <-done
	if errors.Is(native, context.Canceled) {
		native = nil
	}
	return errors.Join(e, native)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
