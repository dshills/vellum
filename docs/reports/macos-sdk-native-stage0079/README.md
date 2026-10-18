# Current private SDK: four native minimum-Go cases

The exact private SDK from clean `abe67c6` now passes four serial native
source-assertion/readback cases on the local Mac: starter, drawing, 3D viewport
and imaging. These use the already verified Go 1.25.6 non-race binaries from the
[current SDK consumption checkpoint](../macos-sdk-stage0079/README.md). No source,
SDK, prepared dependency or executable is rebuilt or changed for these checks.

The calling checkouts remain clean at `9dc2d9c` at terminal observation; their
changes since SDK source `abe67c6` are documentation only. The artifact retains
its original source identity. Its 7,879 payload files plus manifest, saved archive,
four exact executable hashes and nine controller/helper pins remain unchanged.

| Case | Result | Bounded run seconds | Owned process group | stdout / stderr bytes |
| --- | --- | ---: | ---: | ---: |
| hello | Pass | 3.073588 | 52517 (absent) | 243 / 0 |
| drawing | Pass | 2.561991 | 52610 (absent) | 1460 / 38 |
| viewport3d | Pass | 2.454067 | 52701 (absent) | 12491 / 123 |
| imaging | Pass | 2.820113 | 52843 (absent) | 4044 / 0 |

These are complete fixture durations, not frame, input, startup or presentation
latency qualifications. Each process exits zero, reaches complete log EOF and
has its owned process group observed absent. A shared live-test lease is held
through cleanup, validation and terminal input checks. Each case has a fresh
stop/result/log scope; all four are sealed after completion. No failed native
case is hidden or retried.

## What executed

The starter verifies a nonempty GPU Snapshot, the fixture's center-color
predicate and three further successful readbacks, then closes after its short
idle interval. Its native summary reports actual `go1.25.6`. Snapshot/readback
readiness is not a first interactive presented-frame measurement.

Drawing exercises gradient/linear-alpha pixels, transparent output, retained
mesh pan, independent windows, clear, budget refusal, error callbacks, resource
rebuild and repeated resize. Its stored schematic has 100,000 source objects,
its waveform one million source samples, and its 4096-square image initially
uploads 1,024 tiles; a changed tile uploads 65,536 bytes. Those counters do not
measure native/GPU live memory. The one stderr line is the deliberately tested
budget diagnostic.

The 3D fixture checks camera/orbit/depth behavior, independent buffer namespaces,
controlled preparation/draw failures, resource rebuild, resize and closure.
Its three stderr lines are exactly the deliberate budget/preparation/draw faults.
It uses the corrected joint startup targets already present in the SDK rather
than qualifying provisional window geometry.

Imaging executes compute/HDR float targets and tone-mapped straight-alpha
readbacks: exposures `(213,156,0,128)` and `(231,188,0,128)`. Sync/async capture,
recursive/busy refusal, transparent pixels, admission refusal, independent
resources, resize, cancellation/rebuild/minimize and closure pass. All six normal
close predicates pass and canceled/closed capture callback count is zero. Actual
resize/delivery scheduling is retained as observed, not required to have one
particular ordering. Hello and imaging stderr are empty.

All four native summaries confirm process-local geometry persistence is disabled.
The source fixtures neither restore nor save user window geometry. No pointer or
key operator, clipboard, input source, VoiceOver, preference or external-message
operation is performed. Native GUI windows and actual GPU work do execute.

## Scope, provenance and limits

The previously verified module graphs select the SDK's own framework, driver,
prepared Cogent/typesetting and source-built GPU capsule. Preparation additionally
checks thin ARM64 executable headers and records `otool -L` output with no WGPU
dylib dependency. These checks do not establish an exact static archive-member
link map; no 141-member claim from an earlier build is reused here.

The saved SDK remains 29,816,117 bytes, SHA256
`0bce4f09f8a51a7b196b2bba57c92d1c3fc4eaa999093fdef5c50edbea7bb7f9`.
Its caller-supplied source-built GPU archive/notice provenance and distribution
limits remain those of the preceding checkpoint. Complete asset/generated-code
and runtime-subcomponent attribution and owner distribution decisions remain open.

The recorded local OS/display configuration accompanies the raw evidence. These
cases establish native fixture assertions, GPU readback and closure on this Mac.
They do not establish physical/human input, screen-reader/IME task completion,
external native services, cross-display behavior, presentation/photons, ordinary
frame targets, native/GPU memory, sustained activity/endurance, older macOS
execution, signing/notarization, public distribution or full Mac completion.
The binaries are non-race release builds; historical race-enabled SDK cases
remain separately identified, not rerun or relabeled.

Internal reviews cover the exact native plan, terminal results and final archive.
No additional Prism call is made for this source-unchanged evidence continuation:
the SDK implementation checkpoints and immediately preceding packaging checkpoint
already received scoped Anthropic Claude Sonnet 5.5 review. Original headless
cache/inventory failures remain in the preceding checkpoint. `records.tar.gz`
and `records.json` bind this campaign's exact plan, controller/validator/executor
sources, admission/reviews, owned receipts, raw stdout/stderr and final identity
proof. Executables, fonts and the separately saved SDK are excluded. Full M2–M7
acceptance remains open; four successful fixtures do not replace those gates.
