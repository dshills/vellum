# Mac Shell exposure correctness repair

The owner requests investigation/fix, focused testing, Prism/Anthropic Claude
Sonnet5.5 review, commit and replacement packaging. Features and public APIs,
backend, experimental rendering/pacing and the release gate remain unchanged.

Build11's historical captured-frame mismatch is not consistently reproducible:
two bounded unchanged-binary checks now expose A0 correctly, including after
editing only B to1. Those passes do not erase the earlier failed captures or
prove that the capture service caused them.

Source investigation identifies a deterministic correctness gap in the existing
native rendering path: Raise/focus/show may request paint, but a clean scene can
return before composition. A native destination that needs restoration then
waits for an unrelated model edit. A real retained-scene/software-compositor
regression overwrites only destination pixels, then exercises native-event/close
routing without changing model/scene state. All four cases (raise, close-other,
focus, show), hidden restoration and busy coalescing fail before the fix and pass
afterward. This isolates the missing presentation behavior; it does not claim to
prove the sole cause of every historical screen capture.

Pinned stage0108 requests one coalesced presentation on Raise/focus/show using
the existing per-window forced-frame bit, and consumes it only after visibility
and native ownership admit composition. Hidden passes preserve one request.
Normal unchanged idle still skips composition; no polling loop or command/value
mutation is added. Closed-window registration guards and the native library are
unchanged. The stage changes one runtime Go file plus regression source.

Race regressions for exposure, related busy/display-admission/lifecycle paths
and concurrent retirement/focus pass. Core vet and patch-pin/order checks pass.
The normal preparer replays the stage to exactly the tested hashes, and refuses
a drifted preimage before creating its new file. The local ignored preparation
is updated through that same stage. No historical patch bytes are rewritten.

The source-built ARM64 native archive remains786aaad8…; no dependency/asset/license
input changes. Existing public SDK/apps remain preserved. A replacement Shell
will contain this reviewed prepared-Go overlay, with new executable/notice/archive
identities recorded after the code checkpoint is committed. Existing SDK download
bytes themselves do not gain the fix. Native launch/exposure checks are separate
from the deterministic offscreen evidence. No public upload, tag movement,
signing identity, notarization or global setting/notebook change is authorized.

The code checkpoint Prism review (Anthropic claude-sonnet-5-5) covers all ten
staged files completely and reports no findings. The replacement package/native
observations follow in a separate checkpoint after this source commit.
