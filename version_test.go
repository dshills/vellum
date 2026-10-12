package vellum_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dshills/vellum"
	"github.com/dshills/vellum/automation"
)

func TestVersionMatchesReleaseMetadata(t *testing.T) {
	data, err := os.ReadFile("release/version.json")
	if err != nil {
		t.Fatal(err)
	}
	var release struct {
		Version    string `json:"version"`
		Automation struct {
			Protocol int      `json:"protocol"`
			Targets  []string `json:"targets"`
			OptIn    bool     `json:"opt_in"`
			Status   string   `json:"status"`
		} `json:"automation"`
	}
	if err := json.Unmarshal(data, &release); err != nil {
		t.Fatal(err)
	}
	if release.Version != vellum.Version {
		t.Fatalf("API version %q differs from release %q", vellum.Version, release.Version)
	}
	if a := release.Automation; a.Protocol != automation.Version || len(a.Targets) != 1 || a.Targets[0] != "darwin" || !a.OptIn || a.Status != "experimental" {
		t.Fatal("automation contract differs from release metadata", a)
	}
}
