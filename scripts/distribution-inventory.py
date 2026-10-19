#!/usr/bin/env python3
"""Record Mac SDK files, Go build assets and native link candidates without launching code."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path, PurePosixPath
import re
import tempfile
import struct
import sys

if __name__ == "__main__":
    sys.dont_write_bytecode = True

import native_gpu

from native_archive import archive_members, linker_objects, selected_archive_members, regular_stream, unchanged

from artifact_io import publish_exclusive
from child_process import check_child_output
from sdk_environment import child_environment

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("vellum_release_inventory", ROOT / "scripts/release-sdk.py")
release = importlib.util.module_from_spec(spec)
spec.loader.exec_module(release)
MAX_FILES = 20000
MAX_FILE_BYTES = 1024 * 1024 * 1024
COMMAND_TIMEOUT_SECONDS = 600
FONT_SUFFIXES = {".ttf", ".otf", ".ttc", ".woff", ".woff2"}
NATIVE_SUFFIXES = {".a", ".dylib", ".so", ".dll", ".lib", ".o"}


SDK_MANIFEST = "SDK_MANIFEST.json"
MAX_SDK_MANIFEST_BYTES = 8 << 20
MAX_SDK_MEMBER_BYTES = 64 << 20
MAX_SDK_TOTAL_BYTES = 1 << 30


def sdk_source_snapshot(root, include_untracked=False):
    """Observe a complete immutable SDK payload; its provenance is self-described."""
    manifest_path = root / SDK_MANIFEST
    if not manifest_path.exists() and not manifest_path.is_symlink():
        return None
    if include_untracked:
        raise ValueError("SDK payload inventory does not accept include_untracked")
    raw = native_gpu.read_regular(manifest_path, MAX_SDK_MANIFEST_BYTES)
    manifest = native_gpu.strict_json(raw)
    if (type(manifest) is not dict or type(manifest.get("schema")) is not int
            or manifest["schema"] != 1 or type(manifest.get("release")) is not dict):
        raise ValueError("invalid SDK source manifest")
    module = manifest["release"].get("module")
    version = manifest["release"].get("version")
    if (type(module) is not str or len(module) > 512
            or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._~-]*(?:/[A-Za-z0-9][A-Za-z0-9._~-]*)+", module)
            or type(version) is not str
            or not re.fullmatch(r"v[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?", version)):
        raise ValueError("invalid SDK source module/version")
    declared = manifest.get("source")
    if (type(declared) is not dict or set(declared) != {"commit", "clean"}
            or type(declared.get("commit")) is not str
            or not re.fullmatch(r"(?:[0-9a-f]{40}|[0-9a-f]{64})", declared["commit"])
            or type(declared.get("clean")) is not bool):
        raise ValueError("invalid declared SDK source provenance")
    files = manifest.get("files")
    if type(files) is not dict or not 0 < len(files) < MAX_FILES:
        raise ValueError("SDK source file-count bound exceeded or invalid records")
    if SDK_MANIFEST in files or "go.mod" not in files:
        raise ValueError("invalid SDK source manifest membership")
    for name, checksum in files.items():
        if (type(name) is not str or not name or len(name) > 4096 or "\\" in name
                or any(ord(c) < 32 or ord(c) == 127 for c in name)
                or len(name.split("/")) > 32 or any(part in {"", ".", ".."} for part in name.split("/"))
                or PurePosixPath(name).is_absolute() or not native_gpu.sha_value(checksum)):
            raise ValueError("invalid SDK source path/checksum")
    native_gpu.validate_record(manifest.get("native_gpu"), files, manifest.get("private_build"))
    expected = {**files, SDK_MANIFEST: hashlib.sha256(raw).hexdigest()}
    # Enumerate only declared directories, refusing extras before traversing them.
    directories = {str(parent) for name in expected for parent in PurePosixPath(name).parents
                   if str(parent) != "."}
    pending, observed = [(root, "")], set()
    while pending:
        directory, prefix = pending.pop()
        with os.scandir(directory) as entries:
            for entry in entries:
                name = prefix + entry.name
                if entry.is_symlink():
                    raise ValueError("SDK source cannot contain symlinks")
                if name in expected and entry.is_file(follow_symlinks=False):
                    observed.add(name)
                elif name in directories and entry.is_dir(follow_symlinks=False):
                    pending.append((Path(entry.path), name + "/"))
                else:
                    raise ValueError("SDK source member coverage mismatch: " + name)
    if observed != set(expected):
        raise ValueError("SDK source member coverage mismatch")
    root_module = native_gpu.read_regular(root / "go.mod", 1 << 20).decode("utf-8")
    directives = re.findall(r"^module\s+(\S+)", root_module, re.MULTILINE)
    if len(directives) != 1 or directives[0].strip('"') != module:
        raise ValueError("SDK source module differs from root go.mod")
    total = 0
    for name, checksum in sorted(expected.items()):
        bound = MAX_SDK_MANIFEST_BYTES if name == SDK_MANIFEST else MAX_SDK_MEMBER_BYTES
        path = root / name
        release.require_regular_source(root, path)
        stream, before = regular_stream(path, bound)
        with stream:
            total += before.st_size
            if total > MAX_SDK_TOTAL_BYTES:
                raise ValueError("SDK source aggregate byte bound exceeded")
            digest, size = hashlib.sha256(), 0
            remaining = before.st_size
            while remaining:
                chunk = stream.read(min(65536, remaining))
                if not chunk:
                    raise ValueError("SDK source shortened during observation")
                digest.update(chunk); size += len(chunk); remaining -= len(chunk)
            if stream.read(1):
                raise ValueError("SDK source grew beyond admitted byte bound")
            unchanged(stream, before)
            current = os.stat(path, follow_symlinks=False)
            if (current.st_dev, current.st_ino) != (before.st_dev, before.st_ino):
                raise ValueError("SDK source path changed during observation")
        if size != before.st_size or digest.hexdigest() != checksum:
            raise ValueError("SDK source checksum mismatch: " + name)
    state = {"kind": "sdk-payload", "declared_source": declared,
             "sdk_manifest_sha256": expected[SDK_MANIFEST],
             "integrity": "observed payload matches manifest checksums",
             "authentication": "not independently verified"}
    return state, [(name, root / name) for name in sorted(expected)]


def source_snapshot(root, include_untracked=False):
    sdk = sdk_source_snapshot(root, include_untracked)
    if sdk is not None:
        return sdk
    top = Path(release.git_output(root, ["rev-parse", "--show-toplevel"], text=True).strip()).resolve()
    if top != root:
        raise ValueError("inventory root must be a Git checkout root or a complete SDK payload")
    return release.source_state(root), list(release.source_files(root, include_untracked=include_untracked))


def file_identity(base, name):
    """Reject escaping/nonregular inputs and stream bounded hashes."""
    relative = PurePosixPath(name)
    if not name or relative.is_absolute() or ".." in relative.parts or str(relative) != name:
        raise ValueError("invalid inventory source path: " + name)
    base = base.resolve()
    source = base.joinpath(*relative.parts)
    release.require_regular_source(base, source)
    digest, size = hashlib.sha256(), 0
    with source.open("rb") as stream:
        while chunk := stream.read(1024 * 1024):
            size += len(chunk)
            if size > MAX_FILE_BYTES:
                raise ValueError("inventory source exceeds file bound: " + name)
            digest.update(chunk)
    return {"file": name, "bytes": size, "sha256": digest.hexdigest()}


def nearby_notices(name, module):
    parents = set(PurePosixPath(name).parents)
    return [item["file"] for item in module.get("license_files", [])
            if PurePosixPath(item["file"]).parent in parents]


def normalize_flag(flag, roots):
    # Replace only actual path prefixes, longest first; no substring matching.
    prefix = flag[:2] if flag.startswith(("-L", "-F")) and len(flag) > 2 else ""
    value = flag[len(prefix):]
    if not value.startswith("/"):
        return flag
    path = Path(value).resolve()
    for module, root in sorted(roots.items(), key=lambda pair: len(str(pair[1])), reverse=True):
        try:
            relative = path.relative_to(root)
        except ValueError:
            continue
        return prefix + "module:" + module + "/" + relative.as_posix()
    return flag


def build_metadata(packages):
    """Compare build inputs, excluding Go's changing cache/staleness fields."""
    fields = ("Standard", "ImportPath", "Dir", "Imports", "GoFiles", "CgoFiles", "CFiles", "CXXFiles", "MFiles", "HFiles", "FFiles", "SFiles", "SwigFiles", "SwigCXXFiles", "SysoFiles", "EmbedPatterns", "EmbedFiles", "CgoCFLAGS", "CgoCPPFLAGS", "CgoCXXFLAGS", "CgoFFLAGS", "CgoLDFLAGS", "Error", "DepsErrors", "Incomplete")
    out = []
    for package in sorted(packages, key=lambda value: value["ImportPath"]):
        record = {key: package[key] for key in fields if key in package}
        module = package.get("Module", {})
        record["Module"] = {key: module[key] for key in ("Path", "Version", "Sum", "Dir", "Replace") if key in module}
        out.append(record)
    return out


