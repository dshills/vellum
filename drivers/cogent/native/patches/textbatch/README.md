# Transient passive text batches and sealed glyph masks

Stage 0079 adds an explicit rasterx text-batch capability. A custom frame may
batch contiguous plain passive Text leaves only after their background boxes
were suppressed by the managed pane coverage proof. Visibility, render flags,
pending shaping, selected/highlighted text, debug views, children/parts and
ordinary paint order retain their original behavior. Other renderers decline
batch production. CPU replay and private scene capture understand the compact
command; original per-text admission and glyph placement arithmetic still run.
Shaped lines are borrowed for the same frame lifetime as ordinary render.Text.
Public PaintText mutation remains visible; no layout identity cache is added.

A per-renderer FIFO bank snapshots only immutable zero-origin masks returned by
the private bitmap glyph cache. It retains at most 1,024 entries and 8 MiB of
charged source backing, owned snapshot backing and entry structs. Fixed queue,
map/allocator metadata and other caches are separate. Resize/release clear the
bank. Sealed sceneplan resources expose no image or pixel access; live packets
remain valid after bank eviction. Every packet charges each retained snapshot
once, including its image/resource headers. Public GlyphMask and arbitrary
mutable image sources retain ordinary copy-and-compare ownership. No new shader
operation or translated glyph arithmetic is introduced; small outlines retain
the existing exact-path/placement span route.

Tests compare original CPU and captured replay pixels across five DPI values,
light/dark, scrolling, selected/decorated cells and mutable PaintText layouts.
Resource tests cover clipping/nonzero source origins, detached legacy encoding,
malformed input, byte admission, cache eviction, resize and repeated release.
Native measurements and full Mac acceptance are separate from these checks.
