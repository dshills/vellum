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
python3 scripts/benchmark-native_test.py
go test -race ./...
go vet ./...
cd drivers/cogent
go test ./...
go vet ./...
mkdir -p build
go build -o build/vellum-workbench ./cmd/workbench
# Replacement dependency tests do not run transitively with driver tests.
(cd build/upstream && go test -race ./system ./system/driver/base ./system/driver/desktop ./core -run 'TestComposerPublicationAndWithdrawal|TestLockRechecksAfterWaitingForNativeTeardown|TestSceneShowUsesOwnWindowTitle|TestQueuedCloseStopsLoopAfterComposerWithdrawal|TestNewWindowSizingOwnsContextDuringOtherWindowResize|TestSameWindowStageSizingUsesExistingContext|TestScreenSnapshotDuringCacheRebuild|TestWindowOptionsFixupUsesScreenSnapshot|TestMainQueuePublicationDuringStartupDispatch|TestMainQueueInitializationPreservesIdentityAndCompletion|TestMainThreadRequiredCallbackWaitsForLoopPublication|TestComposerWithdrawnBeforeNativeTeardownEvenWithoutGLFW')

"$VELLUM_REPO_ROOT/scripts/check-example.sh"
