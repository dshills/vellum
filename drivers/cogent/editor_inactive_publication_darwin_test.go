//go:build darwin && cgo

package cogent

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/system/driver/desktop"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
)

// The existing main-dispatch backend seam creates no native view/window. Counts
// observe real publication dispatch; this is not physical IME qualification.
type inactivePublicationBackend struct {
	calls          int
	sequence       uint64
	refuse         bool
	err            error
	clients        []uint64
	active         []bool
	mainDispatched bool
	onPublish      func()
	onClose        func()
	closed         bool
}

func (b *inactivePublicationBackend) Publish(client uint64, state core.TextFieldEditingState, start, end int, active bool, ack uint64, rect [4]float64, geometry textinput.Geometry, maximum int) (bool, error) {
	b.calls++
	if b.closed {
		return false, app.ErrClosed
	}
	b.clients = append(b.clients, client)
	b.active = append(b.active, active)
	var pcs [64]uintptr
	n := runtime.Callers(1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, "system/driver/base.(*App).ExecuteMain") {
			b.mainDispatched = true
		}
		if !more {
			break
		}
	}
	if client == 0 && (active || state.Text != "" || state.Anchor != 0 || state.Focus != 0 || start != -1 || end != -1 || rect != [4]float64{} || len(geometry.Cells) != 0 || len(geometry.Carets) != 0 || maximum != textinput.MaxDocumentBytes) {
		return false, errors.New("changed empty inactive publication payload")
	}
	if b.onPublish != nil {
		b.onPublish()
	}
	return !b.refuse && ack >= b.sequence, b.err
}
func (*inactivePublicationBackend) Settle(uint64) bool { return false }
func (b *inactivePublicationBackend) Close() {
	b.closed = true
	if b.onClose != nil {
		b.onClose()
	}
}

func inactivePublicationFixture(t *testing.T) (*app.Application, *nativeWindow, *textWindow, *record, *inactivePublicationBackend) {
	t.Helper()
	a, w, tw, r, _ := editorFixture(t, "peer")
	backend := &inactivePublicationBackend{}
	if err := a.Do(func() {
		r.widget.AsWidget().SetState(false, states.Focused)
		tw.clients = map[uint64]*record{}
		tw.native = backend
		tw.nativeReady.Store(true)
		tw.setActiveClient(0, false)
	}); err != nil {
		t.Fatal(err)
	}
	return a, w, tw, r, backend
}

func TestEmptyInactivePublicationReusesActualAcceptedMainDispatch(t *testing.T) {
	a, _, tw, _, backend := inactivePublicationFixture(t)
	if err := a.Do(func() {
		for range 256 {
			tw.publish()
		}
	}); err != nil {
		t.Fatal(err)
	}
	if backend.calls != 1 || !backend.mainDispatched || tw.activeClient != 0 {
		t.Fatal("repeated empty publication reentered main or lost accepted activation", backend.calls, backend.mainDispatched)
	}
}

func TestEmptyInactivePublicationRefusalAndErrorsNeverCache(t *testing.T) {
	for _, failure := range []string{"refused", "error"} {
		t.Run(failure, func(t *testing.T) {
			a, _, tw, _, backend := inactivePublicationFixture(t)
			if err := a.Do(func() {
				if failure == "refused" {
					backend.refuse = true
				} else {
					backend.err = errors.New("native failed after apparent acceptance")
				}
				tw.publish()
				tw.publish()
				if backend.calls != 2 {
					t.Fatal("failed attempt was cached", backend.calls)
				}
				backend.refuse = false
				backend.err = nil
				tw.publish()
				tw.publish()
			}); err != nil {
				t.Fatal(err)
			}
			if backend.calls != 3 {
				t.Fatal("successful retry did not establish one reusable result", backend.calls)
			}
		})
	}
}

func TestEmptyInactivePublicationAckAndBackendIdentityInvalidate(t *testing.T) {
	a, _, tw, _, first := inactivePublicationFixture(t)
	second := &inactivePublicationBackend{}
	if err := a.Do(func() {
		tw.publish()
		tw.publish()
		tw.ack = 7
		tw.publish()
		tw.publish()
		if first.calls != 2 {
			t.Fatal("new acknowledgment bypassed native publication", first.calls)
		}
		tw.native = second
		tw.publish()
		tw.publish()
		tw.native = first
		tw.publish()
		tw.publish()
	}); err != nil {
		t.Fatal(err)
	}
	if first.calls != 3 || second.calls != 1 {
		t.Fatal("backend identity reused another backend's result", first.calls, second.calls)
	}
}

