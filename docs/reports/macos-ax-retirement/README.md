# Accessibility retirement ordering and diagnostics

A late accessibility refresh could republish a closed bridge into the window
registry because bind neither checked retirement nor serialized registry
publication with close. An offscreen regression fails on the original source
with a retired window entry still present. Restoring just the original bind
body through a Go overlay also fails the final regression. The fixed bind holds
the bridge mutex through publication and refuses closed bridges. Retirement
removes the exact bridge only, so an old attachment cannot displace a replacement.

Each bridge now stores one copied first-close record before signaling retirement:
window ID, monotonic attachment token, observed semantic generation and the
teardown call-path label. Requests holding that bridge return an
AccessibilityClosedError that unwraps ErrAccessibilityClosed. Callers can inspect
the copied record with errors.As; the error retains no bridge, application,
controls or callbacks. Subsequent cleanup does not overwrite the first record.
No retired global registry or historical diagnostic log is added. A lookup after
the bridge is fully unregistered still reports the window unavailable.

Labels identify native attachment release, native/text attachment setup failure,
driver unlaunched cleanup and driver window cleanup. They describe the teardown
path, not the event that caused it. The native cycle harness emits an optional
AttachmentRetirement record on such a failure and stops; it does not retry a
closed attachment or qualify an admitted but incomplete action.

Regression coverage includes closed/concurrent bind, preserving a replacement,
immutable first-close copies, request cancellation and a ready native result
racing retirement. The native action harness verifies copied provenance without
opening windows. Full offscreen adapter race tests and native build/vet checks
are recorded separately. The complete offscreen adapter race suite passes in
99.27 seconds; native harness race tests, offscreen/native vet and the tagged
native build pass. The native executable is built but not launched. All 6,204
prepared dependency files remain unchanged. No live tests, preferences or
clipboard access are used.

This checkpoint requires internal review between Prism reviews, following the
separately preserved color-inclusive endurance checkpoint. It fixes a verified
registry defect and adds diagnostics; it does not establish the cause of the
343-window endurance failure or pass the required fresh complete campaign.
Full P6/P11, native/GPU memory, human/physical input, display and distribution
gates remain open. The [review manifest](records/qualification.json) and
[evidence index](records/evidence-index.json) record issued reviews, exact sources
and scoped results; future live checks remain deferred while the human uses the
machine.
