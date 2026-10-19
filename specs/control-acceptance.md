# Mac control acceptance map

This maps the frozen catalogue for [M2](macos-delivery.md) to executable
scenarios. The catalogue is the 34 `ui.Kind` constants in `node.go`,
`controls.go`, `grid.go` and `tree.go`, plus the Breadcrumb and ColorPicker
composites. Composite controls use the same public node APIs.

**R** means primitive input routed through the manual driver's actual managed
widgets with measured geometry/completed software frames. **A** means native
adapter/object or owner/model checks; direct setters and semantic handlers are
not physical input. **W** is the full public workspace replay. **N** is a named native Go FIFO
record. **N-partial** means source-asserted observations from a failed native
record with independent validation incomplete; it does not qualify a native pass.
OS/physical input and native pixels are separate evidence. Test names below
identify their containing source files. The [M2 result](../docs/reports/macos-workspace-acceptance/README.md)
records actual execution and qualified reuse. This is a coverage map, not a
claim that all Mac acceptance gates have passed.

| Public kind | Named scenario and checked behavior | Method |
| --- | --- | --- |
| `Column` | `TestFrozenWorkspaceRecordsAcrossBatches`: composed layout, measured bounds and frozen captures | W |
| `Row` | `TestColorPickerManagedKeyboardReconcileAndPixels`: horizontal preview/value composition and completed pixels | R |
| `Label` | `TestRetainedReconcileAndExtensionFailure`: changed public label text reaches the retained native text widget without replacing its identity | A |
| `Button` | `TestIconsManualArtworkInputAndReplacement`: keyboard/pointer activation, completed artwork and retained identity | R |
| `Icon` | `TestIconsManualArtworkInputAndReplacement`: bundled symbol replacement, geometry and nonempty artwork | R |
| `Image` | `TestManagedImageManualCenteredFitAndCompletedPixels`, `TestManagedImageManualScrollClipPreservesSourceMapping`: fit, source mapping and clipped pixels; native `gallery -media-native-check` checks accepted fit/replacement callbacks, role and visible geometry (native pixels separate) | R/N |
| `Extension` | `TestPublicWorkbenchManualConsumer`: independently consumed public dock, repeated panel-local IDs and routed activation; `TestExtensionReplacementDestroysOnlyReplacedInstance`: replacement ownership | R/A |
| `SplitPane` | `TestSplitManualKeyboardPointerGeometryAndRetainedDraft`: divider keys/drag, measured ratio and retained child editing; `gallery -structure-native-check`: full50-input42-state native Go FIFO record, measured focus/geometry/state, retained pages, rejected/accepted close and stale-target retirement | R/N |
| `Toolbar` | `TestChromeManualToolbarOverflowAndStatusUpdates`: overflowing tool reveal, Home/End/arrows and disabled skipping; native `gallery -media-native-check` checks actual overflow reveal, End/Home, disabled skipping and one activation | R/N |
| `StatusBar` | `TestChromeManualToolbarOverflowAndStatusUpdates`: owner change reaches caption, semantics and completed pixels; native `gallery -media-native-check` checks accepted caption after toolbar/path actions | R/N |
| `FormGroup` | `TestChromeManualToolbarOverflowAndStatusUpdates`: literal caption and child focus traversal; native `gallery -media-native-check` checks literal caption and actual field traversal | R/N |
| `Disclosure` | `TestDisclosureManualPointerKeyboardGeometryAndFocus`: collapse/reopen, layout release, focus and retained children; `gallery -structure-native-check`: full50-input42-state native Go FIFO record, measured focus/geometry/state, retained pages, rejected/accepted close and stale-target retirement | R/N |
| `Checkbox` | `TestChoiceManualQuietPointerFocusAdvancesOrdinaryControl`: measured pointer focus and Tab; `TestNativeControlsChangeRetainedCallbacksAndDisabled`: typed change and controlled state | R/A |
| `Switch` | `TestExpandedControlsManualFrameAndInput`: focused key changes accepted boolean and completed frame | R |
| `ToggleButton` | `TestExpandedControlsManualFrameAndInput`, `TestExpandedRejectedToggleRestoresVisibleStyle`: pointer activation and rejected-value styling | R |
| `ProgressBar` | `TestExpandedControlsManualFrameAndInput`: accepted fraction, geometry and completed fill pixels | R |
| `ProgressRing` | `TestProgressShapesCompletedFramesAcrossDPI`: multiple fractions/DPI, changed pixels, retained identity and no passive repaint timer | R |
| `ProgressGauge` | `TestProgressShapesCompletedFramesAcrossDPI`: semicircle fractions/DPI and passive lifetime | R |
| `Separator` | `TestExpandedControlsManualFrameAndInput`: composed decoration; `TestExpandedControlsValidationAndFocus` in `ui/expanded_controls_test.go`: descriptor rules | R/A |
| `Selector` | `TestOrdinarySelectorAndNumberKeyboardControlledRoundTrip`: focused arrows skip disabled options, wrap by stable ID and restore rejected values | R |
| `ListBox` | `TestListBoxActualScrollingNavigationAndPixels`, `TestListBoxActualWorkbenchModalFocusAndClosure`: scrolling selection, focus scopes and retirement; `gallery -navigation-native-check`: native scroll/rejection/read-only/disabled/Tab and exactly-once proposals | R/N |
| `MenuButton` | `TestMenuButtonRealPopupRoutingAndRetirement`: open, measured item click, duplicate labels, dismissal and stale popup refusal; `gallery -navigation-native-check`: native duplicate/check/separator actions, dismissal, modal disable and closure | R/N |
| `RadioGroup` | `TestChoiceManualKeyboardPointerAndFocus`: arrow wrap/disabled skip, measured pointer, one Tab stop and exclusive state | R |
| `SegmentedControl` | `TestChoiceManualKeyboardPointerAndFocus`: same scenario over segmented presentation | R |
| `Slider` | `TestColorPickerManagedKeyboardReconcileAndPixels`: actual focused key changes a managed byte slider; `TestNativeControlsChangeRetainedCallbacksAndDisabled`: typed payload and finite refusal | R/A |
| `NumberField` | `TestOrdinarySelectorAndNumberKeyboardControlledRoundTrip`: focused numeric step, exactly one proposal, rejected-value restoration and disabled refusal | R |
| `TextField` | `TestManualRealManagedProjectInput`: routed text edit and Undo; `TestManagedTextFieldFocusLossWaitsForAdmittedNativeFIFO`: native input handoff; `TestTextCataloguePublicManual`: shared text/editor/modal public routing at160/162DPI; native `gallery -text-native-check`: current9013a73 full68-input33-state record passes; previous failed scopes stay sealed | R/A/N |
| `MultilineText` | `TestFrozenWorkspaceRecordsAcrossBatches`: notes edits/history and UTF-16 selection; `TestManagedEditorCompositionHistoryExactBytes`: composition/committed-byte boundary; `TestTextCataloguePublicManual`: shared text/editor/modal public routing at160/162DPI; native `gallery -text-native-check`: current9013a73 full68-input33-state record passes; previous failed scopes stay sealed | W/A/N |
| `Tabs` | `TestTabCloseActualPointerKeyboardAndGeometry`: controlled pointer/key close and stale target refusal; `TestNativeTabsReorderReparentAndDisabledOption`: page identity; `gallery -structure-native-check`: full50-input42-state native Go FIFO record, measured focus/geometry/state, retained pages, rejected/accepted close and stale-target retirement | R/A/N |
| `Scroll` | `TestManagedImageManualScrollClipPreservesSourceMapping`, `TestPlacementNativeScrolledContentAndPreviewPixels`: actual ancestor scroll clips and transformed content; `gallery -structure-native-check`: full50-input42-state native Go FIFO record, measured focus/geometry/state, retained pages, rejected/accepted close and stale-target retirement | R/N |
| `Dialog` | `TestManualModalDisabledAndStaleRealRouter`: modal refusal of background input; `TestDialogSceneLifecyclePreservesChildrenAndDismissesOnce`: preserved children and single dismissal; `TestTextCataloguePublicManual`: shared text/editor/modal public routing at160/162DPI | R/A |
| `Canvas2D` | `TestPlacementActualRoutingScaledAcrossBatches`, `TestPlacementRoutedCancellationAndRetirement`: transformed placement/history, consumption and cancellation; `TestCanvasCompletedSemanticsAndSelection`: completed object geometry | R/A |
| `DataGrid` | `TestStoredGridRoutedReplayAcrossBatches`, `TestManagedGridRowSelectionActualPointerKeysAllAndPaging`: routed stored data, selection, scrolling and completed frames | R |
| `DataTree` | `TestManagedTreeActualGeometryAndKeyboard`, `TestManagedTreePointerDisclosureAndRetiredPress`: indexed geometry, keys/disclosure and retired presses | R |
| Breadcrumb composite | `TestBreadcrumbManagedInputIdentityAndFrames`, `TestBreadcrumbNarrowOverflowRevealsEnabledAncestors`: stable action IDs, routed input and narrow overflow; native `gallery -media-native-check` checks accepted/rejected/disabled path, current role and stale-frame refusal | R/N |
| ColorPicker composite | `TestColorPickerManagedKeyboardReconcileAndPixels`; `TestGalleryColorPickerAcceptedInvalidRejectedAndNormalized`: byte edit/pixels and model validation; native `gallery -media-native-check` checks byte/hex/reject/normalize/invalid/straight-alpha and disabled/hidden refusal | R/A/N |

