# Fusion-style workbench — phase12-ribbon-fit

Supersedes the intermediate build5 page-minimum-only fix: grouped toolbar pages retain their intrinsic vertical tool height before the host fits them. Scroll and grow defaults for ordinary pages remain native. A strengthened geometry regression requires a nonzero compact ribbon and contained visible action, at160/240/320DPI, in addition to existing tab/settings and compact control routing checks. Build5 was closed and is not the final delivered app.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: The closure finding is a false positive: drivers/cogent/go.mod targets Go1.25.6 with per-iteration loop variables. reconcileScoped retains records and native widgets together or mounts a fresh pair; pages are owned by that retained Tabs record, never remapped to another record. This one-to-one lifetime is also used by existing tab/button stylers. The strengthened reuse test passes with native widget identity retained and the ordinary page minimum restored. No unsupported pre-Go1.22 execution is claimed.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
