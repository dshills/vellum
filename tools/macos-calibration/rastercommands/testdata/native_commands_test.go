// Copyright (c) 2026 Vellum contributors. SPDX-License-Identifier: MIT.
package rasterx

import (
	"bytes"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/pimage"
	"cogentcore.org/core/paint/ppath"
	"cogentcore.org/core/paint/render"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestVellumNativeCommandsDecisionsAndPixels(t *testing.T) {
	VellumCommandReset()
	defer VellumCommandReset()
	bounds := image.Rect(0, 0, 100, 100)
	inner := bounds.Inset(5)
	makeDraw := func(r image.Rectangle, c color.Color, op draw.Op) *pimage.Params {
		return &pimage.Params{Cmd: pimage.Draw, Rect: r, SourcePos: r.Min, Source: image.NewUniform(c), Op: op}
	}
	var path ppath.Path
	path.Rectangle(0, 0, 100, 100)
	noop := &render.Path{Path: path}
	noop.Context.Cumulative = math32.Identity2()
	noop.Context.Bounds.Rect = math32.B2FromRect(bounds)
	rs := New(math32.Vec2(100, 100)).(*Renderer)
	items := render.Render{makeDraw(bounds, color.RGBA{}, draw.Src), makeDraw(bounds, color.NRGBA{255, 0, 0, 255}, draw.Over), noop, &render.ContextPush{}, makeDraw(inner, color.NRGBA{255, 255, 255, 255}, draw.Src), &render.ContextPop{}}
	rs.Render(items)
	s := VellumCommandSnapshot()
	if s.Invalid || s.Overflow || len(s.Frames) != 1 || !s.Frames[0].Complete || !s.Frames[0].OwnedScanner {
		t.Fatalf("capture invalid %+v", s)
	}
	want := map[int]string{0: "covered-next", 1: "cropped-uniform", 2: "render-path", 4: "render-image"}
	for index, value := range want {
		if s.Frames[0].Items[index].Decision != value {
			t.Fatalf("decision%d got%s", index, s.Frames[0].Items[index].Decision)
		}
	}
	if !s.Frames[0].Items[2].CropNoPaintCompatible || rs.vellumCommandFrame != -1 || rs.vellumCommandDepth != 0 {
		t.Fatal("guard/temporary state cleanup")
	}
	expected := image.NewRGBA(bounds)
	draw.Draw(expected, bounds, image.NewUniform(color.NRGBA{255, 0, 0, 255}), bounds.Min, draw.Src)
	draw.Draw(expected, inner, image.White, inner.Min, draw.Src)
	if !bytes.Equal(rs.image.Pix, expected.Pix) {
		t.Fatal("instrumentation changed original crop/clear pixels")
	}
}
func TestVellumNativeCommandsDuplicateAndNestedGuards(t *testing.T) {
	VellumCommandReset()
	defer VellumCommandReset()
	rs := &Renderer{}
	rs.vellumCommandFrame = rs.vellumCaptureCommands(render.Render{&pimage.Params{}})
	rs.vellumCommandDepth = 2
	rs.vellumCommandDecision(999, "render-image")
	if VellumCommandSnapshot().Invalid {
		t.Fatal("nested scope was recorded")
	}
	rs.vellumCommandDepth = 1
	rs.vellumCommandDecision(0, "render-image")
	rs.vellumCommandDecision(0, "render-image")
	if !VellumCommandSnapshot().Invalid {
		t.Fatal("duplicate decision accepted")
	}
	VellumCommandReset()
	if VellumCommandSnapshot().Invalid {
		t.Fatal("invalid state not reset")
	}
}

func TestVellumNativeCommandsAdmissionBoundary(t *testing.T) {
	VellumCommandReset()
	defer VellumCommandReset()
	rs := &Renderer{}
	rs.vellumCaptureCommands(nil)
	if !VellumCommandBeginCohort() || VellumCommandBeginCohort() {
		t.Fatal("boundary must be one shot")
	}
	rs.vellumCaptureCommands(nil)
	s := VellumCommandSnapshot()
	if !s.CohortStarted || s.Frames[0].Cohort || !s.Frames[1].Cohort {
		t.Fatal("boundary rewrote earlier frames")
	}
	VellumCommandReset()
	if VellumCommandSnapshot().CohortStarted {
		t.Fatal("boundary not reset")
	}
}
