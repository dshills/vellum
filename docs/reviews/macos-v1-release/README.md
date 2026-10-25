
## Separate cold native test compilation from execution deadlines

Prism: `mac-v1-release-ci-complete.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage with no final findings. Actual four-package text regressions pass; subprocess cleanup tests pass. Preserve exact test selection/race/offscreen/count and90s execution timeout, separate300s per-package compilation, package cwd and Windows suffix. Required missing binaries explicitly refuse instead of silently skipping. Original workflow40231586819 failed at cold compilation under120s; remote success remains pending a source-bound rerun.

Focused checks: `mac-v1-release-ci-complete-checks.json`, `mac-v1-release-ci-reviewed-checks.json`.

## Adopt the scoped stable v1 API and Mac support contract

Prism: `mac-v1-stable-api-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage. Historical personal_use metadata is now explicit and versioned; current macos_v1 and api_compatibility are distinct. No consumers constrain the descriptive darwin support label; exporter/checker and public-guard tests pass. Final finding adjudicated:1.0.0 identifies the intended committed payload so that exact source can undergo CI and packaging. This commit does not claim those pending gates passed or create tags/assets; final handoff remains blocked until success. Stable source diff changes only Version/comments, no signatures/runtime features. Human/host/experimental and public audit limits retained.

Focused checks: `mac-v1-stable-contract-checks.json`, `mac-v1-stable-metadata-final-checks.json`.

## Record verified private macOS v1 delivery and CI closeout

Prism: `mac-v1-release-delivery-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage with no final findings. Chronology clarified: packages and launch were verified while CI ran; both required workflows subsequently passed exact packaged source382e0e4 before final handoff. Workflow update timestamps labeled accurately. Three apps,29 verification commands, nine extracted-SDK origins, strict signed ZIP/binary hashes and unchanged notebook verified. Notes v1.0.0/build9 launched and was left running. RC physical evidence reused explicitly; no fresh Save/recovery campaign or human/platform/public-distribution qualification claimed. This report-only commit does not relabel built source.

Focused checks: `mac-v1-stable-contract-checks.json`, `mac-v1-release-ci-complete-checks.json`.
