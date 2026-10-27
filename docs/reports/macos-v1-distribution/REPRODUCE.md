# Reconstruct inputs and produce public candidates after approval

This recipe is a later-stage handoff, not authorization to publish or change the
audit. Current metadata intentionally refuses the public commands. Product code
is frozen at v1.0.0 / 4b42fb8; the existing tag must not move. Record the actual
approval/packaging revision separately. No UI, backend or performance changes are
needed for these commands.

## Retained private material

Use the local delivery directory named by the packet. Verify its `SHA256SUMS`.
Extract PRE_SIGN_INPUTS.tar.gz into a new local stage and the SDK candidate into
a separate directory. Preserve the original archives. The pre-sign input archive
contains apps/<id>/binary, apps/<id>/INVENTORY.json, a shared-notices/NOTICES pool
and icons/{forma,luma,pulse}.png. Candidate JSON binds all binary/inventory/icon
hashes; each extracted app must be reconstructed with its own exact notice list.
Do not copy the entire shared pool into an individual notice bundle.

From the repository, with STAGE set to that new extracted input directory:

```sh
python3 - "$STAGE" <<'PYCODE'
from pathlib import Path
import json, shutil, sys
sys.path.insert(0, 'scripts')
import app_notices
stage = Path(sys.argv[1])
for app in ('notes', 'widgets', 'shell', 'forma', 'luma', 'pulse'):
    inputs = stage / 'apps' / app
    inventory = json.loads((inputs / 'INVENTORY.json').read_text())
    output = inputs / 'notices'
    output.mkdir()  # new output only
    shutil.copyfile(inputs / 'INVENTORY.json', output / 'INVENTORY.json')
    for name in inventory['files']:
        app_notices.notice_name(name)
        target = output / name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(stage / 'shared-notices' / name, target)
    app_notices.load(output, inventory['artifact_sha256'])
PYCODE
```

## Existing guarded producers

Only after the owner completes the current-payload audit decision, record that
approval and commit it using the existing release policy. Keep all original
component declaration labels intact; their pending/review-required labels are
historical collection records, not authorization switches. Do not edit the old
alpha.3 approval or remove any producer checks.

Use the pinned Go version, prepared checkout and existing driver environment.
SOURCE_SDK below means the independently extracted audit SDK. These paths name
new outputs; no output may overwrite a previous artifact.

```sh
python3 scripts/release-sdk.py --public   --output "$OUT/vellum-v1.0.0-darwin-arm64-sdk.tar.gz"   --macos-gpu-library "$SOURCE_SDK/drivers/cogent/build/libs-darwin/arm64/libwgpu_native.a"   --macos-gpu-sha256 786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883   --macos-gpu-provenance "$SOURCE_SDK/drivers/cogent/build/libs-darwin/PROVENANCE.json"
```

Example for Forma, **if the disclosed ad-hoc/unnotarized distribution route is
chosen**. Use a separately authorized signing pipeline for Developer ID and
notarization; this command does not provide either.

```sh
python3 scripts/package-app.py --public   --binary "$STAGE/apps/forma/binary"   --notice-bundle "$STAGE/apps/forma/notices"   --icon "$STAGE/icons/forma.png"   --output "$OUT/Forma.app" --name Forma   --identifier com.dshills.vellum.showcase.forma   --version 1.0.0 --build-version 10 --sign-identity -
```

Repeat with the other five identities in candidate-packet.json. Original showcase
icons are retained; existing private-only examples/showcase/package.py is not the
public producer. Use the general guarded producer above with explicit notices.

## Final publication packet

The final public SDK contains approval/source metadata that differs from these
private audit bytes. App public manifests and any chosen signatures may also
change. Preserve code/dependency/native/asset equivalence and record exact final
source, binary, notices and archive hashes. Recheck independent SDK consumption,
ZIP extraction/signatures and a bounded native launch/close on the final apps.
Existing CI and interaction evidence retain their documented scopes. Then prepare
final release notes and obtain explicit publication approval for those exact
assets. The current private SHA256SUMS and blocked draft are not upload inputs.
