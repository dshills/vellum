# Stored-tree frame work and updater correction

The tree adapter now supports the existing optional frame recorder through
EnqueueTreeScroll. It reserves an initialized native MouseScroll in the originating
FIFO, checks completed/current geometry and source state, and links actual input
to owner mutation, reconciliation and native composition. Ordinary input has its
existing path. Cancellation does not retract already delivered input; boundary
no-ops, pending indexes, disabled targets and missing/refused transitions cannot
become successful samples.

The independent loaded-tree consumer stores 10k/100k nodes with a huge branch and
separate deep chain. Initial construction/indexing and post-GC data memory precede
mounting and are reported separately. Wide and deep warm views use at least 64
clipped rows, compact 9-logical-pixel rows and 12-Dp Noto Sans. Every event
checks exact offset/generation, source stamp, viewport/bounds/scale, one callback
and matching completed recorder sample. The first scroll is not cold-cache evidence.
Native geometry persistence is disabled for the selected campaign.

The new native measurements exposed duplicated row updater/style work. Tree sync
updated every row before a full host update traversed those same rows again.
It now installs all row/footer values and states, then invokes one host update.
A retained-row mismatch conservatively marks the host dirty. Same-generation
pending, disabled/help/theme and ordinary geometry behavior remain covered.

The meaningful original-source regression produces 450 duplicate 2/2 updater/style
assertions. The final race regression passes 1/1 per retained text. All 12 detached
software RGBA records match exactly across four DPI values and scroll/selection/
pending phases. These are software snapshots, not new GPU/presentation claims.
Actual-input recorder tests also cover boundary refusal, pending admission,
cancellation after delivery and an earlier owner action disabling the target.
Manual rendering explicitly cannot qualify native frame timings.

The native comparison uses five optimized processes per build for each stored
size and wide/deep case, with alternating build order. Both builds have identical
consumer bytes; the baseline overlay changes only syncTree's updater pattern.
All 40 processes and 9600 warm samples pass independent validation of sequence,
offsets, generations, callback counts, native composition, work-region sums and
matching phase geometry/cardinality. Each has 71 clipped row frames. Mean per-run
p95 is reduced in all four cases:

| Stored nodes / view | Baseline p95 | Optimized p95 | Change |
| --- | ---: | ---: | ---: |
| 10k / wide | 9.304 ms | 8.486 ms | −8.8% |
| 10k / deep | 9.511 ms | 8.562 ms | −10.0% |
| 100k / wide | 9.377 ms | 8.455 ms | −9.8% |
| 100k / deep | 9.518 ms | 8.479 ms | −10.9% |

Recorded mean work is about 10–12% lower. Process-wide Go allocation per observed
interval is about 21–23% lower; that interval includes observers and other Go
runtime/adapter goroutines. Target p95 remains 4 ms and all 20 final runs fail it.
Results at the two stored sizes are similar at the same measured native workload;
this does not independently qualify every interaction or maximum-row behavior.
See [review summary](records/review-summary.json) for per-run distributions and
[the contract](../../../specs/tree-performance.md) for measurement boundaries.

Earlier preflights retained a 64-row guard and refused layouts with only 53 rows.
The selected compact typography/row configuration was established before pairing;
no smaller passing workload was silently substituted. Raw refusals remain archived.
The full adapter race run has unchanged adapter/root/prepared inputs; its broad
source-delta receipt flags a first-scroll JSON/comment rename in the independent
consumer. That raw flag remains with a scoped input resolution, and final consumer
checks/native builds cover the renamed output on the final source.

Prism's complete 15-file Anthropic Claude Sonnet 5.5 review returned two medium
findings. The linked summary/raw evidence is now committed with this report.
A missing initial tree binding may retry for at most five seconds while its logical
origin remains live; after the first successful inspection, ErrTreeClosed is
terminal. Startup/retirement race tests and two final 100k wide/deep native
preflights pass. The 40-run timing campaign predates only this error-path fix;
its consumer/source version and later preflights are recorded separately. Recorded
successful scroll/frame work is unchanged. Raw provider exit one and both findings
remain with explicit resolutions; no full provider repeat was required.

The unchanged table target, complete tree interaction/expansion/cold-cache and
maximum-visible-row cases, native/GPU allocation attribution, physical displays,
full P6/P11, overnight soak and Mac distribution remain open.
