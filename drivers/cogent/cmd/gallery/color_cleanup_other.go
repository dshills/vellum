//go:build !darwin || !cgo || offscreen || !vellum_native_checks

package main

import "errors"

func beginColorCleanup() (func(), error) {
	return nil, errors.New("color-smoke requires a Mac vellum_native_checks build")
}
func finishColorCleanup() error { return errors.New("native color cleanup unavailable") }
