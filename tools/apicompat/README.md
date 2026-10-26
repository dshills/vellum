# Stable root API compatibility guard

Developer-only module for the thirteen root packages named by
[the v1 compatibility contract](../../specs/stable-api.md). The framework and
native driver module graphs are unchanged. Dependencies are pinned to Go's
`x/exp/apidiff`, `x/tools` and `x/mod` versions already used by the prepared
repository. Use the Go version declared in the root go.mod.

From the checkout root:

```sh
(cd tools/apicompat && go test -race ./... && go vet ./...)
(cd tools/apicompat && go run . check -root ../.. -baseline ../../release/api/v1)
```

A check returns0 for a compatible surface,1 for an incompatible or conservative
review-required change, and2 for invalid/missing input or a loader error. Its
JSON lists changes. It compares all baseline packages, even if the current policy
tries to omit one, verifies the bundle hash/membership, and never rewrites the
baseline. Package origins must resolve inside the selected checkout. Controlled
child Go configuration disables workspaces, module edits, external package
drivers, unrelated credentials and cgo; it does not execute application code.
The stable root packages are Go-only: capture/check refuse a cgo import even in
build-excluded source files. Other platform-specific Go APIs are checked on the
CI host matrix. The tool uses public module resolution with readonly modules;
prepare the pinned tool dependencies in the cache before running offline. Custom
proxy/Go flag overrides are intentionally not inherited by the package loader.

The root `vellum.Version` constant may change to match valid current v1 metadata.
Only that value exception is allowed: removing it, changing its type, or changing
a constant in another package still fails. Ordinary function/field/method/type,
interface, constant and package removals are checked with the pinned upstream
semantic diff. Additional generic declaration checks are deliberately
conservative: the pinned upstream tool misses some constraint narrowings.
Generic constraint edits require compatibility review even when they could be
compatible widenings or dependent-parameter renames.

## Baseline ownership

The reviewed baseline belongs to the v1 contract. It was captured from9f05a37,
whose root API is identical to the packaged382e0e4 release. Its export data omits
local source positions. The manifest pins version, selected packages, source
revision, compiler/host and export hash. Capture refuses dirty/untracked stable
API paths or hidden Git index flags and refuses an existing output directory.
Only capture into a **new** reviewed location when adopting another baseline:

```sh
(cd tools/apicompat && go run . capture -root ../.. -baseline /absolute/new/baseline)
```

Do not regenerate v1 to silence a breaking result. Review the cause and preserve
the v1 promise; use a separately adopted major-version policy for incompatible
stable changes. The current checker deliberately accepts v1 policy only.
Capture failures can leave diagnostic partial output; a successful check requires
a matching complete manifest/export pair. Baseline hashes detect accidental
corruption; repository review remains the trust boundary for authorized changes.

## Scope

This first guard covers the **root module API only**. The native facade still
uses the public consumer and native lifetime/input checks; it is not covered by
this export snapshot. Runtime behavior, errors.Is contracts, data-format
compatibility, availability/default behavior and UI/accessibility qualification
retain their existing tests and review obligations. The source diff is an
approximation, with upstream correspondence/embedding/unsafe limitations.

The developer tool is excluded from private SDKs with the other repository-only
tools. Its baseline and contract can be included as release metadata. It neither
changes the shipped v1 app bytes nor broadens supported platforms or distribution
rights. Running this check on CI hosts qualifies the checker, not their native UI.
