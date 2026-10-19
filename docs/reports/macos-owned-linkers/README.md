# Owned Mac SDK linker adapters

The private ARM64 Mac SDK now replaces all five WebGPU linker-module distributions with freshly authored Vellum compatibility modules and explicit MIT notices. This closes their effective-source declaration gap in this SDK profile. Source-checkout builds and historical SDKs retain their existing graphs and original absence findings.

The WebGPU Go binding and prepared Core/typesetting sources are unchanged. Only Darwin ARM64 with CGO selects linkage, naming the exact static archive rather than searching for a WGPU library. The other four modules satisfy the binding's unconditional imports on Darwin and refuse their own native targets at compile time. These adapters provide no Windows/Linux/iOS/Android qualification.

Every generated source/license byte, directory and target policy is bound to the SDK manifest. Native example main modules, the driver, and generated independent applications receive all five explicit replacements. Archive/generator checks reject missing or changed declarations, wrong origins, partial capsules, symlinks and extra files. Legacy one-Darwin-capsule records remain supported without receiving new ownership claims.

The collector inventories the actual effective graph. It records the original requirement version/checksum separately from the authored source identity, whose version/checksum are null. Own notices remain in their capsules rather than being represented as upstream declarations. An unrelated missing declaration stays unresolved. The exact native archive, provenance and component notices remain separate, and distribution audit remains pending.

The [evidence record](evidence.json) identifies 137 passing affected Python test methods, real Go target-selection/refusal checks, fresh package/relocation and optimized example builds, plus the retained failures. The package was made from the disclosed working candidate before this documentation/commit; it does not claim a clean final-release artifact. No applications were launched and no OS settings were changed. The canonical source and both preceding SDKs were preserved during integration.

This is an SDK source/linkage checkpoint. Actual native application acceptance, presentation/60 Hz measurements, resource/endurance and physical/human checks, native/font/asset notice applicability and distribution approval, and the remaining Mac delivery gates are still open. See the [SDK contract](../../../specs/sdk.md) and [Mac delivery plan](../../../specs/macos-delivery.md).