STANDARD_SOURCE_FIELDS = ("GoFiles", "CgoFiles", "CFiles", "CXXFiles", "MFiles",
                          "HFiles", "FFiles", "SFiles", "SwigFiles", "SwigCXXFiles",
                          "SysoFiles", "EmbedFiles")
TOOLCHAIN_DECLARATIONS = ("LICENSE", "COPYING", "NOTICE", "PATENTS")


def standard_library_inventory(packages, toolchain_root):
    """Hash Go-selected sources; nearby declarations do not establish rights.

    This excludes compiler-generated inputs, transitive includes, external
    headers, compiler/linker executables and proof of code surviving linking.
    Paths are GOROOT-relative so equal installations compare after relocation.
    """
    toolchain_root = toolchain_root.resolve()
    result, files, declarations, seen = [], {}, {}, set()
    selected_count = 0

    def admit(name, destination):
        if name not in destination:
            if len(files) + len(declarations) == MAX_FILES:
                raise ValueError("standard-library inventory exceeds file-count bound")
            destination[name] = file_identity(toolchain_root, name)

    for package in sorted(packages, key=lambda value: value["ImportPath"]):
        if "Standard" in package and type(package["Standard"]) is not bool:
            raise ValueError("invalid standard-library package marker")
        if not package.get("Standard"):
            continue
        name = package["ImportPath"]
        if (not name or name in seen or len(seen) == MAX_FILES or package.get("Error")
                or package.get("DepsErrors") or package.get("Incomplete")):
            raise ValueError("duplicate, incomplete or over-budget standard-library package")
        seen.add(name)
        directory = Path(package["Dir"])
        # macOS may spell the same installation through /var or /private/var,
        # or a GOROOT alias. Resolve only the installation prefix; descendant
        # components remain lexical so internal source symlinks are refused.
        # The outermost match prevents an internal link back to GOROOT from
        # masquerading as another installation prefix.
        installation = next((parent for parent in reversed([directory, *directory.parents])
                             if parent.resolve() == toolchain_root), None)
        if installation is None:
            raise ValueError("standard-library package outside selected GOROOT")
        relative = directory.relative_to(installation)
        if (not directory.is_absolute() or not relative.parts or relative.parts[0] != "src"
                or ".." in relative.parts or "\\" in str(relative)):
            raise ValueError("standard-library package outside GOROOT/src")
        # Refuse a redirected package directory before enumerating any inputs.
        current = toolchain_root
        for part in relative.parts:
            current /= part
            if current.is_symlink():
                raise ValueError("standard-library source cannot contain symlinks")
        if not current.is_dir():
            raise ValueError("standard-library package directory is missing")
        for field in STANDARD_SOURCE_FIELDS:
            values = package.get(field, [])
            if type(values) is not list or any(type(value) is not str for value in values):
                raise ValueError("invalid standard-library source list")
        selected = sum(len(package.get(field, [])) for field in STANDARD_SOURCE_FIELDS)
        if selected > MAX_FILES - selected_count:
            raise ValueError("standard-library inventory exceeds file-count bound")
        selected_count += selected
        sources = {}
        for field in STANDARD_SOURCE_FIELDS:
            values = package.get(field, [])
            if len(set(values)) != len(values):
                raise ValueError("duplicate standard-library source")
            sources[field] = []
            for filename in sorted(values):
                # Validate the unjoined spelling too: Path would normalize a//b.
                path = PurePosixPath(filename)
                if (not filename or path.is_absolute() or ".." in path.parts
                        or str(path) != filename or "\\" in filename):
                    raise ValueError("invalid standard-library source path")
                source = (relative / filename).as_posix()
                admit(source, files)
                sources[field].append(source)
            if not sources[field]:
                del sources[field]
        nearby = []
        for parent in [relative, *relative.parents]:
            for filename in TOOLCHAIN_DECLARATIONS:
                candidate = parent / filename
                path = toolchain_root / candidate
                if path.exists() or path.is_symlink():
                    notice = candidate.as_posix()
                    admit(notice, declarations)
                    nearby.append(notice)
        result.append({"package": name, "directory": relative.as_posix(),
                       "source_fields": sources, "nearby_declaration_candidates": sorted(nearby)})
    return {"scope": "Selected standard-library package source/embedded/object files and named ancestor declarations from this installed GOROOT; not complete compiler/native inputs, linked-code membership or notice applicability",
            "rights_status": "review-required", "packages": result,
            "files": [files[name] for name in sorted(files)],
            "declaration_candidates": [declarations[name] for name in sorted(declarations)]}


