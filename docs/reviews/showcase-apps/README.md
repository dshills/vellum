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

## Final native packages and captures

The private final packages were built from clean source commit
`b646a9c72ace856225e88abb373a297100c2303e`. All three ZIPs were extracted into
separate local directories and passed `codesign --verify --deep --strict` with
their full-color ICNS assets present. Source file hashes match the package
manifest. Final Forma/Luma/Pulse bundles were launched and their complete windows
visually inspected; raw unmodified app-window JPEGs are in `docs/screenshots`.
The corrected Luma profile caption was verified by switching Mono then Original
in the final bundle. Forma and Pulse production sources are unchanged from the
interaction checks above. Closing all three preceding disposable demo windows
was followed by successful final launches. No personal documents were opened.

The README describes real functionality and limitations. No full release or
performance campaign was repeated. Physical coordinate targeting did not produce
a reliable control result in this session; native AX actions and offscreen
routing are recorded separately. This is not new human accessibility acceptance.
The apps are private, ad-hoc signed and unnotarized; no public assets/tags were
created. Final app copies remain under `/private/tmp/vellum-showcase-final`.

Final documentation checkpoint: Prism / Anthropic Claude Sonnet 5.5 reviewed
all eight textual files with complete coverage and no findings. The three binary
JPEGs were visually inspected separately, including their native chrome and
selection appearance. All 73 local documentation links resolved and each stored
image SHA256 matched its capture record (1418×931). No pixel edits or conversions
were applied to the captured bytes.
