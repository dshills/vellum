//go:build !darwin || !cgo

package main

import (
	"context"
	"github.com/dshills/vellum/app"
)

func receiveFileURLs(context.Context, string, []string, float64, float64) error {
	return app.ErrUnavailable
}
