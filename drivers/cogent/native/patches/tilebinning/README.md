# Adaptive scene tile-bin counts

Owned stage0075 replaces the encoder's first per-reference counting traversal.
Admission computes the exact original rectangular tile-reference count, including
operations later omitted by EncodeGPU. Original operation, mask/data, encoded-byte
and reference limits are unchanged. Builder-produced rectangles are clipped to
bounded packet extents; per-tile counts are bounded by MaxOps, totals by
MaxTileReferences. All arithmetic fits supported Go integer targets.

When retained references exceed twice the tile count, four rectangle-difference
corners and an in-place two-dimensional prefix pass recover exact counts. Sparse
batches retain direct counting. The same temporary count array subsequently
becomes fill positions. There is no new cache, backing allocation, retained
reference or packet/Encoded ABI. Final painter order, compact operation IDs,
per-tile reference order, detached bytes, CPU replay and reset proof are unchanged.

The frozen pre-change encoder is the byte/admission oracle. Tests cover mixed
masks/pixels/spans, random nonaligned rectangles, both encoding modes, exact
reference limits, threshold edges and 65,536 operations in one tile. Existing
encoded-byte/operation bounds and frozen digests remain required. A wrong corner
sign is rejected by the oracle. The managed-driver offscreen fixture compares
24 actual fixed/variable grid frames at supported manual DPI96/160/192 through
serialized integer execution, packet replay and complete captured pixels.
Manual DPI162 is unsupported; its failed fixture attempt remains separate from
native DPI162 evidence. Microbenchmarks are synthetic, and native/ordinary
latency, GPU completion, memory and full Mac qualification are separate gates.

Modified/new owned Go source is MIT. Preparation pins patch input/output hashes
and patch SHA256 before replacing either prepared module.
