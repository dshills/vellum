# Document failure audit — internal review

Audited FileStore's sync/close/rename/directory-sync and cancellation sequence,
Session's Written versus acknowledged result, and Notes primary/recovery/session
metadata transitions. Existing focused cases cover pre-rename failures preserving
original bytes and cleaning temporary files; post-rename errors; canceled/stale
owner completion; malformed/missing restore records; recovery precedence and
recovery-marker/metadata failure refusing closure. No persistence runtime change
was necessary in this phase.

Added two missing application-level regressions. A real replacement followed by
an injected durability error must retain dirty history, the existing recovery
snapshot and its live window. A successful retry clears recovery and preserves
Undo/Redo around the saved revision. A failed recovery checkpoint (real directory
in place of its file) must remain retryable after repair, capture the new edits
and leave primary notebook bytes and dirty history unchanged. All inputs are
private temporary fixtures; no user's notebook or global settings are touched.

Internal review verified error injection occurs after a real successful primary
write, checks both recovery and primary disk bytes, and does not merely assert
a status label. The checkpoint retry checks both the persisted snapshot and
scheduler bookkeeping. Existing root fault tests remain the evidence for actual
FileStore sync-stage injection; no power-loss guarantee is inferred.

Validation: focused Notes race tests for new cases plus FailedSave,
WindowSession and RecoveryMarkerFailure passed with Go 1.25.6. Focused root
document race tests for Atomic, FailedInstallAndWrittenError,
CancelAfterAdmission, WindowClose and ChooserFailure passed. Notes vet and diff
whitespace checks passed. These are owner/worker and filesystem tests with no
native driver initialization or desktop access. Physical launch/input checks
remain deferred. Stable APIs, formats, rendering and release gates are unchanged.
