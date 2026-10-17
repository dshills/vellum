"""Private, digest-pinned Mac GPU SDK overlays; provenance is caller-declared data."""
import hashlib
import json
import math
from pathlib import Path, PurePosixPath
import posixpath
import re
import struct

from native_archive import archive_members, regular_stream, unchanged

MODULE = "github.com/oliverbestmann/webgpu/libs-darwin"
DIRECTORY = "drivers/cogent/build/libs-darwin"
LIBRARY = "arm64/libwgpu_native.a"
PROVENANCE = "PROVENANCE.json"
MAX_LIBRARY_BYTES = 64 << 20
MAX_PROVENANCE_BYTES = 1 << 20
MAX_MANIFEST_BYTES = 8 << 20
GO_MOD = ("module " + MODULE + "\n").encode()
LIBS_GO = b'''package lib

/*
#cgo darwin,!ios,amd64 LDFLAGS: -L${SRCDIR}/amd64 -lwgpu_native
#cgo darwin,!ios,arm64 LDFLAGS: -L${SRCDIR}/arm64 -lwgpu_native
#cgo darwin LDFLAGS: -framework Metal -framework QuartzCore
*/
import "C"
'''
TRUST = "caller-supplied-candidate-not-attested"
FILES = {"go.mod", "libs.go", LIBRARY, PROVENANCE}


def digest(data):
    return hashlib.sha256(data).hexdigest()


def sha_value(value):
    return type(value) is str and re.fullmatch(r"[0-9a-f]{64}", value) is not None


def read_regular(path, bound):
    stream, before = regular_stream(path, bound)
    with stream:
        data = stream.read(bound + 1)
        unchanged(stream, before)
    if len(data) != before.st_size or len(data) > bound:
        raise ValueError("native GPU input changed or exceeds bound")
    return data


