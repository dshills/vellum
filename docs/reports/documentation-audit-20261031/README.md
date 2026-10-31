# Documentation audit inventory — 31 October 2026

See [the maintained audit/gap report](../../documentation-audit.md) and
[current status](../../status.md).

- [Base document inventory and missing-target findings](inventory-and-links.json):
  classification for 936 tracked Markdown files and 358 archival link targets.
- [Validation receipt](validation.json): maintained navigation scans, SDK overlay checks and test counts.
- [Exact-source CI snapshot](ci-snapshot.json): v1.1.3 workflow state at inspection,
  not a permanent pending status or a claim of final success.

The audit scans file/link metadata across the base inventory and semantically
checks maintained entry points. It does not rewrite embedded source captures,
notices, pinned patch documentation or historical approvals and hashes. New audit
pages are not included in the base count; current-guide validation includes them.

One-off scan methods are preserved as [scan.py](scan.py) and [anchors.py](anchors.py). Run them from the repository root in that order with Python 3; they write their intermediate scan JSON under /private/tmp. They scan the checkout state at execution, not the frozen base inventory. Their regex parsing and untested external URLs are explicit limits, and they are not wired into CI.
