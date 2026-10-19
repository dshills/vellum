# Desktop cursor ownership

The Mac adapter binds `App.Cursor(window)` to the exact requested window.
Independent windows have independent current cursors, size/visibility status and
native caches. Calling another window's cursor API cannot retarget an existing
cursor object. A cursor retained after its window closes becomes inert and drops
its window/application, current enum and custom-size callback references.

## Admission and retention

`cursorimg.Get` accepts comparable, nonnil enum identities and integer raster
sizes 1..256. Unknown SVG names and invalid sizes fail without retaining empty
cache entries. `None` continues to mean no cursor preference and has no image.
All 39 original built-in images/hotspots and custom-enum center fallback remain.
The requested desktop size is converted using published logical DPI, divided by
DPR on macOS, then rounded to an even raster size. Invalid/nonfinite conversions
and results outside the size range fail while preserving the current assignment.
No size is silently clamped.

The managed shared raster LRU retains at most 64 entries / 8 MiB of RGBA pixels.
Each native window retains at most 16 handles / 1 MiB of admitted pixel payload,
including its active and pending replacement. The active handle is pinned until
a replacement is successfully bound. Inactive entries are evicted before native
allocation. The largest image includes its shadow: 272 x 272 x 4 bytes; two such
images fit the native payload limit.

These are retention/admission limits, not measurements of private Cocoa/GLFW
allocation. Native cache keys also retain the corresponding CPU rasters, bounded
separately by each owner's admitted payload. Working renders/copies, open-window
count, application-retained results and direct legacy map mutations are outside
the shared 8 MiB budget. This is not a fixed whole-process memory limit.

`cursorimg.Cursors` remains available for compatibility; direct access requires
caller exclusion from managed API calls. Valid injected rasters are copied to
small owned RGBA storage on first managed admission, so tiny subimages cannot
retain large caller backing arrays through the managed cache. Get results are
read-only during use and retain pointer identity while cached. Eviction permits
re-rendering. `cursorimg.Clear` is the production invalidation API. Legacy direct
clearing is detected at the next managed validation; arbitrary direct writes do
not carry framework synchronization or retention guarantees.

## Theme and dispatch

Official theme setters serialize against cursor SVG construction/rendering.
Color-image callbacks, enum String and custom unit/image callbacks run outside
internal locks. Cursor rendering uses fixed color snapshots, the known CPU
rasterizer and a fontless Go shaper for the embedded shape-only SVGs; it does not
invoke application renderer/shaper factories or scan system fonts. Other direct
readers/writers of exported global theme fields still need caller synchronization.

Each request records its window sequence before callbacks/rendering. Newer
requests and size changes retire older pending requests. Native main validation
checks sequence, target lifetime and both cache/theme generations. Its acceptance
is the linearization point: a later invalidation can overlap an already admitted
assignment. Successful native binding updates Current even if a newer request
starts during binding and subsequently fails. Stale/closed/stopped requests return
errors while preserving the last valid assignment.

GLFW create/bind/evict/destroy operations run only on the recorded initialization
thread or the existing native main queue. Workers wait for queue publication;
main callbacks dispatch directly through the same failure barrier. Window
creation uses this same admission path, including before queue publication. No cursor or
window mutex is held across a queue wait, and the cursor path never acquires the
window mutex. Published DPI/DPR is read atomically. The existing Close path's
window mutex across final queue dispatch is not changed by this checkpoint.

The pinned GLFW binding can return a nonnil wrapper around NULL. The adapter
checks its source-verified one-pointer layout before allocation and rejects nil
or zero wrappers before binding. Pending handles are owned before bind. GLFW
panics retain the existing terminal queue semantics; they are not recoverable
allocation refusals. Terminal cleanup attempts every owned destroy and always
reaches GLFW termination, including when a cleanup panics.

## Teardown and compatibility

Theme changes retire stale inactive native entries at the next Set. A window's
old active handle remains until successful replacement or that window's close.
An internal native-owner registry includes windows already removed from the app's
public window list. Window teardown detaches/destroys cursors before destroying
the GLFW window; app shutdown drains all remaining owners before termination.
No per-window hook consumes the bounded external cleanup registrations.

`TheCursor` remains a legacy facade that snapshots the current context window
when used. Its SetSize supplies defaults for subsequently created windows;
applications should use `App.Cursor(window).SetSize` for exact-window changes.
SetSize changes the requested size; call Set to apply it. Current reports the
applied enum rather than its size. Direct CursorBase field access is retained for source compatibility and requires
caller synchronization. Hide/Show retain their existing visibility-status
behavior; actual native input mode is controlled by Window.SetCursorEnabled.
This phase does not establish native hide/show correctness or disabled/raw-mode
interoperation.

## Qualification

Regression controls exercise the actual managed cache and adapter with a fake
native backend: limits, invalid requests, custom backing ownership, themes and
concurrency, exact-window routing, request ordering, queue publication/shutdown,
main reentry with the window mutex held, allocation refusal/null wrappers,
pending bind panic and complete cleanup. The original 390-case shape/hotspot
matrix remains required.

Mocks, race tests, source inspection and native compilation establish no physical
pointer, Retina/DPI appearance, OS cursor-allocation, native thread instrumentation,
visibility-mode, idle/endurance or measured presentation result. Those remain
Mac completion gates. Measured 60 Hz responsiveness governs release; 4 ms remains
an optimization target. See macos-delivery.md and macos-responsiveness.md.
