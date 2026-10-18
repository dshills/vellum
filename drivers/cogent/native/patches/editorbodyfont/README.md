# Editor document and gutter typography (stage87)

Normalize a fresh body rich/text style to Monospace before Lines admission,
character sizing and body/tab shaping. The widget UI font remains available to
line-number chrome; theme/node size, weight and slant still reach both. Neither
the widget style nor earlier shaped text is mutated through a shared style.
Full relayout clears the tab shape alongside line and gutter shapes so font,
scale and wrap changes cannot retain a tab from an earlier layout.

Actual Manual rendering tests at160/162DPI cover retained font-size changes,
monospaced body/tab vs themed gutter, exact selection and restored pixels.
Two existing direct-scene test fixtures initialize bundled-font shapers before
layout without requiring an external system-font-cache setting; the temporary
factory is restored immediately, and existing geometry/visual assertions stay.
No native input, acquisition, lifecycle, public API or SDK refresh is introduced.
