# Showcase application checkpoints

Requested: original runnable demos with native screenshots for the main README.
Selected Forma (retained design canvas/history), Luma (original images and
collections) and Pulse (virtualized fixture data and retained charts). Scope is
example apps plus the specific styling corrections reported by the owner during
native inspection. No new framework APIs, backend/gate changes or public binaries.

The owner identified cramped inline icons/labels and unusual table selection.
The shared example sidebar now explicitly centers content with a 12dp gap.
The existing grid now honors the scoped semantic palette, insets cell text and
uses continuous read-only selected rows; editable and unselected cursors remain.
Both render routes use the same scoped override. Light/dark 96/160/192dpi tests,
mutable scoped-theme changes and existing reference/surface/packet pixel parity
checks pass. Unthemed rendering remains unchanged. These checks are not a broad
performance campaign or a change in release policy.

Application model and offscreen routing checks, Prism review and the final
native-window screenshot provenance are recorded with each completed checkpoint.
Initial native app captures were inspected; final README captures follow source
review and preserve real app pixels. No personal data or global settings used.

The normal model race/vet checks pass. The managed-control offscreen race/vet
checks pass after supplying the harness's explicit background and replacing
unsupported GPU extension descriptors with empty layout nodes. No production
control callbacks are replaced; actual down/up events route to measured bounds.
GPU geometry has model coverage and the complete surfaces are inspected in the
native apps. The initial routing attempt was a fixture configuration failure,
not an application launch failure. The repository example check now builds all
three binaries and tests this module.

Native accessible actions verified Forma add/Undo/Redo with the same layer ID,
Pulse service selection/inspector, the 93-row attention filter, sample advance
and native-toolbar reset. A coordinate click did not select its intended row;
these checks therefore establish native accessibility routing, not physical
pointer acceptance. No global settings or personal files were changed.

The first implementation Prism response had one malformed, non-actionable
placeholder ("Not applicable; placeholder removed."). It is not a code defect
or a clean review verdict; a fresh final source review follows the tested
fixture correction.

Final source review: Prism / Anthropic Claude Sonnet 5.5, complete coverage of
29 files, no findings. Luma's native favorite filtering (six to three), filmstrip
selection/primary preview and Mono/Original rendering also passed accessible
actions and visual inspection. That inspection found one cosmetic caption left
as "Original color" for Mono; it now names the selected look. The final model
race/vet receipt includes that correction. A final checkpoint review covers it.

Final checkpoint Prism coverage: 32 files, complete, no findings. Source is
committed before the final private app build and README window captures.
