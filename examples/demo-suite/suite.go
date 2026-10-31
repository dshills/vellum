package main

import (
	"context"
	"github.com/dshills/vellum/ui"
	"os/exec"
	"path/filepath"
)

func (d *demo) suiteView() ui.Node {
	n := ui.Node{ID: "suite", Kind: ui.Scroll, Grow: true, Children: []ui.Node{label("suite-intro", "Open a focused demo below. Each is built from the same Vellum source checkpoint as this hub.")}}
	for _, feature := range featureDemos {
		n.Children = append(n.Children, d.featureCard(feature))
	}
	n.Children = append(n.Children,
		ui.Node{ID: "automation-launch", Kind: ui.Button, Text: "Open workbench with local automation enabled", Disabled: d.busy, Activate: func() { d.launchFeature(featureByID("workbench"), true) }},
		ui.Node{ID: "replay-launch", Kind: ui.Button, Text: "Run deterministic UI replay", Disabled: d.busy, Activate: d.replayFeature},
		label("suite-result", d.serviceResult))
	return n
}

func (d *demo) featureCard(feature featureDemo) ui.Node {
	return ui.Node{ID: "suite-card-" + feature.ID, Kind: ui.Column, Style: ui.Style{Gap: ui.Some(float32(6)), Padding: ui.Some(ui.UniformInsets(8))}, Children: []ui.Node{
		{ID: "launch-" + feature.ID, Kind: ui.Button, Text: feature.Name, Disabled: d.busy, Activate: func() { d.launchFeature(feature, false) }}, label("description-"+feature.ID, feature.Description),
	}}
}

func (d *demo) replayFeature() {
	d.service("UI replay", func(ctx context.Context) (string, error) {
		dir, err := suiteDirectory()
		if err != nil {
			return "", err
		}
		out, err := exec.CommandContext(ctx, filepath.Join(dir, "replay"), "-batch", "64").CombinedOutput()
		if err != nil {
			return "", replayFailure(err, out)
		}
		return replayResult(dir, out)
	})
}
