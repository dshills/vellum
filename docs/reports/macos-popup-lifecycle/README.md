# Popup timeout ownership checkpoint

Stage 82 repairs a native popup lifecycle prerequisite found while extending M2
control-catalogue checks. A time.AfterFunc callback previously read and destroyed
live popup widgets outside the render context. Final native event-loop teardown
also destroyed stages outside that context. Both paths now acquire the same owner
used by rendering; timeout callbacks capture it before firing.

A fresh ticket per arming rejects stale callbacks even for the same Stage/Scene.
Rearm, timeout disabling and stage deletion revoke the ticket before nonblocking
Stop. Manual timers retain their existing owned FIFO path. Successful expiry
requests one paint; stale or duplicate completions do not. Event-loop teardown
finishes before the window wait completes, including release of that wait if
destructive cleanup panics.

Validation: final owned Core race suite, complete driver-root package (`.`) race suite, minimum-Go native
gallery compilation, qualified Core vet, independently reproduced dependency
preparation and all eleven preparation tests. Exact outcomes and receipts are
recorded in records/results.json. Native compilation is separate from execution:
this checkpoint launches no OS window and claims no physical-pointer, presentation,
full M2, M4 or Mac completion result. Commit requires internal review and one bounded Anthropic Claude Sonnet 5.5
Prism checkpoint; their outcomes belong in the final evidence archive.

The unfiltered upstream Core suite failed as a diagnostic: its AssertRender
fixtures raced ordinary rendering and hit the admitted 8 MiB log limit after
36 seconds. Its incomplete log remains incomplete. It generated 300 missing
image-test PNGs in the retained SDK. No original payload file changed; only those
SHA-checked new PNGs were moved into a retained private quarantine. The exact
7,880-file original SDK vector was restored. No SDK refresh occurred. Unqualified
Core vet separately failed on unkeyed literals and unreachable code in unchanged
upstream source; qualified vet disables only those two checks. An initial private
preparation-test fixture omitted patch files and failed; the complete fixture
passes unchanged tests. These failures are retained, not counted as passes.

Next: bounded detached actual-popup inspection, then ordinary Selector popup and
NumberField native routing/validation/cleanup checks. Full catalogue and M2–M7
remain open. No OS input-source, VoiceOver, preference or clipboard change occurs.
