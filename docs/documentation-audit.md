# Documentation audit and remaining gaps

Audit date: 31 October 2026. Base source: v1.1.3,
`0b11d3b092455b9967bad891fa42ae4ba5b6f8ae`. Read
[current status](status.md) for source/artifact identities and macOS closeout.
This audit changes documentation, not qualification results or native payloads.

## Coverage and corrections

The base inventory has 936 tracked Markdown documents: 167 maintained guides,
contracts and example/tool READMEs; 677 historical/evidence/decision documents;
76 pinned native documents; and 16 third-party notice documents. Another 1,104
tracked .txt files are predominantly logs, license material, fixtures and source
templates, not current narrative guides. Their original bytes were preserved.
The full [inventory and link findings](reports/documentation-audit-20261031/inventory-and-links.json)
record that classification and its limits.

All 936 Markdown files received inventory/local-file-link scanning. Maintained
status, onboarding, release/API contracts, operational commands and navigation
received focused semantic review against source and receipts. This is not a
line-by-line revalidation of every historical report, an export-by-export Go API
documentation audit, or a new native acceptance campaign. External URLs were not
exhaustively network-tested; GitHub CI was queried for the exact released commit.

Corrections include:

- Central status separates v1.1.3 source, the stable-v1 baseline, private build25,
  demo build5, Work Log and public build13 downloads.
- Stable API/release docs no longer identify current source as v1.1.1 or describe
  the current tutorial as alpha-only. Historical release notes remain unchanged.
- App onboarding starts with the actual generator, including the complete document
  starter. Older private Notes archive commands are explicitly historical.
- Candidate/handoff docs distinguish producer-time deferred labels from later
  scoped native checks. Generated identity/hash blocks remain unchanged.
- Dated completion, preview and performance plans link to current status instead
  of presenting old source snapshots as today's work queue. No old failure or
  owner decision was relabeled.
- Example/navigation indexes include Documents, Mac Shell, Forge, Measure Pad,
  Work Log and demo-suite restoration; current upgrade/recovery guidance is added.
- Dependency triage now has a discoverable index without editing notice inputs.

## Remaining documentation debt

| Gap | Impact and next bounded action |
| --- | --- |
| Archived relative links | The base scan found 358 unresolved targets in 25 archival files, chiefly copied source/review snapshots plus an expired absolute evidence path. They resolve relative to the copied snapshot rather than the original checkout. Preserve captured bytes; add a non-mutating original-root mapping or archive navigation index if these are needed for audit. These are not 358 broken links in current guides. |
| Release/status drift prevention | Several current identities had been copied into prose without an automated cross-check. Status is now centralized, but CI does not yet validate every current-version mention or all maintained Markdown links/anchors. Add a focused documentation check, scoped to maintained docs and explicit archive exceptions. |
| Consumer migration coverage | The new upgrade guide covers safe backups, isolated first run, rollback and existing refusal paths. Supported file-schema migrations remain the named fixture contracts; an application-version/file-format compatibility matrix and tested migration recipes are still needed before promising upgrades across arbitrary app versions. |
| API examples/reference coverage | Guides cover the maintained public workflows, but there is no complete compiled-snippet index for every stable API. A follow-up should map stable exports to Go documentation/examples and compile the supported snippets, rather than invent examples for experimental APIs. |
| Diagnostic and chronological document size | Large implementation/performance ledgers retain many historical “current” statements. Entry-point scope notes now prevent treating them as current release policy. A future index can group checkpoints by date/source; relocating or rewriting immutable evidence is unnecessary. |
| Demo product copy | The running hub still uses “Experimental alpha” in its subtitle while the root API has a stable-v1 scope. This is application UI copy, not a source-version or documentation change made here. Clarify the subtitle separately without removing experimental tool exclusions. |

## Qualification gaps that prose cannot close

- **CI:** the audit snapshot shows API/platform workflows successful and Framework
  still running on v1.1.3. Consult the [snapshot](reports/documentation-audit-20261031/ci-snapshot.json)
  and live exact-commit results; do not call the tag entirely green from local tests.
- **Multi-day use:** the [operator ledger](daily-use/SESSION-LOG.md) is empty. Three
  short automated sessions and a physical focus check do not establish multi-day
  ordinary use. The validator correctly leaves the gate pending.
- **Final matching delivery:** existing private/public SDKs and demos are verified
  at their own sources, not rebuilt v1.1.3 native releases. Selecting a green final
  source, packaging and testing those exact bytes is engineering/release work.
- **Human and distribution scope:** IME, VoiceOver, broader hardware/OS/display
  coverage, downloaded-quarantine launch and notarization lack the named results.
  Either retain explicit exclusions or perform the relevant separately scoped work.
- **Security applicability:** the low alert remains open. The current
  [artifact-bound triage](reviews/application-readiness/item-5-advisory.json) is not
  a general finding for future native builds or upstream development tools.

## Validation and review boundary

The current-guide local-target, anchor-candidate and literal repository-script
reference scans found no unresolved items after the edits. Full base-inventory
exceptions are retained rather than hidden. Candidate index/generated-guide
consistency passed. All 43 SDK/release tests and three generator tests passed. The documentation
overlay was also applied to all 33 changed/new SDK-facing guides; each passed the
archive's link/evidence-exclusion policy. An initial test attempt hit a restricted
default Go cache; rerunning with the writable cache passed. The first discovery
command selected zero tests and is not counted as verification. See the
[validation receipt](reports/documentation-audit-20261031/validation.json).
No desktop actions, new measurements, public release, native-pin change or external
review transmission is part of this documentation audit. Review is internal;
historical Prism requests and receipts do not authorize a new external review.
