package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"github.com/dshills/vellum/app"
	"testing"
)

func TestWindowChromeCreationPolicyAndSceneInset(t *testing.T) {
	for _, policy := range []app.WindowChrome{{}, {Style: app.ChromeUnified, Appearance: app.AppearanceDark}} {
		body := core.NewBody("chrome policy")
		configureWindowChromeContent(body, app.Window{Chrome: policy})
		stage := initialWindowStage(body, app.Window{Width: 230, Height: 408, Chrome: policy})
		if uint8(stage.WindowChrome.Style) != uint8(policy.Style) || uint8(stage.WindowChrome.Appearance) != uint8(policy.Appearance) {
			t.Fatal("policy not copied")
		}
		body.Scene.Stage = stage
		body.Scene.Style()
		padding := body.Scene.Styles.Padding.Top.Value
		if policy.Style == app.ChromeUnified && padding != 0 {
			t.Fatal("unified client inset", padding)
		}
		if policy.Style == app.ChromeStandard && padding != 8 {
			t.Fatal("default scene changed", padding)
		}
	}
}

func TestRootScrollingPolicyPreservesDefaultAndExplicitScroll(t *testing.T) {
	for _, disable := range []bool{false, true} {
		body := core.NewBody("scroll policy")
		host := core.NewFrame(body)
		inner := core.NewFrame(host)
		inner.Styler(func(s *styles.Style) { s.Overflow.Set(styles.OverflowAuto) })
		body.Scene.Stage = body.NewWindow()
		body.Scene.Style()
		body.Style()
		host.Style()
		sceneBefore, bodyBefore, hostBefore := body.Scene.Styles.Overflow, body.Styles.Overflow, host.Styles.Overflow
		configureRootScrolling(body, host, app.Window{DisableRootScrolling: disable})
		body.Scene.Style()
		body.Style()
		host.Style()
		inner.Style()
		if inner.Styles.Overflow.X != styles.OverflowAuto || inner.Styles.Overflow.Y != styles.OverflowAuto {
			t.Fatal("explicit scroll was changed")
		}
		if disable {
			for _, s := range []styles.Style{body.Scene.Styles, body.Styles, host.Styles} {
				if s.Overflow.X != styles.OverflowHidden || s.Overflow.Y != styles.OverflowHidden {
					t.Fatal("implicit root scrolling remains enabled")
				}
			}
		} else if sceneBefore != body.Scene.Styles.Overflow || bodyBefore != body.Styles.Overflow || hostBefore != host.Styles.Overflow {
			t.Fatal("default scrolling behavior changed")
		}
	}
}
