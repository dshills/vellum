//go:build offscreen

package workspace

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/dshills/vellum/replay"
)

func runRecord(t *testing.T, rows, batch int) replay.Result {
	t.Helper()
	record, err := Record(rows) // Parse/validate before creating the backend.
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewBackend(context.Background(), rows, batch)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := b.Close(context.Background()); err != nil {
			t.Error(err)
		}
		if !b.Executor.Snapshot().Idle() || !b.Fixture.Owner.Pending().Idle() {
			t.Error("residual work after teardown")
		}
	})
	runner, err := replay.NewRunner(replay.RunnerConfig{EventBatch: batch, MaxSteps: 100000, MaxCheckpointBytes: 32 << 20}, b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := runner.Run(context.Background(), record)
	if err != nil {
		t.Fatalf("applied %d/%d, residual %+v: %v", got.Applied, len(record.Operations), got.Residual, err)
	}
	if got.Applied != len(record.Operations) || !got.QuiescentNow || !b.Executor.Snapshot().Idle() {
		t.Fatal("incomplete workspace replay", got.Applied, got.Residual)
	}
	if !b.Clock.Now().Equal(record.Manifest.Start.Add(record.Operations[len(record.Operations)-1].At)) {
		t.Fatal("implicit time advancement")
	}
	idle, err := runner.RunUntilIdle(context.Background())
	if err != nil || !idle.QuiescentNow || idle.Progress.Units != 0 {
		t.Fatal("idle dispatched extra work", idle, err)
	}
	t.Logf("rows=%d batch=%d operations=%d checkpoints=%d progress=%+v", rows, batch, got.Applied, len(got.Checkpoints), got.Progress)
	return got
}
func compact(t *testing.T, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := json.Compact(&out, b); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func firstDifference(path string, a, b any) string {
	if reflect.DeepEqual(a, b) {
		return ""
	}
	if am, ok := a.(map[string]any); ok {
		if bm, ok := b.(map[string]any); ok {
			for k, v := range am {
				if d := firstDifference(path+"."+k, v, bm[k]); d != "" {
					return d
				}
			}
		}
	}
	if aa, ok := a.([]any); ok {
		if ba, ok := b.([]any); ok && len(aa) == len(ba) {
			for i := range aa {
				if d := firstDifference(fmt.Sprintf("%s[%d]", path, i), aa[i], ba[i]); d != "" {
					return d
				}
			}
		}
	}
	return fmt.Sprintf("%s: expected %.180v got %.180v", path, a, b)
}
func TestFrozenWorkspaceRecordsAcrossBatches(t *testing.T) {
	for _, rows := range []int{10000, MaxRows} {
		data, err := recordFiles.ReadFile(fmt.Sprintf("records/%d.golden.json", rows))
		if err != nil {
			t.Fatal(err)
		}
		var expected []replay.Checkpoint
		if err := json.Unmarshal(data, &expected); err != nil {
			t.Fatal(err)
		}
		for _, batch := range []int{1, 7, 64} {
			t.Run(fmt.Sprintf("%d/%d", rows, batch), func(t *testing.T) {
				got := runRecord(t, rows, batch)
				if len(got.Checkpoints) != len(expected) {
					t.Fatal("checkpoint count", len(got.Checkpoints), len(expected))
				}
				for i, c := range got.Checkpoints {
					e := expected[i]
					if c.Sequence != e.Sequence || c.Label != e.Label || c.Window != e.Window {
						t.Fatal("checkpoint identity", i, c.Label)
					}
					// Parse and encode both values to preserve JSON meaning independently of
					// formatting and HTML escapes. Native capability tokens alone are normalized
					// by Capture; model content, geometry, semantics and UTF-16 ranges remain.
					var want, actual any
					if err := json.Unmarshal(e.Data, &want); err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal(c.Data, &actual); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(want, actual) {
						t.Errorf("%s %s", c.Label, firstDifference("capture", want, actual))
					}
					if c.ImageSHA256 != e.ImageSHA256 {
						t.Errorf("%s image differs: expected %s got %s", c.Label, e.ImageSHA256, c.ImageSHA256)
					}
				}
				assertScenario(t, rows, got.Checkpoints)
			})
		}
	}
}
func assertScenario(t *testing.T, rows int, points []replay.Checkpoint) {
	t.Helper()
	all := map[string]Capture{}
	images := map[string]string{}
	for _, p := range points {
		var c Capture
		if err := json.Unmarshal(p.Data, &c); err != nil {
			t.Fatal(err)
		}
		all[p.Label] = c
		images[p.Label] = p.ImageSHA256
		if len(c.Windows) == 0 || len(p.ImageSHA256) != 64 || c.Model.Failure != "" {
			t.Fatal("missing observations", p.Label)
		}
	}
	get := func(label string) Snapshot {
		c, ok := all[label]
		if !ok {
			t.Fatal("missing checkpoint", label)
		}
		return c.Model
	}
	opened := get("opened")
	if !opened.Loaded || opened.StoredRows != rows || opened.StoredColumns != 20 || opened.StoredTreeNodes != 100000 || opened.Grid.Stamp.Rows != rows || opened.Tree.VisibleRows != 100000 || opened.Name != "North Star" {
		t.Fatal("wrong stored fixture", opened.StoredRows, opened.Tree)
	}
	// Independently derive stored-byte identity and numeric/filter oracles.
	h := sha256.New()
	var raw [80]byte
	best, bestRow, filterBest, filterRow, filterCount := -1, -1, -1, -1, 0
	for r := 0; r < rows; r++ {
		for c := 0; c < 20; c++ {
			binary.LittleEndian.PutUint32(raw[c*4:], uint32((r*97+c*13+7)%10007))
		}
		h.Write(raw[:])
		v := (r*97 + 7) % 10007
		if v > best {
			best, bestRow = v, r
		}
		if strings.Contains(strconv.Itoa((r*97+20)%10007), "7") {
			filterCount++
			if v > filterBest {
				filterBest, filterRow = v, r
			}
		}
	}
	if opened.DataSHA256 != hex.EncodeToString(h.Sum(nil)) {
		t.Fatal("stored byte identity")
	}
	edited := get("property-edited")
	if edited.Name != "Harbor" || edited.History.UndoCount != 1 || !edited.History.Dirty {
		t.Fatal("property history")
	}
	sorted := get("sorted")
	if !sorted.Sorted || sorted.FirstKey != fmt.Sprintf("row:%07d", bestRow) || sorted.FirstValue != strconv.Itoa(best) {
		t.Fatal("real numeric descending sort", sorted.FirstKey, sorted.FirstValue)
	}
	selected := get("grid-scrolled-selected")
	if selected.Grid.Active.RowKey == "" || selected.Grid.Active.ColumnID != "c00" || selected.Grid.Selection.CapturedRows != 1 || selected.Grid.Viewport.ScrollX <= 0 || selected.Grid.Viewport.ScrollY == sorted.Grid.Viewport.ScrollY {
		t.Fatal("routed selection or two-axis scroll", selected.Grid)
	}
	index, err := strconv.Atoi(strings.TrimPrefix(selected.Grid.Active.RowKey, "row:"))
	if err != nil || selected.ActiveValue != strconv.Itoa((index*97+7)%10007) {
		t.Fatal("active cell value")
	}
	canceled := get("cancel-delivered")
	if canceled.Filtered || canceled.Grid.Stamp != selected.Grid.Stamp || !reflect.DeepEqual(canceled.Grid.Active, selected.Grid.Active) || !canceled.Task.OwnerDone || !canceled.Task.Canceled || len(canceled.Outcomes) < 3 || canceled.Outcomes[2] != (outcome{"query", "canceled"}) {
		t.Fatal("canceled worker mutated installed view", canceled.Task)
	}
	filtered := get("filtered")
	if !filtered.Filtered || filtered.Grid.Stamp.Rows != filterCount || filtered.FirstKey != fmt.Sprintf("row:%07d", filterRow) || filtered.FirstValue != strconv.Itoa(filterBest) {
		t.Fatal("real filtered projection", filtered.Grid.Stamp, filtered.FirstKey)
	}
	if get("tree-collapsed").Tree.VisibleRows != 1 || get("tree-expanded").Tree.VisibleRows != 100000 || get("tree-filtered").Tree.VisibleRows != 2 || get("tree-filtered").Tree.Pending {
		t.Fatal("worker outline indexes")
	}
	notes := get("notes-edited")
	if notes.Editor.Text != "Notes Ω\n" || notes.Editor.Selection.Focus != 8 || notes.Editor.Composing {
		t.Fatal("routed Unicode editor", notes.Editor)
	}
	reordered := get("tabs-reordered")
	if !reflect.DeepEqual(reordered.Layout.Root.Second.First.Tabs, []string{"canvas", "notes", "grid"}) || reordered.Editor.Text != notes.Editor.Text {
		t.Fatal("actual tab drag lost content", reordered.Layout)
	}
	preview, placed := get("placement-preview"), get("placed")
	if len(preview.Canvas.Objects) != 0 || len(placed.Canvas.Objects) != 1 || placed.Canvas.Objects[0].ID != "block-1" || placed.History.UndoCount != 2 || images["placement-preview"] == images["placed"] {
		t.Fatal("placement/preview commit")
	}
	navigated := get("canvas-navigated")
	if navigated.Canvas.View.Zoom == placed.Canvas.View.Zoom || navigated.Canvas.View.Center == placed.Canvas.View.Center || navigated.Canvas.Selected != "block-1" {
		t.Fatal("canvas selection, pan and zoom")
	}
	if len(get("placement-undone").Canvas.Objects) != 0 || !reflect.DeepEqual(get("placement-redone").Canvas.Objects, placed.Canvas.Objects) {
		t.Fatal("placement undo/redo identity")
	}
	undone := get("undone")
	if undone.Name != "North Star" || undone.History.Dirty || undone.History.UndoCount != 0 || len(undone.Canvas.Objects) != 0 || undone.Editor.Text != "Notes Ω\n" {
		t.Fatal("independent document history")
	}
	if !get("modal-open").Modal || get("modal-closed").Modal || get("modal-open").Canvas.View != get("modal-closed").Canvas.View {
		t.Fatal("modal background scroll escaped")
	}
	survivor := all["surviving-inspector"]
	if len(survivor.Windows) != 1 || survivor.Windows[0].ID != "inspector" || survivor.Model.Name != "North Star" {
		t.Fatal("independent window lifetime")
	}
}
