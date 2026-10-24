package cogent

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"testing"
)

func TestAlertReplyValidationAndForeignParent(t *testing.T) {
	o := app.AlertOptions{Title: "Confirm", Buttons: []app.AlertButton{{ID: "ok", Label: "OK", Role: app.AlertDefault}}}
	for _, wire := range []string{`{}`, `{"ButtonID":"foreign","Suppressed":false}`, `{"ButtonID":"ok","Suppressed":true}`, `{"ButtonID":"ok","Suppressed":false} {}`} {
		if _, err := decodeAlertResult(wire, o); err == nil {
			t.Fatal("invalid reply", wire)
		}
	}
	if result, err := decodeAlertResult(`{"ButtonID":"ok","Suppressed":false}`, o); err != nil || result.ButtonID != "ok" {
		t.Fatal(result, err)
	}
	a := dropApp(t)
	other := dropApp(t)
	ref, _ := other.WindowRef("main")
	s := &platformServices{owner: a}
	if _, err := s.ShowAlert(context.Background(), ref, o); !errors.Is(err, app.ErrInvalidWindowRef) {
		t.Fatal("foreign parent admitted", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.ShowAlert(ctx, ref, o); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
