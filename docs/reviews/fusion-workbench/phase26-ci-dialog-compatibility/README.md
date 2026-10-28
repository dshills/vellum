# Fusion-style workbench — phase26-ci-dialog-compatibility

Framework run 40663203450 failed only frozen workspace modal-open image checks after the earlier ordinary-dialog surface fix. Dialog Body/Scene/Close now require an explicit node/kind dialog color before applying theme colors. Generic content defaults retain historical native chrome; MacTheme explicit dialog pairs retain the Forge fix. Both frozen 10k/1M records pass at batch7 without any input, model, image-golden, API, workflow or gate changes; remote full CI is checked after integration.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: Fixed the comment attachment. Theme and ResolveStyle have only Default, Kinds and node Style as color sources; classes, variants and theme state layers mentioned by the reviewer do not exist in this API (ui/style.go). The opt-in tests deliberately exclude generic Default, then ResolveStyle still applies the full supported overlay. Explicit MacTheme cases and generic-theme regression pass. Immutable replay goldens remain byte-for-byte unchanged.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
