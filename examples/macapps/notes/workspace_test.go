package notes

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dshills/vellum/document"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func retireFiles(t *testing.T, x *fileFixture) {
	t.Helper()
	x.a.Close()
	if e := x.f.Stop(); e != nil {
		t.Fatal(e)
	}
	if e := x.r.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestWindowSessionRestoresExternalWithoutChangingDefault(t *testing.T) {
	dir := privateDirectory(t)
	x := filesFixture(t, dir)
	if e := x.f.ConfigureRestoration(x.r); e != nil {
		t.Fatal(e)
	}
	external := filepath.Join(t.TempDir(), "restored.vellumnotes")
	var expected []Note
	x.owner(t, func() {
		if e := x.f.Save(false); e != nil {
			t.Fatal(e)
		}
	})
	x.finish(t)
	before, _ := os.ReadFile(x.r.Path)
	x.owner(t, func() {
		_, _ = x.m.Add()
		expected = x.m.Items()
		if e := x.f.main.Save(context.Background(), document.SaveOptions{Path: external}); e != nil {
			t.Fatal(e)
		}
	})
	x.finish(t)
	state, e := readWindowSession(dir)
	if e != nil || state.Path != external {
		t.Fatal(state, e)
	}
	retireFiles(t, x)
	next := filesFixture(t, dir)
	if e := next.f.ConfigureRestoration(next.r); e != nil {
		t.Fatal(e)
	}
	next.finish(t)
	next.owner(t, func() {
		if !reflect.DeepEqual(next.m.Items(), expected) || next.m.History().Dirty || next.m.History().UndoCount != 0 || next.f.main.State().Path != external {
			t.Fatal("restored state differs")
		}
	})
	after, _ := os.ReadFile(next.r.Path)
	if !bytes.Equal(before, after) {
		t.Fatal("restore overwrote private default")
	}
}
func TestWindowSessionRecoveryPrecedesRestoreAndRetainsOrigin(t *testing.T) {
	dir := privateDirectory(t)
	x := filesFixture(t, dir)
	_ = x.f.ConfigureRestoration(x.r)
	external := filepath.Join(t.TempDir(), "recover.vellumnotes")
	x.owner(t, func() {
		_, _ = x.m.Add()
		_ = x.f.main.Save(context.Background(), document.SaveOptions{Path: external})
	})
	x.finish(t)
	x.owner(t, func() { _, _ = x.m.Add(); x.f.Checkpoint() })
	x.finish(t)
	retireFiles(t, x)
	next := filesFixture(t, dir)
	if e := next.f.ConfigureRestoration(next.r); e != nil {
		t.Fatal(e)
	}
	next.owner(t, func() {
		if next.f.prompt != "recovery" || len(next.m.Items()) != 6 {
			t.Fatal("auto reopen bypassed recovery")
		}
		next.f.promptView().Children[1].Activate()
		if len(next.m.Items()) != 8 || !next.m.History().Dirty || next.f.main.State().Path != external {
			t.Fatal("recovery lost origin")
		}
		if e := next.f.Save(false); e != nil {
			t.Fatal(e)
		}
	})
	next.finish(t)
	data, e := os.ReadFile(external)
	if e != nil {
		t.Fatal(e)
	}
	b, e := DecodeNotebook(data)
	if e != nil || len(b.Notes) != 8 {
		t.Fatal("recovery save missed original file", e)
	}
}
func TestWindowSessionBadMetadataAndMissingFilePreserveModel(t *testing.T) {
	for _, bad := range []bool{true, false} {
		t.Run(map[bool]string{true: "bad metadata", false: "missing file"}[bad], func(t *testing.T) {
			x := filesFixture(t, privateDirectory(t))
			before := x.m.Items()
			data := []byte(`{"version":99}`)
			if !bad {
				data, _ = json.Marshal(WindowSession{Version: 1, Path: filepath.Join(t.TempDir(), "missing.vellumnotes")})
			}
			path := filepath.Join(x.r.Directory, "WindowSession.json")
			if e := os.WriteFile(path, data, 0600); e != nil {
				t.Fatal(e)
			}
			if e := x.f.ConfigureRestoration(x.r); e != nil {
				t.Fatal(e)
			}
			x.finish(t)
			x.owner(t, func() {
				x.f.Activated()
				if !reflect.DeepEqual(before, x.m.Items()) || x.m.History().Dirty {
					t.Fatal("failed restore changed model")
				}
			})
			x.finish(t)
			after, _ := os.ReadFile(path)
			if !bytes.Equal(after, data) {
				t.Fatal("failed metadata overwritten")
			}
		})
	}
}
func TestWindowSessionRejectsUnsafeAndUnboundedState(t *testing.T) {
	for _, s := range []WindowSession{{1, "relative.vellumnotes", 0}, {1, "/tmp/not-a-note.vforge", 0}, {1, "/tmp/a.vellumnotes", -1}} {
		if s.validate() == nil {
			t.Fatal("unsafe session accepted", s)
		}
	}
	if _, e := decodeWindowSession([]byte(`{"version":1,"path":"/tmp/a.vellumnotes","activated":0,"unknown":1}`)); e == nil {
		t.Fatal("unknown metadata accepted")
	}
	if _, e := decodeWindowSession(make([]byte, workspaceLimit+1)); e == nil {
		t.Fatal("unbounded metadata accepted")
	}
}

func TestWindowSessionWriteFailurePreventsCoordinatedQuit(t *testing.T) {
	x := filesFixture(t, privateDirectory(t))
	if e := x.f.ConfigureRestoration(x.r); e != nil {
		t.Fatal(e)
	}
	// Fail the real metadata rename without changing the notebook/recovery stores.
	if e := os.Mkdir(filepath.Join(x.r.Directory, "WindowSession.json"), 0700); e != nil {
		t.Fatal(e)
	}
	group, e := NewQuitGroup(x.a, x.f)
	if e != nil {
		t.Fatal(e)
	}
	x.owner(t, func() { _, _ = x.m.Add(); group.Request(); x.f.settleIntent("quit", true) })
	quitDrain(t, x.a, group)
	x.owner(t, func() {
		if len(x.a.Windows()) != 1 || x.f.quitHeld || x.m.History().Dirty {
			t.Fatal("failed metadata lost saved window")
		}
	})
}
