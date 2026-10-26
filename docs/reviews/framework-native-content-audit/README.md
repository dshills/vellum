# Framework CI native-content audit repair

Framework run 40373420007 at 0c97300 failed scripts/check.sh:
https://github.com/dshills/vellum/actions/runs/40373420007
`TestAccessibilityPendingWireFieldAudit`: unreviewed wire field NativeContent.
The exact audit failure reproduced locally before the edit (receipt retained).

NativeContent already participates in wire equality by descriptor value and is
copied into detached snapshots. Its wire schema/semantic serializer was updated
with the widget, but the separate pending-publication audit list was omitted.

The repair adds the field to the explicit audit and supports only its reviewed
NativeContentInput struct-pointer type. Unknown fields/types still fail. Added
regressions prove unchanged detached native siblings permit pending grid updates,
while every exported descriptor field, addition or removal requires full
publication. This changes test coverage only, with no runtime, API, workflow,
backend or release-gate change.

Focused race-enabled Pending/BuiltinSemantic/Snapshot/SharedBaseline/NativeContent
tests and offscreen driver vet pass. The first new fixture used nil actions;
clone's established nil-to-empty normalization made that fixture unequal, so it
was corrected to an explicit empty action slice. This was a test fixture issue,
not a runtime behavior change. Existing nil-versus-empty guards are retained.

Prism review is recorded alongside this checkpoint. Full GitHub status must be
reported from the new run after push; focused checks do not claim a green remote
workflow before that run completes. No desktop input or app data was accessed.

Prism / Anthropic Claude Sonnet 5.5: complete coverage, no findings.
