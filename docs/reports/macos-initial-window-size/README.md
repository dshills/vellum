# Initial native window sizing — macOS

Source checkpoint `c1343ffed992988fd5bd2d9b608514ee03ba2d34` makes explicit `app.Window.Width`/`Height`
Dp initial client-size requests. They bypass the adapter's former permanent
`Body.Min` and upstream screen-fraction sizing. Saved geometry and OS limits keep
precedence. Conversion, constraints and native creation share a chosen screen
snapshot before Show. See [the contract](../../../specs/windows-services.md).

## Verified

| Request (Dp) | Initial client/render pixels | Later client pixels | Separate AX outer frame |
|---|---|---|---|
| 320 × 240 | 324 × 243 | 162 × 121 | 324 × 275 |
| 640 × 400 | 648 × 405 | 324 × 202 | 648 × 437 |

Both native processes selected screen 0, LG HDR DQHD, 5120 × 1440 raw pixels,
DPR 1 and logical DPI 162. The observed request conversion is `Dp × 162 / 160`.
Each actual creation-options record equals the initial native client and render
size. AX outer frames include decorations and use macOS screen coordinates;
they are recorded separately and never inferred to be physical pixels. Both
race-instrumented native cases and exact-PID/executable/title read-only AX
clients exited 0; all four owned process groups were reaped and absent. Source,
prepared bytes, reviewed pins and log EOF/hash checks stayed stable. No physical
input, AX actions, input-source/VoiceOver/clipboard or OS-setting changes occurred.

The live build contains the exact candidate source plus an explicitly recorded
logging-only desktop overlay for the chosen screen/options before GLFW. Its Go
metadata is the dirty candidate based on b3c1c2, not a clean binary rebuilt at the
new commit. The original source, overlay, binaries' hashes, commands, raw runtime
records and review pins are retained. Compiled executables are excluded from the
archive with their hashes and sizes. The first build remains identified as unused. An earlier two-case campaign
passed; its raw results remain preserved separately. After the provider review,
an adjacent fixture-startup gap was fixed by waiting for native queue publication.
The final fixture was rebuilt and reran the same two validations in a new reviewed
campaign. No runtime receipt or validation was overwritten.

## Automated checks and review

Final focused offscreen race tests exercise the actual creation-options seam,
DPI/DPR/zoom, distinct small requests, rounding, tiny fractional scaling, extreme
finite sizes through final native integer conversion, saved precedence, legacy
unset/fullscreen behavior, independent sizing contexts, and actual adapter body
stylers. Fresh owned preparation verifies stage 0074 input/output/patch hashes.
Adapter-package offscreen race (106.23 s), portable module race (26.65 s), module vet
and preparation tests passed before the final generator tag and diagnostic-only
fixture guards. Focused-v4 race (35.23 s) covers the unchanged sizing implementation. The fresh
v3 native build, native vet and two native cases bind the final queued fixture.

Earlier focused checks exposed a missing math32 import and an uninitialized
system-test app fixture; those logs remain preserved. Internal review then found
a fractional-DPR zero-size and float32 integer-overflow boundary. Final native
bounds and regression coverage resolve those findings. Fixture startup/cleanup
and evidence-producer terminal observations were also corrected before execution.
Internal source/readiness/result and final checkpoint reviews accompany one
Prism review using Anthropic Claude Sonnet 5.5 (12 files, 32,466-byte diff, one
chunk/call). Its two medium findings assumed non-idempotent Close and
unsynchronized native dispatch. Independent inspection refuted those stated
mechanisms: Close is mutex-protected/idempotent, synchronous queue acknowledgment
orders captured output, and SetWinSize marshals GLFW to main. The separate real
pre-publication fixture hazard was fixed and revalidated after that provider call.
Production adapter/patch/spec bytes were unchanged by this last fixture guard. See [result.json](result.json),
[evidence manifest](evidence.manifest.json) and [raw evidence](evidence.tar.gz).

## Limits and next gate

These are two small, minimal-content windows on the current display, followed by
programmatic resizing. Live comparisons admit integral macOS DPR 1/2; this run
observed DPR 1. Fractional scale and synthetic secondary-screen checks are pure
regressions. Physical drag/resize, cross-display movement/hotplug, content stress,
live saved-geometry restoration, screen readers and broad native interoperability
remain open. Manual replay retains its physical viewport and Workbench retains
its OS-client-unit later-resize path; neither was redefined by this change.

All prior 4 ms p95 / 6 ms p99 failures remain failures at their recorded viewport.
A corrected initial viewport changes the workload and does not demonstrate an
algorithmic speedup or retroactively qualify old binaries. The next performance
comparison must record actual initial/client/render dimensions and display scale.
Full macOS completion and public release remain open.
