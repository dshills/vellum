# Retain header focus after rejected tab close

Clicking a controlled document-tab close button now quietly restores keyboard
focus to that live header before posting the existing application close request.
A rejected close preserves the document, accepted selection and page identity;
Delete/Backspace addresses the same header next, including an unselected tab.
The fix runs in the originating native click handler. Queued callbacks drained
by a peer window and semantic close actions do not touch native focus. Current
artwork identity, live capability and enabled/visible control guards remain
required. No new public API, Tab stop, selection proposal or retry queue is added.

## Evidence and preserved failure

The fresh visibility-correlated structural native record at `90fc432` fails
after43 inputs and36 source-asserted completed states. A separate failed snapshot
at frame282 shows one rejected Notes close, both documents retained, Notes still
selected, and no actual tab-header focus. SessionIssued283 retires with Closed
true, Pending false and zero tracked windows. Of130 frame correlations,50 are
startup admission refusals and80 completed Frames; no unavailable/failed Frame
outcome occurs. The process exits1 in5.852 seconds and owned group11325 is absent.
Whole source/prepared/SDK post-boundaries pass. This record stays failed/sealed;
it does not independently qualify a completed prefix or full catalogue result.
The serial campaign stops there and never launches its text/modal case.

Core clears focus before sending a Click to its private passive close button.
The controlled CloseRequest intercept consumes that event and previously queued
only the owner callback, leaving no actual focus after rejection. A new actual
Manual routing regression fails on the unchanged implementation for both selected
and unselected rejected closes. With the repair, both cases retain actual header
and published strip focus, preserve accepted selection/page identity, deliver one
proposal per gesture, and route the following Delete to the same header.

All existing close admission, queue refusal, accepted removal, semantic action,
hidden/disabled/retired scope and stale-artwork checks remain passing. Checks use
Go1.25.6/GOMAXPROCS2, actual checkout Objective-C units, prepared stage86 and the
retained source-built786a native archive through an external manifest. The two
product Go overlay is promoted byte-for-byte; its regression receives only
gofmt formatting before the final focused check. The
checkout dependency manifests, prepared trees and SDK are unchanged.

- focused: PASS, 13.048 seconds.
- driver: PASS, 285.145 seconds.
- gallery: PASS, 34.070 seconds.
- vet: PASS, 11.855 seconds.
- build: PASS, 11.698 seconds.

These checks launch no native application. They do not convert the earlier
failed record into a pass. Full structural50/42 and text/modal68/33 qualification
still require new reviewed native scopes. OS/physical input, IME, human VoiceOver,
native pixels/presentation, resource/endurance, final SDK/rights and full Mac
completion remain open. The significant source delta receives one Prism review
with Anthropic Claude Sonnet5.5 and an internal disposition before commit.

## Review and retained records

Internal review checks the source delta and staged checkpoint, followed by one
Anthropic Claude Sonnet5.5 Prism call on the complete four-file diff. Raw retained,
discarded and suppressed findings are independently assessed before commit. Final
actual-source focused routing passes in13.881 seconds without Go replacements.
No second provider call or native rerun is implied by this evidence packaging.

The [bounded checkpoint archive](records/checkpoint.tar.gz) retains the original
failed native plan/admission/raw record, validators, private diagnostic consumer,
negative and final software chains, exact product source, provider output and
internal dispositions. [Member hashes](records/manifest.json) and
[archive limits](records/archive.json) accompany it. Binary, SDK and previous
archives are excluded. Software receipts compiled the exact promoted runtime
overlay; the actual-source focused run separately checks final formatting.
