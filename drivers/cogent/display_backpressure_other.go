//go:build !darwin

// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import (
	"errors"
	"time"
)

func configureDisplayBackpressure(timeout time.Duration, _ func(error)) (func(), error) {
	if err := validateDisplayBackpressure(timeout); err != nil {
		return nil, err
	}
	if timeout != 0 {
		return nil, errors.New("cogent: experimental display backpressure requires macOS")
	}
	return func() {}, nil
}
