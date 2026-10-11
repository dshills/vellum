# Retained macOS image drawing

The portable `canvas.TiledImage` copies straight-alpha RGBA8 sources into immutable
128×128 tiles. Patches share unchanged tiles without keeping a parent/version
chain. Nonzero origins, padded/subimage strides, malformed headers, crossing tile
boundaries, unchanged patches, copying accessors and concurrent versions are tested.

`MeshCanvas.Image` draws a retained image below mesh geometry, or by itself, on
the UI window's shared device and 4× MSAA target. Its sRGB texture decodes to linear
texels; filtering premultiplies before interpolation to avoid transparent color
halos. The completed-frame key includes source, placement and filter, independently
from the per-window uploaded-texel baseline. Partial upload or submission failure
withdraws publication and forces complete texel repair on the next attempt.
Controlled writer-fault tests cover this baseline seam; they are not physical GPU
failure injection. Checked allocation/upload APIs and budget/device admission
include image pixels and the 32-byte uniform, including snapshot readback.

The independent, race-built native consumer passes with an actual 4096×4096 source:
1,024 initial tiles and 67,108,864 bytes. A 128×128 patch uploads exactly one tile
and 65,536 bytes without texture reallocation. Native snapshots verify base/patch
pixels, unchanged neighboring pixels, image-below-mesh ordering, independent
image-only windows, nearest selection, halo-free half-alpha red filtering and
translated fractional placement. Acknowledged chrome visits and camera/placement/
filter changes retain texels. Snapshot admission includes source pixels; explicit
rebuild repairs all 1,024 tiles, three resizes retain texels/texture allocation,
and removal clears live image bytes and leaves a transparent empty frame.
The one injected render budget refusal is expected in the successful native log.
Fixture windows close and no input-source or VoiceOver preferences change.

Complete `scripts/check.sh` passes 87 Python regressions, portable/native race
checks and vet, owned-backend regressions, native builds, independent consumers and
public SDK contracts. Five focused native image/drawing race repetitions pass.
Internal review corrected malformed-rectangle overflow admission and required a
chrome acknowledgment baseline captured immediately before invalidation. It also
made two acceptance assertions precise: snapshot budget includes the uniform and
excludes only texels; the independent consumer mutates the actual copied patched
pixel. Those two test-only corrections followed the full checks; affected consumer
race/vet/build and the final native fixture pass on corrected bytes. The exact
inputs, commands, exit status and log hashes are in
[source-and-check-proof.json](source-and-check-proof.json).

Three portable benchmark runs measured a one-pixel patch to a 4096² source at
2.61–3.12 µs/op, 75,080 B/op and four allocations/op. This copies one 64 KiB tile
and the 1,024-entry pointer list; source construction is outside the timed loop.
These runs overlap other checks and do not measure GPU or sustained performance.

[Image contract](../../../specs/images.md) describes bounded API behavior. Native
scroll-clipped compositor/MSAA edge pixels, sustained frame/input/CPU/GPU and global
resource acceptance, physical GPU loss, 3D/compute/float/asynchronous-readback and
other [Mac gates](../../../specs/macos-completion.md) remain open. This checkpoint
is not completion of P4, the Mac framework or public distribution qualification.
