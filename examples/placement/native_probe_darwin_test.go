//go:build darwin && cgo && !offscreen && vellum_placement_probe

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func privateProbeDir(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	return root
}
func TestNativeProbeWriterPrivateAtomicBounds(t *testing.T) {
	root := privateProbeDir(t)
	w, err := newNativeProbeWriter(root)
	if err != nil {
		t.Fatal(err)
	}
	defer w.log.Close()
	if err := w.write(map[string]int{"value": 1}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "latest.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"record", "total", "foreign-temp"} {
		t.Run(mode, func(t *testing.T) {
			value := any(map[string]int{"value": 2})
			switch mode {
			case "record":
				value = strings.Repeat("x", 256<<10)
			case "total":
				w.bytes = nativeProbeByteLimit
			case "foreign-temp":
				w.bytes = 0
				if err := os.WriteFile(filepath.Join(root, "latest.tmp"), []byte("foreign"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := w.write(value); err == nil {
				t.Fatal("write bound/collision accepted")
			}
			after, err := os.ReadFile(filepath.Join(root, "latest.json"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("refusal replaced latest state", err)
			}
		})
	}
	if data, err := os.ReadFile(filepath.Join(root, "latest.tmp")); err != nil || string(data) != "foreign" {
		t.Fatal("foreign temporary file changed", err)
	}
	if _, err := newNativeProbeWriter(root); err == nil {
		t.Fatal("nonempty directory accepted")
	}
	other := privateProbeDir(t)
	if err := os.Chmod(other, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := newNativeProbeWriter(other); err == nil {
		t.Fatal("nonprivate directory accepted")
	}
	parent := privateProbeDir(t)
	link := filepath.Join(parent, "link")
	if err := os.Symlink(other, link); err != nil {
		t.Fatal(err)
	}
	if _, err := newNativeProbeWriter(link); err == nil {
		t.Fatal("symlink directory accepted")
	}
}
func TestNativeProbeRefusesTestDriver(t *testing.T) {
	if err := nativePlacementPlatform(); err == nil {
		t.Fatal("offscreen test app accepted as native")
	}
}
func TestNativeProbeObserverRetirement(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	held, started := make(chan struct{}), make(chan struct{})
	o := startNativeProbeObserver(cancel, func() error { close(started); <-held; return errors.New("retained observer failure") })
	heldObserver := o
	released := false
	release := func() {
		if !released {
			close(held)
			released = true
		}
	}
	t.Cleanup(func() {
		release()
		select {
		case <-heldObserver.done:
		case <-time.After(time.Second):
			t.Error("held control did not join")
		}
	})
	<-started
	cleaned := false
	cleanup := func() error { cleaned = true; return nil }
	if err := retireNativeProbe(cancel, o, time.Millisecond, cleanup); !errors.Is(err, errNativeObserverNotJoined) || cleaned || ctx.Err() == nil {
		t.Fatal("held observer authorized shared teardown", err, cleaned)
	}
	release()
	if err := retireNativeProbe(cancel, o, time.Second, cleanup); err == nil || !strings.Contains(err.Error(), "retained observer failure") || !cleaned {
		t.Fatal("joined observer failure lost", err, cleaned)
	}
	cleaned = false
	o = startNativeProbeObserver(cancel, func() error { panic("inert panic control") })
	if err := retireNativeProbe(cancel, o, time.Second, cleanup); err == nil || !strings.Contains(err.Error(), "inert panic control") || !cleaned {
		t.Fatal("observer panic/join not retained", err, cleaned)
	}
}
