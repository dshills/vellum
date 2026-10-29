# Installed coordinated Quit — internal review and handoff

The installed private build22 from `4f447e85aa97dbcfa3b36d2707a679c173791777`
passed the native coordinated Quit check on 29 October 2026. The SDK and all
five apps retain that exact source identity; this is a later evidence record,
not another runtime change or a replacement app build.

## Observations

The app-restricted native helper opened the A/B Documents example with disposable
repositories. Each window changed from six to seven notes. Native Command-Q
presented A's decision. Save in A followed by Cancel in B kept both participating
windows open with their seven-note documents. Reading A's persisted fixture
confirmed seven notes: the successful Save remained saved after the later Cancel.
This establishes the corrected native input lifetime and the intended delayed
closure behavior for this bounded case, alongside the existing race regressions.

Before the next Save B action, foreground switched to Outlook. The helper
refused input; no unrelated application received that action. The owner then
requested remaining checks stay offscreen. Desktop work stopped immediately;
the fixture windows were left open. This interruption is not an application
failure and does not qualify the remaining workflows.

## Offscreen checks and review

Rechecked the installed payload's exact manifest binding and document declarations
for Notes, Document Starter and Forge. All three declare the expected extension,
Editor role, exported UTI and Alternate handler rank. No system default was
changed. This metadata inspection is separate from actual cold/warm Finder
behavior. Reuse the unchanged focused application, generated-consumer, native
callback and packaging verification already bound to the build; no broad suite,
performance campaign or new framework feature was introduced.

Internal review checked the structured native receipt against its action order,
assertions and source identity. Only coordinated Quit is recorded as a native
pass. Remaining native restoration/active-window/geometry, Finder Open With,
Find/Replace interactions and final post-test relaunch observations are deferred
by the owner. Ordinary installed launches had passed in the prior handoff.
The corrected native run retired those prior sessions normally before testing;
this continuation does not claim all five apps were relaunched afterward.

The local private delivery carries separate native-Quit, stopped-run and
registration receipts. Public tags/assets, audit/signing/notarization, backend,
performance gate and human IME/VoiceOver qualification remain unchanged.
No Prism or other external code-review service was used. Documentation-only
checkpoint: existing source tests and package identity are reused.
