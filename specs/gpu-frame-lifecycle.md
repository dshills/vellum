# GPU frame ownership and failure

A native frame has one serialized owner. Successful surface acquisition owns its
texture/view and holds the surface lock until Present or AbortCurrentTexture.
Idle abort and idle presentation do nothing. Retirement first detaches references,
then releases view, texture and finally the lock. Teardown retires a held frame
before releasing render/surface/device state; callers stop encoding first.

Ordinary and measured rendering use the same checked preparation. A temporary
yield returns no pass; failed encoder or pass creation releases every returned
partial handle and aborts the frame. Partial native acquisition/view errors also
release returned resources. Nil encoder/pass results without an error are refused.
Surface outdated/lost recovery remains bounded by the existing acquisition policy.

Submission consumes the pass and encoder. It releases the pass before Finish and
releases the finished buffer before the encoder, including partial-buffer failures.
Only a valid finished buffer reaches queue submission. Failed submission aborts
the frame and prevents presentation; image capture does not read an unsubmitted
frame. Checked End and ordinary/measured drawer paths return refusal; the legacy
void End method logs it. Measured presentation remains separate from encoding and
returns zero presentation time on refusal. Queue.Submit admission does not prove
GPU completion, and this change does not add device-loss recreation.

Counted-reference tests exercise the actual shared control functions and actual
Surface mutex, including repeated idle abort, teardown and Go release/queue panic
unwind. Production End refusal is checked for both ordinary and measured paths.
These are code-level ownership checks, not physical backend fault injection or
Cocoa/IOSurface reclamation evidence. Stale/concurrent owners, abandoned active
passes, broad panic recovery, physical device loss and native memory/performance
acceptance require separate work. Desktop checks stay deferred while the user is
using the computer. See the phase report in docs/reports/macos-frame-cleanup.

## Texture and synchronous staging ownership

Texture creation publishes an owned live texture/view pair together. Failed view
creation retires the created texture; nil/invalid factory results are refused.
Same-format configuration reuses only an owned live pair on the same live device
and queue with the required texture usage bits. Image/render usage transitions
recreate incompatible textures. Explicit device changes retire texture/view/sampler/staging references.
Retiring an empty borrowed binding clears its shared flag, so later owned creation
cannot inherit borrowed mode. Borrowed source objects retain their ownership.

Readback configuration reuses a matching live buffer or allocates a fresh owned
replacement. Allocation refusal preserves the prior buffer/layout; success
explicitly retires it, rather than waiting for finalizers. Replacement admits an
old-plus-new ownership interval; native reclamation may lag. Serialized caller
ownership covers configuration, copying, mapping, unmapping and retirement.

This synchronous API accepts positive single-layer/single-sample RGBA8/BGRA8
resolves and validates uint32 layout fields and Go map-length arithmetic. It does
not promise a physical-memory budget; managed canvas/view capture has separate
admission limits. ConfigReadBuffer precedes encoding CopyToReadBuffer, and caller
submission/completion precedes reading. ReadDataMapped refuses stale/missing
staging without allocating uncopied data. GrabImage finishes rendering but returns
no image if copy admission fails. ReadData keeps native channels; ReadGoImage
normalizes detached BGRA bytes to NRGBA order while preserving alpha/transfer bytes.

The pinned native factories already clean their own error references. Counted
partial-result tables are defensive control tests. A 1,000-replacement fixture
proves explicit release calls and maximum two owned handles during replacement;
it does not prove GPU heap retirement or sustained native readback performance.
Original-compatible wrapper-only regressions, layout/channel vectors and actual
invalid-input/device-change guards accompany source review and compilation.


An opt-in tiny headless GPU fixture checks actual image/render usage transitions,
32 staging reference retirements, device changes and exact 3x2 RGBA/BGRA padded
copy/map roundtrips. Old wrappers stay strongly held while their IsValid state
proves explicit reference release. These observations remain separate from
physical GPU heap reclamation, native surfaces/presentation and performance.

Headless device admission refuses a missing adapter and cleans partial GPU/device
references on failure. The shared global instance remains owned by final
application shutdown. Counted factory probes verify error and retirement order.

## Private compute-system ownership

A ComputeSystem borrows its GPU adapter and owns its separately requested device
and queue. NewComputeSystemChecked returns no partial system on adapter/device
refusal; the legacy constructor logs and returns nil. Owner-serialized shutdown
first detaches all system state, abandons pending recording, completes prior work,
retires pipelines/values and then releases the private device/queue. Empty and
repeated shutdown are safe. GraphicsSystem still borrows its renderer's device.

Compute recording admits a complete live command/pass pair. Partial/refused
factories retire acquired references. The pinned native BeginComputePass can
panic; cleanup unwinds but does not convert that panic to a returned error.
Submission requires caller End, detaches the recording, releases pass before
Finish and retires command/buffer on success/refusal. Failure never reaches the
queue, and old bind groups retire on both paths. Shutdown never submits pending
recording. No broad panic recovery or concurrent-owner contract is added.

Device.Release clears ownership before cleanup and releases its independently
owned GetQueue reference even when the device is missing/invalid. Polling requires
a live device. Wrapper retirement and small headless shader-output checks remain
separate from GPU/native heap reclamation, physical device loss, presentation,
sustained performance and full Mac acceptance.

## Storage-value readback and temporary layouts

The legacy Value buffer APIs validate nonnegative shape operands and multiplication
before native allocation or Go slicing. Storage readback uses positive four-byte
aligned ranges, live source/staging handles, exact read sizes and required usage.
Buffer-backed StorageTexture roles retain their prior buffer route; real texture
objects require Texture copy/layout methods. Automatic staging errors propagate
through CreateBuffer, SetFromBytes and WriteDynamicBuffer; matching GPU buffers
still repair absent staging. A staged upload is not rolled back on later refusal.
Staging replacement is transactional with respect to its owned reference and
checks staging size independently of the GPU AllocSize field.

Mapping has one serialized owner and initially unmapped buffers. Values/native
ownership are not copied, and caller-native maps are outside this contract. Every
batch member, known device/queue identity and duplicate is checked before any map.
Callbacks capture only per-attempt buffered channels. The owner waits for statuses
through channels after synchronous Poll; callback delivery then has a 30-second
bound. This does not bound Poll or qualify physical-loss recovery. Any batch error
cancels all attempted exact buffers; a refused rollback quarantines its staging.
Late callbacks cannot publish a map or change a replacement buffer's owner state.

ReadToBytes accepts only this API's completed mapping and exact allocated byte
length (including dynamic-offset padding). Short destinations preserve a ready
map for retry; range/unmap refusal quarantines staging until explicit ReleaseRead.
Held/quarantined maps refuse upload, copying and non-retiring reconfiguration.
Explicit zero/readonly configuration retires staging. Source bytes are detached
before unmapping; no partial copied output is successful when cleanup refuses.

Vars.Config keeps its native bind-layout validation and releases those temporary
owned handles immediately. Pipelines independently create their own layouts.
Source release calls, fault kernels and tiny native words/padding checks remain
separate from actual native/GPU memory, latency/endurance and Mac acceptance.
