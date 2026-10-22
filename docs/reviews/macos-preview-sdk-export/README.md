# Preview SDK export blocker

The R1 source export failed on an existing specs link to excluded tracked
`tools/macos-calibration/stagingreuse/TIMING.md`. The exporter now treats tools
as repository-only documentation references using its existing existence,
tracked-source and symlink safeguards. Tools remain outside the SDK payload.

All 43 exporter tests pass, including the new real-reference/missing/untracked
regression. Prism using Anthropic Claude Sonnet 5.5 covered all three changed
files with zero retained/discarded/suppressed findings and no coverage omissions.
No runtime/API changes. The failed export log remains in local session records;
full export and SDK relocation are the next R2 checks.
