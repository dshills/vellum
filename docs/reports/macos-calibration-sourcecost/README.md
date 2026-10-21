# Matched-rectangle source-cost diagnostic preparation

The first constructor-only diagnostic is sealed and failed at its first input:
Core's grouped detailed tracer lacked its per-ticket source-detail container.
Its native process exits1 with "source diagnostic composition details unavailable"
after1.237 seconds; no complete cohort or timing acceptance exists. The raw data,
original36ad689 source snapshot and [refusal](refusal-analysis.json) are retained.

Stage0106 initializes that container in BeginGroup, mirroring the existing Begin
path only when source details are enabled. Ordinary traces remain unchanged.
The two new Core regression methods cover prefix/final detached data, shared32
record cap/33 refusal, cancellation with live composition and ordinary isolation.
They fail by assertions on original105 and pass with the existing group/source
trace tests under race detection. All34 preparation checks and a full fresh
prepare pass. Initial test assumptions about last-stage105/patch-only stages were
corrected and failed outputs retained. This repairs diagnostics, not latency.

The fresh [one-process protocol](../../../tools/macos-calibration/sourcecost/PLAN.md)
uses three exact private compiler-overlay locations: existing detailed constructor,
existing detached result field, and the identical grouped initialization repair.
Original SDK Go/native files are untouched, original786aa native archive is reused,
and no renderer/pacing/default/public API or release policy changes. Five wrapper
methods with six nested negative subcases pass; earlier view/routing/pixel and
existing scalar-validator checks are reused. Extra clocks perturb elapsed time.

Native build/overlay identities are in [original](build-identity.json) and
[corrected](build-identity-v2.json) records. The actual stage106 patch is owned
under native/patches/groupsourcedetails. Internal/Prism review and source commit
precede a fresh single-process60-second scope. No ordinary latency qualification
or Mac milestone closes from this diagnostic preparation.
