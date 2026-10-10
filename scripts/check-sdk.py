#!/usr/bin/env python3
"""Verify and build a relocated SDK consumer without original-checkout replacements."""
import argparse
import hashlib
import gzip
import io
import json
import os
from pathlib import Path, PurePosixPath
import re
import shutil
import subprocess
import sys

from child_process import run_child, check_child_output
import tarfile
import tempfile
import zlib

from sdk_environment import CHILD_ENVIRONMENT, GO_CONFIGURATION, child_environment


# Bounded verification applies before tarfile parses PAX/long-name metadata or
# extraction allocates disk space. Hash ordinary file bodies in small chunks.
MAX_COMPRESSED_BYTES = 512 * 1024 * 1024
MAX_MEMBER_BYTES = 64 * 1024 * 1024
MAX_EXPANDED_BYTES = 1024 * 1024 * 1024
MAX_MEMBERS = 20_000
MAX_MANIFEST_BYTES = 8 * 1024 * 1024
MAX_METADATA_BYTES = 1024 * 1024
MAX_GO_MOD_BYTES = 1024 * 1024
CHUNK_BYTES = 64 * 1024
COMMAND_TIMEOUT_SECONDS = 600
GENERATOR_TIMEOUT_SECONDS = 60
RUNTIME_TIMEOUT_SECONDS = 30


def validate_pax_path(payload):
    """The SDK producer emits path-only PAX records, never framing overrides."""
    position = 0
    keys = set()
    while position < len(payload):
        separator = payload.find(b" ", position)
        if separator < 0 or not re.fullmatch(rb"[1-9][0-9]{0,8}", payload[position:separator]):
            raise ValueError("invalid SDK PAX record")
        length = int(payload[position:separator])
        end = position + length
        if end > len(payload) or end <= separator + 2 or payload[end - 1:end] != b"\n":
            raise ValueError("invalid SDK PAX record")
        key, equal, value = payload[separator + 1:end - 1].partition(b"=")
        if not equal or key != b"path" or key in keys:
            raise ValueError("unsupported SDK PAX metadata key or duplicate")
        if not value or b"\0" in value:
            raise ValueError("invalid SDK PAX path")
        keys.add(key)
        position = end
    if not keys:
        raise ValueError("empty SDK PAX metadata")


def preflight_tar(archive_bytes):
    """Bound raw framing; accept regular files and one path-only PAX header."""
    count = total = 0
    pending_path = False
    with gzip.GzipFile(fileobj=io.BytesIO(archive_bytes), mode="rb") as stream:
        while True:
            header = stream.read(tarfile.BLOCKSIZE)
            if header == b"\0" * tarfile.BLOCKSIZE:
                if pending_path:
                    raise ValueError("SDK PAX metadata has no following file")
                return
            if len(header) != tarfile.BLOCKSIZE:
                raise ValueError("truncated SDK tar header")
            member = tarfile.TarInfo.frombuf(header, "utf-8", "surrogateescape")
            count += 1
            if count > MAX_MEMBERS:
                raise ValueError("SDK archive member count limit exceeded")
            if member.size < 0 or member.size > MAX_MEMBER_BYTES:
                raise ValueError("SDK archive member size limit exceeded")
            if member.type not in {tarfile.REGTYPE, tarfile.AREGTYPE, tarfile.XHDTYPE}:
                raise ValueError("unsafe SDK archive member type")
            if member.type == tarfile.XHDTYPE and pending_path:
                raise ValueError("SDK PAX header chains are unsupported")
            if member.type == tarfile.XHDTYPE and member.size > MAX_METADATA_BYTES:
                raise ValueError("SDK tar metadata size limit exceeded")
            total += member.size
            if total > MAX_EXPANDED_BYTES:
                raise ValueError("SDK archive expanded size limit exceeded")
            remaining = (member.size + tarfile.BLOCKSIZE - 1) // tarfile.BLOCKSIZE * tarfile.BLOCKSIZE
            metadata = bytearray() if member.type == tarfile.XHDTYPE else None
            while remaining:
                data = stream.read(min(CHUNK_BYTES, remaining))
                if not data:
                    raise ValueError("truncated SDK tar body")
                if metadata is not None:
                    metadata.extend(data)
                remaining -= len(data)
            if metadata is not None:
                validate_pax_path(bytes(metadata[:member.size]))
                pending_path = True
            else:
                pending_path = False


