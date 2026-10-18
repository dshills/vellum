# Mac control acceptance map

This maps the frozen catalogue for [M2](macos-delivery.md) to executable
scenarios. The catalogue is the 34 `ui.Kind` constants in `node.go`,
`controls.go`, `grid.go` and `tree.go`, plus the Breadcrumb and ColorPicker
composites. Composite controls use the same public node APIs.

**R** means primitive input routed through the manual driver's actual managed
widgets with measured geometry/completed software frames. **A** means native
adapter/object or owner/model checks; direct setters and semantic handlers are
not physical input. **W** is the full public workspace replay. **N** is a named native Go FIFO
record; OS/physical input and native pixels are separate evidence. Test names below
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
| `Image` | `TestManagedImageManualCenteredFitAndCompletedPixels`, `TestManagedImageManualScrollClipPreservesSourceMapping`: fit, source mapping and clipped pixels | R |
| `Extension` | `TestPublicWorkbenchManualConsumer`: independently consumed public dock, repeated panel-local IDs and routed activation; `TestExtensionReplacementDestroysOnlyReplacedInstance`: replacement ownership | R/A |
| `SplitPane` | `TestSplitManualKeyboardPointerGeometryAndRetainedDraft`: divider keys/drag, measured ratio and retained child editing | R |
| `Toolbar` | `TestChromeManualToolbarOverflowAndStatusUpdates`: overflowing tool reveal, Home/End/arrows and disabled skipping | R |
| `StatusBar` | `TestChromeManualToolbarOverflowAndStatusUpdates`: owner change reaches caption, semantics and completed pixels | R |
| `FormGroup` | `TestChromeManualToolbarOverflowAndStatusUpdates`: literal caption and child focus traversal | R |
| `Disclosure` | `TestDisclosureManualPointerKeyboardGeometryAndFocus`: collapse/reopen, layout release, focus and retained children | R |
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
| `TextField` | `TestManualRealManagedProjectInput`: routed text edit and Undo; `TestManagedTextFieldFocusLossWaitsForAdmittedNativeFIFO`: native input handoff | R/A |
| `MultilineText` | `TestFrozenWorkspaceRecordsAcrossBatches`: notes edits/history and UTF-16 selection; `TestManagedEditorCompositionHistoryExactBytes`: composition/committed-byte boundary | W/A |
| `Tabs` | `TestTabCloseActualPointerKeyboardAndGeometry`: controlled pointer/key close and stale target refusal; `TestNativeTabsReorderReparentAndDisabledOption`: page identity | R/A |
| `Scroll` | `TestManagedImageManualScrollClipPreservesSourceMapping`, `TestPlacementNativeScrolledContentAndPreviewPixels`: actual ancestor scroll clips and transformed content | R |
| `Dialog` | `TestManualModalDisabledAndStaleRealRouter`: modal refusal of background input; `TestDialogSceneLifecyclePreservesChildrenAndDismissesOnce`: preserved children and single dismissal | R/A |
| `Canvas2D` | `TestPlacementActualRoutingScaledAcrossBatches`, `TestPlacementRoutedCancellationAndRetirement`: transformed placement/history, consumption and cancellation; `TestCanvasCompletedSemanticsAndSelection`: completed object geometry | R/A |
| `DataGrid` | `TestStoredGridRoutedReplayAcrossBatches`, `TestManagedGridRowSelectionActualPointerKeysAllAndPaging`: routed stored data, selection, scrolling and completed frames | R |
| `DataTree` | `TestManagedTreeActualGeometryAndKeyboard`, `TestManagedTreePointerDisclosureAndRetiredPress`: indexed geometry, keys/disclosure and retired presses | R |
| Breadcrumb composite | `TestBreadcrumbManagedInputIdentityAndFrames`, `TestBreadcrumbNarrowOverflowRevealsEnabledAncestors`: stable action IDs, routed input and narrow overflow | R |
| ColorPicker composite | `TestColorPickerManagedKeyboardReconcileAndPixels`; `TestGalleryColorPickerAcceptedInvalidRejectedAndNormalized`: byte edit/pixels and model validation | R/A |

Driver scenarios are in `drivers/cogent/*_test.go`, the public consumer scenarios
are in `examples/replay/*_test.go` and `examples/replay/workspace/*_test.go`,
and the gallery model scenarios are in `drivers/cogent/cmd/gallery/main_test.go`.
The portable root package tests retain the detailed validation contracts in
[controls.md](controls.md) and the linked per-control specifications.

## Remaining Mac acceptance

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
