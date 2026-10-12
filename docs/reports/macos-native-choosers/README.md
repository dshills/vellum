# Real macOS document chooser acceptance

The existing ordinary editor passes an external native Open / Save As / Cancel
workflow on owned temporary files. Both windows then close normally with process
exit 0. This supplies actual panel-selection evidence beyond the earlier
injected-chooser document smokes. No production editor or native-service change
was needed.

The qualifying run, `live-v20`, uses a fresh optimized editor build whose module
metadata identifies `41a4888`. The build receipt binds selected Go, Objective-C
and C/header inputs plus all 6,167 prepared dependency files. These selected
source and prepared identities remain unchanged through the run. The app launches with no smoke flags, positional paths, injected
chooser or store. An owned temporary `.app` wrapper preserves the exact executable
bytes and gives native automation an identifiable application. Swift is an
optional external acceptance client, not a runtime dependency.

## Verified workflow

The final run creates a unique directory, input/destination basenames and input
text. They cannot match the earlier attempts' fixtures. The destination is absent
before selection. The client checks existing accessibility trust without a
prompt, then restricts AX actions to the launched PID and exact executable.

- Open uses the actual Cocoa `Open document` panel and its Go to Folder input.
  Native Open installs the exact unique file text and basename. The input bytes
  remain unchanged; Open/Save As are enabled again and Cancel operation disables.
- Save As selects a new filename in the unique document directory already shown
  by the native panel. Native Cancel restores the original filename/text and
  controls, reports cancellation, and creates no destination file.
- A second Save As accepts the same previously absent destination through native
  Save. The expected absolute file is created with the immutable original fixture
  bytes, and the editor reports the saved basename and successful completion.
- Native Open Cancel preserves the saved file, filename and document text and
  restores the controls. This observes actual panel cancellation.
- The first window closes through its native close button. The remaining window
  stays usable, with its empty text and Ready state unchanged. Its own native
  close button ends the application normally. Standard error is empty.

There are 11 ordered observation stages and 47 hashed command/output receipts.
The 150-second harness completed in 8.49 seconds. It observes completed UI states
after action admission; an AX action returning success alone does not qualify a
step. Save acceptance is also verified at the actual absolute path. The input
and saved file hashes equal the original fixture hash; the oracle does not use
universal-newline decoding or compare only two potentially mutated files.

## Harness corrections and evidence limits

Earlier attempts remain historical records. The initial desktop tool returned
after the 420-second watchdog had retired its process; its late UI result cannot
qualify a workflow. A manual CLI attempt similarly exhausted its watchdog. Later
bounded corrections addressed asynchronous registration/publication, foreground
activation, native Go to Folder focus, separate filename/tag fields, compact Save
panels, and survivor lookup. Older repeated names/text cannot alone prove an
absolute Open path.

The earlier full workflows `live-v16` and `live-v17` remain observations from their
own harness generations. Internal review subsequently found gaps in critical
window reads, late error capture, focus checks, byte comparisons, recursive
retries and failed-command records. They do not substitute for the corrected
`live-v20` qualification. The stricter `live-v18` and `live-v19` attempts timed out
before file operations and were retired. Posting to the owned PID did not
activate the native Go sheet in `live-v18`; normal foreground routing alone did
not settle the transition in `live-v19`. The final client adds bounded panel
activation/event-delivery intervals and records the focused window. This is
harness evidence, not a demonstrated framework defect.

Critical AX window membership and title reads preserve error status, including
unscoped teardown observations. Failed reads do not imply a missing window. Dump
records include errors from the completed state reads. A known missing or
transiently unavailable scope before action dispatch can be polled within both
the overall deadline and a 64-attempt limit. Uncertain or admitted actions are
never replayed. Command records begin before dispatch, retaining raw outputs on
child timeout, launch failure and malformed JSON. Eight fault probes verify
these failure paths, bounded retries and rejection of line-ending mutations.

Go to Folder uses only Cmd-Shift-G on the explicit Open/Save panel and Return on
the explicit Go sheet. The helper raises that target, waits briefly, and records
its focused-window identity. Both qualifying shortcuts observed the expected
panel/sheet as focused. It checks and records the owned foreground PID immediately
before each key-down and key-up event. These use normal global foreground event
routing; the focus check and event post are not atomic. A second Return is only
attempted while the live Go sheet and entered path are still observed, with that
sheet required again by the helper. No clipboard injection, permission prompt,
input-source, VoiceOver or display-setting change is used.

Save As normally defaults to the opened document's directory. The client verifies
that unique directory in the panel and changes only `saveAsNameTextField`, leaving
the tag editor alone. It navigates only if necessary. Private session
path/revision/result counters are not exposed externally, and no claim is made
about those internal values.

Separate internal source and evidence reviews cover this acceptance checkpoint
between significant Prism reviews. The archive includes scoped qualifying
records, inert acceptance-client/build recipes and historical outcome/hash
summaries. Earlier broad UI snapshots remain private; executable binaries,
application bundles and compiler caches are excluded. Per-member hashes bind the
published records.

This closes the observed actual editor panel-selection/persistence/cancellation
workflow and normal two-window teardown at that scope. Native menu activation,
general clipboard preservation/roundtrip, physical Finder drag, human IME and
VoiceOver, cross-display, sustained/native-memory and full performance/public
SDK distribution acceptance remain open. Windows/Linux qualification is deferred.
