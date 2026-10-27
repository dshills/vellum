# Current-payload public packaging checkpoints

The owner explicitly approves the current distribution-audit payload on
27 October 2026 and authorizes final local public packaging, launch checks,
Prism/Anthropic Claude Sonnet 5.5 review and commits. The new approval record
binds the unchanged db5d80b packet, product baseline4b42fb8 and exact pre-sign
binaries, notices/icons and native archive. Historical alpha.3 approval is
preserved. Existing producer guards remain unchanged.

This checkpoint records the owner decision and updates current release metadata.
No runtime/framework code, native dependency, asset or original notice text is
changed. Final downloads are not authorized for upload by this request. Only
existing ad-hoc signing is used; no signing identity/notarization is requested.

Approval checkpoint Prism review: complete coverage, one medium metadata-status
finding. Addressed by retaining macos_v1.status=private-release for the existing
delivery; no new status enum is introduced. Local public producer eligibility
continues to use the existing approved audit, clean-source/module/license and
notice guards. public_publication_authorized stays false. The focused version
regression and exact packet/binary/notice binding checks pass.

Final public packaging checkpoint: guarded SDK and six app production succeeds;
the SDK independent consumer passes; five app render/close observations pass.
Mac Shell is held for dark contrast and excluded from recommended hashes.
All seven archive hashes, six strict signatures, signed executables and exact
original notice bindings pass after closure; candidate processes are absent.
No runtime code or sealed package byte was changed to obtain those results.

[Final Prism review](final-prism.json): complete text coverage (15 files;
the evidence tar is opaque binary with a separately hashed member index).
One medium finding asks to verify Forma checksum consistency, while itself
observing identical hashes. No actual digest mismatch exists. The added
packet checker confirms production/README/post-launch/all/recommended checksum
agreement and actual archive bytes. Negative controls reject an altered README
hash, inclusion of held Shell and an altered evidence hash. See
[the disposition/check receipt](packet-check-receipt.json). No unchanged product
qualification was repeated and no upload was performed.

The [checker follow-up review](packet-checker-prism.json) has complete coverage
and two medium findings. The README digest check used a weak substring match;
it now parses each named app row, notice count, disposition and digest. A new
swapped-Forma/Luma-row negative control rejects misattribution. The missing-file
finding identified a diagnostic improvement, not acceptance of invalid bytes:
missing assets already failed closed; they now produce an explicit missing-file
error without a traceback. Both fixes pass focused controls, including absent
asset files. Final review is limited to these concrete checker corrections.

[Final focused checker review](checker-corrected-prism.json) covers the one
corrected helper completely. Its two medium diagnostic findings were addressed
using the review's own suggested fixes: an explicit malformed-checksum error,
a defensive non-None evidence stream check and bounded member read. The former
now has a negative control; the latter supplements the existing regular-member
and size/hash checks. Neither finding exposed acceptance of invalid artifacts.
The final check receipt passes actual seven-file integrity and six negative
controls. All findings are dispositioned; no further unchanged broad review or
product test campaign is required. Artifact bytes and source tag remain unchanged.
