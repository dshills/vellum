# Toolbars, status bars and form groups

Vellum now provides `ui.Toolbar`, `ui.StatusBar` and `ui.FormGroup` as retained
containers. Toolbars use horizontal/default or vertical main-axis scrolling,
keeping native controls, private editing state and stable child identity during
overflow. The cross axis keeps natural control size subject to explicit limits.
Tab uses child managed stops; unused main-axis arrows wrap through enabled stops,
and Home/End reveal the first/last. Focused text/choice handlers retain precedence.
Nested layouts route navigation through the current nearest toolbar; reparenting
clears the old scope. Modal content and extension-owned trees keep their own input
scope. Recognized traversal with a disabled toolbar/ancestor or no enabled stop
is consumed without generic disabled-child fallback. Ordinary trees without an
active/old toolbar scope avoid native traversal and binder map allocations.

FormGroup has an escaped visible Text heading and retained column content.
StatusBar places an escaped status Text beside a row of children and exposes the
literal message as semantic Value. Stable Accessibility.Name can name that region.
Private captions do not duplicate accessible nodes or add focus stops. Empty Text
hides a caption; nonempty Text restores it. Managed Gap/Align/Justify style content
as well as the outer structure; font attributes inherit, with weight restored
after native Text defaults. The caption keeps normalized inherited size so global
font scaling applies once. The Mac bridge maps Toolbar to AXToolbar and status/form
to AXGroup, publishing ordinary managed children/actions. No live-region announcement
or automatic field-label association is implied.

See [the catalog](../../controls.md) and [contract](../../../specs/controls.md).
The independent gallery demonstrates all three using accepted application state.

## Verification and corrections

Final portable and complete offscreen adapter race suites pass. Focused chrome
and gallery tests pass three race repetitions; adapter vet and native gallery
compilation also pass. These five final-v3 receipts bind the sixteen-leaf baseline and unchanged 6,198
prepared identities. A final single test-only caption assertion change then passes
three focused race repetitions against all sixteen final-v4 source identities.
The fifteen other leaves and all production code remain unchanged; earlier full
qualification supports that same production. The final targeted run compiles the
entire affected test package, including the strengthened caption fixture. The gallery was
compiled only. Native objects verify AXToolbar and two group roles without a window.

Actual Manual fixtures verify both toolbar orientations, constrained overflow,
keyboard reveal and activation of the last tool, Home, disabled-aware wrapping
arrows, ordinary Tab order, child text Home/End precedence and completed status
pixels. They also verify nested-row traversal, nearest-inner toolbar scope and
current scope removal when a layout moves out. The moved-out Home check directly
sends a Core event to the retained frame, separately from the full Manual router.
Retained tests cover private-host reorder/reparent, uncommitted field drafts,
latest callbacks, disabled delivery, title escaping and style removal/restoration.

Internal review identified nested layouts using generic disabled navigation,
captions remaining Invisible after Text returns, and native Text resetting font
weight. All three fail in the exact preserved pre-review sixteen-leaf candidate:
nested Right focuses disabled B, returning captions have empty geometry, and
caption weight is normal versus accepted semibold. They pass after dynamic layout
fallback binding, explicit two-way visibility restoration and weight inheritance.
Real Manual empty→nonempty→empty→nonempty tests check geometry and completed pixels.

Review of the first font fix exposed a whole-Font copy rescaling inherited Size.
The corrected125% in-process fixture checks the active scalar and25dp parent size,
fails the preserved old Chrome overlay at31.25dp caption size and passes the
Weight-only fix. The initial pre-bootstrap fixture PASS is explicitly weak evidence:
body settings initialization could reset its scalar. Tests call no persistent
settings API and restore the in-process scalar. Another genuine overlay regression
shows all-disabled children, toolbar and ancestor modes focus a disabled sibling
without the final recognized-traversal guard; the unchanged final fixture passes.

34 raw receipts and 68 logs retain 10 exit-1
attempts. Switch insertion and test-only-helper compilation mistakes,
both-axis toolbar collapse, native disabled arrows, normal-handler non-bubbling
attempts, the initial review regressions, corrected scaling red and disabled-scope
reds remain historical attempts. The five final-v3 baseline checks plus final-v4 targeted race run support
this unchanged-production checkpoint with its one strengthened test leaf. Historical source maps and available payloads explicitly list any gaps;
exact pre-review sources and old overlay inputs are retained. The initial review
NEEDS_FIX snapshot remains immutable. Binaries, caches and complete compiler/SDK
closure are excluded.

## Review and remaining gates

Final internal source review approves the exact sixteen leaves after corrections.
Prism runs once at this checkpoint using Anthropic Claude Sonnet5.5: complete
sixteen-file baseline coverage, 4 chunks, 40319 input/
5227 output tokens, raw exit1 with one medium test-coverage finding.
The retention fixture previously guarded captions by nil presence, permitting
silent skips in that single fixture. Other rendering/font tests dereference the
caption and were already independent coverage. The final test now explicitly
requires a caption for status/form and absence for toolbar, and runs checks by
kind. Internal review verifies that sole test-only correction, and three targeted
race repetitions pass. The finding is resolved; its original raw provider report
is preserved. Prism was not repeated for this trivial assertion correction.
Provider baseline coverage and internal final-delta coverage remain distinct. Separate internal evidence
review binds the report, raw receipts, source families, review/provider records and
sealed archive before commit. The [checkpoint](records/checkpoint.json) binds that
review and the [archive](records/evidence.tar.xz) with its [member index](records/evidence-index.json).

No OS desktop UI/input/settings or general clipboard writes occurred. Software and
inert-object evidence does not qualify physical Mac input, human VoiceOver,
native/GPU memory, sustained performance/endurance or full framework completion.
Those original gates remain open while the public UI vocabulary expands.
