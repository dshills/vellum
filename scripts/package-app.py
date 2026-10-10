#!/usr/bin/env python3
"""Package a prebuilt private app without research prototypes, publishing or notarization."""
import argparse
import hashlib
import gzip
import json
import os
from pathlib import Path
import platform
import plistlib
import posixpath
import re
import shutil
import tarfile
import tempfile
import zipfile

from artifact_io import publish_exclusive
from child_process import run_child, check_child_output
from sdk_environment import child_environment

NATIVE_TOOL_TIMEOUT_SECONDS = 60


def valid_name(value):
    if not isinstance(value, str) or value in {"", ".", ".."} or value.upper() == "APP_MANIFEST.JSON" or any(c in value for c in '/\\:*?"<>|') or any(ord(c) < 32 or ord(c) == 127 for c in value) or value.endswith((" ", ".")) or value.startswith("-"):
        return False
    stem = value.split(".", 1)[0].upper()
    return stem not in {"CON", "PRN", "AUX", "NUL", *("COM" + str(i) for i in range(1, 10)), *("LPT" + str(i) for i in range(1, 10))}


def validate(binary, name, identifier, version, target, identity, build_version):
    if binary.is_symlink() or not binary.is_file():
        raise ValueError("app binary must be a regular file")
    if not valid_name(name) or not re.fullmatch(r"[A-Za-z0-9]+(?:[.-][A-Za-z0-9]+)+",identifier):
        raise ValueError("invalid app name or bundle identifier")
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+",version):
        raise ValueError("app version must be numeric major.minor.patch")
    if not re.fullmatch(r"[0-9]+(?:\.[0-9]+){0,2}", build_version):
        raise ValueError("build version must contain one to three numeric components")
    with binary.open("rb") as source:
        magic = source.read(4)
    expected = {"darwin":{b"\xcf\xfa\xed\xfe",b"\xfe\xed\xfa\xcf",b"\xca\xfe\xba\xbe"},"linux":{b"\x7fELF"}}
    if target == "windows":
        with binary.open("rb") as source:
            header = source.read(64)
            offset = int.from_bytes(header[60:64],"little") if len(header)==64 else 0
            source.seek(offset)
            signature = source.read(4)
        if not magic.startswith(b"MZ") or offset < 64 or signature != b"PE\x00\x00":
            raise ValueError("Windows app requires a PE executable")
    elif magic not in expected.get(target,set()):
        raise ValueError("binary format does not match app target")
    if identity is not None and target != "darwin":
        raise ValueError("signing is implemented only for explicitly selected macOS identities")
    if target == "darwin" and platform.system() != "Darwin":
        raise ValueError("macOS dependency inspection/signing needs a macOS host")


def darwin_dependencies(binary):
    output = check_child_output(["otool","-L",str(binary)],text=True, env=child_environment(), timeout=NATIVE_TOOL_TIMEOUT_SECONDS)
    dependencies = set()
    for line in output.splitlines():
        if not line.startswith(("\t", " ")):
            continue  # Each architecture of a universal binary has its own header.
        value = line.strip().split(" (",1)[0]
        if value and not posixpath.normpath(value).startswith(("/System/Library/","/usr/lib/")):
            dependencies.add(value)
    if dependencies:
        raise ValueError("non-system native dependencies require an audited bundle closure: " + ", ".join(sorted(dependencies)))


def normalize_layout(stage):
    stage.chmod(0o755)
    for path in stage.rglob("*"):
        path.chmod(0o755 if path.is_dir() or path.stat().st_mode & 0o111 else 0o644)


def normalized_tar_member(member, executable_name):
    member.uid = member.gid = member.mtime = 0
    member.uname = member.gname = ""
    member.mode = 0o755 if member.isdir() or member.name == executable_name else 0o644
    return member


def application_archive(stage, output, name, target):
    if target == "windows":
        with zipfile.ZipFile(output, "w", compression=zipfile.ZIP_DEFLATED) as archive:
            for source in [stage, *sorted(stage.rglob("*"))]:
                relative = source.relative_to(stage).as_posix()
                filename = name + ("/" + relative if relative != "." else "")
                if source.is_dir():
                    filename += "/"
                record = zipfile.ZipInfo(filename, (1980, 1, 1, 0, 0, 0))
                record.create_system = 3
                record.compress_type = zipfile.ZIP_DEFLATED
                mode = 0o040755 if source.is_dir() else (0o100755 if source.name == name + ".exe" else 0o100644)
                record.external_attr = mode << 16
                if source.is_dir():
                    record.external_attr |= 0x10
                archive.writestr(record, b"" if source.is_dir() else source.read_bytes())
    else:
        with output.open("wb") as destination:
            with gzip.GzipFile(fileobj=destination, mode="wb", filename="", mtime=0) as zipped:
                with tarfile.open(fileobj=zipped, mode="w|", format=tarfile.PAX_FORMAT) as archive:
                    archive.add(stage, arcname=name, filter=lambda member: normalized_tar_member(member, name + "/" + name))


