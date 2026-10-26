package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureModule = "example.com/contract"

func typed(t *testing.T, path, body string) *types.Package {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "/Users/private/API-source.go", "package sample\n"+body, 0)
	if err != nil {
		t.Fatal(err)
	}
	p, err := (&types.Config{}).Check(path, fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSourceCompatibilityChanges(t *testing.T) {
	cases := []struct {
		name, old, new string
		breaking       bool
	}{
		{"unchanged", "func F(x int) string { return \"\" }", "func F(x int) string { return \"\" }", false},
		{"parameter rename", "func F(x int) {}", "func F(count int) {}", false},
		{"add function", "func F() {}", "func F() {};func G() {}", false},
		{"remove function", "func F() {}", "", true},
		{"parameter type", "func F(int) {}", "func F(string) {}", true},
		{"result type", "func F() int { return 0 }", "func F() string { return \"\" }", true},
		{"optional field addition", "type S struct{ X int }", "type S struct{ X int;Y string }", false},
		{"remove field", "type S struct{ X int;Y string }", "type S struct{ X int }", true},
		{"field type", "type S struct{ X int }", "type S struct{ X string }", true},
		{"interface requirement", "type I interface{ F() }", "type I interface{ F();G() }", true},
		{"method removal", "type S int;func(S) F() {}", "type S int", true},
		{"constant value", "const Kind=1", "const Kind=2", true},
		{"constant type", "const Kind int=1", "const Kind int64=1", true},
		{"generic constraint", "func F[T any](x T) {}", "func F[T ~int](x T) {}", true},
		{"generic type constraint", "type S[T any] struct{ Value T }", "type S[T ~int] struct{ Value T }", true},
		{"generic widening requires review", "func F[T ~int](x T) {}", "func F[T any](x T) {}", true},
		{"generic parameter rename", "func F[T any](x T) {}", "func F[Value any](x Value) {}", false},
		{"named generic constraint", "type C interface{~int};func F[T C](x T) {}", "type C interface{~string};func F[T C](x T) {}", true},
		{"embedded generic constraint", "type A interface{~int};type C interface{A};func F[T C](x T) {}", "type A interface{~string};type C interface{A};func F[T C](x T) {}", true},
		{"nominal constraint review", "type A interface{~int};type B interface{~int};func F[T A](x T) {}", "type A interface{~int};type B interface{~int};func F[T B](x T) {}", true},
		{"generic parameter count", "func F[T any](x T) {}", "func F[T any, U any](x T) {}", true},
		{"root version value exception", "const Version=\"v1.0.0\"", "const Version=\"v1.1.0\"", false},
		{"root version type still checked", "const Version=\"v1.0.0\"", "const Version string=\"v1.1.0\"", true},
		{"root version removal", "const Version=\"v1.0.0\"", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := compare([]*types.Package{typed(t, fixtureModule, tc.old)}, []*types.Package{typed(t, fixtureModule, tc.new)}, fixtureModule)
			if got := len(r.Incompatible) > 0; got != tc.breaking {
				t.Fatalf("breaking=%v: %+v", got, r)
			}
		})
	}
}

func TestVersionExceptionCannotHideAnotherPackageOrRemoval(t *testing.T) {
	root := typed(t, fixtureModule, "const Version=\"v1.0.0\"")
	old := typed(t, fixtureModule+"/app", "const Version=1")
	next := typed(t, fixtureModule+"/app", "const Version=2")
	if r := compare([]*types.Package{root, old}, []*types.Package{root, next}, fixtureModule); len(r.Incompatible) != 1 || r.VersionValueChanged {
		t.Fatal(r)
	}
	if r := compare([]*types.Package{root, old}, []*types.Package{root}, fixtureModule); len(r.Incompatible) == 0 {
		t.Fatal("package removal hidden", r)
	}
}

func fixtureBaseline(t *testing.T, dir string, pkgs []*types.Package) baseline {
	t.Helper()
	paths := []string{}
	for _, p := range pkgs {
		paths = append(paths, p.Path())
	}
	info := baseline{Schema: 1, Module: fixtureModule, Version: "v1.0.0", SourceCommit: strings.Repeat("a", 40), Packages: paths}
	if err := writeBaseline(dir, pkgs, &info); err != nil {
		t.Fatal(err)
	}
	return info
}

