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
