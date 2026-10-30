#!/bin/sh
set -eu
VELLUM_CHECK_STAGES='tooling portable native dependencies examples offscreen'
usage() { echo 'Usage: scripts/check.sh [--list-stages | --stage STAGE]' >&2; exit 2; }
if [ "$#" -eq 0 ]; then
 VELLUM_CHECK_STAGE=all
elif [ "$#" -eq 1 ] && [ "$1" = --list-stages ]; then
 for VELLUM_STAGE_NAME in $VELLUM_CHECK_STAGES; do echo "$VELLUM_STAGE_NAME"; done
 exit 0
elif [ "$#" -eq 2 ] && [ "$1" = --stage ]; then
 VELLUM_CHECK_STAGE=$2
 case "$VELLUM_CHECK_STAGE" in tooling|portable|native|dependencies|examples|offscreen) ;; *) usage ;; esac
else
 usage
fi
VELLUM_REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$VELLUM_REPO_ROOT"
. ./drivers/cogent/env.sh
if [ ! -f drivers/cogent/build/upstream/go.mod ] || [ ! -f drivers/cogent/build/typesetting/go.mod ]; then
 echo 'Run drivers/cogent/prepare.sh before native checks.' >&2
 exit 1
fi

check_tooling() (
cd "$VELLUM_REPO_ROOT"
python3 -B scripts/native-patch-inventory.py --check
python3 -B scripts/check-stages_test.py
python3 -B scripts/ci-cache-budget_test.py
python3 drivers/cogent/prepare_test.py
python3 scripts/check-text-backend_test.py
python3 scripts/check-text-backend.py
python3 scripts/check-editor-accessibility_test.py
python3 scripts/benchmark-native_test.py
python3 scripts/benchmark-grid_test.py
python3 scripts/child_process_test.py
python3 scripts/build-macos_test.py
python3 scripts/release-sdk_test.py
python3 -B scripts/new-app_test.py
python3 scripts/distribution-inventory_test.py
python3 scripts/check-sdk_test.py
python3 -B scripts/check-sdk-ci_test.py
python3 -B scripts/private-macos-bundle_test.py
python3 -B scripts/private_bundle_preflight_test.py
python3 -B scripts/install-private-macos_test.py
python3 -B scripts/verify-private-macos_test.py
python3 scripts/package-app_test.py
python3 scripts/app_icons_test.py
)

check_portable() (
cd "$VELLUM_REPO_ROOT"
go test -race ./...
go vet ./...
)

check_native() (
cd "$VELLUM_REPO_ROOT"
cd drivers/cogent
go test -race -timeout=20m ./...
go vet ./...
mkdir -p build
go build -o build/vellum-workbench ./cmd/workbench
)

check_dependencies() (
cd "$VELLUM_REPO_ROOT"
cd "$VELLUM_REPO_ROOT/drivers/cogent"
# Replacement dependency tests do not run transitively with driver tests.
# Use the driver module so both pinned replacement modules are applied.
go test -race -tags offscreen cogentcore.org/core/paint/renderers/rasterx/scan cogentcore.org/core/paint/renderers/rasterx
go test -race -tags offscreen -timeout=120s cogentcore.org/core/paint/renderers/sceneplan cogentcore.org/core/paint/renderers/gpuscene
go test -race -timeout=120s github.com/go-text/typesetting/font -run "^TestVellum(BoundedOutline|OutlineCost)"
(cd build/upstream && go test -race ./system ./system/driver/base ./system/driver/desktop ./system/composer ./gpu ./gpu/gpudraw ./core ./xyz/xyzcore -run 'TestVellumSourceTiming|TestVellumFrameTrace|TestVellumMeasuredComposition|TestVellumReviewTextCursor|TestComposerPublicationAndWithdrawal|TestLockRechecksAfterWaitingForNativeTeardown|TestSceneShowUsesOwnWindowTitle|TestQueuedCloseStopsLoopAfterComposerWithdrawal|TestNewWindowSizingOwnsContextDuringOtherWindowResize|TestSameWindowStageSizingUsesExistingContext|TestScreenSnapshotDuringCacheRebuild|TestWindowOptionsFixupUsesScreenSnapshot|TestMainQueuePublicationDuringStartupDispatch|TestMainQueueInitializationPreservesIdentityAndCompletion|TestMainThreadRequiredCallbackWaitsForLoopPublication|TestComposerWithdrawnBeforeNativeTeardownEvenWithoutGLFW|TestDeviceLoss|TestSurfaceAcquisition|TestRetiredDirectRenderSource|TestComposition|TestInlineComposition|TestIdleBarrier|TestVellumScreenPoll|TestVellumScreenAccessors|TestVellumCurrentWindowSnapshot|TestVellumRedraw|TestVellumComposer|TestVellumUnsafeResizeReplay|TestVellumRenderedGeometryScrollAndWrap|TestVellumFocuslessShortcutPhases|TestVellumMacIngress|TestVellumMainQueueStop|TestVellumMainQueueFailure|TestVellumNativeEventLifecycle|TestVellumTerminalMainRecovery|TestVellumQuitClean|TestVellumTextureCopy|TestVellumWindowList')
)

check_examples() (
cd "$VELLUM_REPO_ROOT"
"$VELLUM_REPO_ROOT/scripts/check-example.sh"

(cd "$VELLUM_REPO_ROOT/examples/sdk-contract" && go test -race ./... && go vet ./...)
)

check_offscreen() (
cd "$VELLUM_REPO_ROOT"
# Explicit routing uses the offscreen driver; these processes do not open GUI windows.
(cd "$VELLUM_REPO_ROOT/drivers/cogent" && go test -race -tags offscreen -timeout=20m ./... && go vet -tags offscreen ./...)
(cd "$VELLUM_REPO_ROOT/drivers/cogent/build/upstream" && go test -race -tags offscreen ./core ./events ./text/lines ./text/textcore ./system/driver/offscreen -run 'TestVellumSourceTiming|TestVellumSpriteSource|TestVellumReviewTextCursor|TestManual|TestVellumManual|TestVellumExposure|TestVellumSourceClock|TestVellumDefaultRepeatClick|TestVellumOutputBuffer|TestVellumPointer')
(cd "$VELLUM_REPO_ROOT/examples/replay" && go test -race ./... && go vet ./... && go test -race -tags offscreen ./... && go vet -tags offscreen ./... && mkdir -p build && go build -o build/vellum-workspace ./cmd/workspace && if [ "$(go env GOOS)" = darwin ]; then go build -tags vellum_native_checks -o build/vellum-workspace-native ./cmd/workspace-native; fi)
(cd "$VELLUM_REPO_ROOT/examples/grid" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)
(cd "$VELLUM_REPO_ROOT/examples/loadedgrid" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)
(cd "$VELLUM_REPO_ROOT/examples/loadedtree" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)

(cd "$VELLUM_REPO_ROOT/examples/placement" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)

(cd "$VELLUM_REPO_ROOT/examples/personal-mac" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)

# Forge is an independent consumer; these checks do not open GUI windows.
(cd "$VELLUM_REPO_ROOT/examples/forge" && go test -race -tags offscreen ./... && go vet -tags offscreen ./...)
)

if [ "$VELLUM_CHECK_STAGE" = all ]; then
 for VELLUM_STAGE_NAME in $VELLUM_CHECK_STAGES; do "check_$VELLUM_STAGE_NAME"; done
else
 "check_$VELLUM_CHECK_STAGE"
fi
