# Safe resumption of deferred Mac checks

Status on 30 October 2026: **desktop checks remain deferred**. This checklist is
a handoff, not permission to execute its native steps. Keep features, backend,
rendering/pacing and release/performance policy unchanged. Do not change global
input, accessibility, security, quarantine or application document settings.

## Offscreen identity checks available now

The [current handoff](mac-document-workflows-handoff.md) distinguishes recorded
candidate identity from current source integration. These commands come from the
[validated candidate index](../release/private-macos-candidate.json):

<!-- vellum-candidate:commands:start -->
```sh
python3 -B scripts/verify-private-macos.py \
  --bundle "$HOME/Documents/ChatGPT/Go Native UI/deliveries/mac-build24-20261030/final-build24-b9212489" \
  --installation "$HOME/Applications/Vellum/build24-b9212489afc9" \
  --manifest-sha256 516456cc39bc0b3c2b8faf14871cc820afda813a0a5f269e23da15fd97c77727
python3 -B scripts/check-private-macos-linkage.py \
  --installation "$HOME/Applications/Vellum/build24-b9212489afc9" \
  --installation-sha256 1c8a8b8425759434214631ff73b8ec9293af9c89dd82847f111cdea544a39fcd
```
<!-- vellum-candidate:commands:end -->

Run `python3 -B scripts/private-candidate.py` to check index/report/guide consistency;
add `--verify-local` to verify the recorded local archive/installation bytes and
signatures. Neither mode starts an app or accesses notebooks. The index check is
checkout-only because historical reports are intentionally absent from SDK exports;
the standalone verifier/linkage tools remain independently consumable from SDKs.

The commands use the recorded owner's local delivery convention. Other hosts must
select their explicitly transferred complete payload and reviewed identities.
Do not modify selected inputs during verification or substitute a similarly named
old build. Build24 includes e589f36e; build23 and older installations remain retained.

All five build24 executables require macOS 27.0; recorded inspection host is 27.2.
They cannot run on historically qualified 26.6.2, and no 27.2 launch qualification
is claimed. Signatures/source tests do not establish physical/native behavior.

## Before any native check, after desktop deferral is lifted

1. Confirm the approved app set and bounded desktop window. Use only authorized
   app-restricted input methods; stop if the foreground app changes. Do not send
   fallback input to another app or silently extend the testing window.
2. Inspect the actual existing app/version and data directory. The old build22
   Documents fixture was last seen with A saved and B dirty, both seven notes.
   That is historical evidence, not a statement about today's processes or files.
   Do not force-quit, overwrite, migrate or blindly rerun the original six-note
   fixture. Preserve unsaved work and resolve any uncertain ownership first.
3. Admit only a verified selected candidate. Do not run two versions with the same
   bundle ID against the same repository. Keep previous app directories intact.
4. Use newly created private, nonsynced disposable data directories. Record their
   absolute paths, source/build/SDK/manifest identities and intended A/B files.
   Pass explicit `-data-dir` overrides for Notes, Documents and Document Starter;
   never use the personal Notes notebook as a test fixture. Use a disposable
   Forge study and its documented isolated-state options. If fixture provenance
   is uncertain or temporary helpers expired, recreate them before continuing.
5. State the expected initial counts, dirty flags, saved file bytes and selected
   document. Store evidence in a new directory; keep original native logs unchanged.
   Check app-restricted targeting before each gesture. Refusal is a stopped check,
   not an application failure or a reason to send unrestricted input.

## Bounded remaining cases

| Case | Required observation |
| --- | --- |
| Coordinated Quit | Save A then Cancel B preserves both windows and B edits; finish B's decision safely before retiring the fixture. |
| Restoration | External paths, active A/B ordering and native geometry survive a normal restart; recovery retains precedence. |
| Native opening | Cold/warm Finder delivery and native chooser focus reach the selected app; dirty Save/Discard/Cancel behaves correctly. Never change default file handlers. |
| Find/Replace | Existing shortcuts reach the editor, dismissal restores focus, and Replace All is one Undo/Redo step. |
| Independent consumers | Document Starter and Forge open their disposable documents and restore only documented state. |

Run only cases justified by the selected candidate's changed paths and outstanding
checks; reuse earlier unchanged evidence. A desktop interruption or protected
system dialog stops the affected case. Preserve fixture data and record the exact
last completed step. Do not turn a bounded check into a performance campaign.

## Evidence and finish

Record the exact tested source/package and separate physical-pointer/keyboard,
accessible-action and automated/offscreen results. A code-level window-ID reuse
test does not prove native focus; an archive signature does not prove launch.
The lifetime regression at `e589f36e` currently has offscreen evidence only.
Human IME/VoiceOver, other hardware/hosts and quarantine/notarization remain
unqualified. A successful local check grants no new distribution approval.

At the end, leave ordinary apps launched only if requested. Otherwise preserve
remaining fixture decisions and report the handoff. Update current instructions
and add new evidence without rewriting historical hashes or claiming unexecuted
steps passed.
