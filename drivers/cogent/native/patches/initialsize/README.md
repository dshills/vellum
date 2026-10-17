# Initial desktop client size

Stage 0074 carries a finite positive Dp request from Stage to NewWindowOptions.
Explicit requests bypass content measurement and screen-fraction sizing. Desktop
creation selects one locked screen snapshot after first-window settings update,
then uses that same snapshot for conversion/constraints and GLFW setup before
Show. Saved geometry clears the request. The request never becomes a persistent
body minimum; manual physical viewports use their existing options.

Conversion uses logical DPI / 160 with nearest-pixel rounding, screen bounds and
a representable margin below GLFW's signed 32-bit dimension limit. A final bound
preserves at least one native coordinate unit after the legacy fractional-DPR
round trip. Core/system regressions exercise real creation options, chosen-screen
DPI, saved precedence, tiny fractional sizes, extreme finite requests, native
integer limits and unchanged unset/fullscreen cases. The Vellum adapter regression
checks the actual body stylers and independent application minimum styles.

Patch input/output file hashes and the ordered manifest permit fresh preparation.
Live initial/client/render geometry and external decorated AX frames are separate
observations. The smoke fixture's programmatic resize is not physical pointer,
cross-display, full-size-content, screen-reader or performance qualification.
