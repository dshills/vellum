//go:build darwin && cgo

package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/system/driver/base"
	"cogentcore.org/core/system/driver/desktop"
	"context"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent/internal/platformio"
	"github.com/dshills/vellum/drivers/cogent/internal/textinput"
	"math"
	"time"
)

type cocoaTextInput struct {
	input  *textinput.Input
	closed bool
} // Cocoa main only
func (b *cocoaTextInput) Publish(client uint64, s core.TextFieldEditingState, start, end int, active bool, ack uint64, rect [4]float64, g textinput.Geometry, maximum int) (bool, error) {
	if b.closed {
		return false, app.ErrWindowClosed
	}
	a, _ := textinput.UTF16Offset(s.Text, s.Anchor)
	f, _ := textinput.UTF16Offset(s.Text, s.Focus)
	b.input.RequestLimit(client, maximum)
	// An absent layout stays a geometry client with no cells: a stale/missing
	// range must remain unavailable rather than claim a fabricated caret match.
	if g.Viewport[3] <= 0 {
		g.Viewport = [4]float64{0, 0, 0, 1}
		g.FocusLine = -1
	}
	return b.input.UpdateGeometryForClient(client, s.Text, a, f, start, end, active, ack, rect, g)
}
func (b *cocoaTextInput) Settle(client uint64) bool { return !b.closed && b.input.Settle(client) }
func (b *cocoaTextInput) Close() {
	if !b.closed {
		b.closed = true
		b.input.Close()
	}
}
func attachTextInput(a *app.Application, w *desktop.Window) (*textWindow, error) {
	t, err := newTextWindow(a, &w.Events().Deque)
	if err != nil {
		return nil, err
	}
	input := textinput.Attach(w.Glw.GetCocoaView(), t.preflightInput, func() {
		if err := t.commitInput(); err != nil {
			panic(err)
		}
	})
	if input == nil {
		t.pump.Close()
		return nil, fmt.Errorf("cogent: native text input attachment failed")
	}
	input.EnableDeltas()
	t.native = &cocoaTextInput{input: input}
	return t, nil
}
func nativeTextAttachment(w *nativeWindow) *textWindow {
	if w == nil || w.body == nil || w.body.Scene == nil || w.body.Scene.RenderWindow() == nil {
		return nil
	}
	native, ok := w.body.Scene.RenderWindow().SystemWindow.(*desktop.Window)
	if !ok {
		return nil
	}
	attachment, ok := native.Native.(*accessibilityAttachment)
	if !ok {
		return nil
	}
	return attachment.text
}
func nativeTextScale(w *nativeWindow) float64 {
	if w != nil && w.body.Scene.RenderWindow() != nil {
		if n, ok := w.body.Scene.RenderWindow().SystemWindow.(*desktop.Window); ok {
			n.Mu.Lock()
			ratio := n.DevicePixelRatio
			n.Mu.Unlock()
			if ratio > 0 {
				return float64(ratio)
			}
		}
	}
	return 1
}
func publishTextOnMain(t *textWindow, client uint64, s core.TextFieldEditingState, start, end int, active bool, ack uint64, rect [4]float64, g textinput.Geometry, maximum int) (bool, error) {
	if !t.nativeReady.Load() {
		return false, nil
	}
	var accepted bool
	var err error
	if !desktop.TheApp.RunOnMainIfOpen(func() { accepted, err = t.native.Publish(client, s, start, end, active, ack, rect, g, maximum) }) {
		return false, app.ErrClosed
	}
	return accepted, err
}
func textMain[T any](ctx context.Context, t *textWindow, action func() (T, error)) (T, error) {
	if ctx == nil {
		var zero T
		return zero, fmt.Errorf("cogent: query requires context")
	}
	if t.closed.Load() {
		var zero T
		return zero, app.ErrWindowClosed
	}
	if !t.nativeReady.Load() {
		var zero T
		return zero, app.ErrUnavailable
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	queue, stopped := desktop.TheApp.InitMainQueue()
	return platformio.Dispatch(ctx, stopped, func(ctx context.Context, f func()) error {
		select {
		case queue <- base.FuncRun{F: f}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-stopped:
			return app.ErrClosed
		}
	}, func() (T, error) {
		var zero T
		if t.closed.Load() {
			return zero, app.ErrWindowClosed
		}
		return action()
	})
}
func settleTextOnMain(ctx context.Context, t *textWindow, client, id uint64) error {
	_, err := textMain(ctx, t, func() (bool, error) {
		if !t.acceptsNativeInput() {
			return false, app.ErrWindowClosed
		}
		if t.native == nil || !t.native.Settle(client) {
			return false, ErrTextInputStale
		}
		// Both unmark's preflight/commit and acknowledgment are admitted on this
		// main callback, directly to the originating FIFO, never Application.Post.
		return true, t.pump.Send(textMessage{Kind: "settle", Request: id})
	})
	return err
}

// InspectTextInput is an in-process Cocoa protocol observation, not live IME or
// external accessibility qualification. Call these blocking probes in a worker.
func InspectTextInput(ctx context.Context, window string) (string, error) {
	return ProbeTextInput(ctx, window, 0)
}

// ProbeTextInput performs deterministic native selectors. Steps 1..10 are the
// extracted protocol fixtures; 0 only inspects. It changes no OS input settings.
func ProbeTextInput(ctx context.Context, window string, step int) (string, error) {
	if ctx == nil || step < 0 || step > 10 {
		return "", fmt.Errorf("cogent: invalid input probe")
	}
	t := lookupTextWindow(window)
	if t == nil {
		return "", app.ErrWindowClosed
	}
	return textMain(ctx, t, func() (string, error) {
		b, ok := t.native.(*cocoaTextInput)
		if !ok || b.closed {
			return "", app.ErrUnavailable
		}
		value := b.input.Probe(step)
		if value == "" {
			return "", app.ErrUnavailable
		}
		return value, nil
	})
}
func QueryTextInput(ctx context.Context, window string, start, length int, point [2]float64) (string, error) {
	if ctx == nil || start < 0 || length < 0 || start > math.MaxInt32 || length > math.MaxInt32 || math.IsNaN(point[0]) || math.IsNaN(point[1]) || math.IsInf(point[0], 0) || math.IsInf(point[1], 0) {
		return "", fmt.Errorf("cogent: invalid text query")
	}
	t := lookupTextWindow(window)
	if t == nil {
		return "", app.ErrWindowClosed
	}
	return textMain(ctx, t, func() (string, error) {
		b, ok := t.native.(*cocoaTextInput)
		if !ok || b.closed {
			return "", app.ErrUnavailable
		}
		value := b.input.QueryProbe(start, length, point)
		if value == "" {
			return "", app.ErrUnavailable
		}
		return value, nil
	})
}

func startNativeTextPublication(t *textWindow) {
	t.awaitNativeTextMain(func(f func()) { desktop.TheApp.RunOnMainQueued(f) })
}
