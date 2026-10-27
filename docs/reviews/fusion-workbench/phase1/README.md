# Fusion-style workbench — phase1

Add typed GPU input without changing legacy routing, plus portable camera/ray/box and gesture helpers. Real Core manual routing verifies capture, exactly-once delivery, Escape and swallowed release; owner cancellation is queued rather than reentrant.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: dfc0c9dd1c84d1aa addressed with explicit captured/uncaptured routing guards and the passing real router regression. Core capture already precedes ordinary routing; the extra guard also refuses direct synthetic duplicate delivery and unowned primary up/drag. Different event types in the listener map cannot change same-type listener order.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
