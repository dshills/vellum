# Joined CPU/native submission/GPU/presentation diagnostic

Continue the existing Cogent/WebGPU/Metal design. Features, SDK, app, allocation,
cache, render/pacing defaults and the 33.3 ms release gate remain unchanged.
Combine the exact reviewed CPU-upload and Metal-backend observers in a fresh
private source build. Reuse the joined raster clock and the three reviewed
source-detail overlays. Never install this diagnostic library into the retained
SDK or packaged app. Keep all dependency name/versions unchanged; only the exact
wgpu-core and wgpu-hal registry packages become private path copies.

Keep the original 248 routed inputs (128 warmup,120 measured,120 ms period),
locked output/geometry/configuration, native host clock and explicit opaque
scope/tag/generation/drawable joins. Reuse both strict parent classifiers, including
all retained startup records, callback drain, foreground checks and full terminal
accounting. Join each raster/upload row to the actual backend drawable. Attribute
all render commits between the preceding present and this present. Refuse missing,
ambiguous, after-present/unattributed commits, invalid bounds, overflow or drift.

Partition each input's post-acquisition interval at first render commit, first
render GPU start and final render GPU end. GPU span includes gaps/overlap between
buffers: do not sum per-buffer GPU times. CPU submission and GPU execution may
interleave. Keep present-request, present-buffer commit/completion/GPU clocks and
signed request-minus-render-GPU-end separately; never impose invented CPU/GPU
callback ordering. Final-render-GPU-end to shown includes any remaining present
buffer, driver and display work; it is NOT an isolated compositor measurement.
Keep the actual p95-ranked input and its full command identities. Do not add
independent quantiles or infer an ordinary-build speedup from this observer build.

Focused regression checks must cover same-ticket partition closure, sparse tags,
multiple render buffers, signed overlapping present request, missing/duplicate
joins, unassigned trailing submits, bad GPU bounds and inherited refusals. Reuse
the raster race and full-pixel/routing verification; build fresh binaries offline,
pin inputs and reproduce the preparation. Internal review and Prism Anthropic
Claude Sonnet5.5 precede commit and any live run.

One newly approved quiet desktop attempt,40 s native /60 s overall, exclusive
lease and owned executor. Stop on the first invalidity; retain it without retry.
No new broad campaign, no physical-input or displayed-pixel qualification and no
release acceptance. Review and commit actual findings separately after the run.
