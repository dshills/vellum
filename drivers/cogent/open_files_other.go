//go:build !darwin || !cgo

package cogent

import "github.com/dshills/vellum/app"

func configureOpenFiles(owner *app.Application, callback func(OpenFilesEvent)) (*openFilesReceiver, func(), error) {
	if callback != nil {
		return nil, func() {}, app.ErrUnavailable
	}
	return nil, func() {}, nil
}
