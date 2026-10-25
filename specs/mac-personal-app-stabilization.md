# Personal Mac app stabilization

Keep framework features, backend policy and qualification gate frozen. Work on
five finite checkpoints: (1) repair the concrete merged-main CI failures,
(2) classify the four dependency alerts and update the maintained legacy Gio
prototype pins where needed without changing original native notices,
(3) finish the Dark-bundle launch and physical Find checks in a newly authorized
desktop window, (4) make the existing Notes consumer durable using public
document/history/editor APIs, (5) document the consumer architecture and package
the completed application. Review checkpoints with Prism/Anthropic Claude Sonnet
5.5, run focused checks and commit separately.

Notes remains bounded plain text with 64 notes, existing folders and unified
Undo/Redo. Add a versioned, strictly validated whole-notebook file; native
Open/Save/Save As; a private local notebook path; atomic replacement via existing
FileStore; close protection; bounded interrupted-session recovery with explicit
restore/discard and safe error behavior. Worker IO receives detached snapshots;
all model/history changes belong to the owner. Native input settles before save,
open and close. Invalid files and late results leave the live notebook unchanged.
Undo history stays session-local; opening/recovering restores note identities and
clears stale undo closures. No rich text, accounts, networking, new framework
widgets, editor-presentation API or rendering experiments. Preserve original
icons. No human IME/VoiceOver qualification is added.

Existing full portable race/vet and focused native compilation cover CI fixes.
Remote CI verifies the older macOS SDK and actual Windows behavior; local tests
do not claim those environments. Native pointer checks are separate from routed
regressions. Use disposable app files in live checks; preserve global settings.

## Notes visual correction, 25 October 2026

Use the owner's Apple Notes comparison for proportions and presentation only; do
not copy personal reference contents into fixtures, reports or reviews. Correct
the existing app's compact folder/count rows, list density and selection colors,
search sizing, dividers, title sizing and unused header space. Keep document
bytes, stable control/editor identities, settlement and history unchanged. Use
standard native chrome with no client header reservation. Assert allocated
geometry in offscreen regressions before packaging. The code-editor body font
and gutter require a separate scope decision; they are not represented as solved
by app styling. No new framework controls or renderer policies are introduced.

The owner explicitly approved the focused prose-editor exception on 25 October:
existing `MultilineText` may opt into proportional typography without the code
gutter. Keep the existing document model, native input, settlement and app Undo
path. This exception does not reopen widget or rendering-policy scope. The Notes
consumer must use an accurately identified private SDK snapshot containing that
change; the earlier frozen SDK cannot compile the new presentation option.
