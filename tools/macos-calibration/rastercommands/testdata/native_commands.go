// Copyright (c) 2026 Vellum contributors. SPDX-License-Identifier: MIT.
package rasterx

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/pimage"
	"cogentcore.org/core/paint/render"
	"fmt"
	"image"
	"image/color"
	"sync"
)

type VellumCommand struct {
	Index                 int
	Decision              string
	CropNoPaintCompatible bool
	Scroll                bool
	Type                  string
	Rect                  image.Rectangle
	SourcePos             image.Point
	Operation             int
	Command               int
	Mask                  bool
	SourceType, ColorType string
	HasUniformRGBA        bool
	UniformRGBA           [4]uint32
	PathLength            int
	PathTransform         math32.Matrix2
	HasFill, HasStroke    bool
}
type VellumCommandFrame struct {
	Cohort       bool
	Complete     bool
	OwnedScanner bool
	Target       image.Rectangle
	Items        []VellumCommand
}
type VellumCommandCapture struct {
	Overflow      bool
	Invalid       bool
	CohortStarted bool
	Frames        []VellumCommandFrame
}

var commandCapture struct {
	sync.Mutex
	Frames        []VellumCommandFrame
	Overflow      bool
	Invalid       bool
	CohortStarted bool
	Items         int
}

func VellumCommandReset() {
	commandCapture.Lock()
	defer commandCapture.Unlock()
	commandCapture.Frames = nil
	commandCapture.Overflow = false
	commandCapture.Invalid = false
	commandCapture.CohortStarted = false
	commandCapture.Items = 0
}
func VellumCommandSnapshot() VellumCommandCapture {
	commandCapture.Lock()
	defer commandCapture.Unlock()
	out := VellumCommandCapture{Overflow: commandCapture.Overflow, Invalid: commandCapture.Invalid, CohortStarted: commandCapture.CohortStarted, Frames: make([]VellumCommandFrame, len(commandCapture.Frames))}
	for i, f := range commandCapture.Frames {
		out.Frames[i] = VellumCommandFrame{Cohort: f.Cohort, Complete: f.Complete, OwnedScanner: f.OwnedScanner, Target: f.Target, Items: append([]VellumCommand(nil), f.Items...)}
	}
	return out
}
func commandUniform(src image.Image, out *VellumCommand) {
	out.SourceType = fmt.Sprintf("%T", src)
	u, ok := src.(*image.Uniform)
	if !ok || u == nil {
		return
	}
	out.ColorType = fmt.Sprintf("%T", u.C)
	switch u.C.(type) {
	case color.RGBA, color.RGBA64, color.NRGBA, color.NRGBA64, color.Alpha, color.Alpha16, color.Gray, color.Gray16:
		r, g, b, a := u.C.RGBA()
		out.HasUniformRGBA = true
		out.UniformRGBA = [4]uint32{r, g, b, a}
	}
}
func (rs *Renderer) vellumCaptureCommands(commands render.Render) int {
	commandCapture.Lock()
	defer commandCapture.Unlock()
	if len(commandCapture.Frames) >= 32 || len(commands) > 1024-commandCapture.Items {
		commandCapture.Overflow = true
		return -1
	}
	f := VellumCommandFrame{Cohort: commandCapture.CohortStarted, OwnedScanner: rs.Scanner != nil && rs.Raster != nil && rs.Raster.Filler.Scanner == rs.Scanner, Items: make([]VellumCommand, len(commands))}
	if rs.image != nil {
		f.Target = rs.image.Bounds()
	}
	for i, raw := range commands {
		row := VellumCommand{Index: i, Type: fmt.Sprintf("%T", raw)}
		switch x := raw.(type) {
		case *render.ContextPush:
			if x != nil {
				row.Scroll = x.Scroll != nil
			}
		case *pimage.Params:
			if x != nil {
				row.Command = int(x.Cmd)
				row.Rect = x.Rect
				row.SourcePos = x.SourcePos
				row.Operation = int(x.Op)
				row.Mask = x.Mask != nil
				commandUniform(x.Source, &row)
			}
		case *render.Path:
			if x != nil {
				row.CropNoPaintCompatible = rs.cropNoPaintPath(x)
				row.PathLength = len(x.Path)
				row.PathTransform = x.Context.Cumulative
				row.HasFill = x.Context.Style.HasFill()
				row.HasStroke = x.Context.Style.HasStroke()
			}
		}
		f.Items[i] = row
	}
	commandCapture.Items += len(commands)
	commandCapture.Frames = append(commandCapture.Frames, f)
	return len(commandCapture.Frames) - 1
}

// Decisions are emitted only at root renderItems depth, after the unchanged
// branch executes. Nested scroll/text replay is outside this fixture's scope.
func (rs *Renderer) vellumCommandDecision(index int, decision string) {
	if rs.vellumCommandDepth != 1 || rs.vellumCommandFrame < 0 {
		return
	}
	commandCapture.Lock()
	defer commandCapture.Unlock()
	f := rs.vellumCommandFrame
	if f >= len(commandCapture.Frames) || index < 0 || index >= len(commandCapture.Frames[f].Items) || commandCapture.Frames[f].Items[index].Decision != "" {
		commandCapture.Invalid = true
		return
	}
	commandCapture.Frames[f].Items[index].Decision = decision
}
func vellumFinishCommands(index int) {
	if index < 0 {
		return
	}
	commandCapture.Lock()
	defer commandCapture.Unlock()
	if index >= len(commandCapture.Frames) || commandCapture.Frames[index].Complete {
		commandCapture.Invalid = true
		return
	}
	commandCapture.Frames[index].Complete = true
}

// Mark only the already-locked geometry boundary; this is not an input tag.
func VellumCommandBeginCohort() bool {
	commandCapture.Lock()
	defer commandCapture.Unlock()
	if commandCapture.CohortStarted {
		return false
	}
	commandCapture.CohortStarted = true
	return true
}