func TestBaselineRoundTripHashAndNonOverwrite(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "baseline")
	old := []*types.Package{typed(t, fixtureModule, "const Version=\"v1.0.0\"; type S struct{X int};func (S) F() {};type Generic[T ~int] struct{Value T}")}
	info := fixtureBaseline(t, dir, old)
	pkgs, read, err := readBaseline(dir)
	if err != nil {
		t.Fatal(err)
	}
	if read.ExportSHA256 != info.ExportSHA256 || len(compare(old, pkgs, fixtureModule).Incompatible) != 0 {
		t.Fatal(read)
	}
	data, err := os.ReadFile(filepath.Join(dir, "root.export"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("/Users/private")) {
		t.Fatal("local source position leaked into baseline")
	}
	if err := writeBaseline(dir, old, &info); err == nil {
		t.Fatal("baseline overwritten")
	}
	data[len(data)-1] ^= 1
	if err := os.WriteFile(filepath.Join(dir, "root.export"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err = readBaseline(dir); err == nil || !strings.Contains(err.Error(), "hash") {
		t.Fatal("corrupted export accepted", err)
	}
}

func TestBaselineMembershipAndBounds(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "baseline")
	fixtureBaseline(t, dir, []*types.Package{typed(t, fixtureModule, "const Version=\"v1.0.0\"")})
	name := filepath.Join(dir, "manifest.json")
	data, _ := os.ReadFile(name)
	var info baseline
	_ = json.Unmarshal(data, &info)
	info.Packages = append(info.Packages, fixtureModule+"/missing")
	data, _ = json.Marshal(info)
	_ = os.WriteFile(name, data, 0644)
	if _, _, err := readBaseline(dir); err == nil {
		t.Fatal("manifest/package mismatch accepted")
	}
	file := filepath.Join(t.TempDir(), "large")
	_ = os.WriteFile(file, []byte("12345"), 0644)
	if _, err := regularBytes(file, 4); err == nil {
		t.Fatal("oversized baseline accepted")
	}
	linked := file + "-link"
	if err := os.Symlink(file, linked); err != nil {
		t.Skip(err)
	}
	if _, err := regularBytes(linked, 10); err == nil {
		t.Fatal("redirected baseline accepted")
	}
}

func TestCommandChecksUseReadOnlyBaselineAndRetainPackageSet(t *testing.T) {
	root := t.TempDir()
	_ = os.Mkdir(filepath.Join(root, "release"), 0755)
	_ = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+fixtureModule+"\n\ngo 1.25.6\n"), 0644)
	_ = os.WriteFile(filepath.Join(root, "api.go"), []byte("package sample\nconst Version=\"v1.0.0\"\nfunc F() {}\n"), 0644)
	p := policy{Module: fixtureModule, Version: "v1.0.0"}
	p.API.Status = "stable-v1"
	p.API.Packages = []string{fixtureModule}
	metadata, _ := json.Marshal(p)
	_ = os.WriteFile(filepath.Join(root, "release/version.json"), metadata, 0644)
	dir := filepath.Join(t.TempDir(), "baseline")
	fixtureBaseline(t, dir, []*types.Package{typed(t, fixtureModule, "const Version=\"v1.0.0\";func F() {}")})
	before, _ := os.ReadFile(filepath.Join(dir, "root.export"))
	digest := sha256.Sum256(before)
	var out bytes.Buffer
	code, err := run(context.Background(), []string{"check", "-root", root, "-baseline", dir}, &out)
	if code != 0 || err != nil {
		t.Fatal(code, err, out.String())
	}
	after, _ := os.ReadFile(filepath.Join(dir, "root.export"))
	if got := sha256.Sum256(after); got != digest {
		t.Fatal("check rewrote baseline")
	}
	_ = os.WriteFile(filepath.Join(root, "api.go"), []byte("package sample\nconst Version=\"v1.0.0\"\n"), 0644)
	out.Reset()
	code, err = run(context.Background(), []string{"check", "-root", root, "-baseline", dir}, &out)
	if code != 1 || err != nil {
		t.Fatal("breaking API did not fail", code, err, out.String())
	}
	// A policy shrink must not erase previously stable package coverage.
	// Use a valid matching export bundle with an extra old package.
	oldDir := filepath.Join(t.TempDir(), "old")
	fixtureBaseline(t, oldDir, []*types.Package{typed(t, fixtureModule, "const Version=\"v1.0.0\""), typed(t, fixtureModule+"/app", "type S int")})
	code, err = run(context.Background(), []string{"check", "-root", root, "-baseline", oldDir}, &out)
	if code != 2 || err == nil || !strings.Contains(err.Error(), "omitted") {
		t.Fatal("policy shrink bypassed guard", code, err)
	}
}

