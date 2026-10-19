package app

import "fmt"

// ChromeStyle selects native window decoration. Standard preserves backend defaults.
type ChromeStyle uint8

const (
	ChromeStandard ChromeStyle = iota
	// ChromeUnified keeps OS-owned window buttons and a native unified toolbar.
	// On macOS content extends behind a transparent, title-hidden header. The
	// native header is reserved for window dragging; place controls below it.
	ChromeUnified
)

// WindowAppearance affects this native window only, independently of ui.Theme.
type WindowAppearance uint8

const (
	AppearanceAutomatic WindowAppearance = iota
	AppearanceLight
	AppearanceDark
)

// WindowChrome is an immutable creation policy. Its zero value changes nothing.
// Explicit appearance or Unified chrome currently requires the macOS desktop
// driver; unsupported native backends return an error rather than approximate it.
// Width/Height continue to request client-view Dp, including the reserved header
// in Unified mode. They are not outer-frame dimensions. ui.Theme styles content
// separately; no application-wide or operating-system appearance is changed.
type WindowChrome struct {
	Style      ChromeStyle
	Appearance WindowAppearance
}

func (c WindowChrome) Validate() error {
	if c.Style > ChromeUnified || c.Appearance > AppearanceDark {
		return fmt.Errorf("vellum: invalid window chrome")
	}
	return nil
}
