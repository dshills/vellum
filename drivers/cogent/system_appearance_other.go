//go:build !darwin || !cgo

package cogent

import "fmt"

func validateSystemAppearanceSupport(enabled bool) error {
	if enabled {
		return fmt.Errorf("cogent: system appearance observation requires macOS with cgo")
	}
	return nil
}
func nativeStartSystemAppearance(uint64) {}
func nativeStopSystemAppearance(uint64)  {}
