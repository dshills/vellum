//go:build !darwin || !cgo || offscreen

package cogent

func configureServiceCleanup(*platformServices) (func(), error) { return func() {}, nil }
