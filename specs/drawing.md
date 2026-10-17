# Retained GPU drawing

The portable `canvas` package now exposes immutable indexed triangle meshes,
world-space polyline strokes, normalized waveform min/max reduction and a
paint-ordered rectangle spatial index. `cogent.MeshCanvas` is the native Mac
adapter. The original bounded `GPUCanvas` rectangle/picking API remains available.

## Portable geometry

`NewMesh` copies and validates finite positions within ±1,000,000, triangle-list
indices and index bounds. Limits are 1,000,000 vertices and 3,000,000 indices.
Accessors copy; a mesh identity cannot change content. Empty geometry is valid.
Vertices use straight-alpha sRGB `color.NRGBA`. Indexed order determines paint
order, with no depth test or back-face culling. Applications supply arbitrary
triangle geometry; this API does not implement polygon fills, text shaping or a
shader language.

`StrokePolyline` removes consecutive duplicates and produces an open stroke with
butt caps and miter joins capped at four half-widths. Width is in world units;
screen-width strokes require regeneration after zoom. Self-intersections and
reversals can overlap and blend more than once. It is not a polygon fill rule.

`Waveform` copies normalized finite samples into an immutable min/max hierarchy.
Construction is O(N), at most 4,000,000 samples. Each envelope queries a chosen
half-open sample interval into at most 16,384 columns in O(columns × log N),
preserving isolated peaks. The output is an envelope, not a decimated line;
constant bands have zero area. Construction retains about 16 MiB for a million
samples (2 × the next power of two × 8 bytes), separately from caller input.

`RectIndex` copies up to 1,000,000 bounded, uniquely identified rectangles and
builds a bounding-volume hierarchy. Input colors are premultiplied `color.RGBA`.
Queries conservatively include intersecting bounds and restore input paint order.
`Mesh` constructs geometry only for the visible query; over-budget visible
geometry returns an error. Fully overlapping datasets can require O(N) queries
and are not promised to become small through indexing. Query statistics report
total objects, tested objects and visible objects separately.

## Native ownership and retention

Mutable adapter fields and methods belong to `Application.Do`/`PostWait`. Owner
and window identity cannot change while mounted. Arrow keys pan when focused. A
mutable adapter mounts once in a live widget. Immutable meshes and indexes may be
shared between workers and windows; native GPU resources belong to each window.
Retirement and originating window cancellation reject resource access and
preparation. The native window handle is published by the originating render
callback; status queries do not traverse unpublished native stage state.

`View` dimensions are full actual content pixels, including content cropped by
scrolling. They are not dp dimensions or the clipped visible rectangle. Dragging
pans and wheel zoom anchors the pointer. Optional `Prepare(Viewport)` runs when
the viewport or application-managed `Revision` changes. It returns a mesh;
expensive source construction belongs on a worker before publication. It must
not recursively call renderer methods. Changing preparation inputs/callbacks
requires advancing `Revision` and invalidating the window.

The adapter borrows the UI window's WebGPU device/queue and allocates one retained
offscreen sRGB target with 4× MSAA resolve, a flat shader and indexed buffers.
It creates no separate device. It converts colors to premultiplied linear RGB
before GPU interpolation and blending, and composites the texture with `Over`.
Background is transparent black by default, or an opaque `color.NRGBA`.
Other background alpha values are rejected.

Vertices are rebased against a mesh-local origin in float64 before float32 upload;
camera offset is also computed in float64. This preserves small translated meshes
and fractional camera changes. GPU positions and transforms are float32; detail
across very large mesh-local spans still has that precision limit. Indexed/cropped
geometry can keep the local span small.

Mesh identity controls vertex/color/index upload. Camera changes upload only the
16-byte transform for a retained mesh. Unchanged visits reuse the completed
offscreen frame. Visible mesh changes may resize/reallocate buffers; the current
backend allocates exact byte lengths rather than retaining padded capacity.
Normal composition performs no CPU rasterization or GPU readback.

