# Reviewed API baselines

`v1/manifest.json` and `v1/root.export` record the reviewed stable root API from
9f05a37 (same API as the private382e0e4 release). The export hash and explicit
package set are checked by the separate [developer guard](../../tools/apicompat/README.md).
A baseline is not regenerated during checks. Stable declarations require the
[API compatibility policy](../../specs/stable-api.md); native facade and behavior
remain separately tested. This is a compatibility snapshot, not source or binary
authentication, runtime qualification, or distribution approval.