def package(binary, output, name, identifier, version, target, identity=None, build_version="1"):
    validate(binary,name,identifier,version,target,identity,build_version)
    sdk_release = json.loads((Path(__file__).resolve().parents[1]/"release/version.json").read_text())["version"]
    if output.exists() or output.is_symlink():
        raise ValueError("app output already exists")
    if target == "darwin":
        if output.suffix.lower() != ".app":
            raise ValueError("macOS app output must have a .app suffix")
    output.parent.mkdir(parents=True,exist_ok=True)
    with tempfile.TemporaryDirectory(prefix=".vellum-app-",dir=output.parent) as directory:
        stage = Path(directory)/("application.app" if target=="darwin" else name)
        stage.mkdir()
        payload = stage/"Contents/MacOS" if target=="darwin" else stage
        payload.mkdir(parents=True,exist_ok=True)
        executable = payload/(name+".exe" if target=="windows" else name)
        shutil.copyfile(binary,executable);executable.chmod(0o755)
        validate(executable,name,identifier,version,target,identity,build_version)
        if target == "darwin":
            darwin_dependencies(executable)
        manifest = {"schema":1,"name":name,"identifier":identifier,"version":version,"build_version":build_version,"sdk_release":sdk_release,"target":target,"source_binary_sha256":hashlib.sha256(executable.read_bytes()).hexdigest(),"private_build":True,"signed_identity":identity,"notarized":False,"native_distribution_audit":"pending","target_runtime":"unverified"}
        if target=="darwin":
            (stage/"Contents/Resources").mkdir()
            info={"CFBundlePackageType":"APPL","CFBundleName":name,"CFBundleDisplayName":name,"CFBundleExecutable":name,"CFBundleIdentifier":identifier,"CFBundleVersion":build_version,"CFBundleShortVersionString":version,"NSHighResolutionCapable":True}
            (stage/"Contents/Info.plist").write_bytes(plistlib.dumps(info,sort_keys=True))
            (stage/"Contents/Resources/APP_MANIFEST.json").write_text(json.dumps(manifest,indent=2)+"\n")
            normalize_layout(stage)
            if identity is not None:
                run_child(["codesign","--force","--sign",identity,str(stage)],check=True, env=child_environment(), timeout=NATIVE_TOOL_TIMEOUT_SECONDS)
                run_child(["codesign","--verify","--strict",str(stage)],check=True, env=child_environment(), timeout=NATIVE_TOOL_TIMEOUT_SECONDS)
            # Exclusive reservation keeps preexisting directories/symlinks safe.
            output.mkdir()
            try:
                output.chmod(0o755)
                for child in stage.iterdir():
                    shutil.move(str(child),output/child.name)
                if identity is not None:
                    run_child(["codesign","--verify","--strict",str(output)],check=True, env=child_environment(), timeout=NATIVE_TOOL_TIMEOUT_SECONDS)
            except BaseException:
                shutil.rmtree(output);raise
        else:
            (stage/"APP_MANIFEST.json").write_text(json.dumps(manifest,indent=2)+"\n")
            artifact = Path(directory)/"application.archive"
            normalize_layout(stage)
            application_archive(stage, artifact, name, target)
            artifact.chmod(0o644)
            publish_exclusive(artifact,output)
    return {"output":str(output),"private_build":True,"target":target,"signed":identity is not None,"notarized":False,"native_distribution_audit":"pending"}


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary",required=True,type=Path)
    parser.add_argument("--output",required=True,type=Path)
    parser.add_argument("--name",required=True)
    parser.add_argument("--identifier",required=True)
    parser.add_argument("--version",required=True,help="numeric app release version: major.minor.patch")
    parser.add_argument("--build-version",default="1",help="numeric bundle build version (one to three components); increment for distribution")
    parser.add_argument("--platform",dest="target",choices=["darwin","windows","linux"],default=platform.system().lower())
    parser.add_argument("--sign-identity",help="explicit macOS code-signing identity; '-' requests ad-hoc signing")
    args=parser.parse_args()
    print(json.dumps(package(args.binary.absolute(),args.output.absolute(),args.name,args.identifier,args.version,args.target,args.sign_identity,args.build_version),indent=2))


if __name__=="__main__":main()
