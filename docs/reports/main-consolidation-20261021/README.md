# Main consolidation — 2026-10-21

The canonical Vellum checkout now uses `main`. It started at `f35c5a0477b4b0d30df9ae4116cbb0efd44e0cce`, the tested Internal22 checkpoint. Every one of the 14 pre-existing local branch tips and all 14 registered worktree HEADs were already ancestors of that checkpoint; no source merge or conflict resolution was necessary. There is no configured Git remote.

Three historical worktrees retained drafts. The exact modified/untracked files, index-only blobs and applicable binary patches are preserved in `historical-worktree-drafts.tar.gz`; `manifest.json` records their SHA256 hashes and original branch/HEAD/path. Their feature work subsequently landed through `bd0e1de` (icon provenance), `ad91bc6` (desktop input measurement) and `60eaad1` plus later fixes (idle soak). These snapshots are recovery material, not inputs to the build. Current implementation, frozen framework behavior and Internal22 app remain unchanged.

Several old temporary checkouts contain directories and Git metadata but none of their tracked files. Their missing-file statuses were recorded, not committed as mass deletions. The idle-soak index still held nine changed blobs, which are included in the archive even though their working files were missing. Original branches and worktrees are retained without resetting or deleting anything.

Verification: all branch/worktree HEADs are reachable from main; archive member contents match their captured source/index bytes and hashes. Only these consolidation documents/archive are added to main. Reuse the preceding tested application checkpoint: no executable source changed. The separate research/delivery workspace has an unborn repository and no branches or commits to merge into Vellum.
