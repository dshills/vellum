# Checked RGBA upload rows

RGBA subimages retain their parent's stride. The previous texture upload declared
four times the subimage width, so later GPU rows came from incorrect byte offsets.
The correction validates row layout and sends the actual stride and complete
pixel range, with zero extra row packing. Empty/reversed/overflow/short storage
is refused; a single row ignores its unused inter-row stride.

An internally reviewed candidate passes 600 deterministic layout cases under
race testing and GPU-package vet. Four actual headless texture upload/copy/map
cases pass, including odd padding, missing final padding, negative origins and
post-upload input mutation. Restoring the original tight-row assumption fails
all three multirow cases while the single-row control passes. These tests created
no windows. Original source, malformed-bound review repair and fault evidence
are retained.

The source is adopted as owned patch0061 on the committed progress-control base.
Normal preparation reproduces the four pinned outputs: 6,211 to 6,214 dependency
leaves, with every other leaf unchanged. Eight final code/build gates pass:
preparation tests, regeneration, full GPU-package race/vet, both native texture
roots with no skipped tests, full adapter race/vet and native Gallery build.
Preparation unit tests use mocked installations and stable pre-regeneration
snapshots; runtime tests separately bind the complete regenerated tree. The
native Gallery binary is built but was not launched. Final internal source and
evidence review and one bounded Anthropic Claude Sonnet 5.5 Prism review precede
commit.
This correction does not explain the procedural-grid timing failure or claim a
performance gain. Image composition geometry, native surfaces/presentation,
physical GPU memory, fullP6/P11 and Mac distribution remain open.
