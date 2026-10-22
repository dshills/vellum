# Rejected private uniform-fill candidate

Do not adopt this prototype or run native tests for it. The frozen SDK, shipping
renderer, app and native libraries remain untouched. The helper/test source lives
under testdata so normal Go package discovery ignores it; apply.py only targets a
fresh private Core copy. Indexed evidence retains the exact experiment.

Hypothesis: replace opaque built-in uniform rectangle fills with a seeded row and
copy. Original-source inspection after the benchmark confirms Go1.25.6 already
copies rows; the proposed difference is first-row initialization and dispatch
checks. Three focused pixel methods (including1200 fixed-seed cases) pass against
stdlib, but bounded serial warm-cache benchmark ranges overlap: stdlib142556–162019
ns/op, prototype142023–147623ns/op, zero allocations both. No useful gain established.
The raw benchmark reports suffix14 despite launcher GOMAXPROCS2; do not call this
an application-thread-limit experiment, extrapolate to cold/raster/native latency,
repeat unchanged or adopt it based on these microbenchmark figures.

No new native run, release acceptance or performance gate change. Retain the
negative result and inspect actual original raster command work before choosing
another repair. Features frozen, experimental rendering/pacing disabled,33.3ms
gate unchanged, M3–M7 open. Review with Prism Anthropic Claude Sonnet5.5 and commit
this rejected private checkpoint/evidence; no production renderer modifications.
