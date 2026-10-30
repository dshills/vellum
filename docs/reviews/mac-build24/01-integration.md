# Checkpoint 1: integrate reviewed maintenance for private build24

Internal review, 30 October 2026. Integration of the five reviewed commits
42873dc8, faa14efe, e589f36e, 2888cb31 and 44fbf1df. Features, backend,
release/performance policies and desktop deferral are unchanged.

Compared main fa85ebe with the maintenance branch. Main and origin/main agree;
the integration is a fast-forward with no conflicts or unrelated source changes.
The only application runtime change is the exact-window lifetime guard in the
Notes action boundary. Other changes are regression/tooling/documentation work.
All individual internal reviews remain in mac-offscreen-maintenance24.

Reused unchanged focused verification: installer failures/kill/retry; read-only
verifier synthetic tests and actual build23 verification; full Notes race/vet,
generated document overlap/Quit tests and vet; disposition/notice guard; SDK
exporter documentation tests and stage dispatch. Whitespace and ancestry checks
passed. No reason to repeat a local broad campaign before existing hosted CI.

Reviewed source integration/push only. No release tag/asset, audit change, native
library update, app launch or desktop access is authorized by this checkpoint.
After pushing this checkpoint, check all three workflows on its exact commit;
retain results outside the source tree. Packaging must wait for that commit to
be green. CI is pending at commit creation, not preclaimed successful here.
