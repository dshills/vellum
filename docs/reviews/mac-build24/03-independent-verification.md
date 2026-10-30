# Checkpoint 3: independent existing-delivery verification

Internal review, 30 October 2026. Ran the committed read-only verifier against
the durable completed build24 copy and explicit installed version, with caller
manifest pin 516456cc39bc0b3c2b8faf14871cc820afda813a0a5f269e23da15fd97c77727.
It did not consume the producer's intermediate extraction or rebuild anything.

All checks passed: SDK all-file/source integrity; six named archive hashes and
SHA256SUMS; manifest/preflight/delivery/installation links; installed byte equality;
app plist/resource manifest identity; five freshly extracted and five installed
strict signatures. Verified source b9212489afc921171fc5abbd5275d8919ae930dd,
build 24, SDK 5a825eb07c480f2df248435b7c0a55e0b0cb5104ccd2de3b8f146c55c2798382.

Reviewed report against packaging identities and source-bound receipts. No
inconsistency or correction needed. Original report is retained in private
artifacts/delivery; the repository summary normalizes the installation home path.
The verifier's unit/fault coverage is unchanged and green in the source's CI;
reused it instead of rerunning an unrelated campaign.

No compiler, installer, app process, notebook reader, input helper, desktop or
LaunchServices call. Temporary snapshots are cleaned. Result is local integrity
and signature evidence, not independent publisher authenticity, native runtime
acceptance, quarantine assessment, notarization or a distribution decision.
Next checkpoint inspects linkage statically; no native execution is needed.
