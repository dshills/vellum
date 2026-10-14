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

from native_archive import archive_members, linker_objects, selected_archive_members

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
    fields = ("ImportPath", "Dir", "Imports", "GoFiles", "CgoFiles", "CFiles", "CXXFiles", "MFiles", "HFiles", "FFiles", "SFiles", "SwigFiles", "SwigCXXFiles", "EmbedPatterns", "EmbedFiles", "CgoCFLAGS", "CgoCPPFLAGS", "CgoCXXFLAGS", "CgoFFLAGS", "CgoLDFLAGS", "Error", "DepsErrors", "Incomplete")
    out = []
    for package in sorted(packages, key=lambda value: value["ImportPath"]):
        if package.get("Standard"):
            continue
        record = {key: package[key] for key in fields if key in package}
        module = package.get("Module", {})
        record["Module"] = {key: module[key] for key in ("Path", "Version", "Sum", "Dir", "Replace") if key in module}
        out.append(record)
    return out


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
    configuration = json.loads(check_child_output(["go", "env", "-json", "GOOS", "GOARCH", "GOVERSION", "CGO_ENABLED", "GOROOT"], cwd=root, env=env, text=True, timeout=COMMAND_TIMEOUT_SECONDS))
    if configuration["GOOS"] != "darwin" or configuration["CGO_ENABLED"] != "1":
        raise ValueError("inventory requires the native macOS cgo build configuration")
    source_state = release.source_state(root)
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
    builds = package_inventory(packages, roots, notice_by_module)
    link_observation = native_link_observation(binary, link_map, roots, builds) if binary is not None else None
    if link_observation is not None and link_observation["architecture"] != {"arm64": "arm64", "amd64": "x86_64"}.get(configuration["GOARCH"]):
        raise ValueError("link artifacts differ from selected Go build architecture")
    sources = []
    for name, path in release.source_files(root, include_untracked=include_untracked):
        if len(sources) == MAX_FILES:
            raise ValueError("SDK inventory exceeds file-count bound")
        record = file_identity(root, name)
        suffix = path.suffix.lower()
        record["kind"] = "font-candidate" if suffix in FONT_SUFFIXES else "native-file-candidate" if suffix in NATIVE_SUFFIXES else "sdk-file"
        sources.append(record)
    # Capture patch/pin identity in addition to hashes of the prepared files.
    pin_record = file_identity(root, "drivers/cogent/native/dependencies.json")
    toolchain_root = Path(configuration.pop("GOROOT")).resolve()
    toolchain_roots = {"GOROOT": toolchain_root, "GOROOT-parent": toolchain_root.parent}
    toolchain_declarations = []
    for label, base in toolchain_roots.items():
        for filename in ("LICENSE", "COPYING", "NOTICE"):
            if (base / filename).exists() or (base / filename).is_symlink():
                toolchain_declarations.append({"source_root": label, **file_identity(base, filename)})
    inputs = [(root, record) for record in sources]
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
    if [name for name, _ in release.source_files(root, include_untracked=include_untracked)] != [record["file"] for record in sources]:
        raise ValueError("SDK source set changed during inventory collection")
    latest = check_child_output(["go", "list", "-deps", "-json", "./cmd/workbench"], cwd=root / "drivers/cogent", env=env, text=True, timeout=COMMAND_TIMEOUT_SECONDS)
    if build_metadata(release.decode_objects(latest)) != build_metadata(packages):
        raise ValueError("Go build metadata changed during inventory collection")
    if release.source_state(root) != source_state:
        raise ValueError("Git state changed during inventory collection")
    if link_observation is not None:
        if native_link_observation(binary, link_map, roots, builds) != link_observation:
            raise ValueError("native link observation changed during collection")
    result = {"schema": 1, "scope": "SDK source payload and selected macOS workbench Go build inputs; native files are linker search candidates, not proof of final binary inclusion or a complete embedded Rust/C dependency closure", "distribution_audit": "pending", "runtime_executed": False, "source": source_state, "include_untracked": include_untracked, "build_configuration": configuration, "package_pattern": "./cmd/workbench", "pin_manifest": pin_record, "framework_declarations": declarations, "toolchain_declaration_candidates": toolchain_declarations, "go_notices": notices, "sdk_files": sources, **builds}
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
