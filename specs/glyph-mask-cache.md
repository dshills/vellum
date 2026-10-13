# Shared glyph-mask retention

The private Cogent raster cache retains at most 32 strong font-face keys and
8 MiB of requested alpha-mask pixel data. It evicts whole face groups in FIFO
admission order. A hit returns the existing mask and placement without refreshing
that order. A miss renders first, evicts enough oldest groups, relooks up the
requesting face map (which eviction may have removed), then stores the new mask.
Eviction clears vacated backing-slice slots as well as map keys. All cache
membership, accounting and shared rasterization remain under its existing mutex.

Previously returned masks remain valid through caller references. Eviction
changes neither their pixels nor native resource state. Glyph key quantization,
cacheable dimensions, fractional placement and mask generation retain the pinned
parent behavior. Refused uncacheable glyphs retain no face keys or mask data.

Tests reproduce the parent's 33-key and 8 MiB overrun, then cover exact byte
admission, same-face eviction, FIFO ordering and warm hits, slice-key retirement,
pixel ownership and concurrent reads/renders. An independently instantiated copy
of the pinned original cache methods compares cold, warm and regenerated real
font masks and placement in 64 size/glyph/subpixel cases. Existing raster tests
continue to verify broader scene rendering.

The limits cover face-key count and cached mask data only. Parsed font storage,
map/slice/allocator overhead, transient allocations, other caches and caller-held
masks have separate costs. FIFO can evict a recently used group; bounding retained
resources takes precedence over keeping its warm hit. The native overlay's lower
Go heap is diagnostic evidence; substantial IOSurface growth remains unresolved.
The final owned patch still needs a new quiet-host native memory campaign. Full
P11 endurance, process physical footprint, GPU memory and latency remain open.
