//go:build offscreen

package notes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/text/fonts"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/ui"
)

func settleViewport(t *testing.T, m *cogent.Manual) {
	t.Helper()
	ctx := context.Background()
	for range 128 {
		s, err := m.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ready := s.Application.Ordinary+s.Application.ReadyCompletions > 0
		for _, w := range s.Windows {
			ready = ready || w.Events > 0 || w.ReadyPaint
		}
		if ready {
			if _, err = m.Step(ctx, 7); err != nil {
				t.Fatal(err)
			}
			continue
		}
		pending := false
		for _, w := range s.Windows {
			if w.LayoutPending {
				if _, err = m.Frame(ctx, w.ID); err != nil {
					t.Fatal(err)
				}
				pending = true
			}
		}
		if !pending {
			return
		}
	}
	t.Fatal("viewport settlement exhausted")
}

// Check allocated output: a compact search field and folder rows, note-card
// density and non-overlapping three-column geometry at both supported sizes.
// This exercises real layout and rendering, not physical/native pointer input.
func TestNotesMacLayout(t *testing.T) {
	for _, size := range []image.Point{image.Pt(1100, 740), image.Pt(1380, 900)} {
		t.Run(size.String(), func(t *testing.T) { notesViewport(t, size) })
	}
}
func notesViewport(t *testing.T, size image.Point, checks ...func(*Model, *app.Application, *cogent.Manual)) {
	ctx := context.Background()
	model := New()
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "notes", Title: "Notes viewport", Width: float32(size.X + 16), Height: float32(size.Y + 16), View: model.View}); err != nil {
		t.Fatal(err)
	}
	clock, err := execution.NewManualClock(time.Unix(1, 0), 256, 256)
	if err != nil {
		t.Fatal(err)
	}
	defer clock.Close()
	executor, err := execution.NewManual(8)
	if err != nil {
		t.Fatal(err)
	}
	defer executor.Close()
	theme := Theme()
	theme.Default.Background = ui.Some(color.NRGBA{})

	config := cogent.ManualConfig{Commands: model.Commands(), Clock: clock, Executor: executor, Theme: theme, LogicalDPI: 160, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128}
	err = fs.WalkDir(fonts.Default, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".ttf") && !strings.HasSuffix(path, ".otf") {
			return nil
		}
		data, err := fs.ReadFile(fonts.Default, path)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(data)
		config.Fonts = append(config.Fonts, cogent.ManualFont{Name: path, Data: data, SHA256: hex.EncodeToString(hash[:])})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	m, err := cogent.NewManual(ctx, a, config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := m.Close(ctx); err != nil {
			t.Error(err)
		}
	}()
	settleViewport(t, m)
	snapshot, err := m.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	elements := map[string]cogent.ManualElement{}
	for _, e := range snapshot.Windows[0].Elements {

		elements[e.ID] = e
	}
	get := func(id string) image.Rectangle {
		t.Helper()
		e, ok := elements[id]
		if !ok {
			t.Fatal("missing allocated element", id)
		}
		return e.Bounds
	}
	content := get("notes-content")
	if content.Min.Y > 10 {
		t.Errorf("unused header band remains: %v", content)
	}
	for _, id := range []string{"notes-folders", "notes-list", "notes-document"} {
		if b := get(id); !b.In(content) || b.Dy() < size.Y-30 {
			t.Errorf("pane %s: %v within %v", id, b, content)
		}
	}
	if a, b := get("notes-folders"), get("notes-list"); a.Max.X >= b.Min.X || a.Dx() != 210 || b.Dx() != 320 {
		t.Errorf("column geometry %v %v", a, b)
	}
	if b := get("notes-search"); b.Dy() > 32 || b.Dx() < 270 {
		t.Errorf("oversized/collapsed search: %v", b)
	}
	for _, id := range []string{"all", "notes", "projects", "personal"} {
		row, name, count := get("folder-row-"+id), get("folder-"+id), get("folder-count-"+id)
		if row.Dy() > 32 || name.Max.X > count.Min.X || !count.In(row) || count.Max.X < row.Max.X-2 {
			t.Errorf("folder spacing/count collision %s: %v %v %v", id, row, name, count)
		}
	}
	for _, v := range model.Visible() {
		card := get("note-card-" + v.ID)
		if card.Dy() > 80 || card.Dy() < 52 {
			t.Errorf("note card density %s: %v", v.ID, card)
		}
		for _, id := range []string{"note-select-", "note-preview-", "note-date-"} {
			if b := get(id + v.ID); !b.In(card) {
				t.Errorf("clipped card child %s %v within %v", id, b, card)
			}
		}
	}
	title, body := get("note-title-n1"), get("note-editor-n1")
	if title.Dy() > 32 || title.Max.Y > body.Min.Y || body.Dy() < size.Y-200 || !body.In(get("notes-document")) {
		t.Errorf("writing area clipped: title=%v body=%v", title, body)
	}
	if output := os.Getenv("VELLUM_NOTES_LAYOUT_PNG"); output != "" && size.X == 1100 {
		pixels, err := m.Capture(ctx, "notes")
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(output)
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, pixels)
		closeErr := f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	for _, check := range checks {
		check(model, a, m)
	}

}
