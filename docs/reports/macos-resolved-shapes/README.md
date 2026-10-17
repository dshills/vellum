# Resolved shaping reuse: rejected performance prototype

The private resolved-shaping cache passes correctness and shows positive reuse,
but fails the frozen adoption requirement in both renderers. It is not added to
the framework. Production source, stage0079 and the ordinary baseline remain
unchanged at `4b0cfd7`; no full Mac gate closes. No cache tuning or another small
shaping, row-fragment, AX, upload or culling variant is queued.

## Decision and ordinary measurements

Five alternating baseline/candidate pairs per CPU/GPU mode use the existing
loaded-grid runner, fixed 10k rows, 128 warmup events and 120 measured events per
run, normal runtime parallelism and no race, profile or source-detail observer.
All 2,400 ordinary raw samples are recomputed using integer nearest-rank p95/p99.
Acceptance uses per-run values, not pooled samples. The frozen adoption bar is
at least 0.4 ms lower median run-p95 and four of five paired improvements in each
mode, with no median-p99 or p99-failure-count regression. It does not replace the
full 4 ms p95 / 6 ms p99 matrix.

| Mode | Baseline median run-p95 | Candidate | Reduction | Pairs improved | Baseline/candidate p99 failures >6 ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| GPU | 4.596166 ms | 4.522458 ms | 0.073708 ms | 3/5 | 0 / 0 |
| CPU | 5.480209 ms | 5.404167 ms | 0.076042 ms | 4/5 | 3 / 2 |

Every ordinary p95 fails 4 ms: ten GPU and ten CPU runs. Preserve all tails,
including CPU candidate pair 3 (p99 6.264207 ms), even though the median run-p99
and failure count improve. Run-p95/run-p99 values below are milliseconds.

| Mode/pair | Baseline p95 | Candidate p95 | Baseline p99 | Candidate p99 |
| --- | ---: | ---: | ---: | ---: |
| GPU 1 | 4.538124 | 4.580332 | 4.698040 | 5.213460 |
| GPU 2 | 4.648501 | 4.497292 | 4.920211 | 4.586709 |
| GPU 3 | 4.589001 | 4.522458 | 5.522751 | 4.832416 |
| GPU 4 | 4.596166 | 4.494085 | 5.125208 | 4.699414 |
| GPU 5 | 4.605249 | 4.669333 | 5.517501 | 4.897374 |
| CPU 1 | 5.517458 | 5.345167 | 5.880919 | 5.633248 |
| CPU 2 | 5.456999 | 5.424709 | 6.100502 | 5.865207 |
| CPU 3 | 5.541208 | 5.980875 | 6.059836 | 6.264207 |
| CPU 4 | 5.446291 | 5.207459 | 5.722375 | 5.366290 |
| CPU 5 | 5.480209 | 5.404167 | 6.111585 | 6.099416 |

Separate novel-value controls add one baseline/candidate pair per mode (480
samples total). They pass the frozen regression guard, but cannot qualify five-run
performance or rescue ordinary adoption failure. Their p95/p99 values are:

| Novel mode | Baseline p95 / p99 | Candidate p95 / p99 |
| --- | ---: | ---: |
| GPU | 4.308707 / 4.521544 | 4.405000 / 4.712541 |
| CPU | 4.976417 / 5.344917 | 4.668039 / 5.372916 |

Ordinary allocations barely change: median of per-run p50 process-allocation
bytes is GPU 5,299,728 → 5,298,080 and CPU 3,693,752 → 3,691,264. These are not
live memory or the separate 1,000-control allocation acceptance workload.

## Selected boundary and input proof

Existing ordinary matrix logs show horizontal-scroll frames dominate the slowest
seven samples per run (GPU 29/35, CPU 32/35). This is phase correlation, not a
callsite or causality proof. `tail-analysis.json` records the exact prior archive
members and raw hashes; those logs are not duplicated in this archive.

The prototype retains resolved HarfBuzz `shaping.Output` after ordinary
font/bidi/script resolution, so recycled widgets can reuse previously visited
values. Font-map lookup/segmentation, wrapping, bounds, layout, paint and encoding
still run. No work is moved asynchronously outside the measured frame.

Keys prove the full binary rune context and range, exact fixed size, direction,
script, language, feature sequence, resolved immutable Font and requested/effective
Ppem. The existing HarfBuzz font LRU can retain an earlier mutable Face for the
same Font. `PreparedFace` performs its original admission/touch on every lookup;
keys include the actual retained Face's Ppem. Unsupported/variable/coordinate
inputs use original shaping. Retained entries own glyph arrays and key strings,
restore the current output Face and clone glyphs again on return. No mutable
Face, widget, model, callback or input slice is retained in the new cache.

Bounds are 8 MiB owned keys/glyphs plus conservative entry metadata, 2,048
entries, eight immutable Font references, and per-request bounds. This does not
bound immutable font backing tables or process memory. Explicit reset and owning
window retirement clear the cache; previously returned outputs remain detached.
The source overlay and tests are archived, not installed into prepared sources.

