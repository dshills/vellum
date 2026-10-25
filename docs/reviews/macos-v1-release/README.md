
## Separate cold native test compilation from execution deadlines

Prism: `mac-v1-release-ci-complete.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage with no final findings. Actual four-package text regressions pass; subprocess cleanup tests pass. Preserve exact test selection/race/offscreen/count and90s execution timeout, separate300s per-package compilation, package cwd and Windows suffix. Required missing binaries explicitly refuse instead of silently skipping. Original workflow40231586819 failed at cold compilation under120s; remote success remains pending a source-bound rerun.

Focused checks: `mac-v1-release-ci-complete-checks.json`, `mac-v1-release-ci-reviewed-checks.json`.
