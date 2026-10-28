# Generate complete isolated document applications with tested persistence

Prism: Anthropic Claude Sonnet 5.5, complete selected diff coverage.

Added deterministic nonempty source admission and explicit UTF-8. Replaced main.go string surgery with a dedicated template. Copying app-owned tests is intentional and generated-module race/vet passed. SDK verification now generates and checks the document consumer as well as the minimal one; generation itself does not implicitly build or run applications.

Checks are offscreen/source checks; no new native qualification is claimed.
