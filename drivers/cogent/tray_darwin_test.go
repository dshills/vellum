//go:build darwin && cgo

package cogent

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"strings"
	"testing"
	"time"
)

func TestTrayNativeDecoderBitmapAndProducerOffscreen(t *testing.T) {
	desc := nativeTray{Title: "Vellum", AccessibilityLabel: "Vellum status", Tooltip: "Résumé 😀", Template: true, Items: []nativeTrayItem{{Label: "Show", Index: 0}, {Separator: true, Index: -1, Disabled: true}}}
	if err := trayPixels(&desc, trayTestPNG(t)); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(desc)
	if !nativeTrayValidate(data) {
		t.Fatal("native tray payload rejected")
	}
	if got := nativeTrayImageFixture(data); got != 7 {
		t.Fatal("bitmap/alpha/aspect fixture", got)
	}
	if got := nativeServicePayloadFixture(data); got != 7 {
		t.Fatal("producer lifetime", got)
	}
	clear, _ := json.Marshal(nativeTray{Remove: true})
	if !nativeTrayValidate(clear) {
		t.Fatal("clear payload rejected")
	}
	for _, bad := range []nativeTray{{}, {Title: "bad\x00"}, {Title: "Vellum", Width: 1, Height: 1, RGBA: []byte{1, 2, 3}}, {Title: "Vellum", Template: true}, {Title: "Vellum", Items: []nativeTrayItem{{Label: "Show", Index: 1}}}, {Remove: true, Title: "extra"}} {
		if !bad.Remove {
			bad.AccessibilityLabel = "Vellum status"
		}
		wire, _ := json.Marshal(bad)
		if nativeTrayValidate(wire) {
			t.Fatal("bad native metadata accepted", bad.Title)
		}
	}
	missingName, _ := json.Marshal(nativeTray{Title: "Vellum"})
	if nativeTrayValidate(missingName) {
		t.Fatal("missing native accessible name")
	}
	for _, bad := range [][]byte{append(append([]byte(nil), data...), []byte(" trailing")...), []byte(`{"Title":1}`), data[:len(data)-1]} {
		if nativeTrayValidate(bad) {
			t.Fatal("bad native envelope")
		}
	}
}

func TestNativeServiceCancellationRetainsAdmissionAndAcknowledges(t *testing.T) {
	withMenuCommitState(t)
	tokens := make([]uint64, nativeServicePendingMax)
	cancellations := 0
	for i := range tokens {
		ctx, cancel := context.WithCancel(context.Background())
		_, err := nativeOperationRun(ctx, func(token uint64) { tokens[i] = token; cancel() }, func(uint64) { cancellations++ })
		if !errors.Is(err, context.Canceled) || nativeServiceActive(tokens[i]) {
			t.Fatal(err)
		}
	}
	if cancellations != nativeServicePendingMax {
		t.Fatal(cancellations)
	}
	if _, err := nativeOperationRun(context.Background(), func(uint64) { t.Fatal("over-capacity dispatched") }, nil); !errors.Is(err, app.ErrServiceBusy) {
		t.Fatal(err)
	}
	nativeServiceFinish(tokens[0], nativeResult{})
	value, err := nativeOperationRun(context.Background(), func(token uint64) { nativeServiceFinish(token, nativeResult{value: "ok"}) }, nil)
	if err != nil || value != "ok" {
		t.Fatal(value, err)
	}
	for _, token := range tokens {
		nativeServiceFinish(token, nativeResult{})
	}
	nativeCallbacks.Lock()
	pending := len(nativeCallbacks.operations)
	nativeCallbacks.Unlock()
	if pending != 0 {
		t.Fatal(pending)
	}
}

