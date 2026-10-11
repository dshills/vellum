#!/usr/bin/env python3
"""Build an exclusive, deterministic SDK archive from recorded source and owned dependencies."""
import argparse
from contextlib import contextmanager, nullcontext
import gzip
import hashlib
import io
import json
import os
import re
import shutil
from pathlib import Path
import subprocess
import tarfile
import tempfile
import sys

from child_process import run_child, check_child_output

from artifact_io import publish_exclusive
from sdk_environment import child_environment

ROOT = Path(__file__).resolve().parents[1]
COMMAND_TIMEOUT_SECONDS = 600


def decode_objects(raw):
    decoder = json.JSONDecoder()
    rest = raw.strip()
    objects = []
    while rest:
        value, end = decoder.raw_decode(rest)
        objects.append(value)
        rest = rest[end:].lstrip()
    return objects


def git_output(root, arguments, text=False):
    # The same allowlist drops GIT_DIR/WORK_TREE/INDEX_FILE and GIT_CONFIG_*
    # redirection, so source provenance always describes cwd's repository.
    return check_child_output(["git", *arguments], cwd=root,
                                   env=child_environment(), text=text,
                                   timeout=COMMAND_TIMEOUT_SECONDS)


def source_state(root):
    revision = git_output(root, ["rev-parse", "HEAD"], text=True).strip()
    status = git_output(root, ["status", "--porcelain", "-z", "--untracked-files=normal"])
    return {"commit": revision, "clean": not bool(status)}


def hidden_index_paths(root):
    # Git status can hide worktree changes under these index flags. Never
    # certify committed public bytes while either flag is present.
    entries = git_output(root, ["ls-files", "-v", "-z"]).decode().split("\0")
    return [entry[2:] for entry in entries if entry and (entry[0].islower() or entry[0] == "S")]


def release_gates(root, metadata, public):
    if not public:
        return
    failures = []
    if metadata.get("namespace_status") != "selected" or metadata.get("module", "").startswith("example.com/"):
        failures.append("public module namespace has not been selected")
    if metadata.get("license_status") != "selected" or not (root / "LICENSE").is_file():
        failures.append("owner must select the framework license")
    if metadata.get("third_party_distribution_audit") != "approved":
        failures.append("third-party/native binary distribution audit is incomplete")
    module_file = root / "go.mod"
    declared = re.search(r"^module\s+(\S+)",module_file.read_text(),re.MULTILINE) if module_file.is_file() else None
    if declared is None or declared.group(1).strip('"') != metadata.get("module"):
        failures.append("release module metadata differs from go.mod")
    if not source_state(root)["clean"]:
        failures.append("public SDK source must be committed and clean")
    if not failures and hidden_index_paths(root):
        failures.append("public SDK refuses assume-unchanged or skip-worktree index flags")
    if failures:
        raise ValueError("Public release blocked: " + "; ".join(failures))


def require_regular_source(root, source):
    relative = source.relative_to(root)
    current = root
    for part in relative.parts:
        current = current / part
        if current.is_symlink():
            raise ValueError(f"SDK source cannot contain symlinks: {relative}")
    if not source.is_file():
        raise ValueError(f"SDK source must be a regular file: {relative}")


def source_files(root, include_untracked=False, dependencies=None):
    arguments = ["git", "ls-files", "-z", "--cached"]
    if include_untracked:
        arguments += ["--others", "--exclude-standard"]
    tracked = git_output(root, arguments[1:]).decode().split("\0")
    allowed = {"app", "ui", "canvas", "command", "history", "document", "editor", "execution", "replay", "tasks", "virtual", "workbench", "drivers", "examples", "scripts", "release"}
    for name in sorted(set(filter(None, tracked))):
        path = Path(name)
        if name.startswith("drivers/cogent/build/") or "__pycache__" in path.parts:
            continue
        if path.parts[0] not in allowed and name not in {"go.mod", "go.sum", "README.md", "LICENSE"} and not (len(path.parts) == 1 and path.suffix == ".go"):
            if not (name.startswith("specs/") or name.startswith("docs/") and not name.startswith(("docs/reviews/", "docs/reports/", "docs/research/", "docs/assets/"))):
                continue
        source = root / path
        require_regular_source(root, source)
        yield name, source
    for dependency in ("upstream", "typesetting"):
        base = dependencies[dependency] if dependencies is not None else root / "drivers/cogent/build" / dependency
        if not (base / "go.mod").is_file():
            raise ValueError("run drivers/cogent/prepare.sh before creating the SDK")
        require_regular_source(root if dependencies is None else base.parent, base / "go.mod")
        for source in sorted(base.rglob("*")):
            if source.is_symlink():
                raise ValueError(f"SDK dependencies cannot contain symlinks: {source}")
            if source.is_file():
                require_regular_source(base, source)
                yield "drivers/cogent/build/" + dependency + "/" + source.relative_to(base).as_posix(), source


