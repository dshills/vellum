package notes

import (
	"encoding/json"
	"github.com/dshills/vellum/document"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeNotebookFixture(t *testing.T, count int) string {
	t.Helper()
	m := New()
	for len(m.Items()) < count {
		_, _ = m.Add()
	}
	snapshot, e := m.CaptureNotebook()
	if e != nil {
		t.Fatal(e)
	}
	data, e := EncodeNotebook(snapshot)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "fixture.vellumnotes")
	if e = os.WriteFile(path, data, 0600); e != nil {
		t.Fatal(e)
	}
	return path
}
func TestFinderOpenDirtyCancelSaveAndDiscard(t *testing.T) {
	for _, save := range []bool{true, false} {
		t.Run(map[bool]string{true: "Save", false: "Discard"}[save], func(t *testing.T) {
			x := filesFixture(t, privateDirectory(t))
			_ = x.f.ConfigureRestoration(x.r)
			target := writeNotebookFixture(t, 8)
			x.owner(t, func() {
				_, _ = x.m.Add()
				if e := x.f.OpenPath(target, false); e != nil {
					t.Fatal(e)
				}
				if x.f.prompt != "open" || len(x.m.Items()) != 7 {
					t.Fatal("Finder bypassed dirty decision")
				}
				x.f.promptView().Dismiss()
				if len(x.m.Items()) != 7 || !x.m.History().Dirty {
					t.Fatal("Cancel lost edits")
				}
				if e := x.f.OpenPath(target, false); e != nil {
					t.Fatal(e)
				}
				x.f.settleIntent("open", save)
			})
			x.finish(t)
			x.owner(t, func() {
				if len(x.m.Items()) != 8 || x.f.main.State().Path != target || x.m.History().Dirty {
					t.Fatal("Finder target not installed")
				}
			})
			if save {
				data, e := os.ReadFile(x.r.Path)
				if e != nil {
					t.Fatal(e)
				}
				b, e := DecodeNotebook(data)
				if e != nil || len(b.Notes) != 7 {
					t.Fatal("prior notebook not saved", e)
				}
			}
		})
	}
}
func TestFinderInitialSupersedesAutomaticRestore(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	old := writeNotebookFixture(t, 7)
	wanted := writeNotebookFixture(t, 8)
	data, _ := json.Marshal(WindowSession{1, old, 0})
	_ = os.WriteFile(filepath.Join(x.r.Directory, "WindowSession.json"), data, 0600)
	if e := x.f.ConfigureRestoration(x.r); e != nil {
		t.Fatal(e)
	}
	x.owner(t, func() {
		if e := x.f.OpenPath(wanted, true); e != nil {
			t.Fatal(e)
		}
	})
	x.finish(t)
	x.owner(t, func() {
		if len(x.m.Items()) != 8 || x.f.main.State().Path != wanted {
			t.Fatal("automatic restore overwrote Finder request")
		}
	})
}
func TestFinderInvalidBatchAndRecoveryPreserveLiveNotebook(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	before := x.m.Items()
	target := writeNotebookFixture(t, 8)
	x.owner(t, func() {
		x.f.HandleOpenFiles([]string{target, target}, false)
		if !reflect.DeepEqual(x.m.Items(), before) {
			t.Fatal("batch changed notebook")
		}
		if e := x.f.OpenPath("relative.vellumnotes", true); e == nil {
			t.Fatal("relative path")
		}
		x.f.prompt = "recovery"
		if e := x.f.OpenPath(target, true); e != document.ErrBusy {
			t.Fatal("recovery bypassed", e)
		}
	})

}
