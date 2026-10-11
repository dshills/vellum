# Supported public editor acceptance

This phase adds the portable `editor.Model`, managed `ui.MultilineText`, one
shared document journal and an independent two-window document application.
Text preserves exact UTF-8, CRLF and terminal blank lines. Preedit stays outside
history until an originating-window settlement commits it. Save and Close
settles again after IO so input admitted during a write cannot be discarded.

## Verification scope

[Acceptance](acceptance.json) records the frozen source inputs, binary identity,
commands, exit status and evidence hashes. The [full repository gate](full-check.json) and [final driver correction checks](driver-final-checks.log) pass, including the driver race suite, vet,
portable model/contracts and independent consumer. The corrected field
handoff, refocus and active-client regressions also pass 20 repeated race runs. Reproduced pinned
backend tests use offscreen initialization for exact text and shaped geometry.

Two consecutive [native](native-smoke-1.log) [runs](native-smoke-2.log) pass against
the same race binary and unchanged [source inputs](source-inputs.json). They cover
composition/history, controlled-field handoff, exact-byte save/open, independent
windows, bounded admission, wrapped/RTL geometry, native input during Save and
Close, and terminal teardown. These use deterministic Cocoa protocol injection.

A separate read-only [external AX process](external-ax-final/acceptance.json)
passes against that same binary. It checks exact bytes, UTF-16 selections/ranges,
invalid surrogate rejection, screen bounds, point lookup, terminal blank lines
and RTL glyph positions. The 60-second fixture closes normally with exit0. The
inspector uses existing trust and changes no text or OS settings. Human IME
conversion/candidate panels and VoiceOver editing remain separate acceptance.

## Historical failures and corrections

[The startup diagnostic](historical-native-startup.log) showed a correctly
focused accessible editor but no active Cocoa input client. Ordinary Custom
wake compression had replaced the deferred publication node, leaving later
publication requests stranded. [The pre-fix regression](historical-deque-before.log)
reproduces that failure. Pinned patch27 protects deferred FIFO nodes while
retaining ordinary compression and priority window events. Patch26 publishes
legacy edits at accepted event boundaries before following commands or input.

[A fresh two-window run](historical-native-main-thread.log) exposed a background
Cocoa call through the backend's pre-loop inline fallback. A queued main-loop
readiness gate now prevents premature native calls without blocking UI-owner
construction. [The field-refresh failure](historical-field-refresh.log) exposed
retained widget refresh discarding active preedit; pinned patch28 preserves it.
The Find fixture also initially awaited controlled Change before focus loss;
it now follows existing field semantics. These historical logs are not passing
acceptance records.

[Post-review run 11](historical-post-review-native-11.json) failed while awaiting
Undo delivery. [The diagnostic run](historical-action-diagnostic-run-12.json)
showed the operation executing and passed. Native action admission acknowledges
queue admission; it does not prove delivery after a semantic generation changes.
The fixture now checks fresh accessibility publication before sending one action
and observes actual owner delivery without retries. The earlier failure was not
a demonstrated model/history Undo defect.

[Run 13](historical-focus-handoff-native-13.json) then exposed a real controlled
field ordering defect: an ordinary view wake applied focus loss before previously
admitted native composition records reached the owner. Pinned patch29 and the
adapter defer the field's final Change until outgoing native input, including
unmark, is acknowledged and the client is inactive. Only then is the next focused
client published. [The core regression](field-focus-loss-core-red.log) and
[adapter regression](field-focus-loss-adapter-red.log) fail before this correction.
[A second regression](field-focus-loss-refocus-red.log) reproduces a late unmark
erasing a fresh refocus selection; same-visible sealing now preserves it.
[Repeated corrected tests](field-focus-loss-repeat20-final.log), the full driver suite
and final live runs pass. This is an admitted-native-field handoff guarantee;
programmatic focus requests retain existing legacy key-event routing semantics.

The [historical pre-review acceptance](historical-pre-prism/acceptance.json) and
its raw evidence remain in a separate directory. [Final source proof](source-inputs.json)
records 218 source, 108 prepared and 26 qualification inputs. The final two native
runs and external AX run use binary
`75a2d76b8c3015b6ad0667ffeb2c2ec35184dd1f65f27dd9382eb2b77cab9c97`.

[Initial external AX](external-ax-initial.json) passed against an earlier binary.
Its [source identity](historical-external-ax-source.json) is historical evidence,
separate from final-source acceptance. No input sources, VoiceOver or privacy
preferences were changed in any of these checks.

The targeted Prism review led to three final corrections. Republish of the same
client now writes its actual active flag once; [the earlier code fails this
observer regression](field-atomic-snapshot-red.log). Outgoing retirement publishes
an empty inactive client rather than the outgoing field's text/geometry, preserving
the native acknowledgment gate even when legacy Go editing state is invalid.
[The final correction record](field-final-corrections.json) includes the checked
main-dispatch payload fixture and test identities. The standalone Cocoa test pins
the process startup thread during test initialization, as recorded by its
[init-thread check](input-init-review.json). Independent review approved these
narrow corrections, followed by final-source repository, two native and external
AX checks. [Earlier source13 passes](historical-correction-pass-13/record.json)
remain historical evidence.

## Limits

These checks do not certify sustained performance, retained process memory,
physical monitor/GPU-loss workflows, human assistive technology or public
distribution. The native editor is bounded to 4 MiB; portable models may choose
larger bounds. See the [editor contract](../../../specs/editor.md) and
[Mac completion audit](../MACOS_COMPLETION_AUDIT.md). A clean committed private SDK
must separately reproduce pinned dependencies and run the independent editor
and external AX after relocation; its commit identity cannot be embedded in
this earlier precommit evidence.
