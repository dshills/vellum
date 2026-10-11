# Managed software UI replay checkpoint

The public `replay` package provides bounded value records and a synchronous
runner. `drivers/cogent` adds an `offscreen` manual adapter that uses the real
retained widget router and software compositor. The independent
[`examples/replay`](../../../examples/replay/README.md) module freezes one
51-operation record and nine captures; batches 1, 7 and 64 agree on domain state,
actual focus and directional UTF-16 selection, measured geometry/semantics and
completed image digests. Event/frame transport counts may differ by batch.

The fixture edits a project property through physical pointer/key input, sorts
and selects visible rows, cancels and explicitly delivers a worker result, then
undoes the edit. Command-only fixtures and record authoring are separate from
acceptance of the bundled record. Modal/disabled input, stale lifetimes, readiness,
frame admission and closure have separate actual-router tests.

Caller-supplied font bytes, theme/DPI and one shared explicit clock identify the
rendering inputs. Only explicit clock advancement and selected-worker steps
execute future work. `RequestFrame` enqueues paint; bounded `Step` charges
originating events plus owner callbacks. Capture is a fresh read of completed
output. Future timers are reported without forcing a time advance; current
blocked, running and awaiting work cannot become successful quiescence.

Native preparation stages 30 and 31 add origin-owned input/render scheduling and
widget timing. Completion debounce, snackbar/popup expiry and cursor painting use
the scene clock. Text buffers accept one checked live-view execution lease;
manual markup runs in originating dispatch. Shared/conflicting leases, existing
ambient work, output readers and filesystem autosave refuse instead of launching
unaccounted work. Desktop execution retains its real-clock behavior. Stage 32 preserves its
repeat-click render ownership/update bookkeeping, drops retired or moved timer
targets, checks frame-request extension support and retains loud rejection of
invalid output-buffer setup. Separate regressions cover these production paths.

Independent internal review found and resolved constructor cleanup, process
ownership, closed-window observation, unqualified manager shortcuts, stale
manifest incarnations, closed-owner global worker admission, residual readiness,
invalid times and retained JSON backing-capacity defects. Historical red evidence
is preserved separately from accepted frozen proofs. Screenshot, zoom, refresh
and other unqualified manager shortcuts now refuse before queue admission.

[qualification.json](qualification.json) binds final integrated source inputs,
commands, exits, prepared output hashes and actual native editor regression.
[evidence.json](evidence.json) preserves raw logs and proof inputs as exact,
round-trippable UTF-8 strings. [component-evidence.json](component-evidence.json)
retains independently qualified component checks and historical review evidence.
Full upstream vet still has its recorded pre-existing warnings; own-module vet
checks are distinct. The combined source/specification review uses Prism with
Anthropic Claude Sonnet 5.5; its ledger is in
[the review directory](../../reviews/macos-replay/README.md).

This checkpoint supports one active manual application per process and multiple
windows sharing its clock. Callers hold Cogent global settings/palette/keymap/
highlighting fixed and inject the borrowed executor into domain managers.
Arbitrary caller goroutines, callback CPU time and custom extensions are outside
its deterministic accounting. It does not qualify the complete P9 professional
workload, large-data views, external automation, Cocoa IME/VoiceOver, external AX,
GPU presentation, cross-display behavior, sustained performance or public release.
Mac input sources, privacy and accessibility settings were not changed.
