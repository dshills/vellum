# Desktop pointer modes and positioning

This extends the cursor ownership contract with per-window native visibility,
relative movement and positioning. The first backend is the pinned GLFW/Cocoa
adapter. Public Go interfaces remain `Cursor.Hide`, `Show`, `IsVisible`,
`Window.SetCursorEnabled`, `IsCursorEnabled`, and `SetMousePos`.

## Native ownership and status

Each window owns its pointer backend alongside its cursor image/handle owner.
Operations use the existing native-thread dispatcher: direct native callbacks
run through the main execution barrier, workers wait for queue publication.
They never call GLFW from workers or hold cursor/window mutexes across native
callbacks. Startup shutdown, failed creation, window retirement and terminal
cleanup use the existing registry. Mode and positioning each have their own
bounded sequence; superseded queued operations have no native effect. Cursor
image/size operations have a separate sequence.

`Hide` selects enabled-hidden mode; `Show` selects enabled-normal mode. Disabled
mode stays disabled through either method and remembers the shown preference
for re-enabling. `IsVisible` reports a live, known normal GLFW mode.
`IsCursorEnabled` reports a live, known normal/hidden mode. Unknown readback
makes both false. These describe logical driver state, not focus, occlusion,
global pointer location or compositor verification. Inherited exported fields
and direct GLFW access remain caller-controlled; use supported methods for
synchronized access. `CursorBase.Vis` retains the requested shown preference.

Raw motion is optional and only enabled after disabled mode is read back.
Leaving disabled mode or disabling raw first clears raw and confirms it cleared.
Capability is queried once per owner; unsupported platforms receive no raw setter
calls, including `false`. Cocoa currently falls back to ordinary disabled motion.
Known partial/refused native changes publish their observed logical state;
requested intent is retained for explicit retry. The existing public methods
have void signatures and can refuse unavailable, stopped, invalid or nested
operations. Private checked paths expose refusals to regression controls.

Backend callbacks may inspect synchronized status and retire the window.
Lifetime is revalidated before and after each callback. Nested mode/position
mutations during a native transaction are refused; their requested mode intent
is retained for explicit retry. No nested success is reported. Cleanup captures
backend callbacks, marks ownership closed and clears references before restoring
raw/normal mode. If retirement reenters an active backend callback, one bounded
restoration closure runs after that callback unwinds, so its final write cannot
overwrite restoration. Native window destruction retains the existing queued
lifecycle ordering. Panic-resilient drains attempt pointer restoration and cursor
image destruction before window destruction; native panics retain the main
barrier's fail-stop semantics.

## Coordinates and producer state

`SetMousePos` accepts window pixels, including finite fractional, negative and
outside-window coordinates. On macOS it divides by the latest published DPR on
native main. Both pixel and converted native values must fit the architecture's
`image.Point` integer interval before a native setter can store virtual positions.
Invalid scale/nonfinite values and inactive/hidden/iconified native windows are
refused before positioning. Readback is a logical backend observation; normal
warp notifications and physical pointer placement remain backend-dependent.

Incoming positions use one atomic scale snapshot, float64 arithmetic and floor
for negative fractions. Invalid/overflowing positions, integer-overflowing
previous/start deltas and nonfinite/float32-overflowing scrolls are dropped before
producer state changes. GLFW owns disabled virtual positions/recentering; the
adapter never adds a per-movement warp.

Coordinate-changing mode setter paths and explicit disabled virtual positioning
refresh only producer `Last.MousePos` from checked readback. They generate no
synthetic mouse event. Failed/unknown mode readback and failed disabled-warp
readback request a fresh baseline for the next admitted position while retaining
observed applied mode status. No-op mode/raw-only changes do not silently rebase it.
Focus regain requests a fresh baseline on the next admitted position because
Cocoa may finish changing mode after its focus callback. DPR snapshots and focus
epochs are acknowledged by the native producer without clearing concurrent
publications. Button/drop position updates acknowledge their snapshots so the
first following drag keeps its real delta. Scroll does not acknowledge a new
position baseline. Original drag-start position and button/down timestamps are
preserved; applications should cancel/restart a drag if their coordinate space
changes mid-drag. No new cross-window/global capture contract is introduced.

## Qualification

Fake native backends exercise actual window/cursor dispatch, state transitions,
mode/raw ordering/refusal, coordinates, routing and retirement with no GLFW init,
window creation or pointer movement. Fresh pinned preparation, race/vet and
private SDK compile/link/package checks establish source/packaging evidence.
They do not establish physical movement, Cocoa visibility/relative capture,
Retina transitions, actual raw support, allocation/endurance or measured display
responsiveness. Those live gates remain open. Mac release requires measured
60 Hz responsiveness; 4 ms remains an optimization target.
