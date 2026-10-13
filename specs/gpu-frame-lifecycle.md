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
