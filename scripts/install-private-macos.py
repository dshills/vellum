#!/usr/bin/env python3
"""Install an identified private bundle into a new durable versioned directory."""
import argparse
import fcntl
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import sys
import tempfile
import zipfile

sys.dont_write_bytecode = True
from child_process import run_child

MAX_ARCHIVE = 512 << 20
MAX_EXPANDED = 1 << 30
APP_NAMES = {'Vellum Forge', 'Vellum Measure Pad', 'Vellum Documents',
             'Vellum Document Starter', 'Vellum Notes'}


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        while data := stream.read(1 << 20):
            h.update(data)
    return h.hexdigest()


def regular(path, limit):
    info = path.lstat()
    if not stat.S_ISREG(info.st_mode) or info.st_size > limit:
        raise ValueError('expected bounded regular input: ' + str(path))


def signature(app):
    run_child(['codesign', '--verify', '--deep', '--strict', app], check=True, timeout=30)


def entries(archive, app_name):
    """Admit only this app's bounded regular files/directories, never links."""
    result = {}; total = 0
    with zipfile.ZipFile(archive) as source:
        infos = source.infolist()
        if not infos or len(infos) > 2000:
            raise ValueError('app ZIP member count exceeded')
        for info in infos:
            name = info.filename.rstrip('/')
            parts = PurePosixPath(name).parts
            mode = info.external_attr >> 16
            if ('\\' in name or not parts or parts[0] != app_name + '.app'
                    or any(p in ('', '.', '..') for p in name.split('/'))
                    or name in result or stat.S_IFMT(mode) not in (0, stat.S_IFREG, stat.S_IFDIR)
                    or info.flag_bits & 1):
                raise ValueError('unsafe app ZIP member')
            total += info.file_size
            if total > MAX_EXPANDED or info.file_size > MAX_ARCHIVE:
                raise ValueError('app ZIP expanded size exceeded')
            if info.is_dir():
                result[name] = None
            else:
                h = hashlib.sha256()
                with source.open(info) as stream:
                    while data := stream.read(1 << 20):
                        h.update(data)
                result[name] = h.hexdigest()
    return result


def verify_tree(directory, expected):
    actual = {}
    for path in directory.rglob('*'):
        mode = path.lstat().st_mode
        if stat.S_ISLNK(mode) or not (stat.S_ISREG(mode) or stat.S_ISDIR(mode)):
            raise ValueError('unexpected installed link or special file')
        if path.is_file():
            actual[path.relative_to(directory).as_posix()] = digest(path)
    if actual != {k: v for k, v in expected.items() if v is not None}:
        raise ValueError('installed app bytes differ from verified ZIP')


def extract(archive, destination):
    with zipfile.ZipFile(archive) as source:
        for info in source.infolist():
            path = destination / info.filename
            if info.is_dir():
                path.mkdir(parents=True, exist_ok=True)
            else:
                path.parent.mkdir(parents=True, exist_ok=True)
                with source.open(info) as data, path.open('xb') as out:
                    while chunk := data.read(1 << 20):
                        out.write(chunk)
                path.chmod((info.external_attr >> 16) & 0o777 or 0o644)


def install(bundle, destination):
    if sys.platform != 'darwin':
        raise ValueError('private installation requires macOS')
    bundle = bundle.absolute(); destination = destination.expanduser().absolute()
    # Refuse known synced locations and redirected parents; do not strip metadata
    # or quarantine attributes from user files to make signature checks succeed.
    for parent in [destination, *destination.parents]:
        if parent.is_symlink():
            raise ValueError('install destination has a symlink parent')
    home = Path.home()
    for synced in [home/'Documents', home/'Desktop', home/'Library/Mobile Documents', home/'Library/CloudStorage']:
        if destination == synced or synced in destination.parents:
            raise ValueError('use a nonsynced installation directory such as ~/Applications/Vellum')
    manifest_path = bundle/'BUNDLE_MANIFEST.json'; regular(manifest_path, 1 << 20)
    manifest_hash = digest(manifest_path)
    manifest = json.loads(manifest_path.read_text(encoding='utf-8'))
    commit = manifest.get('source_commit', ''); build = manifest.get('build_version', '')
    apps = manifest.get('apps', [])
    if (not re.fullmatch('[0-9a-f]{40}', commit) or not re.fullmatch('[0-9]{1,6}', build)
            or manifest.get('private') is not True or not 1 <= len(apps) <= len(APP_NAMES)):
        raise ValueError('invalid private bundle identity')
    names = [item.get('name') for item in apps]
    if len(set(names)) != len(names) or not set(names) <= APP_NAMES:
        raise ValueError('unexpected or duplicate app')
    destination.mkdir(parents=True, exist_ok=True)
    lock = os.open(destination/'.install.lock', os.O_CREAT | os.O_RDWR | os.O_NOFOLLOW, 0o600)
    with os.fdopen(lock, 'w') as handle:
        fcntl.flock(handle, fcntl.LOCK_EX)
        target = destination/('build'+build+'-'+commit[:12])
        if target.is_symlink():
            raise ValueError('existing version is a symlink; preserved')
        with tempfile.TemporaryDirectory(prefix='.install-', dir=destination) as temporary:
            stage = Path(temporary); payload = stage/'apps'; payload.mkdir()
            all_entries = {}; archives = {}
            for item in apps:
                name = item['name']; filename = item.get('archive', '')
                expected = manifest.get('files', {}).get(filename)
                if (filename != name.replace(' ', '-')+'.zip'
                        or not isinstance(expected, str) or not re.fullmatch('[0-9a-f]{64}', expected)):
                    raise ValueError('invalid app archive identity')
                original = bundle/filename; regular(original, MAX_ARCHIVE)
                archive = stage/filename
                # Copy first; verification and extraction consume the same owned snapshot.
                with original.open('rb') as inp, archive.open('xb') as out:
                    while chunk := inp.read(1 << 20): out.write(chunk)
                if digest(archive) != expected:
                    raise ValueError('app archive SHA256 mismatch')
                admitted = entries(archive, name); all_entries.update(admitted)
                archives[filename] = expected
                if not target.exists(): extract(archive, payload)
            receipt = dict(source_commit=commit, build_version=build, manifest_sha256=manifest_hash,
                           archives=archives, apps=names, signing='ad-hoc', notarized=False)
            if target.exists():
                regular(target/'INSTALLATION.json', 1 << 20)
                if json.loads((target/'INSTALLATION.json').read_text()) != receipt:
                    raise ValueError('existing installation has different identity; preserved')
                verify_tree(target/'apps', all_entries)
                for name in names: signature(target/'apps'/(name+'.app'))
            else:
                verify_tree(payload, all_entries)
                for name in names: signature(payload/(name+'.app'))
                (stage/'INSTALLATION.json').write_text(json.dumps(receipt, indent=2)+'\n')
                # Publish only app bytes and receipt; never overwrite previous versions.
                ready = stage/'ready'; ready.mkdir()
                payload.rename(ready/'apps')
                (stage/'INSTALLATION.json').rename(ready/'INSTALLATION.json')
                ready.rename(target)
            return target


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--bundle', required=True, type=Path)
    parser.add_argument('--destination', type=Path, default=Path.home()/'Applications/Vellum')
    args = parser.parse_args()
    print(install(args.bundle, args.destination))


if __name__ == '__main__': main()
