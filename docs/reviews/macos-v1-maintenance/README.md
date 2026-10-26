
## Guard the stable v1 root API with a reviewed semantic baseline

Prism: `mac-v1-root-api-guard-cgo.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Initial review: physical directory identity and named/embedded generic constraint checks implemented with regressions. Second review: Go-only root policy now explicitly rejects cgo imports even in build-excluded stable-package files. Final review dispositions: 15e408ddb7136cc8 is not reproduced: vellum.go declares Version="v1.0.0", matching release/version.json, and the real baseline check passes. 9521d5961afcd33c is an acknowledged pinned-library coupling covered by TestSourceCompatibilityChanges/root_version_value_exception, which executes the actual pinned apidiff, plus type/removal/other-package negatives; dependency upgrades must retain these tests. No exception widening or arbitrary environment/proxy inheritance added. The generated root.export binary was outside LLM text review and separately verified by recorded SHA256, clean source revision, package membership, semantic roundtrip/current comparison, and absence of local source paths. The root module graph and accepted v1 build9 delivery remain unchanged; native facade and behavior retain separate checks.

Focused checks: `mac-v1-api-guard-final-checks.json`, `mac-v1-api-guard-review-fixes-checks.json`, `mac-v1-api-guard-cgo-checks.json`.
