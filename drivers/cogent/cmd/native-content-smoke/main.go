// Command native-content-smoke opens only synthetic content in an owned window.
// It inspects actual AppKit embedding/lifetime; it does not inject physical input.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
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
	ref, e := a.OpenWindow(app.Window{ID: "native-content-smoke", Title: "Vellum synthetic content check", Width: 760, Height: 600, View: func() ui.Node {
		n := ui.Node{ID: "root", Kind: ui.Column, Grow: true}
		if visible {
			v, e := p.View()
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
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		await := func(count int) error {
			var last string
			for ctx.Err() == nil {
				wire, e := cogent.InspectNativeContent(ctx, a, ref)
				last = wire
				var v struct {
					Hosts []struct {
						Class            string `json:"class"`
						Loaded, Attached bool
						Pages            int
					}
				}
				if e == nil && json.Unmarshal([]byte(wire), &v) == nil && len(v.Hosts) == count && (count == 0 || v.Hosts[0].Loaded && v.Hosts[0].Attached && v.Hosts[0].Pages == 1) {
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
		e := await(1)
		if e == nil {
			e = a.PostWindow(ref, func() { visible = false; a.Invalidate(ref.ID()) })
		}
		if e == nil {
			e = await(0)
		}
		if e == nil {
			e = a.PostWindow(ref, func() { visible = true; a.Invalidate(ref.ID()) })
		}
		if e == nil {
			e = await(1)
		}
		if e == nil {
			e = a.PostWindow(ref, func() { visible = false; a.Invalidate(ref.ID()) })
		}
		if e == nil {
			e = await(0)
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