Rendering follows application owner → native window mutex → resource mutex.
Composition holds native GPU ownership and takes only the resource mutex;
it never calls the application owner or prepares/renders again. Sources capture
rectangles, texture identity and resource generation. Resize, rebuild and
destruction retire borrowed identities; cached direct-texture replay remains
disabled by the owned backend. Clipping is the intersection of full content
allocation and effective content bounds, translated independently on both axes.

## Resource APIs and limits

`Stats` reports native render visits, successful offscreen submissions, nonempty geometry
upload operations/bytes, uniform bytes, snapshots, allocation generation, size
and submitted triangles. These are renderer counters and calculated resource
sizes, not GPU timestamps, resident process memory or presentation latency.
Image counters report live texel/uniform bytes, successful tile uploads/bytes and
texture allocations separately. Submitted triangle counts include the image quad
(two triangles) when present; uniform upload bytes include its 32-byte transform.

`MaxGPUBytes` defaults to 512 MiB for this renderer's tracked allocation estimate:
20 bytes per target pixel (RGBA8 resolve + 4× MSAA), position/color/index buffers,
16-byte mesh transform, optional [image texture plus 32-byte uniform](images.md),
and optional row-aligned readback staging. Device/driver internal
overhead, UI textures, Go data and transient allocator behavior are separate.
The renderer checks logical-device texture/buffer limits before admission and
uses checked texture APIs for actual target/MSAA allocation. Budget refusal
preserves a previously completed frame and reports the error through `OnError`.
An upload/submission failure prevents publication of a partial frame. `OnError`
runs synchronously on the application owner after releasing native/resource
locks. It may directly change model state, query this adapter or post work; it
must not nest `Do`/`Drain`. Callback replacement belongs to the same owner.
`Mount` and its validation reports share the caller's owner, like other adapter
methods. Normal application closure suppresses a rejected late Render admission.

`Snapshot` copies the last completed frame without flushing pending layout,
preparation or model changes. It must run on the application owner. Readback is
explicit and synchronous; straight-alpha `image.NRGBA` output decodes sRGB,
unpremultiplies in linear space and re-encodes RGB. A snapshot staging buffer
persists until target resize/rebuild/destruction. `RebuildResources` retires live
resources and invalidates for recreation on the healthy shared device.
`RequestFrameAfter` requests one paint; it starts no continuous ticker.
Device loss returns `ErrDeviceLost` under the existing shutdown policy;
automatic physical-device recreation remains unimplemented.

## Acceptance scope

The independent [drawing example](../examples/drawing/README.md) consumes public
packages from its own module. Its native fixture checks offscreen gradient/alpha
pixels, transparent snapshot output, retained resources after acknowledged
native visits, independent windows, 100,000 source rectangles, a million-sample
waveform envelope, budget refusal, rebuild, repeated resize, empty-frame clear
owner-held error callbacks and closure.
Portable tests compare indexed queries with brute force and waveform envelopes
with exact source ranges, including isolated peaks and zoomed sample selection.

The clipping math and generation/lifetime rules have deterministic tests. Native
scroll-clipped compositor pixels and MSAA edge quality require separate evidence;
full-frame offscreen snapshots do not certify them. The rectangle index is not a
general triangle spatial index. [Retained image drawing](images.md) adds the
4096² changed-tile workload with native pixels and resource checks. P4's 3D viewport,
compute preparation, float targets and asynchronous screenshot readback remain
open, along with sustained GPU/frame/input and global resource acceptance.
This phase does not qualify P4 as a whole or the Mac framework as complete.

## Isolated native smoke geometry

The independent example's `--smoke` mode disables backend window-geometry loading
and saving before creating the application. The process-local setting remains
active through native teardown and restores its prior value on normal return.
Interactive mode retains the normal persistence behavior. The smoke result emits
its actual `geometry_persistence_disabled_in_fixture` flag, so external validation
can reject a run that did not isolate fixture geometry. Geometry isolation does
not establish physical input, compositor presentation, native/GPU memory, sustained
performance or complete Mac acceptance.
