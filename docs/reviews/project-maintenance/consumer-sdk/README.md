# Verify relocated SDK consumers separately from native diagnostics

Prism: Anthropic Claude Sonnet 5.5, complete selected diff coverage.

The generated consumer passes independent race/vet/build with replacement origins restricted to the extracted SDK. Optional runtime now explicitly runs a separate SDK diagnostic; added its own origin validation and race/vet before build. CLI help distinguishes diagnostic runtime from minimal consumer verification.

Checks are offscreen/source checks; no new native qualification is claimed.
