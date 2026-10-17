# Adaptive scene tile bins — macOS checkpoint

Vellum now counts dense scene tile bins with rectangle differences and an
in-place prefix pass. Sparse batches keep direct counting. Exact original
operation, byte and tile-reference limits remain enforced before compaction;
serialized bytes, painter order and detached ownership are unchanged. The owned
dependency stage is 0075, committed as `29ec8c98cd9303d37cca0d035245627642e191cc`.

## Correctness and review

Preparation and its unit suite passed. The affected scene/compositor packages
and the root adapter package passed their offscreen race checks; vet passed.
These commands used exact source/prepared bytes.
A frozen original encoder checks admission and every serialized byte, including
dense/sparse branch edges, high counts, mixed masks/pixels/spans and public
Builder fixtures immediately below, at and above reference/encoded-byte limits.
Encoded-byte cases differ by four bytes, the smallest reachable alignment.
A deliberately wrong difference-corner sign fails the byte oracle.

Twenty-four actual retained-grid software frames cover fixed/variable heights,
scrolling and supported manual DPI 96/160/192. Original/candidate encoded and
pixel vectors match; integer encoded replay, packet replay and complete captured
pixels agree. This is procedural-source correctness coverage. Unsupported manual
DPI 162, an initial compile failure and earlier strict wrappers that failed on
17 generated reference PNGs remain preserved. Generated missing goldens are not
regression evidence; final fresh checks passed with exact prepared-file sets.

One Prism review used Anthropic Claude Sonnet 5.5. Its medium concern about
`t.Fatal` in `Application.Do` was independently refuted: `Do` directly invokes
the closure synchronously on the calling test goroutine. The original finding,
exit 1, complete nine-file payload and precise disposition remain in evidence.
Internal source, commit, native-readiness and result reviews are preserved.

Six explicit native GPU roots passed with actual execution/readback: integer
compositing, shared drawer ordering/fallback, text/paths, covered uniforms,
checked input refusal/recovery and three-batch shared-to-owned texture reuse.
Their GPU evidence applies to those fixtures. Ordinary grid measurements below
do not prove GPU completion for individual grid tickets.

## Matched ordinary native measurements

Clean optimized CGO binaries compare baseline `77da1dd` and candidate `29ec8c9`
with `GOMAXPROCS=2` and without overlays, race or profiler/detail flags. Only the
encoder production file changes in prepared dependencies; the other prepared delta adds tests.
Five repeat sets alternate CPU/GPU, fixed/variable and adjacent build order.
Forty included processes warm 128 routed events and measure 120 each: 4,800 frames, with
2,400 complete per-sample grid/native endpoint geometry comparisons.

The initial campaign stopped when a draft report edit changed a reviewed input
hash. Its source/binary/native exit were unchanged, but the guard-failed result
remains failed. Two GPU checks and ten ordinary runs in complete pairs were
independently revalidated and carried forward; a reviewed continuation freshly
ran both members of the interrupted pair and the 28 unexecuted cases. Forty
ordinary runs enter this comparison. All 42 physical ordinary runs are preserved:
one unpaired valid run and one guard-failed run are excluded, totaling 44 native
processes including the two GPU checks. No original result was relabeled.

Both cohorts use the corrected current viewport, the same million stored rows
and unchanged minimum 400 clipped cells / 50 rows / eight columns. This is a
matched implementation comparison; old viewport failures retain their status.
Times below are milliseconds, pooled across 600 samples per group.

| Build | Mode | Heights | P95 ms | P99 ms |
|---|---|---|---:|---:|
| baseline | GPU requested | fixed | 6.169 | 6.644 |
| candidate | GPU requested | fixed | 6.057 | 6.652 |
| baseline | GPU requested | variable | 6.078 | 6.337 |
| candidate | GPU requested | variable | 5.952 | 6.257 |
| baseline | CPU | fixed | 6.951 | 7.359 |
| candidate | CPU | fixed | 6.784 | 7.303 |
| baseline | CPU | variable | 6.710 | 6.947 |
| candidate | CPU | variable | 6.853 | 7.199 |

Per-pair candidate/baseline p95 ratios show repeat variation. Lower is faster;
these are total UI-work ratios and do not isolate encoder cost. CPU control
pairs also vary in both directions, so these results do not establish a general
native frame-speed improvement.

| Mode | Heights | Median ratio | Pair range |
|---|---|---:|---:|
| CPU | fixed | 0.963 | 0.960–1.012 |
| CPU | variable | 1.008 | 0.986–1.071 |
| GPU requested | fixed | 0.982 | 0.945–0.994 |
| GPU requested | variable | 0.983 | 0.971–0.991 |

The unchanged 4 ms p95 limit fails in 20/20 baseline and
20/20 candidate processes; 6 ms p99 fails in
18/20 and 18/20. Complete table/tree performance
qualification remains open. Synthetic dense encoder microbenchmarks had
24–34% lower medians; sparse medians were 4–6% higher with substantial ordering
variation. Those separate results do not establish general frame gains.

The preceding two source-cost diagnostic processes are also preserved. Their
nested raster/copy/capture/encode/lease/resource/input intervals describe attempted
host work; they must not be added or treated as GPU execution/transfer time.

## Scope and evidence

Every owned process was reaped and observed absent. Reviewed source, preparation,
binary and launch bindings, bounded streams/EOF, watchdogs, exclusive lease and
human-return stops governed collection. All native campaigns described here are sealed. No OS input
source, VoiceOver, clipboard, power or permission settings were changed.
Grid scrolling entered the native event deque programmatically; physical
pointer/key interaction and presentation latency remain unverified.

Native/GPU memory, the complete table/tree workload, 1,000 controls plus canvas,
endurance, cross-display, broader accessibility and full macOS qualification
remain open. No public release is qualified by this checkpoint.

See [structured results](result.json), [manifest](evidence.manifest.json),
[raw evidence](evidence.tar.gz) and [performance contract](../../../specs/grid-performance.md).
