//go:build !darwin || !cgo || ios

// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
package core

type renderPriority struct{}

func beginRenderPriority() *renderPriority { return nil }
func (*renderPriority) end()               {}
