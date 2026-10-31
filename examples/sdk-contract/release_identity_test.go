package sdkcontract_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dshills/vellum"
)

// This consumer runs in both the checkout and the relocated SDK. The selected
// release metadata belongs to that same root, not a historical source tag.
func assertReleaseIdentity(t *testing.T) {
	t.Helper()
	// Both layouts place this package at <root>/examples/sdk-contract and
	// metadata at <root>/release/version.json; Go tests run in the package directory.
	data, err := os.ReadFile(filepath.Join("..", "..", "release", "version.json"))
	if err != nil {
		t.Fatalf("SDK layout requires ../../release/version.json: %v", err)
	}
	var release struct{ Version string }
	if err = json.Unmarshal(data, &release); err != nil {
		t.Fatal(err)
	}
	if release.Version != vellum.Version {
		t.Fatalf("release metadata out of sync with SDK: package=%q release=%q", vellum.Version, release.Version)
	}
	if !strings.HasPrefix(vellum.Version, "v1.") {
		t.Fatalf("SDK no longer on stable-v1 major line: package=%q", vellum.Version)
	}
}