Driver scenarios are in `drivers/cogent/*_test.go`, the public consumer scenarios
are in `examples/replay/*_test.go` and `examples/replay/workspace/*_test.go`,
and the gallery model scenarios are in `drivers/cogent/cmd/gallery/main_test.go`.
The portable root package tests retain the detailed validation contracts in
[controls.md](controls.md) and the linked per-control specifications.

## Remaining Mac acceptance

The [current catalogue and assertion audit](../docs/reports/macos-current-catalogue/README.md)
renews all five native Go FIFO records on `1338b11`/stage87: 255 inputs and
169 observations, independently reviewed. The audit covers 36 rows, 29 requirement
groups and 205 exact executed test functions, with source assertion predicates
and raw PASS locations retained. Three explicit M2 methods remain open: attached
Cocoa-selector tab closing (eight stages), Breadcrumb (twelve stages) and
ColorPicker (eight steps, seven edits and one invalid proposal). The current
Go FIFO records do not replace those dedicated methods. Their existing fixtures
are built and source-reviewed, with no new native launch while the owner is
online. M2 remains partial; physical/human and M3–M7 requirements remain open.

The older entries below retain their original source/configuration and failure
identities. They are historical checkpoints, not outstanding copies of already
qualified current records.

The map supplies a named automated scenario for every catalogue entry. The
[full native workspace checkpoint](../docs/reports/macos-native-workspace/README.md)
now passes all 216 operations and 27 captures at both 10k/1M rows. It covers the
record's command/history, workers, edits, tab reorder, placement, modal precedence
and two-window survival through native Go event routing. It does not establish
native catalogue completeness, physical input or human assistive-technology use.

