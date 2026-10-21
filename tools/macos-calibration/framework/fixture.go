// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/styles"
	"context"
	"errors"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
	"image"
	"image/color"
)

const windowID = "calibration"

type patch struct {
	owner    *app.Application
	ref      app.WindowRef
	sequence uint64
	commands *command.Registry
}

func shade(n uint64) color.NRGBA {
	if n%2 == 1 {
		return color.NRGBA{R: 255, A: 255}
	}
	return color.NRGBA{G: 255, A: 255}
}
func (p *patch) node() ui.Node {
	return ui.Node{ID: "patch", Kind: ui.Column, Grow: true, Style: ui.Style{Background: ui.Some(shade(p.sequence)), Padding: ui.Some(ui.UniformInsets(0))}}
}
func newFixture() (*app.Application, *patch, error) {
	core.TheApp.SceneInit = func(sc *core.Scene) {
		sc.FinalStyler(func(s *styles.Style) { s.Background = image.NewUniform(color.NRGBA{R: 255, G: 255, B: 255, A: 255}) })
	}
	a, e := app.New(app.Config{QueueCapacity: 64, CompletionCapacity: 8})
	if e != nil {
		return nil, nil, e
	}
	p := &patch{owner: a, commands: command.New()}
	e = p.commands.Register(command.Command{ID: "toggle", Label: "Toggle calibration patch", Shortcut: string(key.NewChord(' ', key.CodeSpacebar, 0)), Execute: func(context.Context) error { p.sequence++; return a.Invalidate(windowID) }})
	if e != nil {
		a.Close()
		return nil, nil, e
	}
	p.ref, e = a.OpenWindow(app.Window{ID: windowID, Title: "Vellum matched calibration", Width: 1500, Height: 1250, FixedSize: true, DisableRootScrolling: true, View: p.node})
	if e != nil {
		a.Close()
		return nil, nil, e
	}
	return a, p, nil
}
func (p *patch) expect(sequence uint64) (uint64, error) {
	if p.sequence != sequence {
		return 0, errors.New("calibration: missing or duplicate routed transition")
	}
	return sequence, nil
}
