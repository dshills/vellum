# Storage readback and temporary layouts

Patch 0051 validates buffer-size operands before multiplication and slicing.
Storage readback requires positive four-byte-aligned bytes, live source/staging
size and usage, and the originating device/queue. Buffer-backed legacy
StorageTexture variables retain the existing buffer route; real texture-backed
values use the Texture layout/copy API instead. Staging reuse checks its own
actual size, independently of GPU AllocSize. Fresh replacement is published only
on success and retires the old reference; allocation refusal preserves it.
CreateBuffer, SetFromBytes and WriteDynamicBuffer return automatic staging errors
and repair missing staging on otherwise-reusable GPU-buffer paths. A valid GPU
upload may precede a staging refusal; this is not atomic content rollback.

Mapping preflights all members, known owner identities and duplicates before any
native map. Values and buffers have one serialized owner and enter unmapped;
external native maps and copied Value ownership are unsupported. Callback channels
capture only a per-attempt status channel. After synchronous Poll, callback delivery
has a separate 30-second wait; Poll itself has no new deadline. Failure cancels
all attempted exact buffers. Refused rollback quarantines staging until explicit
ReleaseRead. Late delivery cannot mutate owner state or a replacement buffer.

ReadToBytes requires this API's completed mapping and an exact full-byte target.
Short-target refusal preserves a ready map for retry; range/unmap refusal
quarantines it before another read. Upload/copy/reconfiguration refuse held or
quarantined mappings. Explicit zero/readonly retirement drops staging.
Vars.Config preserves native layout validation and immediately releases its
otherwise-discarded temporary layout references. Pipeline layouts remain distinct.

Fault kernels, original-compatible refusal and opt-in tiny native readback/shader
checks exercise the contract. Counted mapping/allocation failures are not physical
GPU faults. Wrapper retirement and source release loops do not prove native/GPU
heap reclamation, sustained performance, presentation or full Mac acceptance.
Modified upstream source retains BSD-style notices; new helpers/tests are MIT.
The module cache is not edited. Every stage input/output and patch is hash-pinned.
