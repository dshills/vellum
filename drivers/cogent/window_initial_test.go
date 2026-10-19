package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"github.com/dshills/vellum/app"
	"testing"
)

func TestInitialWindowSizeDoesNotBecomePersistentBodyMinimum(t *testing.T) {
	body := core.NewBody("initial size")
	configureInitialWindow(body)
	stage := initialWindowStage(body, app.Window{Width: 1100, Height: 800})
	if stage.UseMinSize || stage.InitialSizeDP != math32.Vec2(1100, 800) {
		t.Fatal("initial creation request not configured")
	}
	body.Style()
	if body.Styles.Min.X.Value != 0 || body.Styles.Min.Y.Value != 0 || body.Styles.Grow.X != 1 || body.Styles.Grow.Y != 1 {
		t.Fatalf("initial dimensions became lasting layout constraints: %+v", body.Styles)
	}
	// Application content constraints remain independent of the initial request.
	body.Styler(func(s *styles.Style) { s.Min.Set(units.Dp(120), units.Dp(80)) })
	body.Style()
	if body.Styles.Min.X.Value != 120 || body.Styles.Min.Y.Value != 80 {
		t.Fatal("application minimum overwritten")
	}
}

func TestInitialWindowFixedSizePolicy(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		fixed, wantResizable bool
	}{
		{"default resizable", false, true},
		{"fixed", true, false},
		{"default after fixed", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := core.NewBody("fixed-size creation policy")
			stage := initialWindowStage(body, app.Window{Width: 230, Height: 408, FixedSize: tc.fixed})
			if stage.Resizable != tc.wantResizable {
				t.Fatalf("Resizable=%v want%v for FixedSize=%v", stage.Resizable, tc.wantResizable, tc.fixed)
			}
			if stage.UseMinSize || stage.InitialSizeDP != math32.Vec2(230, 408) {
				t.Fatal("fixed-size policy altered creation dimensions or content-minimum policy")
			}
		})
	}
}
