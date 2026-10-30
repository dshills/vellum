# Lifecycle boundaries — internal review

Two defects reproduced before fixes in deterministic owner/worker regressions:

- Quit A awaits native settlement; B retires; another Quit aborts the original
  epoch. A late successful acknowledgment formerly reopened an orphan Quit
  prompt. Recheck the epoch/liveness before flush or decision creation and
  return cancellation for obsolete work. Duplicated old acknowledgments cannot
  disturb a new independent Quit; dirty edits and surviving windows are retained.
- Finder Open queues behind window-session persistence. If that write failed,
  pendingExternal was left retained with no continuation. On failure explicitly
  cancel/drop the dependent request and report retry Open. Do the same cleanup
  for a deferred Save-then-Open destination. Do not silently replay after an
  unrelated metadata write, bypass a dirty decision or retry unboundedly.

The two regressions failed against the prior source and pass after the fixes.
An overlap test also confirms cold-marked Finder requests cannot bypass an
active Quit and that retry still uses the dirty Open/Cancel path. Reviewed
callbacks remain owner-confined, epoch checking occurs before mutation, cleanup
precedes returning from the metadata failure callback, and no native backend or
stable root API was changed. Existing document.Session window-scoped completion
retirement remains the guard against late installs after closure.

Validation: full dependency-free Notes Go 1.25.6 race suite and Notes vet passed;
all existing Quit, recovery, session-restoration and Finder tests were included.
Diff whitespace checks passed. Tests initialize no native driver/window and use
only private temporary files. No physical Finder, native dialog or desktop
acceptance is inferred. Those checks remain deferred by the owner.
