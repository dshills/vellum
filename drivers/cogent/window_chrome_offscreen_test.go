//go:build offscreen

package cogent

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
)

// Native decoration cannot be faithfully approximated in a manual session.
// Reject the entire desired topology before evaluating even an earlier valid
// window's View; refusal must release the manual owner for a valid retry.
func TestManualRejectsNativeChromeBeforeWindowConstruction(t *testing.T) {
	for _, policy := range []app.WindowChrome{{Style: app.ChromeUnified}, {Appearance: app.AppearanceLight}, {Appearance: app.AppearanceDark}} {
		t.Run(fmt.Sprintf("style-%d-appearance-%d", policy.Style, policy.Appearance), func(t *testing.T) {
			config := manualConfig(t, 7)
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			views := 0
			view := func() ui.Node {
				views++
				return ui.Node{ID: "root", Kind: ui.Label, Text: "Visible only after valid retry"}
			}
			if err = a.Open(app.Window{ID: "valid", Title: "Valid manual", Width: 400, Height: 240, View: view}); err != nil {
				t.Fatal(err)
			}
			bad, err := a.OpenWindow(app.Window{ID: "native", Title: "Unsupported native", Width: 400, Height: 240, Chrome: policy, View: view})
			if err != nil {
				t.Fatal(err)
			}
			before := a.Pending()
			windows := a.Windows()
			if manual, err := NewManual(testContext, a, config); err == nil || manual != nil || !strings.Contains(err.Error(), "native window chrome unavailable") {
				t.Fatalf("native chrome admitted/incorrect refusal: manual=%v err=%v", manual, err)
			}
			if views != 0 || !reflect.DeepEqual(before, a.Pending()) || len(a.Windows()) != len(windows) || a.Windows()[0].Ref() != windows[0].Ref() || a.Windows()[1].Ref() != bad || a.Context().Err() != nil {
				t.Fatal("refusal evaluated View or mutated domain topology/readiness", views, a.Pending())
			}
			if _, err = a.Services(); !errors.Is(err, app.ErrUnavailable) {
				t.Fatal("failed manual construction installed services", err)
			}
			if err = a.CloseWindow(bad); err != nil {
				t.Fatal(err)
			}
			manual, err := NewManual(testContext, a, config)
			if err != nil {
				t.Fatal("refusal retained exclusive manual claim", err)
			}
			if err = manual.Close(testContext); err != nil {
				t.Fatal(err)
			}
			if views == 0 {
				t.Fatal("valid retry did not construct its managed window")
			}
		})
	}
}
