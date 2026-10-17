# Opt-in local semantic automation

The public `automation` package exposes explicitly named application queries and
typed domain actions. macOS is the implemented endpoint target; Windows/Linux
`Start` returns `ErrUnsupported`. Importing the package and normal driver startup
open no endpoint. The application registers a `Registry`, then explicitly calls
`Start` before Run or on its owner. Start freezes the registry (at most 256 total
bindings); later registration and a second Start with that registry refuse.
Registry and Server are constructor-owned pointer handles and must not be copied.
Registration/Start/Close refuse copied or uninitialized handles; inspection
methods return empty metadata (Done returns nil) for an invalid server handle.
Copying an active synchronizing value itself is unsupported; these method guards
do not make a concurrent value copy safe.

## Access and protocol

Start creates a new private `/tmp/vellum-api-*` directory with mode 0700, a Unix
socket with mode 0600, and a cryptographically random 256-bit bearer capability.
`Endpoint` carries socket path, protocol version and token. The application gives
that descriptor privately to authorized clients. The same OS account or root can
access process/filesystem state; this capability is not a same-user sandbox.
No TCP listener, automatic service discovery or implicit external UI access exists.

Each connection sends the lowercase 64-character hex token and newline, then one
UTF-8 JSON request and newline. Clients keep their write side open until receiving
the response; disconnect, half-close or a second message cancels the request.
The server sends one JSON response and newline and closes. `Call` implements this
contract with context cancellation and a maximum one-minute client deadline.
Server request deadlines default to five seconds and cannot exceed one minute.

Version 1 requests have `version`, `id`, `operation`, `target`, `params`, and an
`expected_revision` required for `invoke` and forbidden for `inspect`. IDs and
targets are 1–128 ASCII letters/digits or `._:-`. Params are JSON objects, including
`{}` for argument-free commands. Maximum request JSON is 32 KiB; parsing rejects
unknown/case-varied envelope fields, duplicate keys at all depths, trailing
values, invalid UTF-8, lone escaped UTF-16 surrogates and nesting beyond 32.
Typed params use Go JSON field decoding with unknown-field refusal; applications
validate required fields and domain constraints in their shared callbacks.

Responses preserve the request ID and contain a result or a fixed public error
code/message. A revision is included when available. Callback errors and panic
text, parameters, results and credentials never enter the audit ring. Callback
error text is never transmitted. DTO authors explicitly omit/redact sensitive
values; the transport cannot infer which arbitrary application fields are secret.
Results must fit the 64 KiB response limit; oversized results refuse. That limit
bounds retained/transmitted JSON, not application callback/JSON marshaler
allocations. Application callbacks own their result size and execution time.

## Owner, revision and cancellation

Socket goroutines parse/framing-check before admission, then use bounded
`Application.Post`. Domain queries, typed decoding, revision reads and command
availability/execution run only on the application owner. `RegisterCommand`
uses the existing `command.Registry.Invoke`, including its availability check.
Typed actions should share the same validation and mutation functions as UI
callbacks; external transport grants no bypass around those domain rules.

The application supplies a revision identity for all exposed state/availability.
It must change on every exposed mutation, including target/model replacement and
an ABA change back to an earlier value. Equality is rechecked on owner admission
immediately before invoking the selected handler. A stale invocation refuses.
An optional Window lifetime scopes the endpoint; the framework never reopens a
retired window ID. Registries/named targets are immutable while the endpoint runs.

There are eight combined client/owner-work credits. Disconnect/deadline cannot
return an accepted owner's credit until that callback explicitly skips or returns,
or application closure drops its queue. Repeated disconnected requests therefore
cannot accumulate more than eight retained owner envelopes from this endpoint.
Queue-full admission refuses. No background thread enters native UI state.

Cancellation suppresses unstarted callbacks and is cooperative during execution.
It cannot undo a committed action or preempt arbitrary application code. Request
IDs are correlations, not deduplication keys; a lost/canceled response may follow
a committed action. Clients must inspect again and never automatically retry an
invoke. Handler panics propagate through the application's existing owner failure
contract, closing the application; this transport does not recover domain panics.

