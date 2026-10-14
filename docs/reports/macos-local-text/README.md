# Passive grid text layout

Owned patch 0055 removes duplicate shaping from an admitted local grid update.
Previously the ordinary Text updater shaped changed content using the old
allocation and the scheduled local sizing pass immediately shaped it again.
Passive labels now parse/update normally, clear old shaped runs before new rich
styles, and shape in the required sizing pass. The inline editor and ordinary
full-layout updates retain immediate shaping.

The Core helper refuses interactive text, full-layout fences, unrequested roots,
foreign/retired/detached roots or text, and wrong ancestors before mutation.
Caller must keep passive styles and size every updated label before paint.
Grid's existing styleDirty-to-layoutDirty relationship guarantees that pass.
The scoped updater flag restores on return; pending text shapes normally in
Render if a queued root is canceled before sizing. This changes no painter order,
GPU upload, row cache, document data or public Vellum API.

Internal review found and fixed the pre-paint input hazard by refusing interactive
Text. Real-shaper tests cover one scheduled shape, untouched interactive refusal,
ordinary fallback with immediate shape-consuming input, cancellation recovery,
foreign/root/global fences and restored updater scope. Grid race tests cover
local/full pixels, retained membership, editing and accessibility publication.
Formal preparation checks all 55 patch stages and produces exactly three reviewed
Core leaves plus the adapter change; other prepared leaves remain byte-identical.

The native release comparison uses 40 sequential Cocoa processes, five
alternating pairs per case, 128 guarded warmup callbacks and 120 samples per
process (4,800 total). All callback guards, paired geometry and exits pass;
every run and outlier is retained. Both binaries use the same observer and
opt-in GPU scene rendering. [Review summary](records/review-summary.json) records
binary identities, pooled results and scope.

| Case | Baseline / candidate p95 (ms) | Baseline / candidate p99 (ms) | Mean UI work (ms) | Mean Go allocation (MiB/sample) |
| --- | --- | --- | --- | --- |
| fixed-10000 | 6.560 / 6.218 | 7.224 / 6.614 | 5.614 / 5.457 | 9.339 / 9.143 |
| fixed-1000000 | 6.383 / 5.856 | 6.643 / 6.399 | 5.416 / 5.246 | 8.890 / 8.752 |
| variable-10000 | 6.643 / 6.314 | 7.407 / 6.911 | 5.421 / 5.261 | 9.195 / 9.033 |
| variable-1000000 | 6.507 / 5.953 | 6.799 / 6.420 | 5.269 / 5.099 | 8.780 / 8.610 |

Observed mean UI work improves 2.8–3.2%, p95 improves 4.9–8.5%, mean process Go
allocation falls 1.5–2.1%, and allocation count falls about 4.8–4.9%. These are
local pooled observations, not statistical proof across hosts. Both binaries
still miss the unchanged 4 ms p95 / 6 ms p99 targets. Host UI work excludes queue
wait and drawable acquisition, and does not measure GPU completion or physical
presentation. Default CPU rendering was not timed in this campaign.

The checkpoint uses internal source review and one bounded Prism review with
Anthropic Claude Sonnet 5.5. Raw overlay sources, superseded failed fixture run,
passing receipts, all native records and reviews are preserved in a SHA256-indexed
local archive after review. Complete P6, composed P11/endurance, native memory,
human input/accessibility, physical GPU/display and distribution gates remain open.
