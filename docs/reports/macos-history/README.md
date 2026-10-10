# macOS document history and native command lifecycle

The framework now exposes a bounded document journal, stable dirty/save revision
identity, sparse workbench property changes and document Undo/Redo after focused
native text handling. A label-only native window also receives both shortcut
phases. The independent consumer exercises the public history/command contract.

The fresh integrated check exposed a native startup shutdown deadlock: StopMain
captured a nil stop channel before loop publication. Owned patch21 retains a
stable close-once stop signal, acknowledges accepted refusal and preserves
synchronous callback output. Native window creation is one admitted callback;
geometry constraining uses direct helpers to avoid recursive queue dispatch.
The initial failure and final passing logs are retained separately.

Prism's first batch review found two medium items. Source inspection and an
independent internal review refuted the synthetic-event claim: window lookup and
event enqueueing use their locks, and KeyChord does not mutate shared source state.
The stop wake ordering finding was valid and is fixed: StopMain does not post a
native wake, and the desktop loop's existing 100 ms event wait observes shutdown.
A nil identity/forbidden wake regression and all final checks pass. The follow-up
Prism review covers this corrective delta rather than re-reviewing unchanged code.

Fresh pinned dependency preparation and ./scripts/check.sh pass. The latter
includes 86 Python regressions, portable/native race tests and vet, selected
backend race tests, native build and independent consumers. Main-queue tests
also pass 20 race-enabled repetitions. Four serial native smokes pass: text and
document history; 100,000-row workbench/GPU readback; last-window handoff; and
float/redock/close session persistence. See source-and-check-proof.json for
commands, exit status, exact input hashes and internal review dispositions.

The synthetic native chord test observes AppKit focus publication; it does not
establish physical input/IME or external screen-reader acceptance. OS preferences
were unchanged. Document persistence, unsaved-close workflows and the other
Mac completion gates remain open in specs/macos-completion.md. A preexisting
callback-panic acknowledgment gap is separate from this shutdown fix.

Final SDK integration also fixes the exporter whitelist so the new history
package ships. The added regression first failed with the missing package, then
passed after the export fix: it creates an archive, relocates it, deletes the
original source and runs Undo/Redo through a separate Go module. All 26 exporter
regressions pass. This focused archive check uses fixture native dependency
metadata and does not regenerate or qualify a complete native SDK release.
