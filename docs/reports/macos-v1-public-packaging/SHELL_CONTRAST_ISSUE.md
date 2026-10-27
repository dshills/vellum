# Held fixture: Mac Shell dark-system contrast

The exact final `Vellum Mac Shell.zip` hash is in [SHA256SUMS](SHA256SUMS).
Both windows launch and close, but their body text is nearly black over a dark
native background. Both screenshots (`native/shell-a.jpg`, `native/shell-b.jpg`)
and AX trees are preserved in the evidence archive. Raising the B window did not
correct contrast. This fixture is excluded from RECOMMENDED_SHA256SUMS.

## Reproduction

On the observed dark-appearance desktop, extract the recorded ZIP, verify its
signature and open Vellum Mac Shell.app. Inspect both windows' title/status/
checkbox labels. Do not change a global appearance setting to reproduce it.
The existing example sets `cogent.Driver{Theme: ui.MacTheme(false)}` while both
`app.Window` values omit explicit native chrome appearance:
[fixture source](../../../examples/macapps/cmd/macshell/main.go).
That mismatch is the source-based explanation; a corrective build has not been tested.

## Smallest next action

Make an app-only fixture correction: explicitly select Light native appearance to
match its existing fixed light managed theme, or synchronize both through the
existing system-appearance path. Run a focused fixture check and Prism review,
commit, then rebuild/repackage only Mac Shell. Record its new pre-sign binary and
notice bindings as an audit delta before recommending the replacement ZIP. Reuse
all unchanged SDK/other-app verification. Do not relabel this current ZIP as fixed.

No runtime code change was made in the final-packaging task: the owner's exact
approved payload was preserved, and the fixture is held instead of silently
replacing its audited binary. This is a visible defect in that optional fixture,
not a new framework feature or a changed performance gate.
