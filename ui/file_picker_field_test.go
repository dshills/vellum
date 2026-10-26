package ui

import (
	"path/filepath"
	"testing"
)

func TestFilePickerControlledActionsAndAdmission(t *testing.T) {
	calls := []FilePickerEvent{}
	p := &FilePickerField{ID: "path", Kind: FilePickerFile, Extensions: []string{"txt"}, Request: func(e FilePickerEvent) error { calls = append(calls, e); return nil }}
	n, e := p.View()
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "a.txt")
	if !n.FileDrop(FileDropEvent{Paths: []string{path}}) || len(calls) != 1 || calls[0].Path != path || p.Path != "" {
		t.Fatal(calls)
	}
	n.FileDrop(FileDropEvent{Paths: []string{filepath.Join(t.TempDir(), "a.png")}})
	if len(calls) != 1 {
		t.Fatal("bad suffix proposed")
	}
	p.Disabled = true
	n.FileDrop(FileDropEvent{Paths: []string{path}})
	if len(calls) != 1 {
		t.Fatal("disabled proposed")
	}
	if ValidatePickerPath("relative", FilePickerFile, nil) == nil || ValidatePickerPath(path, FilePickerFolder, []string{"txt"}) == nil {
		t.Fatal("bad picker metadata")
	}
}
