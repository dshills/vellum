# Mac widget essentials

Implement in five reviewed, tested, separately committed phases: image well,
searchable full-color icon picker, settings pane, AppKit alert service, and an
expanded reusable calendar. User authorization covers these public widget APIs.
Keep renderer/pacing unchanged. Tests remain offscreen; package/relaunch at end.

All managed components use stable IDs, one owner, controlled callbacks, bounded
models, current eligibility checks and shared themes. Preserve immutable color
pixels. Image drop transports references only; decode on workers. Picker popovers
revoke stale sessions and bound recent IDs. Settings retain inactive content.
Alert service is optional and asynchronous from workers, with bounded options,
stable response IDs, full-color PNG icon, cancellation/shutdown cleanup and native
exception fences. No offscreen test may display an alert or file chooser.
Calendar reuses civil-date/DST validation, with a reusable month grid, Today,
disabled dates and controlled selection; do not create an unrelated date engine.

Use portable and focused actual managed-routing tests; record offscreen rasters
separately from desktop evidence. Prism uses Anthropic Claude Sonnet 5.5 for each
checkpoint. Existing framework-scope freeze in CONTRIBUTING predates and is
superseded by the user's explicit widget/theming expansion request.

Offscreen integration acceptance: render all five examples in light/dark themes,
including the open icon-picker popup. Calendar weekday/day columns must align
across sparse, enabled and disabled rows; both month arrows must have hit geometry
inside a compact header. Selected dates use the scoped accent/contrast colors.
Managed popovers use the local Mac popup surface rather than a startup theme.
An independent SDK consumer must construct the components without native imports.
Package both personal gallery variants from a clean committed source; native
alert, file chooser/external drag and physical-pointer checks remain deferred.
