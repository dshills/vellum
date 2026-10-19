# Paced grid/tree presentation workload

Both standalone stored-data consumers now expose `-perf-60hz` with a new raw
presentation journal. The mode runs at immutable rational 60 Hz deadlines and
retains every changing state, late backlog and unstarted slot on failure. It
keeps serial/ordinary branches and the existing public routing/recorder APIs.
See the [contract](../../../specs/paced-presentation.md) for runnable commands,
measurement boundaries, bounds and qualification limits.

The timed cohort has no per-event stdout, memory sampling or profiling. Existing
CPU polling, owner probes, clipped geometry hashing and source/callback checks
remain included observer work. Compact detached view summaries retain the
expected nonzero trajectory, current generation and frame identity. The fixed
planned end, actual last iteration and idle held tail remain distinct. Final
capture keeps the previously reviewed frozen-prefix and retirement protocol.

The strict version 2 analyzer uses one final armed actual presentation per
scheduled state in request order. N states yield N-1 intervals; intermediate
draws do not dilute gaps. Equal/reordered final times remain anomalies, and
late/failed/unstarted work cannot disappear through a shortened successful
prefix. Numeric pacing and latency predicates are separate from authenticity
and full release qualification. Same-clock native latency remains direct;
cross-clock endpoint bounds use enclosing receipt call brackets and explicit
origin/rate assumptions and rounding uncertainty.

## Checks and internal corrections

The final source passes 45 grid and 31 tree top-level race methods, 18 Python
methods, both offscreen vets and actual optimized Darwin ARM64 builds. Owned
command times are 23.494s/17.875s for race, 2.475s for Python and 7.652s/6.646s for
native builds. Network access and source/SDK writes are denied to test/build
children; outputs/timeouts are bounded and every owned process group retires.
Existing AX deprecation and duplicate-objc linker warnings remain unchanged.
No compiled application is launched.

Schedule tests use deterministic clocks for rational deadlines, late backlog,
cancellation/work failure, unstarted slots, early clocks/reentry and detached/
bounded evidence. Schema tests cover both formats, trajectory/geometry identity,
slot/tag omissions, impossible clock brackets, duplicate/reordered final states,
long-gap dilution by intermediate draws and explicit final overrun. Synthetic
fixtures remain synthetic; these are not measurements of rendering performance.

Internal review caught a mismatch before native execution: the first Python
trajectory used horizontal phases 0/8/16/24, while the Go consumer uses 7/15/23/31.
The first fixture called that same Python helper, concealing the mismatch despite
passing tests. The final fixture has an independent literal trajectory and Go/
Python checks assert eight published cumulative checkpoints plus the initial
seven vertical-only steps. All original passing receipts/source remain retained
as intermediate evidence, not final trajectory qualification.

Review also required presentation endpoint rounding in addition to receipt-origin
uncertainty. The final bounds conservatively include both seconds-to-nanoseconds
conversion and independent first/final float rounding. Local pinned Go runtime
and QuartzCore source declarations support the common-rate mach_absolute_time
model, with hashes retained; actual clock delivery still needs live proof.
The two correction paths receive affected race/vet/Python checks. Unchanged Go
production bytes retain the existing native build/tree checks. Version 1 remains
accepted without paced-only fields. Shared helpers/tests are byte-identical in
both independently consumable examples.

[Evidence summaries](evidence.json) retain exact receipt/log hashes, method lists,
unused build identities, clock-source hashes and terminal process observations.
Full raw/source-pinned logs, original source and intermediate records stay in
development checkpoint `macos-paced-presentation-cc2ec6a`; these summaries do not
reconstruct them. Precommit build outputs stay unused and are not admitted for a
later live campaign.

## Remaining qualification

No real native drawable, callback, timing journal or continuous pacing result is
produced by this checkpoint. The native callback/clock/loss/lifecycle proof,
affected stage88 renewals and five measured runs on an available quiet desktop
remain required. External source/build/archive/binary/launch/workload/display/
power provenance remains mandatory. Grid/tree do not replace composed workspace,
text/editing, canvas or resize release workloads. Physical/human verification,
M4-M7 and Mac completion remain open. The backend, prepared sources, retained
SDK and release gates are unchanged.
