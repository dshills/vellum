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
locks. `StopMain` itself needs no wake: the existing desktop event wait is bounded
at 100 ms. This guard does not cover unrelated GLFW calls or standalone GPU helper
initialization/termination outside the desktop application driver.

Owned patch22 contains exact input/output hashes and its regressions. Run the
selected backend race suite through `scripts/check.sh`; `cmd/mainqueue-smoke`
provides separate `-case normal` and `-case panic` real desktop-loop probes. Neither
probe opens windows or changes OS input/accessibility preferences. Windowed
shutdown remains a separate native acceptance check.
