package cogent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/app"
	"io"
)

type alertWindowContext struct {
	context.Context
	window context.Context
}

func (c alertWindowContext) Err() error {
	if err := c.window.Err(); err != nil {
		return err
	}
	return c.Context.Err()
}
func (s *platformServices) ShowAlert(ctx context.Context, parent app.WindowRef, o app.AlertOptions) (app.AlertResult, error) {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return app.AlertResult{}, err
	}
	defer cancel()
	if err = o.Validate(); err != nil {
		return app.AlertResult{}, err
	}
	owned := false
	for _, w := range s.owner.Windows() {
		owned = owned || w.Ref() == parent
	}
	if !owned || parent == (app.WindowRef{}) {
		return app.AlertResult{}, app.ErrInvalidWindowRef
	}
	if parent.Context().Err() != nil {
		return app.AlertResult{}, app.ErrWindowClosed
	}
	joined, end := context.WithCancel(scope)
	stop := context.AfterFunc(parent.Context(), end)
	defer func() { stop(); end() }()
	return nativeShowAlert(alertWindowContext{Context: joined, window: parent.Context()}, s.owner, parent, o)
}
func alertParentToken(owner *app.Application, parent app.WindowRef) (uint64, error) {
	accessibilityRegistry.Lock()
	bridges := make([]*accessibilityBridge, 0, len(accessibilityRegistry.bridges))
	for _, b := range accessibilityRegistry.bridges {
		bridges = append(bridges, b)
	}
	accessibilityRegistry.Unlock()
	for _, b := range bridges {
		if b.owner != owner || b.closed.Load() {
			continue
		}
		b.mu.Lock()
		same := b.windowRef == parent
		b.mu.Unlock()
		if same && parent != (app.WindowRef{}) && parent.Context().Err() == nil {
			return b.token, nil
		}
	}
	return 0, app.ErrUnavailable
}
func decodeAlertResult(wire string, o app.AlertOptions) (app.AlertResult, error) {
	var v struct {
		ButtonID   *string
		Suppressed *bool
	}
	d := json.NewDecoder(bytes.NewBufferString(wire))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || v.ButtonID == nil || v.Suppressed == nil {
		return app.AlertResult{}, fmt.Errorf("cogent: malformed alert reply")
	}
	var trailing any
	if d.Decode(&trailing) != io.EOF {
		return app.AlertResult{}, app.ErrUnavailable
	}
	for _, b := range o.Buttons {
		if b.ID == *v.ButtonID {
			if *v.Suppressed && o.SuppressionLabel == "" {
				return app.AlertResult{}, app.ErrUnavailable
			}
			return app.AlertResult{ButtonID: *v.ButtonID, Suppressed: *v.Suppressed}, nil
		}
	}
	return app.AlertResult{}, app.ErrUnavailable
}
