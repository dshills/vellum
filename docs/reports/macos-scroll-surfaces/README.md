# macOS ordered scroll-surface checkpoint

The native raster backend can reuse scrolling pane pixels at their original
painter position, guarded by detached actual paint operations. Pixel/race checks
pass. Both measured experiments regress stored-table latency, so final managed
rendering keeps ordinary replay by default. Driver.ExperimentalGridScrollReuse
and ManualConfig.ExperimentalGridScrollReuse provide an explicit experiment in
owned main scenes; separate modal/popup scenes retain ordinary replay. This
checkpoint does not complete Mac/P6 performance or recommend enabling the cache.

## Implementation

Each admitted group has complete opaque rectangular coverage and a bounded
private RGBA cache captured before later siblings. A toroidal mapping restores
the overlap with up to four Src copies and captures only disjoint exposed strips.
Full repair resets the mapping. Every operation affecting the old overlap must
equal the current ordered operation after checked integer translation: actual
mask bytes or fixed outline control points, colors, background/selection geometry
and effective clips. Whole intersecting masks/outlines can conservatively reject
reuse. The accepted plan also drives full/strip replay, avoiding a second
font/color read between validation and drawing. Native adapters always request
VerifyPaint; other callers may explicitly make the original epoch promise.

Owned retention per renderer is bounded by 64 groups, 32 MiB of RGBA pixels and
8 MiB of descriptor accounting (operation slots, fixed path arrays and owned alpha
headers/pixels). Identical masks may share owned bytes within one capture after
current-byte comparison; mutations preserve older copies. The borrowed lookup is
cleared before retention. No font faces, shaped runs, widgets, model references or
render commands survive capture. Map/Go allocator metadata and replacement/builder
transients are separate; this is not a process/native/GPU memory bound. Resize,
eviction and clock exhaustion retire retained pixels and descriptors.

Unknown commands/colors/images, uncertain coverage, finite uniform source
clipping, transforms/masks/filters, painting paths, unsupported text and large
jumps fall back to ordinary replay. Disabled fill-and-stroke box paths are
harmless. Zero text clips are refused because the legacy scanner treats them as
unclipped. Capture preserves SVG/bitmap/outline precedence, refusing SVG/bitmap,
stroked/decorated/vertical/math text, and stops at its byte/primitive limit.
Detached replay preserves covered-uniform-clear elision for full repairs.
The original flat commands remain available to SVG/PDF/other renderers; normal
GPU image upload stays complete. CapturedPixels counts CPU cache writes.

Shared quadratic/cubic subdivision now uses coordinates relative to its integer
control hull, restoring the origin and terminal endpoint exactly. Ordinary fills,
strokes and dashes use this kernel even when scroll reuse is disabled. Some prior
antialias edge pixels intentionally change. Logical paths, text shaping, layout
and semantic geometry remain unchanged. Exact emitted-edge and new full-replay
pixel comparisons establish translation stability, not whole old-pixel parity.

## Checks and review boundaries

The preceding canonical source is `56d394f`. Preparation adds six new prepared Go
files and changes six existing ones: 6,146 exact prepared files. 13 broader v3
preparation/portable/driver/native race/vet/build checks pass. A first attempt
failed setup because it used the sandbox-denied default Go cache; the accepted
retry uses controlled writable caches. V4 and v5 each have seven affected
preparation/raster race/vet/native-pixel/build checks. The final v6 policy has ten
preparation/full-driver race/vet/consumer/policy-pixel/build checks, with prepared
files exactly equal to v5. These distinct source versions and receipts are
retained; earlier checks are not relabelled as running on the final policy.

Primitive tests exercise dirty parent repaint, later overlaps, mixed-axis motion,
repair/resize, signed extreme offsets, coverage and bounded eviction. Shaped
selected/bidi/combining text is compared to full replay at four DPIs and four
fractional phases. Guard tests change ink, background, selection, clipping,
subpixel phase, glyph output and the exact consumed mask; retain owned capacities;
stop at budgets; and test wrapped storage/nonzero bounds/one-pixel dimensions and
strided same-mask A-B-A capture. Restoring the original curve kernel (plus an unused
test-only helper) compiles and fails the real curve/text translation assertions.

