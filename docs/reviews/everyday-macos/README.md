# Everyday Mac use — 2026-10-29

Feature-frozen maintenance, internal reviews and a separate commit per phase.

| Phase | Scope | Result |
| --- | --- | --- |
| [1](01-installation.md) | Durable versioned private install | Installer tests and real repeated build20 installation passed |
| [2](02-lifecycle.md) | Quit, Cancel, surviving windows and shutdown | Focused race and native installed checks passed |
| [3](03-files.md) | Native and automated external file workflows | Save As/later Save/private restart/explicit Open passed |
| [4](04-presentation.md) | Existing Mac look, focus and layout | Geometry/focus passed; Documents Light chrome aligned |
| [5](05-acceptance.md) | Notes/Forge acceptance and private build21 | Focused app regressions and native Forge smoke passed |

Use the [installation guide](../../private-macos-installation.md). Final SDK/app
hashes and installed launch observations belong to each private delivery's
manifest, checksums, verification and installation receipt. Existing archive
identities remain unchanged. Current source commits have not been pushed by this
task; prior green CI applies only to its stated commit.

Quit can close a clean window before another dirty window cancels. Last-window
close ends Run; relaunch opens the usual startup set. External notebooks require
explicit reopen after restart. Notes/Forge intentionally use Light presentation.
Recovery is the latest completed snapshot and Undo history is session-local.
No new public distribution, notarization or signing identity is implied. Human
accessibility, other hosts and downloaded quarantine launches remain unqualified.
