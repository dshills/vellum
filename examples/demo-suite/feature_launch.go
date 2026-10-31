package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

func (d *demo) launchFeature(feature featureDemo, automation bool) {
	d.service("Open "+feature.Name, func(ctx context.Context) (string, error) {
		dir, err := suiteDirectory()
		if err != nil {
			return "", err
		}
		bundle := filepath.Join(dir, feature.Bundle)
		if _, err = os.Stat(bundle); err != nil {
			return "", err
		}
		if automation {
			return launchAutomation(dir, bundle)
		}
		args, err := featureArguments(bundle, feature)
		if err != nil {
			return "", err
		}
		// Context scopes only the launch helper; focused apps intentionally outlive the hub.
		if err = exec.CommandContext(ctx, "open", args...).Run(); err != nil {
			return "", err
		}
		return feature.Name + " opened in its own application. Close its windows when finished.", nil
	})
}
