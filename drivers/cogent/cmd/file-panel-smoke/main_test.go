//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
)

func TestPanelLifecycleEvidenceRejectsUnretiredAndWrongMode(t *testing.T) {
	empty := cogent.NativeServiceCheckState{MainThread: true}
	open := cogent.NativeServiceCheckState{MainThread: true, NativeOutstanding: 1, FilePanels: cogent.NativeServiceFilePanelCheck{Tracked: 1, Visible: 1, Open: 1}}
	if !panelMatches(empty, app.OpenFile, false) || !panelMatches(open, app.OpenFile, true) || panelMatches(open, app.SaveFile, true) || panelMatches(open, app.OpenFile, false) {
		t.Fatal("baseline lifecycle classification")
	}
	for _, field := range []string{"hidden", "untracked", "extra-panel", "no-ack", "still-admitted", "wrong-thread", "mixed-class", "negative-count"} {
		t.Run(field, func(t *testing.T) {
			s := open
			switch field {
			case "hidden":
				s.FilePanels.Visible = 0
			case "untracked":
				s.FilePanels.Tracked = 0
			case "extra-panel":
				s.FilePanels.Tracked = 2
			case "no-ack":
				s.NativeOutstanding = 0
			case "still-admitted":
				s.NativeOutstanding = 2
			case "wrong-thread":
				s.MainThread = false
			case "mixed-class":
				s.FilePanels.Save = 1
			case "negative-count":
				s.FilePanels.Save = -1
			}
			if panelMatches(s, app.OpenFile, true) {
				t.Fatalf("accepted incomplete/wrong active evidence: %+v", s)
			}
		})
	}
	for _, s := range []cogent.NativeServiceCheckState{{MainThread: true, NativeOutstanding: 1}, {MainThread: true, FilePanels: cogent.NativeServiceFilePanelCheck{Tracked: 1}}, {MainThread: true, FilePanels: cogent.NativeServiceFilePanelCheck{Visible: 1}}, {MainThread: true, FilePanels: cogent.NativeServiceFilePanelCheck{Save: 1}}, {}} {
		if panelMatches(s, app.SaveFile, false) {
			t.Fatalf("accepted nonterminal inventory: %+v", s)
		}
	}
	open.FilePanels.Open = 0
	open.FilePanels.Save = 1
	if !panelMatches(open, app.SaveFile, true) {
		t.Fatal("save class rejected")
	}
}
func TestConfigAndCallerJoinBoundaries(t *testing.T) {
	valid := config{cycles: 100, shutdown: "save", timeout: 5 * time.Minute}
	if err := valid.validate(); err != nil {
		t.Fatal(err)
	}
	for _, c := range []config{{0, "open", time.Minute}, {101, "open", time.Minute}, {1, "other", time.Minute}, {1, "open", time.Second}, {1, "save", 6 * time.Minute}} {
		if c.validate() == nil {
			t.Fatalf("invalid config accepted: %+v", c)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := receive(ctx, make(chan reply)); !errors.Is(err, context.Canceled) {
		t.Fatalf("join did not stop: %v", err)
	}
	ch := make(chan reply, 1)
	ch <- reply{err: app.ErrClosed}
	v, err := receive(context.Background(), ch)
	if err != nil || !errors.Is(v.err, app.ErrClosed) || v.path != "" {
		t.Fatalf("lost terminal caller reply: %+v %v", v, err)
	}
}

func TestInFlightCancellationIsDistinctFromNewAdmissionRefusal(t *testing.T) {
	if !canceledReply(reply{err: context.Canceled}) || !canceledReply(reply{err: errors.Join(context.Canceled, errors.New("detail"))}) {
		t.Fatal("in-flight cancellation refused")
	}
	for _, value := range []reply{{}, {err: app.ErrClosed}, {err: app.ErrCanceled}, {err: context.DeadlineExceeded}, {path: "/tmp/unexpected-selection", err: context.Canceled}} {
		if canceledReply(value) {
			t.Fatalf("accepted a selection or wrong lifetime boundary: %+v", value)
		}
	}
}
func TestPartialEvidenceSurvivesConcurrentSnapshots(t *testing.T) {
	r := &recorder{}
	var workers sync.WaitGroup
	for i := 0; i < 32; i++ {
		workers.Add(1)
		go func(i int) { defer workers.Done(); r.add(event{Stage: "observed", Cycle: i}); _ = r.snapshot() }(i)
	}
	workers.Wait()
	first := r.snapshot()
	if len(first) != 32 {
		t.Fatalf("lost partial progress: %d", len(first))
	}
	first[0].Stage = "mutated outer copy"
	if r.snapshot()[0].Stage != "observed" {
		t.Fatal("report shares mutable event slice")
	}
}
