// Command apicompat checks the named stable root API against a reviewed v1 baseline.
// It is a separate developer module; framework consumers acquire no dependencies.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"golang.org/x/exp/apidiff"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	"golang.org/x/tools/go/gcexportdata"
	"golang.org/x/tools/go/packages"
)

const maxMetadata = 1 << 20
const maxExport = 16 << 20

type policy struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	API     struct {
		Status   string   `json:"status"`
		Packages []string `json:"root_packages"`
	} `json:"api_compatibility"`
}

type baseline struct {
	Schema       int      `json:"schema"`
	Module       string   `json:"module"`
	Version      string   `json:"version"`
	SourceCommit string   `json:"source_commit"`
	Packages     []string `json:"packages"`
	ExportSHA256 string   `json:"export_sha256"`
	GoVersion    string   `json:"go_version"`
	GOOS         string   `json:"goos"`
	GOARCH       string   `json:"goarch"`
}

type result struct {
	BaselineVersion     string   `json:"baseline_version"`
	CurrentVersion      string   `json:"current_version"`
	Packages            int      `json:"packages"`
	Compatible          []string `json:"compatible_changes"`
	Incompatible        []string `json:"incompatible_changes"`
	VersionValueChanged bool     `json:"version_value_changed"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	code, err := run(ctx, os.Args[1:], os.Stdout)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "apicompat:", err)
	}
	os.Exit(code)
}

func run(ctx context.Context, args []string, out io.Writer) (int, error) {
	if len(args) == 0 || (args[0] != "capture" && args[0] != "check") {
		return 2, errors.New("usage: apicompat capture|check -root REPOSITORY -baseline DIRECTORY")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	rootArg := flags.String("root", ".", "framework repository root")
	dirArg := flags.String("baseline", "", "reviewed baseline directory (capture never overwrites)")
	if err := flags.Parse(args[1:]); err != nil {
		return 2, err
	}
	if flags.NArg() != 0 || *dirArg == "" {
		return 2, errors.New("baseline directory required; unexpected arguments refused")
	}
	root, err := filepath.Abs(*rootArg)
	if err != nil {
		return 2, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return 2, err
	}
	p, err := readPolicy(root)
	if err != nil {
		return 2, err
	}
	if args[0] == "capture" {
		revision, err := cleanAPISource(ctx, root, p)
		if err != nil {
			return 2, err
		}
		pkgs, err := load(ctx, root, p.Module, p.API.Packages)
		if err != nil {
			return 2, err
		}
		if err = checkVersion(pkgs, p); err != nil {
			return 2, err
		}
		info := baseline{Schema: 1, Module: p.Module, Version: p.Version, SourceCommit: revision, Packages: p.API.Packages, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
		if err = writeBaseline(*dirArg, pkgs, &info); err != nil {
			return 2, err
		}
		return 0, json.NewEncoder(out).Encode(info)
	}
	old, info, err := readBaseline(*dirArg)
	if err != nil {
		return 2, err
	}
	if info.Module != p.Module || semver.Major(info.Version) != semver.Major(p.Version) {
		return 2, errors.New("baseline module/major differs from current policy")
	}
	// Load every baseline package even if someone removes it from the current policy.
	// A changed policy must never silently shrink the compatibility check.
	for _, path := range info.Packages {
		if !slices.Contains(p.API.Packages, path) {
			return 2, fmt.Errorf("stable baseline package omitted from current policy: %s", path)
		}
	}
	current, err := load(ctx, root, p.Module, p.API.Packages)
	if err != nil {
		return 2, err
	}
	if err = checkVersion(current, p); err != nil {
		return 2, err
	}
	changes := compare(old, current, p.Module)
	changes.BaselineVersion = info.Version
	changes.CurrentVersion = p.Version
	changes.Packages = len(info.Packages)
	if err = json.NewEncoder(out).Encode(changes); err != nil {
		return 2, err
	}
	if len(changes.Incompatible) > 0 {
		return 1, nil
	}
	return 0, nil
}

func readPolicy(root string) (policy, error) {
	var p policy
	data, err := regularBytes(filepath.Join(root, "release/version.json"), maxMetadata)
	if err != nil {
		return p, err
	}
	if err = json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	if p.API.Status != "stable-v1" || !semver.IsValid(p.Version) || semver.Major(p.Version) != "v1" {
		return p, errors.New("expected an explicit stable-v1 policy and valid v1 version")
	}
	if err = validatePackages(p.Module, p.API.Packages); err != nil {
		return p, err
	}
	moduleData, err := regularBytes(filepath.Join(root, "go.mod"), maxMetadata)
	if err != nil {
		return p, err
	}
	parsed, err := modfile.Parse("go.mod", moduleData, nil)
	if err != nil {
		return p, err
	}
	if parsed.Module == nil || parsed.Module.Mod.Path != p.Module {
		return p, errors.New("release policy module differs from selected root go.mod")
	}
	slices.Sort(p.API.Packages)
	return p, nil
}

func validatePackages(mod string, paths []string) error {
	if err := module.CheckImportPath(mod); err != nil {
		return err
	}
	if len(paths) == 0 || len(paths) > 128 {
		return errors.New("stable package list must contain1..128 entries")
	}
	seen := map[string]bool{}
	for _, p := range paths {
		if err := module.CheckImportPath(p); err != nil {
			return err
		}
		if (p != mod && !strings.HasPrefix(p, mod+"/")) || seen[p] {
			return fmt.Errorf("invalid or duplicate stable package: %s", p)
		}
		seen[p] = true
	}
	if !seen[mod] {
		return errors.New("stable package list omits root Version package")
	}
	return nil
}

// The stable root surface is deliberately Go-only. Inspect even build-excluded
// files so CGO_ENABLED=0 cannot silently hide a new public cgo declaration.
func rejectStableCgo(root, mod string, paths []string) error {
	for _, path := range paths {
		dir := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(strings.TrimPrefix(path, mod), "/")))
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			source := filepath.Join(dir, name)
			data, err := regularBytes(source, maxExport)
			if err != nil {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), source, data, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imp := range file.Imports {
				if imp.Path.Value == `"C"` || imp.Path.Value == "`C`" {
					return fmt.Errorf("stable root package %s contains cgo file %s; this Go-only baseline cannot qualify its API", path, name)
				}
			}
		}
	}
	return nil
}

func load(ctx context.Context, root, mod string, paths []string) ([]*types.Package, error) {
	if err := rejectStableCgo(root, mod, paths); err != nil {
		return nil, err
	}
	cfg := &packages.Config{Context: ctx, Dir: root, Env: childEnv(), Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedModule}
	loaded, err := packages.Load(cfg, paths...)
	if err != nil {
		return nil, err
	}
	if len(loaded) != len(paths) {
		return nil, errors.New("package loader returned an incomplete stable package set")
	}
	wanted := map[string]bool{}
	for _, p := range paths {
		wanted[p] = true
	}
	var result []*types.Package
	for _, p := range loaded {
		if len(p.Errors) > 0 {
			return nil, fmt.Errorf("load %s: %v", p.PkgPath, p.Errors[0])
		}
		if p.IllTyped {
			return nil, fmt.Errorf("stable package has an invalid dependency graph: %s", p.PkgPath)
		}
		if !wanted[p.PkgPath] || p.Types == nil || !p.Types.Complete() || p.Module == nil || p.Module.Path != mod {
			return nil, fmt.Errorf("unexpected package/module origin: %s", p.PkgPath)
		}
		if !sameDirectory(root, p.Module.Dir) {
			return nil, fmt.Errorf("stable package escaped selected checkout: %s", p.PkgPath)
		}
		delete(wanted, p.PkgPath)
		result = append(result, p.Types)
	}
	if len(wanted) > 0 {
		return nil, errors.New("package loader omitted stable packages")
	}
	slices.SortFunc(result, func(a, b *types.Package) int { return strings.Compare(a.Path(), b.Path()) })
	return result, nil
}

func checkVersion(pkgs []*types.Package, p policy) error {
	for _, pkg := range pkgs {
		if pkg.Path() != p.Module {
			continue
		}
		v, ok := pkg.Scope().Lookup("Version").(*types.Const)
		if !ok || v.Val().Kind() != constant.String || constant.StringVal(v.Val()) != p.Version {
			return errors.New("root Version constant differs from current release metadata")
		}
		return nil
	}
	return errors.New("root Version package was not loaded")
}

func compare(old, current []*types.Package, mod string) result {
	report := apidiff.ModuleChanges(&apidiff.Module{Path: mod, Packages: old}, &apidiff.Module{Path: mod, Packages: current})
	var r result
	exception := versionChange(old, current, mod)
	for _, change := range report.Changes {
		if exception != "" && change.Message == exception {
			r.VersionValueChanged = true
			continue
		}
		if change.Compatible {
			r.Compatible = append(r.Compatible, change.Message)
		} else {
			r.Incompatible = append(r.Incompatible, change.Message)
		}
	}
	r.Incompatible = append(r.Incompatible, genericConstraintEdits(old, current)...)
	slices.Sort(r.Compatible)
	slices.Sort(r.Incompatible)
	return r
}

// The pinned apidiff revision does not report every constraint narrowing.
// Treat changed generic declarations conservatively as requiring review, even
// when the edit might be a compatible widening or dependent-parameter rename.
func genericConstraintEdits(old, current []*types.Package) []string {
	next := map[string]*types.Package{}
	for _, p := range current {
		next[p.Path()] = p
	}
	params := func(o types.Object) *types.TypeParamList {
		switch o := o.(type) {
		case *types.Func:
			return o.Type().(*types.Signature).TypeParams()
		case *types.TypeName:
			switch t := o.Type().(type) {
			case *types.Named:
				return t.TypeParams()
			case *types.Alias:
				return t.TypeParams()
			}
		}
		return nil
	}
	fingerprint := func(list *types.TypeParamList) []string {
		items := []string{fmt.Sprintf("parameters:%d", list.Len())}
		for i := 0; i < list.Len(); i++ {
			items = append(items, constraintDescription(list.At(i).Constraint(), map[types.Type]bool{}))
		}
		return items
	}
	var changes []string
	for _, a := range old {
		b := next[a.Path()]
		if b == nil {
			continue
		}
		for _, name := range a.Scope().Names() {
			if !token.IsExported(name) {
				continue
			}
			before, after := a.Scope().Lookup(name), b.Scope().Lookup(name)
			if after == nil {
				continue
			}
			if !slices.Equal(fingerprint(params(before)), fingerprint(params(after))) {
				changes = append(changes, a.Path()+"."+name+": generic constraints changed; compatibility review required")
			}
		}
	}
	return changes
}

func constraintDescription(t types.Type, seen map[types.Type]bool) string {
	qualifier := func(p *types.Package) string { return p.Path() }
	name := types.TypeString(t, qualifier)
	if seen[t] {
		return name + " (recursive)"
	}
	seen[t] = true
	defer delete(seen, t)
	text := name + " underlying=" + types.TypeString(t.Underlying(), qualifier)
	if face, ok := t.Underlying().(*types.Interface); ok {
		face.Complete()
		for i := 0; i < face.NumEmbeddeds(); i++ {
			text += " embedded={" + constraintDescription(face.EmbeddedType(i), seen) + "}"
		}
	}
	return text
}

func sameDirectory(a, b string) bool {
	first, err := os.Stat(a)
	if err != nil || !first.IsDir() {
		return false
	}
	second, err := os.Stat(b)
	return err == nil && second.IsDir() && os.SameFile(first, second)
}

// Only the documented root Version value may change. Its removal, type change,
// or a Version constant in any other package remains an incompatibility/error.
func versionChange(old, current []*types.Package, mod string) string {
	lookup := func(pkgs []*types.Package) *types.Const {
		for _, p := range pkgs {
			if p.Path() == mod {
				c, _ := p.Scope().Lookup("Version").(*types.Const)
				return c
			}
		}
		return nil
	}
	a, b := lookup(old), lookup(current)
	if a == nil || b == nil || !types.Identical(a.Type(), b.Type()) || a.Val().Kind() != constant.String || b.Val().Kind() != constant.String {
		return ""
	}
	return fmt.Sprintf("Version: value changed from %s to %s", a.Val(), b.Val())
}

func regularBytes(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("not a bounded regular file: %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("file grew beyond read budget")
	}
	return data, nil
}

func readBaseline(dir string) ([]*types.Package, baseline, error) {
	var info baseline
	data, err := regularBytes(filepath.Join(dir, "manifest.json"), maxMetadata)
	if err != nil {
		return nil, info, err
	}
	if err = json.Unmarshal(data, &info); err != nil {
		return nil, info, err
	}
	if info.Schema != 1 || !semver.IsValid(info.Version) || semver.Major(info.Version) != "v1" || len(info.SourceCommit) != 40 {
		return nil, info, errors.New("invalid baseline identity")
	}
	if _, err = hex.DecodeString(info.SourceCommit); err != nil {
		return nil, info, errors.New("invalid baseline commit")
	}
	if err = validatePackages(info.Module, info.Packages); err != nil {
		return nil, info, err
	}
	data, err = regularBytes(filepath.Join(dir, "root.export"), maxExport)
	if err != nil {
		return nil, info, err
	}
	hash := sha256.Sum256(data)
	if hex.EncodeToString(hash[:]) != info.ExportSHA256 {
		return nil, info, errors.New("baseline export hash differs from manifest")
	}
	pkgs, err := gcexportdata.ReadBundle(bytes.NewReader(data), token.NewFileSet(), map[string]*types.Package{})
	if err != nil {
		return nil, info, err
	}
	actual := make([]string, 0, len(pkgs))
	for _, p := range pkgs {
		actual = append(actual, p.Path())
	}
	slices.Sort(actual)
	expected := slices.Clone(info.Packages)
	slices.Sort(expected)
	if !slices.Equal(actual, expected) {
		return nil, info, errors.New("baseline export package set differs from manifest")
	}
	return pkgs, info, nil
}

func writeBaseline(dir string, pkgs []*types.Package, info *baseline) error {
	var data bytes.Buffer
	// Positions are deliberately omitted: baseline APIs need no local source paths.
	if err := gcexportdata.WriteBundle(&data, token.NewFileSet(), pkgs); err != nil {
		return err
	}
	if data.Len() > maxExport {
		return errors.New("API export exceeds baseline budget")
	}
	hash := sha256.Sum256(data.Bytes())
	info.ExportSHA256 = hex.EncodeToString(hash[:])
	if err := os.Mkdir(dir, 0755); err != nil {
		return fmt.Errorf("capture refuses an existing baseline directory: %w", err)
	}
	// The manifest is written last. Partial failed captures remain diagnostics,
	// never a successful baseline; check cannot accept one without both hashes.
	if err := os.WriteFile(filepath.Join(dir, "root.export"), data.Bytes(), 0644); err != nil {
		return err
	}
	metadata, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), append(metadata, '\n'), 0644)
}

func cleanAPISource(ctx context.Context, root string, p policy) (string, error) {
	paths := []string{"go.mod", ":(top,glob)*.go", "release/version.json"}
	for _, path := range p.API.Packages {
		if path != p.Module {
			paths = append(paths, strings.TrimPrefix(path, p.Module+"/"))
		}
	}
	args := append([]string{"status", "--porcelain", "-z", "--untracked-files=all", "--"}, paths...)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	cmd.Env = childEnv()
	status, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(status) > 0 {
		return "", errors.New("capture requires clean committed stable API sources and release policy")
	}
	// Index flags must not conceal a changed API when binding the baseline to Git.
	cmd = exec.CommandContext(ctx, "git", append([]string{"ls-files", "-v", "-z", "--"}, paths...)...)
	cmd.Dir = root
	cmd.Env = childEnv()
	indexed, err := cmd.Output()
	if err != nil {
		return "", err
	}
	for _, entry := range bytes.Split(indexed, []byte{0}) {
		if len(entry) > 0 && (entry[0] == 'S' || entry[0] >= 'a' && entry[0] <= 'z') {
			return "", errors.New("capture refuses hidden assume-unchanged/skip-worktree API paths")
		}
	}
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "--verify", "HEAD")
	cmd.Dir = root
	cmd.Env = childEnv()
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	revision := strings.TrimSpace(string(out))
	if len(revision) != 40 {
		return "", errors.New("expected a full Git baseline revision")
	}
	return revision, nil
}

func childEnv() []string {
	allowed := map[string]bool{}
	for _, key := range strings.Fields("PATH HOME USERPROFILE SYSTEMROOT WINDIR SYSTEMDRIVE APPDATA LOCALAPPDATA TEMP TMP TMPDIR LANG LC_ALL LC_CTYPE GOPATH GOCACHE GOMODCACHE GOMAXPROCS GOMEMLIMIT GORACE HTTP_PROXY HTTPS_PROXY NO_PROXY http_proxy https_proxy no_proxy SSL_CERT_FILE SSL_CERT_DIR") {
		allowed[key] = true
	}
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if allowed[key] {
			env = append(env, entry)
		}
	}
	return append(env, "GOWORK=off", "GOENV=off", "GOFLAGS=-mod=readonly", "GOTOOLCHAIN=local", "GO111MODULE=on", "CGO_ENABLED=0", "GOPROXY=https://proxy.golang.org,direct", "GOSUMDB=sum.golang.org")
}
