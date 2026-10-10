package ui

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFileDropPreservesPathsAndDoesNotAlias(t *testing.T) {
	path := filepath.Join(t.TempDir(), "space Ω\nfilename")
	input := []string{path, path, filepath.Join(filepath.Dir(path), "uncreated")}
	got, err := CopyFileDropPaths(input)
	if err != nil || !reflect.DeepEqual(got, input) {
		t.Fatalf("%q %v", got, err)
	}
	input[0] = "caller mutation"
	if got[0] != path {
		t.Fatal("input slice aliased")
	}
	got[1] = "callback mutation"
	if input[1] != path {
		t.Fatal("output slice aliased")
	}
}
func TestFileDropRejectsWholeBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "valid")
	many := make([]string, MaxFileDropPaths+1)
	for i := range many {
		many[i] = path
	}
	for _, paths := range [][]string{nil, {}, {path, ""}, {path, "relative"}, {path, path + "\x00"}, {path, path + "\xff"}, many, {path + strings.Repeat("x", MaxFileDropBytes)}, {path, path + strings.Repeat("x", MaxFileDropBytes-len(path))}} {
		if got, err := CopyFileDropPaths(paths); err == nil || got != nil {
			t.Fatalf("partial or invalid batch admitted: %d paths", len(paths))
		}
	}
	exact := path + strings.Repeat("x", MaxFileDropBytes-len(path))
	if _, err := CopyFileDropPaths([]string{exact}); err != nil {
		t.Fatal("exact byte limit refused", err)
	}
	if _, err := CopyFileDropPaths(many[:MaxFileDropPaths]); err != nil {
		t.Fatal("exact count limit refused", err)
	}
}
