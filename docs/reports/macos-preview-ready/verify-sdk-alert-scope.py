"""Verify only the four recorded alert scopes in the exact preview SDK archive."""
import hashlib
import json
import sys
import tarfile
from pathlib import Path, PurePosixPath

archive = Path(sys.argv[1])
expected = "28a95b5c2354673d97161df30e828008a5aa18465ce18545a42e8caf0a82f269"
with archive.open("rb") as stream:
    actual = hashlib.file_digest(stream, "sha256").hexdigest()
if actual != expected:
    raise SystemExit("SDK archive hash does not match the reviewed payload")
def load_json(sdk, member):
    try:
        stream = sdk.extractfile(member)
    except KeyError:
        raise SystemExit("Required inventory member is missing: " + member) from None
    if stream is None:
        raise SystemExit("Required inventory member is not a file: " + member)
    with stream:
        return json.load(stream)


root = "vellum-v0.1.0-alpha.3/"
with tarfile.open(archive, "r:gz") as sdk:
    names = sdk.getnames()
    prototype_paths = [name for name in names if "prototypes" in PurePosixPath(name).parts]
    if prototype_paths:
        raise SystemExit("Unexpected prototype path in SDK")
    modules = load_json(sdk, root + "THIRD_PARTY_INVENTORY.json")["modules"]
    versions = {item["module"]: item["version"] for item in modules}
    selected = {name: versions.get(name) for name in ("golang.org/x/image", "golang.org/x/net")}
    if selected != {"golang.org/x/image": "v0.44.0", "golang.org/x/net": "v0.57.0"}:
        raise SystemExit("Selected versions differ from the alert triage")
    native = load_json(sdk, root + "release/native-components/wgpu-darwin-arm64/INVENTORY.json")
    components = native["components"]
    if len(components) != 82 or any(item["name"] == "rand" for item in components):
        raise SystemExit("Selected native component list differs from the triage")
print(json.dumps({
    "sdk_sha256": actual,
    "archive_members": len(names),
    "prototype_paths": prototype_paths,
    "selected_go_versions": selected,
    "native_component_count": len(components),
    "rand_in_selected_native_components": False,
    "scope": "Archive exclusion and selected inventory versions only; not a security audit or prototype safety claim.",
}, indent=2))
