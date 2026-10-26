
## Guard the stable v1 root API with a reviewed semantic baseline

Prism: `mac-v1-root-api-guard-cgo.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Initial review: physical directory identity and named/embedded generic constraint checks implemented with regressions. Second review: Go-only root policy now explicitly rejects cgo imports even in build-excluded stable-package files. Final review dispositions: 15e408ddb7136cc8 is not reproduced: vellum.go declares Version="v1.0.0", matching release/version.json, and the real baseline check passes. 9521d5961afcd33c is an acknowledged pinned-library coupling covered by TestSourceCompatibilityChanges/root_version_value_exception, which executes the actual pinned apidiff, plus type/removal/other-package negatives; dependency upgrades must retain these tests. No exception widening or arbitrary environment/proxy inheritance added. The generated root.export binary was outside LLM text review and separately verified by recorded SHA256, clean source revision, package membership, semantic roundtrip/current comparison, and absence of local source paths. The root module graph and accepted v1 build9 delivery remain unchanged; native facade and behavior retain separate checks.

Focused checks: `mac-v1-api-guard-final-checks.json`, `mac-v1-api-guard-review-fixes-checks.json`, `mac-v1-api-guard-cgo-checks.json`.

## Keep API test failures visible in Windows CI

Prism: `mac-v1-api-workflow-steps.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism reports no findings with complete coverage. Only the workflow command grouping changes: test and vet run in independently failing steps. The existing e2530b1 API source/baseline passed the initial Mac/Linux/Windows matrix (40258277100). Its passing focused Go checks are reused because no Go source changed. The corrected workflow still needs its own CI observation; no native app or accepted release bytes changed.

Focused checks: `mac-v1-api-guard-cgo-checks.json`.

## Check the stable Mac facade from an external Go consumer

Prism: `mac-v1-native-facade.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism reports no findings with complete coverage. External-package compile assertions cover only the adopted native facade whitelist, including both field read/write directions, Run/app.Driver, settlement signatures, appearance, customization and the retained toolbar alias. The fixture is never executed and cannot open windows or touch settings. Existing toolbar/appearance/settlement regressions and native offscreen vet pass. Runtime implementation and signatures are unchanged; stale package and legacy/plural toolbar comments are corrected. This does not extend the root semantic snapshot to the entire native package or establish behavior/default/error/physical-platform qualification. Accepted v1 build9 archives remain unchanged.

Focused checks: `mac-v1-native-facade-checks.json`.

## Record green v1 maintenance CI and integration handoff

Prism: `mac-v1-maintenance-ci.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism reports no findings with complete coverage. The final maintenance source 20659c9 passed Framework 40258851542, Stable root API 40258851606 and Platform adapters 40258851622. This checkpoint changes reports only and reuses the exact passing focused checks; it does not claim a future main CI result or change the accepted private 382e0e4/build9 packages. No broad campaign was repeated and no distribution or human/hardware qualification claim was expanded.

Focused checks: `mac-v1-api-guard-cgo-checks.json`, `mac-v1-native-facade-checks.json`.
