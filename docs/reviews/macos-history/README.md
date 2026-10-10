# Document history and native shutdown review

Prism reviewed the significant completed batch with Anthropic Claude Sonnet 5.5:
38 initial source/specification/test/evidence files. Its two medium findings were
resolved before committing: the synthetic queue-producer claim was refuted by
source inspection and independent internal review; the StopMain native wake
ordering issue was fixed and tested.

The nine-file corrective review is clean. A final six-file SDK integration review
is also clean; it includes the exporter whitelist fix, relocated consumer
regression and updated evidence. These focused follow-ups avoid another full
batch review. Every report declares complete selected-input coverage with no
exclusions, skipped chunks or truncation. Review-chain.json verifies that the
initial reviewed hashes, corrective hashes and SDK integration hashes compose
exactly to all 42 final staged input hashes before these immutable records.

Internal agent review was used between Prism runs. Fresh pinned preparation,
portable/native race tests and vet, selected backend regressions, native build,
independent consumers, 20 queue race repetitions and four native smokes pass.
All 26 exporter regressions pass after the packaging correction. See
../../reports/macos-history/README.md for commands, source identities, test
scope, the initial failures and qualification limits.

No OS input sources or VoiceOver preferences changed. Document persistence,
physical IME/assistive workflows, sustained performance and other Mac completion
gates remain open. Preexisting callback-panic acknowledgment and initial
RunOnMain wake-lifetime gaps are documented separately; this checkpoint does
not claim them fixed or a complete production framework.

The corrective diff is stored as base64 in corrections.diff.json to preserve its exact context whitespace. Decode its data field to verify the raw diff hash in correction-source-proof.json.
