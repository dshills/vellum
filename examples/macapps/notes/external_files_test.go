package notes

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
)

type externalFileChooser struct {
	path  string
	calls chan app.FileOptions
}

func (s externalFileChooser) ChooseFile(_ context.Context, options app.FileOptions) (string, error) {
	select {
	case s.calls <- options:
		return s.path, nil
	default:
		return "", app.ErrUnavailable
	}
}
func (externalFileChooser) ReadClipboard(context.Context) (string, error) {
	return "", app.ErrUnavailable
}
func (externalFileChooser) WriteClipboard(context.Context, string) error   { return app.ErrUnavailable }
func (externalFileChooser) SetMenu(context.Context, []app.MenuGroup) error { return nil }

func TestExternalSaveAsSubsequentSaveAndExplicitReopen(t *testing.T) {
	directory := privateDirectory(t)
	x := filesFixture(t, directory)
	x.owner(t, func() {
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	original, err := os.ReadFile(x.r.Path)
	if err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(privateDirectory(t), "External notebook.vellumnotes")
	if err := os.MkdirAll(filepath.Dir(external), 0700); err != nil {
		t.Fatal(err)
	}
	calls := make(chan app.FileOptions, 2)
	chooser := externalFileChooser{external, calls}
	x.a.SetServices(chooser)
	x.owner(t, func() {
		if _, err := x.m.Add(); err != nil {
			t.Fatal(err)
		}
		if err := x.m.Commands().Invoke(context.Background(), "notes.save-as"); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	choice := nextFileChoice(t, calls)
	if choice.Mode != app.SaveFile || choice.Directory != directory {
		t.Fatal("Save As chooser lost current directory", choice)
	}
	x.owner(t, func() {
		if x.f.main.State().Path != external || x.m.History().Dirty {
			t.Fatalf("Save As path=%q expected=%q state=%+v history=%+v", x.f.main.State().Path, external, x.f.main.State(), x.m.History())
		}
		if err := x.m.Update(x.m.Selected(), "External only", "Preserved 😀 content"); err != nil {
			t.Fatal(err)
		}
		if err := x.f.Save(false); err != nil {
			t.Fatal(err)
		}
	})
	x.finish(t)
	if len(calls) != 0 {
		t.Fatal("ordinary Save reopened chooser")
	}
	unchanged, err := os.ReadFile(x.r.Path)
	if err != nil || !bytes.Equal(unchanged, original) {
		t.Fatal("external Save overwrote private default", err)
	}
	data, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := DecodeNotebook(data)
	if err != nil {
		t.Fatal(err)
	}
	x.a.Close()
	if err := x.f.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := x.r.Close(); err != nil {
		t.Fatal(err)
	}
	next := filesFixture(t, directory)
	next.owner(t, func() {
		if len(next.m.Items()) == len(saved.Notes) {
			t.Fatal("restart silently reopened external file")
		}
	})
	next.a.SetServices(chooser)
	next.owner(t, func() {
		if err := next.m.Commands().Invoke(context.Background(), "notes.open"); err != nil {
			t.Fatal(err)
		}
	})
	next.finish(t)
	if nextFileChoice(t, calls).Mode != app.OpenFile {
		t.Fatal("wrong Open chooser mode")
	}
	next.owner(t, func() {
		if !reflect.DeepEqual(next.m.Items(), saved.Notes) || next.m.History().Dirty || next.m.History().UndoCount != 0 || next.f.main.State().Path != external {
			t.Fatal("explicit Open lost content/path/history boundary")
		}
		if err := next.m.Commands().Enabled("notes.undo"); err == nil {
			t.Fatal("fresh external history availability", err)
		}
	})
}

func nextFileChoice(t *testing.T, calls <-chan app.FileOptions) app.FileOptions {
	t.Helper()
	select {
	case options := <-calls:
		return options
	case <-time.After(time.Second):
		t.Fatal("expected chooser was not invoked")
		return app.FileOptions{}
	}
}