@contextmanager
def public_dependencies(root):
    """Regenerate pinned, checksum-verified sources without trusting ignored trees."""
    driver = root / "drivers/cogent"
    metadata = json.loads((driver / "native/dependencies.json").read_text())
    dependencies = metadata["dependencies"]
    if {entry["name"] for entry in dependencies} != {"upstream", "typesetting"} or len(dependencies) != 2:
        raise ValueError("public SDK requires exactly the pinned native dependencies")
    for entry in dependencies:
        if not re.fullmatch(r"[A-Za-z0-9_./-]+", entry["module"]) or not re.fullmatch(r"v[0-9A-Za-z.+-]+", entry["version"]):
            raise ValueError("invalid pinned dependency module/version")
    env = child_environment()
    with tempfile.TemporaryDirectory(prefix="vellum-sdk-pinned-") as temporary:
        directory = Path(temporary)
        verification = directory / "verification"
        verification.mkdir()
        go_version = re.search(r"^go\s+(\S+)", (root / "go.mod").read_text(), re.MULTILINE)
        if go_version is None:
            raise ValueError("framework go.mod requires a Go version")
        requirements = "\n".join("  " + entry["module"] + " " + entry["version"] for entry in dependencies)
        (verification / "go.mod").write_text("module example.com/vellum-sdk-source-verification\ngo " + go_version.group(1) + "\nrequire (\n" + requirements + "\n)\n")
        for entry in dependencies:
            downloaded = json.loads(check_child_output(["go", "mod", "download", "-json", entry["module"] + "@" + entry["version"]], cwd=verification, env=env, text=True, timeout=COMMAND_TIMEOUT_SECONDS))
            if downloaded.get("Error") or downloaded.get("Sum") != entry["sum"]:
                raise ValueError("pinned dependency checksum mismatch: " + entry["module"])
        # Download Sum alone can trust a cache record. Verify hashes of the actual
        # cached zip AND extracted module directories before copying/patching.
        run_child(["go", "mod", "verify"], cwd=verification, env=env, check=True, stdout=subprocess.PIPE, timeout=COMMAND_TIMEOUT_SECONDS)
        # Execute the owned bootstrap in a child with the same controlled Go
        # environment; its nested Go/patch processes inherit that environment.
        helper = "import importlib.util,json,sys;from pathlib import Path;s=importlib.util.spec_from_file_location('prepare',sys.argv[1]);m=importlib.util.module_from_spec(s);s.loader.exec_module(m);m.prepare_dependency(json.loads(sys.argv[2]),Path(sys.argv[3]))"
        paths = {}
        for entry in dependencies:
            destination = directory / entry["name"]
            run_child([sys.executable, "-c", helper, str(driver / "prepare.py"), json.dumps(entry), str(destination)], cwd=verification, env=env, check=True, timeout=COMMAND_TIMEOUT_SECONDS)
            paths[entry["name"]] = destination
        # A persistent cache modification during copying also blocks publication.
        run_child(["go", "mod", "verify"], cwd=verification, env=env, check=True, stdout=subprocess.PIPE, timeout=COMMAND_TIMEOUT_SECONDS)
        yield paths


def declared_notice(path):
    """Conventional text notices only; source/test fixtures are not declarations."""
    name = path.name.lower()
    match = re.fullmatch(r"(?:licen[cs]e|copying|notice)(?:[._-][a-z0-9._-]+)?", name)
    if not match or path.suffix.lower() in {".go", ".c", ".h", ".cc", ".cpp", ".rs", ".py", ".js", ".ts", ".json", ".yaml", ".yml", ".xml", ".html", ".svg", ".png", ".pdf"}:
        return False
    return not any(part.lower() in {"test", "tests", "testdata", "fixture", "fixtures", "example", "examples"} for part in path.parts[:-1])


