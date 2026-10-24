package cogent

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"testing"
)

func TestNativeColorReplyAndParentAdmission(t *testing.T) {
	for _, wire := range []string{`{}`, `{"R":-1,"G":0,"B":0,"A":255}`, `{"R":256,"G":0,"B":0,"A":255}`, `{"R":1,"G":2,"B":3,"A":255,"extra":0}`, `{"R":1,"G":2,"B":3,"A":255} {}`} {
		if _, err := decodeColorPanel(wire, true); err == nil {
			t.Fatal(wire)
		}
	}
	wire := `{"R":1,"G":2,"B":3,"A":128}`
	if _, err := decodeColorPanel(wire, false); err == nil {
		t.Fatal("opaque panel admitted alpha")
	}
	if c, err := decodeColorPanel(wire, true); err != nil || c.A != 128 {
		t.Fatal(c, err)
	}
	a, other := dropApp(t), dropApp(t)
	ref, _ := other.WindowRef("main")
	s := &platformServices{owner: a}
	if _, err := s.ChooseColor(context.Background(), ref, app.ColorPanelOptions{Title: "Colors"}); !errors.Is(err, app.ErrInvalidWindowRef) {
		t.Fatal(err)
	}
	ctx, end := context.WithCancel(context.Background())
	end()
	if _, err := s.ChooseColor(ctx, ref, app.ColorPanelOptions{Title: "Colors"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
