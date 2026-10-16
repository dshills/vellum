package main

import (
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestPNGReplacementPreservesFileOnEncodingFailureOrCancellation(t *testing.T) {
	for _, cancelDuringWrite := range []bool{false, true} {
		dir := t.TempDir()
		path := filepath.Join(dir, "capture.png")
		original := []byte("existing document")
		if err := os.WriteFile(path, original, 0640); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		failed := errors.New("controlled encode refusal")
		encoder := func(w io.Writer, _ image.Image) error {
			_, err := w.Write([]byte("partial PNG"))
			if err != nil {
				return err
			}
			if cancelDuringWrite {
				cancel()
				return nil
			}
			return failed
		}
		err := replacePNG(ctx, path, image.NewNRGBA(image.Rect(0, 0, 2, 2)), encoder)
		if cancelDuringWrite && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if !cancelDuringWrite && !errors.Is(err, failed) {
			t.Fatal(err)
		}
		got, e := os.ReadFile(path)
		if e != nil || string(got) != string(original) {
			t.Fatalf("existing destination changed: %q %v", got, e)
		}
		files, e := os.ReadDir(dir)
		if e != nil || len(files) != 1 {
			t.Fatalf("temporary leaked: %v %v", files, e)
		}
	}
}
func TestPNGReplacementCommitsCompleteImageAndMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "capture.png")
	if err := os.WriteFile(path, []byte("old"), 0640); err != nil {
		t.Fatal(err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(1, 1, color.NRGBA{2, 3, 4, 128})
	if err := savePNG(context.Background(), path, img); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoded, err := png.Decode(file)
	if err != nil || decoded.Bounds() != img.Bounds() {
		t.Fatalf("incomplete PNG: %v %v", decoded, err)
	}
	got := color.NRGBAModel.Convert(decoded.At(1, 1)).(color.NRGBA)
	if got != img.NRGBAAt(1, 1) {
		t.Fatal(got)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0640 {
		t.Fatalf("mode changed: %v %v", info, err)
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 1 {
		t.Fatalf("temporary leaked: %v %v", files, err)
	}
}

// observedWaitContext exposes PostWait's blocking admission branch without sleeps.
type observedWaitContext struct {
	context.Context
	once    sync.Once
	waiting chan struct{}
}

func (c *observedWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

type saveServices struct {
	app.Services
	path string
}

func (s saveServices) ChooseFile(context.Context, app.FileOptions) (string, error) {
	return s.path, nil
}

func TestSaveCompletionWaitsForMailboxCredit(t *testing.T) {
	a, err := app.New(app.Config{QueueCapacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err := a.Open(app.Window{ID: "save", Title: "Save", Width: 100, Height: 100, View: func() ui.Node { return ui.Node{Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	windowCtx, err := a.WindowContext("save")
	if err != nil {
		t.Fatal(err)
	}
	ctx := &observedWaitContext{Context: windowCtx, waiting: make(chan struct{})}
	path := filepath.Join(t.TempDir(), "capture.png")
	a.SetServices(saveServices{path: path})
	w := &imageWindow{application: a, view: &cogent.GPUView{Window: "save"}, captureBusy: true}
	if err := a.Post(func() {}); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { w.save(ctx, image.NewNRGBA(image.Rect(0, 0, 2, 2))); close(done) }()
	select {
	case <-ctx.waiting:
	case <-done:
		t.Fatal("save returned with completion dropped from full mailbox")
	case <-time.After(2 * time.Second):
		t.Fatal("save did not reach mailbox wait")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("completion not admitted after mailbox credit")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := a.Do(func() {
		if w.captureBusy || w.status != "PNG capture saved" {
			t.Errorf("capture terminal state lost: busy=%v status=%q", w.captureBusy, w.status)
		}
	}); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err := png.Decode(file); err != nil {
		t.Fatal(err)
	}
}
func TestSaveCompletionDropsAfterWindowLifetimeEnds(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	for _, id := range []string{"save", "survivor"} {
		if err := a.Open(app.Window{ID: id, Title: id, Width: 100, Height: 100, View: func() ui.Node { return ui.Node{Kind: ui.Label} }}); err != nil {
			t.Fatal(err)
		}
	}
	ref, err := a.WindowRef("save")
	if err != nil {
		t.Fatal(err)
	}
	a.SetServices(saveServices{path: filepath.Join(t.TempDir(), "capture.png")})
	w := &imageWindow{application: a, view: &cogent.GPUView{Window: "save"}, captureBusy: true, status: "original"}
	// Save is a worker method; no native chooser or UI owner is needed here.
	w.save(ref.Context(), image.NewNRGBA(image.Rect(0, 0, 2, 2)))
	if err := a.CloseWindow(ref); err != nil {
		t.Fatal(err)
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if err := a.Do(func() {
		if !w.captureBusy || w.status != "original" {
			t.Errorf("closed window received terminal UI mutation: busy=%v status=%q", w.captureBusy, w.status)
		}
	}); err != nil {
		t.Fatal(err)
	}
}