def module_license(path):
    """A module's own root/license-directory declaration, not vendor notices."""
    return declared_notice(path) and re.match(r"^(?:licen[cs]e|copying)(?:[._-]|$)", path.name.lower()) is not None and (len(path.parts) == 1 or len(path.parts) == 2 and path.parts[0].lower() in {"license", "licenses", "licence", "licences"})


def license_inventory(root, dependencies=None):
    env = child_environment()
    # Listing the module graph alone leaves Dir absent for cold-cache modules.
    # Materialize and verify every external source before collecting declarations.
    # Work on copied graph files: Go downloads may add sums even in readonly mode.
    with tempfile.TemporaryDirectory(prefix="vellum-sdk-license-graph-") as temporary:
        graph = Path(temporary)
        shutil.copyfile(root / "drivers/cogent/go.mod", graph / "go.mod")
        shutil.copyfile(root / "drivers/cogent/go.sum", graph / "go.sum")
        module_path = json.loads((root / "release/version.json").read_text())["module"]
        owned = dependencies if dependencies is not None else {name: root / "drivers/cogent/build" / name for name in ("upstream", "typesetting")}
        replacements = {module_path: root.resolve(), "cogentcore.org/core": owned["upstream"].resolve(), "github.com/go-text/typesetting": owned["typesetting"].resolve()}
        run_child(["go", "mod", "edit", *["-replace=" + name + "=" + str(path) for name, path in replacements.items()]], cwd=graph, env=env, check=True, timeout=COMMAND_TIMEOUT_SECONDS)
        run_child(["go", "mod", "download", "all"], cwd=graph, env=env, check=True, timeout=COMMAND_TIMEOUT_SECONDS)
        run_child(["go", "mod", "verify"], cwd=graph, env=env, check=True, stdout=subprocess.PIPE, timeout=COMMAND_TIMEOUT_SECONDS)
        raw = check_child_output(["go", "list", "-m", "-json", "all"], cwd=graph, text=True, env=env, timeout=COMMAND_TIMEOUT_SECONDS)
    modules = decode_objects(raw)
    pinned_file = root / "drivers/cogent/native/dependencies.json"
    pinned = {item["module"]: item for item in json.loads(pinned_file.read_text())["dependencies"]} if pinned_file.is_file() else {}
    inventory = []
    notices = {}
    for module in sorted(modules, key=lambda item: item["Path"]):
        if module.get("Main") or module["Path"] == json.loads((root / "release/version.json").read_text())["module"]:
            continue
        effective = module.get("Replace", module)
        directory = Path(effective["Dir"]) if "Dir" in effective else None
        copies = []
        own_licenses = []
        if directory is None or not directory.is_dir():
            raise ValueError("dependency source unavailable after verified download: " + module["Path"])
        for path in sorted(directory.rglob("*")):
            if path.is_file() and not path.is_symlink() and declared_notice(path.relative_to(directory)):
                relative = path.relative_to(directory).as_posix()
                key = hashlib.sha256(module["Path"].encode()).hexdigest()[:16] + "/" + relative
                content = path.read_bytes()
                notices["THIRD_PARTY_NOTICES/" + key] = content
                copies.append({"file": relative, "notice": key, "sha256": hashlib.sha256(content).hexdigest()})
                if module_license(Path(relative)):
                    own_licenses.append(relative)
        source_pin = pinned.get(module["Path"]) if module.get("Replace") and effective.get("Dir") and not effective.get("Version") else None
        version = source_pin["version"] if source_pin else effective.get("Version", module.get("Version"))
        source_sum = source_pin["sum"] if source_pin else effective.get("Sum", module.get("Sum"))
        inventory.append({"module": module["Path"], "version": version, "sum": source_sum, "license_files": copies, "module_license_files": own_licenses, "status": "declared-files-copied" if own_licenses else "manual-review-required"})
    return {"schema": 1, "scope": "Go module graph and declared license/notice files; not a legal approval or complete embedded native/Rust/font audit", "distribution_audit": "pending", "modules": inventory}, notices


def add_member(archive, name, data, executable=False):
    member = tarfile.TarInfo(name)
    member.size = len(data)
    member.mode = 0o755 if executable else 0o644
    member.uid = member.gid = member.mtime = 0
    archive.addfile(member, io.BytesIO(data))


