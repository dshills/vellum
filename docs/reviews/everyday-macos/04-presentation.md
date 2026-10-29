# Existing Mac presentation and focus — 2026-10-29

Reviewed an actual installed Documents/Notes-layout window capture with selected
prose, native full-color toolbar, folder/list spacing, selection and typography.
Forward/back Tab traversal returned focus to the Note text AX element. The
native Save/Open round trip had dismissed both panels back to the parent. No
unrelated click, text or global preference changes were used.

Corrected Documents native chrome to explicit Standard/Light, matching its
existing fixed Light Notes theme and the Notes/Forge executables. Previously
its default native appearance could diverge from its fixed Light client under
a Dark host. This aligns the existing presentation; it adds no theme or widget.
The correction is checked again in the final packaged executable.

## Internal review and checks

Reviewed app-local appearance ownership, chrome/client consistency, public
Window configuration, note selection/focus lifetime and original-color artwork.
Passed the existing rendered Notes layout at two sizes; Forge geometry at three
DPIs and three sizes while reproducing a process-local dark host palette; Forge
focused Backspace routing; and Documents active-command/lifetime tests, all with
the race detector. No new framework API or dependency patch. No external review.

Apps in this bundle intentionally use Light presentation. Full-app Dark mode
is not added or claimed; existing framework palette tests remain distinct from
native app observations. Human IME/VoiceOver and other hosts remain unqualified.
The native helper's observations and captures are kept in the private delivery
record; a rendered layout test is not physical pointer qualification.
