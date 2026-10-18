# Managed passive cell surface admission

Stage 0080 adds `Scene.PassiveTextSurfaceAllowed` for the managed grid value
surface. It checks the renderer's explicit text-batch capability and declines
custom widget initialization, render tracing and both private Inspector modes.
The original widget route remains authoritative for these configurations.
The helper does not change painting, the renderer, source publication or resource
ownership. Its test checks positive rasterx and each refusal without a window.

The driver owns only the current projection's body-cell values and shapes;
headers, panes, scrollbars, host input/AX and inline editing remain native.
Unsupported or late-invalidated values revoke completed measurement and request
owner reconstruction of ordinary cells. Rendering never publishes stale styles.
A late paint-only refusal can omit body values in that invalidated paint before
the owner rebuilds them; the complete original frame restores normal content.

See `specs/grid-performance.md` and `docs/reports/macos-grid-cell-surface/README.md`
from the repository root for evidence, reference routing and scope limits.
