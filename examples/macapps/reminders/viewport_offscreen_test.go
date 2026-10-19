//go:build offscreen

package reminders

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"io/fs"
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

// Match the native demo's 1000x740 client content with the standard offscreen
// scene's extra 8dp on each side. This checks real allocated/public geometry,
// not just Grow style records, and makes no physical pointer claim.
func TestRemindersWideViewportAllocatesHeadingRowsAndComposer(t *testing.T) {
	ctx := context.Background()
	model := New()
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "reminders", Title: "Reminders viewport", Width: 1016, Height: 756, View: model.View}); err != nil {
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
	theme.Default.Background = ui.Some(white)
	config := cogent.ManualConfig{Clock: clock, Executor: executor, Theme: theme, LogicalDPI: 160, OwnerBatch: 7, MaxWindows: 1, EventCapacity: 128}
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
	get := func(id string) cogent.ManualElement {
		t.Helper()
		e, ok := elements[id]
		if !ok {
			t.Fatal("missing allocated element", id)
		}
		return e
	}
	viewport := image.Rect(8, 8, 1008, 748)
	content := get("reminders-content").Bounds
	if content.Dx() < 700 || !content.In(viewport) {
		t.Errorf("content does not use available viewport: %v", content)
	}
	for _, id := range []string{"reminders-title-row", "reminders-heading", "reminders-list", "row-r1", "reminders-compose-row"} {
		if bounds := get(id).Bounds; bounds.Dx() < 600 || !bounds.In(content) {
			t.Errorf("content width collapsed at %s: %v within %v", id, bounds, content)
		}
	}
	if title := get("reminders-title").Bounds; title.Dx() < 250 || title.Dy() > 60 {
		t.Errorf("list title wraps instead of occupying heading width: %v", title)
	}
	if count := get("reminders-count").Bounds; count.Max.X < content.Max.X-80 {
		t.Errorf("heading count did not reach trailing column: %v content %v", count, content)
	}
	if summary := get("reminders-completed-count").Bounds; summary.Dy() > 26 {
		t.Errorf("completed summary wrapped: %v", summary)
	}
	if input := get("reminders-new-title").Bounds; input.Dx() < 350 || !input.In(content) {
		t.Errorf("composer input collapsed: %v", input)
	}
	if details := get("details-r1").Bounds; details.Max.X < content.Max.X-80 {
		t.Errorf("row details did not reach trailing column: %v content %v", details, content)
	}
	for _, id := range []string{"today", "scheduled", "all", "flagged", "completed", "assigned"} {
		icon, count := get("smart-icon-"+id).Bounds, get("smart-count-"+id).Bounds
		if icon.Max.X+8 > count.Min.X || count.Dy() > 42 || !count.In(get("smart-card-"+id).Bounds) {
			t.Errorf("smart tile icon/count overlap or wrap: %s icon=%v count=%v", id, icon, count)
		}
	}
	for _, id := range []string{"reminders", "family", "work", "camping", "groceries"} {
		name, count := get("list-"+id), get("list-count-"+id)
		if name.Bounds.Max.X > count.Bounds.Min.X || !count.Bounds.In(get("list-row-"+id).Bounds) {
			t.Errorf("sidebar name and count overlap: %s name=%v count=%v", id, name.Bounds, count.Bounds)
		}
	}
}
