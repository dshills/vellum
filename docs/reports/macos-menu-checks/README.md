# Controlled menu check actions and separators

MenuButton Options now add Checkable/Checked boolean actions and passive
Separator entries. Application-owned Checked is accepted state; choosing an
action submits the inverse state through Change.Checked and closes the popup.
The application can accept, reject or normalize the proposal. Reopening derives
its marker from accepted state. Descriptor/value changes revoke queued requests
through the existing popup epoch and originating-window retirement.

Separators retain unique descriptor IDs, count toward the 64-entry bound, and
render with the existing native separator widget. They have no label, focus,
action callback, accessible child or manual input target. All-separator or
otherwise passive menus cannot enter focus/action input. Other control kinds
refuse the new metadata. Ordinary action-menu behavior remains unchanged.

The managed popup shows a bundled checkmark for accepted true. Its AXMenuItem
publishes an accepted boolean value with Checked/Unchecked value description.
An inert fixture invokes production Cocoa selectors on unattached elements;
it creates no OS window, successful action dispatch or native attachment.
[Apple's accessibility protocol](https://developer.apple.com/documentation/appkit/nsaccessibilityprotocol)
defines these value/description properties. This evidence does not establish
human VoiceOver announcements or physical Mac menu behavior.

## Scoped qualification

Portable validation checks false/true states, malformed separator combinations,
empty/duplicate IDs, other-kind metadata refusal and passive focus exclusion.
An actual manual-backend popup routes keyboard and pointer check actions, skips
separator/disabled entries, preserves rejected accepted values, renders both
check states, and refuses stale queued proposals after descriptor value changes.
Existing flat-menu lifetime, peer-window drain/wake and ordinary action tests
remain in the targeted suite. Software captures were inspected: literal labels,
separator line, disabled checked action and accepted marker render correctly.
Two intentionally broken overlays fail exact proposal/separator assertions.

Final portable/adapter race tests, vet, native tagged build and gallery tests
must pass. Prepared dependencies retain all 6,207 baseline bytes. The new controls
are demonstrated on the gallery's Icons and actions page. The significant
source/spec/report checkpoint requires internal review and Prism using Anthropic
Claude Sonnet 5.5 before commit; indexed evidence preserves producers, captures,
all receipts, negative controls and reviews. Native binaries are built only.

Live tests remain paused while the human uses the computer. Physical input/menu,
human screen-reader, cross-display placement and complete Mac qualification
remain open. Full P6/P11 and current interacting lifecycle acceptance remain open;
this control expansion does not qualify a frame-rate or memory target.
