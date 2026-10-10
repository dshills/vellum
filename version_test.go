package vellum_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dshills/vellum"
)

func TestVersionMatchesReleaseMetadata(t *testing.T) {
	data, err := os.ReadFile("release/version.json")
	if err != nil {
		t.Fatal(err)
	}
	var release struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &release); err != nil {
		t.Fatal(err)
	}
	if release.Version != vellum.Version {
		t.Fatalf("API version %q differs from release %q", vellum.Version, release.Version)
	}
}
