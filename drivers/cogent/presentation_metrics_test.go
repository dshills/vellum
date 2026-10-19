// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
package cogent

import (
	"cogentcore.org/core/system/driver/desktop"
	"reflect"
	"testing"
)

type presentationAttachFixture struct {
	calls    int
	observer desktop.PresentationObserver
}

func (f *presentationAttachFixture) SetPresentationObserver(s desktop.PresentationObserver) error {
	f.calls++
	f.observer = s
	return nil
}
func TestPresentationMetricsAdapterRetiresDetachedScope(t *testing.T) {
	m, err := NewPresentationMetrics(4)
	if err != nil {
		t.Fatal(err)
	}
	sink, err := m.bind("main")
	if err != nil {
		t.Fatal(err)
	}
	w := &nativeWindow{presentation: sink}
	target := &presentationAttachFixture{}
	if attachPresentation(w, target) != nil || attachPresentation(w, target) != nil || target.calls != 1 {
		t.Fatal("show repeated observer installation")
	}
	target.observer.Receipt(1, 1)
	sequence := target.observer.Acquired(desktop.PresentationFrame{Tag: 1, Generation: 2, Composition: 3, Armed: true}, 2)
	target.observer.Presented(sequence, 0, 3)
	target.observer.Unavailable(desktop.PresentationFrame{Composition: 4}, 4)
	sink.Closed()
	before := m.Snapshot()
	target.observer.Presented(sequence, 9, 9)
	target.observer.Receipt(2, 5)
	if after := m.Snapshot(); !reflect.DeepEqual(before, after) || after.Failed || len(after.Samples) != 2 || after.Unavailable != 1 || after.Samples[0].DrawableID != 0 || after.Samples[0].Presented != 3 || !after.Scopes[0].Closed {
		t.Fatal(after)
	}
}
func TestPresentationMetricsConfigurationAndCopiedWrapper(t *testing.T) {
	for _, n := range []int{-1, 0, 65537} {
		if _, err := NewPresentationMetrics(n); err == nil {
			t.Fatal(n)
		}
	}
	m, err := NewPresentationMetrics(1)
	if err != nil {
		t.Fatal(err)
	}
	copied := *m
	for _, invalid := range []*PresentationMetrics{nil, {}, &copied} {
		if invalid.valid() || !invalid.Snapshot().Failed {
			t.Fatal(invalid)
		}
	}
	if err := attachPresentation(&nativeWindow{}, struct{}{}); err != nil {
		t.Fatal("disabled path", err)
	}
	sink, _ := m.bind("main")
	if err := attachPresentation(&nativeWindow{presentation: sink}, struct{}{}); err == nil {
		t.Fatal("unsupported observation accepted")
	}
}
