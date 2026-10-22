# Packaged alpha.3 application acceptance

The owner made the desktop available on 22 October 2026. Checks used the actual
app extracted from the retained alpha.3 private zip, not a source-run substitute.
The signature and signed binary hash still match after testing. Framework features,
default backend, rendering/pacing policy and runtime code remain unchanged.

Completed observations:

- One native Connector drag creates connector-2 at (-220,-60), object count two,
  Undo two. Escape and rejected Note drops leave both counts unchanged. Empty
  drag pans. CUA observed the dedicated toolbar change to 94% after wheel zoom,
  corroborated by saved zoom 0.9417645335842487, without a document edit. The helper
  reads only the stale last-action sentence; its old 100% text is not the toolbar.
  Undo removes connector-2; Redo restores that exact identity.
- Select All, Copy and Paste create block-3/connector-4 together in one Undo step.
  Undo removes both; Redo restores those IDs. Duplicate creates block-5/connector-6
  together. Delete removes the pair in one step; Undo restores them. Undoing paste
  and duplicate restores the original pair. Relative-geometry edge cases retain
  the already passing unchanged-source automated coverage; this native check
  observed identities/count/history, not a fresh numeric paste-coordinate audit.
- Layers Forward changes paint order through one undoable command; Undo restores
  it. Group makes group-1. Direct corner resize changes both members together,
  then native dragging moves the group; each adds one history step. Shift-marquee
  selects both members without a document edit. Ungroup works; Undo restores
  group-1 and membership.
- Save As writes a new temporary drawing. New clears it; native Open restores
  exact identities/groups, dimensions and camera, with history reset to zero.
  The saved JSON is retained. Native Save As cancellation preserves the drawing;
  opening an invalid temporary file reports failure and preserves the current
  clean document. Disk-write failure retains existing passing automated coverage
  (TestCanceledSaveAndFailedWriteKeepDirtyDrawing and related close/save tests),
  rather than a new OS-permission failure claim. The earlier same-product SVG/PNG
  content check remains separately scoped R2/R3 evidence and was not rerun.
- An unsaved third component, block-7, triggers close protection; Cancel keeps it.
  After verifying the private recovery copy contains only these three disposable
  test objects and the original saved pair, one exact owned app process is killed
  deliberately. This is a controlled interruption, not a discovered crash. On
  relaunch, Restore offers an unsaved drawing with all three IDs, group geometry
  and 94% zoom. Saving it produces JSON exactly equal to the pre-termination copy.
  Ordinary Quit then leaves no exact app process and removes the recovery copy.

A path initially entered in Save As's filename field became a colon-separated
leaf in the prior owned temporary directory. This was test entry error; using
native Go to Folder plus a filename selected the intended destination. The
hash-matching misnamed disposable file was removed; no user drawing was replaced.
An earlier focus-lost drag remains an inconclusive separate attempt. The renewed
quiet drag succeeds: its initial snapshot is background, then the helper activates
and guards foreground before every posted event and observes foreground afterward.
The initial false value precedes activation; it is not evidence of input sent to a
background app. Pointer evidence is synthetic native
mouse input, not human hardware verification; AX/keyboard actions are identified
separately. Global input/accessibility settings were not changed.

Packaging notice discovery reuses existing bounded inventory functions against
the actual Placement graph: 470 packages, 4,441 nonstandard embeds in 35 modules,
and 42 candidate ancestor declarations. It found two missing conservative notice
candidates: the Go BoringSSL-related declaration and the additional Material Symbols
NOTICE. The latter attribution is present in the SDK source but absent from the
initial app notice bundle. A separate final bundle now has 418 files/175 entries;
its membership and binary binding validate. Final packaging must use that bundle.
This does not establish that optional BoringSSL code is linked, settle asset/legal
applicability, approve rights or silently change the already tested zip.

The TUINSWindow base class is implemented by Apple's TextInputUIMacHelper image,
as established through public runtime metadata in a separate no-window process.
This establishes class origin, not the retaining reference or deallocation. The
one-window terminal count and other resource limitations remain separately scoped.

Nonblocking backlog: after a pan/zoom and canceled gesture, the last-action status
sentence may retain the old zoom while the dedicated toolbar value is current.
Invalid-file messages expose parser wording. These did not block the listed
workflows; no frozen-scope UI polish was added. Native active-editing/idle/soak
acceptance, final packaging/CI and distribution decisions remain open here.
