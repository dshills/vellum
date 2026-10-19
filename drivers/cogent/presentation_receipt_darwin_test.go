//go:build darwin

// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import "testing"

func TestPresentationReceiptDoesNotTraverseNativeWindow(t *testing.T) {
	m, err := NewPresentationMetrics(1)
	if err != nil {
		t.Fatal(err)
	}
	sink, err := m.bind("main")
	if err != nil {
		t.Fatal(err)
	}
	// Deliberately no body/scene/stage/native window. Only recorder+host clock.
	recordPresentationReceipt(&nativeWindow{presentation: sink}, 7)
	v := m.Snapshot()
	if v.Failed || len(v.Inputs) != 1 || v.Inputs[0].Tag != 7 || v.Inputs[0].Received <= 0 {
		t.Fatal(v)
	}
	sink.Closed()
	recordPresentationReceipt(&nativeWindow{presentation: sink}, 8)
	if len(m.Snapshot().Inputs) != 1 {
		t.Fatal("retired receipt")
	}
}
