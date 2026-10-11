# Imaging acceptance

The initial reviewed compiler inputs passed two consecutive macOS/arm64 runs of
the same race-built native binary, with no race diagnostics:
[initial run 1](native-smoke-1.txt) and [initial run 2](native-smoke-2.txt).
Their logs and [initial source-hashes.json](source-hashes.json) remain unchanged.
They establish the original GPU checks, not the later review corrections.

The final review corrections passed [corrected race tests](correction-tests.txt),
[vet](correction-vet.txt), [race build](correction-build.txt), and two consecutive
native runs with no race diagnostics: [corrected run 1](correction-native-smoke-1.txt)
and [corrected run 2](correction-native-smoke-2.txt).
[correction-source-hashes.json](correction-source-hashes.json) records their compiler,
adapter, binary, native preparation and all seven reproduced patch-25 output hashes.
The initial proof remains separate and byte-identical.

Every diagnostic result now uses one nonblocking bounded mailbox helper. A regression
fills its slot and exercises actual synchronous SnapshotAsync rejection, callback
error, and callback success on the UI owner; all clear capture busy/update status
without blocking or replacing the unread prior diagnostic. The final native check
also observes normal last-window domain/application cancellation, absent registration,
completed native teardown (IsClosed followed by drawing-lock refusal), consumer
program release, and a zero callback count. The native lock is not context-cancelable;
the process watchdog bounds a stuck lock, while the independent terminal context
bounds the preceding polling. This barrier does not join the private decoder worker
or measure its reservation/private buffer destruction. The force-close scenario
checks native closure, unavailable Stats, suppressed callback and survivor pixels
before explicitly removing that window's domain registration.

The [failed corrected attempt](historical-window-registry-race.txt) remains preserved.
It exposed a shared render-window registry race and an incorrect assumption that
native force-close automatically removed domain registration. Pinned backend patch
25 synchronizes registry lifecycle; the fixture performs the domain cleanup explicitly.
These final verification runs intentionally use pre-commit race builds
(`vcs.modified=true`), with source/binary/preparation hashes to identify the inputs.
They are not claims about a release binary or a clean committed SDK build.

[Race tests](tests.txt), [vet](vet.txt), and [race build](build.txt) passed.
Tests include independent HDR/sRGB reference calculations, partial-alpha and exact
transparent probe rejection, uniform sizing, exposure validation, partial cleanup,
a changed resize request after native size restoration, atomic PNG preservation on
encode failure/cancellation, complete PNG/mode replacement, full-mailbox completion
admission and closed-window delivery suppression. The mailbox regression observes
PostWait's actual wait branch before releasing one queue credit; it does not rely
on a sleep to establish ordering.

The live fixture checked compute-written 256² HDR storage, a single-sample
RGBA16Float attachment, tone mapping and premultiplication into the framework's
four-sample sRGB target. Actual GPU readbacks at exposures 1 and 2 match independent
straight-alpha sRGB probes (±4 per nontransparent channel); the transparent probe
must be exactly zero. Expected left exposure-1 is `(213,156,0,128)` and exposure-2
is `(231,188,0,128)`. The right HDR probe distinguishes preserving a value above 1
from clamping it in the float pass.

Native framework buttons changed exposure/reset and initiated asynchronous Capture.
The capture flow checked one reservation through callback return, recursive busy
refusal, nil/canceled/budget admission refusals without allocation, explicit
synchronous and asynchronous staging coexistence, capture-time pixels/metadata
independent of later model/target changes, quiet cancel/rebuild, separate resources
in two windows, direct-native-force-close delivery suppression,
and successful surviving-window readback. Old-frame delivery happened after target
replacement in both initial and corrected runs (`deliveryAfterResize=true`); this is an observation, not
a required scheduling guarantee. All minimized statistics were captured before
restore: polling/completion advanced without another frame. Chrome-only changes
retained the completed image, and each changed frame uploaded only the 32-byte
uniform. Native resize retained the compute/storage/float program resources.

The PNG chooser was not driven by this fixture: save regressions install a fake
platform service and execute the real worker/save/mailbox code. Native buttons
were activated through the framework accessibility action API, not physical
pointer input. Closed Stats cannot expose private staging counters; the consumer
observes callback suppression and survivor health, while adapter regressions
establish reservation control flow and ownership review checks the private staging
reference. This evidence does not measure private buffer destruction after force
closure, native presentation latency, device-loss recovery, other platforms, or
sustained performance. No OS preferences or input sources were changed.

Historical failures remain visible: [initial shader refusal](historical-shader-refusal.txt)
rejected the WGSL reserved field `target`; renaming it to `viewport` fixed checked
creation. [A repeat fixture failure](historical-resize-no-transition.txt) requested
the restored current native size, so no target transition occurred. The final
fixture reads GLFW dimensions on native main and chooses a different bounded size.
These historical logs are not evidence for the final implementation's success.
