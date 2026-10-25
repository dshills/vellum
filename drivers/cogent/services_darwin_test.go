//go:build darwin && cgo

package cogent

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"testing"
	"time"
)

func withMenuCommitState(t *testing.T) {
	t.Helper()
	nativeCallbacks.Lock()
	previousStopped := nativeCallbacks.stopped
	nativeCallbacks.stopped = false
	previousOperations, previousMenus, previousToken := nativeCallbacks.operations, nativeCallbacks.menus, nativeCallbacks.menuToken
	previousTrays, previousTrayToken := nativeCallbacks.trays, nativeCallbacks.trayToken
	nativeCallbacks.operations = make(map[uint64]nativeOperation)
	nativeCallbacks.menus = make(map[uint64]nativeMenuCallback)
	nativeCallbacks.menuToken = 0
	nativeCallbacks.trays = make(map[uint64]nativeMenuCallback)
	nativeCallbacks.trayToken = 0
	nativeCallbacks.Unlock()
	t.Cleanup(func() {
		nativeCallbacks.Lock()
		nativeCallbacks.stopped = previousStopped
		nativeCallbacks.operations = previousOperations
		nativeCallbacks.menus = previousMenus
		nativeCallbacks.menuToken = previousToken
		nativeCallbacks.trays = previousTrays
		nativeCallbacks.trayToken = previousTrayToken
		nativeCallbacks.Unlock()
	})
}

// Use the production service path: native activation may only enqueue owner work.
func menuCommitService(t *testing.T, commands ...command.Command) (*app.Application, *platformServices) {
	t.Helper()
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	registry := command.New()
	for _, cmd := range commands {
		if err := registry.Register(cmd); err != nil {
			t.Fatal(err)
		}
	}
	return a, &platformServices{owner: a, commands: registry, onError: func(err error) { t.Error(err) }}
}

