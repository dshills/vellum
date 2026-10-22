# Uniform-fill control: rejected before native testing

The proposed private opaque-uniform row-copy helper passes three pixel-oracle
methods, including1200 seeded Params.Render cases, explicit padded subimages and
built-in color/source-clipping/custom-color fallback checks. The frozen SDK,
shipping renderer, app and native libraries are unchanged.

The bounded serial warm-cache rectangle microbenchmark overlaps: stdlib142,556–
162,019ns/op, prototype142,023–147,623ns/op; zero allocations each. The raw runtime
suffix is14 despite the launcher setting GOMAXPROCS2, so no controlled application
thread-limit or latency claim follows. No unchanged retry is needed for rejection.

Inspection of pinned Go1.25.6 image/draw.go411–439 confirms the standard renderer
already copies a first filled row across the remaining rows. The initial scalar-
fill hypothesis was incomplete. The prototype changes only row initialization
and dispatch; no useful improvement is established. Do not adopt it or spend a
native run on it. The [protocol](../../../tools/macos-calibration/uniformfill/PLAN.md)
and indexed evidence retain source/tests/raw receipts and limitations.

Next inspect the actual original raster command stream rather than assume each
frame is a single uniform fill. This control does not attribute the observed
0.85ms raster time. Features/defaults/SDK/app remain frozen,33.3ms gate unchanged,
M3–M7 open. No Mac completion, ordinary equivalence or new UI feature is claimed.
