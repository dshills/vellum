# Current private SDK consumption — macOS

The clean framework at `238b3b619e5ef274c218cb62d0b348e44e3fcaf8` packages as
a private development SDK. A generated independent consumer passes race tests,
vet and native compilation after extraction into a path containing spaces. The
public API contract module also passes race tests and vet. Nine demo applications
and four documented replay/workspace commands compile from the same extracted
SDK. Those checks compiled the native executables without launching them. A
subsequent scoped native hello check is recorded below.

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


## Native public hello follow-up

The already compiled hello executable from that exact relocated SDK opens a
native window, reads back its public GPUCanvas rectangle and exits normally
after a one-second quiet interval. Its single `vellum.native.ready` record
confirms the fixture disabled geometry persistence and completed three further
Snapshot calls. The check uses the example's center-pixel color predicate;
it does not compare the full image or observe physical screen presentation.

The reported image is 630×344 pixels. Raw startup, readback and
heap values are retained as single-process observations, with no latency or
memory-budget qualification. The executable SHA256 is
`1d9d56dafeebd8019f3e210613a9bd326bcbebc8d407abf08f4018d2f7d16fa9`.

Runtime source remains `238b3b6`; `e88941b` is the documentation checkpoint
present in both clean checkouts during this test. All 7,411 SDK file hashes,
the private archive, compiled binary and current tracked/prepared source remain
unchanged. The owned process exits with status zero, both output streams reach
EOF, its group is observed absent, and the local campaign stop marker is sealed.

Internal prelaunch review blocked an initial controller whose error handling
could skip cleanup or lose terminal evidence. The corrected controller passes
six ordinary cases (five launched children), six record-parser refusal checks,
four additional actual-child fault checks and one isolated fake-child check.
Three finalization cases and two stale-output admission refusals pass separately.
The initial blocked controller and its inaccurate original child-count field
are preserved separately. No native application ran under that blocked version.
The accepted controller uses the shared native lease, universal/local stop
markers, a 45-second watchdog, bounded raw logs and independent cleanup attempts.

This is one public example's native startup/readback/shutdown check. It does not
launch the other packaged demos or qualify physical pointer/keyboard input,
accessibility interaction, GPU memory, sustained performance, public distribution
or full Mac completion. No input-source, VoiceOver, clipboard or OS preference
settings are changed. Existing stage 0075 Prism review remains the implementation
review; this test-only follow-up receives internal result/documentation reviews.

See [native result](native-smoke.json), [native evidence manifest](native-evidence.manifest.json)
and [native raw evidence](native-evidence.tar.gz). The earlier consumption report
and its original evidence archive retain their separate scope.
