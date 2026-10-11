#!/bin/sh
set -eu
VELLUM_REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$VELLUM_REPO_ROOT"
. ./drivers/cogent/env.sh
if [ ! -f drivers/cogent/build/upstream/go.mod ] || [ ! -f drivers/cogent/build/typesetting/go.mod ]; then
 echo 'Run drivers/cogent/prepare.sh before native checks.' >&2
 exit 1
fi
python3 drivers/cogent/prepare_test.py
python3 scripts/check-text-backend.py
python3 scripts/check-editor-accessibility_test.py
python3 scripts/benchmark-native_test.py
python3 scripts/benchmark-grid_test.py
python3 scripts/child_process_test.py
python3 scripts/release-sdk_test.py
python3 scripts/check-sdk_test.py
python3 scripts/package-app_test.py
go test -race ./...
go vet ./...
cd drivers/cogent
go test -race ./...
go vet ./...
mkdir -p build
go build -o build/vellum-workbench ./cmd/workbench
# Replacement dependency tests do not run transitively with driver tests.
# Use the driver module so both pinned replacement modules are applied.
go test -race -tags offscreen cogentcore.org/core/paint/renderers/rasterx/scan cogentcore.org/core/paint/renderers/rasterx
(cd build/upstream && go test -race ./system ./system/driver/base ./system/driver/desktop ./system/composer ./gpu ./gpu/gpudraw ./core ./xyz/xyzcore -run 'TestVellumSourceTiming|TestVellumFrameTrace|TestVellumMeasuredComposition|TestVellumReviewTextCursor|TestComposerPublicationAndWithdrawal|TestLockRechecksAfterWaitingForNativeTeardown|TestSceneShowUsesOwnWindowTitle|TestQueuedCloseStopsLoopAfterComposerWithdrawal|TestNewWindowSizingOwnsContextDuringOtherWindowResize|TestSameWindowStageSizingUsesExistingContext|TestScreenSnapshotDuringCacheRebuild|TestWindowOptionsFixupUsesScreenSnapshot|TestMainQueuePublicationDuringStartupDispatch|TestMainQueueInitializationPreservesIdentityAndCompletion|TestMainThreadRequiredCallbackWaitsForLoopPublication|TestComposerWithdrawnBeforeNativeTeardownEvenWithoutGLFW|TestDeviceLoss|TestSurfaceAcquisition|TestRetiredDirectRenderSource|TestComposition|TestInlineComposition|TestIdleBarrier|TestVellumScreenPoll|TestVellumScreenAccessors|TestVellumCurrentWindowSnapshot|TestVellumRedraw|TestVellumComposer|TestVellumUnsafeResizeReplay|TestVellumRenderedGeometryScrollAndWrap|TestVellumFocuslessShortcutPhases|TestVellumMainQueueStop|TestVellumMainQueueFailure|TestVellumNativeEventLifecycle|TestVellumTerminalMainRecovery|TestVellumQuitClean|TestVellumTextureCopy|TestVellumWindowList')

"$VELLUM_REPO_ROOT/scripts/check-example.sh"

(cd "$VELLUM_REPO_ROOT/examples/sdk-contract" && go test -race ./... && go vet ./...)

# Explicit routing uses the offscreen driver; these processes do not open GUI windows.
(cd "$VELLUM_REPO_ROOT/drivers/cogent" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)
(cd "$VELLUM_REPO_ROOT/drivers/cogent/build/upstream" && go test -race -tags offscreen ./core ./events ./text/lines ./text/textcore ./system/driver/offscreen -run 'TestVellumSourceTiming|TestVellumSpriteSource|TestVellumReviewTextCursor|TestManual|TestVellumManual|TestVellumSourceClock|TestVellumDefaultRepeatClick|TestVellumOutputBuffer')
(cd "$VELLUM_REPO_ROOT/examples/replay" && go test -race ./... && go vet ./... && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)
(cd "$VELLUM_REPO_ROOT/examples/grid" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)
(cd "$VELLUM_REPO_ROOT/examples/loadedgrid" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)
(cd "$VELLUM_REPO_ROOT/examples/loadedtree" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)
