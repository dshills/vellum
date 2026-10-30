# Developer handoff — internal review

Add one current Mac document-workflow handoff and link it from the repository,
documentation index and developer/private-build guides. Record previous green
main separately from local maintenance commits, old build22 artifacts separately
from the new app lifecycle fixes, and offscreen installation separately from
native runtime acceptance. No new source build, hosted CI, merge or push is
claimed. The matched new-source build procedure is ready without opening apps.

Review found stale consumer statements that restart always returns to the local
notebook, recovery always saves to the local default, and current Notes reserves
a 52 Dp unified-chrome header. Correct these against ConfigureRestoration,
restorePath and the current ChromeStandard/TitlebarToolbars host wiring. Also
remove the blanket claim that unfinished lifecycle handoffs were superseded:
some native checks are genuinely deferred, and earlier reports do not qualify
later source. The top-level private recipe now names all five apps including
Notes. Historical reports and original payload identities remain unchanged.

The deferred-check plan preserves the partially used A/B fixture, warns against
replaying a fresh six-note script on seven-note windows, and requires inspection
of app/source/data identity when desktop access is newly authorized. It preserves
unsaved B and saved A, treats temporary fixtures as potentially expired and makes
no current-liveness or final-relaunch claim. Limits retain literal current-note
Find, fixed-window restoration, one-file Finder requests, plain text, no cloud,
and unqualified human IME/VoiceOver/other hosts/quarantine launch.

Validation: 141 local Markdown link targets across eight current guides exist.
All sh/bash code blocks in seven affected guides passed sh -n without execution.
Cross-checked native host flags, restoration code, committed phase hashes, the
prior exact-main CI receipt and build22 SDK/source identity. Diff checks passed.
Reused the preceding focused tests and real offscreen existing-bundle installation;
no runtime change in this phase justifies another broad test/build campaign.
Features, API baseline, desktop deferral and release/backend policies unchanged.
