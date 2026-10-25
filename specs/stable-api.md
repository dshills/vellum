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

Root and native driver are separate Go modules. This private SDK retains explicit replacement origins and no public module tags are created. Public Go module release would require separately coordinated root/driver version tags and authorization. SDK consumers must preserve every generated native-linker replacement.

## Verification and change review

The final pass compares source changes against the accepted RC: no exported declarations are removed or reshaped; the API version constant changes deliberately. Existing root tests and the independent sdk-contract/Notes/shell consumers exercise public usage. Run focused tests for actual changes, review checkpoints with Prism using Anthropic Claude Sonnet5.5, and bind final SDK/app manifests to committed source. The final CI report records exact workflow outcomes. New compatibility obligations guide subsequent v1 maintenance; existing historical reports remain immutable.
