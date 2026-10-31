package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func suiteDirectory() (string, error) {
	exe, e := os.Executable()
	if e != nil {
		return "", e
	}
	if filepath.Base(filepath.Dir(exe)) == "MacOS" {
		return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(exe)))), nil
	}
	data, e := os.ReadFile("build/latest-suite.json")
	if e != nil {
		return "", fmt.Errorf("run examples/demo-suite/run.sh from the checkout, or build the suite first: %w", e)
	}
	var selected struct{ Suite, Version string }
	if e = json.Unmarshal(data, &selected); e != nil {
		return "", e
	}
	if selected.Version != demoVersion() || !filepath.IsAbs(selected.Suite) {
		return "", fmt.Errorf("rebuild the focused demos for %s before source launch", demoVersion())
	}
	return selected.Suite, nil
}
