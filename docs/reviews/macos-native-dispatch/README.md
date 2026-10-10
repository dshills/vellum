# Native dispatch review

Internal design review identified a terminal recovery deadlock: Cogent attempted
to open a crash window on its stopped queue. The new terminal-state guard logs and
rethrows immediately. Internal review of all nine prepared source/test files and
then the canonical patch, hashes, manifest, regression selectors, specs and native
fixture found no actionable defects.

One significant Prism batch reviewed the completed source/specification/test/
evidence checkpoint with Anthropic Claude Sonnet 5.5. The report has no findings
at the medium/high threshold and declares complete coverage of all 23 files,
with no exclusions, skipped chunks or truncated bytes. No unchanged follow-up
Prism run was needed.

`staged-inputs.json` captures the exact index hashes, parent commit, staged diff
hash and Prism report hash before these immutable review records were added.
Those inputs were checked unchanged before review completion and again before
commit. The final commit verifies the same 23 file hashes. The
[evidence](../../reports/macos-native-dispatch/README.md) records fresh full checks,
twenty race repetitions, separate native normal/panic probes, window handoff and
workbench/GPU shutdown, with source and log identities.

The contract covers synchronous failure and desktop wake termination, not broad
GPU/AX recovery or other GLFW operations. Physical OS workflows, sustained
performance and public distribution remain outside this checkpoint. OS input and
VoiceOver preferences were unchanged; bounded native windows closed.
