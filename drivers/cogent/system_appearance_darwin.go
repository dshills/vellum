//go:build darwin && cgo

package cogent

/*
#include <stdint.h>
void vellumStartSystemAppearance(uint64_t);
void vellumStopSystemAppearance(uint64_t);
*/
import "C"
import "image/color"

func nativeStartSystemAppearance(token uint64) { C.vellumStartSystemAppearance(C.uint64_t(token)) }
func nativeStopSystemAppearance(token uint64)  { C.vellumStopSystemAppearance(C.uint64_t(token)) }

//export vellumSystemAppearanceActive
func vellumSystemAppearanceActive(token C.uint64_t) C.int {
	appearanceRegistry.Lock()
	s := appearanceRegistry.current
	appearanceRegistry.Unlock()
	if s != nil && s.id == uint64(token) {
		return 1
	}
	return 0
}

//export vellumSystemAppearance
func vellumSystemAppearance(token C.uint64_t, dark, contrast, motion, transparency C.int, red, green, blue C.int) {
	if red < 0 || red > 255 || green < 0 || green > 255 || blue < 0 || blue > 255 {
		return
	}
	receiveSystemAppearance(uint64(token), SystemAppearance{Known: true, Dark: dark != 0, HighContrast: contrast != 0, ReduceMotion: motion != 0, ReduceTransparency: transparency != 0, Accent: color.NRGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: 255}})
}

func validateSystemAppearanceSupport(bool) error { return nil }
