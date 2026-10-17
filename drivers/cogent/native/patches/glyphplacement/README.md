# Exact-location small glyph coverage

Stage 0076 adds a renderer-owned index for repeated small glyph scene coverage.
It applies after the original scene work checks and bitmap cache return. An exact
identity matrix, current private immutable outline, final effective scale bits and
final position bits identify the original fixed path. No rounding approximation
or integer-translation assumption authorizes reuse.

Each static-outline cache has a separate small epoch token, with no font or
outline references. Per-outline uint32 IDs occupy the legacy entry's trailing
padding on supported 64-bit targets. IDs never wrap; zero disables reuse for
later admissions after exhaustion while existing live IDs remain valid. A live
font/GID/slice ownership check precedes every lookup. Custom, replaced and evicted
outlines cannot reuse an old ID.

The independent placement FIFO contains at most 8192 keys and values. This is an
additional metadata limit, separate from the unchanged 512-entry/1-MiB coverage
array limit; allocator/map overhead is not an exact physical byte guarantee.
Values retain only small tokens, the coverage fingerprint, original origin and
guarded hull. A distinct coverage-entry token invalidates eviction and hash
reuse without retaining evicted paths or spans. Learning requires successful
original drawing and full fixed-path equality with the actual retained entry.

Every hit rechecks current target/clip, known uniform color, owned spanner,
scanner dimensions, clean scanner/filler state and guarded hull. Replay shares
the original winding, color, packed-span fallback and cleanup. All original
logical glyph, outline, traversal and extent charges precede reuse. Bitmap,
transformed, stroked, unsupported and custom paths retain their prior behavior.
Resize and Release discard the index through the existing coverage owner.

After a full index produces fewer than 1024 hits in 8192 probes, the next 262144
eligible glyphs use the original path without index lookup or learning. The
bounded pause then permits another probe window. An underfull index can continue
to miss, including on changed clips; this policy does not promise zero overhead
for every workload. Cold and clipped-miss costs must remain visible in results.

Tests compare exact original-point bounds, fresh packet bytes in both encodings,
seeded replay pixels, actual warm grid captures at DPI96/160/192, source/cache
replacement, FIFO bounds, collisions, clipping, work charges, exhaustion and
cooldown routing. Intentional guard-removal experiments must fail their targeted
regressions. Software capture timings do not qualify native frame budgets, GPU
completion, retained/native memory, physical input or full macOS completion.

Modified/new owned Go source is MIT. Preparation verifies input, patch and output
hashes before replacing either prepared dependency module.
