// Copyright (c) 2026 Vellum contributors. SPDX-License-Identifier: MIT.
package rasterx

import (
	"cogentcore.org/core/paint/pimage"
	"cogentcore.org/core/paint/render"
	"image"
	"image/color"
	"testing"
)

type commandSideEffectColor struct{ calls *int }

func (c commandSideEffectColor) RGBA() (uint32, uint32, uint32, uint32) {
	*c.calls++
	return 0, 0, 0, 65535
}
func TestVellumCommandsNoCustomCallbacksAndDetachedSnapshot(t *testing.T) {
	VellumCommandReset()
	defer VellumCommandReset()
	calls := 0
	rs := &Renderer{image: image.NewRGBA(image.Rect(0, 0, 4, 3))}
	rs.vellumCaptureCommands(render.Render{&pimage.Params{Source: image.NewUniform(commandSideEffectColor{&calls})}, &pimage.Params{Source: image.NewUniform(color.NRGBA{3, 5, 7, 255})}})
	first := VellumCommandSnapshot()
	if calls != 0 || first.Overflow || len(first.Frames) != 1 || first.Frames[0].Items[0].HasUniformRGBA || !first.Frames[0].Items[1].HasUniformRGBA {
		t.Fatal("custom callback or missing scalar metadata")
	}
	first.Frames[0].Items[1].SourceType = "changed"
	if VellumCommandSnapshot().Frames[0].Items[1].SourceType == "changed" {
		t.Fatal("snapshot aliases recorder")
	}
}
func TestVellumCommandsCapsAndReset(t *testing.T) {
	VellumCommandReset()
	defer VellumCommandReset()
	rs := &Renderer{}
	for i := 0; i < 32; i++ {
		rs.vellumCaptureCommands(make(render.Render, 32))
	}
	s := VellumCommandSnapshot()
	if s.Overflow || len(s.Frames) != 32 {
		t.Fatal("exact32frames/1024items rejected")
	}
	rs.vellumCaptureCommands(nil)
	s = VellumCommandSnapshot()
	if !s.Overflow || len(s.Frames) != 32 {
		t.Fatal("frame cap not retained")
	}
	VellumCommandReset()
	rs.vellumCaptureCommands(make(render.Render, 1025))
	s = VellumCommandSnapshot()
	if !s.Overflow || len(s.Frames) != 0 {
		t.Fatal("item cap partial capture")
	}
	VellumCommandReset()
	s = VellumCommandSnapshot()
	if s.Overflow || len(s.Frames) != 0 {
		t.Fatal("reset did not release records")
	}
}
