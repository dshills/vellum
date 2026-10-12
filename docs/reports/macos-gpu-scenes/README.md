# Ordered GPU scene checkpoint

The opt-in scene adapter compiles painter-ordered detached operations and blends
exact premultiplied RGBA bytes on the originating window's existing GPU device.
Eligible resident frames avoid the ordinary Go-image scene upload. Recovery,
device changes and mutable image exposure may seed the full CPU base. The adapter
preserves partial repaints, ordinary source order, CPU fallback and mutable CPU
image semantics.
It remains disabled by default while representative performance, eligibility and
native memory acceptance are open. This is a renderer implementation checkpoint,
not completed Mac/P4/P6 acceptance or a latency speedup claim.

Stage 41 binds 32 upstream files; bounded-outline stage 3 binds three font files.
Fresh preparation reproduces all 6,166 prepared files exactly. The normal font
outline API remains unchanged; bounded static TrueType capture refuses encoded,
CFF and variable fonts before unbounded retrieval. Detached packets own no font,
widget, model, shaped-run or borrowed pixel references. Unsupported whole batches
reconstruct accepted history before ordinary CPU rendering. Shared-device source
leases retain budgets and borrowed views through safe submission/retirement.

The final gate contains 17 passing portable/driver/consumer/raster/GPU
race/vet/preparation checks, including the affected Core/composer scope. A
broader Core offscreen selection timed out in an inherited desktop run-and-show
geometry test; that receipt is preserved and no full Core test pass is claimed.
Compiler tests cover actual paths, filled/stroked/clipped coverage, real outline
text/selection/transforms, math and RGBA icons, unknown hook refusal, winding,
aggregate work/decode guards, packet ownership and bounded replay. Managed software
frames match ordinary rendering through scrolling at 96/120/192 DPI, exercising the
same adapter's actual CPU fallback.

Four explicitly admitted native GPU test groups pass with individual RUN/PASS logs:
96 ordered randomized kernel batches at three sizes, including odd padded rows,
images/masks and overlapping coverage16 spans; actual real outline text/path/
selection/icon byte parity; shared-device ordered sibling composition, partial
updates, fallback recovery, retained mutable snapshots and resize/retirement;
and three-batch borrowed-to-owned slot recycling with exact source/target pixels
and reuse after source retirement. Borrowed redraw refusal remains, and a later
owned CPU batch is eligible. Test readbacks are explicit qualification work.

Early development native runs accidentally filtered out their opt-in variable,
so those tests skipped. The correction is recorded; old outputs remain historical
and unqualified. The corrected final runner explicitly merges the native flag and
controlled Go cache paths, uses verbose test execution logs, and verifies source
stability. No initial skipped run is used as native parity evidence.

The final actual Cocoa fixed 10k diagnostic process uses 128 warmups/120 samples.
Its 124 main-scene records use the GPU route with zero Go-image upload attempts.
This tests that ordinary startup/size checks retain deferred rendering.
Its diagnostic timing is not an ordinary matched campaign. No physical transfer
bytes, GPU completion/presentation timing, broad modal/sprite/window eligibility,
full native memory or general latency qualification follows from this checkpoint.

Internal source reviews cover the frozen packet/compiler/font, renderer/Core and
GPU seams. They resolved winding initialization, aggregate/decode bounds, buffer
capacity accounting, mutable image exposure, queued retirement, texture aliasing
and abnormal cleanup. Prism used Anthropic Claude Sonnet 5.5 for this significant
implementation batch. A large literal patch timed out twice; successful outer
chunks and a completed review of its exact 32 constituent sources provide the
combined coverage. Raw findings and failed receipts remain preserved. The review
found an unknown span-operator admission gap; it is fixed and covered by a
regression. Defensive span-source and locked drawer-owner checks were also added.
The six-file delta was independently reviewed and tested after the provider run;
no second full provider review is claimed. Source-backed dispositions address the
remaining findings.

The private SDK archive verifies 6,868 files and passes a relocated consumer
race/vet/build check. It uses the prepared sibling Core/typesetting module graph.
This is a private packaging check, not public release or native-runtime acceptance.

The next work is representative native source eligibility and an ordinary matched
performance/memory campaign, followed by targeted CPU compiler allocation work
where measurements justify it. Human input/accessibility, native services,
cross-display behavior, third-party distribution audit and remaining Mac workload
acceptance remain separate open gates. No input sources, VoiceOver or OS
preferences were changed.
