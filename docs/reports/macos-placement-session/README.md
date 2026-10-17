# Session-owned placement probe and failed native delivery check

17 October 2026. This checkpoint fixes cancellation scope in the opt-in
`vellum_placement_probe` executable. It watches `stop-live` only inside its own
validated private output directory. A historical test-session path no longer
controls new probes; old session seals remain intact. The controller must create
the current session marker when desktop access is withdrawn or the user returns.
`Lstat` treats any marker, including a dangling symlink, as a stop; other inspection
errors fail closed. Normal framework and example builds exclude this probe.

## Source checks

Race-enabled probe tests pass, including two-session isolation, cancellation of
the real observer, joined retirement/cleanup and marker inspection errors.
The full placement example offscreen routing race suite, probe vet and a fresh
race-enabled native probe build also pass. These are automated checks; the
routing suite does not prove physical pointer delivery. Internal code and native
admission reviews approve the scoped change and one bounded attempt. A single
Prism Anthropic Claude Sonnet 5.5 review covers all six changed source/document
files, with zero retained, discarded or suppressed findings.

## Native outcome: failed, no retry

The one current-source attempt uses clean base `3801970` plus the three reviewed
probe/example-documentation edits and unchanged stage0079 dependencies (6,260
files). A fresh source-bound plan pins all compiled source leaves, inputs and
executables, and the prepared aggregate. Its exact SHA256 is
`f334081e0789a6d3beceaadceb9d2d61459093bd6392163e3cab70c608983e0d`.
The shared live lease, fresh stop scope and 240-second outer bound were used.
No old sealed plan or SDK probe was launched. The reused operator implementation
changes only its fixed session stop path and exact target executable allowlist;
its actual read-only kernel identity self-check passes before admission.

The native probe reaches a completed Cocoa accessibility snapshot. Owned process
identification and AX frontmost activation succeed. The operator reports zero
exit after posting a click and Return to the kernel-identified owned PID, with
foreground/window/bounds guards. The expected object does not appear within the
five-second observation bound, so the campaign fails after 9.10 seconds. The
controller also fails. Posting receipts do not establish application delivery.
No cause is inferred as a framework defect from these observations.

The last sample contains no objects, no history entry, no source activation,
no canvas change, no active gesture and zero reserved queue entries; its AX focus
ID is empty. The campaign check named “pointer click only focuses palette” tests
only no object/no activation: it **does not prove focus**. The later Return has
no observed placement. Drag threshold/previews, transformed drops, native
Undo/Redo, rejection/cancellation/removal and gesture-time window closure are
unreached and remain unqualified.

After timeout the campaign terminates its owned probe (exit -15); cleanup and
finalization error lists are empty. A subsequent process-table observation finds
owned PID 95894 and group 95893 absent. There is no natural `probe.closed` record
or normal-close acceptance. No physical input, continuous preview visibility,
display/DPI transitions, input-source/VoiceOver, presentation, performance,
resource budget or complete Mac claim follows. No settings or clipboard were
changed. The fresh attempt is sealed and will not be retried or expanded into
another operator effort. Placement now needs direct/manual verification.

## Evidence

[records.json](records.json) hashes each member of [records.tar.gz](records.tar.gz).
The archive retains the exact source diff, tests and build logs, current-source
plan and build/module metadata, internal reviews, operator/campaign sources,
read-only kernel check, controller stdout/stderr/result, raw campaign/action
receipts and probe samples under `native-evidence/`, endpoint absence, assessment,
and final Prism response/disposition. Executable payloads and prepared dependency
copies are omitted; their hashes and local paths are recorded. The initial
unexecuted plan is retained separately from the reviewed v2 plan.

The cancellation-scope fix is retained, with the failed workflow outcome intact.
Continue [Mac delivery](../../../specs/macos-delivery.md) through independent
M3 measurement and M6 component/packaging work. M2 native, M3, M4, M5,
M6 rights/final SDK and M7 remain open. No native input retry is queued.