def verify_archive(path, expected_sha256=None):
    with path.open("rb") as source:
        archive_bytes = source.read(MAX_COMPRESSED_BYTES + 1)
    if len(archive_bytes) > MAX_COMPRESSED_BYTES:
        raise ValueError("SDK compressed archive size limit exceeded")
    digest = hashlib.sha256(archive_bytes).hexdigest()
    if expected_sha256 is not None and digest != expected_sha256:
        raise ValueError("SDK archive SHA-256 mismatch")
    preflight_tar(archive_bytes)
    hashes = {}
    manifest = None
    prefix = None
    root_module = None
    total = 0
    with tarfile.open(fileobj=io.BytesIO(archive_bytes), mode="r|gz") as archive:
        for member in archive:
            parts = PurePosixPath(member.name).parts
            if not member.isfile() or "\\" in member.name or len(parts) < 2 or any(part in {"", ".", ".."} for part in member.name.split("/")):
                raise ValueError("unsafe SDK archive member")
            if prefix is None:
                prefix = parts[0]
                if not re.fullmatch(r"vellum-v[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?", prefix):
                    raise ValueError("invalid SDK archive prefix")
            if parts[0] != prefix:
                raise ValueError("unsafe SDK archive member")
            name = member.name[len(prefix) + 1:]
            if name in hashes:
                raise ValueError("duplicate SDK archive member")
            if len(hashes) >= MAX_MEMBERS or member.size < 0 or member.size > MAX_MEMBER_BYTES:
                raise ValueError("SDK archive member count or size limit exceeded")
            total += member.size
            if total > MAX_EXPANDED_BYTES:
                raise ValueError("SDK archive expanded size limit exceeded")
            source = archive.extractfile(member)
            if name == "SDK_MANIFEST.json":
                if member.size > MAX_MANIFEST_BYTES:
                    raise ValueError("SDK manifest size limit exceeded")
                manifest = json.loads(source.read(MAX_MANIFEST_BYTES + 1))
                hashes[name] = None
            else:
                if name == "go.mod" and member.size > MAX_GO_MOD_BYTES:
                    raise ValueError("SDK root go.mod size limit exceeded")
                root_module_bytes = bytearray() if name == "go.mod" else None
                content_hash = hashlib.sha256()
                while data := source.read(CHUNK_BYTES):
                    content_hash.update(data)
                    if root_module_bytes is not None:
                        root_module_bytes.extend(data)
                if root_module_bytes is not None:
                    directives = re.findall(r"^module\s+(\S+)", root_module_bytes.decode("utf-8"), re.MULTILINE)
                    root_module = directives[0].strip('"') if len(directives) == 1 else None
                hashes[name] = content_hash.hexdigest()
    if not isinstance(manifest, dict) or manifest.get("schema") != 1 or not isinstance(manifest.get("release"), dict) or prefix != "vellum-" + str(manifest["release"].get("version")):
        raise ValueError("invalid SDK manifest")
    module_path = manifest["release"].get("module")
    if not isinstance(module_path, str) or len(module_path) > 512 or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._~-]*(?:/[A-Za-z0-9][A-Za-z0-9._~-]*)+", module_path):
        raise ValueError("invalid SDK manifest module path")
    if root_module != module_path:
        raise ValueError("SDK manifest module differs from archived root go.mod")
    files = manifest.get("files")
    if not isinstance(files, dict) or "SDK_MANIFEST.json" in files:
        raise ValueError("invalid SDK checksum records")
    if set(hashes) - {"SDK_MANIFEST.json"} != set(files):
        raise ValueError("SDK manifest/member coverage mismatch")
    for name, expected in files.items():
        if not isinstance(expected, str) or not re.fullmatch(r"[0-9a-f]{64}", expected):
            raise ValueError("invalid SDK member checksum")
        if hashes[name] != expected:
            raise ValueError(f"SDK member checksum mismatch: {name}")
    return {"archive_sha256": digest, "prefix": prefix, "manifest": manifest, "_archive_bytes": archive_bytes}


def extract_verified(path, destination, verified):
    # Create each verified regular file exclusively; tarfile.extractall is never
    # used and links, traversal paths, duplicates and special files were rejected.
    if destination.exists() or destination.is_symlink():
        raise ValueError("SDK extraction destination exists")
    destination.mkdir()
    try:
        with tarfile.open(fileobj=io.BytesIO(verified["_archive_bytes"]), mode="r:gz") as archive:
            for member in archive:
                target = destination / member.name
                target.parent.mkdir(parents=True, exist_ok=True)
                with target.open("xb") as output:
                    shutil.copyfileobj(archive.extractfile(member), output)
                target.chmod(0o755 if member.mode & 0o111 else 0o644)
        return destination / verified["prefix"]
    except BaseException:
        shutil.rmtree(destination)
        raise


def decode_objects(raw):
    decoder = json.JSONDecoder()
    objects = []
    rest = raw.strip()
    while rest:
        value, end = decoder.raw_decode(rest)
        objects.append(value)
        rest = rest[end:].lstrip()
    return objects