def package_inventory(packages, roots, notices):
    roots = {name: path.resolve() for name, path in roots.items()}
    result, assets, native, unresolved = [], [], [], []
    source_count = 0
    cgo_fields = ("CgoFiles", "CFiles", "CXXFiles", "MFiles", "HFiles", "FFiles", "SFiles", "SwigFiles", "SwigCXXFiles")
    seen_packages = set()
    for package in sorted(packages, key=lambda value: value["ImportPath"]):
        if package.get("Error") or package.get("DepsErrors") or package.get("Incomplete"):
            raise ValueError("incomplete Go package graph: " + package["ImportPath"])
        if package.get("Standard"):
            continue  # Toolchain declaration is recorded separately.
        if not package.get("Module"):
            raise ValueError("nonstandard package has no module identity: " + package["ImportPath"])
        name, module = package["ImportPath"], package["Module"]["Path"]
        if name in seen_packages or module not in roots:
            raise ValueError("duplicate package or unverified module: " + name)
        seen_packages.add(name)
        directory = Path(package["Dir"]).resolve()
        try:
            relative_directory = directory.relative_to(roots[module])
        except ValueError as error:
            raise ValueError("package outside verified module: " + name) from error
        selected = len(package.get("EmbedFiles", []))
        if package.get("CgoFiles"):
            selected += sum(len(package.get(field, [])) for field in cgo_fields)
        if selected > MAX_FILES - source_count:
            raise ValueError("build inventory exceeds file-count bound")
        source_count += selected
        entries = []
        for filename in sorted(package.get("EmbedFiles", [])):
            record = file_identity(roots[module], (relative_directory / filename).as_posix())
            record.update({"package": name, "module": module, "kind": "font" if Path(filename).suffix.lower() in FONT_SUFFIXES else "embedded-asset", "nearby_notice_candidates": nearby_notices(record["file"], notices.get(module, {})), "rights_status": "review-required"})
            assets.append(record)
            entries.append(record["file"])
        flags = package.get("CgoLDFLAGS", [])
        cgo_sources = []
        for field in cgo_fields:
            if not package.get("CgoFiles"):
                break
            for filename in sorted(package.get(field, [])):
                record = file_identity(roots[module], (relative_directory / filename).as_posix())
                record["source_field"] = field
                cgo_sources.append(record)
        searches, libraries, frameworks, direct_files = [], [], [], []
        position = 0
        while position < len(flags):
            flag = flags[position]
            if flag in {"-L", "-l", "-framework", "-F"}:
                if position + 1 == len(flags):
                    raise ValueError("incomplete native linker flag: " + name)
                value = flags[position + 1]
                position += 2
                if flag == "-L": searches.append(value)
                elif flag == "-l": libraries.append(value)
                elif flag == "-framework": frameworks.append(value)
                else: unresolved.append({"package": name, "flag": flag, "value": normalize_flag(value, roots), "reason": "framework-search-path-not-resolved"})
                continue
            if flag.startswith("-L") and len(flag) > 2: searches.append(flag[2:])
            elif flag.startswith("-l") and len(flag) > 2: libraries.append(flag[2:])
            elif flag.startswith("-F") and len(flag) > 2:
                unresolved.append({"package": name, "flag": normalize_flag(flag, roots), "reason": "framework-search-path-not-resolved"})
            elif not flag.startswith("-") and Path(flag).suffix.lower() in NATIVE_SUFFIXES:
                direct_files.append(flag)
            else:
                unresolved.append({"package": name, "flag": normalize_flag(flag, roots), "reason": "unparsed-native-linker-option"})
            position += 1
        for search in searches:
            path = Path(search)
            path = (path if path.is_absolute() else directory / path).resolve()
            if not any(path.is_relative_to(base) for base in roots.values()):
                unresolved.append({"package": name, "flag": "-L", "value": normalize_flag(str(path), roots), "reason": "search-path-outside-verified-modules"})
        for direct in direct_files:
            path = Path(direct)
            if not path.is_absolute(): path = directory / path
            # Preserve the final component so file_identity refuses source links.
            path = path.parent.resolve() / path.name
            candidates = []
            for owner, base in sorted(roots.items(), key=lambda pair: len(str(pair[1])), reverse=True):
                if path.is_relative_to(base):
                    if path.exists() or path.is_symlink():
                        record = file_identity(base, path.relative_to(base).as_posix())
                        record.update({"module": owner, "nearby_notice_candidates": nearby_notices(record["file"], notices.get(owner, {}))})
                        candidates.append(record)
                    break
            if not candidates:
                unresolved.append({"package": name, "flag": normalize_flag(str(path), roots), "reason": "direct-native-file-missing-or-outside-verified-modules"})
            native.append({"package": name, "direct_input": normalize_flag(str(path), roots), "file_candidates": candidates, "link_selection": "unverified", "embedded_native_component_provenance": "unreviewed", "rights_status": "review-required"})
        for library in libraries:
            if not re.fullmatch(r"[A-Za-z0-9_+.-]+", library):
                raise ValueError("unsupported native library name: " + library)
            candidates = []
            for search in searches:
                path = Path(search)
                if not path.is_absolute():
                    path = directory / path
                path = path.resolve()
                for owner, base in sorted(roots.items()):
                    try:
                        relative = path.relative_to(base)
                    except ValueError:
                        continue
                    for suffix in (".a", ".dylib", ".so"):
                        filename = (relative / ("lib" + library + suffix)).as_posix()
                        candidate = base / filename
                        if candidate.exists() or candidate.is_symlink():
                            record = file_identity(base, filename)
                            record.update({"module": owner, "nearby_notice_candidates": nearby_notices(filename, notices.get(owner, {}))})
                            if record not in candidates: candidates.append(record)
            if not candidates:
                unresolved.append({"package": name, "flag": "-l" + library, "reason": "no-module-file-candidate"})
            native.append({"package": name, "library": library, "searches": [normalize_flag(value if value.startswith("/") else str(directory / value), roots) for value in searches], "file_candidates": candidates, "link_selection": "unverified", "embedded_native_component_provenance": "unreviewed", "rights_status": "review-required"})
        result.append({"package": name, "module": module, "embedded_files": entries, "cgo_source_files": cgo_sources, "cgo_ldflags": [normalize_flag(flag, roots) for flag in flags], "framework_references": frameworks, "framework_resolution": "unverified"})
    return {"packages": result, "embedded_assets": assets, "native_link_candidates": native, "unresolved_native_inputs": unresolved}


