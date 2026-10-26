# Existing widget interaction checkpoint

Source e261cc616a92587ed0b96c3149a81a84d5c8c096, private baseline gallery build11,
26 October 2026. Ad-hoc signed; no accepted release archive replaced. Fixtures:
two-page local PDF, sixty-second silent WAV, and reference.txt in an owned temp
folder. Recreate with `python3 scripts/content-fixtures.py /private/tmp/new-folder`.

Evidence types are deliberately separate. Computer-use accessibility actions
selected pages, set the PDF Find field and accepted native file choosers. The
owner-authorized native mouse helper supplied actual window-local mouse events;
it checked the exact app bundle/window and foreground, refusing when Chrome took
focus. No user documents were opened. No clipboard contents were read or changed.

## Observed native results

- Physical PDF Next click changed page1/2 to2/2. Find "page one" selected that
  phrase and returned to page1/2. Zoom changed the native scrollbar layout and
  magnification; Fit and thumbnail toggle responded.
- Physical media Play removed the poster. Dragging its timeline sought to
  35.9615 seconds; volume changed from1.0 to0.4023. A second native Play click
  showed play/pause=on and elapsed time43.8462 seconds. Audio was synthetic silence.
  Leaving the page removed the native player controls.
- Physical Browse opened the exact-parent file chooser. Accepting reference.txt
  updated the field and enabled Reveal/Clear. Reveal selected reference.txt in
  Finder's fixture folder, with no file modification.
- Two physical calendar clicks selected October12–17. Both fields updated and
  the accepted-action count increased exactly once, from1 to2.

## Reproduced defects, assigned to checkpoint2

1. Empty-source PDF title wraps into a very narrow column.
2. Managed placeholder text shows through the native PDF toolbar.
3. PDF Find collapses to a two-point native field width in the ordinary gallery.
4. Found text can lie above the clipped viewport after Zoom/Fit; selection is
   present in native AX but its pixels are outside the visible content region.
5. PDF/media successful opens leave a previous foreground-refusal error in status.

The background AX-only initial chooser attempt was refused by the native active
window guard. Activating the exact window with the mouse helper made the chooser
work. This is distinct from an installed service being missing.

PDF Copy is pending explicit approval for a bounded in-memory original-clipboard
save/restore check; automatic approval review blocked executing that mechanism.
Finder-to-field native dropping, keyboard focus traversal, narrow/Dark layout and
corrected PDF result visibility will be completed during polish/package checks.
Human VoiceOver/IME and other hardware remain unqualified. This checkpoint records
initial results and defects; it does not certify complete native usability.
