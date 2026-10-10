# Visual text geometry review

10 October 2026. Canonical repository: Vellum, branch `codex/foundation-completion`, parent `1b20cffe5f1d0e3d0bf0e45fa2ee84a86470aa86`.

Prism reviewed staged source changes with `--no-auto-exclude --format json --fail-on medium`. Final coverage reports are complete, with no exclusions, truncation or skipped files. The following scopes have no unresolved findings:

- Toolkit: `vellum-visual-prism-source-second.json`, all 14 decoded patch files, zero findings. A temporary local source snapshot contains the exact pre/post-0011 sources; this avoids interpreting a patch inside a Git diff. Its staged diff and all decoded file hashes are recorded in `scope-hashes.json` and checked against research and canonical preparation.
- Native bridge, fixture and lifetime checks: nine files in `vellum-visual-prism-code-fourth.json` had no findings and have not changed since that review. Its three findings concern the toolkit patch and are superseded by the final decoded-source review.
- Audit, preparation, verifier and input hashes: `vellum-visual-prism-audit-second.json`, four files, zero findings.
- Documentation and evidence claims: `vellum-visual-prism-docs-final.json`, eight files, zero findings.

Earlier reports are retained with each finding's disposition in `dispositions.json`. Two findings did not require their proposed code changes: invalid patch counts were disproved by mechanical hunk counting, dry-run application and two successful preparations; split graphemes were already merged at the layout aggregation, and new regressions establish complete advance and outer caret stops. All other findings were addressed and retested.

The final full offscreen core package and focused field/shaper/raster tests pass. Native geometry matches 15 visible rendered graphemes in both final builds. Bridge/transport/semantics race tests and vet pass; 1,000 Cocoa window lifetimes drain 11,000 updates. Protocol saturation retains refusal atomicity. Independent evidence checks and schema-2 negative verifier cases pass. Actual raster screenshots were manually inspected; raw logs, benchmark samples and checked artifact hashes are committed under `prototypes/foundationintegration/evidence/20261010-visual-text/`.

`git diff --cached --check` passes outside the generated unified patch. Blank context lines inside that patch are required unified-diff formatting; patch validity is checked separately. Global foundation G1 acceptance remains open, including the multiline code editor's native adapter, live IME/VoiceOver workflows, other platforms and broader resource/release acceptance.
