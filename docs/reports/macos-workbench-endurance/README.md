# Composed Mac workbench activity

The workbench inspector now exposes controlled value TextField, package
RadioGroup, rotation SegmentedControl, disclosed FormGroup and StatusBar.
Property requests use its existing sparse document journal. Tests cover full
uint32 values, invalid-request atomicity, undo/redo and current owner selection.
Disclosure expansion is transient UI state; document persistence is unchanged.

A new dedicated `-endurance` mode retains private raw observations, Go execution
trace and final post-GC heap profile. See the
[contract](../../../specs/workbench-endurance.md). Queued owner callbacks are
revoked after timeout; already executing callbacks and worker/owner completions
join before output closure or native shutdown. Geometry reads use the native
window mutex. Output paths must be new; trace/heap limits are 256/8 MiB.

## Actual ten-minute diagnostic run

One optimized default-renderer native process uses one million stored records
and 100,000 outline nodes. After initial GPU readiness, 600.113 seconds of
profiled composed activity completes 2,384 steps. Each step selects a component,
checks sparse property undo/replay, moves the legacy table viewport, changes
canvas fit/zoom and invalidates the native view. These are owner/model activities;
they do not prove a completed pixel update for each canvas mutation or synthetic
scroll event delivery. Periodic actual Cocoa value requests yield 596 edits
confirmed in the accepted domain model. New semantic generations can revoke
individual requests; bounded retries prove eventual acceptance, not delivery of
every request. Both 1280×900 and 1100×780 effective window-manager/pixel sizes are
sampled across 198 resize requests; requests are not individual resize receipts.

All 1,787 workers terminate and deliver owner completion: 597 fresh results pass
checksum/revision checks and 1,190 stale results are discarded. The final task
slot count is zero. The final owner snapshot has zero ordinary messages,
reservations, ready completions and cleanup messages; Dispatching=1 is the
currently executing observation callback. This is not a claim of render idle.
The native process exits successfully and the controller joins.

119 periodic Go observations record 6–8 goroutines, 52.4–100.9 MiB sampled live
HeapAlloc, and 140.3–146.9 MiB GoSys. GC counts increase from 11 to 306. Five
vmmap/ps observations show 390.3–399.9 MiB physical footprint and 256.4–258.9 MiB
RSS. IOSurface virtual-region sizes are 10.3 MiB four times and 13.8 MiB once.
These observations include application, profiler, caches and opaque driver memory;
no native heap or framework-controlled GPU allocation is attributed. A finite
sample range is not proof against long-session growth. The execution trace is
143,656,154 bytes, with its final post-GC heap profile retained separately.

This is one profiled diagnostic run at its recorded geometry, not five ordinary
latency runs or the original 1920×1080 roughly-1,000-visible-control fixture.
Owner round-trip observations are not CPU execution, frame work or presentation
latency. Dropped-frame/input-backlog/native-allocation/GPU attribution is
unavailable. Existing P6 targets remain failed; all-component primary/display
variants, integrated 1,000 lifecycle cycles, overnight idle/background and full
P11/Mac/release acceptance remain open.

## Corrections and review

Early native preflights checked Cocoa queue admission before model delivery and
failed. The selected harness reinspects/retries and confirms domain acceptance.
Internal review also found queued timeout callbacks could execute after resource
closure and native geometry getters read unlocked fields. The final source
revokes queued work, joins running work and uses the geometry writer mutex.
Tests preserve the timeout-delivery and running-callback regressions. Final vet
also caught a missing operation cancellation on a geometry-read refusal path.
That branch is corrected after the ten-minute run; the v9 source vector is
retained separately. The final native short run checks the corrected source.
A later Prism finding also showed controller cancellation could hide a native
loop failure. Final shutdown joins the controller and preserves the native cause
alongside secondary errors; its regression covers cancellation and join timeout.
Steady activity is unchanged, and no claim equates the distinct binary hashes.
An ambient profile-summary invocation could not find pprof; the controlled SDK
invocation parses both scheduler and heap profiles successfully. All failed
preflights and intermediate build/check receipts remain in the evidence archive.
One repeated preflight used unchanged code after a runner path error; it is
retained separately and is not selected evidence.

The checkpoint requires final adapter race/vet and execution-trace validation,
internal source/evidence review and one bounded Prism review with Anthropic
Claude Sonnet 5.5. Raw artifacts and exact reviewed sources will be archived with
a member hash index. No OS input source, VoiceOver, clipboard or persistent
window-geometry setting was changed by this fixture.