func TestCaptureRefusesDirtyOrHiddenAPISource(t *testing.T) {
	root := t.TempDir()
	runGit := func(args ...string) {
		t.Helper()
		args = append([]string{"-c", "core.hooksPath=" + filepath.Join(root, "no-hooks"), "-c", "commit.gpgsign=false"}, args...)
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = childEnv()
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatal(err, string(out))
		}
	}
	runGit("init", "-q")
	_ = os.Mkdir(filepath.Join(root, "release"), 0755)
	_ = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+fixtureModule+"\n"), 0644)
	_ = os.WriteFile(filepath.Join(root, "api.go"), []byte("package sample\n"), 0644)
	_ = os.WriteFile(filepath.Join(root, "release/version.json"), []byte("{}"), 0644)
	runGit("add", ".")
	runGit("-c", "user.name=API fixture", "-c", "user.email=fixture@example.com", "commit", "-qm", "baseline")
	p := policy{Module: fixtureModule}
	p.API.Packages = []string{fixtureModule}
	revision, err := cleanAPISource(context.Background(), root, p)
	if err != nil || len(revision) != 40 {
		t.Fatal(revision, err)
	}
	for _, flag := range []string{"--assume-unchanged", "--skip-worktree"} {
		t.Run(flag, func(t *testing.T) {
			runGit("update-index", flag, "api.go")
			if _, err := cleanAPISource(context.Background(), root, p); err == nil {
				t.Fatal("hidden API capture accepted")
			}
			undo := "--no-assume-unchanged"
			if flag == "--skip-worktree" {
				undo = "--no-skip-worktree"
			}
			runGit("update-index", undo, "api.go")
		})
	}
	_ = os.WriteFile(filepath.Join(root, "api.go"), []byte("package sample\nvar Changed int\n"), 0644)
	if _, err := cleanAPISource(context.Background(), root, p); err == nil {
		t.Fatal("dirty API baseline accepted")
	}
}

func TestPolicyRefusesModuleMismatchBeforeLoading(t *testing.T) {
	root := t.TempDir()
	_ = os.Mkdir(filepath.Join(root, "release"), 0755)
	p := policy{Module: fixtureModule, Version: "v1.0.0"}
	p.API.Status = "stable-v1"
	p.API.Packages = []string{fixtureModule}
	data, _ := json.Marshal(p)
	_ = os.WriteFile(filepath.Join(root, "release/version.json"), data, 0644)
	_ = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/other\n"), 0644)
	if _, err := readPolicy(root); err == nil || !strings.Contains(err.Error(), "go.mod") {
		t.Fatal("module mismatch accepted", err)
	}
}

func TestOriginIdentityAcceptsAliasesAndRefusesOtherDirectories(t *testing.T) {
	root := t.TempDir()
	if !sameDirectory(root, filepath.Join(root, ".")) || sameDirectory(root, t.TempDir()) {
		t.Fatal("directory origin identity")
	}
	file := filepath.Join(root, "file")
	_ = os.WriteFile(file, []byte("not a directory"), 0644)
	if sameDirectory(root, file) || sameDirectory(root, file+"-missing") {
		t.Fatal("invalid origin accepted")
	}
	link := root + "-alias"
	if err := os.Symlink(root, link); err != nil {
		t.Skip(err)
	}
	t.Cleanup(func() { _ = os.Remove(link) })
	if !sameDirectory(root, link) {
		t.Fatal("same physical directory refused through alias")
	}
}

func TestChildEnvironmentRefusesRedirectsAndCredentials(t *testing.T) {
	t.Setenv("GIT_WORK_TREE", "other")
	t.Setenv("GOPACKAGESDRIVER", "arbitrary-command")
	t.Setenv("ANTHROPIC_API_KEY", "test-not-a-secret")
	t.Setenv("GOFLAGS", "-modfile=other")
	values := map[string]string{}
	for _, entry := range childEnv() {
		k, v, _ := strings.Cut(entry, "=")
		values[k] = v
	}
	for _, k := range []string{"GIT_WORK_TREE", "GOPACKAGESDRIVER", "ANTHROPIC_API_KEY"} {
		if _, ok := values[k]; ok {
			t.Fatal("unrelated environment retained", k)
		}
	}
	if values["GOFLAGS"] != "-mod=readonly" || values["GOWORK"] != "off" || values["CGO_ENABLED"] != "0" {
		t.Fatal(values)
	}
}

func TestStableCgoCannotDisappearBehindBuildConstraints(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "api.go"), []byte("package sample\nconst Version=\"v1.0.0\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := rejectStableCgo(root, fixtureModule, []string{fixtureModule}); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{
		"//go:build never_selected\n\npackage sample\nimport \"C\"\nfunc OnlyCgo() {}\n",
		"//go:build never_selected\n\npackage sample\nimport `C`\nfunc OnlyCgo() {}\n",
	} {
		if err := os.WriteFile(filepath.Join(root, "api_cgo.go"), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		if err := rejectStableCgo(root, fixtureModule, []string{fixtureModule}); err == nil || !strings.Contains(err.Error(), "contains cgo file") {
			t.Fatalf("build-excluded cgo source was not refused: %v", err)
		}
	}
}
