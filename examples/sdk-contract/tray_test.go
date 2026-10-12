package sdkcontract_test

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"testing"
)

// Public API consumption only; this adapter never creates an OS status item.
type consumerTray struct {
	memoryClipboard
	options app.TrayOptions
	present bool
}

func (*consumerTray) TrayCapabilities() app.TrayCapabilities {
	return app.TrayCapabilities{StatusItem: true, CommandMenu: true}
}
func (m *consumerTray) SetTray(ctx context.Context, o app.TrayOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := o.Validate(); err != nil {
		return err
	}
	m.options = o
	m.present = true
	return nil
}
func (m *consumerTray) ClearTray(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.present = false
	return nil
}
func TestPublicTrayConsumer(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	fake := &consumerTray{}
	a.SetServices(fake)
	tray, err := a.Tray()
	if err != nil || !tray.TrayCapabilities().CommandMenu {
		t.Fatal(err)
	}
	o := app.TrayOptions{Title: "Vellum", Tooltip: "Vellum board workspace", Menu: []app.TrayMenuItem{{Command: "window.show"}, {Separator: true}, {Command: "document.save"}}}
	if err := tray.SetTray(context.Background(), o); err != nil || !fake.present || len(fake.options.Menu) != 3 {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := tray.ClearTray(ctx); !errors.Is(err, context.Canceled) || !fake.present {
		t.Fatal(err)
	}
	if err := tray.ClearTray(context.Background()); err != nil || fake.present {
		t.Fatal(err)
	}
}