func TestEmptyInactivePublicationLifetimeReadinessAndCloseInvalidate(t *testing.T) {
	a, w, tw, _, backend := inactivePublicationFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := a.Do(func() {
		tw.publish()
		tw.publish()
		tw.nativeReady.Store(false)
		tw.publish()
		tw.nativeReady.Store(true)
		tw.publish()
		tw.publish()
		if backend.calls != 2 {
			t.Fatal("not-ready lifetime reused a cached result", backend.calls)
		}
		tw.lifetime.Store(&textLifetime{ctx: ctx})
		tw.publish()
		tw.publish()
		if backend.calls != 3 {
			t.Fatal("new lifetime reused prior lifetime's result", backend.calls)
		}
		cancel()
		tw.publish()
		if backend.calls != 3 {
			t.Fatal("canceled lifetime dispatched native publication")
		}
		tw.lifetime.Store(&textLifetime{ctx: context.Background()})
		tw.publish()
		tw.publish()
		closeTextWindow(w)
		tw.publish()
	}); err != nil {
		t.Fatal(err)
	}
	if backend.calls != 4 {
		t.Fatal("fresh/closed lifetime publication count", backend.calls)
	}
}

func TestEmptyInactivePublicationKeepsGoAdmissionRulesFresh(t *testing.T) {
	a, _, tw, client, backend := inactivePublicationFixture(t)
	if err := a.Do(func() {
		tw.publish()
		client.control.disabled = true
		tw.clients[client.control.client] = client
		tw.publish() // cache hit must still run setActiveClient/publishGridRules
		rules := tw.rules.Load()
		if rules.retiring != 0 || rules.clients[client.control.client].allowed {
			t.Fatal("cached publication retained stale enabled admission rules")
		}
		client.control.disabled = false
		tw.publish()
		if !tw.rules.Load().clients[client.control.client].allowed {
			t.Fatal("cached publication failed to refresh restored rules")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if backend.calls != 1 {
		t.Fatal("Go-only rules refresh needlessly reentered native main", backend.calls)
	}
}

func TestEmptyInactivePublicationActiveAttemptsAndExplicitDeactivationStayReal(t *testing.T) {
	a, _, tw, client, backend := inactivePublicationFixture(t)
	if err := a.Do(func() {
		tw.publish()
		tw.publish()
		backend.refuse = true
		if accepted, err := tw.publishClient(client, true); accepted || err != nil {
			t.Fatal("active refusal fixture", accepted, err)
		}
		backend.refuse = false
		tw.publish()
		tw.publish()
		if backend.calls != 3 {
			t.Fatal("refused active attempt did not invalidate empty cache", backend.calls)
		}
		if accepted, err := tw.publishClient(client, true); !accepted || err != nil {
			t.Fatal("active publication", accepted, err)
		}
		if accepted, err := tw.deactivateNativeClient(); !accepted || err != nil {
			t.Fatal("explicit handoff", accepted, err)
		}
		if backend.calls != 5 || backend.clients[3] != client.control.client || !backend.active[3] || backend.clients[4] != 0 || backend.active[4] {
			t.Fatal("activation/deactivation bypassed actual native handoff", backend.clients, backend.active)
		}
		tw.publish()
		tw.publish()
		if backend.calls != 6 {
			t.Fatal("deactivation result was confused with cached no-client branch", backend.calls)
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestEmptyInactivePublicationReservedAndCommittedInputCannotUseCache(t *testing.T) {
	a, w, tw, client, backend := inactivePublicationFixture(t)
	if err := a.Do(func() {
		tw.publish()
		tw.publish()
		backend.sequence = 1
		ticket, err := tw.pump.Preflight(textMessage{Kind: "input", Input: fixtureInput(client, 1, "unmark", "peer", "peer", 0, 0, -1, -1)})
		if err != nil {
			t.Fatal(err)
		}
		tw.publish() // pending reservation: stale ack must reach native refusal
		if backend.calls != 2 || tw.ack != 0 {
			t.Fatal("reserved input was hidden by cached inactive state", backend.calls, tw.ack)
		}
		if err := ticket.Commit(); err != nil {
			t.Fatal(err)
		}
		tw.publish()
		if backend.calls != 3 {
			t.Fatal("committed input credit was ignored", backend.calls)
		}
		for range 16 {
			if tw.pump.Stats().Queue.Outstanding == 0 {
				break
			}
			e, ok := w.source.Load().Deque.PollEvent()
			if !ok {
				t.Fatal("committed FIFO event was lost")
			}
			if _, err := tw.pump.HandleEvent(e, func(message textMessage) {
				if message.Kind == "input" {
					tw.apply(message.Input)
				}
			}); err != nil {
				t.Fatal(err)
			}
		}
		if tw.ack != 1 || tw.pump.Stats().Queue.Outstanding != 0 {
			t.Fatal("ordered input acknowledgment was not preserved")
		}
		tw.publish()
		tw.publish()
	}); err != nil {
		t.Fatal(err)
	}
	if backend.calls != 4 {
		t.Fatal("acknowledged empty state did not resume bounded reuse", backend.calls)
	}
}

func TestEmptyInactivePublicationAcceptanceRevalidatesReadiness(t *testing.T) {
	a, _, tw, _, backend := inactivePublicationFixture(t)
	if err := a.Do(func() {
		backend.onPublish = func() { tw.nativeReady.Store(false) }
		tw.publish()
		backend.onPublish = nil
		tw.nativeReady.Store(true)
		tw.publish()
		tw.publish()
	}); err != nil {
		t.Fatal(err)
	}
	if backend.calls != 2 {
		t.Fatal("acceptance from an invalidated readiness epoch was cached", backend.calls)
	}
}

// An interface-bearing value whose dynamic member is a slice is not safely
// comparable even though reflect.Type.Comparable is true; stay uncached.
type noncomparableInactiveBackend struct {
	calls  *int
	values any
}

func (b noncomparableInactiveBackend) Publish(uint64, core.TextFieldEditingState, int, int, bool, uint64, [4]float64, textinput.Geometry, int) (bool, error) {
	(*b.calls)++
	return true, nil
}
func (noncomparableInactiveBackend) Settle(uint64) bool { return false }
func (noncomparableInactiveBackend) Close()             {}
func TestEmptyInactivePublicationNoncomparableBackendStaysUncached(t *testing.T) {
	a, _, tw, _, _ := inactivePublicationFixture(t)
	calls := 0
	if err := a.Do(func() {
		tw.native = noncomparableInactiveBackend{calls: &calls, values: []int{1}}
		tw.publish()
		tw.publish()
	}); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatal("noncomparable backend was incorrectly cached", calls)
	}
}

// Only this fresh App becomes terminal. The real desktop singleton is restored
// before fixture cleanup, and no main queue/OS event loop is started.
func TestEmptyInactivePublicationStoppedMainDoesNotPublishGoRules(t *testing.T) {
	a, _, tw, _, backend := inactivePublicationFixture(t)
	original := desktop.TheApp
	desktop.TheApp = &desktop.App{}
	defer func() { desktop.TheApp = original }()
	if err := a.Do(func() {
		tw.publish()
		before := tw.rules.Load()
		if before == nil || backend.calls != 1 {
			t.Fatal("initial accepted publication missing")
		}
		desktop.TheApp.StopMain()
		tw.publish()
		if tw.rules.Load() != before {
			t.Fatal("stopped native main accepted a cached Go publication")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

func TestEmptyInactivePublicationNativeAttachmentRetiresBeforeOwnerCleanup(t *testing.T) {
	a, _, tw, _, backend := inactivePublicationFixture(t)
	if err := a.Do(func() {
		tw.publish()
		before := tw.rules.Load()
		stopBeforeAX := new(int)
		backend.onClose = func() {
			// A fake backend marks itself closed before this callback. Exercise
			// the actual attachment retirement prefix, then stop before AXDetach's
			// real Cocoa-main dispatch. This constructs no native view or window.
			tw.publish()
			panic(stopBeforeAX)
		}
		func() {
			defer func() {
				if recovered := recover(); recovered != stopBeforeAX {
					t.Fatalf("retirement sentinel mismatch: %v", recovered)
				}
			}()
			(&accessibilityAttachment{text: tw}).Close()
			t.Fatal("retirement failed to reach backend close")
		}()
		backend.onClose = nil
		if !backend.closed || tw.closed.Load() || !tw.acceptsNativeInput() {
			t.Fatal("native-close-before-owner-cleanup fixture was not preserved")
		}
		tw.publish()
		if tw.rules.Load() != before {
			t.Fatal("retired native attachment accepted cached Go publication")
		}
	}); err != nil {
		t.Fatal(err)
	}
}
