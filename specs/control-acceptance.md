# Mac control acceptance map

This maps the frozen catalogue for [M2](macos-delivery.md) to executable
scenarios. The catalogue is the 34 `ui.Kind` constants in `node.go`,
`controls.go`, `grid.go` and `tree.go`, plus the Breadcrumb and ColorPicker
composites. Composite controls use the same public node APIs.

**R** means primitive input routed through the manual driver's actual managed
widgets with measured geometry/completed software frames. **A** means native
adapter/object or owner/model checks; direct setters and semantic handlers are
not physical input. **W** is the full public workspace replay. Test names below
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
| `ListBox` | `TestListBoxActualScrollingNavigationAndPixels`, `TestListBoxActualWorkbenchModalFocusAndClosure`: scrolling selection, focus scopes and retirement | R |
| `MenuButton` | `TestMenuButtonRealPopupRoutingAndRetirement`: open, measured item click, duplicate labels, dismissal and stale popup refusal | R |
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

The map supplies a named automated scenario for every catalogue entry. It does
not close the native half of M2 or M5. The required next native scenarios are:

1. Run the professional workspace native integration variants at both 10k and
   1M rows, checking command/history, cancellation, document persistence,
   dock/float/redock, two-window survival and data edits. Its existing 18-action
   smoke is a native supplement; the 216-operation record is qualified separately.
2. In the gallery, exercise the ordinary Selector popup and NumberField typed,
   pointer and keyboard edits, including rejection/disabled behavior. The new
   deterministic keyboard regression does not qualify an actual desktop popup.
3. Exercise the supported catalogue with real keyboard/pointer input and human
   assistive technology, including modal/focus boundaries, text/IME and virtual
   data. Follow the native workflow matrix in `macos-delivery.md`; software DPI
   changes do not qualify physical display transitions.

These remain open while live testing is stopped. No records or image goldens
are regenerated to obtain a passing result. Performance, idle, memory and
endurance belong to M3/M4 and have their original thresholds.
