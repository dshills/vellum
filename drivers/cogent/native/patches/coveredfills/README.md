# Covered uniform rectangle draws

Stage 38 changes raster replay and adds differential tests, with exact input,
patch and output hashes. An adjacent Draw/Src clear can be skipped only when
the next item is Draw/Over with identical nonempty rectangles and source
positions, no masks, and exact nonnil image.Uniform sources backed by built-in
value colors. The second color must be fully opaque. Replay checks live colors;
it leaves captured items, painter recording and gradient updates intact.

Unknown image/color implementations, gradients, wrappers, translucent fills,
transforms, blur, pixel operations, intervening paths and the legacy zero-rect
full-image sentinel follow the original dispatch. Borrowed images retain their
existing lifetime and concurrency contract. No additional retained resources,
invalidation, native upload, GPU image or submission behavior is introduced.

Tests replay the pre-change dispatch as an independent pixel oracle against
dirty, padded, nonzero-origin RGBA destinations. They cover built-in colors,
clipped and extreme source coordinates, borrowed-color mutation and observable
custom-color calls, masks, transforms, ordered blur/pixel/path operations and
the actual StandardBox painter at fractional scales. Border/shadow replacement,
resize/regrowth, rounded-parent clipping and gradient Update/At calls are
checked separately. New Vellum tests are MIT; upstream renderer attribution
and BSD license remain in force. A pure 500-cell microbenchmark isolates draw
work; only fresh ordinary native observations can assess complete UI latency.
