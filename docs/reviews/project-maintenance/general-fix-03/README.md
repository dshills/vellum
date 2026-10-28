# Report Forge recovery failures without hiding ordinary save results

Prism: Anthropic Claude Sonnet 5.5, complete selected diff coverage.

Invalidate is synchronized by Application.mu and is worker-safe. Addressed redundant repaint concern by invalidating only when the recovery error text changes. Encode/write failure and successful clearing tests pass.

Checks are offscreen/source checks; no new native qualification is claimed.
