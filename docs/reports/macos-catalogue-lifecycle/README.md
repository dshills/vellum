# Shared catalogue shutdown ownership

The older Selector/NumberField, choice, navigation and structural gallery runners
still had the retirement race found and fixed in the media checkpoint. After
`Application.Run` returned, an admitted but undispatched owner callback could be
dropped while its worker waited for a private reply. The worker then exhausted
the six-second join. Conversely, closing the last window could end Run and cancel
the scenario before the final close acknowledgement and closed-reference check.

All five modes now share `catalogueWorkerResult` and `catalogueFinishWindow`.
Run return cancels scenario waits before the unchanged six-second join. Successful
scenario teardown first requires a live scenario, then gives the owner close
acknowledgement and original-reference refusal check one independent one-second
deadline. The fatal unjoined-worker path still avoids session/application locks.
Each mode preserves its typed report, callback/state script, input payloads,
terminal error propagation and nonzero failure behavior.

Four actual Application regressions cover retirement dropping an admitted owner
callback; last-window closure with the owner acknowledgement deliberately held
until after Run-return cancellation; cancellation before terminal authority is
acquired; and retirement dropping the final-close callback, which expires its
independent deadline. The earlier media tests are migrated to the shared helper.
This closes the known cross-catalogue shutdown defect in the consumer source.

The focused race run, full gallery offscreen race suite, gallery vet and native
executable build pass using Go1.25.6/GOMAXPROCS2 and the immutable retained SDK
with source overlays. All four owned process groups are absent, logs are complete
and admitted source pins stay unchanged. [Checks and exact commands](records/checks.json),
[bounded controller](records/check.py) and [overlay](records/overlay.json) retain
raw execution evidence. No native application is launched in this checkpoint.
The earlier successful and failed native records retain their original identities;
this source change does not upgrade their qualification or replay sealed scopes.

Prepared stage86, the retained SDK and backend/runtime behavior are unchanged.
Full structural/catalogue M2, physical/OS input, native pixels, human accessibility,
M3–M7 and full Mac completion remain open. Continue the remaining structural and
text/modal acceptance with a fresh reviewed scope; unavailable frames remain
terminal, with no availability inference or retry loop.

Internal source review approves the lifetime changes. One complete uncached
Anthropic Claude Sonnet5.5 Prism call reviews11 changed file entries/26,576 diff
bytes: raw CLI0, zero retained findings and one discarded medium finding. The
tool discards its quoted evidence because it says `return false` where the actual
code returns `closed`; independently, its underlying `%v` error-chain concern is
valid. The narrow post-provider fix uses `errors.Join` to preserve an unexpected
Input cause while still rejecting inconsistent nil-error validation. Internal
review approves that fix; final full gallery race/vet/build pass again. The
[raw provider result](records/prism.json), [exact provider diff](records/prism-staged.diff),
[original reviewed helper](records/reviewed-catalogue_lifecycle.go.txt) and
[disposition](records/post-provider-update.json) preserve this distinction. No
second provider call occurs. The pre-provider source certificate is historical:
its helper and executable bindings precede the final narrow fix; final checkpoint
review binds the current source and repeat software receipts before commit.