The historical [popup/Selector integration checkpoint](../docs/reports/macos-native-controls/README.md)
repairs owner wake starvation and actual Selector click focus. Its final native
attempt reaches the source-asserted Selector states and one accepted typed
number, then fails the next focused-key guard after numeric Enter. All four
native attempts remain failed/sealed; no full native subcase closes. The original check reached 24/11 of its 49/21 record. The corrected checker now
authors four routed numeric refocus clicks: 57 inputs/21 unchanged observations;
that checkpoint left native acceptance pending. The subsequent result below
completes this record.

The subsequent [retained-state and multi-click checkpoint](../docs/reports/macos-control-refocus/README.md) completes the unchanged native
Selector/NumberField record: 57 routed inputs and 21 independently expected
observations. Go adapter fixes restore inherited disabled styling and actual
focus before ordinary double/triple-click handling, preserving selection and
exactly-once edits. Two preceding 57-input cases remain failed at 54 receipts/19
source-asserted observations; the styling-only case did not fix actual focus.
Controlled Go2/SDK786a CPU-scene routing is qualified for this record only. Full
catalogue/M2, physical/human M5 and M3–M7 remain open; no SDK refresh occurs.

The [choice catalogue checkpoint](../docs/reports/macos-choice-catalogue/README.md)
adds a complete 30-input/23-observation native record for Checkbox, Switch,
ToggleButton, RadioGroup, SegmentedControl and Slider keyboard behavior. Shared
state, rejected changes, disabled skipping/refusal, focus and exactly-once
callbacks pass. ProgressBar/Ring/Gauge and status publish expected values and
fully visible geometry; native progress pixels and Slider drag are not claimed.
Original-reference/session/process retirement passes. This advances named
catalogue coverage in controlled Go2/SDK786a CPU-scene routing; remaining catalogue,
full M2/M5, physical/human input and M3–M7 remain open. No SDK refresh occurs.

