//go:build !darwin || !cgo || offscreen || !vellum_native_checks

package main

import "errors"

func beginProgressCleanup() (func(), error) {
	return nil, errors.New("progress-smoke requires macOS native checks build")
}
func finishProgressCleanup() error {
	return errors.New("progress-smoke requires macOS native checks build")
}
