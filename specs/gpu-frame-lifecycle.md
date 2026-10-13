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
