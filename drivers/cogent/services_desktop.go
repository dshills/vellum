//go:build (windows || (linux && !android)) && cgo

package cogent

import (
	"context"
	"fmt"

	"cogentcore.org/core/system/driver/base"
	"cogentcore.org/core/system/driver/desktop"
	"example.com/vellum/app"
	"example.com/vellum/cogent/internal/platformio"
	"github.com/go-gl/glfw/v3.4/glfw"
)

func nativeClipboardCall(ctx context.Context, action func() (string, error)) (string, error) {
	if err := nativeServiceContextErr(ctx); err != nil {
		return "", err
	}
	queue, stopped := desktop.TheApp.InitMainQueue()
	enqueue := func(ctx context.Context, run func()) error {
		select {
		case queue <- base.FuncRun{F: run}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-stopped:
			return app.ErrClosed
		}
	}
	return platformio.Dispatch(ctx, stopped, enqueue, func() (value string, err error) {
		if err = nativeServiceContextErr(ctx); err != nil {
			return
		}
		defer func() {
			if p := recover(); p != nil {
				if e, ok := p.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("GLFW clipboard: %v", p)
				}
			}
		}()
		return action()
	})
}

// The Go GLFW wrappers drain their Go error channel but can log/suppress native
// platform errors. GetError reads a separate C error slot on this same OS thread.
// Clear its previous error before the operation, then preserve this operation's
// native failure. A missing read text flavor is the documented empty text result.
func nativeClipboardOperation(reading bool, operation func() string, getError func() (glfw.ErrorCode, string)) (string, error) {
	_, _ = getError()
	value := operation()
	code, description := getError()
	if code == glfw.NoError {
		return value, nil
	}
	if reading && code == glfw.FormatUnavailable {
		return "", nil
	}
	return "", &glfw.Error{Code: code, Desc: description}
}
func nativeReadClipboard(ctx context.Context) (string, error) {
	return nativeClipboardCall(ctx, func() (string, error) {
		return nativeClipboardOperation(true, glfw.GetClipboardString, glfw.GetError)
	})
}
func nativeWriteClipboard(ctx context.Context, value string) error {
	_, err := nativeClipboardCall(ctx, func() (string, error) {
		return nativeClipboardOperation(false, func() string { glfw.SetClipboardString(value); return "" }, glfw.GetError)
	})
	return err
}
func nativeShutdown() { clearDesktopMenus() }
