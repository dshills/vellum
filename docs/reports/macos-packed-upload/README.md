# Rejected packed GPU input upload experiment

The current runtime remains the reviewed text-batch implementation at 6e366fc,
with prepared stage 0079 and five checked input writes. This experiment was
implemented, tested and reviewed, then reverted after ordinary native timing
failed to show a worthwhile gain. Candidate stage 0080 and its diagnostic-reader
changes exist only in the archived rejected-source.diff, not active preparation.

The candidate packed configuration and four encoded arrays into one GPU buffer
with five aligned, exact-length read-only binding ranges. The shader, scene
operations, admission and CPU/journal fallback remained unchanged. Padding was
zeroed and counted as attempted write payload; one real checked write replaced
five. Packed scalar scratch was renderer-owned, bounded to 1 MiB in addition to
existing 1 MiB encoding scratch. Deferred GPU retirement held no CPU scratch.

## Decision from ordinary native results

Five interleaved release pairs use fixed 10k rows, 400 nonempty clipped cells,
128 warm and 120 measured events per process. All paired geometry checks pass at
162 logical DPI/1519x1266 client, 477 retained cells/nine headers on this Apple M4 Pro /
LG HDR DQHD 60 Hz Mac. Runtime settings remain ordinary: no race detector, profiler,
compiler overlay, GOMAXPROCS or GOMEMLIMIT override.

| Run | Baseline p95 ms | Candidate p95 ms |
| --- | ---: | ---: |
| 1 | 4.945000 | 5.178168 |
| 2 | 5.235417 | 5.234249 |
| 3 | 5.188873 | 5.299748 |
| 4 | 5.088210 | 5.264583 |
| 5 | 5.474541 | 5.302708 |

Three of five candidate p95s are higher. Median run-p95 is 5.189→5.265 ms
(1.46% higher); observed median per-run median process Go allocation is
5,893,496→5,902,600 B (0.15% higher). These observations do not establish a
universal regression, but they do not justify the added copy and retained MiB.
All p95s still fail 4 ms. **Reject the candidate; no Mac gate closes.**

Separate source-details processes record 120 input captures per build, all GPU
accepted with zero fallback/error. Five writes become one; median input phase
is 171.4→145.9 microseconds. Attempted bytes rise 560,022→560,664 B due to alignment
padding, excluding unused GPU capacity. This small diagnostic phase reduction
cannot substitute for the complete ordinary result. These instrumented samples
are never pooled into ordinary timing or presentation/physical-input evidence.
The earlier three-pair campaign was stopped before the decoder/integration and
scratch-lifetime corrections; its abandoned raw records are separate.

## Correctness, reviews and restoration

Race unit coverage compares every config/source/range byte and zero padding,
binding and total limits, overflow-resistant admission, larger-than-cache
payloads, source-copy ownership and resize/release/drawer-detach cleanup.
Actual native GPU tests pass integer compositor/readback, shared-device order,
text/path pixels, rejected-write and recovery. Both the borrowed encoding and
packed write source are overwritten before waiting; readback remains correct.
A real buffer missing CopyDst rejects exactly one full padded write with zero
submissions; restoration recovers exact pixels. The Core compositor ticket test
passes exact input-write counts and nested accounting. The native command takes
10.325s. Fifty-four diagnostic-runner regressions, eleven preparation regressions,
release build and affected vet pass. Fresh candidate preparation matches all
6,262 files, aggregate d4aa3c3663dd7298cf69618edd27c42c110bfc67b3a0fc2e3fac158b36bfd57b.

Internal review found and resolved a stale Core five-write assertion and the
original lease-owned scratch overlap risk. Final source review finds no blocker.
One complete fresh Prism Anthropic Claude Sonnet 5.5 call reviews 8 staged files,
36,609 B in one chunk, no truncation. CLI exit 0 has **zero retained, one discarded
medium and zero suppressed**. The discarded concern alleges an old layout can
reach a bind group after upload failure. Independent review refutes it: errors
return before binding; layout is assigned before the checked write; stale
metadata after reserve failure is never consumed, and the next successful upload
replaces it. Ownership/budget remain tracked and fallback invalidates readiness.
No additional production edit, test or provider call follows. Rejection is based
on measured performance irrespective of the source review.

The baseline uses the previously verified release binary SHA
38a98dcaddb7855a2476943aa468ae6f8294f11f665e0b88946ea65e677fb9da;
the final candidate identity is in native-v2/native-comparison.json. Neither is
claimed as a fresh historical baseline rebuild. The candidate diff is preserved,
then all runtime, test, decoder and manifest edits are reverted. Fresh stage 0079
preparation equals canonical source: 6,260 files with aggregate
1d9861ac10a0d83f607a36fd70dc79f00ce51d427c792b30b6a1331c0c3451ba.
An endpoint query finds no owned benchmark executables. Strict per-PID/PGID
absence proof is not claimed; no system/input/accessibility preferences changed.

## Remaining bottleneck evidence

Separate committed-build profiles contain 600 CPU-profile events and 120 sampled
allocation events after 128 warmups. The sampled Go allocation delta is 693 MB:
syncAccessibility is 332.42 MB cumulative (47.97%); ScenePacket is 186.87 MB (26.96%).
Accessibility serialization alone accounts for 130.17 MB cumulative. These are
sampled process-wide observations, including fixture/runtime work, not native/
GPU memory or ordinary frame qualification. Nested cumulative entries are not
added. The CPU profile attributes much time to native wait/cgo; those entries
cannot be read as an additive frame-time budget.

Next investigate the synchronous accessibility snapshot/serialization/publication
path and full scene capture, preserving canonical semantics, pending/ready
transaction authority, ownership and fallback. The unchanged full Mac targets,
M2 native/M3/M4/M5/M6/M7 and final soak remain open. The archive contains raw
profiles, samples, source, commands, failures and reviews; it contains no copied
SDK or native executable.
