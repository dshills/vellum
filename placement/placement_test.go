package placement_test

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/history"
	"github.com/dshills/vellum/placement"
)

var ctx = context.Background()

type component struct {
	Kind          string
	Width, Height float64
}

func fixture(t *testing.T, execute func(context.Context, placement.Request[component]) error) (*placement.Controller[component], *placement.Source, *placement.Target) {
	t.Helper()
	c, err := placement.New(placement.Config[component]{Items: []placement.Item[component]{{ID: "resistor", Label: "Resistor", Payload: component{"resistor", 60, 20}, Preview: canvas.Rect{Width: 60, Height: 20, Color: color.RGBA{40, 100, 200, 255}}}}, Command: execute})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	target, err := c.NewTarget(placement.TargetConfig[component]{ID: "drawing", Snap: func(r placement.Request[component]) (canvas.Point, error) {
		return canvas.Point{X: math.Round(r.Position.X/20) * 20, Y: math.Round(r.Position.Y/20) * 20}, nil
	}, Accept: func(r placement.Request[component]) error {
		if r.Position.X < 0 {
			return placement.ErrRejected
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	return c, c.Source("resistor"), target
}
func geometry(scale float64) placement.Geometry {
	return placement.Geometry{Window: "main", Origin: canvas.Point{X: -120 * scale, Y: -30 * scale}, Clip: canvas.Bounds{Min: canvas.Point{}, Max: canvas.Point{X: 400 * scale, Y: 300 * scale}}, Scale: scale, View: canvas.Viewport{Width: 600, Height: 400, Zoom: 2, Center: canvas.Point{X: 100, Y: 40}}}
}
func TestScrolledScaledAndZoomedCoordinates(t *testing.T) {
	for _, scale := range []float64{1, 1.5, 2, 3} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			g := geometry(scale)
			p, err := g.World(canvas.Point{X: 100 * scale, Y: 50 * scale})
			if err != nil || p != (canvas.Point{X: 60, Y: -20}) {
				t.Fatal("scrolled origin or display scale applied incorrectly", p, err)
			}
			for _, v := range []struct {
				zoom float64
				want canvas.Point
			}{{.5, canvas.Point{X: -60, Y: -200}}, {1, canvas.Point{X: 20, Y: -80}}, {4, canvas.Point{X: 80, Y: 10}}} {
				g.View.Zoom = v.zoom
				p, err = g.World(canvas.Point{X: 100 * scale, Y: 50 * scale})
				if err != nil || p != v.want {
					t.Fatal("pan/zoom conversion", p, err, v)
				}
			}
			if _, err = g.World(canvas.Point{X: -1, Y: 50}); !errors.Is(err, placement.ErrRejected) {
				t.Fatal("clipped point accepted", err)
			}
			if _, err = g.World(canvas.Point{X: 400 * scale, Y: 50}); !errors.Is(err, placement.ErrRejected) {
				t.Fatal("far edge accepted", err)
			}
		})
	}
}
func TestOneCommandUndoRedoSameObject(t *testing.T) {
	type object struct {
		ID        int
		Position  canvas.Point
		Component component
	}
	var objects []object
	next := 0
	journal, err := history.New(8)
	if err != nil {
		t.Fatal(err)
	}
	registry := command.New()
	var pending *placement.Request[component]
	err = registry.Register(command.Command{ID: "place", Label: "Place component", Available: func() error {
		if pending == nil {
			return placement.ErrRejected
		}
		return nil
	}, Execute: func(context.Context) error {
		next++
		o := object{next, pending.Position, pending.Payload}
		return journal.Apply(history.Change{Label: "Place " + o.Component.Kind, Apply: func() error {
			for _, old := range objects {
				if old.ID == o.ID {
					return errors.New("duplicate object")
				}
			}
			objects = append(objects, o)
			return nil
		}, Revert: func() error {
			if len(objects) != 1 || objects[0] != o {
				return errors.New("stale document")
			}
			objects = nil
			return nil
		}})
	}})
	if err != nil {
		t.Fatal(err)
	}
	c, source, target := fixture(t, func(ctx context.Context, r placement.Request[component]) error {
		pending = &r
		defer func() { pending = nil }()
		return registry.Invoke(ctx, "place")
	})
	drag, err := source.Begin("main", canvas.Point{X: 2, Y: 2}, 8)
	if err != nil {
		t.Fatal(err)
	}
	if err = drag.Move(target, geometry(2), canvas.Point{X: 5, Y: 5}); err != nil || drag.State().Dragging || target.State().Accepted {
		t.Fatal("preview before threshold", drag.State(), err)
	}
	if err = drag.Move(target, geometry(2), canvas.Point{X: 200, Y: 100}); err != nil || !drag.State().Accepted || !drag.State().Dragging || drag.State().Position != (canvas.Point{X: 60, Y: -20}) {
		t.Fatal("missing accepted preview", drag.State(), err)
	}
	if len(objects) != 0 {
		t.Fatal("preview committed model")
	}
	committed, err := drag.Release(ctx, target, geometry(2), canvas.Point{X: 200, Y: 100})
	if err != nil || !committed || len(objects) != 1 || next != 1 || journal.Snapshot().UndoCount != 1 {
		t.Fatal("placement command", committed, objects, err)
	}
	original := objects[0]
	if original.Component.Kind != "resistor" || original.Position != (canvas.Point{X: 60, Y: -20}) {
		t.Fatal("typed request corrupted", original)
	}
	if c.State().Active || target.State().Active || drag.State().Active {
		t.Fatal("completed drag retained previews")
	}
	if committed, err = drag.Release(ctx, target, geometry(2), canvas.Point{X: 200, Y: 100}); committed || !errors.Is(err, placement.ErrClosed) || next != 1 {
		t.Fatal("duplicate release committed", committed, err)
	}
	if err = journal.Undo(); err != nil || len(objects) != 0 {
		t.Fatal("one undo did not remove placement", objects, err)
	}
	if err = journal.Redo(); err != nil || len(objects) != 1 || objects[0] != original || next != 1 {
		t.Fatal("redo recreated identity", objects, err)
	}
}
func TestCancellationAndFinalPositionRevalidation(t *testing.T) {
	for _, variant := range []string{"escape", "source-removed", "target-removed", "outside", "rejected", "cross-window", "capture-loss", "window-close", "context"} {
		t.Run(variant, func(t *testing.T) {
			calls := 0
			c, source, target := fixture(t, func(context.Context, placement.Request[component]) error { calls++; return nil })
			drag, err := source.Begin("main", canvas.Point{}, 8)
			if err != nil {
				t.Fatal(err)
			}
			if err = drag.Move(target, geometry(1), canvas.Point{X: 100, Y: 50}); err != nil || !drag.State().Accepted {
				t.Fatal("fixture", err, drag.State())
			}
			g := geometry(1)
			p := canvas.Point{X: 100, Y: 50}
			releaseCtx := ctx
			switch variant {
			case "escape", "capture-loss":
				drag.Cancel()
			case "source-removed":
				source.Cancel()
			case "target-removed":
				target.Cancel()
			case "outside":
				p.X = 500
			case "rejected":
				g.View.Center.X = -100
			case "cross-window":
				g.Window = "other"
			case "window-close":
				c.Close()
			case "context":
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()
				releaseCtx = cancelCtx
			}
			committed, _ := drag.Release(releaseCtx, target, g, p)
			if committed || calls != 0 || c.State().Active || target.State().Active || drag.State().Active {
				t.Fatal("cancel/rejection left command or retained gesture", committed, calls, c.State(), drag.State())
			}
		})
	}
}
func TestKeyboardUsesSameSnapValidationAndCommand(t *testing.T) {
	var requests []placement.Request[component]
	_, source, target := fixture(t, func(_ context.Context, r placement.Request[component]) error {
		requests = append(requests, r)
		return nil
	})
	if err := source.PlaceKeyboard(ctx, target, canvas.Point{X: 21, Y: 31}); err != nil || len(requests) != 1 || requests[0].Position != (canvas.Point{X: 20, Y: 40}) {
		t.Fatal(requests, err)
	}
	if err := source.PlaceKeyboard(ctx, target, canvas.Point{X: -40}); !errors.Is(err, placement.ErrRejected) || len(requests) != 1 {
		t.Fatal("keyboard skipped rejection", requests, err)
	}
}
func TestReentrantValidationAndCommandRefuse(t *testing.T) {
	var source *placement.Source
	var target *placement.Target
	var nested error
	c, err := placement.New(placement.Config[int]{Items: []placement.Item[int]{{ID: "x", Label: "X", Preview: canvas.Rect{Width: 1, Height: 1}}}, Command: func(context.Context, placement.Request[int]) error {
		nested = source.PlaceKeyboard(ctx, target, canvas.Point{})
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	source = c.Source("x")
	target, err = c.NewTarget(placement.TargetConfig[int]{ID: "canvas", Accept: func(placement.Request[int]) error {
		if e := source.PlaceKeyboard(ctx, target, canvas.Point{}); !errors.Is(e, placement.ErrBusy) {
			t.Fatal("reentrant validation admitted", e)
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err = source.PlaceKeyboard(ctx, target, canvas.Point{}); err != nil || !errors.Is(nested, placement.ErrBusy) {
		t.Fatal("reentrant command admitted", nested, err)
	}
}
func TestCloseDuringAcceptanceCannotResurrectGesture(t *testing.T) {
	calls := 0
	c, source, _ := fixture(t, func(context.Context, placement.Request[component]) error { calls++; return nil })
	target, err := c.NewTarget(placement.TargetConfig[component]{ID: "closing", Accept: func(placement.Request[component]) error { c.Close(); return nil }})
	if err != nil {
		t.Fatal(err)
	}
	drag, err := source.Begin("main", canvas.Point{}, 8)
	if err != nil {
		t.Fatal(err)
	}
	_, err = drag.Release(ctx, target, geometry(1), canvas.Point{X: 100, Y: 50})
	if !errors.Is(err, placement.ErrClosed) || calls != 0 || drag.State().Active || source.Valid() || target.Valid() {
		t.Fatal("closure resurrected gesture", calls, drag.State(), err)
	}
}

func TestCopiedAndZeroHandlesCannotMutateOrDoubleCommit(t *testing.T) {
	calls := 0
	c, source, target := fixture(t, func(context.Context, placement.Request[component]) error { calls++; return nil })
	copiedController := *c
	copiedController.Close()
	if copiedController.Source("resistor") != nil || !source.Valid() {
		t.Fatal("copied controller changed original")
	}
	copiedSource := *source
	if copiedSource.Valid() {
		t.Fatal("copied source admitted")
	}
	copiedSource.Cancel()
	copiedTarget := *target
	if copiedTarget.Valid() {
		t.Fatal("copied target admitted")
	}
	var zero placement.Gesture
	if err := zero.Move(target, geometry(1), canvas.Point{X: 100, Y: 50}); !errors.Is(err, placement.ErrClosed) {
		t.Fatal("zero gesture admitted", err)
	}
	drag, err := source.Begin("main", canvas.Point{}, 8)
	if err != nil {
		t.Fatal(err)
	}
	copyDrag := *drag
	if committed, err := copyDrag.Release(ctx, target, geometry(1), canvas.Point{X: 100, Y: 50}); committed || !errors.Is(err, placement.ErrClosed) || calls != 0 {
		t.Fatal("copied gesture committed", committed, err)
	}
	copyDrag.Cancel()
	if !c.State().Active {
		t.Fatal("copied gesture canceled original")
	}
	if committed, err := drag.Release(ctx, target, geometry(1), canvas.Point{X: 100, Y: 50}); !committed || err != nil || calls != 1 {
		t.Fatal("original gesture corrupted", committed, calls, err)
	}
}
func TestContextCanceledDuringValidationNeverCommands(t *testing.T) {
	for _, keyboard := range []bool{false, true} {
		t.Run(fmt.Sprint(keyboard), func(t *testing.T) {
			calls := 0
			c, source, _ := fixture(t, func(context.Context, placement.Request[component]) error { calls++; return nil })
			life, cancel := context.WithCancel(ctx)
			defer cancel()
			target, err := c.NewTarget(placement.TargetConfig[component]{ID: "canceling", Accept: func(placement.Request[component]) error { cancel(); return nil }})
			if err != nil {
				t.Fatal(err)
			}
			if keyboard {
				err = source.PlaceKeyboard(life, target, canvas.Point{})
			} else {
				drag, e := source.Begin("main", canvas.Point{}, 8)
				if e != nil {
					t.Fatal(e)
				}
				_, err = drag.Release(life, target, geometry(1), canvas.Point{X: 100, Y: 50})
			}
			if !errors.Is(err, context.Canceled) || calls != 0 || c.State().Active {
				t.Fatal("canceled validation commanded", calls, err, c.State())
			}
		})
	}
}
func TestGestureMutationFromAcceptRefusesBeforeChangingState(t *testing.T) {
	c, source, _ := fixture(t, func(context.Context, placement.Request[component]) error { return nil })
	var drag *placement.Gesture
	target, err := c.NewTarget(placement.TargetConfig[component]{ID: "reentrant", Accept: func(placement.Request[component]) error {
		before := drag.State()
		if err := drag.Move(nil, placement.Geometry{}, canvas.Point{}); !errors.Is(err, placement.ErrBusy) {
			t.Fatal("nested move admitted", err)
		}
		if committed, err := drag.Release(ctx, nil, placement.Geometry{}, canvas.Point{}); committed || !errors.Is(err, placement.ErrBusy) {
			t.Fatal("nested release admitted", committed, err)
		}
		if drag.State() != before {
			t.Fatal("refused mutation changed outer preview")
		}
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	drag, err = source.Begin("main", canvas.Point{}, 8)
	if err != nil {
		t.Fatal(err)
	}
	if err = drag.Move(target, geometry(1), canvas.Point{X: 100, Y: 50}); err != nil || !drag.State().Accepted {
		t.Fatal("outer move corrupted", drag.State(), err)
	}
	if committed, err := drag.Release(ctx, target, geometry(1), canvas.Point{X: 100, Y: 50}); !committed || err != nil {
		t.Fatal("outer release corrupted", committed, err)
	}
}
func TestSnapRetirementStopsAcceptance(t *testing.T) {
	for _, variant := range []string{"close-pointer", "cancel-pointer", "close-keyboard"} {
		t.Run(variant, func(t *testing.T) {
			accepted, commanded := 0, 0
			c, source, _ := fixture(t, func(context.Context, placement.Request[component]) error { commanded++; return nil })
			var drag *placement.Gesture
			target, err := c.NewTarget(placement.TargetConfig[component]{ID: "retiring", Snap: func(r placement.Request[component]) (canvas.Point, error) {
				if variant == "cancel-pointer" {
					drag.Cancel()
				} else {
					c.Close()
				}
				return r.Position, nil
			}, Accept: func(placement.Request[component]) error { accepted++; return nil }})
			if err != nil {
				t.Fatal(err)
			}
			if variant == "close-keyboard" {
				err = source.PlaceKeyboard(ctx, target, canvas.Point{})
			} else {
				drag, err = source.Begin("main", canvas.Point{}, 8)
				if err != nil {
					t.Fatal(err)
				}
				_, err = drag.Release(ctx, target, geometry(1), canvas.Point{X: 100, Y: 50})
			}
			if !errors.Is(err, placement.ErrClosed) || accepted != 0 || commanded != 0 || c.State().Active {
				t.Fatal("retired snap reached later callbacks", accepted, commanded, err)
			}
		})
	}
}