Final manual tests compare 56 opted-in managed Grid frames against full scene
replay at 96/120/144/192 DPI, including mixed scroll, repeat position, source
replacement, active selection and disable/enable. Captured cache area must equal
redrawn/exposed area. 12 default-mode frames at four DPIs prove no hints or retained
pane cache. They do not directly exercise actual modal caches, native atlas/GPU
pixels or human VoiceOver workflows. Three live Cocoa checks each qualify v4 and
v5 artifacts: fixed-10k and variable-million 400-cell accessibility/identity/action
smokes, plus generic accessibility. Three final v6 default-mode Cocoa checks pass and are recorded separately.
No input source, VoiceOver or OS preference is changed.

## Matched experiments

Each experiment retains 40 interleaved ordinary native processes, five per build
for fixed/variable 10k/million stored tables. Every process completes 128 guarded
warmups and 120 measured frames, retaining all outliers: 9,600 total observations
across the two experiments. Diagnostics are disabled. Paired geometry is exact
across builds and row counts. Candidate artifacts match their qualified build
receipts and exact selected source/prepared maps throughout. The preceding
release artifact is reused with its clean canonical/source/prepared/build link;
it was not freshly built for this checkpoint. Pooled nearest-rank p95 UI work,
in milliseconds (positive change means slower):

| Experiment | Case | Baseline | Experimental cache | Change |
| --- | --- | ---: | ---: | ---: |
| v4 | fixed-10000 | 11.968 | 12.680 | +5.95% |
| v4 | fixed-1000000 | 12.117 | 12.503 | +3.19% |
| v4 | variable-10000 | 12.051 | 12.713 | +5.50% |
| v4 | variable-1000000 | 11.706 | 12.544 | +7.16% |
| v5 | fixed-10000 | 12.054 | 12.366 | +2.59% |
| v5 | fixed-1000000 | 11.912 | 12.245 | +2.79% |
| v5 | variable-10000 | 12.127 | 12.498 | +3.05% |
| v5 | variable-1000000 | 11.880 | 12.624 | +6.27% |

V4 captures the complete pane after reuse. V5 uses toroidal storage, copied-mask
sharing and preserves covered-clear elision. Each experiment remains slower at
p95 in all four cases; comparisons between experiments are different host runs.
All 80 processes fail both unchanged 4 ms p95 /6 ms p99 targets. Full phase/queue
and p50/p95/p99/max distributions and per-process variation remain in the raw
summaries. Separate nested-source diagnostics on the preceding artifact motivated
this work; they do not qualify warm targets or physical GPU transfers.

The final v6 artifact gates this same v5 renderer behind an explicit opt-in.
The v4/v5 experimental measurements are not latency measurements of final default
v6 rendering. No final default speedup is claimed. Reconciliation/layout, actual
GPU rendering/upload changes and broader Mac acceptance remain open; the CPU
cache experiment must not be substituted for those gates.

## Evidence

The archive/index retain frozen source versions, original inputs, checks and
logs, generated test images, source/prepared/binary links, live smokes, negative
proofs, both full comparisons and internal review certificates. Source approvals
and the v3 capacity finding/correction remain separate. Final report/receipt
review, significant-batch Anthropic Claude Sonnet 5.5 Prism output and final audit
are kept in review/ and audit.json. Selected endpoint bindings do not claim a
complete compiler/SDK closure, human input/accessibility, native/GPU allocation
reclamation or physical presentation qualification.

## Final review disposition

Prism reviewed the 16 implementation, patch, test, specification and report files
with Anthropic Claude Sonnet 5.5. It exited 0 with no retained findings and complete
coverage. Its raw output also contains one discarded medium comment about the
shared curve kernel changing ordinary raster output. Independent source review
confirmed that scope as the intentional compatibility change described above;
it did not rely on Prism's evidence-location discard rationale, which was not
reproduced. No unresolved defect was demonstrated. Existing raster tests pass,
but their generated PNG fixtures do not establish parity with prior antialias
pixels. The raw provider record and separate internal disposition are preserved.

This review paragraph, the explicit fill/stroke/dash scope clarification and a
spacing correction were added after Prism. They receive a separate internal
closure review; the original provider input and source approvals remain intact.
