package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func featureArguments(bundle string, feature featureDemo) ([]string, error) {
	base := ""
	if feature.ID == "documents" || feature.ID == "notes" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return nil, err
		}
	}
	return featureArgumentsAt(bundle, feature, base)
}

func featureArgumentsAt(bundle string, feature featureDemo, base string) ([]string, error) {
	args := []string{"-n", bundle}
	switch feature.ID {
	case "documents", "notes":
		if !filepath.IsAbs(base) {
			return nil, fmt.Errorf("absolute config directory required")
		}
		data := filepath.Join(base, "Vellum Demo Suite", demoVersion(), feature.ID)
		if err := os.MkdirAll(data, 0700); err != nil {
			return nil, err
		}
		args = append(args, "--args", "-data-dir", data)
	case "grid":
		args = append(args, "--args", "-variable-rows", "-row-selection", "-editable-cells")
	case "gallery":
		args = append(args, "--args", "-mac-widgets", "-native-titlebar", "-mac-system-appearance")
	case "forge", "reminders", "forma", "luma", "pulse", "placement", "workspace", "workbench", "editor", "loadedgrid", "loadedtree", "drawing", "imaging", "viewport3d":
		// These catalog entries deliberately use their ordinary startup behavior.
	default:
		return nil, fmt.Errorf("no launch policy declared for %q", feature.ID)
	}
	return args, nil
}
