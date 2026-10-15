# Bounded uniform background subtraction

Large nested widget backgrounds repeatedly filled pixels that a later opaque
child background immediately replaced. Owned stage 0064 reduces that CPU replay
work while keeping the original widget painter and captured commands intact.
It inspects at most eight following items and paints at most four uncovered
strips. The eligibility and fallback contract is in [rendering](../../../specs/rendering.md).

The optimization applies to unclipped CPU replay, including full scroll-cache
repaints. Strip-clipped replay and scene-packet capture keep their existing
paths. It adds no retained image, worker, wakeup or GPU policy change.

## Evidence and scope

The source base is `60eaad15165b75ac5b596e5583c9718342dfc552` with 6,217 prepared
dependency files. The final patch adds four Go files and changes only
`paint/renderers/rasterx/renderer.go` in that dependency: 6,221 final files.
All checks are offscreen or compilation only. The human is using the computer;
live GUI, native GPU, quiet-host frame and soak testing remain paused.

The real loaded-grid consumer owns 10,000 rows and 20 numeric columns. At
162 DPI, 64 warm and 1,024 profiled routed scroll inputs retained 486 materialized
cells, a 767×624 logical viewport, grid bounds `(9,32)-(1325,1106)`, and 1,089
callbacks including the initial layout callback. Every input checked measured
cell geometry and at least 400 visible cells. Pristine and candidate final
1500×1250 RGBA images have the same SHA256:
`e984e91c4c85be13c95ff713a0d09821077434dd94cbcada96a8e3b5754b8fdb`.
This profiling fixture also performs manual snapshots and observation work;
its elapsed time and CPU samples are not ordinary native frame measurements.

An identical fixed benchmark body ran against pristine and candidate renderers,
five samples of 256 replays each. It constructs six nested large backgrounds
and 486 small cell backgrounds before timing. Median CPU replay elapsed time
was 1,061,727 ns/op versus 346,036 ns/op; both report zero allocations and bytes
per replay. These are isolated rendering microbenchmarks on an occupied Apple
M4 Pro, run baseline then candidate. They establish neither a controlled native
speedup nor a whole-grid frame budget. The observed profiled fill reduction
supports investigating this optimization in the next quiet native comparison.
The existing 500-small-cell control benchmark has medians of 73,810 versus
75,087 ns/op (five samples of 512 replays; zero allocations), an observed 1.7%
increase. Occupied-host microbenchmarks do not resolve that small difference.

Original-dispatch differential tests check dirty nonzero-origin images with
padded stride and guard bytes; partial and complete overlaps; translucent Src
and opaque/translucent Over; source-bound clipping; masks and custom colors;
repeated borrowed-source chains; exact lookahead limits; custom Scanner.Clear
observers; and recovered malformed-path panic pixels. Direct intermediate-image
assertions prove the overlap was actually omitted. Separate cache tests require
three full repairs and 21 reused/strip repaints, plus explicit clipped replay.
Three injected defects are rejected: disabling subtraction, drawing the omitted
overlap, and crossing a custom scanner observer.

The full adapter, loaded-grid consumer, raster and scene-plan race suites pass.
The first full suites ran with the final implementation and three added Go
files (the helper and two tests); the last cache test file was then prepared and checked with the complete
raster suite. Preparation checks, vet and native compilation are recorded
separately. Native compilation does not execute the binary.
The upstream image assertions create 17 missing reference PNGs during the full
raster suites. Those generated images are preserved separately, then removed
from the prepared tree. They do not qualify visual correctness. All original
prepared source bytes remain unchanged; the final 6,221-file source census
excludes only those known generated outputs. The new differential pixel oracles
do not use the generated references. This evidence paragraph was added after
the code checks; implementation and test sources were unchanged.

Early profiling probes are retained as failed harness attempts: one int/uint64
comparison did not compile; two 162-DPI assertions assumed the wrong viewport
width or omitted the initial resize callback. Correcting those assertions did
not change production source. The 96-DPI and paint-command probes are diagnostic
only. They are not pooled with the successful 162-DPI workload.

## Review and qualification

Internal source/evidence review and one Prism checkpoint using Anthropic
`claude-sonnet-5-5` gate this commit. Before commit, the
[qualification](records/qualification.json) and [archive index](records/archive-index.json)
will be sealed to bind the exact source, preparation, commands, raw logs, profiles, injected
variants, reviews and provider coverage. Generated evidence receives a final
internal review after the provider review.

Native performance gain, current complete lifecycle, static idle/overnight,
native GPU behavior and memory, P4/P6/P9/P11, physical input/displays, human
accessibility and distribution remain unqualified. Existing 4/6 ms frame targets
are unchanged; the earlier native comparison still fails them. The failed
123-of-1,000-window current lifecycle run remains failed. This source checkpoint
does not advance those gates.
