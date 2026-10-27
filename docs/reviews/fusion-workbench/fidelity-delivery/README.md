# Forge Fusion fidelity delivery — private build6

The owner-approved presentation pass is complete on the existing Fusion-inspired
example. Features, commands, renderer, stable API and release policy stay frozen.

- Ribbon: shorter flat full-color actions, tighter group captions and separators.
  Whole-group overflow remains. The final ribbon measures 113 points high on the
 1458×972point window, versus 146 in the first build4 visual check. The native
  content-tab minimum correction also retains intrinsic tool height; ordinary
  content tabs retain their existing default minimum and scroll/grow behavior.
- Browser: one heading, compact 29-point rows, aligned original artwork/text,
  quiet independent eye/activation/disclosure targets and a restrained selection.
  The duplicate project heading and outer scroll container are removed.
- Timeline: one 37-point icon strip, blue marker and existing previous/next plus
  earlier/later commands. Names and paused state remain in tooltips/accessibility;
  icon-less consumers retain text fallback. Paused selected items retain gray
  fill plus a blue selection outline.

## Verification

Focused checks and Prism Anthropic Claude Sonnet 5.5 receipts are preserved in
[phase10](../phase10-fidelity/README.md), [phase11](../phase11-ribbon-height/README.md)
and [phase12](../phase12-ribbon-fit/README.md). These cover UI contracts/races,
160/240/320 DPI native geometry/routing and normal tab/settings regressions,
Forge command-path Undo/Redo, vet and the unchanged stable API baseline.
The minimum-only intermediate build5 was rejected by native checking and closed;
the strengthened test now requires nonzero height and visible contained actions.

Physical checks on the exact ZIP extraction used the already-approved Forge-only
mouse helper: ribbon Add and Undo/Redo; independent browser selection, eye,
activation and disclosure; feature selection, marker, reorder and Undo/Redo.
AX records confirmed actual state and feature order. A native Tab reached the
named earlier button. This is separate from synthetic/offscreen routing evidence.

The first combined sequence later showed unrequested changes between helper
steps, and the window closed. Attempts 27–32 are excluded from controlled evidence;
no cause is attributed. The owner then granted a fresh three-minute quiet desktop
window. Attempts 33–40 passed exact marker/reorder and Undo/Redo assertions, with
one journal entry per requested action and no intervening state changes. Final
fresh launch was verified at 21:18:43 UTC, within 76 seconds of that quiet window's
start. See [native checks](native-checks.json) for bounds/status, raw-record hashes,
invalidated attempts and screenshot provenance. Human accessibility is unqualified.

The final unedited JPEG was inspected against the prior Forge capture and the
fixed Autodesk desktop reference used in the spec. It shows all ribbon groups,
compact outlined rows, selection, single icon timeline and the live 3D viewport
without clipping in these areas. This is a qualitative fidelity improvement;
Fusion versions, native controls, artwork and scale differ. The inspector,
orientation controls and modal action panel remain outside this pass.

## Package and handoff

Source: `ad207e218e4e16106067920b4d82f135baa459de`. [Manifest](manifest.json) and
[producer check](package-check.json) bind the clean committed source to the ZIP.
Both the built app and the exact extracted ZIP app passed strict deep signature
verification; the fresh [signature receipt](signature-check.json) records exact
commands, outputs and observation time. The manifest and producer evidence are
used together; the manifest is not a standalone source attestation.
Archive SHA256: `5f4ad911d19a0dfca2e889bc801629b95fe8a725dbb25659977e9f5669b08815`.

Durable private delivery: `/Users/dshills/Documents/ChatGPT/Go Native UI/deliveries/forge-20261027-build6`.
Running app: `/private/tmp/vellum-forge-delivery-20261027-build6/extracted/Vellum Forge.app`.
It was relaunched with the original seven-body in-memory study, Undo 0 / Redo 0;
PID 6534 is an observation, not a permanent process identity. The final_relaunch
object in native-checks.json links raw record42 and its SHA256 to that new process,
source bundle, observed time and reset journal; it is separate from the prior
process 5760 used for quiet input checks.

This is an Apple Silicon private app, ad-hoc signed and unnotarized. No SDK,
release tags, public assets or audit records changed. Existing box proxies,
mock creation history, in-memory-only documents, 48-body limit and experimental
viewport remain as documented in the [Forge guide](../../../../examples/forge/README.md).
No new CAD solver, feature operations, timeline dragging or playback were added.
