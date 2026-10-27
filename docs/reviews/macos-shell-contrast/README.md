# Mac Shell fixed-light contrast correction

The owner authorizes fixing the held Mac Shell contrast issue, focused testing,
Prism/Anthropic Claude Sonnet 5.5 review, a commit and replacement packaging.
Features, stable APIs, dependencies, backend and release/performance policy remain
frozen. The fixture already uses ui.MacTheme(false); both native windows now
explicitly use Light appearance, and their root surfaces use the same theme’s
window palette color instead of inheriting a dark backend surface.

Only examples/macapps/cmd/macshell/main.go changes product behavior. Existing
commands, per-window toolbar ownership and lifecycle are unchanged. This is a
fixed-light fixture, not an automatic Light/Dark theme feature. No global setting
or user notebook is accessed. App instructions document the intended behavior.

An isolated copy of the verified public SDK (8,856 listed hashes checked) differs
in exactly that one source file; the source-built native archive hash786aaad8…
is unchanged. Focused final executable offscreen compilation/race wiring and vet
pass. Existing window policy ownership/validation and unsupported offscreen
native-chrome refusal regressions also pass. The first two checks cover the
initial chrome-only candidate; final compile/vet include its explicit surface.
No new test merely restates the appearance assignment; native contrast will be
checked in both actual replacement windows after packaging.

Replacement binary, notice-inventory and ZIP hashes will be recorded separately.
Historical audit approval, original binary hashes and failed visual evidence stay
unchanged. Reuse all other SDK/app verification. No upload, tag movement,
Developer ID signing, notarization or expansion of qualification is authorized.

The fix checkpoint Prism review (Anthropic claude-sonnet-5-5) has complete
coverage of all four staged files and no findings. Replacement packaging/native
results will follow in a separate evidence checkpoint.

The committed fix is repackaged as build11. Final extracted native checks verify
readable contrast in B and in A after its command repaint, independent document
counters, toolbar availability and normal final-window shutdown. Post-launch
signatures, notice bindings and archive hashes pass. A separate exposure/capture
mismatch is reproduced and explicitly retained as unresolved, so public upload
recommendation for this optional fixture remains held. No original archive,
audit approval, SDK or other app is changed. See the final report for identities.

Replacement evidence Prism review (Anthropic claude-sonnet-5-5) has complete
text coverage and one medium finding f34226a04f7498c4: the pre-launch production
snapshot's not-yet-run field could be mistaken for final status. Addressed using
the review's suggested correction: the report's package record now states the
observed contrast/command pass with unresolved exposure/capture mismatch and
references replacement-native.json. The original pre-launch receipt is retained
unchanged in raw evidence and documented as a capture-time snapshot. The durable
local delivery manifest already has the completed native results. No sealed
app/archive byte or audit record was changed by this documentation correction.
