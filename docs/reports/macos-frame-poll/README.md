# Frame ticket polling

`FrameMetrics.Poll(tag)` avoids copying the complete retained history on every
iteration of the loaded-grid completion wait. The native trace, timing regions,
scroll trajectory, admission limits and final full-history validation remain
unchanged. This reduces diagnostic overhead; it does not isolate a shipping
framework runtime improvement or complete Mac qualification.

## Contract and checks

Poll and Snapshot share the synchronized collector. A private reservation map
stores zero for pending tags and a one-based index for collected samples. Map
membership still refuses duplicate tags; history and reservations are never
removed. Poll returns one detached sample plus global dropped/pending status.
Unknown/pending tags are distinct from collected refused tickets. Optional source
details remain deeply detached. Nil, zero/copied recorders and zero tags are
refused. The consumer preserves refusal, drop, completion and generation checks,
and still inspects the final full snapshot.

Focused recorder/input race checks pass three times, and the complete offscreen
adapter race suite, adapter vet and loaded-grid example race tests pass. The first
new unit test incorrectly expected cancellation to collect before its input/owner
terminal. Its failed receipt is retained. The corrected test verifies that the
ticket stays pending, then supplies a retirement terminal and checks the original
refusal survives. This changes the fixture, not the native trace contract.

## Local observer benchmark

Five sequential observations per method/history size use already-collected
ordinary records without an active native window. Median results are:

| Retained records | Snapshot ns/op | Snapshot bytes/op | Poll ns/op | Poll bytes/op |
| --- | ---: | ---: | ---: | ---: |
| 128 | 2,220 | 20,480 | 14.37 | 0 |
| 512 | 8,654 | 81,920 | 14.42 | 0 |
| 4,096 | 49,733 | 622,594 | 14.77 | 0 |

Snapshots report one allocation/op; polls report zero. This excludes native
collection work and source-detail copying, and is not a frame-latency result.

## Matched native comparison

The original optimized binary is built at
`61976ee557d46550b3aed8e2fa6b6cf1c5198a15`. The candidate changes four Go leaves:
the recorder, its two test leaves and the loaded-grid waiting loop. Each of four
stored-data cases uses five interleaved original/candidate process pairs,
128 guarded warmups and 120 measured events per process: 40 processes and
4,800 ordinary samples. All processes pass; paired geometry and callback work
match exactly. No profiles, source-detail diagnostics or external process
sampling run during measured frames. Code checks, builds and the local benchmark
finish before this campaign. Source receipts bind 710 selected source files and
6,198 prepared files; they are not a complete compiler/SDK/cache closure or
continuous proof of host isolation.

Pooled elapsed UI-work results, retaining all observations:

| Case | Original p95 / p99 ms | Candidate p95 / p99 ms | Median process allocation reduction |
| --- | ---: | ---: | ---: |
| Fixed 10k | 8.920 / 9.599 | 8.801 / 9.646 | 13.3% |
| Fixed 1m | 8.640 / 9.084 | 8.694 / 9.103 | 19.3% |
| Variable 10k | 9.021 / 9.702 | 9.028 / 9.738 | 18.0% |
| Variable 1m | 8.821 / 9.329 | 8.899 / 9.386 | 17.2% |

The original and candidate both fail the unchanged 4 ms p95 / 6 ms p99 targets.
Timing is mixed. Allocation observations include helpers and runtime activity;
the reduction cannot be assigned entirely to framework code or native/GPU
memory. Observer allocation/GC pressure changed between variants. UI-work totals
retain their existing elapsed-work definition, including pauses in charged
regions; they are not thread CPU, GPU execution or input-to-photon measurements.

## Evidence and remaining gates

Internal source, protocol and raw-evidence reviews precede one bounded Prism
review using Anthropic Claude Sonnet 5.5. The
[checkpoint](records/checkpoint.json), [archive](records/evidence.tar.xz) and
[member index](records/evidence-index.json) retain the original failed unit
fixture, corrected tests, exact source variants, producer scripts, terminal
receipts, all native observations and review results.

Full P6/table/tree latency, native allocation, integrated P11/endurance,
human input/VoiceOver, physical display/GPU cases and release remain open.
The campaign requires no input-source, VoiceOver, clipboard or persistent
preference changes. See the [timing contract](../../../specs/grid-performance.md).
