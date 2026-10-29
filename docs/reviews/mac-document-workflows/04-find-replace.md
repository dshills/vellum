# Notes Find/Replace — internal review

Integrated existing FindBar/editor operations into Notes, Documents and generated
Document Starter; no root API or rendering change. Compact Mac controls retain
space in the writing column. Native Edit menus and active-window shortcuts use
the same model commands. Prose focus returns on Escape/Done.

Reviewed selected-note lifetime, Unicode scalar offsets, history/save identity,
composition, bounded fields, settled callbacks and native field ordering.
Review corrected two input-ordering hazards: the Find fields must remain able
to admit native Change while an action awaits settlement, and Command-G must
settle a newly typed query even before its controlled value is published.
Queries/replacements are read after the FIFO acknowledgment. A retained action
cannot replace a different selected note. Shared app-host settlement sequences
fields/body once and stops on failure; duplicate completions cannot advance it.

Focused application race tests and vet passed, covering Unicode navigation and
wrap, no-match/stale/composition refusal, latest controlled fields, one-step
Replace All Undo/Redo, repeated acknowledgments and focus/tree validity. The
generated-app recipe is checked separately; actual native shortcuts and final
layout remain packaged-app observations. No external review service was used.
