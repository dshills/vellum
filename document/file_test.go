package document

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteBoundedReadAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc")
	if err := os.WriteFile(path, []byte("old"), 0640); err != nil {
		t.Fatal(err)
	}
	written, err := (FileStore{}).Write(context.Background(), path, []byte("new bytes"), 16)
	if !written || err != nil {
		t.Fatal(written, err)
	}
	data, err := (FileStore{}).Read(context.Background(), path, 16)
	if err != nil || string(data) != "new bytes" {
		t.Fatal(string(data), err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0640 {
		t.Fatal(info.Mode())
	}
	if _, err = (FileStore{}).Read(context.Background(), path, 3); !errors.Is(err, ErrTooLarge) {
		t.Fatal(err)
	}
}
func TestAtomicFailuresPreserveOriginalAndCleanTemporaryFiles(t *testing.T) {
	failure := errors.New("injected")
	for _, stage := range []string{"sync", "rename", "directory", "cancel-before", "cancel-after"} {
		t.Run(stage, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "doc")
			os.WriteFile(path, []byte("old"), 0600)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ops := defaultAtomicFileOps
			switch stage {
			case "sync":
				ops.syncFile = func(*os.File) error { return failure }
			case "rename":
				ops.rename = func(string, string) error { return failure }
			case "directory":
				ops.syncDirectory = func(*os.File) error { return failure }
			case "cancel-before":
				ops.syncFile = func(f *os.File) error { cancel(); return f.Sync() }
			case "cancel-after":
				ops.syncDirectory = func(f *os.File) error { cancel(); return f.Sync() }
			}
			written, err := writeAtomic(ctx, path, []byte("new"), 16, ops)
			post := stage == "directory" || stage == "cancel-after"
			if err == nil || written != post {
				t.Fatal(written, err)
			}
			b, _ := os.ReadFile(path)
			want := "old"
			if post {
				want = "new"
			}
			if string(b) != want {
				t.Fatal(string(b), want)
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 1 {
				t.Fatal("temporary files leaked", entries)
			}
		})
	}
}
func TestFileAdmissionRejectsSpecialPathsAndBounds(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	os.WriteFile(target, []byte("old"), 0600)
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if written, err := (FileStore{}).Write(context.Background(), link, []byte("new"), 16); err == nil || written {
		t.Fatal("symlink save accepted")
	}
	b, _ := os.ReadFile(target)
	if string(b) != "old" {
		t.Fatal("followed symlink")
	}
	if _, err := (FileStore{}).Read(context.Background(), dir, 16); err == nil {
		t.Fatal("directory read admitted")
	}
	for _, max := range []int64{0, -1, MaximumMaxBytes + 1} {
		if _, err := (FileStore{}).Read(context.Background(), target, max); err == nil {
			t.Fatal("limit accepted", max)
		}
	}
	if _, err := (FileStore{}).Read(context.Background(), "relative", 16); err == nil {
		t.Fatal("relative read")
	}
	if written, err := (FileStore{}).Write(context.Background(), target, []byte("long"), 3); written || !errors.Is(err, ErrTooLarge) {
		t.Fatal(written, err)
	}
}
