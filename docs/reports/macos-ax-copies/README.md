# Accessibility publication copies

Owner accessibility publication now consumes fresh snapshot containers, groups
scoped grid records by indices and assembles the complete native envelope from
already validated JSON. The matched native campaign observes
0.97–1.10 MB less mean process-wide Go allocation per event.
Pooled p95 is lower in 6 of eight cases. All per-mode warm target checks fail for both artifacts.
The user reported concurrent computer use during this work. Background load
and host exclusivity were not controlled or measured, so these latency results
are shared-host diagnostic observations and are ineligible for performance
acceptance. CPU/GPU competition can affect timing; pointer/keyboard/focus changes
can affect live UI state. No cause is assigned to the earlier paging failure.
Process-specific Go counters do not include other applications' allocations,
but additional Vellum input/paint can change process samples. Structural wire,
ownership and helper-allocation checks remain separate from that timing scope.
Further live window/input/performance checks are deferred until the user says
the computer is free; code, offscreen checks and reviews may continue.

This advances the Mac performance work; it does not complete the framework's
sustained performance, native memory, human accessibility or distribution gates.

## Implementation

Only the ordinary syncAccessibility traversal transfers ownership of freshly
created node and handler containers. Ready publications reuse those containers.
Borrowed internal entries keep separate output and handler containers. Both
routes still detach last-completed cache records, action slices and the existing
mutable pointer fields. Value's existing scalar contract is unchanged; arbitrary
new deep-copy support or general concurrent publication is not introduced.

Scope grouping records indices instead of repeatedly copying large node records
as a group grows. Interleaved scopes retain original per-scope order. Pending
replacement still makes detached read-only output and removes both current and
retained handlers. Scope/incarnation/parent changes, removal and attachment
closure retain their existing retirement behavior.

Whole-node JSON encoding still precedes cache retention. A pending replacement
is encoded again before publication. The envelope adds only fixed delimiters and
a generated unsigned decimal generation around those exact encoded semantic
bytes; application strings are already encoded. Identical semantics still replace
callbacks without advancing generation. Actual changes retire old queued actions.
Native validation, full commits, object identity, notifications and C/queued
payload ownership retain their existing paths. No publication is skipped based
on an unchanged portable grid generation.

## Source and native checks

Internal source review binds six changed source/test/spec files with no findings.
Full offscreen driver and root race checks, driver vet and focused AX race checks
pass. Twelve native-only Cocoa envelope/selector/text/payload groups pass with no
skips, including the actual owned preparation and production envelope in
startup/ready/pending transaction validation. The separately tagged managed-grid
native selector is covered by the complete offscreen driver suite.

The new reference-wire test compares exact bytes with the former stdlib
RawMessage envelope for empty/nil records, escaped/invalid-UTF8/Unicode text,
scalar values, completed/pending grid records, a 400-cell fixture and generations
through uint64 maximum. Interleaved-scope and owned/borrowed lifecycle checks cover
input/cache separation, mixed transitions, callback replacement, stale queued
actions, rejected encoding, removal and closure.

Two actual general Cocoa accessibility consumer processes pass native controls,
roles, values, focus, modal/keyboard workflows, paging and workbench namespaces.
Four actual stored-grid processes pass fixed/variable 10k/million semantics,
completed geometry, bounded visible cells, native identity reuse/replacement,
current actions and attachment retirement. These are in-process native checks;
human VoiceOver, external clients and physical display workflows remain separate.

The evidence preserves three earlier failures. The first wrapper expected 13
native groups despite one selector requiring the offscreen tag; twelve selected
groups passed. A sandboxed app launch could not connect to macOS window services
and timed out. The first unsandboxed candidate control smoke failed waiting for
row-0 after the viewport showed row-1..4. The unchanged committed baseline passed,
and the same candidate binary subsequently passed both traced runs without a
production or assertion change. That failure's origin remains unproved and is
not presented as fixed, attributed to user input, or substituted by later passes.

One Prism review uses Anthropic claude-sonnet-5-5 over six staged implementation,
test and contract files: complete coverage and no findings. Runtime histories and
host metadata are excluded from its input. Final measurement/report evidence is
independently reviewed separately.

## Matched ordinary measurements

