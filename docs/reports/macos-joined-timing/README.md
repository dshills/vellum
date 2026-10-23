# Joined raster/upload/scheduling: native result

The approved single capture passes complete joined accounting. Its instrumented
receipt-to-actual-presentation p95 is **35.353 ms**, with17of120measured observations
over33.3ms. The largest interval on the actual p95-ranked input occurs **after
drawable acquisition**. This narrows the next boundary; it is not a renderer repair,
ordinary-build qualification, or permission to change the release gate.

## The same input at the p95 rank

Nearest rank114of120 selects accepted revision228, scope1, opaque tag458,
generation231, presentation sequence232 and drawable237. The following are
contiguous parts of that one input's receipt-to-shown span, not phase percentiles:

| Part of the selected input | Elapsed |
| --- | ---: |
| Raster bracket | 0.332 ms |
| Native upload function | 1.460 ms |
| Other time before drawable acquisition | 0.320 ms |
| Drawable acquisition to actual presentation | 33.241 ms |
| **Receipt to actual presentation** | **35.353 ms** |

“Other” combines receipt→raster start, raster end→upload start and upload end→
acquisition from this input. Full raw clocks and allninewake-to-shown segments
remain in [result-summary.json](result-summary.json) and indexed raw evidence.
For context, separately computed population p95values are0.844ms for the raster
bracket and3.104ms for native upload; these must not be added to each other or
to the selected input's values.

Acquisition-to-shown includes remaining encoding, submission, GPU execution and
compositor/display scheduling. **This capture does not separate those phases.**
Calling the whole interval GPU or compositor delay would exceed the evidence.
No hot-microbenchmark subtraction, inferred raster timestamp, or counterfactual
speedup is used. A future diagnostic should join backend submit/GPU/presentation
boundaries to these same input/drawable identities before proposing policy changes.

## Accounting and verification

All248requests (128warmup/120measured) uniquely join one CPU scene, raster bracket,
native upload and positive armed presentation. There are256raster records and254
native upload records, including8and6configured startup records respectively.
Every unused record precedes first receipt; startup/failure data are not discarded.
Sparse tags remain opaque. Allninecontiguous native-clock intervals close exactly
at floating precision in this capture; the maximum per-input partition error is0ns.
An independent raw-clock audit recomputes0.0ns signed residual for both wake and
receipt partitions on all248inputs. The classifier’s1µs refusal tolerance is
separate from that observed value. This is arithmetic consistency, not clock
accuracy or observer calibration.

Wake/check/post-check and raster/upload/acquisition/presentation use native host
time. Nominal request target/lateness, FIFO queue duration, nested Go work and
receipt clock-read widths remain separate; there is no guessed target-clock
conversion. Read brackets range209–15,042ns across all248requests, without being
interpreted as total observer overhead. No phase quantiles are summed.

The run takes37.695s overall and31.273s in the native command, within60/40s bounds.
Allnineowned command groups retire; native stderr is empty. Independent replay
matches both sealed validation copies, complete logs/pins, original-policy CPU
observer archive, unchanged configuration/persistent geometry and all8,321SDKfiles.
Process retirement does not claim whole-GPU allocation balance. No physical input
or actual displayed-byte readback is performed.

The [protocol](../../../tools/macos-calibration/joined/PLAN.md),
[renderer patch](renderer.patch), [fixture patch](native-main.patch),
[preparation checks](preparation-checks.json) and indexed preparation/result archives
retain the basis. Source checkpoint is `ffb7aea`. Five focused recorder race methods,
the original routed full-pixel/state test with the clock enabled, and five classifier
methods/23refusals pass. Recipe reproduction is exact. Superseded v1 preparation is
retained and was never run natively; only the reviewed final v2 binary ran once.
No broad campaign or unchanged native retry occurred.

Prism Anthropic Claude Sonnet5.5 preparation review led to explicit Begin publication
state and corrected observer-on usage. Its conditional clock-panic concern is
dispositioned against the fixed native-clock contract: unexpected clock failure
aborts/refuses the attempt; incomplete pending evidence must not be erased. Source
and result reviews and their dispositions are retained in the archives.

Features, SDK/application, original allocation/cache and shipping render/pacing
defaults stay frozen; no pool/cache candidate or production repair is adopted.
Extra clocks and journal locks perturb execution, and one case does not establish
repeatability or ordinary-build equivalence. The33.3ms release gate is unchanged;
M3–M7 remain open. The desktop capture is finished and no native job remains active.
