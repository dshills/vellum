# Pending native accessibility publication checkpoint

Vellum can publish a pending managed grid by copying its existing native node
dictionaries when the complete prepared publication proves unchanged outside the
pending fields. The full canonical payload accompanies every patch and supplies
the authoritative fallback. See the
[contract](../../../specs/accessibility.md#pending-only-native-publication).
Public Go APIs and application interaction behavior remain unchanged.

## Source and checks

The implementation is committed as `e4d398a24fff8f9506ea837c206be7e6416aba7f`. Changed owned publications
retain one detached immutable wire baseline, without handlers or activation
guards. Exact membership, order, scalar types, pointer values and every current
wire field constrain reuse. Host labels/help, values, geometry, identity changes,
ordinary controls and ready sibling changes require full publication. Native
preflight checks the preceding epoch and all affected scope hosts before freezing
and installing every replacement dictionary. Completion uses full admission.

The final adapter package's complete offscreen race suite, module vet and both
native fixture builds pass. Earlier full-module race checks pass at the
pre-hardening source reviewed by Prism. Foundation comparisons reproduce full
publication's node dictionaries and identity relationships through pending,
repeated target, completion, rejected updates, removal and closure. Controls
cover malformed/stale patches, late scope refusal, full fallback after native
rejection, exact large uint64 epochs and overflow. A Go compiler overlay removing
only the RowCount comparison builds and fails the intended field-audit assertion.

Internal review found three issues that are now covered explicitly:

- A pending completed-subtree cache could restore custom/mutable values into the
  new baseline. Both raw and prepared values are now checked before retention.
- The legacy full counter can wrap. A permanent exhaustion latch disables new
  baseline/patch reuse before wrap can alias an old native epoch; the full path's
  existing behavior remains unchanged. A seeded actual-publication test covers
  wrap and repeated ready/pending updates after it.
- Full commit prunes geometry and text-geometry entries absent from membership.
  Pending commit now shares that cleanup. Differential cache/node/identity checks
  cover repeated orphan entries, current Unicode text geometry and atomic refusal.
  The first native compiler-overlay mutation did not fail and is retained as an
  unsuccessful control. An isolated module copy removing only pending cleanup
  compiles and fails the intended ancillary-state assertion. Original source
  bytes remain unchanged during both controls.

Prism reviewed all eight changed source/spec files using Anthropic
`claude-sonnet-5-5`, in one complete 39,815-byte chunk and one provider call. Its
raw medium finding and exit status 1 are retained. The ordinary native-refusal
scenario was independently refuted using serialized unique publication epochs;
internal review identified and corrected the formal wrap boundary separately.
The final exhaustion/cache corrections are internally reviewed and checked after
that provider run. A second provider review is not claimed. The evidence retains
the original provider payload, original endpoint bytes, every correction and
[the final review](result.json).

## Native correctness

The normal native accessibility smoke passes its reachable Cocoa tree, native
window identity, selectors, controlled actions/disabled refusal, focus/Tab,
virtual paging, outline, modal and dock namespace checks. Its final snapshot has
27 reachable elements. The dedicated stored-million-row Cocoa grid check passes
the 400-visible-cell minimum, native press admission, 2,268 overlap identity-reuse
observations, 459 replacement retirements and attachment-query refusal at close.
Reuse observations do not count distinct objects.

The pre-hardening 22-case campaign and a separate final six-case sanity campaign
finish with complete bounded logs and clean owned-group
retirement. Source, prepared dependencies and reviewed inputs remain stable
through each execution; one shared lease serializes the runs. The fixtures
disable their geometry persistence and report no preference changes. There are
no physical pointer/key actions or VoiceOver checks in this campaign.

Native runs use the exact frozen uncommitted candidate, subsequently committed
unchanged. Runtime metadata correctly records its `vcs.modified=true`; the
baseline is a clean `c5045bc` executable. Both actual binary digests, metadata and
complete source/prepared manifests are retained. A later commit identity is not
retroactively assigned to those executables.

## Ordinary frame measurements

Five fresh alternating AB/BA pairs run the fixed-million-row, stored numeric
fixture in each rendering mode, with 128 warmup and 120 measured frames per
process. There are 20 ordinary processes, 2,400 validated completed tickets and
1,200 completely identical paired geometry records. No profiler, diagnostic
detail mode or owned test/build job overlaps these selected timings. This does
not establish continuous global host isolation.

| Mode | Baseline median process p95 | Candidate median process p95 | Runs over 4 ms p95, baseline / candidate | Runs over 6 ms p99, baseline / candidate |
| --- | ---: | ---: | ---: | ---: |
| CPU renderer | 7.460498 ms | 7.038084 ms | 5 / 5 | 5 / 5 |
| GPU requested | 5.336209 ms | 5.317541 ms | 5 / 5 | 1 / 2 |

In both modes two of five paired candidate/base p95 ratios exceed one. These
results establish neither consistent speedup nor broad no-regression, and the
ordinary 4/6 ms targets remain unmet. This cohort covers the fixed-million case;
variable rows, other row counts and the professional workspace are separate
acceptance workloads. A GPU request does not prove actual GPU execution.

Those five pairs precede the final counter-exhaustion and cache-parity corrections. The final source
repeats native AX/grid acceptance and one fresh base/candidate pair per rendering
mode, with 480 further tickets and 240 full geometry pairs. The CPU base/candidate p95 values are 6.877835/7.415833 ms; GPU-requested
values are 5.406873/5.451626 ms. All four runs miss the 4 ms p95 target.
CPU p99 exceeds 6 ms in both runs; GPU-requested p99 exceeds it only in
the baseline run.
This four-process sanity cohort is kept separate and does not establish sustained
performance or broad no-regression.

## Earlier native stack diagnostic

Before implementation, two separate unchanged `c5045bc` million-row fixtures
capture eight-second `/usr/bin/sample` wall stacks against their exact owned
PIDs, after warmup. Each completes 1,024 measured tickets. Independent analysis
counts each inclusive main-thread category once per stack branch:

| Mode | Main-thread observations | AX update | Commit | Validator | JSON reader |
| --- | ---: | ---: | ---: | ---: | ---: |
| CPU renderer | 6,002 | 1,443 | 1,005 | 545 | 426 |
| GPU requested | 6,009 | 1,432 | 1,010 | 546 | 414 |

Categories overlap and cannot be added. These wall stacks locate native parsing,
validation and commit work; they do not attribute CPU time, distinguish pending
frequency or prove causal savings. Sampled timings do not qualify ordinary
latency. The stack diagnostic's whole-process footprint includes the stored
dataset and cannot qualify the empty-app, native or GPU memory budgets.

## Evidence and remaining gates

[Result summary](result.json) binds the
[implementation/check/native/provider archive](evidence.tar.gz) and its
[member manifest](evidence.manifest.json), plus the separate
[baseline stack archive](baseline-stacks.tar.gz) and
[member manifest](baseline-stacks.manifest.json). The archives retain producers,
raw JSONL/logs/stacks, source/build identities, reviews and failed control output;
they exclude executables. Source and result reviews describe their exact scope.

Native fast-path frequency and causal native CPU/memory savings remain
unmeasured. Inert fixtures do not observe asynchronous notification delivery.
Physical placement/pointer/keyboard/display workflows, external AX clients,
human IME/VoiceOver, full P4/P6/P9/P11, sustained/native/GPU memory, idle/overnight,
platform interop and Mac distribution remain open. The earlier placement app
binding timeout remains failed with its cause unassigned; this campaign did not
retry CUA or qualify that workflow. The Mac completion goal remains active.
