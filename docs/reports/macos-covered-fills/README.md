# macOS covered uniform fill checkpoint

Stage 38 skips an immediately overwritten rectangular clear during raster replay.
Software pixels and observable source calls match the original dispatch. The
isolated 500-cell CPU replay median falls from 129.916 to 73.617
microseconds (43.3% less elapsed work), with zero allocations in all ten
benchmark observations. Across the four native cases, pooled p95 UI work is 6.0–6.7% lower for the
candidate on this host/workload. Native timing results below are separate from
the microbenchmark;
40 of 40 ordinary processes fail the unchanged 4 ms p95 / 6 ms p99 targets.
Mac/P6 completion remains open.

## Implementation and checks

Only adjacent image Draw/Src then Draw/Over items with identical nonempty
rectangles and source positions, nil masks, exact nonnil Uniform sources and
built-in value colors qualify. The second color must be opaque. Checks run at
replay, so changed borrowed colors are respected. Captured items, painter
recording, gradient updates and operation order are retained. Unknown/custom
sources, translucent fills, masks and operation boundaries use original replay.
This changes CPU rectangle writes without adding a retained image, changing
native resources or implementing partial GPU uploads.

Twelve source-bound preparation/race/vet checks pass. The repeated new suite
executes the elision branch 1,254 times. Its pre-change dispatch oracle compares
whole dirty/strided RGBA buffers, 360 eligible color/coverage combinations,
borrowed-color changes, custom color calls, masks, transforms, blur/pixel/path
ordering and actual StandardBox painting at fractional scales. Decorated-to-plain
replacement, resize/regrowth, rounded-parent clipping and gradient Update/At
calls are covered. Three deliberately faulty overlays fail meaningful pixel or
call-count assertions. Eight actual retained-grid RGBA frames at 96/120/144/192DPI
match the Stage37 renderer overlay exactly. Baseline file/overlay hashes bind
before and after replay/builds and the comparison.

Fresh pinned preparation reproduces 6,140 files. Compared with the baseline
prepared map, only renderer.go changes and covered_draw_test.go is added.
Actual Cocoa 400-cell accessibility publication/identity/action/close checks
pass for fixed10k and variable-million data. These checks do not establish human
screen-reader behavior or display latency.

## Ordinary native comparison

Forty fresh ordinary processes retain 4,800 samples: four cases, five repetitions
per build, 120 samples following 128 guarded warmup callbacks. Baseline/candidate
order alternates within each case. Every sample has matching full geometry
across builds and paired 10k/million cases. Diagnostics are off; no observation
is discarded or selectively rerun. These are elapsed CPU-side UI-work
measurements; phase percentiles cannot be added and physical GPU execution and
compositor presentation are unavailable.

| Case | Build | p50 ms | p95 ms | p99 ms | Max ms |
| --- | --- | ---: | ---: | ---: | ---: |
| fixed-10000 | baseline | 12.870 | 13.714 | 13.936 | 14.512 |
| fixed-10000 | candidate | 12.023 | 12.860 | 13.103 | 13.317 |
| fixed-1000000 | baseline | 12.507 | 13.512 | 14.085 | 15.560 |
| fixed-1000000 | candidate | 11.761 | 12.613 | 12.922 | 13.528 |
| variable-10000 | baseline | 13.020 | 13.925 | 14.128 | 14.749 |
| variable-10000 | candidate | 12.180 | 13.097 | 13.497 | 15.537 |
| variable-1000000 | baseline | 12.739 | 13.526 | 13.958 | 14.419 |
| variable-1000000 | candidate | 11.878 | 12.720 | 13.333 | 13.533 |

Baseline release SHA256: `16c7276e76fa07d558480b65e7b5905302c4b57c1706549c86971b9725b65c06`. Candidate release SHA256:
`ed650c66156a4cf784be1db7f93583f22fc1fe3dff5f4392007ef792d602c2cb`. The baseline is a fresh build with only the
renderer replaced by exact Stage37 bytes. Selected source/prepared endpoint
maps bind both builds, tests and measurements; they do not establish a complete
compiler, SDK or ambient runtime closure. Raw per-run distributions and all
maxima are retained. Observed differences support only this host/workload;
a microbenchmark gain does not qualify overall warm latency. Table/tree
endurance, broader native/font/GPU memory and professional Mac acceptance remain
open. Initial baseline CPU/source-cost diagnostics are kept separately.

## Evidence and review

[records.tar.gz](records.tar.gz) and [records-index.json](records-index.json)
retain raw observations, final receipts, source snapshots, diagnostics and inert
producers. Earlier failed development/cache/generated-output and sandbox Cocoa
attempts are historical records, not accepted qualification. The full raster
suite's 17 generated PNG files are copied into evidence and removed from the
prepared tree before checking source endpoints. The checks-v3 producer note
explicitly distinguishes its end-state copy from loaded-script byte identity.

Internal source and protocol reviews approve the change; final receipt review
binds this report and completed measurements. One significant 11-file patch/test/
spec/report batch is reviewed by Prism using Anthropic Claude Sonnet5.5; raw
review records are in [review](review/process.json). [audit.json](audit.json)
binds reviewed sources, current prepared bytes and the archive. No public
release, push/tag or macOS preference change is included.