def native_link_observation(binary, link_map, roots, builds):
    """Observe supplied artifacts, without claiming they came from one build."""
    binary, link_map = Path(binary).absolute(), Path(link_map).absolute()
    binary_identity = file_identity(binary.parent.resolve(), binary.name)
    with binary.open("rb") as stream:
        header = stream.read(8)
    if len(header) != 8 or header[:4] not in {b"\xcf\xfa\xed\xfe", b"\xfe\xed\xfa\xcf"}:
        raise ValueError("link observation requires a thin 64-bit Mach-O binary")
    cpu = struct.unpack("<I" if header[:4] == b"\xcf\xfa\xed\xfe" else ">I", header[4:])[0]
    architecture = {0x100000c: "arm64", 0x1000007: "x86_64"}.get(cpu)
    mapping = linker_objects(link_map)
    if architecture != mapping["architecture"]:
        raise ValueError("binary and linker-map architecture differ")
    candidates = {}
    for library in builds["native_link_candidates"]:
        for candidate in library["file_candidates"]:
            if candidate["file"].endswith(".a"):
                path = (roots[candidate["module"]] / candidate["file"]).resolve()
                candidates[path] = candidate
    archives = []
    observed = set()
    unresolved = {}
    for name in mapping["objects"]:
        match = re.fullmatch(r"(.+\.a)\(([^()]+)\)", name)
        if match:
            if not Path(match[1]).is_absolute():
                unresolved[match[1]] = "relative-linker-directory-unbound"
                continue
            path = Path(match[1]).resolve()
            if path in candidates:
                observed.add(path)
            else:
                unresolved[normalize_flag(str(path), roots)] = "not-a-verified-module-file-candidate"
    for path in sorted(observed):
        candidate = candidates[path]
        archive = archive_members(path)
        if archive["bytes"] != candidate["bytes"] or archive["sha256"] != candidate["sha256"]:
            raise ValueError("native archive differs from collected candidate")
        selected = selected_archive_members(mapping, path, archive)
        archives.append({"module": candidate["module"], "file": candidate["file"], "bytes": archive["bytes"], "sha256": archive["sha256"], "archive_member_count": len(archive["members"]), "selected_members": selected, "selection": "reported-by-linker-object-table", "component_source_and_rights": "unreviewed"})
    if file_identity(binary.parent.resolve(), binary.name) != binary_identity:
        raise ValueError("binary changed during link observation")
    map_identity = file_identity(link_map.parent.resolve(), link_map.name)
    if map_identity["bytes"] != mapping["bytes"] or map_identity["sha256"] != mapping["sha256"]:
        raise ValueError("linker map changed during observation")
    return {"binary": binary_identity, "link_map": map_identity, "architecture": architecture, "object_count": len(mapping["objects"]), "archives": archives, "unresolved_archives": [{"file": name, "reason": unresolved[name]} for name in sorted(unresolved)], "artifact_pair_binding": "caller-supplied-unverified", "scope": "Exact archive-member identities reported by absolute paths in the supplied linker object table; relative paths have no bound linker working directory and remain unresolved. Member payload bytes need not all survive dead stripping. Binary/map hashes and matching architecture do not prove they came from one build. A separate owned build receipt must bind them. Object names do not prove Rust component versions, source or redistribution rights."}


