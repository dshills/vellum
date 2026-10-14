# Managed breadcrumb navigation

The public `ui.Breadcrumb` component adds controlled location/hierarchy navigation
without native backend types. It composes managed toolbar, quiet-button, current
label and decorative-icon records. Paths admit 1..64 valid unique domain IDs;
encoded targets remain stable independently of labels. Accepted navigation is an
owner path update, and rejection preserves the current path and ancestor identity.
Read-only paths have no action buttons. Queued requests check the latest callback,
root/path identity, current validity and disabled state before delivery.

The Gallery demonstrates navigation acceptance/rejection, reset and disabling.
The relocated SDK consumer exercises the public component API. The contract is
[breadcrumb.md](../../../specs/breadcrumb.md).

Focused race tests pass for actual managed Enter/pointer routing, rejected button
identity, accepted current-label retirement, literal labels, read-only behavior,
narrow overflow Home/End focus revelation and four-DPI completed software frames
at 96/120/144/192. Inert Cocoa selectors read the real component's button/static
roles and action metadata; decorative separator IDs are omitted. A same-sized
path reorder before admitted owner work drains is refused. This creates no OS
window and does not qualify human VoiceOver or physical input.

Two deliberate source overlays fail meaningful checks: removing complete-path
identity comparison delivers an old Cocoa-owner request to reordered content;
removing the count guard admits 65 otherwise-valid locations. The faults preserve
production source and all 6,214 prepared dependency leaves. Initial fixture
failures are retained: a content-sized parent allocated no scroll width, the
400-pixel frame clipped the current label at DPI192, and a new overflow assertion
used WidgetBase instead of the containing Frame. The corrected frame uses a
width-allocated parent and720-pixel fixture; narrow overflow is verified separately.
The first native fault attempt only failed to compile and is not counted as a
detected guard fault. No backend layout/renderer change follows these observations.

Final portable/adapter race and vet, all 28 SDK exporter regressions and native
Gallery build pass. Internal final review and one bounded Prism review using
Anthropic Claude Sonnet 5.5 precede commit. The qualification record binds exact
final source and six gate receipts. This report was updated after the checks;
all executable source and test bytes remained unchanged. A source-only Gallery build does not qualify a
live breadcrumb sequence. The independent current endurance campaign retains
its earlier `e7f8e30` endpoint and cannot qualify this later component. Full Mac,
P6/P11, physical input/displays, human accessibility, native/GPU memory and public
distribution remain open.
