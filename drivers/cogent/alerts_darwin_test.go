//go:build darwin && cgo

package cogent

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"testing"
)

func TestNativeAlertRecipeWithoutPresentation(t *testing.T) {
	valid := `{"Title":"Confirm","Message":"No UI","Buttons":[{"ID":"ok","Label":"Keep","Role":"default"},{"ID":"cancel","Label":"Cancel","Role":"cancel"},{"ID":"delete","Label":"Discard","Role":"destructive"}]}`
	if nativeAlertRecipeFixture(valid) == "" {
		t.Fatal("production recipe rejected")
	}
	for _, wire := range []string{`[]`, `{}`, `{"Title":"T","Message":"M","Buttons":[]}`} {
		if nativeAlertRecipeFixture(wire) != "" {
			t.Fatal("malformed recipe admitted")
		}
	}
}
func TestNativeAlertWindowLifetimeRevokesOperationOffscreen(t *testing.T) {
	withMenuCommitState(t)
	window, cancelWindow := context.WithCancel(context.Background())
	joined, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := context.AfterFunc(window, cancel)
	defer stop()
	ctx := alertWindowContext{Context: joined, window: window}
	_, err := nativeOperationRun(ctx, func(token uint64) {
		cancelWindow()
		if nativeServiceActive(token) {
			t.Fatal("closed parent still active")
		}
		nativeServiceFinish(token, nativeResult{value: `{"ButtonID":"ok","Suppressed":false}`})
	}, func(uint64) {})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("retired parent admitted result", err)
	}
	nativeCallbacks.Lock()
	remaining := len(nativeCallbacks.operations)
	nativeCallbacks.Unlock()
	if remaining != 0 {
		t.Fatal("operation leaked")
	}
}
func TestNativeAlertBusyReply(t *testing.T) {
	withMenuCommitState(t)
	token := serviceToken.Add(1)
	result := make(chan nativeResult, 1)
	nativeCallbacks.Lock()
	nativeCallbacks.operations[token] = nativeOperation{ctx: context.Background(), result: result}
	nativeCallbacks.Unlock()
	nativeServiceResultFixture(token, nil, 5)
	if !errors.Is((<-result).err, app.ErrServiceBusy) {
		t.Fatal("native busy status lost")
	}
}

func TestNativeAlertClosedParentReply(t *testing.T) {
	withMenuCommitState(t)
	token := serviceToken.Add(1)
	result := make(chan nativeResult, 1)
	nativeCallbacks.Lock()
	nativeCallbacks.operations[token] = nativeOperation{ctx: context.Background(), result: result}
	nativeCallbacks.Unlock()
	nativeServiceResultFixture(token, nil, 6)
	if !errors.Is((<-result).err, app.ErrWindowClosed) {
		t.Fatal("closed parent status lost")
	}
}
