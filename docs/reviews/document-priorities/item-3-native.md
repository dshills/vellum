# Item 3 — completed native Save-A/Cancel-B check

The admitted build24 Documents app ran against the existing private synthetic
A/B fixture on macOS 27.2. Both notebooks began with seven saved notes. Each
received a distinct eighth Unicode note. Command-Q opened A's coordinated Quit
prompt. Save wrote A, leaving A visible with held controls while B's decision
was pending. The user selected B; Cancel preserved B's eighth note and exact
Unicode body, returned Edited status and enabled its toolbar. Independent reads
confirmed A had eight saved notes and B's seven-note file was byte-identical to
its pre-Quit hash. B was then saved normally to resolve that dirty decision safely.

The user finally selected A. Its native snapshot confirms eight notes, the exact
A eighth-note title/body, Saved notebook status, enabled controls and editor focus.
Both documents therefore survived the canceled Quit in their expected states at cancellation (A saved, B dirty), with the later B Save explicitly separate from cancellation acceptance.
[Structured observations](item-3-native-result.json) preserve the sequence and
saved closeout hashes. The earlier regression and default Prism review remain
in [item-3.md](item-3.md).

App-bound window switching did not select the other window and a Dock binding
timed out; there was no unrestricted input fallback. Window selection was human
assisted, while the inspected prompts and edits used app-bound CUA actions.
This is not a physical-input, process-memory, release or current-source rebuild
qualification. No app code, build artifact, personal notebook or global setting changed;
the synthetic fixture notebook files were saved as described.
Restart and recovery are the next, separately recorded item.

Default Prism review completed with full coverage and exit 0. Four low documentation/evidence-structure findings were addressed by distinguishing memory from disk, explicitly mapping A/B states, adding structured B post-cancel observations, and clarifying build-artifact versus fixture writes. [Original receipt](item-3-native-prism.json).
