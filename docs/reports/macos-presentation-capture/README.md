# Stored grid/tree presentation capture

This checkpoint connects both standalone consumers to the optional public native
presentation API. `-perf-presentation NEWFILE` exports bounded pre-close/post-close
journals and a strict validator joins routed receipts, CPU generations and actual
presentation records. [The contract](../../../specs/presentation-capture.md) defines
flags, frozen acquisition prefix, explicit tail/loss accounting, worker retirement,
count/file/time limits and external execution provenance.

This is implementation and headless evidence. No native window, real drawable,
live callback, display interval or real presentation journal was produced here.
The existing scroll loops are serial. Successful synthetic JSON validation cannot
qualify the owner's measured 60 Hz gate, ordinary timings or physical input latency.
Every capture-enabled performance event explicitly carries diagnostic/noncadence
labels; the disabled path retains its ordinary output.

## Verification

Go 1.25.6 / GOMAXPROCS=2 checks used the existing stage88 prepared sources and retained
native archive 786aaad8…2ecec4883. IP networking and source/SDK writes were denied to
build/test children. Each owned child has a bounded timeout/output, terminal
result and observed process-group absence. No application binary was launched.

- Final Go revision: 38 grid and 25 tree top-level methods pass with race detection
  (25.565s and 16.649s owned commands); both default offscreen vets pass.
- Actual optimized Darwin ARM64 binaries compile (7.013s and 6.975s). They are
  precommit outputs, retained unused, and are not admitted for later native launch.
- Final Python revision: 10 test methods pass (2.434s), including valid grid/tree
  fixtures, rejection/identity/clock/loss controls, paired composition mutations,
  actual CLI argument handling, duplicate/nonfinite JSON and nonblocking refusal
  of directory/symlink/FIFO inputs. These are synthetic controls, not Metal proof.
- Helpers test exclusive creation, failed evidence/idempotence, an unjoined worker,
  fixed prefix versus explicit pending tail, worker acknowledgement, executable
  bounds/hash, disabled output and incompatible modes. Helper sources are identical
  in the two independently consumable example modules.

Internal review found that the first validator allowed a composition ordinal to
be reused with a changed tag/generation/armed stamp. The final validator enforces
nondecreasing ordinals and immutable per-composition identity. Paired pre/post
mutations fail specifically at that rule, while repeated acquisitions with the
same stamp pass. Initial source and all intermediate passing receipts remain
retained; they are not relabeled as final-revision checks.

Internal review also found a pre-Run setup failure could populate a premature
post-close snapshot. Capture now records separate `WorkerRetired` and
`RuntimeReturned` acknowledgements and writes `AfterClose` only when both hold.
A failed setup retains evidence with no runtime/teardown claim; regression tests
cover that path and the validator refuses a missing runtime acknowledgement.

Prism reviewed the implementation once with Anthropic Claude Sonnet 5.5 and found
that a missing final armed acquisition could leave a captured journal for later
offline rejection. The focused correction requires exact completed CPU/receipt
tag parity and one matching final armed acquisition per tag in the frozen prefix.
Pending callbacks for those existing acquisitions may drain; later acquisitions
cannot repair the prefix. The offline validator already refused the incomplete
case. The correction and regression tests receive internal review; no second
provider call is used. Original reviewed source and provider output are retained.

The first focused regression fixture failed compilation because it initialized
promoted Frame fields directly. The embedded literal was corrected; the original
four failed race/vet commands remain recorded. Final race/vet checks pass, and
the successful native builds use unchanged production source.

Actual native build metadata inspection showed external-modfile builds can lack
VCS fields. Capture now hashes the executable outside measured work and the
validator requires its expected hash plus external source/build/launch binding.
Present VCS fields must still match clean source. Neither a hash nor self-reported
metadata alone authenticates a native execution.

[Evidence summaries](evidence.json) bind receipts, raw log hashes, method names,
build output identities and process retirement. Full source-pinned receipts,
source revisions and logs remain in development checkpoint
`macos-presentation-capture-7cbbb73`; this summary does not reconstruct those files.
Existing AX deprecation and duplicate-objc linker warnings remain unchanged.

## Prospective startup observation correction

The shared capture and strict validator now distinguish terminal skipped native
callbacks acquired before the first input as startup observations only when they
are in the same scope with no armed/tag/generation identity and zero presentation
time. Apple's drawable contract permits zero time for unpresented/skipped frames;
such callbacks do not prove display. Full frozen prefixes and all raw states stay
retained. New captures explicitly count prefix records, genuine presentations and
startup skips separately. Existing failed journals remain failed and refused.

This source correction changes diagnostic population accounting, not rendering,
latency or pacing behavior. Positive synthetic serial/paced grid/tree controls
compare complete measured populations against the original fixture; negative
controls cover the exact receipt boundary, wrong identities, other lost states,
missing final acquisitions, corrupted counts and immutable failed files. Capture
helper controls preserve the skipped record and the frozen CPU/receipt cohort.
These controls are headless evidence only. Actual callback/cadence qualification
and a fresh externally bound native run remain separate.

## Remaining work

Qualify actual Metal callback/clock/loss identity and native teardown first. Then
use a continuously changing, representative workload for five measured pacing
runs with retained display/configuration/provenance. The serial validator never
reports a cadence pass. Other input boundaries, physical input, full resources/
endurance, SDK/rights and final Mac audit remain open. This checkpoint changes no
prepared native sources or private SDK and closes no full Mac milestone.
