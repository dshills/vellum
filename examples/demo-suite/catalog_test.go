package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEveryLauncherHasMatchingPackagedDemo(t *testing.T) {
	data, err := os.ReadFile("demos.json")
	if err != nil {
		t.Fatal(err)
	}
	var entries []struct{ ID, Name string }
	if err = json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	bundles := map[string]string{}
	for _, entry := range entries {
		if _, exists := bundles[entry.ID]; exists {
			t.Fatalf("duplicate build id %s", entry.ID)
		}
		bundles[entry.ID] = entry.Name + ".app"
	}
	for _, feature := range featureDemos {
		if bundles[feature.ID] != feature.Bundle {
			t.Fatalf("launcher %s expects %s, build provides %s", feature.ID, feature.Bundle, bundles[feature.ID])
		}
		delete(bundles, feature.ID)
	}
	delete(bundles, "hub")
	delete(bundles, "replay")
	if len(bundles) != 0 {
		t.Fatalf("packaged demos missing from launcher: %v", bundles)
	}
}

func TestFeaturePoliciesCoverCatalogAndIsolateNotebookData(t *testing.T) {
	base := t.TempDir()
	for _, feature := range featureDemos {
		args, err := featureArgumentsAt("/private/tmp/Demo.app", feature, base)
		if err != nil {
			t.Fatalf("%s: %v", feature.ID, err)
		}
		if len(args) < 2 || args[0] != "-n" || args[1] != "/private/tmp/Demo.app" {
			t.Fatalf("invalid launch: %v", args)
		}
		if feature.ID == "documents" || feature.ID == "notes" {
			path := args[len(args)-1]
			expected := filepath.Join(base, "Vellum Demo Suite", demoVersion(), feature.ID)
			if path != expected {
				t.Fatalf("notebook isolation: got %s want %s", path, expected)
			}
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0700 {
				t.Fatalf("private notebook directory: %v %v", info, err)
			}
		}
	}
	if _, err := featureArgumentsAt("/private/tmp/Demo.app", featureDemo{ID: "unclassified-new-demo"}, base); err == nil {
		t.Fatal("unknown launch policy silently admitted")
	}
}
