# Mac UI element expansion

The public framework now includes controlled `ui.Switch` and `ui.ToggleButton`,
read-only `ui.ProgressBar`, and decorative `ui.Separator`. Switches and toggles
use accepted Checked values and typed owner Change requests. Progress uses a
finite 0..1 Number fraction. Progress/separators accept horizontal/default or
vertical Orientation, reject input callbacks, and stay outside keyboard focus.
Separators are omitted from the accessible hierarchy; progress publishes read-only
fraction/bounds and the Cocoa progress role. Both boolean additions publish
checkable roles and boolean press/set actions.

Retained synchronization preserves identity, latest callbacks, disabled ancestry
and queued retirement refusal. Explicit size/grow styles run after native final
defaults and restore those defaults when removed. A native boolean mismatch
requests restyling before resetting Checked, so an application rejection cannot
leave a toggle's selected background visible. The independent gallery demonstrates
all four additions; [the public catalog](../../controls.md) lists available kinds.
The [controls spec](../../../specs/controls.md) describes their contracts and
prioritized future vocabulary. Grouped radio/segmented choices and indeterminate
progress are separate work.

## Verification

Final portable and complete offscreen adapter race suites, adapter vet and desktop
gallery compilation pass. Focused control/gallery tests pass three race repetitions.
They exercise payloads, retained callbacks/identity, rejected state restoration,
disabled/retired queued changes, finite progress bounds, passive action refusal,
orientation, explicit style override/removal and accessible values/actions.
Native Cocoa object fixtures verify three role mappings and NSNumber boolean
set conversion without a window, native attachment or system input.

The Manual software fixture routes Enter to a focused switch and a measured
pointer click to a toggle. It observes accepted state, completed raster changes,
progress geometry/value and changed progress pixels. A final-source
[software frame](records/controls.png) is stored. This fixture uses explicit fonts,
colors, workers and clock; it does not measure physical presentation or latency.
The decorative separator's native type/orientation is checked separately.

Internal review identified stale checked-background styling after an application
rejects a toggle request. The corrected direct Core fixture routes a real Switch
Click and explicitly completes the queued pre-owner post-event Styler pass;
it reproduces that visible style assertion on the preceding candidate. This is
separate from the Manual accepted input fixture and from physical OS input.
The same final test passes after the native-mismatch redraw fix.

Nineteen command receipts preserve fifteen passes and four raw exit-1 attempts:
the initial portable suite's temporary Unix sockets were sandbox-denied and pass
with authorized local socket access; the first rejection fixture violated Manual's
required explicit colors; its replacement dereferenced the intentionally nil
unchecked background and panicked in Go; the corrected pre-fix fixture then failed
the genuine stale-background assertion. All sources/logs are retained with their
scope, and none of these raw failures is a passing acceptance result. Final six
required checks bind the exact fourteen source leaves and unchanged 6,198 prepared
dependency identities. The earlier thirteen-source review snapshot is preserved
by hash with an explicit original-path mapping after final snapshots were refreshed.

## Review and evidence

Internal source review approved the final fourteen files after the styling fix.
Prism ran once using Anthropic Claude Sonnet 5.5 with complete fourteen-file
coverage: 3 chunks, 32621 input/3350
output tokens, raw exit 0 and no findings. A separate internal evidence review
binds this report, source variants, raw receipts, producers and archive before
commit. The [checkpoint](records/checkpoint.json) binds the reviews and
[archive](records/evidence.tar.gz) with its [member index](records/evidence-index.json).
Native binaries/caches and complete compiler/SDK closure are excluded.

No desktop windows were opened, OS input sent, settings changed or general
clipboard written. Human VoiceOver and physical input acceptance, native/GPU heap,
sustained performance/endurance and full Mac qualification remain open. Mac UI
vocabulary expansion continues alongside those foundation checks.
