# Item 3 — coordinated Quit data-safety regression

The existing Save-first/Cancel-last regression now verifies distinct exact Unicode
notebooks, including CRLF, not only note counts and dirty flags. After A approves
Save and B cancels Quit, both original window references must remain live, both
holds must be released, A must be clean, B dirty, and every live Note value must
match its captured pre-Quit value. The first file must decode to all approved A
notes. B's initially absent primary file must remain absent; cancel cannot save it.

All application/owner errors in this path are checked. The asynchronous file
worker is still drained to its existing three-second bound; missing second
prompt fails rather than being reported as a cancellation pass. The complete
focused Quit race population passes under Go 1.25.6 and offscreen routing.

No production change or new stable API is introduced. This is real temporary
file/model/lifecycle evidence, not a native dialog-input pass. The build24 native
Save-A/Cancel-B case remains pending because key-window focus after the earlier
native chooser was not reliably established by the external desktop tool. The
user has been asked to foreground the admitted app or explicitly select offscreen
scope. Neither a native pass nor completion of the five-priority acceptance plan
is claimed until that scope is resolved.

The initial default Prism review reported one medium diagnostic-context finding
and one low expected-content weakness. Both were fixed: primary paths are checked,
read/decode errors identify A and its path, and new-note expectations are composed
from literal Unicode/CRLF values before Update. The focused Quit race population
passes after correction. A second default staged Prism review completed with
full coverage, exit 0 and zero findings. Original and final receipts are retained
as item-3-prism-initial.json and item-3-prism-final.json. This commit completes
the regression work only; the native item remains pending foreground access.