The [list/menu catalogue checkpoint](../docs/reports/macos-navigation-catalogue/README.md)
adds stage84 support for direct Button embeddings in popup inspection and
47 native Go FIFO inputs/41 independent observations. The original 17-input
case remains failed/sealed; only the distinct corrected scope passes. ListBox scrolling,
explicit repeat proposals, rejection/read-only/disabled behavior and Tab pass;
MenuButton duplicate-label actions, actual popup pointer/keys, disabled/stale
refusal, Escape/outside consumption, modal owner disable, check/separator/reset
and original-window closure pass. Callback totals are 8 list, 2 menu, 3 check/reset
and 1 outside action. This is controlled Go2/SDK786a CPU-scene routing. Remaining
catalogue, full M2/M5, physical/human input and M3–M7 stay open; no SDK refresh.

The [structural catalogue checkpoint](../docs/reports/macos-structure-catalogue/README.md)
adds tested controlled Tabs keyboard navigation. The new 50-input/42-state
native scenario fails after41 inputs/35 source-asserted observations, before
document selection/close results. The measured-frame acquisition reason is
unknown. Review establishes that disabled styling changes metrics, invalidating
the original cross-style geometry equality guard. The failed scope remains sealed. Full structural qualification and
physical/pixels/M2/M5/M3–M7 remain open; N-partial means a failed record prefix.

The remaining native catalogue work is:

1. Exercise the rest of the supported catalogue's focus, input, state publication
   and lifetime through actual native routing, with named evidence for each
   entry. Keep automated Go routing separate from OS and physical input.
2. Complete real keyboard/pointer and human assistive-technology workflows,
   including modal/focus boundaries, text/IME and virtual data. Follow the M5
   native matrix in `macos-delivery.md`; software DPI changes do not qualify
   physical display transitions.

Full M2/M5 remain open. Live tests require current desktop authorization. No
records or image goldens are regenerated to obtain a passing result. Performance,
idle, memory and endurance belong to M3/M4 with their original thresholds.

The structural native prefix above predates the subsequent Prism tab activation
and disabled-focus fixes. Those fixes pass actual Core queue/manager routing and
complete driver-root/gallery race suites. At that prior checkpoint they had not
received a new native run. The failed diagnostic attempt below now exercises
those fixes but does not complete structural qualification. Both failed scopes
remain sealed and the full structural/native gate stays open.


### Measured acquisition checkpoint

The [acquisition diagnostics result](../docs/reports/macos-acquisition-diagnostics/README.md)
adds measured-only actual source/stage/status/attempt detail without changing
ordinary retry/release behavior or strict completed-frame criteria. A distinct
native50/42 scope fails after16 inputs/15 source-asserted observations with
actual `occluded(196609)` on one acquisition attempt. Cleanup and separate whole
post-failure boundaries pass; full native validation remains open. This result
does not identify the older failure's cause. Both scopes remain sealed.
The new style-aware oracle has separate public Manual160/162DPI and41 synthetic
controls, with no complete native oracle pass. Complete driver-root (`.`)/gallery race, focused
GPU/Core regressions, build, prepare and affected-package vet pass; default Core
vet retains12 exact baseline diagnostics. That checkpoint proposed visibility-aware
Frame scheduling. The typed-terminal decision below preserves the existing
contract while an authoritative drawable-ready event remains unestablished.
Full M2/M5 and M3–M7 remain open; performance targets and release gates are unchanged.


### Typed unavailable Frame checkpoint

The [typed terminal result](../docs/reports/macos-frame-unavailable/README.md)
preserves the first-composition contract and classifies checked known surface
yields through `DesktopCompositionError`/`ErrDesktopFrameUnavailable`. No failed
frame token, automatic acquisition/requeue or readiness inference is introduced.
Cancellation identity is checked only after the actual trace terminal; capacity
and lifetime rules remain bounded. Late cancellation may race a committed reply
and does not retract prior work. Contradictory metadata and source-detail
faults fail closed. Full driver-root/gallery race, focused controls, affected vet,
build and deterministic stage86 preparation pass; default Core vet retains12
baseline diagnostics. Synthetic measured Manual results are software evidence.
No new native case runs; the earlier16/15 and41/35 failures stay sealed. Automatic
pending/resumption needs an authoritative drawable-availability event that is
not established. Continue remaining native catalogue/typed-error acceptance with
explicit availability admission and no retry loop. M2/M5/M3–M7 and all unchanged
performance/release gates remain open.

