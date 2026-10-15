# Covered uniform background pixels

Owned patch 0064 subtracts uniform destination pixels that the first later
pixel-producing command must replace. At most eight items are inspected. The
first rectangle must contain at least 4,096 pixels, both rectangles must lie
inside the target with fully available uniform sources, and overlap must cover
at least half the first rectangle. Remaining pixels use at most four disjoint
ordinary draws with rebased source positions. Captured items stay unchanged.

Only builtin value colors qualify. Src replaces at any alpha; Over replaces only
when opaque. Lookahead crosses plain context markers and bounded, validated
identity-transform move/line/close paths with neither fill nor stroke and the
renderer-owned scanner. Paths still execute in their original order. Malformed,
transformed, long or unknown paths, custom scanner callbacks, masks, unknown
sources, scroll hints and intervening paint retain ordinary replay.

The helper runs only for unclipped CPU replay, including full scroll-cache
repaints. Strip-clipped cache replay and scene-packet capture use their existing
paths. StandardBox still emits its original clears, preserving removed border
and shadow erasure. No GPU admission, encoding or rendering policy changes.

Independent original dispatch checks dirty padded images, source bounds, live
borrowed colors, repeated chains, scanner observers and recovered path panics.
The identical benchmark body runs against pristine and patched renderers.
Preparation verifies original/output file hashes and patch bytes. New files are
MIT; renderer.go retains its upstream BSD notice. Native performance remains
unqualified. See the [checkpoint report](../../../../../docs/reports/macos-uniform-crop/README.md).
