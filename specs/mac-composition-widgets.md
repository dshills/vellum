# Mac composition widgets and theming

User-authorized expansion after the native toolbar work. Preserve current
renderer/pacing policy and native ownership. Each widget has a focused test,
Prism Anthropic Claude Sonnet5.5 review and commit; package the gallery afterward.
All additions are independently usable through public Go APIs. Full-color
artwork must remain immutable and visible alongside Mac-style typography,
spacing, focus and disabled states. Source values/callbacks belong to one owner.

## Shared scoped theming

Theme scopes are value snapshots at accepted View boundaries. Nested scopes
replace inherited defaults; explicit node style wins. Removing a scope restores
inheritance, not a previously scoped default. Invalid scopes reject atomically.
Retained control identity/value/focus survive changes; transient menus retire.
Provide semantic opaque colors and light/dark/accent Mac presets. Do not mutate
OS appearance or global renderer colors. Demonstrate live local theme choices.

## 1. Cascading menus

Extend flat actions with bounded immutable branches, shortcut captions, rightward
submenu indicators, disabled ancestry and controlled check leaves. Reuse native
popup/keyboard ownership and action epochs. Parent branches never invoke an
application action. Escape/Left close appropriate popup scope; leaf selection
commits once, closes the chain and resolves the current callback. Removed,
disabled/replaced or closed menus cannot invoke old actions. Full-color artwork
must work at every level. Expose opened submenu hierarchy to accessibility.

## 2. Split button

A separate primary action and adjacent disclosure segment with a shared compact
Mac outline. The main segment executes only the primary callback; disclosure
opens existing action menus, including cascades. Support small/regular sizes,
full-color icons, disabled state and keyboard access to both affordances.

## 3. Inspector panel

A reusable property panel with title, optional colored artwork, aligned labels,
collapsible retained sections and mixed-value field presentation. Compose existing
forms/disclosures/scroll primitives. Controlled section changes preserve editor
values and focus scope; no document model or CAD-specific behavior is added.

## 4. Column browser

Bounded hierarchical navigation through resizable columns, keyboard accessible
row selection/navigation, icons/full-color thumbnails and optional final preview.
Controlled selection carries stable domain path IDs. Invalid/stale paths,
disabled ancestry and removed source rows cannot reach callbacks. Reuse lists,
scroll containers and split panes. No filesystem access or unbounded tree walk.

## 5. Filter/rule editor

Typed criterion/operator/value rows with add/remove actions and bounded nested
All/Any groups. Return controlled structural proposals to the application, with
stable row/group IDs and validation of operator/value compatibility. Keep values
retained across theme/structural updates. No evaluation of arbitrary code or
unbounded predicate execution is implied. Use native-feeling controls and spacing.

## Evidence

Meaningful portable and focused managed-routing regressions cover acceptance,
rejection/stale callbacks, bounds, theme retention, keyboard behavior and cleanup.
Use independent consumer checks where APIs are added. Separate deterministic
routing evidence from packaged native screenshots/input checks. Reuse passing
checks when implementation has not changed; no exhaustive release campaign.
