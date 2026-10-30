#!/usr/bin/env python3
"""Verify an existing private delivery and installation without building or launching."""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import plistlib
import macos_deployment
import re
import sys
import tempfile

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent
IDENTIFIERS = {
    'Vellum Forge': 'com.dshills.vellum.forge.demo',
    'Vellum Measure Pad': 'com.dshills.vellum.measurepad.demo',
    'Vellum Documents': 'com.dshills.vellum.documents.demo',
    'Vellum Document Starter': 'com.dshills.vellum.documentstarter.demo',
    'Vellum Notes': 'com.dshills.vellum.notes.demo',
}


def load(name):
    spec = importlib.util.spec_from_file_location(name.replace('-', '_'), ROOT/(name+'.py'))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


installer = load('install-private-macos')
checker = load('check-sdk')


def read_regular(path, limit=1 << 20):
    installer.regular(path, limit)
    with path.open('rb') as stream:
        data = stream.read(limit+1)
    if len(data) > limit:
        raise ValueError('input grew beyond size limit')
    return data


def record(path):
    data = read_regular(path)
    value = json.loads(data)
    if not isinstance(value, dict):
        raise ValueError('expected object receipt: '+path.name)
    return value, hashlib.sha256(data).hexdigest()


def no_links(path):
    for parent in (path, *path.parents):
        if parent.is_symlink():
            raise ValueError('verification path has a symlink parent')


def require_fields(actual, expected, label):
    # Strict bool handling avoids accepting 0 as False or 1 as True in receipts.
    if not isinstance(actual, dict) or any(type(actual.get(k)) is not type(v) or actual.get(k) != v for k, v in expected.items()):
        raise ValueError(label+' identity mismatch')


def verify_app(app, item, build):
    info = plistlib.loads(read_regular(app/'Contents/Info.plist'))
    require_fields(info, dict(CFBundleName=item['name'], CFBundleExecutable=item['name'],
                             CFBundleIdentifier=IDENTIFIERS[item['name']], CFBundleVersion=build), 'app plist')
    manifest, _ = record(app/'Contents/Resources/APP_MANIFEST.json')
    require_fields(manifest, dict(name=item['name'], identifier=IDENTIFIERS[item['name']],
        build_version=build, source_binary_sha256=item['binary_sha256'],
        private_build=True, signed_identity='-', notarized=False), 'app manifest')
    deployment=manifest.get('macos_deployment')
    macos_deployment.verify(info,deployment,macos_deployment.inspect(app/'Contents/MacOS'/item['name']) if deployment is not None else None)
    # The source binary hash is pre-signing; codesign binds the final executable
    # and resource manifest. Do not compare a pre-signing hash to signed bytes.
    installer.signature(app)


