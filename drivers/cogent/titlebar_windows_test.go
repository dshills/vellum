package cogent

import (
	"context"
	"errors"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
)

func TestTitlebarPerWindowCommandsAvailabilityAndRetirement(t *testing.T) {
	a, original, _ := prototypeFixture(t)
	second, err := a.OpenWindow(scopeDeclaration("second"))
	if err != nil {
		t.Fatal(err)
	}
	enabled := false
	invoked := 0
	registry := command.New()
	if err := registry.Register(command.Command{ID: "save", Label: "Save", Available: func() error {
		if !enabled {
			return errors.New("not dirty")
		}
		return nil
	}, Execute: func(context.Context) error { invoked++; return nil }}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	factory := func(ref app.WindowRef) (*TitlebarToolbar, error) {
		calls++
		return &TitlebarToolbar{Window: ref, Items: []TitlebarToolbarItem{{ID: "save", Command: "save", Kind: original.Items[0].Kind, PNG: original.Items[0].PNG}}}, nil
	}
	var firstConfig, secondConfig *preparedTitlebarToolbar
	if err := a.Do(func() {
		firstConfig, err = preparePerWindowTitlebar(a, original.Window, factory, registry)
		if err == nil {
			secondConfig, err = preparePerWindowTitlebar(a, second, factory, registry)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err != nil || calls != 2 || firstConfig.ref == secondConfig.ref {
		t.Fatal("window capabilities not independent", err, calls)
	}
	first, other := newTitlebarToolbarBinding(a, firstConfig), newTitlebarToolbarBinding(a, secondConfig)
	first.setAllowed(true)
	other.setAllowed(true)
	if first.deliver("save") || other.deliver("save") {
		t.Fatal("disabled registry command admitted")
	}
	enabled = true
	_ = a.Do(func() { first.refreshAvailability(); other.refreshAvailability() })
	if !first.deliver("save") {
		t.Fatal("enabled command refused")
	}
	_ = a.Drain()
	if invoked != 1 {
		t.Fatal("same command path not used", invoked)
	}
	// Availability ABA revokes the queued capability even if re-enabled later.
	if !first.deliver("save") {
		t.Fatal("queue")
	}
	enabled = false
	_ = a.Do(first.refreshAvailability)
	enabled = true
	_ = a.Do(first.refreshAvailability)
	_ = a.Drain()
	if invoked != 1 {
		t.Fatal("stale activation crossed availability ABA")
	}
	// Availability is rechecked even if an app forgot to invalidate its UI.
	if !first.deliver("save") {
		t.Fatal("queue")
	}
	enabled = false
	_ = a.Drain()
	if invoked != 1 {
		t.Fatal("live command refusal bypassed")
	}
	enabled = true
	first.close()
	if first.deliver("save") || !other.deliver("save") {
		t.Fatal("one window close affected another")
	}
	_ = a.Drain()
	if invoked != 2 {
		t.Fatal("surviving window command lost")
	}
	if err := a.CloseWindow(second); err != nil {
		t.Fatal(err)
	}
	reopened, err := a.OpenWindow(scopeDeclaration("second"))
	if err != nil {
		t.Fatal(err)
	}
	if other.deliver("save") {
		t.Fatal("retired opening admitted action")
	}
	var renewed *preparedTitlebarToolbar
	_ = a.Do(func() { renewed, err = preparePerWindowTitlebar(a, reopened, factory, registry) })
	if err != nil || renewed.ref == secondConfig.ref {
		t.Fatal("reopened window reused capability", err)
	}
}

func TestTitlebarFactoryRejectsForeignOpeningAndAmbiguousActions(t *testing.T) {
	a, original, _ := prototypeFixture(t)
	second, err := a.OpenWindow(scopeDeclaration("other"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := preparePerWindowTitlebar(a, second, func(app.WindowRef) (*TitlebarToolbar, error) { return original, nil }, nil); !errors.Is(err, app.ErrInvalidWindowRef) {
		t.Fatal("foreign opening accepted", err)
	}
	registry := command.New()
	_ = registry.Register(command.Command{ID: "run", Label: "Run", Execute: func(context.Context) error { return nil }})
	original.Commands = registry
	original.Items[0].Command = "run"
	if _, err := prepareTitlebarToolbarScoped(a, original, false); err == nil {
		t.Fatal("Activate plus Command accepted")
	}
	original.Items[0].Activate = nil
	original.Items[0].Command = "missing"
	if _, err := prepareTitlebarToolbarScoped(a, original, false); err == nil {
		t.Fatal("missing command accepted")
	}
}

func TestTitlebarCommandErrorIsReportedWithoutClosingApplication(t *testing.T) {
	a, spec, _ := prototypeFixture(t)
	registry := command.New()
	rejected := errors.New("validation refused")
	_ = registry.Register(command.Command{ID: "save", Label: "Save", Execute: func(context.Context) error { return rejected }})
	spec.Commands = registry
	spec.Items = spec.Items[:1]
	spec.Items[0].Activate = nil
	spec.Items[0].Command = "save"
	p, err := prepareTitlebarToolbar(a, spec)
	if err != nil {
		t.Fatal(err)
	}
	binding := newTitlebarToolbarBinding(a, p)
	binding.setAllowed(true)
	var reported error
	binding.onError = func(err error) { reported = err }
	if !binding.deliver(spec.Items[0].ID) {
		t.Fatal("command not admitted")
	}
	_ = a.Drain()
	if !errors.Is(reported, rejected) || a.Context().Err() != nil {
		t.Fatal("command validation became application failure", reported, a.Context().Err())
	}
}

func TestTitlebarAvailabilityFlushKeepsBothPublicationOrders(t *testing.T) {
	b := &titlebarToolbarBinding{}
	old, next := &titlebarAvailability{json: "old"}, &titlebarAvailability{json: "next"}
	b.availability.Store(old)
	b.updateQueued.Store(true)
	// Publication while a flush is pending must appear in that flush.
	b.availability.Store(next)
	if b.updateQueued.CompareAndSwap(false, true) {
		t.Fatal("duplicate native flush queued")
	}
	if b.takeAvailability() != next || b.updateQueued.Load() {
		t.Fatal("pending publication lost")
	}
	// Publication after a flush load must retain a new native wake.
	b.availability.Store(old)
	b.updateQueued.Store(true)
	if b.takeAvailability() != old {
		t.Fatal("flush snapshot")
	}
	b.availability.Store(next)
	if !b.updateQueued.CompareAndSwap(false, true) {
		t.Fatal("follow-up wake lost")
	}
	if b.takeAvailability() != next {
		t.Fatal("follow-up state lost")
	}
}
