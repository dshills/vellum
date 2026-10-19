# Text allocation alignment

Patch 0095 positions a Text's shaped block within its actual horizontal content
allocation. Existing shaping still aligns shorter lines against the longest
line; the shared origin adds only the spare block-to-allocation distance.
Rendering, pointer-to-rune selection mapping, and hyperlink bounds use the same
origin. Existing nearest-rune hyperlink hit semantics, including linked spaces,
are preserved. Omitted/default Start alignment preserves its old origin.

Text's final sizing also honors an explicit positive Grow while retaining its
ordinary GrowWrap measurement policy. Previously GrowWrap blocked final growth,
leaving a wide allocation around an intrinsic-width label; a centered parent
could then position that intrinsic block twice, while paragraph alignment had
no actual content space. Growing labels now receive their final content width
before parent positioning and shared-origin alignment.

Offscreen Core regressions use the real shaper to check all three alignments,
selection/rune mapping, link bounds, large-font linked spaces, and multiline
right edges. A real-shaper regression checks final growth with wrapping retained.
Vellum's driver regressions reproduce the wide centered Weather heading using
completed glyph pixels and check that the ink is centered in its parent, plus keyed
updates, style removal, and native Button/MenuButton content justification.
The patch changes no global fonts, preferences, window state, or application
data. It supplies no physical pointer or native desktop verification claim.

New tests/helper changes are MIT; upstream text.go retains its original notice.
Preparation verifies the original/output file hashes and patch SHA256.
