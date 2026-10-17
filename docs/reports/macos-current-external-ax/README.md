# Current external macOS accessibility checks

A separate, already-trusted macOS process can read the current native editor's
Unicode text/ranges and the placement demo's palette/canvas tree. This advances
read-only external accessibility evidence. It does not qualify physical pointer
or keyboard input, placement workflows, human IME or VoiceOver.

## Build and runtime scope

The framework source is committed `0f636c850ae1a66263bdcc1ecbaa6d8b2a1b0e9f`, with
4,625 tracked files and 6,243 prepared dependency files. The exact framework and
dependency bytes remain unchanged through both serial executions. Both fixtures
use explicit Go overlays, retained beside their original source:

- The editor fixture disables/restores window geometry persistence and reports
  the actual flag in its ready record. Its normal lifetime remains 60 seconds.
- The placement probe shortens its diagnostic deadline from ten minutes to
  twenty seconds. Its no-input deadline failure is expected for this diagnostic.

Go's metadata reports the clean checkout's revision and `vcs.modified=false`;
that field alone does not describe the overlaid fixture bytes. The compiler
commands, complete source/prepared maps, overlays, bundle metadata and binary
hashes bind the actual instrumented fixtures. The bundles use diagnostic
identities. These runs do not repeat the earlier CUA app-resolution operation.

The native app and external client share one reviewed lease/stop protocol, with
bounded streams, readiness and process deadlines, complete EOF/hash checks,
reaped children and strict owned-group absence. No permission prompt, input,
activation, clipboard, input-source or VoiceOver change is requested. The
fixtures disable their window geometry saver.

## External results

The editor client passes exact UTF-8 values with CRLF/final empty lines,
UTF-16 emoji selection/character count, combining-grapheme ranges,
split-surrogate refusal, final-line ranges, positive emoji bounds and screen-point
lookup. Hebrew range bounds have the expected visual RTL order. This exercises
the existing native text client through the external AX API, without editing.

The placement client checks the exact owned PID, executable and window title,
then walks the bounded native tree. Its 30 observed nodes include uniquely
identified Block, Connector and Note buttons and the Component canvas, with the
expected roles, labels and positive native screen frames. It reports no
unexpected property-read errors. The four target identifiers follow the existing
`node:` plus unpadded URL-base64 namespace from the adapter's source.

| Fixture | Native lifetime observed by controller | Client lifetime observed by controller | Client's internal query time |
| --- | ---: | ---: | ---: |
| Editor | 60.688408 s | 0.242933 s | Not recorded by this client |
| Placement | 20.690507 s | 0.046955 s | 0.040682417 s |

These are single diagnostic observations, not a latency distribution or
performance qualification. Controller totals, including cleanup, are retained
separately. Editor exits normally with status 0. Placement exits with its expected
no-input deadline status 1; its final native/observer errors are nil. Every
placement snapshot retains zero objects, history entries, activations, canvas
changes, removals and reservations, with no active drag.

## Original collector failure and correction

The original placement collector remains **failed** in the evidence. Its client
returned `pass`, but the collector compared the `/tmp` spelling returned by
Foundation with a `/private/tmp` argv string. Its later target check also assumed
raw widget IDs. The adapter already encodes those IDs.

A separate derivative validates exact file identity plus the pinned binary digest,
uses the source-defined native ID namespace and additionally requires exact labels.
Every other original predicate remains unchanged. The path aliases identify the
same device/inode and digest. All four encoded targets and their positive frames
are present in the preserved client output. Independent review accepts this
corrected interpretation; no native process is replayed, and neither original
receipt nor any original reviewed input is rewritten. The initial derived
validator's missing Python import is recorded separately and corrected before the
successful reconciliation.

## Evidence and limits

[Result summary](result.json) binds the [raw evidence archive](evidence.tar.gz)
and its [member manifest](evidence.manifest.json). The archive retains the original
failed placement receipt, exact runtime producers and raw logs, the corrected
validator, independent reviews, fixture overlays and source/build identities.
Four compiled executables are excluded with their hashes retained.

This current-source evidence covers read-only text/range and palette/canvas
properties. It neither diagnoses the earlier automation timeout nor qualifies
clicks, dragging, capture loss, keyboard event routing, placement Undo/Redo,
external notification delivery, human accessibility, multiple displays or full
Mac completion. No CUA retry is performed. The existing ordinary performance,
native/GPU memory, full P4/P6/P9/P11, idle/overnight, interop and distribution gates
remain open. See the [Mac completion contract](../../../specs/macos-completion.md).
