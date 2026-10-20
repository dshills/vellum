//go:build darwin

// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import (
	"cogentcore.org/core/system/driver/desktop"
	"errors"
	"time"
)

func configureDisplayBackpressure(timeout time.Duration, fail func(error)) (func(), error) {
	if err := validateDisplayBackpressure(timeout); err != nil {
		return nil, err
	}
	if timeout == 0 {
		return func() {}, nil
	}
	if fail == nil || desktop.TheApp.NativeAttach == nil {
		return nil, errors.New("cogent: early native attachment/failure owner required for display experiment")
	}
	previous := desktop.TheApp.NativeAttach
	desktop.TheApp.NativeAttach = func(w *desktop.Window) (desktop.NativeAttachment, error) {
		attachment, err := previous(w)
		if err != nil {
			return attachment, err
		}
		owned, ok := attachment.(*accessibilityAttachment)
		if !ok {
			return attachment, errors.New("cogent: display experiment requires existing accessibility attachment")
		}
		monitor, err := attachDisplayBackpressure(w, timeout, fail)
		if err != nil {
			return attachment, err
		}
		owned.display = monitor
		return attachment, nil
	}
	return func() { desktop.TheApp.NativeAttach = previous }, nil
}
