# Input ordering fix: review record

10 October 2026. Prism staged reviews using OpenAI/gpt-6.1-sol; both areas have complete coverage, zero truncation/skips/exclusions and zero findings. Base commit: `e1c8e1b12f16388407d5c8e40d2e583bf9c93e4c`; branch: `codex/foundation-completion`.

| Area | Files | Run ID | Staged diff SHA-256 |
|---|---:|---|---|
| code | 7 | `2aa495b406a580ea7be20c86db4cc7fa` | `950148f1b63fffa9e473a618fded4b7f7abfa73cbfe518a66ad9abbf0cfbdd8e` |
| report | 7 | `124829d54a234166584d446b30d2b6bb` | `3990e41bcf71a402f4ca5dd1fd09816c05a5012f44b550a0be8998b2c01d63b0` |


Invocation per area: `prism review staged --paths <explicit comma-separated paths> --no-auto-exclude --format json --out <area report> --fail-on medium`.

Reviewed scopes:

- **code**: `prototypes/foundationintegration/transport/ordered.go`, `prototypes/foundationintegration/transport/ordered_test.go`, `prototypes/foundationintegration/main.go`, `prototypes/foundationintegration/cmd/foundationaudit/main.go`, `prototypes/foundationintegration/prepare.py`, `prototypes/foundationintegration/patches/0010-ordered-deferred-events.patch`, `prototypes/foundationintegration/patches/original-hashes.json`.

- **report**: `docs/reports/INPUT_ORDERING_FIX.md`, `docs/decisions/0009-ordered-native-events.md`, `docs/reports/FOUNDATION_ACCEPTANCE_AUDIT.md`, `docs/reports/FOUNDATION_INTEGRATION_RESULTS.md`, `prototypes/foundationintegration/README.md`, `prototypes/foundationintegration/verify_acceptance.py`, `prototypes/foundationintegration/evidence/20261010-ordering/manifest.json`.

No findings required fixes. Canonical source/evidence/preparation hashes match the committed manifest. Raw evidence was validated locally rather than submitted as implementation code. Authored whitespace checks pass; the nested patch retains standard unified-diff context markers and tabs verbatim, which Git whitespace checking otherwise flags.

Validation: ten repeated race-test runs of transport/input/semantics, offscreen core suite, vet, actual native fixture with saturation checks, and 1,000 hidden-window lifecycle/query cycles pass. Fixed audit observes 32/32 earlier updates at the shortcut across two batches of 16; its old-pump control retains the 16/32 failure. Later native admissions stay after legacy shortcut boundaries, including arrivals during dispatch. Acceptance runner exit 1 explicitly preserves the other open G1 gates.
