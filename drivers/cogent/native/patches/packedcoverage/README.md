# Packed cached glyph spans

Owned patch 0067 copies bounded, ascending cached glyph coverage into one exact
Builder-owned span backing. The cache keeps its existing 16-byte sample shape,
512-entry/1 MiB limits and geometry/collision validation. CPU image rendering
retains the ordinary coverage path. There is no new cache, pool or format.

PackedSpans accepts at most 256 samples into an empty active stream and preflights
the original sequential geometric header/span charges. Unsupported/clipped or
under-budget input declines without mutation, letting ordinary Span preserve
partial-refusal behavior. Operation admission and spare-header refunds still
occur at flush. Packed flush detaches headers and transfers owned spans; packed
subrows never enter the separately charged row scratch. A later ordinary Span
validates/reserves before detaching every packed row. Seal retires all state.

The independent pre-reuse builder oracle, boundary/lifecycle checks, original
source overlays, exact encoded/pixel matrix and software allocation comparisons
qualify this source path. Native GPU, ordinary latency and Mac/P6 completion
remain separate acceptance gates.
