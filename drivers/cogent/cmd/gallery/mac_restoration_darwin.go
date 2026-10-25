//go:build darwin && cgo

package main

/*
void vellumGalleryFreshCocoaSession(void);
int vellumGalleryRestorationFixture(void);
*/
import "C"

// This demo owns its model; Cocoa cannot restore worker-scoped panel sessions.
func configureMacWidgetRestoration()  { C.vellumGalleryFreshCocoaSession() }
func galleryRestorationFixture() bool { return C.vellumGalleryRestorationFixture() != 0 }
