# Reversed-order display-clock control: preparation

The [earlier off/on pair](../macos-display-scheduling/README.md) remains sealed:
p95 was 51.101 ms off and 33.947 ms on. Both miss the unchanged 33.3 ms gate.
That one ordered pair suggests a scheduling lead but cannot establish observer
causation. This checkpoint prepares one new observer-on then observer-off pair.

The exact final v2 executable, native archive, Core, source-detail overlays,
view/commands and classifier are reused. Copied executable bytes and execution
permissions verify; no Go, Rust or SDK rebuild occurs. Only control order changes.
The original controller and completed evidence remain untouched. CPU-upload/GPU
observers stay on in both cases; the final argument alone toggles the display clock.
Both runs keep 128 warmup and 120 measured requests, 120 ms spacing and the same
50 ms held tail outside the cohort. Every startup/callback/presentation remains
accounted for by the unchanged strict classifier.

Five focused orchestration methods cover literal command order, seven failure
subcases, stop before the second run, local and cross-case configuration drift,
and retention of failed observations. Syntax, fresh-scope overwrite refusal,
all compiled source/module/overlay/consumer pins and all 8,321 SDK files verify.
Existing native ownership, JSON, pixel and eight-method classifier checks are
reused through exact source/binary identity, rather than repeated unchanged builds.
See [preparation checks](preparation-checks.json).

The [protocol](../../../tools/macos-calibration/displayclock/REVERSED.md) permits
one newly approved quiet scope after review and commit: at most 100 seconds
overall and 40 seconds per native case, using the existing exclusive lease and
owned executor. Stop on first invalidity, retain the attempt and do not retry.
No native run has started yet. Internal review and Prism with Anthropic Claude
Sonnet 5.5 precede execution; actual results receive a separate review and commit.

Report run statistics separately and calculate on-minus-off p95 by mode,
independent of execution order. Keep the prior pair visible without pooling
samples, adding quantiles or joining timestamps across processes. Reversal alone
does not establish calibrated observer overhead, optical scanout/latch behavior,
compositor-only cause or ordinary release qualification. Features, original
backend, SDK/app, allocation/cache/render/pacing defaults and release gate remain
unchanged. No permanent display link or production fix is adopted; M3–M7 open.

Preparation review is complete: Prism (Anthropic Claude Sonnet 5.5) covered all
nine changed files without exclusions or truncation. Its two medium concerns
about cleanup and the overall deadline are resolved by the inherited guards,
confirmed with an actual headless admission/stop test: expired Host commands and
probes refuse launch; a five-second child stops and retires in 0.234 seconds
when the overall stop fires at 0.2 seconds. `Host.finish` asserts integrity;
resource retirement is in the owned executor, lease context and runner finally.
See [guard audit](host-guard-audit.json) and the retained review dispositions.
No desktop use occurred during these checks.
