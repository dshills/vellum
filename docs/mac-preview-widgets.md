# Mac preview and editing widgets

## Quick Look

`Application.QuickLook()` is optional. On a worker call `PreviewFiles(ctx, window,
QuickLookOptions{Items, SelectedID})`. Items contain stable ID, absolute local Path
and Title, 1–16 regular files. The service checks paths outside the UI owner; the
OS reads previews. Keep files available until the call completes. Escape/Space or
close returns nil. Caller/window/application cancellation revokes the session.
Unavailable platforms return `ErrUnavailable`. Foreign preexisting Quick Look
panels refuse rather than taking over another responder's state. One native
auxiliary session is active at a time; Colors/Fonts and Quick Look don't overlap.
The parent must be visible in the active application and be its main or key
window. Background requests return `ErrUnavailable`; the service does not
activate an unrelated window. AppKit owns preview geometry and animation.
The actual AppKit panel provides its system navigation and supported formats.
Full-screen preview and external preview-provider behavior are OS-owned; no
cross-window handoff, editor, file chooser or file-writing API is added.

Set `DocumentList.Preview` or `OutlineView.Preview` for a Space-key proposal.
Selection and Return activation retain their existing behavior. Call the service
on a worker, using a reserved window completion for result publication, as the
gallery does. Quick Look page provides generated text and original-color PNG
samples through both a document list and outline; temporary files are removed
when the session completes. Native/physical verification is separate from
request, stale-event, parser and lifetime regression tests.

## Find / Replace

`ui.FindBar` retains controlled Query/Replacement/Open/options and takes an existing
`editor.Model` for counts. Change edits one field; Request supplies `FindRequest`.
Call `request.Apply(model)` on the owner, or handle the request in your search
engine. FindNext/Previous wrap; FindReplace replaces a selected match or selects
the next; FindReplaceAll is one undoable journal edit and maps the existing selection through replacements. IME composition refuses.
Search is literal Unicode simple folding with scalar offsets and optional word
boundaries (letters/digits/combining marks/underscore), bounded to 4 MiB documents, 1024 query
scalars, 10000 matches and 64 KiB replacement. Excessive queries refuse instead of
partially replacing. Matches are nonoverlapping and cached by model identity, content revision, query and options; selection-only navigation reuses them. No regex, language-specific expansion or all-match highlighting.
Use Small for 24pt or regular 30pt controls. Host command routing owns Cmd-F/G and
Escape; the demo shows focused query entry, editor focus return and Undo/Redo.

## Search scopes

Retain `ui.SearchScopeBar` with stable Items, Value{ScopeID, Query}, and caller-owned
counts (−1 hides the count). Accept Change via `Value = event.Apply(Value)` so
queued query/scope proposals merge. Actions compare against current accepted state: an All→Project→All queue is preserved, while choosing the already accepted scope is a no-op. 1–16 scopes, <=4096-byte query, optional
full-color IconImage, disabled items and Small (24pt) / regular (30pt). Arrows
move enabled toolbar focus; Return/Space selects; narrow bars scroll to reveal
focused items. SearchBox supplies query and clear through the same path.
SelectedStyle/ItemStyle integrate the page's scoped accent. Selected state is
announced in button help. Results, counts, filtering and persistence are owned
by the application; no local filesystem/network search or debounce is implied.

## Gradient well

Retain `ui.GradientWell` with `GradientValue{Stops, Angle}`. Accept typed edits
with `Value, err = edit.Apply(Value)`; Apply clones/reorders stops by position,
preserving stable IDs and unrelated color channels. 2–16 stops, fractions 0..1,
angle wraps at 360, byte sRGB interpolation over a checkerboard. Drag stops on the
shared track; Up/Down chooses a stop, Left/Right moves it by 1%, Shift ten; numeric
position/angle and the managed RGBA editor supply alternatives. Add chooses the
largest existing gap; Remove retains at least two stops. Handles may cross and
keep identity. Same-position stops keep their prior accepted slice order (stable sorting) and produce an ordered hard transition; the last equal-position stop supplies the color at that exact point.

The well opens an anchored scrollable editor with small/regular track sizes.
Optional NativeColor receives a stop identity, original color and owner-only
Accept callback. The inline editor closes before opening a native panel. Publish
Accept through a reserved completion; reopen/disable/removal revokes that result.
The gallery uses actual AppKit Colors without editing any document. Values are
live proposals; history and persistence stay application-owned. No perceptual
color-space interpolation, radial gradients or imported gradient-file format.

## Range slider

Retain `ui.RangeSlider` with Value{Lower, Upper}, Numeric bounds/Step, optional
IconImage, Small, Width and TickCount (0 or 2–21). Accept Change with
`Value, err = event.Apply(Value, Numeric)`; endpoint proposals merge with current
state, snap relative to Min and clamp at the other handle. Equal values are valid.
The shared track captures a pointer, converts actual painted bounds under DPI and
scrolling, consumes release, and cancels on loss/disable/source or geometry changes.
Live proposals do not create undo transactions. Caller rejection restores accepted
values. Up/Down selects one endpoint, Left/Right changes one step (Shift ten),
Home/End requests its limits. When thumbs overlap, use Up/Down or numeric fields
to choose the endpoint. The track exposes the active handle as an accessible slider;
the two labeled numeric fields expose each endpoint independently. Default regular
track is 320×40pt and small 240×32pt; Width may be 200–600pt. Horizontal only.

## Verification boundary

Portable race tests cover public models, Unicode offsets/undo, bounded values and
callback revocation. Native offscreen routing separately covers scaled captured
pointer gestures and release behavior; gallery raster checks cover all five pages
and the open gradient editor in light/dark. Actual AppKit provider rendering and
physical pointer/shortcut checks are recorded separately with the packaged build.
Quick Look installs a temporary responder after the exact parent, retains and
restores the prior responder, and yields control before releasing references.
Windowed close and fullscreen exit both use the same cleanup path. Loss of
preview control cancels and releases the session. `VELLUM_QUICK_LOOK_TRACE=1`
adds lifecycle phase and visibility/key/controller booleans to stderr for local
diagnostics; it does not log file names or paths.
