# Unchanged native grid refresh

This checkpoint reuses a completed native grid projection when its validated
model generation is unchanged. It also fixes a general retained-parent identity
comparison that treated an embedded accessibility-host Frame as a different
parent from the concrete native tree node on every refresh.

The implementation keeps source-health checks, legitimate style/theme/topology
updates, effective disabled and help-tooltip styling, constant-time canonical
scrollbar restoration on refresh, and the private editor's independent
selection/restore/preedit/focus lifecycle. An incomplete projection sync cannot
be reused. Exact built-in worker projections validate their borrowed input stamp
even when an empty or fixed-height view exposes unchanged constant metadata.
Unstamped source mutations still violate the immutable source contract.

## Correctness evidence

The original grid/reconciliation overlay produces 1,497 expected regression
assertions. The final candidate passes all five new native regression groups
three times under the race detector with unchanged source and prepared inputs.
These exercise actual managed refresh and completed-capture tails, zero Cell
reads and retained update/style traversals, changed model projections,
style/help/disabled/topology, private selection/restore/focus, and no-op/refused
scrollbar restoration. The fixture uses the software backend; Manual still
conservatively reports layoutPending after owner refresh, so this does not
promise that public Capture immediately accepts a no-op refresh.

A separate real embedded-host extension regression preserves the original
Mount parent argument and Sync callbacks while eliminating fake parent moves,
OnChildAdded and update/style work. Real style and parent changes still apply.
Its original reconciliation overlay fails and the candidate passes. The full
native adapter package passes the offscreen race suite and default/offscreen
vet. All portable virtual tests pass under race and vet; the new projection
health regression passes three times and its original overlay fails with 180
expected assertions. Eight final software grid frames at 96/120/144/192 DPI
match the previous checkpoint's 24 repeated pixel digests exactly.

Earlier native fixture assumptions, their unmodified failure logs, diagnostic
parent/update stacks, and strict producer-parser refusals are retained. Separate
source-bound adjudications preserve the raw logs rather than rewriting them.
The first portable full-suite attempt failed before compilation on a protected
default Go cache; its successful final run uses the writable Vellum cache.

## Release comparison

The scoped comparison uses three fresh optimized fixed-10k processes per build,
interleaved A/B/B/A/A/B, with 128 warm-up events and 120 measured events each.
The baseline is the previous single-host-update executable. The candidate has
only the grid, retained-parent and portable source-health production changes.
All 6,128 prepared dependency files retain their exact pinned bytes.

The pooled elapsed UI-work observations are nearest-rank percentiles of 360
samples per build. Times below are milliseconds; no outlier was removed.

| Build | p50 | p95 | p99 | Maximum | Mean | Sample SD |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Baseline `32e9ec9` | 15.785 | 17.233 | 18.954 | 71.056 | 15.994 | 3.025 |
| Candidate | 13.518 | 14.624 | 15.566 | 16.161 | 13.625 | 0.571 |

Pooled p95 decreases by 15.14%. Per-process p95 in execution order
A/B/B/A/A/B is 16.334 / 15.304 / 14.194 / 18.282 / 17.148 / 14.492 ms.
The 71.056 ms baseline outlier is retained; its cause is unclassified. The
baseline and candidate coefficients of variation are 0.1891 and 0.0419.

| Measured stage | Baseline pooled p95 | Candidate pooled p95 |
| --- | ---: | ---: |
| Reconciliation | 6.066 | 3.343 |
| Layout/paint | 2.972 | 2.857 |
| Raster/upload | 8.748 | 9.026 |

These independent stage percentiles are not additive. The change reduces
reconciliation work; raster/upload remains the largest measured stage and does
not improve in this comparison. Both candidate latency targets still fail.

Actual samples, geometry checks, stage timings and source/binary identities are
recorded with the completed comparison. The elapsed UI-work measure includes GC
and matching intermediate composition work; it excludes queue/surface waits and
actual backend acquisition/present calls. It is not physical display latency.
External process observation occurs only after the completed end record.

This is an exploratory optimization comparison, not the complete P6 campaign:
it has fewer than five processes per case and omits variable heights, million-row
cases, cold caches, native/GPU memory attribution and endurance. The unchanged
4 ms p95 and 6 ms p99 targets and all broader Mac acceptance gates remain in force.

## Next rendering investigation

Read-only renderer review identified an unconditional full-window transparent
sprite raster/upload even when there are no drawable active sprites. That is a
separate candidate optimization, requiring sprite activation/removal, composition
and pixel regressions plus release measurements. This checkpoint makes no
sprite or partial-texture-upload change and no process-wide wakeup claim.

## Evidence and review

[records-index.json](records-index.json) lists each raw record and SHA-256 in
[records.tar.gz](records.tar.gz). The archive preserves UTF-8 logs/JSON and inert
producer/source snapshots, including earlier failures and parser refusals. Extract
it to a new temporary directory to inspect or reproduce the recipes; executable
scripts are stored with `.txt` suffixes. Original absolute paths in raw records
identify where the runs occurred; the index supplies their archived member paths.
The archive contains:

- `qualification/comparison.json`, `baseline-inputs.json`,
  `candidate-build-final.json`, machine metadata and all six native run logs/results;
- `qualification/native-regression/qualified-v1/adjudication-v1/qualification.json`,
  its original refused proof and unchanged regression logs/source overlays;
- `qualification/final-adapter-checks/qualification.json`, full native race/vet
  logs and the original embedded-parent overlay failure;
- `qualification/portable/qualification-v2.json`, original overlay failure,
  focused/full race/vet logs and initial protected-cache failure;
- `qualification/pixel-comparison/qualification.json`, current frame log and
  the previous checkpoint's pinned pixel-reference log;
- independent raw measurement verification, internal source review and
  producer recipes. Earlier exploratory proofs are historical, not substitutes
  for the final qualifications above.

[Provider review](review/prism.json), [input scope](review/inputs.json),
[process record](review/process.json) and [disposition](review/disposition.json)
record the requested Anthropic Claude Sonnet 5.5 review. Raw measurement/test
archives and historical source snapshots are excluded from its explicit staged
implementation/test/spec/report payload. The final [internal review](final-review.json)
and [audit](audit.json) bind current source/prepared bytes and archived evidence.
These are checkpoint qualifications, not a public distribution certificate.
