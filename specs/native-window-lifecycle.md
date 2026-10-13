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

The portable 1,000-cycle regression separately proves shared canceled-context
retention and survivor-only cancel-function storage while retaining ID tombstones.
Native cycles establish observed mounting/readback/normal closure at this scope.
They do not qualify the full P11 integrated component mix, 1,000-control baseline,
continuous workload latency, overnight soak, native heap or physical GPU memory.
Historical ID strings and caller-held objects remain separate retention costs.

An optional `-heap-profile` writes a diagnostic after worker checks and before
application closure, including after a safety-limit failure. It requires a new
clean absolute path, uses exclusive creation and caps compressed output at
8 MiB. Write/close errors remain failures. Profiling adds another forced GC and
is diagnostic only; allocation sites do not prove the retaining reference.
