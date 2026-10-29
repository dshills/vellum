package cogent

import (
	"github.com/dshills/vellum/app"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestOpenFilesOwnerBoundsReadinessAndRetirement(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	var calls atomic.Int32
	events := make(chan OpenFilesEvent, 20)
	r := &openFilesReceiver{owner: a, callback: func(e OpenFilesEvent) { calls.Add(1); events <- e }}
	directory := t.TempDir()
	want := filepath.Join(directory, "a.vellumnotes")
	paths := []string{want}
	if !r.enqueue(OpenFilesEvent{paths, false}) {
		t.Fatal("warm refusal")
	}
	paths[0] = "changed"
	time.Sleep(10 * time.Millisecond)
	if calls.Load() != 0 {
		t.Fatal("warm delivered before native readiness")
	}
	r.markReady()
	select {
	case e := <-events:
		if e.Paths[0] != want {
			t.Fatal("aliased request")
		}
	case <-time.After(time.Second):
		t.Fatal("owner delivery absent")
	}
	r.close()
	if r.enqueue(OpenFilesEvent{[]string{filepath.Join(directory, "late")}, false}) {
		t.Fatal("late request accepted")
	}
	q := &openFilesReceiver{owner: a, callback: func(OpenFilesEvent) {}}
	for range 16 {
		if !q.enqueue(OpenFilesEvent{[]string{filepath.Join(directory, "pending")}, false}) {
			t.Fatal("early capacity")
		}
	}
	if q.enqueue(OpenFilesEvent{[]string{filepath.Join(directory, "over")}, false}) {
		t.Fatal("unbounded queue")
	}
	q.close()
	unsafeReceiver := &openFilesReceiver{owner: a, callback: func(OpenFilesEvent) {}}
	for _, p := range []string{"relative", filepath.Join(directory, "invalid") + "\x00"} {
		if unsafeReceiver.enqueue(OpenFilesEvent{[]string{p}, true}) {
			t.Fatal("unsafe path")
		}
	}
}
func TestOpenFilesInitialDeliveryBeforeReadiness(t *testing.T) {
	a, _ := app.New(app.Config{})
	defer a.Close()
	done := make(chan bool, 1)
	r := &openFilesReceiver{owner: a, callback: func(e OpenFilesEvent) { done <- e.Initial }}
	if !r.enqueue(OpenFilesEvent{[]string{filepath.Join(t.TempDir(), "start.vforge")}, true}) {
		t.Fatal("cold refusal")
	}
	select {
	case initial := <-done:
		if !initial {
			t.Fatal("initial flag lost")
		}
	case <-time.After(time.Second):
		t.Fatal("initial delivery absent")
	}
	r.close()
}
