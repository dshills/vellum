package cogent

import (
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"testing"
)

func TestNativeFontReplyAndParentAdmission(t *testing.T) {
	for _, wire := range []string{`{}`, `{"Name":"HelveticaNeue","Family":"Helvetica Neue","Size":0}`, `{"Name":"HelveticaNeue","Family":"Helvetica Neue","Size":513}`, `{"Name":"HelveticaNeue","Family":"Helvetica Neue","Size":13,"Extra":0}`, `{"Name":"HelveticaNeue","Family":"Helvetica Neue","Size":13} {}`} {
		if _, err := decodeFontPanel(wire); err == nil {
			t.Fatal(wire)
		}
	}
	if v, err := decodeFontPanel(`{"Name":"HelveticaNeue-Bold","Family":"Helvetica Neue","Size":24,"Bold":true,"Italic":false}`); err != nil || v.Size != 24 {
		t.Fatal(v, err)
	}
	a, other := dropApp(t), dropApp(t)
	ref, _ := other.WindowRef("main")
	s := &platformServices{owner: a}
	if _, err := s.ChooseFont(context.Background(), ref, app.FontPanelOptions{Title: "Fonts", Name: "HelveticaNeue", Size: 13}); !errors.Is(err, app.ErrInvalidWindowRef) {
		t.Fatal(err)
	}
}
