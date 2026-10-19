package cogent

import (
	"fmt"

	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/system"
	"github.com/dshills/vellum/app"
)

// initialWindowChrome carries only immutable values into native creation.
func initialWindowChrome(stage *core.Stage, spec app.Window) *core.Stage {
	return stage.SetWindowChrome(system.WindowChrome{Style: system.ChromeStyle(spec.Chrome.Style), Appearance: system.WindowAppearance(spec.Chrome.Appearance)})
}

// validateWindowChromeSupport is called before entering the native run loop and
// before accepting newly registered windows. Offscreen rendering does not
// approximate native decoration or appearance.
func validateWindowChromeSupport(spec app.Window) error {
	if err := spec.Chrome.Validate(); err != nil {
		return err
	}
	if spec.Chrome == (app.WindowChrome{}) {
		return nil
	}
	if system.TheApp == nil || system.TheApp.Platform() != system.MacOS {
		return fmt.Errorf("cogent: explicit window chrome requires the macOS desktop driver")
	}
	return nil
}

// Unified chrome exposes the full client surface without the backend Scene's
// default 8dp inset. Applications still reserve the native header themselves.
func configureWindowChromeContent(body *core.Body, spec app.Window) {
	if spec.Chrome.Style == app.ChromeUnified {
		body.Scene.Styler(func(s *styles.Style) { s.Padding.Zero() })
	}
}

// Only the implicit outer viewport is affected. An application's explicit
// ui.Scroll controls retain their own scrolling behavior.
func configureRootScrolling(body *core.Body, host *core.Frame, spec app.Window) {
	if !spec.DisableRootScrolling {
		return
	}
	hide := func(s *styles.Style) { s.Overflow.Set(styles.OverflowHidden) }
	body.Scene.Styler(hide)
	body.Styler(hide)
	host.Styler(hide)
}
