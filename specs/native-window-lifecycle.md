# Native window cycle acceptance

The explicitly opted-in Mac `window-cycle-smoke` consumer exercises fresh keyed
window identities while one native survivor remains alive. Every cycle mounts
a label, controlled text field, checkbox, button and shared-device rectangle
canvas. A widget OnShow observation precedes owner-posted readback so the fixture
cannot pump a window while its native topology is still being constructed.
Readback must match the completed viewport dimensions and expected scene pixel.

`Application.PostWait` waits for queue admission, not callback completion. Every
owner probe therefore returns a buffered immutable reply, which the worker must
receive before evaluating the result. Admission or an expired undrained callback
cannot qualify readback or retirement. A five-second context bounds admission,
polling and reply waits; an already executing synchronous GPU operation remains
subject to the outer process watchdog rather than being preempted by that context.

Close cancels captured original contexts and descendants, revokes a reserved
completion, and rejects new delivery and late result publication. The fixture
waits for the originating desktop composer's atomic withdrawal and render-window
registry removal, then observes only the survivor in visible Cocoa titles and
the Go accessibility bridge registry. These are separate observations, not one
atomic snapshot or proof of native/GPU deallocation. The retired canvas must
refuse native-window access and readback. The survivor is read back again at each
sampling interval. Main-loop cleanup must have no visible windows, retained
service operations or accessibility bridge membership.

Post-GC Go heap/in-use/system/allocation/GC/goroutine observations and self-process
`ps` RSS are recorded at declared intervals. The default interval is 25 cycles;
samples over 512 MiB live Go heap or 1 GiB RSS stop the run as failed after retaining
the offending record. Those sampled safety limits do not bound unsampled peaks
or define performance acceptance. RSS is resident pages, not private footprint
or native allocation attribution. Geometry persistence is disabled within this
disposable fixture so unique titles do not write user preferences. No input,
clipboard, permission or OS setting change is part of the test.

The alpha.2 portable 1,000-cycle regressions cover both unique-name and same-name
retirement with live-only lifetime/cancel maps. Exact opening references reject
stale work after name reuse; see [window lifetimes](window-lifetimes.md).
Historical native campaigns ran their recorded source with ID tombstones; they
do not qualify this API transition. Fresh native same-name lifecycle evidence
remains required. Native cycles establish observed mounting/readback/normal
closure only at their recorded scope. They do not qualify the full P11 integrated
component mix, 1,000-control baseline, continuous workload latency, overnight
soak, native heap or physical GPU memory. Caller-held objects remain separate
retention costs.

## Sampled Cocoa window metadata

Opt-in `vellum_native_checks` snapshots include `IdentityAvailable`,
`WindowClass`, `ContentViewClass` and signed 64-bit `WindowNumber` for every
sampled process window. The class names come from the actual runtime classes;
an absent content view has an empty class name. Each class name is limited to
256 UTF-8 bytes. A missing window class or an oversized class refuses the whole
snapshot, preserving the existing window-count and wire bounds. The observation
creates no historical identity registry and retains no window/content-view
references beyond the main-thread sample.

`IdentityAvailable=false` in old records means metadata was not observed. A
current observed zero/negative window number is preserved rather than classified.
Cocoa window numbers can be reused, and class names describe types: these fields
are not lifetime capabilities, managed ownership, allocation/deallocation or leak
proof. Matching known framework windows and explaining the historical hidden
untitled window still require fresh native evidence. Historical raw records stay
unchanged.

The same serializer and Go window DTO have inert Foundation coverage for boolean
types, exact signed numbers, absent views, UTF-8 byte limits and legacy missing
metadata. These fixtures create no NSApplication/NSWindow, dispatch no native
operation and query no window server. Native sampling and cleanup commands must
also compile, but compilation and inert fixtures do not qualify their runtime
behavior or continuous state coverage.

An optional `-heap-profile` writes a diagnostic after worker checks and before
application closure, including after a safety-limit failure. It requires a new
clean absolute path, uses exclusive creation and caps compressed output at
8 MiB. Write/close errors remain failures. Profiling adds another forced GC and
is diagnostic only; allocation sites do not prove the retaining reference.

The [glyph-mask cache](glyph-mask-cache.md) bounds one shared Go cache. The native
retention report preserves production failures and an earlier diagnostic overlay;
it does not qualify the adopted patch's native memory behavior. Future acceptance
must observe physical footprint and native graphics allocations as well as RSS.


## Accessibility temporary ownership

Native accessibility attachment and teardown own local autorelease pools even
when called directly from Go's main callback queue. Registry/view association
ownership keeps a live attachment valid after its producer pool drains. Teardown
revokes every element, clears geometry, restores the view class when still owned,
and removes association/registry ownership before draining collection temporaries.
Logical registry removal alone does not establish native object retirement.

The inert ARC ownership check requires weak node/state expiry before an enclosing
pool drains and preserves the original-code failing assertion. The full composed
campaign additionally observes native interaction and cleanup, sampled Go/RSS,
and source-bound native allocation snapshots. Physical graphics allocation and
all remaining P11 budgets remain separate from these observations. See the
[accessibility retention checkpoint](../docs/reports/macos-ax-retention/README.md).


## Retired construction contexts

Closing a render window withdraws it from all render registries and replaces its
current construction context with a registered survivor, or nil when none remain.
After the last manual window closes, a subsequent ordinary window must open and
deliver Show through its ordinary event loop. A retired manual context must not
make that construction silently fail the live manual-mode restrictions.

Native focus membership checks, focus ordering and current-context publication
are atomic with registry withdrawal. Raise refuses retired windows and publishes
before invoking native code, so a reentrant close remains withdrawn. Native
focus/raise calls run outside the registry mutex. This does not extend native
object lifetimes or replace owner-thread requirements.

Owned patch 0070 includes offscreen manual-to-ordinary construction, retirement,
reentrant Raise and concurrent focus/close regression coverage. Existing mixed
live manual/ordinary mode refusals still apply. Offscreen routing evidence does
not establish physical desktop focus, native object deallocation or full Mac
acceptance.


## Occlusion notification ownership

The Mac accessibility/native attachment owns one object-specific Cocoa occlusion
observer, installed before its initial main-thread sample. A weak window capture
prevents extending native lifetime. Detach invalidates the observer and clears its
handler before native window destruction. Go bridge retirement independently
revokes the bounded owner FIFO state, including failure/partial-attach paths.

The state retains only detached visibility/sequence and one exact pending redraw
wake. Initial samples flush on OnShow binding; duplicate notifications coalesce.
Retirement removes undelivered nodes and clears the source. Dequeued retired or
foreign wakes cannot repaint or consume a later incarnation's pending request.
The owner marks scene dirt before requesting rendering, so a clean scene can be
recomposed after a visible transition. Existing Frame completion and unavailable
semantics remain unchanged. See [integration contract](desktop-integration.md).

Foundation-only observer fixtures and Manual owner/composition regressions do not
establish real NSWindow notification timing, Metal availability or presentation.
The [checkpoint](../docs/reports/macos-window-visibility/README.md) records the
current qualification and preserves earlier failed native scopes.