func TestNativeServiceResultLimitsAndShutdown(t *testing.T) {
	withMenuCommitState(t)
	for _, tc := range []struct {
		value string
		want  error
	}{{strings.Repeat("x", nativeServiceResultMax+1), app.ErrServiceTooLarge}, {"\xff", app.ErrUnavailable}} {
		_, err := nativeOperationRun(context.Background(), func(token uint64) { nativeServiceResultFixture(token, []byte(tc.value), 0) }, nil)
		if !errors.Is(err, tc.want) {
			t.Fatal(err)
		}
	}
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := nativeOperationRun(context.Background(), func(uint64) { close(started) }, nil)
		done <- err
	}()
	<-started
	stopNativeCallbacks()
	select {
	case err := <-done:
		if !errors.Is(err, app.ErrClosed) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown stranded service")
	}
}

func TestTrayCommitReplacementRetirementAndAvailability(t *testing.T) {
	withMenuCommitState(t)
	enabled := true
	clicked := 0
	a, s := menuCommitService(t, command.Command{ID: "show", Label: "Show", Available: func() error {
		if !enabled {
			return errors.New("disabled")
		}
		return nil
	}, Execute: func(context.Context) error { clicked++; return nil }})
	reported := 0
	s.onError = func(error) { reported++ }
	register := func(ctx context.Context) uint64 {
		token := serviceToken.Add(1)
		nativeCallbacks.Lock()
		nativeCallbacks.operations[token] = nativeOperation{ctx: ctx, result: make(chan nativeResult, 1)}
		nativeCallbacks.trays[token] = nativeMenuCallback{ids: []string{"show"}, service: s}
		nativeCallbacks.Unlock()
		return token
	}
	old := register(context.Background())
	if !nativeTrayCommit(old, false) {
		t.Fatal("initial commit")
	}
	nativeServiceFinish(old, nativeResult{})
	ctx, cancel := context.WithCancel(context.Background())
	replacement := register(ctx)
	cancel()
	if nativeTrayCommit(replacement, false) {
		t.Fatal("canceled replacement committed")
	}
	nativeTrayCleanup(replacement)
	nativeServiceFinish(replacement, nativeResult{})
	nativeTrayAction(old, 0)
	if clicked != 0 {
		t.Fatal("executed outside owner")
	}
	a.Drain()
	if clicked != 1 {
		t.Fatal("previous callback lost")
	}
	nativeTrayAction(old, 0)
	newToken := register(context.Background())
	if !nativeTrayCommit(newToken, false) {
		t.Fatal("replacement failed")
	}
	nativeServiceFinish(newToken, nativeResult{})
	a.Drain()
	if clicked != 1 {
		t.Fatal("queued stale action ran")
	}
	a.Do(func() { enabled = false })
	nativeTrayAction(newToken, 0)
	a.Drain()
	if clicked != 1 || reported != 1 {
		t.Fatal("disabled command ran")
	}
	// Availability errors normally reach onError; use a quiet handler for refusal.
	s.onError = nil
	a.Do(func() { enabled = true })
	nativeTrayAction(newToken, 0)
	a.Drain()
	if clicked != 2 {
		t.Fatal("new callback lost")
	}
	clear := register(context.Background())
	if !nativeTrayCommit(clear, true) {
		t.Fatal("clear refused")
	}
	nativeServiceFinish(clear, nativeResult{})
	nativeTrayAction(newToken, 0)
	a.Drain()
	if clicked != 2 {
		t.Fatal("cleared callback active")
	}
}

func TestTrayCorruptPNGRefusedBeforeCocoaDispatch(t *testing.T) {
	withMenuCommitState(t)
	a, _ := app.New(app.Config{})
	defer a.Close()
	s := &platformServices{owner: a}
	bad := trayTestPNG(t)
	bad[len(bad)-1] ^= 1
	if err := s.SetTray(context.Background(), app.TrayOptions{Title: "Vellum", IconPNG: bad}); err == nil {
		t.Fatal("corrupt PNG accepted")
	}
	nativeCallbacks.Lock()
	pending, trays := len(nativeCallbacks.operations), len(nativeCallbacks.trays)
	nativeCallbacks.Unlock()
	if pending != 0 || trays != 0 {
		t.Fatal("refusal retained service state", pending, trays)
	}
}
