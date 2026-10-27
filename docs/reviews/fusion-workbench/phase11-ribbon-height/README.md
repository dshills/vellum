# Fusion-style workbench — phase11-ribbon-height

The build4 native visual check found an empty band from Core Tabs 96dp content-page minimum. Grouped-toolbar-only page hosts now clear that floor and the ribbon clears its inter-row gap. The focused native suite covers ribbon geometry/overflow, normal tab navigation, settings navigation and compact browser/timeline routing. New three-DPI test verifies compact total height, retained native identity, and restoration of the ordinary96dp minimum on descriptor reuse. Input, command/history, public API and release policy are unchanged.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: The possible stale-record closure finding is a false positive in this reconciler: reconcileScoped retains the same record together with its compatible native widget, or mountShallowScoped creates a new record and widget. Native Tabs/frame widgets are never transferred to a replacement record. Existing tab-bar and button stylers use the same record-lifetime pattern. The new ordinary/ribbon descriptor-reuse test passes at160/240/320DPI, with native widget identity retained and normal minimum restored. Original Prism finding is preserved.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
