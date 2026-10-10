package cogent

import (
	"cogentcore.org/core/core"
	"context"
	"example.com/vellum/app"
	"example.com/vellum/command"
	"testing"
	"time"
)

func TestMenuConstructionDoesNotReenterOwner(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	registry := command.New()
	if err := registry.Register(command.Command{ID: "test", Label: "<literal>", Execute: func(context.Context) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- a.Do(func() {
			host := core.NewBody("headless")
			bar := &CommandBar{Owner: a, Registry: registry}
			widget := bar.Mount(host)
			widget.AsWidget().Update()
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("menu builder nested owner")
	}
}
