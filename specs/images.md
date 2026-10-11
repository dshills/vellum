# Retained image drawing

The portable `canvas.TiledImage` stores immutable straight-alpha sRGB RGBA8 pixels
in 128×128 tiles. `NewTiledImage` copies validated NRGBA input, including nonzero
origins, subimages and row strides. Dimensions are bounded to 1..8192 (at most
256 MiB of retained pixels). Large source construction belongs on a worker.

`WithPatch` requires a fully contained nonempty normalized image rectangle and
matching source dimensions. It clones only intersecting tiles plus the bounded
tile pointer list. Identical tiles keep their identity; a completely unchanged
patch returns the original image. Versions share immutable data and retain no
parent chain. `ChangedTiles(previous)` compares tile identity in O(tile count),
without pixel scans, and returns opaque tiles with copying pixel accessors.
Source data and placement can be shared between windows under their ordinary
application ownership rules.

`ImageLayer` places an image in bounded world coordinates and selects nearest
or linear filtering. `MeshCanvas.Image` draws an optional layer below its meshes,
using the same UI device, retained offscreen target, clipping, input navigation,
resource lifetime and snapshot contracts. An image-only canvas is allowed.
The source texture uses sRGB RGBA8; shaders decode texels to linear values and
premultiply before filtering/blending. Linear filtering interpolates four
premultiplied texels explicitly, preventing straight-alpha color halos.

On first allocation or resource rebuild, the renderer uploads the complete tiled
source. Subsequent versions update only changed tiles. Camera, placement, filter
and UI chrome changes retain source texels. Normal composition does not read back
or rasterize on the CPU. Per-window resources have independent upload baselines.
Partial upload or submission failure prevents publication and clears the baseline
so the next attempt repairs the complete source. Removal releases image resources.

GPU admission includes image texture bytes and its 32-byte uniform, alongside the
existing target, mesh and snapshot staging estimate. Logical device texture and
buffer limits are checked before allocation; texture/pipeline/bind/upload APIs
return checked errors. Counters distinguish image allocation, tile uploads and
bytes from mesh work. CPU source/version/temporary upload copies, driver internals
and transient allocator residency are outside the tracked GPU estimate.

Acceptance requires actual 4096×4096 source data, native GPU color/alpha/filter
pixels, single-tile changed upload counts, unchanged/camera retention, independent
window baselines, removal, rebuild, resizing and closure. A full-image replacement
from unrelated immutable tile identities intentionally requires a complete upload.
Native clipped composition, broader 3D/compute/float/asynchronous-readback P4 work,
sustained performance and physical device-loss qualification remain separate.