`Close` cancels clients/listener and removes only its owned socket and empty
directory. It does not wait for the owner, so owner commands may call it safely.
`Done` means transport goroutines ended; admitted owner envelopes still follow
the mailbox contract. No broad RemoveAll or takeover of existing socket paths is
performed. `Audit` returns a detached latest-256 terminal event ring and dropped
count, containing only ID, operation, target, code, handler-started and available
before/after revisions. Started cancellation does not imply rollback; these are
transport observations, not transaction or durable audit guarantees.

Transient accept errors (including descriptor pressure and aborted connections)
retry with cancellation-aware exponential backoff from 5 ms through a 1 s cap.
A successful accept resets it. Persistent transient pressure remains cancellable
without a spin loop or automatic endpoint retirement. Permanent failures retire
the endpoint; `Err` preserves fatal accept/owned-cleanup causes. Normal closure
has no terminal error. Retry fixtures inject errors/waits without changing OS or
process descriptor limits or sleeping in the logic checks.

## Workbench integration and evidence

`vellum-workbench -automation` explicitly starts a main-window-scoped endpoint
and emits its capability descriptor to the invoking process's stdout. That output
is a credential and is intended for the controlling trusted client, not shared
logs. Ordinary startup remains unchanged. `component` inspects one stable
component ID (or selection with `{}`), including virtual/unrendered records;
`workbench` reports selection, record/reference counts, four command availability
flags and a literal `[redacted]` document path. These are domain DTOs, not completed
native geometry, AX inspection or physical rendering assertions.

The action allowlist is `component.select`, `component.rotate`,
`component.raise-value`, `edit.undo`, `edit.redo`. Selection shares the UI selection
function; the others invoke the existing command registry and history validation.
No arbitrary file/chooser, clipboard or app-close command is exposed. Successful
actions invalidate the application's windows. Workbench revisions combine
monotonic source/selection/reference epochs and current command availability;
undo/redo/selection ABA cannot reuse an observed revision. Epoch exhaustion
refuses before mutation rather than wrapping.

Headless real Unix socket tests cover credentials/permissions, typed actions and
command availability, stale requests, malformed/oversized inputs, disconnect
credit retention, cancellation, owner serialization, queue refusal, owner-safe
close, panic/lifetime teardown, redaction, response and audit limits. Workbench
headless tests exercise its actual domain bindings, history and virtual ID query.
These checks create no windows, synthesize no desktop input and do not change OS
preferences. Full P9/P10 professional workload, native runtime, human accessibility,
distribution and sustained performance acceptance remain open.

## External native consumer fixture

`examples/automation` is an independent standard-library-only consumer of public
`Call`. It launches the native workbench with the explicit diagnostic
`-automation -automation-check -exit-after=20s`; 100k and million-row cases are
supported. The check flag admits only 5–30-second dedicated processes without
other smoke/endurance modes and disables process-local geometry persistence.
A signal context exists before endpoint or worker creation. Readiness requires a
real native window and a successful nonempty public GPU snapshot. It makes no
pixel, presentation or frame-budget assertion. Result receipt and worker
retirement are separate: closure follows snapshot-worker cleanup, and the main
path joins worker retirement and transport Done/Err before process exit, including
readiness or descriptor-write failures.

The client must validate one canonical private descriptor, ready/closed markers,
40 serial query/action responses and their corresponding terminal audit records.
The suite checks command/history and virtual identity semantics, explicit typed
refusals, redaction, availability and stale revisions after both domain and
selection ABA changes. Read-only queries omit expected_revision; invokes include
it and are never retried on transport failure. Normal success additionally
requires zero native exit, socket/owned-directory removal and post-close refusal.
All raw native output is discarded, not published through error paths. Stream
and line limits bound capture; direct-child TERM, Kill fallback and final reap
have independent bounded attempts. An outer native test runner owns the inherited
controller/child group and separately proves its absence. The client cannot
promise graceful endpoint cleanup after SIGKILL or OS/process failure.

Non-GUI model/transport regressions and native execution evidence are distinct.
No physical input, native accessibility/geometry, cross-window interaction,
performance, memory or complete macOS acceptance follows from this fixture.
