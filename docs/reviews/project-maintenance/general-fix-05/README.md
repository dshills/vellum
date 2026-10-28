# Generate minimal applications from dedicated public API templates

Prism: Anthropic Claude Sonnet 5.5, complete selected diff coverage.

Addressed template coupling: main/model/test are now dedicated templates. Text suffix keeps native consumer code outside the root portable Go package walk. Generated independent app race tests pass; hello diagnostics remain available separately.

Checks are offscreen/source checks; no new native qualification is claimed.
