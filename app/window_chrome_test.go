package app

import (
	"github.com/dshills/vellum/ui"
	"testing"
)

func TestWindowChromeValidationAndValueOwnership(t *testing.T) {
	for _, policy := range []WindowChrome{{}, {ChromeUnified, AppearanceDark}, {ChromeStandard, AppearanceLight}} {
		a, _ := New(Config{})
		spec := Window{ID: "one", Title: "One", Width: 230, Height: 408, Chrome: policy, View: func() ui.Node { return ui.Node{Kind: ui.Column} }}
		if err := a.Open(spec); err != nil {
			t.Fatal(err)
		}
		spec.Chrome = WindowChrome{Style: 255}
		if got := a.Windows()[0].Chrome; got != policy {
			t.Fatalf("copied policy=%v want%v", got, policy)
		}
		a.Close()
	}
	for _, bad := range []WindowChrome{{Style: 2}, {Appearance: 3}, {Style: 255, Appearance: 255}} {
		a, _ := New(Config{})
		err := a.Open(Window{ID: "bad", Title: "Bad", Width: 1, Height: 1, Chrome: bad, View: func() ui.Node { return ui.Node{Kind: ui.Column} }})
		if err == nil || len(a.Windows()) != 0 {
			t.Fatal("invalid policy admitted", bad, err)
		}
		a.Close()
	}
}

func TestWindowChromePoliciesArePerWindow(t *testing.T) {
	a, _ := New(Config{})
	defer a.Close()
	for i, appearance := range []WindowAppearance{AppearanceLight, AppearanceDark} {
		id := []string{"light", "dark"}[i]
		if err := a.Open(Window{ID: id, Title: id, Width: 230, Height: 408, Chrome: WindowChrome{Style: ChromeUnified, Appearance: appearance}, View: func() ui.Node { return ui.Node{Kind: ui.Column} }}); err != nil {
			t.Fatal(err)
		}
	}
	got := a.Windows()
	if got[0].Chrome.Appearance != AppearanceLight || got[1].Chrome.Appearance != AppearanceDark {
		t.Fatal("window policy leaked", got)
	}
}
