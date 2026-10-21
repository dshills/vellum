# Actual backend timing observer result

The corrected campaign completes three valid private-backend observer off/on
pairs. **All six p95 response times remain above the unchanged 33.3 ms gate.**
The measured render GPU span is small; these data do not establish a shipping
framework cause, observer neutrality, binary equivalence, or an accepted repair.

| Pair / order | Observer off p95 | Observer on p95 | On minus off p95 |
|---|---:|---:|---:|
| 1: off/on | 36.547 ms | 35.950 ms | -0.597 ms |
| 2: on/off | 36.421 ms | 35.805 ms | -0.617 ms |
| 3: off/on | 36.719 ms | 35.481 ms | -1.239 ms |

These are differences of per-run p95 values, not quantiles of per-frame
differences. Observer-on has lower p95 in these three pairs; that does not prove
zero observer overhead or a latency improvement to adopt. Both modes run the same
private diagnostic archive and binary, not the retained original SDK binary.
The following original-versus-derivative control is required before using this
instrumented timing to choose a shipping-path repair.

For the three observer-on runs, receipt-to-first-render-submit p95 is
4.086/4.158/4.226 ms; first submit to earliest render GPU start p95 is
0.793/0.795/0.824 ms. The min-start to max-end GPU span across exactly five
render command buffers per measured frame has p95 0.284/0.276/0.278 ms.
That span includes inter-buffer gaps; it is not a sum of GPU busy durations.
Last render GPU completion to actual presentation p95 is 31.923/32.033/31.400 ms.
Present requests are asynchronous: GPU-end-to-request p95 is negative
−0.675/−0.670/−0.585 ms. Callback timestamps separately record CPU observations.
Do not add these quantiles or infer a universal compositor floor.

Tail data remain included: observer-off maxima 45.927/51.886/50.953 ms;
observer-on maxima 39.740/37.221/38.090 ms. The third on-run GPU span has a
5.968 ms maximum despite its low p95. No tail, failed scope or startup feedback
is removed to improve the numbers.

The 21 October 2026 campaign runs 21:09:54.517–21:13:11.251 UTC
(16:09:54–16:13:11 Chicago), 196.734 seconds within 240 seconds. Six cases retain
768 warmup and 720 measured accepted states, including 360 measured states with
actual GPU clocks. All six native and 48 probe process groups terminate/retire;
stdout/stderr hashes, source/binary/library/match identities, unchanged SDK
8,321-file inventory, configuration and persistent geometry are verified.

All 252 admitted public presentations in each on-run match exact Metal drawable
IDs and timestamps. Six lower-ID startup rows per run precede public admission,
with requests before first public acquisition and positive feedback before the
first cohort receipt. All 18 prefix rows remain in raw output: five zero and
thirteen positive feedback records, IDs 0–5 per run. Every admitted/cohort record
is matched; callback pending is zero, buffer status is completed and no overflow
occurs. Fixed trace storage never evicts rows. Measured membership uses explicit
scope/tag/generation keys; armed setup is not assumed to be measured input.

## Preparation, refused scopes and review

The diagnostic is a private source-built derivative of pinned wgpu-native and
wgpu-hal 29.0.3. Only submit/present observation and a private ABI are added;
no archive is installed in the retained SDK, framework or app. Source, dependency
package identity, build logs/warnings and binaries are bound in the
[original](build-identity.json) and [revised](build-identity-v2.json) identities.
The same retained SDK and byte-identical view/routing/pixel regression sources
justify reuse of their prior verification. Fifteen backend plus18 parent
classifier methods pass. Focused internal and Prism Anthropic Claude Sonnet 5.5
reviews precede each source checkpoint; final membership review has no findings.
Source checkpoints are f9462e8, c18f74f and 66df735.

Three earlier attempts remain sealed and rejected, with no accepted pair:

- f9462e8 observes 258 backend presents versus252 public records. All252 public
  timestamps match; six earlier startup requests (two zero/four positive feedback)
  precede first public acquisition. Extra actual IDs were unrecorded, never
  retrofitted. [First refusal](refusal-analysis.json) and original source snapshots
  remain; two native plus16 probe groups retire.
- c18f74f passes ID/prefix matching but its checker raises KeyError `(1,1)` on an
  armed startup setup frame without measured receipt. Explicit membership fixes
  this and sparse tags. [Second refusal](refusal-analysis-2.json) and its source
  stay rejected. A corrected-checker posthoc raw regression is not acceptance.
  Two native plus16 probe groups retire.
- The first 66df735 binary copy loses its executable bit. `env` exits126 before
  running the native program; no timing population exists. The fresh valid scope
  uses `copy2`, exact binary hash, mode0755 and execute-access verification. All
  nine wrapper/probe groups retire; this local copy correction changes no source
  and reuses prior tests/review rather than rerunning them.

The sparse 120 ms fixture does not qualify 60 Hz changing workloads, professional
workflows, physical input/pixel readback, resize, endurance, allocation balance,
or public distribution. M3–M7 remain open, features remain frozen, and the 33.3 ms
receipt-to-actual-presentation gate is unchanged. Next compare the retained
original framework binary with this private observer-off binary under a bounded
matched control; no GPU/backend rewrite is selected from these data.

See the [protocol](../../../tools/macos-calibration/backend/PLAN.md),
[exact summary](summary.json) and indexed compressed evidence. Executables,
native libraries, SDK/cache/target trees, credentials and prior archives are
excluded from archives. Result validation and Prism disposition are retained.
