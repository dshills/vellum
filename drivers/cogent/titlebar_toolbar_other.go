//go:build !darwin || !cgo

package cogent

import "fmt"

func validateTitlebarToolbarSupport(p *preparedTitlebarToolbar) error {
	if p != nil {
		return fmt.Errorf("cogent: titlebar prototype requires macOS with cgo")
	}
	return nil
}
func titlebarToolbarAllowed(*accessibilityBridge, bool) {}

func nativeTitlebarCustomize(uint64) {}

func nativeTitlebarRequestAttach(uint64, string, string) {}
func nativeTitlebarStateChanged(uint64)                  {}

func nativeTitlebarReset(uint64) {}
