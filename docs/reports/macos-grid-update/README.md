# One recursive grid update per projection sync

Pooled grid cells and headers previously ran their updater and style work once
in placement and again in the final recursive host update. Placement now assigns
all text, geometry, visibility, selection and cursor properties before that
host traversal. The host update, private inline editor, layout/paint, raster,
upload, frame tracing and retirement paths retain their existing behavior.
There is no public API change or new cache of unchanged frames.

The actual native offscreen regression instruments retained Core updaters and
stylers around one owner-held sync. The original produces 1,626 assertions with
two passes instead of one, and no other failed assertions. The corrected code
passes three race repetitions at logical DPI 96, 120, 144 and 192. Each slot has
one updater/style pass; escaped text, rounded clips, frozen/scrolling panes,
selection and recycled cursor borders remain current. Eight detached software
RGBA frame hashes match the original in all three repetitions. These are
software frame comparisons, not GPU or physical display pixel comparisons.

The affected existing grid and frame-metrics offscreen race suite and both
default/offscreen native vets pass. This checkpoint does not repeat unrelated
portable/tree checks or claim a complete compiler/module-cache/system SDK
closure. The 6,128 prepared dependency files remain identical to the qualified
timing checkpoint.

## Local optimized release comparison

Three fresh fixed-10k native processes per build ran in interleaved order
baseline/candidate/candidate/baseline/baseline/candidate. Every process had
128 warmup scrolls and 120 measured samples. All six retained the unchanged
400 actually clipped cells, 50 rows, eight columns and 767×624 logical viewport
at scale 1.6875. Their normalized clip hashes match at every trajectory phase.
All native tickets complete, with one callback per sample and no dropped or
pending tickets. Neither build is race-instrumented or CPU-profiled. External
process observations run only after the completed end record.

These nearest-rank pooled values each contain 360 same-ticket samples:

| Elapsed UI work | Original | One host pass |
| --- | ---: | ---: |
| p50 | 18.37 ms | 15.80 ms |
| p95 | 19.55 ms | 17.07 ms |
| p99 | 20.41 ms | 18.00 ms |
| maximum | 21.88 ms | 18.96 ms |

Pooled p95 decreases by 12.68%. The three individual candidate p95 values are
17.13, 17.07 and 16.93 ms; original values are 19.29, 20.02 and 19.09 ms.
Reconciliation p95 changes from 8.80 to 5.91 ms. Raster/upload p95 changes from
8.66 to 9.03 ms; this trial does not optimize that path or demonstrate fewer
runtime wakeups. Independent phase percentiles are never added together.
Timing is elapsed work including GC, not OS-thread CPU or GPU completion.

The unchanged P6 targets, p95≤4 ms and p99≤6 ms, remain failed. This is an
exploratory A/B comparison, with fewer than five fresh processes per build and
only fixed-10k rows. It does not replace the required fixed/variable 10k/million
campaign, tree timing, memory/endurance, cold-cache, external accessibility or
physical presentation checks. It supports retaining this isolated removal of
proven duplicate work; it does not certify production readiness.

## Evidence and next work

`archive.json` maps byte-preserved local raw logs, regression proofs, selected
production/prepared hashes, build records, native sample records, machine
metadata and inert producers to this directory. Binaries are identified by SHA
and excluded from Git. The independent review and narrowly scoped Prism result
are recorded in `review/`; the review excludes the raw evidence archive from its
provider payload. Namespace and license remain `github.com/dshills/vellum` and
MIT; no public release, signing or notarization occurs here.

Next, investigate whether geometry publication causes a redundant owner refresh
for an already accepted projection, and measure raster/upload invalidation
separately. Changed scroll images must never be marked unchanged to avoid
uploads. Any further change must preserve geometry, style resets, field focus,
lifetime and current-generation frame attribution before repeating the same
workload and unchanged targets.
