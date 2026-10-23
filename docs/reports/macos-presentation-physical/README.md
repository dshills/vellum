# Physical native-source phase-control preparation

This repairs the private diagnostic build after the [first phase-control
attempt](../macos-presentation-phase/README.md) failed at framework startup.
No new native application has run. The earlier failure remains sealed and is
not reclassified; no matched native/framework comparison exists yet.

The previous linked acquisition method lacked the supplied pending hooks.
The new [recipe](../../../tools/macos-calibration/presentationphase/physical_prepare.py)
copies6,031Core files to a fresh private tree and verifies their bytes before
applying exactly three known replacements: presentation_darwin.m/.go and the
already-used frame_trace.go. Original Core and all8,321SDK files stay unchanged.
The module replacement targets that physical tree, with no native Core source
left in the compiler overlay. Existing SDK session/measurement Go overlays stay
as reviewed. The native archive is reused unchanged; no Rust or SDK rebuild.

An [actual linked-method audit](linked-hook-audit.json) now proves calls to
pc_acquire_begin,pc_acquired,pc_unavailable and the original presentation acquire
observer. This checks the compiled ARM64 method, rather than merely finding
unused hook symbols. A headless test of the actual physical layer factory
verifies the subclass and scalar observer stamp publication/revocation without
NSApplication, a window or drawable acquisition. Those checks do not by themselves
prove successful runtime startup or presentation.

The private framework observer additionally records the actual Metal/view layer
class, so any startup failure has more diagnostic context. It does not alter the
layer or presentation policy. The earlier missing presentation feedback is not
fully explained solely by the absent pending calls; the next capture must still
establish startup readiness and complete phase/pending attribution.

The standalone native executable is copied byte-for-byte from the reviewed
first attempt:44c85638e4fb3415fd979869cd4d25806d8fe7626660eb9a3b9381505aedb91d.
Its source/header/shared observer bytes verify equal. The new private framework
binary is6825d90efe89f75ce1361c8bf75542ff2866c2869061b5ae95405a576646466a.
[Checks](preparation-checks.json) verify owned build/test groups, every log and
compiled source pin, original/source-copy vectors, exact consumer/Core/module/
overlay reproduction and fresh-output refusal. Existing eight classifier/controller
methods, scalar bounds/concurrency/JSON controls and full routed pixels are reused
through unchanged source identity. Only the physical native integration has changed.

The original [phase protocol](../../../tools/macos-calibration/presentationphase/PLAN.md),
classifier, populations, matching criteria, same100second overall/40second per
process bound and fail-first/seal rules are unchanged. Native reference then
framework,128warmup plus120measured with30per phase, four fixed phase fractions;
actual receipt error0..2ms and per-phase p50/p95 error difference<=1ms determine
comparability, not release acceptance. Pending means acquisition-awaiting-callback;
no Metal-pool, reference-lifetime, compositor-depth or optical claim follows.

Internal review and the complete11file Prism review using Anthropic Claude
Sonnet5.5 are complete, with no exclusions/truncation. Its changed-set concern
is source-dispositioned: the physical trace file equals the intended overlay,
missing input refuses, and currentall3changes/fullvectors/reproduction verify.
No actionable findings remain; see the [review receipt](review-completion.json).
A fresh desktop approval is pending for the corrected capture; the first quiet
scope was consumed. Preparation does not launch a native app. Existing design,
features, SDK/app, allocation/cache/render/presentation/pacing defaults and33.3ms
release gate stay unchanged. No permanent display link or production repair is
adopted; M3–M7 remain open. Native results or rejection require their own review
and commit. Evidence excludes binaries, native library, full private Core/SDK/cache.
