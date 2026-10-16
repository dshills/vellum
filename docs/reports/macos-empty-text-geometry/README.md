# Empty native text geometry and isolated accessibility acceptance

Native correctness checks pass on committed `cdc0111`, which includes the
`8dadc48` empty-text-geometry optimization. The 80-process ordinary comparison
against `1ba954c` establishes no consistent latency gain. Both arms still miss
the unchanged UI-work targets: p95 ≤4 ms and p99 ≤6 ms. All 40 runs per arm miss
p95; 37 per arm miss p99. This checkpoint does not establish no regression.

The bridge omits nil/empty text-range geometry updates, which are native no-ops.
Nonempty updates and semantic retirement retain their previous paths. Both
standalone accessibility acceptance entry points now disable the process-local
geometry saver before creating their app, and report strict `true` while it is
disabled. The earlier 83-process readiness plan was refused before launch when
review found these two paths could restore/save user window geometry. That
unlaunched refusal is retained independently from the corrected execution.

The corrected native queue completed all 83 processes and 166 run/validator
commands, from 2026-10-16 23:13:13 to 23:27:19 UTC. Three Cocoa selector tests
exercise UTF-16, stale geometry and refused/accepted values. Main/modal controls
exercise native actions, focus, Tab/Escape, paging and workbench namespaces.
The million-row grid checks at least 400 clipped cells, 2,268 overlapping-object
reuse observations, 459 replacement retirements, native selection and closed
attachment refusal. These are automated native routing checks; physical pointer,
screen-reader and human input acceptance remain separate.

Ordinary timing uses eight CPU/GPU-requested × fixed/variable × 10k/million
cases, five fresh interleaved process pairs each, 128 guarded warmups and 120
measured frames per process. Profiling and detailed input/source instrumentation
are disabled. There are 9,600 measured frames and 9,840 ordinary JSONL records
including setup/ready/end. All 4,800 baseline/candidate frame pairs have identical
complete geometry; another 4,800 public row-count comparisons match. All raw
samples, eleven metric distributions, allocation observations and per-run
variation are retained in [the comparison](records/comparison.json).

Ranges below cover all five per-run percentiles, in milliseconds.

| Case | Baseline p95 | Candidate p95 | Candidate p99 |
| --- | ---: | ---: | ---: |
| cpu-fixed-10000 | 7.838–8.232 | 7.837–8.279 | 8.131–8.910 |
| cpu-fixed-1000000 | 7.580–7.795 | 7.489–7.769 | 7.842–8.130 |
| cpu-variable-10000 | 7.469–9.872 | 7.874–8.192 | 8.140–8.739 |
| cpu-variable-1000000 | 7.286–8.078 | 7.296–9.329 | 7.641–12.065 |
| gpu-fixed-10000 | 6.029–6.286 | 5.962–6.238 | 6.195–6.638 |
| gpu-fixed-1000000 | 5.466–5.872 | 5.552–11.938 | 5.730–13.499 |
| gpu-variable-10000 | 5.984–6.897 | 5.938–6.734 | 6.289–8.811 |
| gpu-variable-1000000 | 5.984–6.398 | 5.958–6.410 | 6.467–7.133 |

The candidate GPU-requested fixed-million first pair reaches p95 11.937835 ms
and p99 13.498958 ms, against baseline p95 5.872041 ms. Its other four candidate
p95s are 5.552–6.007 ms. The bad run is included, with cause unresolved; neither
pooling nor a later good observation removes it. Multiple UI phases increase,
and most slow frames have no reported GC cycle, so GC alone is not an explanation.

Each owned process was reaped with its process group absent, capture logs
complete, and source/binary bindings stable. That is process-retirement evidence,
not native-service inventory or an OS preferences audit. GPU mode is a request;
fallback, executed GPU work, pixels and presentation are not qualified here.
Process-wide allocations do not attribute retained, native or GPU memory.
Full P4/P6/P9/P11, idle/overnight, physical displays/input, human IME/VoiceOver,
interoperability and distribution acceptance remain open.

Focused/full adapter race tests and native vet cover `8dadc48`; the isolation
correction additionally passes loaded-grid offscreen race and both affected
native vets. Each source change received internal review and one focused Prism
review using Anthropic Claude Sonnet 5.5, with no retained/discarded/suppressed
findings. The completed native evidence received independent internal review.
The source fixes were fast-forwarded into the clean canonical checkout after
that review. This documentation does not transfer native qualification to later
source changes.

The [bounded evidence archive](records/evidence.tar.xz) and
[index](records/archive-index.json) preserve producers, build/source vectors,
raw output, reviews, the unlaunched refusal and integration receipt. Seven compiled
executables are excluded by hash and size. Historical absolute paths are labels;
inspection does not require executing archived controllers or extracting them
over an existing checkout. The archive is evidence, not a redistributable SDK.
