package cogent

import (
	"bytes"
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"image"
	"image/color"
	"image/png"
	"reflect"
	"testing"
)

func trayTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, G: 127, B: 17, A: 64})
	img.SetNRGBA(1, 0, color.NRGBA{R: 23, G: 24, B: 25, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestTrayPNGWorkerDecodeAndDetachedPixels(t *testing.T) {
	data := trayTestPNG(t)
	var desc nativeTray
	if err := trayPixels(&desc, data); err != nil {
		t.Fatal(err)
	}
	want := []byte{255, 127, 17, 64, 23, 24, 25, 255}
	if desc.Width != 2 || desc.Height != 1 || !bytes.Equal(desc.RGBA, want) {
		t.Fatal(desc)
	}
	for i := range data {
		data[i] = 0
	}
	if !bytes.Equal(desc.RGBA, want) {
		t.Fatal("pixels alias caller PNG")
	}
	corrupt := trayTestPNG(t)
	corrupt[len(corrupt)-1] ^= 1
	if err := (app.TrayOptions{Title: "Vellum", IconPNG: corrupt}).Validate(); err != nil {
		t.Fatal("fixture header invalid", err)
	}
	if err := trayPixels(&nativeTray{}, corrupt); err == nil {
		t.Fatal("corrupt full PNG accepted")
	}
}
func TestTrayOwnerCaptureAndPreflight(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	r := command.New()
	enabled := true
	executions := 0
	if err := r.Register(command.Command{ID: "show", Label: "Show workbench", Available: func() error {
		if !enabled {
			return errors.New("disabled")
		}
		return nil
	}, Execute: func(context.Context) error { executions++; return nil }}); err != nil {
		t.Fatal(err)
	}
	s := &platformServices{owner: a, commands: r}
	o := app.TrayOptions{Title: "Vellum", Menu: []app.TrayMenuItem{{Command: "show"}, {Separator: true}, {Command: "show"}}}
	desc, ids, err := s.captureTray(o)
	if err != nil || !reflect.DeepEqual(ids, []string{"show", "show"}) || len(desc.Items) != 3 || desc.Items[0].Index != 0 || !desc.Items[1].Separator || desc.Items[2].Index != 1 {
		t.Fatal(desc, ids, err)
	}
	a.Do(func() { enabled = false })
	desc, _, err = s.captureTray(o)
	if err != nil || !desc.Items[0].Disabled || executions != 0 {
		t.Fatal(desc, err, executions)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.SetTray(ctx, o); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.ClearTray(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.SetTray(context.Background(), app.TrayOptions{Title: "bad\x00"}); err == nil {
		t.Fatal("invalid title")
	}
	if err := s.SetTray(context.Background(), app.TrayOptions{Title: "Vellum", Menu: []app.TrayMenuItem{{Command: "unknown"}}}); err == nil {
		t.Fatal("unknown command")
	}
	a.Close()
	if err := s.ClearTray(context.Background()); !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
}
