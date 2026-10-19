# Native macOS example checkpoint

Four independent Go commands demonstrate Calculator, Notes, Weather and
Reminders using public Vellum APIs. Models are bounded, in memory, and contain
fictional data. They access no Apple application stores, accounts or location.
See [the specification](../../../specs/macos-app-examples.md) and
[consumer instructions](../../../examples/macapps/README.md).

## Framework behavior

The checkpoint adds per-window native unified chrome and appearance, explicit
fixed-size creation, optional suppression of implicit root scrolling, axis growth,
text alignment, rounded surfaces, borders and public symbols. Pinned Core stages
94–96 preserve native view/layer ownership, align rendered text and interaction
geometry, honor positive label growth, and preserve explicit fixed initial sizes
when saved geometry exists. The zero-value policies preserve existing consumers.

Calculator uses a fixed 240 × 420Dp client without root scrolling. A software
viewport test checks every key and routes pointer down/up for 1 + 2 = 3. A native
capture separately verifies the complete keypad, absent scrollbars and disabled
zoom control. The other examples have native resizable windows.

## Evidence boundaries

Root package race tests, focused driver/Core race tests, the independent example
module's race/offscreen tests and vet, native command builds, and all 29 recipe
tests form the automated checkpoint. A fresh preparation verifies the pinned
input/output hashes for all 96 Core and six typesetting stages.

Live keyboard/accessibility actions exercise Notes pending title admission before
Undo/Redo and body replay; Weather city/unit/search changes; and Reminders pending
composer creation, completion/replay, and native details text settlement. These
are distinct from physical hardware pointer qualification. No IME/input source,
VoiceOver, system appearance or personal Apple application data was changed.

A hidden AppKit fixture verified per-window chrome and original content view and
layer identity through style restoration. Cocoa may recreate its own close
button; the replacement retained its owning window, action and target. This does
not qualify real Metal fullscreen or display migration. Native automation tool
launch latency is not an application startup performance measurement.

Development evidence (logs, source/implementation captures, exact symbol origin
responses and private unsigned app manifests) is retained externally under the
workspace artifacts/macos-app-examples and artifacts/macos-window-chrome areas.
Apple reference screenshots are not committed or shipped as app assets. Three
new weather symbols have exact commit/path provenance and Apache-2.0 attribution;
that narrow verification does not close the broader distribution audit.

## Remaining acceptance

These examples demonstrate working UI capabilities; visual acceptance remains
open. Native app-specific toolbar items, adaptive safe-area insets, Tahoe glass
and sidebar vibrancy are missing. Notes still uses a monospace plain-text editor
with a line gutter instead of native rich Notes presentation. Weather uses a flat
background and fixed sample forecasts. Reminders omits persistence and services.
Physical pointer, native accessibility/IME, presentation, endurance and SDK rights
checks remain governed by the Mac completion plan. This checkpoint does not
approve public distribution or declare the Mac framework complete.
