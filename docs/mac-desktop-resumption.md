# Safe resumption of deferred Mac checks

Status on 30 October 2026: **desktop checks remain deferred**. This checklist is
a handoff, not permission to execute its native steps. Keep features, backend,
rendering/pacing and release/performance policy unchanged. Do not change global
input, accessibility, security, quarantine or application document settings.

## Offscreen identity checks available now

The [current handoff](mac-document-workflows-handoff.md) distinguishes green main,
local maintenance and installed package source. The final installed build23 uses
source `e0e011c605f8ee6ea84827414a1636bcca74c97e`; its SDK SHA256 is
`a9857bff74590959d688e2827af2918d87468b1667773b6a278123f8fb4cbc75`.
From the current checkout, verify that completed private payload without launch:

```sh
python3 -B scripts/verify-private-macos.py \
  --bundle "$HOME/Documents/ChatGPT/Go Native UI/deliveries/mac-build23-20261030/final-build23-e0e011c" \
  --installation "$HOME/Applications/Vellum/build23-e0e011c605f8" \
  --manifest-sha256 f0d3a2ac660fd68e6bbace8adc19096dc07d056af8753ded8fc9d4da794da6d4
```

The command uses caller-selected paths, temporary archive snapshots and codesign;
it does not execute the apps or access notebook contents. Do not modify those
inputs concurrently. Paths are the recorded owner's local delivery convention;
on another host select the explicitly transferred complete delivery/installation
and their reviewed identities. Do not substitute a similarly named old build.

The newer app-local fix at `e589f36e` is source-only. Installed build23 cannot
validate it. When a newer candidate is requested, package committed source through
the existing matched SDK/five-app producer with a fresh output and build version,
then record its receipts and exact installation. Do not overwrite build23 or
relabel old archives. A new package requires its own physical results; passing
source tests or old signatures cannot supply them. No rebuild is required simply
to keep this checklist current.

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
