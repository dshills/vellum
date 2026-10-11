# Retained-image review

One significant Prism batch used Anthropic Claude Sonnet 5.5 on 29 staged
source/specification/test/evidence files: 115,145 bytes, six chunks and eight
calls. Coverage is complete with no exclusions, skipped chunks or truncation.
It reported one medium finding and no high findings; its medium threshold returned
exit status 1. The report is preserved unchanged in [prism.json](prism.json).

Independent internal review refuted the finding that the restore step must reset
the patched source. The fixture deliberately returns from the 2×1 halo image to
the retained edited 4096² image. It restores placement/filter/mesh binding without
claiming pristine pixels. Subsequent budget, rebuild and resize checks retain that
edited version; removal checks transparent output and closure ends the fresh,
private smoke application. No state persists into a user document or later run.
Resetting to the original would discard the version meant to survive resource
recreation. [adjudication.json](adjudication.json) records call paths and hashes.
No source change or repeated unchanged Prism run was required.

Internal implementation review resolved reversed-rectangle overflow admission,
owner-transaction chrome acknowledgment, precise texel-only snapshot budget
refusal and the copied patched-pixel mutation in the independent SDK contract.
Its final read-only review found no remaining actionable production defect.
The [validation report](../../reports/macos-images/README.md) distinguishes full
checks from two subsequent test-only corrections and their affected reruns.

`staged-inputs.json` records the 29 exact reviewed index inputs, parent, staged
diff and raw report hash before these immutable review records were added. The
commit verifies every input hash. Postcommit SDK relocation evidence remains
outside the canonical checkout so it can record the actual clean commit identity.
Native clipped compositor/MSAA edge quality, broader P4, sustained performance,
physical GPU loss and other Mac gates remain open. No OS preferences changed.
