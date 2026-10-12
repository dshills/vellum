# Renderer-owned static outline reuse

Stage 43 follows sparse span bands. Each Rasterx renderer owns a private LRU,
shared only with its temporary scene/scroll captures. Keys are immutable Font
identity plus full GID; raw font-unit geometry is independent of paint, size,
DPI and transform. Static TrueType admission excludes variable/CFF/encoded fonts.

Entries come only from successful new-owned bounded retrieval, and are read-only
inside Rasterx. Public font retrieval ownership and CPU SVG/bitmap/outline order
stay unchanged. CPU admission refusal uses ordinary retrieval. An oversized
successful GPU outline still renders without retention. Stage 4 in typesetting
supplies opaque exact successful work costs; every warm GPU hit replays its
original aggregate point/traversal charge before unchanged geometry/scanner work.

The cache retains at most 2MiB segment capacity,1024 entries,8 Font references,
and64KiB segment capacity per entry, per renderer. Metadata, transient decoding,
Font/backing storage and the existing glyph-mask cache are separate. Release
clears this new cache in both CPU and GPU retirement while retaining snapshot
pixels. These limits are not total process/native/GPU memory qualification.

Regressions exercise original geometry/pixels, public output independence,
aggregate cold/warm admission, capacity/slot/font eviction, variable CPU fallback,
read-only concurrent cache calls, capture ownership, resize and retirement.
