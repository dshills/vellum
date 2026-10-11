# Explicit execution and completion admission

The portable `execution` package supplies real and manually stepped executors and
clocks. Applications inject `execution.Executor` through `tasks.Config.Executor`
and `document.Config.Executor`; absent injection uses a real goroutine executor.
These hooks control worker execution and portable timers. Native Scene timers,
input, layout and frame scheduling require the separate replay adapter before
P9 can be qualified.

## Accepted work and terminal ownership

`Executor.Submit` refuses without executing, or admits one callback exactly once.
It never executes inline and never waits for execution, so an owner callback can
submit without blocking its own completion pump. Successful admission retains
work through cancellation and executor closure. The callback performs its own
cancellation and cleanup. Arbitrary third-party executors must obey this contract.

`NewManual(capacity)` bounds pending plus running jobs. `PendingIDs` returns a
fresh slice in admission order. `Step(ctx,id)` selects one job; `StepNext` selects
the earliest pending job. Each callback executes in an isolated goroutine and the
step waits for its terminal result. Canceling the wait leaves the callback running,
visible and occupying capacity. `Job.Done/Result` observe eventual completion.
Concurrent callers may step different jobs, but the same identity starts once.
Close rejects new jobs and retains accepted jobs for explicit stepping. IDs refuse
on exhaustion. There is no retained unbounded completion history.

Panic and runtime.Goexit become observable terminal failures, after callback
defers finish. Refunding the execution slot precedes publishing its Done signal.
`Goroutine` has no executor-local limit; tasks/document admission supplies the
application bound. Go cannot terminate a callback that ignores cancellation.

## Application completion reservations

`app.Config.CompletionCapacity` defaults to 256, independently of the ordinary
`QueueCapacity` default of 256. Both must be positive when explicitly set.
`ReserveCompletion(window, cleanup)` binds one exact ticket to the window's
immutable lifetime (or the application for an empty window). Reserve before
worker submission. It never inserts a placeholder into the ready queue.

`Completion.Publish(fn)` admits the final result without waiting for ordinary
queue space; duplicate, copied, aborted and retired tickets refuse. Ordinary
Post and successful Publish append to one FIFO under the same mutex. If A and B
are reserved, then P posts, B publishes, Q posts and A publishes, dispatch is
P/B/Q/A regardless of reservation order. An unresolved reservation never blocks
ordinary ready input. Sequence credit is reserved to guarantee later publication;
new admission refuses rather than reusing identities at exhaustion.

`Abort` withdraws reserved/ready delivery, but cannot interrupt dispatch.
Application/window close revokes unstarted results without user callbacks.
An already dispatching callback retires in its own terminal defer. Logical
completion credit lasts through reserved, ready and dispatching states; it is
refunded at callback termination/revocation. This bounds admitted deliveries, not
concurrently retiring resources or arbitrary callback heap use.

The optional synchronous cleanup hook runs once outside app.mu. It is for
resource/accounting cleanup, must return normally without blocking, and must not
enter Do or mutate UI/domain state. Owner callbacks belong in Publish. Cleanup
remains separately observable in `Application.Pending` until it finishes; Done
is published after that cleanup. If a hook breaches normal-return requirements,
unrelated terminal hooks still retire and native wake notification still occurs.

`DrainN(max)` dispatches at most max entries from the initial arrival cutoff.
Work posted by a callback remains in the next batch even if another original
entry is aborted. Zero is a checked no-op; negatives refuse. `Drain` preserves
its existing complete-initial-batch behavior. Neither method may nest Do/Drain.
An owner callback that panics or calls runtime.Goexit closes the app after
releasing ownership; a panic also propagates. `Pending`
distinguishes ordinary ready, reserved, ready completions, dispatching and cleanup.
DirtyWindows is separate; mailbox Idle does not imply a native frame is settled.

## Tasks and document IO

`tasks.NewWithConfig` adds optional window lifetime and Executor injection; `New`
retains its existing default behavior. Task slots remain occupied until both
worker termination and owner completion/revocation. `Task.Done` retains worker
plus completion-admission meaning. `OwnerDone` reports delivery termination;
Snapshot exposes WorkerDone, OwnerDone and AwaitingOwner. Slot refund occurs
synchronously before the final terminal signal. Canceling/closing cannot free a
still-running worker slot prematurely. Progress remains coalesced best-effort
ordinary admission; the final result uses its reserved completion.

`document.Config.Executor` controls Encode/Decode/store/chooser work. Save/Open
reserve completion before submission and return submission refusal while restoring
session busy/action/operation state. No manual step waits for PostWait while the
owner is waiting for that step. State exposes WorkerPending and OwnerPending in
addition to Busy/Closed; closed owner delivery may finish before its worker.
The existing history-revision and exact persistence contracts remain unchanged.

## Explicit clocks

Both `NewRealClock(capacity)` and `NewManualClock(epoch, capacity, maxCallbacks)`
bound pending plus running occurrences, including a reset during a callback.
Stop removes a pending occurrence; it cannot stop running callbacks. Context
cancellation suppresses unstarted work; Close stops pending occurrences and
keeps running work visible. Reset refuses canceled/closed/negative/capacity/
identity/deadline exhaustion atomically.

Manual time moves only through `Advance` or `AdvanceContext`. Due callbacks run
by deadline then latest arm sequence, outside the clock mutex, and observe their
own scheduled instant. Nested timers/reset may run before the same fixed target.
The positive per-advance callback budget bounds zero-delay feedback. Budget
exhaustion leaves remaining work pending and time at the last reached instant.
Callbacks report panic/Goexit; later due work still runs within the budget.

Canceling an advance wait keeps the reached time and running occurrence; advance
admission remains occupied until that callback really terminates. A later advance
refuses meanwhile. There is no hidden resumption, time jump or reaper goroutine.
Advance without a cancelable context waits for terminating callbacks. A process
watchdog is still required for native tests with uncooperative callbacks.

Clock lifetime Idle requires no pending/running callback and no active advance.
A replay runner must distinguish ready-at-current-time timers from future armed
ones, and current-instant quiescence from complete lifetime termination. Inspection
must never advance time or dispatch a future timer. Native replay must also account
for deferred origin events, frames, workers and owner delivery before claiming idle.

`RealClock.ReadySnapshot` and `ManualClock.ReadySnapshot` inspect under one lock
without firing or advancing. ClockReadiness adds captured Now, Due, Future,
NextDeadline and HasDeadline to ClockSnapshot; Due+Future equals Pending.
CurrentIdle requires no due/running callback and no active advance, while inherited
Idle also requires no future pending timer. A canceled occurrence awaiting its
stop observer stays registered. Nil/uninitialized clock inspection refuses; a
closed initialized clock remains inspectable through running teardown.

Task.Done can become ready before the surrounding executor callback's terminal
defer finishes. A document OnComplete can start a new operation and replace the
Session's currently reported operation. Aggregate quiescence must therefore also
inspect the chosen Executor.Snapshot; task/session status alone cannot certify
that every accepted callback has terminated.
