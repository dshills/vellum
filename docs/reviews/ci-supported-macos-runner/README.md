# CI runner/selection maintenance — 27 October 2026

Checked source c8ef1f79022a00966f9e5577a7d5e22b6f93462a after its push to main.
Stable root API run40501491633 passes all three hosts; Platform adapters
run40501491610 passes Linux X11 and Windows (optional self-hosted qualification
is intentionally skipped). Framework run40501491700 is queued without a runner
or executed step, not a failed test. No source failure is invented from that state.

GitHub's [official macOS14 notice](https://github.com/actions/runner-images/issues/13518)
announces retirement2 November2026 and warns of extended queues while deprecated.
The [supported-image table](https://github.com/actions/runner-images/blob/main/README.md)
lists macos-15 as the corresponding standard ARM64 label. This is a concrete
upcoming CI availability defect; the notice does not prove the exact cause of
this individual queued run.

Move Framework and the API matrix macOS job from pinned macos-14 to pinned
macos-15. Existing commands, platforms, permissions, timeouts and release gates
are unchanged. The CI host is not an added macOS runtime-support claim. Linux
and Windows stay on their existing pins. No runtime/SDK/app bytes are changed.

The new stage0108 exposure tests were absent from the explicit dependency-test
selector (dependency tests do not run transitively). Add only TestVellumExposure
to that existing offscreen selector; retain every prior selected test. Shell
syntax and all workflow YAML parse, and the three newly selected exposure tests
pass with race checking. No duplicate broad local campaign is started. The real
new-head workflow results will be checked after the reviewed fix is pushed.

Features, APIs, native backend/pacing, tags, audit records, packaged assets and
signing policy remain frozen. Only CI configuration/selection and this evidence
change. Review uses Prism with Anthropic Claude Sonnet5.5 before commit/push.

Prism review with Anthropic claude-sonnet-5-5 covers all six staged files
completely and reports no findings. Local checks pass; full GitHub verification
is evaluated against the resulting pushed main commit, not the predecessor.
