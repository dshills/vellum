# Window lifetime references

The alpha.2 API replaces historical-name retention with live registration state.
Portable application/task/document code, native/manual adapters and independent
SDK consumers retain exact opening references. This source contract does not
establish Mac/P11 acceptance; fresh native same-name lifecycle and broader memory
qualification remain separate gates.

`Window.ID` is a logical name, unique among current registrations. A closed name
may be reopened. `OpenWindow(Window) (WindowRef, error)` returns the exact newly
registered lifetime atomically. Convenience `Open(Window) error` discards that
reference. A later `WindowRef(name)` lookup captures the current opening; it is
not evidence of continuity with an earlier Open. Successful registration does not
promise a concurrent close has not already canceled its returned reference.

`WindowRef` is a comparable opaque value. Copies identify the same opening.
`Window.Ref()` in a `Windows()` snapshot comes from that same locked snapshot.
`WindowRef.ID()` returns only its logical name. `Context()` is the original
captured context and stays canceled after retirement; a zero reference has an
already-canceled context. References cannot be serialized or rebuilt from names.

Close/request/check-close and `PostWindow` require an exact reference. While the
application is live, zero/foreign references refuse with `ErrInvalidWindowRef`
and retired references refuse with `ErrWindowClosed`. After application shutdown,
these APIs refuse with `ErrClosed` before validating the reference. A zero reference passed to
`ReserveWindowCompletion` remains invalid even after shutdown. Close-policy
recursion and post-policy validation use the exact
reference, so an old policy that replaces its own window cannot approve or close
the replacement. Already-running callbacks finish cooperatively; cancellation
does not preempt arbitrary code.

`ReserveCompletion(cleanup)` is explicitly application-wide.
`ReserveWindowCompletion(ref, cleanup)` requires a valid live window reference;
zero never falls back to application scope. Revocation compares exact lifetimes
and retains the existing bounded mailbox, cleanup and dispatch rules.

`WindowContext(name)` is a current-live convenience lookup and does not retain or
answer historical names. Task managers and document sessions capture their named
opening during construction and retain its reference/context. Restarting an old
manager/session cannot rebind to a new opening. Creating a new manager/session
with a name deliberately selects the then-current registration. Automation keeps
its original startup context. Global `Post`/`Do` remain application-wide APIs.

`Invalidate(name)` explicitly notifies the current named view. Drivers consume
dirty state with `TakeDirty(ref)`; membership validation and deletion share one
application lock. An old refresh cannot consume a replacement's dirty bit.

Application lifetime/cancel maps contain live windows only. A close cancels the
original context and deletes its entries; application shutdown releases window
declarations, maps and dirty state. Active policies and bounded pending work can
retain references until they return or drain. Caller-held references retain their
small immutable lifetime record/context/issuer, without an Application, View or
native-object backpointer. This is a bookkeeping bound, not total Go/native/GPU
memory qualification or a bound on objects the application deliberately retains.

Native/manual adapters compare the snapshot reference through construction,
OnShow, input, dirty consumption, publication and cleanup. Retire the old native
owner before constructing a same-name replacement. Delayed cleanup deletes only
its exact originating mapping and cannot close the replacement. Logical UI IDs,
persisted Workbench names and replay schemas remain strings; retain lifetime
capabilities in the owning application/adapter instead.

Acceptance requires unique-name and same-name churn with live-only bookkeeping,
stale/foreign/zero refusals, copied-handle behavior, concurrent retirement,
reentrant policy replacement, old task/session/automation refusal, managed Manual
input across replacement, independent SDK/template consumption and a fresh native
same-name lifecycle run. Historical native campaigns keep their original sources
and do not qualify this transition.

## Migrating from alpha.1

- Replace `Open(spec)` followed by a lifetime lookup with
  `ref, err := owner.OpenWindow(spec)` when the caller needs the original opening.
- Pass `ref` to `CloseWindow`, `RequestCloseWindow`, `CheckCloseWindow`,
  `PostWindow` and `TakeDirty`; use `ref.Context()` for workers and OS services.
- Replace `ReserveCompletion(name, cleanup)` with
  `ReserveWindowCompletion(ref, cleanup)`. Replace the old empty-name form with
  `ReserveCompletion(cleanup)` for explicitly application-wide work.
- Task/document constructor `Config.Window` names still select the current
  opening once at construction. Construct a new manager/session for a new opening;
  do not retain an old controller and resolve its name again after closure.
- Managed extension `Mount` implementations can capture
  `cogent.WindowScope(parent)` and retain its owner/reference for deferred work.
  Mount descendants inherit the original scope; an expired parent never adopts
  a same-name replacement. Adapter cleanup remains unconditional so retirement
  releases resources even when model work is refused.

`WindowRef` is process-local and opaque. Persist logical names in workspace or
replay records, then resolve them during intentional new-window setup. Captured
references fence later work; they do not make arbitrary application callbacks
preemptible or turn name-based model notifications into lifetime-scoped commands.

The native cycle/soak fixture reports
`historical_id_tombstones_retained: false` for this implementation. This describes
the live-registration contract, not a sampled Go/native/GPU memory result.
Earlier traces with the flag set to true retain their original source scope;
current campaign validators must require the alpha.2 value rather than accept
those historical traces as evidence for current window lifetimes.

## Same-name native lifecycle fixture

The opt-in window-cycle command adds `-same-name -composed -cycles=2..1000`.
This mode refuses soak and a single cycle. All openings use
`native-cycle-reused`; the default keeps its distinct per-cycle names and its
33-action native plan. Openings capture the exact reference returned atomically
by `OpenWindow`, including the separate soak path.

After each replacement reaches native readback, owner probes require a distinct
live reference, a canceled former context, refusal of old posts, closes, policies,
completion publication and reservations, and preservation/rearming of the new
dirty bit. A later owner acknowledgment checks that no old callback ran. The old
native window, GPU canvas, grid and outline are checked after replacement readback
and again after its normal 33 native actions; the replacement then closes
independently. These sampled checks do not establish continuously observed state
or Cocoa object deallocation.

The fixture retains only the immediately preceding closed opening's probe
references and releases the last capsule before the final memory sample. This
bounded diagnostic retention is part of the workload, not an empty-window memory
baseline or a full P11 memory result. One `vellum.window-lifetime` event follows
each completed close; only cycles after the first report a prior-opening check.
Terminal `same_name_mode` records the chosen workload and
`same_name_replacements_verified` counts completed replacements.

Pure Go probe/configuration regressions run with
`go test -race -tags=offscreen,vellum_native_checks ./cmd/window-cycle-smoke`.
They do not initialize a native window or GPU. Actual native readback, old resource
retirement and the composed interactions require a separately authorized live
run against the exact built source. No prior unique-name campaign qualifies this
new mode, and a same-name run alone does not complete Mac/P11 acceptance.