def collect(root, include_untracked=False, binary=None, link_map=None):
    if (binary is None) != (link_map is None):
        raise ValueError("binary and linker map must be supplied together")
    root = root.resolve()
    env = child_environment()
    source_state, selected_sources = source_snapshot(root, include_untracked)
    configuration = json.loads(check_child_output(["go", "env", "-json", "GOOS", "GOARCH", "GOVERSION", "CGO_ENABLED", "GOROOT"], cwd=root, env=env, text=True, timeout=COMMAND_TIMEOUT_SECONDS))
    if configuration["GOOS"] != "darwin" or configuration["CGO_ENABLED"] != "1":
        raise ValueError("inventory requires the native macOS cgo build configuration")
    graph = release.verified_module_graph(root)
    notices, _ = release.license_inventory(root, modules=graph, copy_notices=False, max_notice_bytes=MAX_FILE_BYTES)
    notice_by_module = {record["module"]: record for record in notices["modules"]}
    # The verified graph's main module lives in a temporary metadata copy.
    # Target package source belongs to the actual driver, not that retired copy.
    roots = {module["Path"]: (root / "drivers/cogent" if module.get("Main") else Path(module.get("Replace", module)["Dir"])).resolve() for module in graph}
    declarations = []
    for module in sorted(graph, key=lambda value: value["Path"]):
        if module.get("Main") or roots[module["Path"]] == root:
            record = file_identity(roots[module["Path"]], "LICENSE")
            declarations.append({"module": module["Path"], **record})
            notice_by_module[module["Path"]] = {"license_files": [record]}
    raw = check_child_output(["go", "list", "-deps", "-json", "./cmd/workbench"], cwd=root / "drivers/cogent", env=env, text=True, timeout=COMMAND_TIMEOUT_SECONDS)
    packages = release.decode_objects(raw)
    toolchain_root = Path(configuration.pop("GOROOT")).resolve()
    standard = standard_library_inventory(packages, toolchain_root)
    builds = package_inventory(packages, roots, notice_by_module)
    link_observation = native_link_observation(binary, link_map, roots, builds) if binary is not None else None
    if link_observation is not None and link_observation["architecture"] != {"arm64": "arm64", "amd64": "x86_64"}.get(configuration["GOARCH"]):
        raise ValueError("link artifacts differ from selected Go build architecture")
    sources = []
    for name, path in selected_sources:
        if len(sources) == MAX_FILES:
            raise ValueError("SDK inventory exceeds file-count bound")
        record = file_identity(root, name)
        suffix = path.suffix.lower()
        record["kind"] = "font-candidate" if suffix in FONT_SUFFIXES else "native-file-candidate" if suffix in NATIVE_SUFFIXES else "sdk-file"
        sources.append(record)
    # Capture patch/pin identity in addition to hashes of the prepared files.
    pin_record = file_identity(root, "drivers/cogent/native/dependencies.json")
    toolchain_roots = {"GOROOT": toolchain_root, "GOROOT-parent": toolchain_root.parent}
    toolchain_declarations = []
    for label, base in toolchain_roots.items():
        for filename in TOOLCHAIN_DECLARATIONS:
            if (base / filename).exists() or (base / filename).is_symlink():
                toolchain_declarations.append({"source_root": label, **file_identity(base, filename)})
    inputs = [(root, record) for record in sources]
    inputs += [(toolchain_root, record) for record in standard["files"] + standard["declaration_candidates"]]
    inputs += [(roots[record["module"]], record) for record in builds["embedded_assets"]]
    inputs += [(roots[package["module"]], record) for package in builds["packages"] for record in package["cgo_source_files"]]
    inputs += [(roots[record["module"]], record) for library in builds["native_link_candidates"] for record in library["file_candidates"]]
    inputs += [(roots[record["module"]], record) for record in declarations]
    inputs += [(toolchain_roots[record["source_root"]], record) for record in toolchain_declarations]
    for module in notices["modules"]:
        for notice in module["license_files"]:
            record = file_identity(roots[module["module"]], notice["file"])
            if record["sha256"] != notice["sha256"]:
                raise ValueError("notice changed during collection: " + notice["file"])
            inputs.append((roots[module["module"]], record))
    for base, record in inputs:
        current = file_identity(base, record["file"])
        if current["bytes"] != record["bytes"] or current["sha256"] != record["sha256"]:
            raise ValueError("inventory input changed during collection: " + record["file"])
    latest = check_child_output(["go", "list", "-deps", "-json", "./cmd/workbench"], cwd=root / "drivers/cogent", env=env, text=True, timeout=COMMAND_TIMEOUT_SECONDS)
    latest_packages = release.decode_objects(latest)
    if build_metadata(latest_packages) != build_metadata(packages):
        raise ValueError("Go build metadata changed during inventory collection")
    if standard_library_inventory(latest_packages, toolchain_root) != standard:
        raise ValueError("standard-library inputs or declarations changed during collection")
    latest_state, latest_sources = source_snapshot(root, include_untracked)
    if [name for name, _ in latest_sources] != [record["file"] for record in sources]:
        raise ValueError("SDK source set changed during inventory collection")
    if latest_state != source_state:
        raise ValueError("Git or SDK source state changed during inventory collection")
    if link_observation is not None:
        if native_link_observation(binary, link_map, roots, builds) != link_observation:
            raise ValueError("native link observation changed during collection")
    result = {"schema": 1, "scope": "SDK source payload and selected macOS workbench Go build inputs; native files are linker search candidates, not proof of final binary inclusion or a complete embedded Rust/C dependency closure", "distribution_audit": "pending", "runtime_executed": False, "source": source_state, "include_untracked": include_untracked, "build_configuration": configuration, "package_pattern": "./cmd/workbench", "pin_manifest": pin_record, "framework_declarations": declarations, "toolchain_declaration_candidates": toolchain_declarations, "go_standard_library": standard, "go_notices": notices, "sdk_files": sources, **builds}
    if link_observation is not None:
        result["native_link_observation"] = link_observation
    return result


