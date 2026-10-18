# Native choices, toggles and progress publication

The standard gallery now offers `-choice-native-check`: **30 public Go FIFO
inputs and 23 independently expected observations** pass on this Mac. It reuses
the ordinary gallery's public `ui.Node` factories for Checkbox, Switch,
ToggleButton, RadioGroup, SegmentedControl, Slider and three progress kinds.
Applications consume the existing public Go APIs; no new driver API or dependency
patch is introduced. This is one named native catalogue record, not full M2/M5
or Mac completion.

The record checks pointer and Space activation, ancestor disabling/refusal and
re-enabling, actual keyboard focus, exclusive shared choice state, disabled
option skipping, rejected selection restoration, and no duplicate proposal when
Space accepts the already selected segment. Slider keys test accepted/rejected
changes, Home/End bounds and progress/status publication at zero and full scale.
The final ordinary action fires once. Closing the original window invalidates
its reference; later input is refused, and the public session closes with no
pending request or live window. The owned process group is confirmed absent.

The Slider sends one Change proposal per attempted step: Right at 100 proposes
the clamped value 100 once. It therefore ends with five numeric callbacks,
including the rejected increment. Choice Space on the selected option emits no
additional proposal. These distinct existing contracts were authored from source
before native execution; the expected native record was not changed after a run.

## Checks and review

The complete gallery race suite, gallery vet and native build pass on the final
four Go files. An independent public Manual consumer tests geometry, keyboard
focus order, shared selection, rejected changes, progress values and a disabled/
re-enabled round trip. These software checks are separate from native evidence.
The unchanged root driver's complete race suite qualified at base `500cf80` is
reused only for identical affected driver bytes; it is not claimed as a new run.

The external native validator checks all 30 ordered packet identities/types/key
runes, current frames, actual focused keyboard targets and measured pointer
bounds, all 23 literal models/callback counts, group-qualified option identities,
effective enabled flags, controlled semantics, progress/status and closure.
Twenty-four synthetic positive/negative validator controls pass. Synthetic data
is not native proof. This checkpoint requires internal source/controller, exact-plan admission and
final evidence reviews plus one scoped Prism Anthropic Claude Sonnet 5.5 review.

Earlier headless failures remain retained: an initial wrong state type prevented
compilation; an output-directory collision produced an unqualified build without
a receipt; software routing caught missing Space runes before any native launch.
An initial validator test process passed all tests but sandbox restrictions
prevented process-group observation; its distinct owned rerun passes. Only the
final current-source receipts and complete native record qualify this checkpoint.

Prism reviewed all 11 staged files (51,880 bytes) in one Anthropic Claude Sonnet
5.5 call, with no excluded/skipped/truncated input or suppressed findings. The
raw CLI exits 1 with one medium diagnostic finding; one further medium comment
is discarded by Prism's quoted-evidence matching. Both raw comments are retained
and independently assessed. The retained lookup-error condition cannot occur:
the same detached snapshot's full ID set is checked before the later subset
lookups, with no intervening mutation. Ordinary gallery rejection flags remain
false and are toggled only by catalogue checks/tests; the discarded comment
does not establish a normal-mode regression. No source change or provider rerun
is needed. This factual README-only addition receives final internal review;
the original provider payload/scope and reviewed README are preserved separately.

## Limits and next work

This is controlled GOMAXPROCS=2, Go1.25.6 Darwin/ARM64, CPU-scene native routing
over the retained source-built archive `786aaad8...`. All 6267 stage83 prepared
files and all 7880 retained SDK files remain unchanged. It does not qualify a
refreshed SDK, another native archive/configuration or production performance.

Progress bars/ring/gauge, icon/label and composed chrome are checked for numeric
semantics and fully visible layout; this new native record does not assert their
pixels or display presentation. Slider pointer dragging remains separate. Go
FIFO input does not verify physical/OS pointer and keyboard, IME, VoiceOver,
display transitions or Cocoa deallocation. No OS settings/input sources/clipboard
change, SDK refresh or public release is performed. Full M2 and M3–M7 remain open.

Continue native list/menu, tabs/disclosure/scroll/split, text/modal and composite
control scenarios against the [acceptance map](../../../specs/control-acceptance.md),
then the unchanged performance/resource/platform/SDK gates in the
[delivery plan](../../../specs/macos-delivery.md). Use the
[gallery instructions](../../../drivers/cogent/cmd/gallery/README.md) to run it.
The [bounded archive](records/checkpoint.tar.gz) and [member manifest](records/manifest.json)
are required commit artifacts for complete qualifying logs and earlier failures;
binaries and fonts are excluded.
