# Live Mac checks after GPU ownership fixes

The current build completes 1,000 native window cycles with successful readback,
retirement and final tracked teardown. The earlier multi-gigabyte IOSurface growth
pattern does not recur in this fixture. Stored-table latency still fails the
unchanged 4/6 ms targets; Mac completion remains open.

## Build and scope

All runs use commit `1d91d848ac973e2b264c24e61c4a3beac34f7f08`, 710 selected
source identities and the same 6,198 prepared identities. Two fresh native
executables build successfully. Terminal receipts bind the clean source endpoint,
binary identities and log hashes. This is not a complete compiler, SDK, external
module-cache or static-native-library closure.

The user explicitly made the machine available for testing. Read-only host
metadata records 14 CPUs/64 GiB RAM and background indexing activity; the pre-test
load average is about 25, and the post-test one-minute average is about 12.
Human availability does not establish continuous host isolation.

## Window and memory observations

The existing window-cycle fixture passes all 1,000 open/readback/close cycles in
about 188 seconds while a survivor remains usable. The final native observation
has no tracked windows, accessibility bridges, pending callbacks or outstanding
service operations. The fixture disables geometry persistence, accesses no
clipboard and does not modify input-source settings.

Forty owned-process vmmap samples at cycles 0..975 record physical footprint
549.4–693.1 MiB, with a final sampled value of 591.6 MiB. IOSurface resident bytes
are 92.4 MiB in 39 samples and 102.6 MiB once. These are region observations,
not object/reference counts. The final physical sample is skipped to avoid racing
process exit; the native terminal report separately confirms cycle 1,000.

Post-GC RSS grows from 176.67 to 315.17 MiB, or 276.58 to 315.17 MiB after cycle
25. The final Go HeapAlloc is 65,747,656 bytes. Caches, tombstones and caller-held
objects remain separate retention costs. No claim that every allocator is flat
or every native/GPU resource has been reclaimed follows.

The external observer perturbs this diagnostic run. Its sampled current physical
footprint guard is 4 GiB; it does not bound unsampled peaks. Prior large-footprint
results are historical context from a different build, not a matched single-patch
comparison. Integrated P11, native/GPU allocator attribution and overnight
endurance remain open.

## Stored-table measurements

Forty sequential release processes retain all 4,800 ordinary samples. Every case
has five processes, 128 guarded warmups and 120 measurements at the prescribed
viewport with at least 400 clipped cells. Default-renderer processes run first,
then GPU-option processes. No interleaved mode-speedup claim is made.

Pooled p95 / p99 total recorded UI work in milliseconds; phase percentiles are
not added together:

| Case | Default scene | GPU scene option |
| --- | ---: | ---: |
| Fixed 10k | 8.061 / 8.903 | 7.945 / 11.124 |
| Fixed million | 7.952 / 8.402 | 7.719 / 7.905 |
| Variable 10k | 8.313 / 11.725 | 8.144 / 9.051 |
| Variable million | 8.014 / 8.234 | 7.830 / 8.327 |

Both modes fail p95 ≤4 ms / p99 ≤6 ms. Individual runs, maxima, variability,
geometry, source callbacks, process-wide allocation and end-only RSS observations
remain in raw results. Domain/projection memory and observation work do not become
framework-only allocation measurements. GPU execution, compositor presentation
and input-to-photon latency remain unavailable.

Four separate source-detail processes retain 480 main-scene records using the
deferred route, with raster/copy availability false and zero SetFromGoImage
attempts. These diagnostics are excluded from ordinary timings. The renderer
option can fall back on other scenes; these records do not count physical GPU
transfers/completion or qualify broader scene support.

## Review and evidence

Internal protocol review caught a nested process-group timeout hole before any
live launch. The corrected native child and vmmap observer share the outer
runner's owned group. The initial rejected protocol is preserved. Independent
reviews recheck native cycles/memory and performance/route evidence. Prism reviews
this bounded documentation checkpoint with Anthropic Claude Sonnet 5.5.

The [checkpoint](records/checkpoint.json), [archive](records/evidence.tar.xz)
and [member index](records/evidence-index.json) preserve producers, receipts,
raw logs, host metadata and reviews. Binaries/caches are excluded. See the
[acceptance scope](../../../specs/macos-post-ownership.md). Human input/VoiceOver,
cross-display, broader services, GPU-loss policy, sustained performance, full P11
and distribution/release qualification remain open.
