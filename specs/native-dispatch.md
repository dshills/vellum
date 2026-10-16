# Native dispatch and termination

The pinned desktop adapter runs one native event loop per process. Workers submit
synchronous callbacks through its owned queue; a callback must not dispatch back
onto that queue. Before publication, the native startup caller may execute inline.
This is an adapter contract, not a promise that arbitrary worker calls during
startup are safe native operations.

Normal return is the only successful callback outcome. Once delivery is accepted,
the caller waits for one acknowledgement even if shutdown begins. Caller-owned
output remains inaccessible until that acknowledgement. Pending delivery can be
refused after stop; native window creation returns `ErrMainStopped` on refusal.
The stable stop signal is closed once and cannot restart the process loop.

An admitted callback panic closes admission before acknowledging failure. The
exact panic payload is rethrown on the native thread and on the dispatch caller;
partial callback output never becomes a successful result. Inline startup uses
the same admission and failure behavior. `panic(nil)` preserves the runtime's
`PanicNilError`. `runtime.Goexit` retains its effect on the native goroutine, while
its waiting caller panics with `ErrMainCallbackIncomplete`. Legacy
`GODEBUG=panicnil=1` is unsupported because recovery cannot distinguish a nil
panic from `Goexit`. These failures do not resume the native loop.

Both main-loop implementations close admission on exit. Desktop cleanup also
terminates its GLFW lifecycle. After the queue is terminal, Cogent's recovery
handler records the crash and immediately rethrows; it cannot wait on a crash
window whose event loop no longer exists. This is fail-stop behavior, not a
recoverable callback-error API or an assertion of complete GPU/AX cleanup after
arbitrary application panics.

The process-wide desktop event barrier serializes initialization, native empty
posts and termination. A post admitted before termination finishes first. Posts
before initialization or after termination are ignored. Duplicate termination and
restart after termination do not call native functions. All desktop key, character,
pointer and scroll wakes use this barrier, as do dispatched and render-window
wakes. It never spans polling, callback execution, queue delivery or application
locks. The portable `StopMain` closes the stop signal without posting a native
event. Windows and Linux retain the finite 100 ms event wait. The Mac loop owns
the stop wake and joins its ingress before native cleanup. This guard does not
cover unrelated GLFW calls or standalone GPU helper
initialization/termination outside the desktop application driver.

On Mac, owned patch 0069 replaces the periodic wait deadline with `WaitEvents`.
One process-loop ingress receives the public unbuffered queue, publishes each
packet to a private one-slot ready channel, then posts through the event barrier.
The native thread alone executes admitted callbacks. Publication before posting
closes the race in which the public caller's earlier wake was consumed before
its queue send became selectable. Existing public leading/trailing wakes remain.

Ingress retains at most one ready and one local pending packet, in addition to
the native thread's active callback. It clears its local packet after transfer.
Shutdown closes admission, refuses only those already received pending packets
through the existing exactly-once acknowledgement protocol, and joins ingress
before cleanup and GLFW termination. An active callback keeps completion and
output ownership. Stop-before-start, callback panic/Goexit and a failing native
wait use the same retirement order. A native post panic remains fatal and is
re-thrown after refusing ingress-owned packets; it provides no native cleanup
guarantee for a fatal worker panic. No goroutine-per-call or polling timer is added.

Injected-wait regression tests cover routing and retirement without initializing
GLFW. They do not establish native idle CPU, application-wide wakeups, GPU
presentation counts or interactive event delivery. Those require separate,
source-bound Mac acceptance; removing a recurring deadline alone does not prove
that any performance budget has been met.

Owned patches 0022 (callback failure) and 0069 (Mac ingress) contain exact
input/output hashes and their regressions. Run the
selected backend race suite through `scripts/check.sh`; `cmd/mainqueue-smoke`
provides separate `-case normal`, `-case panic`, `-case idle` and `-case stop`
real desktop-loop probes. The idle case delays three synchronous callbacks; the
stop case delays a worker-only stop without submitting any callback. These two
cases omit the deliberate empty-event flood. They exercise wait/delivery and stop
liveness, without qualifying idle CPU or wake budgets. None of these probes
opens windows or changes OS input/accessibility preferences. Windowed
shutdown remains a separate native acceptance check.