## Correctness, reuse and failures retained

- Focused independent original-HarfBuzz output regressions pass, including warm
  first-Face/LRU/eviction state, full-context/range/feature/size/font/Ppem changes,
  vertical/sideways directions, mutable output arrays, actual byte/entry/font
  eviction, unsupported CFF/coords and lifecycle retirement. A genuine Thai GPOS
  device-adjustment discriminator uses existing cached FreeSerif: changing the
  effective first-Face Ppem makes oracle glyphs differ and forces a cache miss.
  The font is not copied into the framework.
- Full shaper race, actual ManualWindow retirement, managed-grid pixel/oracle
  regressions across scales 96/120/144/192 and batches 1/7/64, and complete
  offscreen driver race pass. Driver race command takes 260.337 seconds. Separate
  default driver vet and native/diagnostic builds pass.
- Default upstream vet fails on inherited `run.go:36` unkeyed `textpos.Range`.
  The original failure is retained; affected shaper/typesetting vet passes with
  `-composites=false` explicitly. This is qualified vet evidence. Existing Cocoa
  deprecation and duplicate objc linker warnings remain in raw logs.
- Focused-v2 fails because font discovery seeded candidate HarfBuzz first-Face
  state while the oracle started empty. Controlled fixtures are corrected to
  equal empty initial state, without a production change. Final focused-v5
  passes and strengthens coordinate refusal to unchanged Hits/incremented Refused.
- Actual native GPU tests pass without skips, positively assert cache hits and
  preserve the existing text/path GPU readback and independent CPU oracle. The
  separate offscreen GPU package pass alone is not actual GPU evidence.
- Each CPU/GPU diagnostic run has 120 measured samples and 15 horizontal frames.
  Every horizontal frame satisfies at least ten eligible requests and an 80%
  hit fraction; aggregate horizontal reuse is 2,636 hits and zero misses per
  mode. Counters surround actual dirty-text sizing in a separate diagnostic
  binary; no filtered environment flag controls them. These clocks are excluded
  from performance acceptance.

## Execution and evidence boundaries

The initial native plan is rejected before execution for a terminal prerequisite
validation gap. Its source and review are retained. Corrected v2 validates exact
terminal GPU/diagnostics result, plan hash and prepared identity before dependent
phases. Independent internal review approves 17 source/controller/dependency pins,
three binaries and clean/prepared roots. All final phases complete before the
90-minute reassessment deadline. Source/prepared/binary identities remain equal
afterward; both C/T roots are clean at the recorded terminal snapshots, before
these documentation edits. Sampled process inspection finds no matching
test executable, but runner cleanup attempts plus exit-zero are not a strict
owned-descendant/group absence proof.

This uses existing programmatic grid events and native GPU assertions. No physical
pointer/placement, visible presentation, IME, VoiceOver, display transition or
system preference qualification follows. No input-source, clipboard or preference
changes are made. Hardware is the recorded local M4 Pro/macOS 26.6.2 configuration;
display metadata is retained per campaign. Baseline and candidate geometries match.

`records.tar.gz` and `records.json` retain the exact private source/test overlays,
controllers, internal reviews, original failures, command logs, frozen plan/pins,
raw native logs and assessments. Key members are `assessment.json`,
`comparison-result.json`, `ordinary/{gpu,cpu}/{1..5}-{baseline,candidate}/native.log`,
`novel/{gpu,cpu}/{baseline,candidate}/native.log`, `diagnostic-assessment.json`,
`diagnostics/{gpu,cpu}/reuse.json`, `gpu.stdout` and `post-native-proof.json`.
No executable, SDK archive or font copy is included. One scoped Anthropic Claude
Sonnet 5.5 Prism review covers the completed rejection documentation; internal
source, execution, results and archive reviews are retained. M2 native, full M3,
M4, human M5, final SDK/rights M6 and M7 remain open.

The fresh Prism call covers all three documentation files in one chunk (13,391
reviewed bytes): zero retained, one discarded medium and zero suppressed findings.
Finding `8893c5787f75fbf8` questions the warm-state fixture; the provider's quoted
evidence does not match. Independent disposition examines the existing
`TestVellumResolvedShapesWarmEffectiveFaceAndEviction` and
`TestVellumResolvedShapesDeviceAdjustedEffectiveFace`: each independently warms
the unchanged oracle and candidate, then compares same-Font/different-current-Face
requests from nonempty first-Face state. The real device fixture requires changed
oracle glyphs on effective-Ppem mutation and an additional candidate miss; the
warm test also covers mutation/eviction/zero cache. Both pass in focused-v5 raw
output. Equal initial fixture setup does not replace these warm-state checks.
No new test, product edit or repeat provider call is justified for the rejected
prototype. Raw provider output and the internal disposition remain archived.
