# Mac content integration delivery

The authorized five checkpoints are on codex/mac-integration-polish:

1. 8a0f1a4: bounded original-build interaction record and disposable fixtures.
2. 1787de2: title layout, opaque native background, responsive PDF Find, identity-
   scoped search results, page-coordinate selection reveal and open/cancel status.
3. 9eb8f9e: public checklist and scheduled date-range filtering in fictional,
   in-memory Reminders, preserving input settlement and bounded Undo/Redo.
4. 9521ead: independent public history/date/async session examples and guide.
5. Private build12 packaging: SDK and Widgets/Reminders ZIPs use clean source
   9521ead5716d4968614ff8039dfa76d47c5bc149. This report/producer checkpoint does
   not relabel their bytes. Final native interaction results are in `native-final.json`.

The private delivery folder is `deliveries/mac-integration-pass` in the owner's
research workspace. Its README describes extraction, SDK consumption and limits.
`package-manifest.json` pins the SDK, app archives, signed executables, source,
original native archive and relocated module origins. `SHA256SUMS` verifies all
three distributable archives. `package-producer.py.txt` records exact steps.
Generated starter and independent content examples pass race/vet/build; relocated
Reminders passes race/offscreen/vet; extracted and installed app signatures pass.
No broad performance/qualification campaign was run.

`interactions.md` records original mouse/AX results and actual failures; it is
historical evidence, not a claim those initial pixels were correct. The wide and
340dp corrected native checks in `polished-native.json` independently verify Find
width, compact toolbar, opaque background, loading and cleanup. Offscreen and
native inspection do not certify physical input or human usability.

The exact-app native input producer is reviewable in
`native-input-producer.swift.txt`. It admits only the private build12 Widgets or
Reminders app at an exact temporary install path. Finder dragging admits only
reference.txt in the named disposable fixture folder. Input refuses other
foreground apps. The owner explicitly approved the PDF Copy check: at most4 MiB,
16 clipboard items/32 types, saved only in memory and restored; no original
clipboard value or format data is logged, saved or sent to Prism. The source code
is reviewed, not the clipboard data. No clipboard manipulation happened during the
original interaction checkpoint, where automatic approval review refused it.

Native content remains a local-source, bounded-admission integration. Media paths
are mutable; OS decoder costs are not hard-capped. Arbitrary managed popover
stacking and native pixels in offscreen captures remain unsupported. Human
VoiceOver/IME, other hardware and public distribution remain unqualified. Apps
are ad-hoc signed and unnotarized. Accepted v1/build9 artifacts, release/API
baselines, audit status, backend/pacing policy, global OS input/accessibility
settings and user notebooks are unchanged. No public tags/assets are published.

Final build12 checks passed PDF selection/Copy with clipboard restoration, Find/Zoom/Fit visibility, app-scoped Dark appearance, a physical Finder file drop, and physical Reminders completion with AX Undo/Redo. Both packaged apps are running. The Reminders range filter has automated model/routing coverage; it was not physically requalified in the final package. See `native-final.json` for failed attempts, exact evidence boundaries and remaining limitations.