def strict_json(data):
    def unique(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise ValueError("duplicate native GPU JSON key")
            result[key] = value
        return result

    def invalid(value):
        raise ValueError("nonfinite native GPU JSON value")

    def finite_float(value):
        parsed = float(value)
        if not math.isfinite(parsed):
            invalid(value)
        return parsed
    return json.loads(data, object_pairs_hook=unique, parse_constant=invalid, parse_float=finite_float)


def arm64_objects(data, members):
    """Observe thin ARM64 MH_OBJECT headers; not complete Mach-O validation."""
    offset, count = 8, 0
    indices = {"/", "__.SYMDEF", "__.SYMDEF SORTED", "__.SYMDEF_64", "__.SYMDEF_64 SORTED"}
    for member in members:
        header = data[offset:offset + 60]
        size = int(header[48:58])
        name = header[:16].rstrip(b" ")
        name_size = int(name[3:]) if name.startswith(b"#1/") else 0
        payload = data[offset + 60 + name_size:offset + 60 + size]
        if member["name"] not in indices:
            if len(payload) < 32 or payload[:4] != b"\xcf\xfa\xed\xfe" or struct.unpack_from("<I", payload, 4)[0] != 0x0100000C or struct.unpack_from("<I", payload, 12)[0] != 1:
                raise ValueError("native GPU archive requires only thin ARM64 Mach-O objects")
            count += 1
        offset += 60 + size + size % 2
    if not count or offset != len(data):
        raise ValueError("native GPU archive requires nonempty exact object coverage")
    return count


def candidate(library, expected_sha256, provenance):
    if not sha_value(expected_sha256):
        raise ValueError("native GPU archive requires a lowercase SHA-256")
    data = read_regular(library, MAX_LIBRARY_BYTES)
    if digest(data) != expected_sha256:
        raise ValueError("native GPU archive SHA-256 mismatch")
    framed = archive_members(library)
    if framed["sha256"] != expected_sha256:
        raise ValueError("native GPU archive changed between observations")
    objects = arm64_objects(data, framed["members"])
    declaration = read_regular(provenance, MAX_PROVENANCE_BYTES)
    declared = strict_json(declaration)
    if type(declared) is not dict or type(declared.get("schema")) is not int or declared["schema"] != 1 or declared.get("target") != "darwin-arm64" or declared.get("artifact_sha256") != expected_sha256:
        raise ValueError("native GPU provenance must identify schema, target and exact artifact")
    generated = {"go.mod": GO_MOD, "libs.go": LIBS_GO, LIBRARY: data, PROVENANCE: declaration}
    record = {"schema": 1, "target": "darwin-arm64", "module": MODULE, "directory": DIRECTORY,
              "trust": TRUST, "distribution_audit": "pending", "arm64_object_members": objects,
              "files": {name: digest(body) for name, body in sorted(generated.items())},
              "go_mod_overlays": {}}
    return record, {DIRECTORY + "/" + name: body for name, body in generated.items()}



COMPONENT_DIRECTORY = "release/native-components/wgpu-darwin-arm64"
COMPONENT_INVENTORY = COMPONENT_DIRECTORY + "/INVENTORY.json"
KNOWN_COMPONENT_ARCHIVE = "786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883"
MAX_COMPONENT_BYTES = 4 << 20
MAX_COMPONENT_TOTAL = 16 << 20
MAX_COMPONENT_FILES = 512


def component_notice_record(record, files):
    """Check exact SDK member coverage; declarations never grant approval."""
    if type(record) is not dict or set(record) != {"inventory", "files", "components", "distribution_audit"} or record.get("inventory") != COMPONENT_INVENTORY or record.get("distribution_audit") != "pending" or type(record.get("components")) is not int or not 0 < record["components"] <= 256:
        raise ValueError("invalid native component notice record")
    hashes = record.get("files")
    if type(hashes) is not dict or not 1 < len(hashes) <= MAX_COMPONENT_FILES or COMPONENT_INVENTORY not in hashes:
        raise ValueError("invalid native component notice coverage")
    for name, checksum in hashes.items():
        if type(name) is not str or (name != COMPONENT_INVENTORY and not name.startswith(COMPONENT_DIRECTORY + "/NOTICES/")) or "\\" in name or any(part in {"", ".", ".."} for part in name.split("/")) or not sha_value(checksum) or files.get(name) != checksum:
            raise ValueError("native component notice checksum or path mismatch")
    if {name for name in files if name.startswith(COMPONENT_DIRECTORY + "/")} != set(hashes):
        raise ValueError("native component notice member coverage mismatch")
    return record


def component_notices(root, artifact_sha256):
    """Bind checked-in declarations to this exact private archive, when known."""
    root = Path(root)
    path = root / COMPONENT_INVENTORY
    # Refuse symlinked ancestors even if a redirected final path is absent.
    current = root
    for part in PurePosixPath(COMPONENT_INVENTORY).parts:
        current /= part
        if current.is_symlink():
            raise ValueError("native component notices cannot contain symlinks")
    if not path.exists():
        if artifact_sha256 == KNOWN_COMPONENT_ARCHIVE:
            raise ValueError("known native archive requires its component notice inventory")
        return None
    data = read_regular(path, MAX_COMPONENT_BYTES)
    inventory = strict_json(data)
    if type(inventory) is not dict or type(inventory.get("schema")) is not int or inventory["schema"] != 1 or not sha_value(inventory.get("artifact_sha256")) or inventory.get("distribution_audit") != "pending":
        raise ValueError("invalid native component inventory")
    if inventory["artifact_sha256"] != artifact_sha256:
        if artifact_sha256 == KNOWN_COMPONENT_ARCHIVE:
            raise ValueError("known native archive has a mismatched component notice inventory")
        return None  # Never apply one archive's declarations to another binary.
    hashes, components = inventory.get("files"), inventory.get("components")
    if type(hashes) is not dict or type(components) is not list or not 0 < len(components) <= 256:
        raise ValueError("invalid native component inventory coverage")
    expected = {COMPONENT_DIRECTORY + "/" + name: value for name, value in hashes.items()}
    expected[COMPONENT_INVENTORY] = digest(data)
    record = {"inventory": COMPONENT_INVENTORY, "files": expected, "components": len(components), "distribution_audit": "pending"}
    component_notice_record(record, expected)
    referenced, identities, total = set(), set(), len(data)
    if total > MAX_COMPONENT_TOTAL:
        raise ValueError("native component inventory exceeds aggregate byte bound")
    for component in components:
        if type(component) is not dict or type(component.get("id")) is not str or not component["id"] or component["id"] in identities or type(component.get("notice_files")) is not list or not component["notice_files"]:
            raise ValueError("invalid native component identity or notices")
        identities.add(component["id"])
        for name in component["notice_files"]:
            if type(name) is not str or name not in hashes:
                raise ValueError("native component refers to missing notice")
            referenced.add(name)
    if referenced != set(hashes):
        raise ValueError("native component notice is unreferenced")
    for name, checksum in expected.items():
        current = root
        for part in PurePosixPath(name).parts:
            current /= part
            if current.is_symlink():
                raise ValueError("native component notices cannot contain symlinks")
        bound = MAX_COMPONENT_BYTES if name == COMPONENT_INVENTORY else min(MAX_COMPONENT_BYTES, MAX_COMPONENT_TOTAL - total)
        content = read_regular(current, bound)
        total += len(content) if name != COMPONENT_INVENTORY else 0
        if total > MAX_COMPONENT_TOTAL or digest(content) != checksum:
            raise ValueError("native component notice checksum mismatch or byte bound exceeded")
    return record


def overlay_modfile(name, data):
    if name != "drivers/cogent/go.mod" and not (name.startswith("examples/") and name.endswith("/go.mod") and
            re.search(rb"^\s*(?:require\s+)?github\.com/dshills/vellum/drivers/cogent\s+v\S+", data, re.MULTILINE)):
        return None
    if re.search(rb"^\s*(?:replace\s+)?github\.com/oliverbestmann/webgpu/libs-darwin(?:\s+\S+)?\s+=>", data, re.MULTILINE):
        raise ValueError("native GPU module already has a replacement: " + name)
    relative = posixpath.relpath(DIRECTORY, str(PurePosixPath(name).parent))
    if not relative.startswith("."):
        relative = "./" + relative
    return data + ("\nreplace " + MODULE + " => " + relative + "\n").encode()


def validate_record(record, files, private_build):
    if type(files) is not dict:
        raise ValueError("invalid native GPU SDK file map")
    capsule = {name for name in files if name.startswith(DIRECTORY + "/")}
    if record is None:
        if capsule:
            raise ValueError("native GPU capsule has no SDK manifest record")
        return None
    if type(record) is not dict or type(record.get("schema")) is not int or record["schema"] != 1 or private_build is not True or record.get("target") != "darwin-arm64" or record.get("module") != MODULE or record.get("directory") != DIRECTORY or record.get("trust") != TRUST or record.get("distribution_audit") != "pending":
        raise ValueError("invalid private native GPU manifest record")
    hashes = record.get("files")
    if type(hashes) is not dict or set(hashes) != FILES or capsule != {DIRECTORY + "/" + name for name in FILES} or any(not sha_value(value) or files.get(DIRECTORY + "/" + name) != value for name, value in hashes.items()):
        raise ValueError("native GPU manifest/member coverage mismatch")
    if hashes["go.mod"] != digest(GO_MOD) or hashes["libs.go"] != digest(LIBS_GO):
        raise ValueError("native GPU linkage differs from owned module")
    if "component_notices" in record:
        component_notice_record(record["component_notices"], files)
    objects = record.get("arm64_object_members")
    if type(objects) is not int or not 0 < objects < 10000:
        raise ValueError("invalid native GPU object count")
    overlays = record.get("go_mod_overlays")
    if type(overlays) is not dict or "drivers/cogent/go.mod" not in overlays:
        raise ValueError("native GPU driver overlay missing")
    for name, item in overlays.items():
        if type(name) is not str or name != "drivers/cogent/go.mod" and not (name.startswith("examples/") and name.endswith("/go.mod") and all(part not in {"", ".", ".."} for part in name.split("/"))) or type(item) is not dict or set(item) != {"source_sha256", "archived_sha256"} or not all(sha_value(value) for value in item.values()) or files.get(name) != item["archived_sha256"]:
            raise ValueError("invalid native GPU module overlay record")
    return record


def sdk_candidate(root):
    manifest_path = root / "SDK_MANIFEST.json"
    if not manifest_path.exists() and not manifest_path.is_symlink():
        return None
    manifest = strict_json(read_regular(manifest_path, MAX_MANIFEST_BYTES))
    if type(manifest) is not dict:
        raise ValueError("invalid native GPU SDK manifest")
    record = validate_record(manifest.get("native_gpu"), manifest.get("files", {}), manifest.get("private_build"))
    if record is None and ((root / DIRECTORY).exists() or (root / DIRECTORY).is_symlink()):
        raise ValueError("native GPU SDK capsule has no manifest record")
    if record is not None:
        expected = {DIRECTORY + "/" + name: value for name, value in record["files"].items()}
        expected.update({name: value["archived_sha256"] for name, value in record["go_mod_overlays"].items()})
        for name, checksum in expected.items():
            path = root
            for part in PurePosixPath(name).parts:
                path /= part
                if path.is_symlink():
                    raise ValueError("native GPU SDK paths cannot contain symlinks")
            bound = MAX_LIBRARY_BYTES if name == DIRECTORY + "/" + LIBRARY else MAX_COMPONENT_BYTES if name.startswith(COMPONENT_DIRECTORY + "/") else MAX_PROVENANCE_BYTES
            if digest(read_regular(path, bound)) != checksum:
                raise ValueError("native GPU SDK member checksum mismatch")
        if "component_notices" in record:
            observed = component_notices(root, record["files"][LIBRARY])
            if observed != record["component_notices"]:
                raise ValueError("native component inventory does not match selected archive or record")
        # Refuse on the first unexpected entry; never traverse a tampered tree.
        for directory, allowed in [(root / DIRECTORY, {"go.mod", "libs.go", PROVENANCE, "arm64"}),
                                   (root / DIRECTORY / "arm64", {"libwgpu_native.a"})]:
            seen = set()
            for path in directory.iterdir():
                if path.name not in allowed or path.is_symlink():
                    raise ValueError("native GPU capsule contains extra or missing files")
                if not (path.is_dir() if path.name == "arm64" else path.is_file()):
                    raise ValueError("native GPU capsule requires regular files and directories")
                seen.add(path.name)
            if seen != allowed:
                raise ValueError("native GPU capsule contains extra or missing files")
    return record