def write_inventory(root, output, include_untracked=False, binary=None, link_map=None):
    output = output.absolute()
    if output.exists() or output.is_symlink():
        raise ValueError("inventory output already exists")
    try:
        output.resolve().relative_to(root.resolve())
    except ValueError:
        pass
    else:
        raise ValueError("inventory output must be outside the source checkout")
    inventory = collect(root, include_untracked, binary, link_map)
    data = (json.dumps(inventory, indent=2, sort_keys=True) + "\n").encode()
    output.parent.mkdir(parents=True, exist_ok=True)
    pending = None
    try:
        with tempfile.NamedTemporaryFile(prefix=".vellum-inventory-", dir=output.parent, delete=False) as stream:
            pending = Path(stream.name)
            stream.write(data); stream.flush(); os.fsync(stream.fileno())
        pending.chmod(0o644)
        publish_exclusive(pending, output)
    finally:
        if pending is not None: pending.unlink(missing_ok=True)
    return {"event": "vellum.distribution.inventory", "output": str(output), "sha256": hashlib.sha256(data).hexdigest(), "distribution_audit": "pending", "runtime_executed": False, "sdk_files": len(inventory["sdk_files"]), "embedded_assets": len(inventory["embedded_assets"]), "native_link_candidates": len(inventory["native_link_candidates"])}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--include-untracked", action="store_true", help="record nonignored whitelisted private precommit inputs")
    parser.add_argument("--binary", type=Path, help="optional private Mach-O artifact; requires --link-map")
    parser.add_argument("--link-map", type=Path, help="optional ld -map artifact; pair provenance remains caller-supplied")
    args = parser.parse_args()
    print(json.dumps(write_inventory(ROOT, args.output, args.include_untracked, args.binary, args.link_map), indent=2))


if __name__ == "__main__":
    main()
