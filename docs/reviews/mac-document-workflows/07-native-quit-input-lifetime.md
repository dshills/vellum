# Native Quit input lifetime — internal review

The first installed build22 native Command-Q reached the app-owned callback but
aborted with `cogent: native text client is stale` before its dirty prompt. Both
windows and their fixture edits remained open. The Quit hold disabled native
text nodes during their required FIFO settlement, retiring those client IDs.

Keep search/title/body inputs enabled during the coordination hold. Structural
commands remain held; the actual modal decision still disables underlying input.
Late edits continue through normal history, whose approved-revision validation
refuses a stale close. This is an app-only lifetime correction, not a driver
policy relaxation or a release-gate change.

Internal review checked the hold/settlement/modal/Cancel transitions. A focused
regression now checks admitted tree input state while the acknowledgment is
pending, structural command refusal, modal disabling and Cancel restoration.
Existing Quit, settlement and Find race regressions are reused alongside it.
Repackaging and native retry are required before accepting the candidate.
