# Stable v1 API compatibility

The owner approved the stable API/documentation and private macOS v1 packaging pass on25 October2026. This contract adopts source compatibility for the named v1 APIs; it adds no features and does not broaden runtime or distribution qualification. Current identity is `v1.0.0` in release/version.json and vellum.Version.

## Root module

The exported Go API of the root `github.com/dshills/vellum` package and its `app`, `ui`, `command`, `history`, `document`, `editor`, `tasks`, `execution`, `virtual`, `canvas`, `placement` and `workbench` packages is the v1 source-compatibility baseline. The release source commit and SDK manifest identify the exact baseline. Applications must follow each package's documented ownership, validation, resource bounds and lifetime requirements.

Within v1, preserve exported names, signatures, existing struct fields, implemented public interfaces, constant meanings, documented file-format compatibility and errors.Is contracts. Additions may introduce optional APIs/fields without changing the documented zero/default behavior. Applications should use keyed literals for exported configuration/descriptor structs. Do not add required methods to public interfaces or silently change accepted data or command/history semantics in a minor/patch release. Deprecate before removal; an incompatible stable API change requires a new major version and migration instructions. Bug fixes may reject inputs already outside documented contracts. Stable version numbers may change, while existing wire/file format versions change only under their own documented migration policy.

`replay` and `automation` remain explicitly experimental and outside this compatibility promise. Their protocol versions and limits remain recorded separately; their presence in a v1 module does not grant stable protocol semantics.

## macOS driver facade

The separately imported `github.com/dshills/vellum/drivers/cogent` module retains source compatibility for:

- `Driver` construction and `Run`, with normal fields `Theme`, `Commands`, `CommandMenus`, `OnError`, `FollowSystemAppearance`, `OnSystemAppearance`, `TitlebarToolbar`, `TitlebarToolbars` and legacy `TitlebarToolbarPrototype`.
- `TitlebarToolbar`, its retained `TitlebarToolbarPrototype` alias target, `TitlebarToolbarItem`, and `CustomizeTitlebarToolbar`; documented item/window/customization bounds remain.
- `SystemAppearance`, `MacSystemTheme`, `SettleTextField`, `SettleEditor` and `SettleGridEdit`, with exact window/model lifetime and owner-callback requirements.

This is an explicit facade whitelist. Other native exports, backend extension interfaces, Manual/desktop automation, accessibility inspection/action helpers, diagnostic/timing APIs, custom GPU resources/views and opt-in rendering/reuse fields remain experimental. Existing exported fields are retained during this release pass; callers enabling experimental paths accept their separate contracts. Native main runs once per process. App menus are process-wide, and callers own active-document routing when windows use independent registries.

## Behavior and platform limits

Source compatibility does not promise pixel-identical AppKit controls, arbitrary-workload performance, cross-window placement, unrestricted data sizes or qualification of every hardware/display combination. Use the [Mac support guide](../docs/macos-v1.md), [framework contract](framework.md), per-control specifications and their existing limits. The supported personal-use configuration is Apple Silicon on the recorded macOS26.6.2 host. Other configurations remain unqualified. Human IME and VoiceOver usability remain explicitly unqualified under the owner's existing choice.

Root and native driver are separate Go modules. The owner-authorized
26 October source release creates the public root `v1.0.0` tag. Native apps
continue to build from the tagged checkout with explicit local replacements;
no independently consumable native-driver module tag is released. The later
`macos-v1.0.0-build13` packaging release publishes a prepared SDK and six apps
from source `64bf9cf`; see the [publication record](../docs/reports/macos-v1-build13-publication/README.md).
The private SDK retains its recorded identity and replacement origins. Separate
native-module releases require coordinated version tags and authorization. SDK
consumers must preserve every generated native-linker replacement.

## Verification and change review

The final pass compares source changes against the accepted RC: no exported declarations are removed or reshaped; the API version constant changes deliberately. Existing root tests and the independent sdk-contract/Notes/shell consumers exercise public usage. Run focused tests for actual changes, internally review checkpoints and record findings and verification, and bind final SDK/app manifests to committed source. The final CI report records exact workflow outcomes. New compatibility obligations guide subsequent v1 maintenance; existing historical reports remain immutable.

## Root source compatibility guard

The [developer-only API checker](../tools/apicompat/README.md) compares the thirteen
stable root packages with the reviewed [v1 baseline](../release/api/README.md).
Checks are read-only and refuse missing packages, invalid policy/origins or a
corrupted baseline. Its generic-constraint check is conservative because the
pinned upstream semantic diff misses some narrowings. The root Version value is
an explicit metadata-checked exception; its type/removal and other constants
remain guarded. Native facade, behavior, errors and file-format obligations still
use their existing consumer/runtime tests and review. The tool adds no framework
dependency, feature, platform claim or change to the shipped private release.

The native facade has an external-package compile contract in
`drivers/cogent/stable_facade_contract_test.go`. It checks the whitelisted public
field reads/writes, function/method signatures, `app.Driver` implementation and
legacy toolbar alias in both directions. The fixture does not call native APIs or
qualify behavior, defaults, error values or additional platforms; existing native
and lifecycle tests retain those obligations. It deliberately does not lock
experimental exports or extend the root export baseline to the whole driver.