def check_origin(modules, sdk, module_path="github.com/dshills/vellum"):
    origins = {}
    for module in modules:
        replaced = module.get("Replace")
        if replaced is None:
            continue
        if not replaced.get("Dir") or not Path(replaced["Dir"]).is_absolute():
            raise ValueError(f"consumer replacement has no absolute local directory: {module['Path']}")
        directory = Path(replaced["Dir"]).resolve()
        try:
            origins[module["Path"]] = directory.relative_to(sdk.resolve()).as_posix()
        except ValueError:
            raise ValueError(f"consumer replacement is outside relocated SDK: {module['Path']}")
    expected = {module_path: ".", module_path + "/drivers/cogent": "drivers/cogent",
                "cogentcore.org/core": "drivers/cogent/build/upstream",
                "github.com/go-text/typesetting": "drivers/cogent/build/typesetting"}
    for name, relative in expected.items():
        if name not in origins:
            raise ValueError(f"consumer is missing explicit SDK replacement: {name}")
        if origins[name] != relative:
            raise ValueError(f"consumer SDK replacement has wrong directory: {name}; expected {relative}")
    return origins



def check(path, expected_sha256=None, native_runtime=False, allow_self_attested=False):
    if expected_sha256 is None and not allow_self_attested:
        raise ValueError("executing SDK code requires --sha256 from a trusted producer or explicit --allow-self-attested")
    verified = verify_archive(path, expected_sha256)
    env = child_environment()
    with tempfile.TemporaryDirectory(prefix="vellum SDK relocation ") as temporary:
        base = Path(temporary)
        sdk = extract_verified(path, base / "extracted SDK", verified)
        consumer = base / "consumer application"
        run_child(["python3", str(sdk / "scripts/new-app.py"), str(consumer), "--module=example.com/vellum-relocation-consumer"], env=env, check=True, stdout=sys.stderr, timeout=GENERATOR_TIMEOUT_SECONDS)
        modules = decode_objects(check_child_output(["go", "list", "-m", "-json", "all"], cwd=consumer, env=env, text=True, timeout=COMMAND_TIMEOUT_SECONDS))
        origins = check_origin(modules, sdk, verified["manifest"]["release"]["module"])
        run_child(["go", "mod", "verify"], cwd=consumer, env=env, check=True, stdout=sys.stderr, timeout=COMMAND_TIMEOUT_SECONDS)
        for arguments in (["go", "test", "-race", "./..."], ["go", "vet", "./..."], ["go", "build", "-o", str(base / "consumer"), "."]):
            run_child(arguments, cwd=consumer, env=env, check=True, stdout=sys.stderr, timeout=COMMAND_TIMEOUT_SECONDS)
        contract = sdk / "examples/sdk-contract"
        for arguments in (["go", "test", "-race", "./..."], ["go", "vet", "./..."]):
            run_child(arguments, cwd=contract, env=env, check=True, stdout=sys.stderr, timeout=COMMAND_TIMEOUT_SECONDS)
        native = None
        if native_runtime:
            completed = run_child([str(base / "consumer"), "--benchmark", "--idle=1s", "--samples=3"], env=child_environment(runtime=True), check=True, text=True, capture_output=True, timeout=RUNTIME_TIMEOUT_SECONDS)
            reports = []
            for line in completed.stdout.splitlines():
                try:
                    value = json.loads(line)
                except json.JSONDecodeError:
                    continue
                if isinstance(value,dict) and value.get("event") == "vellum.native.ready":
                    reports.append(value)
            if len(reports) != 1:
                raise ValueError("relocated consumer failed to emit exactly one verified GPU report")
            native = reports[0]
        return {"event": "vellum.sdk.relocated", "digest_verified_against_expected": expected_sha256 is not None, "trust_basis": "caller-supplied expected digest" if expected_sha256 is not None else "explicit self-attestation", "independent_authenticity_verified": False, "archive_sha256": verified["archive_sha256"], "version": verified["manifest"]["release"]["version"], "source": verified["manifest"].get("source"), "verified_files": len(verified["manifest"]["files"]), "replacements_within_sdk": origins, "path_contains_spaces": " " in str(sdk), "checks": ["manifest all-file integrity", "external Go module cache verification", "standalone consumer race/vet/build", "public API scenarios race/vet"], "native_runtime": native if native_runtime else "not executed", "public_distribution_audit": "not certified"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--archive", required=True, type=Path)
    parser.add_argument("--sha256", help="trusted archive digest, if available")
    parser.add_argument("--allow-self-attested", action="store_true", help="explicitly execute archive code without an independently trusted digest (for own private builds)")
    parser.add_argument("--runtime", action="store_true", help="also open/verify/close the relocated native consumer (serialize desktop checks)")
    args = parser.parse_args()
    try:
        result = check(args.archive.absolute(), args.sha256, args.runtime, args.allow_self_attested)
    except subprocess.TimeoutExpired as error:
        print(json.dumps({"event": "vellum.sdk.check_failed", "reason": "child command timed out", "timeout_seconds": error.timeout}), file=sys.stderr)
        raise SystemExit(1)
    except (ValueError, OSError, EOFError, tarfile.TarError, zlib.error, subprocess.CalledProcessError) as error:
        print(json.dumps({"event": "vellum.sdk.check_failed", "reason": str(error), "error_type": type(error).__name__}), file=sys.stderr)
        raise SystemExit(1)
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