func TestNativeMenuCancellationBeforeCommitKeepsWorkingMenu(t *testing.T) {
	withMenuCommitState(t)
	oldToken, newToken := serviceToken.Add(1), serviceToken.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	clicked := 0
	a, service := menuCommitService(t,
		command.Command{ID: "old", Label: "Old", Execute: func(context.Context) error { clicked++; return nil }},
		command.Command{ID: "new", Label: "New", Execute: func(context.Context) error { t.Fatal("canceled replacement activated"); return nil }},
	)
	nativeCallbacks.Lock()
	nativeCallbacks.menuToken = oldToken
	nativeCallbacks.menus[oldToken] = nativeMenuCallback{ids: []string{"old"}, service: service}
	nativeCallbacks.operations[newToken] = nativeOperation{ctx: ctx, result: make(chan nativeResult, 1)}
	nativeCallbacks.menus[newToken] = nativeMenuCallback{ids: []string{"new"}, service: service}
	nativeCallbacks.Unlock()
	// The Cocoa block already passed its first activity check and built the menu;
	// its caller cancels before the final main-thread commit/ownership transfer.
	cancel()
	if nativeMenuCommit(newToken) {
		t.Fatal("canceled replacement accepted")
	}
	nativeMenuCleanup(newToken)
	nativeMenuAction(oldToken, 0)
	if clicked != 0 {
		t.Fatal("native callback executed command outside UI owner")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	nativeCallbacks.Lock()
	current := nativeCallbacks.menuToken
	_, pending := nativeCallbacks.menus[newToken]
	nativeCallbacks.Unlock()
	if current != oldToken || pending || clicked != 1 {
		t.Fatal("cancellation damaged the existing menu", current, pending, clicked)
	}
}

func TestNativeMenuCancellationAfterCommitKeepsInstalledCallback(t *testing.T) {
	withMenuCommitState(t)
	token := serviceToken.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	clicked := 0
	a, service := menuCommitService(t, command.Command{ID: "new", Label: "New", Execute: func(context.Context) error { clicked++; return nil }})
	nativeCallbacks.Lock()
	nativeCallbacks.operations[token] = nativeOperation{ctx: ctx, result: make(chan nativeResult, 1)}
	nativeCallbacks.menus[token] = nativeMenuCallback{ids: []string{"new"}, service: service}
	nativeCallbacks.Unlock()
	if !nativeMenuCommit(token) {
		t.Fatal("live replacement rejected")
	}
	// Once ownership was accepted, a caller timing out before native completion
	// must not delete the callback for the menu Cocoa is about to install.
	cancel()
	nativeCallbacks.Lock()
	delete(nativeCallbacks.operations, token)
	nativeCallbacks.Unlock()
	nativeMenuCleanup(token)
	nativeMenuAction(token, 0)
	if clicked != 0 {
		t.Fatal("native callback executed command outside UI owner")
	}
	if err := a.Drain(); err != nil {
		t.Fatal(err)
	}
	if clicked != 1 {
		t.Fatal("installed callback lost after caller cancellation")
	}
}

func TestNativeMenuCommitRejectsCompletedCanceledOperation(t *testing.T) {
	withMenuCommitState(t)
	token := serviceToken.Add(1)
	nativeCallbacks.Lock()
	nativeCallbacks.menus[token] = nativeMenuCallback{ids: []string{"new"}}
	// The caller's operation cleanup has already removed the pending operation.
	nativeCallbacks.Unlock()
	if nativeMenuCommit(token) {
		t.Fatal("orphan callback installed")
	}
	nativeMenuCleanup(token)
}

func TestNativeMenuQueuedActionRetiresBeforeOwnerDrain(t *testing.T) {
	for _, retirement := range []string{"replacement", "native-stop", "service-stop"} {
		t.Run(retirement, func(t *testing.T) {
			withMenuCommitState(t)
			token := serviceToken.Add(1)
			clicked := 0
			a, service := menuCommitService(t, command.Command{ID: "menu", Label: "Menu", Execute: func(context.Context) error { clicked++; return nil }})
			nativeCallbacks.Lock()
			nativeCallbacks.menuToken = token
			nativeCallbacks.menus[token] = nativeMenuCallback{ids: []string{"menu"}, service: service}
			nativeCallbacks.Unlock()
			nativeMenuAction(token, 0)
			switch retirement {
			case "replacement":
				nativeCallbacks.Lock()
				nativeCallbacks.menuToken = serviceToken.Add(1)
				delete(nativeCallbacks.menus, token)
				nativeCallbacks.Unlock()
			case "native-stop":
				stopNativeCallbacks()
			case "service-stop":
				service.stopped.Store(true)
			}
			if err := a.Drain(); err != nil {
				t.Fatal(err)
			}
			if clicked != 0 {
				t.Fatal("retired menu action executed", retirement)
			}
		})
	}
}

func TestNativeShutdownUnblocksUndispatchedBackgroundOperation(t *testing.T) {
	if !nativeServiceDismissFixture() {
		t.Fatal("native cancellation fence failed injected dismissal exceptions")
	}
	withMenuCommitState(t)
	dispatched := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		_, err := nativeOperationRun(context.Background(), func(uint64) { close(dispatched) }, nil)
		finished <- err
	}()
	<-dispatched
	stopNativeCallbacks()
	select {
	case err := <-finished:
		if !errors.Is(err, app.ErrClosed) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown stranded pending operation")
	}
	_, err := nativeOperationRun(context.Background(), func(uint64) { t.Fatal("dispatch after shutdown") }, nil)
	if !errors.Is(err, app.ErrClosed) {
		t.Fatal(err)
	}
	nativeCallbacks.Lock()
	token, pending, menus := nativeCallbacks.menuToken, len(nativeCallbacks.operations), len(nativeCallbacks.menus)
	nativeCallbacks.Unlock()
	if token != 0 || pending != 0 || menus != 0 {
		t.Fatal("shutdown retained native callback state")
	}
}

func TestNativeFilePanelRestorationDoesNotStealFocus(t *testing.T) {
	if !nativeServicePanelRestoreFixture() {
		t.Fatal("inactive, hidden, minimized or unrelated key windows must refuse restoration")
	}
}