def package(root, output, public=False, include_untracked=False):
    if public and include_untracked:
        raise ValueError("public SDK cannot include untracked files")
    # Outputs and their pending archive must never be discoverable as source.
    # Resolve ancestors for this containment check but preserve the caller's
    # final path so the existing/dangling-symlink refusal remains explicit.
    if output.exists() or output.is_symlink():
        raise ValueError("release output already exists")
    try:
        output.resolve().relative_to(root.resolve())
    except ValueError:
        pass
    else:
        raise ValueError("release output must be outside the source checkout")
    metadata = json.loads((root / "release/version.json").read_text())
    version = metadata["version"]
    # The archive prefix is a bounded data value, never a caller-selected path.
    if not isinstance(version,str) or not re.fullmatch(r"v[0-9]+\.[0-9]+\.[0-9]+(?:-[A-Za-z0-9.-]+)?",version):
        raise ValueError("invalid release version")
    release_gates(root, metadata, public)
    state = source_state(root)
    prefix = "vellum-" + version
    output.parent.mkdir(parents=True, exist_ok=True)
    checksums = {}
    untracked_candidates = set(filter(None, git_output(root, ["ls-files", "-z", "--others", "--exclude-standard"]).decode().split("\0"))) if include_untracked else set()
    included_untracked = []
    pending = None
    try:
        with public_dependencies(root) if public else nullcontext(None) as dependencies:
            inventory, notices = license_inventory(root, dependencies)
            if public and any(not item.get("module_license_files") for item in inventory["modules"]):
                raise ValueError("a dependency has no recorded module license declaration")
            with tempfile.NamedTemporaryFile(prefix=".vellum-sdk-", dir=output.parent, delete=False) as temporary:
                pending = Path(temporary.name)
                with gzip.GzipFile(fileobj=temporary, mode="wb", filename="", mtime=0) as zipped:
                    with tarfile.open(fileobj=zipped, mode="w|", format=tarfile.PAX_FORMAT) as archive:
                        for name, source in source_files(root, include_untracked=include_untracked, dependencies=dependencies):
                            data = source.read_bytes()
                            if name in untracked_candidates:
                                included_untracked.append(name)
                            add_member(archive, prefix + "/" + name, data, bool(source.stat().st_mode & 0o111))
                            checksums[name] = hashlib.sha256(data).hexdigest()
                        generated = dict(notices)
                        generated["THIRD_PARTY_INVENTORY.json"] = (json.dumps(inventory, indent=2, sort_keys=True) + "\n").encode()
                        if not public:
                            generated["PRIVATE_BUILD.txt"] = b"Private SDK validation build. Public release, licensing and platform qualification gates have not passed.\n"
                        for name, data in sorted(generated.items()):
                            if name in checksums:
                                raise ValueError(f"generated SDK record collides with source: {name}")
                            add_member(archive, prefix + "/" + name, data)
                            checksums[name] = hashlib.sha256(data).hexdigest()
                        manifest = {"schema": 1, "release": metadata, "source": state, "private_build": not public, "untracked_files": sorted(included_untracked), "files": checksums}
                        add_member(archive, prefix + "/SDK_MANIFEST.json", (json.dumps(manifest, indent=2, sort_keys=True) + "\n").encode())
                temporary.flush()
                os.fsync(temporary.fileno())
        # Public source must still be clean after dependency inventory/read operations.
        if public and source_state(root) != state:
            raise ValueError("public SDK source changed during packaging")
        # Exclusive same-filesystem publication never replaces an existing file.
        pending.chmod(0o644)
        publish_exclusive(pending, output)
        return {"archive": str(output), "sha256": hashlib.sha256(output.read_bytes()).hexdigest(), "version": version, "modules": len(inventory["modules"]), "private_build": not public}
    finally:
        if pending is not None:
            pending.unlink(missing_ok=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--public", action="store_true", help="require the recorded public namespace and license/audit gates")
    parser.add_argument("--include-untracked", action="store_true", help="explicitly include nonignored whitelisted untracked files in a private precommit snapshot; their names are recorded")
    args = parser.parse_args()
    print(json.dumps(package(ROOT, args.output.absolute(), args.public, args.include_untracked), indent=2))


if __name__ == "__main__":
    main()
