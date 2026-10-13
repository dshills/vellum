# Complete texture and readback ownership

Patch 0049 uses the patch48 owned-reference constraint to publish a texture only
with its complete live view. Failure retires acquired references. Configuration
shortcuts require live device/queue and an owned live same-device pair with the required usage bits. Changing
the device retires old texture, view, sampler and staging. Empty borrowed state
is cleared on retirement, and created pairs explicitly acquire owned mode.

Readback staging keeps matching live buffers, preserves old references/layout on
same-device allocation refusal, and releases replaced buffers synchronously.
The prior implementation depended on WebGPU Go finalizers after replacement.
Pinned TryCreate methods already release native handles on their error paths;
partial-return fault tables test defensive control contracts, not those backend
errors. The outer texture must still release a successful texture if view creation
fails. Factory references must be fresh and owned; caller operations are serialized.

Synchronous staging accepts positive single-layer/single-sample four-byte
RGBA8/BGRA8 resolves, with uint32 row/image and Go map-length bounds. Configure and
encode a matching copy before mapping; mapping never allocates an empty uncopied
replacement. Checked copy refuses stale/invalid inputs. GrabImage does not read
a refused copy. Raw ReadData keeps native channels; detached ReadGoImage bytes
normalize BGRA to NRGBA channel order without alpha/transfer conversion.

Tests include seven original-compatible invalid/borrowed/empty-borrowed cases,
counted factory failures, exact layout/refusal checks, 1,000 counted staging
replacements, device-change refusal and four-format image-channel vectors.
General GPU fault/readback and physical memory reclamation remain separate. Native
backend/in-flight retention, sustained performance and live memory tests remain
separate; desktop work is deferred while the human uses the machine. Modified
upstream notices remain BSD-style; new source/tests are MIT. The module cache is
not edited. All inputs/outputs and the patch are hash-pinned.


NoDisplayGPU now refuses a missing adapter before device creation and releases
admission references on failed device creation. An injected five-case table
checks these paths. The global instance remains shared until final shutdown.
An explicitly opted-in headless fixture uses tiny real GPU textures/buffers and
records image-to-render/upload usage replacement, 32 staging reference retirements,
device-change retirement and two 3x2 RGBA/BGRA copy/map/image roundtrips. Held old
Go wrappers verify explicit Release, not finalizer activity. This creates no window,
surface or presentation. Successful native pixel/reference checks do not prove
physical heap reclamation, backend fault recovery or runtime performance.
