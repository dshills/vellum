# Current private SDK consumption — macOS

The clean framework at `238b3b619e5ef274c218cb62d0b348e44e3fcaf8` packages as
a private development SDK. A generated independent consumer passes race tests,
vet and native compilation after extraction into a path containing spaces. The
public API contract module also passes race tests and vet. Nine demo applications
and four documented replay/workspace commands compile from the same extracted
SDK. Native consumer and demo executables were compiled without launching them.

## Archive and dependency identity

The private `v0.1.0-alpha.2` archive has SHA256
`94dd1a1de78085ae705d17e587128b83bff0bbf442fb9f1f12ed348d5ba56f32`
and contains 7,411 hashed files plus its manifest. All 7,310 selected source files
match the clean tracked/prepared inputs; 101 generated marker, inventory and
notice files match their recorded associations. The recorded before/after snapshots of both source checkouts contain identical
4,654 tracked and 6,246 prepared files.

Both owned stages 0074/0075, their checksummed patch files and the current
sceneplan encoder/test outputs are present. Local replacements for Vellum, its
Cogent driver, CogentCore and typesetting resolve to the exact directories inside
the extracted SDK. External Go modules use the existing cache and pass Go module
verification. This is a relocatable development SDK with external dependencies;
it is not an offline runtime installer. The supplied digest establishes integrity
against this local producer, rather than independent artifact authenticity.

## Compiled entrypoints

| Entry | SDK path | Build tags |
|---|---|---|
| drawing | `examples/drawing` | default |
| editor | `examples/editor` | default |
| grid | `examples/grid` | default |
| hello | `examples/hello` | default |
| imaging | `examples/imaging` | default |
| loadedgrid | `examples/loadedgrid` | default |
| loadedtree | `examples/loadedtree` | default |
| placement | `examples/placement` | default |
| viewport3d | `examples/viewport3d` | default |
| workspace | `examples/replay/cmd/workspace` | default |
| workspace-smoke | `examples/replay/cmd/workspace-smoke` | `vellum_native_checks` |
| managed-replay | `examples/replay/cmd/native` | `offscreen` |
| workspace-replay | `examples/replay/cmd/workspace-replay` | `offscreen` |

The first demo runner incorrectly assumed that every example module root was an
executable. Eight applications compiled successfully before its executable-
metadata check refused the replay root, which correctly builds as a Go library
archive. That runner result remains incomplete. A separate run compiled the
remaining viewport application and the four documented replay commands with
their actual build tags. The earlier eight successful builds were revalidated;
their raw logs, the library result and the failed metadata check are preserved.
No framework source fix or reclassification of that failed check was needed.

## Scope and review

Packaging and relocation used bounded owned child commands, read-only Go module
mode, `GOWORK=off`, local toolchain/cache settings and `GOMAXPROCS=2`. The full
source, preparation, SDK contents and archive identities stayed unchanged.
Metadata records the executable import paths; every compiled binary is hashed.
This provides compilation evidence for placement, drawing, imaging, the 3D view,
data views, editor, workspaces and replay consumers through the packaged APIs.
It does not verify their live Cocoa behavior, GPU execution, physical interaction
or performance. Existing compiler deprecation/linker warnings remain in raw logs.

Internal result and documentation reviews cover this evidence-only checkpoint.
No additional Prism call is made: shipping stage 0075 already received one
Anthropic Claude Sonnet 5.5 review in `29ec8c9`. Public namespace and Vellum's MIT
choice remain recorded, while the 89-module declaration inventory still marks
distribution audit pending. Native component/rights closure, public release,
signing/notarization, native/GPU memory and full Mac qualification remain open.

The private SDK, extracted source and compiled outputs remain outside Git. This
report retains their identities plus all raw command output, SDK manifest,
inventory, controlled environment and producer source. Nothing was published.
See [structured results](result.json), [evidence manifest](evidence.manifest.json),
[raw evidence](evidence.tar.gz) and the [SDK contract](../../../specs/release.md).