The [media/chrome catalogue checkpoint](../docs/reports/macos-media-catalogue/README.md)
completes 71 native Go FIFO inputs/42 observations for shared Image, Toolbar,
StatusBar, FormGroup, Breadcrumb and ColorPicker views. Actual focus, measured
pointer targets, overflow, accepted/rejected/disabled/stale/hidden states and
exact callbacks pass; original window/session/process retire. Public Manual
160/162 DPI routing and independently sampled immutable-image replacement pixels,
full gallery race/vet/build and 13 negative validator controls pass separately.
Initial checker/preflight failures remain preserved. This is controlled
Go 1.25.6/GOMAXPROCS2/SDK786a native routing; physical/OS input, IME/human accessibility, native
pixels/presentation and full M2/M5/M3–M7 stay open. No SDK refresh occurs.

Prism then identifies a real undispatched-owner wait on application retirement.
The final media runner cancels after Run and gives final close acknowledgement/
closed-reference validation a separate 1-second context within its 6-second join.
Two actual Application regressions discriminate the missing-cancel and canceled-
close counterfactuals. Final gallery race/vet/build and the distinct corrected
71-input/42-observation native case pass; both native scopes remain sealed.
One original Anthropic Sonnet 5.5 call and independent fix reviews are retained.

The [text/editor/modal result](../docs/reports/macos-text-modal/README.md) supplies
N-partial evidence for TextField/MultilineText through `gallery -text-native-check`:
the intended68/33 record fails after35inputs/19independent observations on actual
typed drawable unavailability. Accepted/rejected/disabled-field behavior,
Unicode/grapheme edits, journal Undo/Redo and editor Tab are validated in this
failed prefix. Disabled-editor/modal/final-close native results remain open;
full software33states pass160/162DPI. The case is permanently sealed without
retry. Physical/OS input, live IME, human accessibility and native pixels remain
separate; full M2/M5/M3–M7 and remaining structural acceptance stay open.

### Complete structural record; text publication fix

The [structural/publication checkpoint](../docs/reports/macos-editor-publication/README.md)
independently validates all50 inputs/42 states for Disclosure, Tabs, SplitPane,
Scroll and document closure at d1232b5. Rejected close retains header focus;
the following Delete requests and accepted removal match the unchanged oracle.
Earlier failed scopes remain failed/sealed. The separate text/modal case fails
at editor-disabled after35 inputs/19 source-asserted states, with952 completed
Frames and a final deadline failure; no disabled snapshot or modal acceptance
is established. Its publication loop is repaired and covered by six actual
publication-path counterfactuals. Fresh complete text/modal68/33 native acceptance
is next, followed by the M2 requirement audit. Physical/human M5 and M3–M7 remain
separate and open; no record/golden or Frame-unavailable policy changes.


### Current body/gutter typography qualification

The [editor typography checkpoint](../docs/reports/macos-editor-gutter-font/README.md)
restores the original six216-operation27-capture software workspace records at
10k/1M rows and owner batches1/7/64 under Go1.25.6. Document body styling remains
monospaced; line-number chrome retains UI theme typography. Original records and
image goldens are unchanged. Retained typography/DPI/tab-cache and selection tests
pass. This is software evidence, separate from native catalogue completion.

A fresh native text/modal attempt on5e30a6c remains failed/sealed at name-rejected
(21 inputs/3 source-asserted states). Its final sampled visibility is known/false,
sequence3, followed by one typed surfaceoccluded196609 acquisition. The disabled
and modal phases are not reached; no native qualification of the prior disabled
publication repair follows. Complete native text/modal68/33 acceptance and the
current-candidate M2 requirement audit remain open. Physical/human M5 and M3–M7
are separate; no acquisition retry or record/golden change is introduced.


### Current workspace and text workflow qualification

The [current workflow checkpoint](../docs/reports/macos-current-workflows/README.md)
qualifies9013a73/stage87 native full workspace432operations54captures and complete
text/editor/modal68inputs33states under Go1.25.6/Go2/SDK786a. The authored disabled
and modal phases now pass. Six original software records1296/162 and portable
root race/defaultvet pass. Every named scenario across34kinds+2composites has
an exact PASS line in current software logs; that is execution coverage, not
complete dedicated-contract or native-method qualification. Historical failed
text/acquisition scopes remain failed/sealed; no causal explanation is inferred.

Current docking source assertions and persisted session layout pass. Document
source assertions and retained output are independently accepted after a separate
source-backed footer correction:1000component nodes+project+group=1002total.
The original document controller remains FAILED_OR_CANCELED; no native rerun.
Injected chooser/native AX actions are separate from real panels/human VoiceOver.
M2 still needs current native catalogue shared-runtime qualification or exact
reuse and its dedicated-contract audit. Physical/native presentation/M5 and
M3–M7/fullMac remain open. No records/goldens or OS settings change.
