# Mac document controls

Five separately tested, Prism-reviewed and committed phases: Finder-style document
list, Mac outline, AppKit Colors panel, AppKit Fonts panel and circular angle dial.
Public API expansion is expressly authorized by the current user request and
supersedes the older internal-app freeze in CONTRIBUTING. Rendering, pacing and
release policy remain unchanged. Preserve original-color artwork, compact Mac
spacing, scoped light/dark themes and keyboard/accessibility routes.

Bound document and outline catalogues to 256 items; reuse managed source-list
selection/routing. List columns sort deterministically and resize through controlled
proposals. Hidden outline selection remains application-owned. Native shared panels
have exact-parent ownership, single-session admission, explicit Apply/Cancel and
lifetime cancellation; return values through reserved window completions. Full-color
demo buttons stay separate from system-owned panel contents. The angle dial keeps
one canonical finite value, wraps angles, supports keys and a numeric field.

Begin offscreen, using public API tests, real managed event routing, light/dark
rasters and native Foundation recipe checks. After packaging/relaunch, the user's
sign-out authorizes bounded desktop tests; keep native observations distinct from
automated evidence. Never interact with protected macOS system dialogs.