def verify(bundle, installation, expected_manifest=None):
    if sys.platform != 'darwin':
        raise ValueError('signature verification requires macOS')
    bundle = bundle.expanduser().absolute()
    installation = installation.expanduser().absolute()
    no_links(bundle); no_links(installation)
    manifest, manifest_hash = record(bundle/'BUNDLE_MANIFEST.json')
    if expected_manifest is not None and (not re.fullmatch('[0-9a-f]{64}', expected_manifest)
                                         or expected_manifest != manifest_hash):
        raise ValueError('caller-supplied manifest SHA256 mismatch')
    commit = manifest.get('source_commit'); build = manifest.get('build_version')
    if (not isinstance(commit, str) or not re.fullmatch('[0-9a-f]{40}', commit)
            or not isinstance(build, str) or not re.fullmatch('[0-9]{1,6}', build)):
        raise ValueError('invalid source/build identity')
    require_fields(manifest, dict(private=True, signing='ad-hoc', notarized=False), 'private bundle')
    apps = manifest.get('apps'); files = manifest.get('files')
    if (not isinstance(apps, list) or len(apps) != len(IDENTIFIERS)
            or any(not isinstance(item, dict) or not isinstance(item.get('name'), str) for item in apps)
            or {item['name'] for item in apps} != set(IDENTIFIERS)):
        raise ValueError('expected complete five-app bundle')
    archives = {}
    for item in apps:
        filename = item['name'].replace(' ', '-')+'.zip'
        if (item.get('archive') != filename or not isinstance(item.get('binary_sha256'), str)
                or not re.fullmatch('[0-9a-f]{64}', item['binary_sha256'])):
            raise ValueError('invalid app identity')
        archives[filename] = None
    if (not isinstance(files, dict) or set(files) != {'Vellum-SDK.tar.gz', *archives}
            or any(not isinstance(h, str) or not re.fullmatch('[0-9a-f]{64}', h) for h in files.values())
            or manifest.get('sdk_sha256') != files['Vellum-SDK.tar.gz']):
        raise ValueError('invalid payload identities')
    sums = ''.join(f'{h}  {name}\n' for name, h in sorted(files.items())).encode()
    if read_regular(bundle/'SHA256SUMS') != sums:
        raise ValueError('SHA256SUMS disagrees with manifest')
    delivery, _ = record(bundle/'DELIVERY.json')
    installed, installed_hash = record(installation/'INSTALLATION.json')
    archives = {name: files[name] for name in archives}
    expected_install = dict(source_commit=commit, build_version=build, manifest_sha256=manifest_hash,
                           archives=archives, apps=[item['name'] for item in apps], signing='ad-hoc', notarized=False)
    require_fields(installed, expected_install, 'installation receipt')
    if set(installed) != set(expected_install):
        raise ValueError('installation receipt field mismatch')
    require_fields(delivery, dict(source_commit=commit, build_version=build, sdk_sha256=manifest['sdk_sha256'],
        manifest_sha256=manifest_hash, installation_sha256=installed_hash,
        private=True, signing='ad-hoc', notarized=False, native_runtime='not executed'), 'delivery')
    # Do not follow a path from an untrusted receipt: only compare with the
    # caller-selected installed directory, then read that directory directly.
    if delivery.get('installation') != str(installation):
        raise ValueError('delivery installation path mismatch')
    if (bundle/'INSTALLATION.json').exists() or (bundle/'INSTALLATION.json').is_symlink():
        if read_regular(bundle/'INSTALLATION.json') != read_regular(installation/'INSTALLATION.json'):
            raise ValueError('copied installation receipt mismatch')
    if 'preflight_sha256' in manifest:
        preflight, preflight_hash = record(bundle/'PREFLIGHT.json')
        if preflight_hash != manifest['preflight_sha256'] or preflight.get('source_commit') != commit:
            raise ValueError('preflight source/hash mismatch')
        require_fields(delivery, dict(preflight_sha256=preflight_hash), 'delivery preflight')
    elif 'preflight_sha256' in delivery:
        raise ValueError('unbound delivery preflight')
    no_links(installation/'apps')
    all_entries = {}
    with tempfile.TemporaryDirectory(prefix='vellum-delivery-verification-') as temporary:
        stage = Path(temporary)
        # Hash and consume the same owned snapshots. No writes to bundle/install.
        for name, expected in files.items():
            source = bundle/name; installer.regular(source, installer.MAX_ARCHIVE)
            copied = stage/name; total = 0
            with source.open('rb') as inp, copied.open('xb') as out:
                while data := inp.read(1 << 20):
                    total += len(data)
                    if total > installer.MAX_ARCHIVE:
                        raise ValueError('archive grew beyond size limit')
                    out.write(data)
            if installer.digest(copied) != expected:
                raise ValueError('payload SHA256 mismatch: '+name)
        sdk = checker.verify_archive(stage/'Vellum-SDK.tar.gz', manifest['sdk_sha256'])
        source = sdk['manifest'].get('source')
        require_fields(source, dict(commit=commit, clean=True), 'SDK source')
        if set(source) != {'commit', 'clean'}:
            raise ValueError('SDK source field mismatch')
        payload = stage/'apps'; payload.mkdir()
        for item in apps:
            archive = stage/item['archive']
            all_entries.update(installer.entries(archive, item['name']))
            installer.extract(archive, payload)
        installer.verify_tree(payload, all_entries)
        installer.verify_tree(installation/'apps', all_entries)
        for item in apps:
            verify_app(payload/(item['name']+'.app'), item, build)
            verify_app(installation/'apps'/(item['name']+'.app'), item, build)
    return dict(source_commit=commit, build_version=build, manifest_sha256=manifest_hash,
                sdk_sha256=manifest['sdk_sha256'], apps_verified=len(apps),
                checks=['SDK all-file integrity/source', 'archive hashes/checksums', 'receipt linkage',
                        'app metadata', 'installed bytes', 'archive and installed strict signatures'],
                trust_basis='caller-supplied manifest digest' if expected_manifest else 'local self-attestation',
                independent_authenticity_verified=False, native_runtime='not executed',
                installation=str(installation))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--bundle', type=Path, required=True)
    parser.add_argument('--installation', type=Path, required=True)
    parser.add_argument('--manifest-sha256', help='optional digest from a separately trusted record')
    args = parser.parse_args()
    try:
        result = verify(args.bundle, args.installation, args.manifest_sha256)
    except (ValueError, OSError) as error:
        parser.exit(1, str(error)+'\n')
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
