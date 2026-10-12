package sdkcontract_test

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/automation"
	"github.com/dshills/vellum/ui"
)

func TestPublicAutomationConsumer(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Mac endpoint")
	}
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "document", Title: "Headless consumer", Width: 200, Height: 100, View: func() ui.Node { return ui.Node{ID: "root", Kind: ui.Label} }}); err != nil {
		t.Fatal(err)
	}
	revision := uint64(1)
	value := 10
	r := automation.NewRegistry(func() (uint64, error) { return revision, nil })
	type edit struct {
		Value int `json:"value"`
	}
	if err = automation.RegisterQuery(r, "property", func(context.Context, struct{}) (any, error) { return value, nil }); err != nil {
		t.Fatal(err)
	}
	if err = automation.RegisterAction(r, "property.set", func(ctx context.Context, p edit) (any, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if p.Value < 0 || p.Value > 100 {
			return nil, errors.New("outside range")
		}
		value = p.Value
		revision++
		return value, nil
	}); err != nil {
		t.Fatal(err)
	}
	s, err := automation.Start(automation.Config{Owner: a, Window: "document", Registry: r})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	wake := make(chan struct{}, 1)
	a.SetWake(func() {
		select {
		case wake <- struct{}{}:
		default:
		}
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-a.Context().Done():
				return
			case <-wake:
				a.Drain()
			}
		}
	}()
	defer func() {
		a.Close()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("owner shutdown")
		}
		select {
		case <-s.Done():
		case <-time.After(5 * time.Second):
			t.Error("transport shutdown")
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	read, err := automation.Call(ctx, s.Endpoint(), automation.Request{Version: 1, ID: "query", Operation: "inspect", Target: "property", Params: json.RawMessage(`{}`)})
	if err != nil || read.Error != nil {
		t.Fatal(read, err)
	}
	write, err := automation.Call(ctx, s.Endpoint(), automation.Request{Version: 1, ID: "edit", Operation: "invoke", Target: "property.set", ExpectedRevision: read.Revision, Params: json.RawMessage(`{"value":20}`)})
	if err != nil || write.Error != nil || string(write.Result) != "20" {
		t.Fatal(write, err)
	}
	refused, err := automation.Call(ctx, s.Endpoint(), automation.Request{Version: 1, ID: "stale", Operation: "invoke", Target: "property.set", ExpectedRevision: read.Revision, Params: json.RawMessage(`{"value":30}`)})
	if err != nil || refused.Error == nil || refused.Error.Code != "stale_revision" {
		t.Fatal(refused, err)
	}
}
