# Exact-payload macOS preview audit decision

23 October 2026. The repository owner answered the concrete decision-packet
question: “The distribution audit is approved for this payload.”
[The approval record](../../../release/macos-preview-audit.json) binds the immutable
decision-input JSON, selected native archive, app source executable and corrected
notice inventory by SHA-256, with 418 notice files and 175 component entries.
These identities were checked again before recording approval.

`release/version.json` now records the existing release-level audit as approved.
No packaging guard is removed or weakened. Original component-notice collections
retain their historical pending/review-required labels; they do not self-approve.
This records the owner's decision, not a new upstream license grant or a claim
that the pinned upstream caveat changed. Future changed payloads need a new audit.

The owner also authorized pushing the reviewed branch and opening/updating a
public draft PR for CI. This is source/report publication only. Release tags,
downloadable assets, signing identities and notarization remain separate. Current
app/SDK behavior and the completed soak are unchanged. The next step is producing
public candidates locally through the existing guards and checking those bytes.
