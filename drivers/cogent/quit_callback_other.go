//go:build !darwin || !cgo

package cogent

func bindNativeQuit(callback *quitCallback) (func(), error) { return func() {}, nil }
