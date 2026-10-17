# Isolated native GPU fixtures and settled startup targets

The independent [drawing](../../../examples/drawing/README.md) and
[3D viewport](../../../examples/viewport3d/README.md) examples pass their scoped
native assertions with the exact private source-built ARM64 GPU library. This
checkpoint isolates fixture window geometry and fixes a demonstrated startup
baseline error. Interactive example behavior and framework runtime libraries
are unchanged.

## Implementation

Both `--smoke` modes disable backend geometry loading/saving before app/window
creation, report the actual flag and preserve it through normal native teardown.
Normal return closes the application before restoring the previous process-local
value. Existing error/watchdog `os.Exit` paths terminate the process and bypass
defers. No OS input source, VoiceOver, clipboard or preference setting is changed;
geometry-file writes were not independently monitored.

The viewport waits for both initial targets to remain ready with positive sizes
and completed frames over the same sampled 100 ms interval. A change to either
frame/target/program-allocation identity restarts the interval. Both baselines are
then checked/captured together in the owner transaction immediately before chrome
mutation. Readiness loss, context cancellation and later drift remain failures.
No baseline is replaced after acceptance actions begin. This operational quiet
barrier does not guarantee that all OS work has drained. All original later
chrome/independence, pixel/depth, budget, fault-repair, resize and closure checks
remain. Eleven bounded owner-thread observations aid independent-window diagnosis.

## Retained failures

The initial viewport run exits 1 at its independent-resource assertion (175-byte
stderr). Its old diagnostic lacks compared values, so its specific cause remains
unassigned. A distinct diagnostic endpoint exits 1 earlier at chrome retention
(899-byte stderr): primary size changes from 5102×1330 to 792×498, Frames 1→2,
TargetGeneration 1→2 and tracked TargetBytes 244,283,760→14,198,976. BufferAllocations
remain 3 and ResourceGeneration 0. This proves that endpoint sampled provisional
primary geometry; source explains its size-only reconfiguration. It does not
prove the initial independent-window failure had the same cause. Both failures
and their exact source endpoints/raw logs remain failed and retained.

## Verification

Source is base `8e88634187f1ce9c80bc8d79858225faa13d2e7d` plus the bound nine
modified example/test/spec leaves. The framework/prepared runtime counterparts
still match 7,027 payload files in the retained clean private SDK from `3da753e`.
Five explicit local module origins select current framework/driver/core/typesetting
and the retained SDK's GPU capsule. This is not a newly generated SDK archive.

Both original examples pass model `go test -race -count=1 ./...`, vet and
race-enabled native compilation. The diagnostic and corrected viewport endpoints
each pass their own model race/vet/build. The corrected suite adds three startup
regression roots, including eight readiness subcases, with deterministic timeline
inputs covering provisional geometry, late changes in the second window,
resource transitions, interrupted readiness and harmless visits/readbacks.
Known Cocoa deprecation and duplicate `-lobjc` compiler warnings remain in raw logs.

Each of the four retained executables is ARM64 MH_EXECUTE built with Go 1.27.2,
race support and a declared minimum macOS 13.0. Each map selects 141 exact members
from static archive SHA `786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`;
none has a WGPU dylib dependency. This does not establish runtime support on an
older Mac version. The current source-specific binaries are:

| Fixture | Executable bytes | SHA-256 |
| --- | ---: | --- |
| Drawing | 68,802,306 | `2e51b7a3b4990e45227a255ccf9eb1d2c530d8a8b0cf7d88f54a0f406cc78905` |
| Corrected viewport | 68,804,594 | `28bc586da1e8e9101fa7ce74922b0c4e91e81ad50990f23697c7dbb95f98642b` |

Drawing exits 0 in 7.571 seconds of owned execution. Its source assertions cover
gradient/linear-alpha and transparent pixel probes, retained meshes/pan,
100,000-object schematic culling, a million-sample waveform reduced to 792
columns/1,584 triangles, a 4096² image/1,024 initial tiles, a single 65,536-byte tile
update, image/filter/placement retention, rebuild/resize, independent windows and
close refusal. Its 1,460-byte stdout has one success summary and observed geometry
flag true; 38-byte stderr is its deliberate budget error. All 487 packages and 7,468
selected source/module/model-test files remain equal after the
viewport-only correction. This reuses the exact drawing result, not a rerun.

The corrected viewport exits 0 in 5.954 seconds of owned execution. Initial
baselines are 792×498 and 630×468, both Frames 2/TargetGeneration 2. Chrome keeps
its frame/preparation/upload counts. Camera actions change pixels without
geometry reallocation; budget refusal retains the completed frame; controlled
Prepare/Draw failures retire partial resources and recover. Rebuild, three
resizes, empty draw, expired helper refusal and primary/secondary closure pass.
Across all eleven observations the independent view stays Frames 2, allocations
3, ResourceGeneration 0, Revision 0 and model releases 0. Its 12,493-byte stdout
has one summary/geometry flag true. The 123-byte stderr contains exactly its three
deliberate budget/Prepare/Draw diagnostics.

All four native processes exit and their owned groups are recorded absent, with
complete bounded logs and no ownership/finalization errors. Internal reviews
separately cover source semantics, build/linkage, native admission and results.
Nineteen earlier summary controls and 27 earlier pure guard/finalization
controls support unchanged validators/guards; they are reused, not newly rerun.
The [previous SDK checkpoint](../macos-source-gpu-sdk-native/README.md) retains
the earlier pure guard/finalization controls.

Prism reviewed all nine source/test/spec leaves (23,256 diff bytes) with one
fresh Anthropic Claude Sonnet 5.5 call, complete coverage and no exclusions.
Its CLI exits 1 with one retained medium finding and no discarded/suppressed
findings. Independent review refutes that finding on the actual Snapshot call
path: it changes only ReadbackBytes/SnapshotReadbacks, which the comparison
excludes. Both actual before/after pairs confirm those are the only changed
fields. Real later layout/frame drift remains a failure. The raw finding,
exit status and disposition are retained; no source edit or second provider call
is made.

## Evidence and remaining limits

[Result](result.json), [manifest](evidence.manifest.json) and
[evidence archive](evidence.tar.xz) retain the three source endpoints, producers,
reviews, raw build/native outputs, module origins, inspections and maps. Large
binaries, the private SDK archive and library are fingerprinted here and remain
in their private run roots. A separate durable local workspace checkpoint will
copy them after commit, together with Git commit/canonical-integration receipts.
Cache/temp data and unrelated user data are excluded.

These are source-assertion/scalar-summary and synchronous GPU-readback checks;
full image dumps and a physical pointer/keyboard session are not retained. They
do not qualify physical interaction/device loss, cross-display behavior,
compositor presentation, ordinary 4/6 ms frame targets, native/GPU memory,
endurance, human accessibility/input, service interoperability, full P4/P6/P9/P11,
complete Mac acceptance or redistribution/release rights. The earlier ordinary
latency failures remain failed. Provisional startup target allocation is not
native/GPU memory qualification. Installed GPU libraries remain unchanged; no
public SDK, signing, notarization, push or release is performed.