Both optimized binaries have successful source-bound build receipts. All prepared
files remain at the preceding static-outline checkpoint; no native patch stage
changes. The matrix contains baseline/candidate × CPU/GPU-selected × fixed/variable
row height ×10k/million stored rows, five repeats per case and 120 observations after
128 closed warmups. Artifact order alternates by repetition. All 80 native processes
exit normally and all 9,600 events validate, with exact paired and small/large-row
geometry. No profiler, native acceptance process, test or build overlaps this
ordinary campaign. Outliers and per-run distributions are retained.

| Mode / height / rows | UI p95 ms, before → after | UI p99 ms, before → after | Process Go MB/event, before → after |
|---|---:|---:|---:|
| cpu-fixed-10000 | 10.91 → 10.85 | 11.91 → 11.84 | 9.409 → 8.353 |
| cpu-fixed-1000000 | 11.81 → 11.19 | 16.67 → 14.05 | 8.937 → 7.966 |
| cpu-variable-10000 | 12.30 → 11.53 | 22.49 → 14.48 | 9.280 → 8.186 |
| cpu-variable-1000000 | 10.60 → 10.72 | 24.16 → 12.37 | 8.824 → 7.804 |
| gpu-fixed-10000 | 11.62 → 11.19 | 24.63 → 13.24 | 14.575 → 13.489 |
| gpu-fixed-1000000 | 10.97 → 12.07 | 23.93 → 14.82 | 13.964 → 12.863 |
| gpu-variable-10000 | 15.31 → 15.18 | 18.54 → 22.39 | 14.360 → 13.293 |
| gpu-variable-1000000 | 11.93 → 11.08 | 22.79 → 13.49 | 13.766 → 12.742 |

All per-mode warm target checks fail for both artifacts. The table's percentiles pool 600 observations per
artifact/case; they do not replace per-run or sustained acceptance. MB uses
1,000,000 bytes. Process-wide Go allocation includes measurement helpers and
concurrent adapter work; it does not measure retained, native or GPU memory.
GPU request metadata does not certify every scene's actual GPU eligibility.
The renderer remains opt-in and physical presentation remains unmeasured.

The producer checks artifact hashes, closed warmup guards and final sources/counts
before its measurement receipt. A final validator additionally requires observed
producer exit 0, recomputes every raw distribution and checks all 80 cases, geometry
and exact current source/prepared maps. Profiling is admitted only after that
terminal receipt and validator, with matching receipt/producer/artifact hashes.

## Separate diagnostics

Four additional native fixed-10k processes record baseline/candidate ×CPU/GPU
sampled allocation profiles,120 observations and 128 warmups each. Eight profiles
use MemProfileRate=524288. Differential pprof uses after-minus-before and a union
of matching AX publication stack names; overlapping inclusive functions must not
be summed. These profiles are diagnostic and excluded from ordinary latency.
Independent after-minus-before decoding measures AX-focused sampled allocation
at 388.30 → 277.21 MiB for CPU and 426.61 → 299.56 MiB for GPU, over each
120-event diagnostic. MiB uses 1,048,576 bytes, matching pprof's MB display scale.
This supports attribution to publication work; it is neither an exact per-frame
allocation nor retained/native/GPU memory. Raw byte totals and complete commands
remain in the archive.

Five isolated microbenchmark repeats compare a synthetic 400-cell fixture using
the original helper preserved through a temporary test-only overlay, the current
borrowed helper and the current owned helper. Prior preparation uses about 651 KB/op;
current borrowed preparation uses 281 KB/op and owned preparation 144 KB/op. The
validated envelope uses 131 KB/op and 2 allocations versus about 274–276 KB/op and 4
for the prior stdlib envelope. These isolate helper behavior, not native frame or
representative UI speedup. The overlay is never used in either ordinary binary.

Raw terminal/command/hash records, profiles, failed attempts, review certificates,
original/current source snapshots and inert recipes are archived. Executables,
compiler caches, app bundles and prepared dependency trees are excluded. External
compiler/module-cache/SDK/dynamic-library closure is not claimed. Native menus,
general clipboard, physical Finder drop, human input/accessibility, cross-display,
sustained/frame/idle/native/GPU/font memory, broad workbench/tree and public SDK
distribution acceptance remain open. Windows/Linux qualification is deferred.
