# External files and recovery — 2026-10-29

Installed build20 passed the actual native Save As/Open panels. Save As navigated
to a disposable external directory and wrote seven notes; adding a note and
Command-S updated that external file to eight while the private default remained
seven. Quit/restart returned to the private default, and native Open explicitly
reopened the eight-note external file. See 03-files.json. No user documents or
clipboard contents were changed. AX filename assignment is not folder navigation:
the check used the native Go to Folder control after the initial harness attempt
created a literal colon-containing name inside the disposable directory.

Added a focused app-owned regression through the registered Save As/Open command
path and document workers. It verifies current-directory chooser options, adopted
path, later Save without another chooser, preserved private bytes, explicit
restart/Open behavior and fresh history with exact Unicode content. Reused
recovery restore/save, corrupt-open, failed-save and stale-edit tests; all passed
with the race detector. Existing recovery limitations remain unchanged.

## Internal review

Reviewed command settlement, worker/owner boundaries, current-path adoption,
separate recovery state and original-default preservation. Corrected the test's
missing external parent directory; this was a fixture failure, not an app defect.
Bounded chooser observations and nonblocking fixture capture prevent a missing
chooser from hanging tests. Model assertions run on the owner. No production
behavior, API or feature change was needed. Native observations are distinct
from the automated regression and from human accessibility qualification.
