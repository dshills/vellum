# Identity outline extrema

Patch 0053 summarizes the used control-point extrema of retained immutable static
TrueType outlines. Exact identity-transform GPU scene preflight checks cached
corners with the original float32 arithmetic. The original segment, retrieval and
extent-product work charges remain unchanged. Other transforms, foreign backing,
eviction and invalid summaries use the original point loop. Cold successfully
retained outlines can use their newly computed summary too.

Font/GID and segment-backing identity guard access under the existing cache mutex.
Four scalar extrema and validity add bounded metadata within the existing entry
limit; the 2 MiB limit continues to count segment capacity only. No public mutable
font output is retained. CPU rendering, glyph-mask paths and GPU admission limits
remain unchanged; the scene renderer remains opt-in.

The independent oracle covers actual glyphs, coordinate boundaries and random
positions, integer extent equivalence, refusal and eviction. Native parity and
matched whole-frame measurements are recorded in the Mac outline-extrema report.
They do not qualify physical GPU faults, presentation or full Mac acceptance.
New helpers/tests are MIT; modified upstream files retain their original notices.
Preparation verifies each original/output hash and the patch SHA256.
