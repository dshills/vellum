#!/usr/bin/env python3
"""Build a private SDK and five apps from that extracted SDK; optionally install without launching; never publishes."""
import argparse
import hashlib
import importlib.util
import json
import platform
import os
import re
from pathlib import Path
import sys
sys.dont_write_bytecode = True
from child_process import run_child, check_child_output
from sdk_environment import child_environment
from private_bundle_preflight import preflight, verify_snapshot
ROOT = Path(__file__).resolve().parents[1]


def clean_source(root, env):
    revision = check_child_output(['git', 'rev-parse', 'HEAD'], cwd=root, env=env, text=True, timeout=20).strip()
    if check_child_output(['git', 'status', '--porcelain', '--untracked-files=normal'], cwd=root, env=env, text=True, timeout=20):
        raise ValueError('commit source changes before private packaging')
    return revision


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_script(name):
    spec = importlib.util.spec_from_file_location('private_bundle_' + name.replace('-', '_'), ROOT/'scripts'/(name+'.py'))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def build_version(value):
    if not re.fullmatch(r'[0-9]{1,6}', value):
        raise argparse.ArgumentTypeError('build version must contain 1..6 decimal digits')
    return value


def install_delivery(out, destination):
    """Verify completed producer bytes and delegate installation without launch."""
    installer = load_script('install-private-macos')
    manifest_path = out/'BUNDLE_MANIFEST.json'
    installer.regular(manifest_path, 1 << 20)
    manifest_digest = installer.digest(manifest_path)
    manifest = json.loads(manifest_path.read_text(encoding='utf-8'))
    files = manifest.get('files', {})
    sdk_name = 'Vellum-SDK.tar.gz'
    apps = manifest.get('apps', [])
    if (not isinstance(apps, list) or len(apps) != len(installer.APP_NAMES)
            or any(not isinstance(item, dict) for item in apps)
            or {item.get('name') for item in apps} != installer.APP_NAMES):
        raise ValueError('expected the complete five-app producer bundle')
    app_names = [item['archive'] for item in apps]
    if (manifest.get('private') is not True or manifest.get('signing') != 'ad-hoc'
            or manifest.get('notarized') is not False or not isinstance(files, dict)
            or set(files) != {sdk_name, *app_names}
            or manifest.get('sdk_sha256') != files.get(sdk_name)):
        raise ValueError('inconsistent private producer manifest')
    for name, expected in files.items():
        if (Path(name).name != name or not isinstance(expected, str)
                or not re.fullmatch(r'[0-9a-f]{64}', expected)):
            raise ValueError('invalid payload identity')
        installer.regular(out/name, installer.MAX_ARCHIVE)
        if installer.digest(out/name) != expected:
            raise ValueError('payload SHA256 mismatch: ' + name)
    installer.regular(out/'SHA256SUMS', 1 << 20)
    checksums = ''.join(f'{value}  {name}\n' for name, value in sorted(files.items()))
    if (out/'SHA256SUMS').read_text(encoding='utf-8') != checksums:
        raise ValueError('SHA256SUMS disagrees with producer manifest')
    if 'preflight_sha256' in manifest:
        installer.regular(out/'PREFLIGHT.json', 1 << 20)
        if installer.digest(out/'PREFLIGHT.json') != manifest['preflight_sha256']:
            raise ValueError('preflight receipt identity mismatch')
    # Verify archive contents and the SDK's recorded clean source, not just a hash.
    checker = load_script('check-sdk')
    verified = checker.verify_archive(out/sdk_name, manifest['sdk_sha256'])
    if verified['manifest'].get('source') != {'commit': manifest['source_commit'], 'clean': True}:
        raise ValueError('SDK source identity mismatch')
    receipt_path = out/'DELIVERY.json'
    if receipt_path.exists() or receipt_path.is_symlink():
        raise ValueError('delivery receipt already exists; preserve the completed delivery')
    target = installer.install(out, destination)
    installation_path = target/'INSTALLATION.json'
    installer.regular(installation_path, 1 << 20)
    installation = json.loads(installation_path.read_text(encoding='utf-8'))
    if (installer.digest(manifest_path) != manifest_digest
            or installation.get('manifest_sha256') != manifest_digest
            or installation.get('source_commit') != manifest['source_commit']
            or installation.get('build_version') != manifest['build_version']
            or installation.get('archives') != {name: files[name] for name in app_names}):
        raise ValueError('installed producer identity changed')
    receipt = dict(source_commit=manifest['source_commit'], build_version=manifest['build_version'],
                   sdk_sha256=manifest['sdk_sha256'], manifest_sha256=manifest_digest,
                   installation=str(target), installation_sha256=installer.digest(installation_path),
                   private=True, signing='ad-hoc', notarized=False, native_runtime='not executed')
    if 'preflight_sha256' in manifest:
        receipt['preflight_sha256'] = manifest['preflight_sha256']
    with receipt_path.open('x', encoding='utf-8') as stream:
        stream.write(json.dumps(receipt, indent=2)+'\n')
    return receipt


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', required=True, type=Path)
    parser.add_argument('--build-version', type=build_version, default='22')
    parser.add_argument('--install-to', type=Path, help='verify and install without launching; use a nonsynced directory')
    args = parser.parse_args()
    if platform.system() != 'Darwin' or platform.machine() != 'arm64':
        parser.error('requires an Apple Silicon Mac')
    env = child_environment()
    revision = clean_source(ROOT, env)
    admission = preflight(ROOT, args.output, env, args.install_to)
    # Use the resolved admitted Go driver even if an ambient PATH alias changes.
    env['PATH'] = str(Path(admission['toolchain']['binaries']['go']['path']).parent) + os.pathsep + env.get('PATH', '')
    out = Path(admission['output'])
    out.mkdir()  # exclusive: failed attempts retain partial output
    (out/'PREFLIGHT.json').write_text(json.dumps(dict(source_commit=revision, **admission), indent=2)+'\n')
    admission_digest = digest(out/'PREFLIGHT.json')
    work = out / "build"; work.mkdir()
    def run(argv, cwd=ROOT):
        run_child([str(a) for a in argv], cwd=cwd, env=env, check=True, timeout=600)
    archive = out / 'Vellum-SDK.tar.gz'
    run([sys.executable, '-B', ROOT/'scripts/release-sdk.py', '--output', archive])
    checksum = digest(archive)
    run([sys.executable, '-B', ROOT/'scripts/check-sdk.py', '--archive', archive, '--sha256', checksum])
    spec = importlib.util.spec_from_file_location('bundle_sdk_checker', ROOT/'scripts/check-sdk.py')
    checker = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(checker)
    verified = checker.verify_archive(archive, checksum)
    source = verified['manifest'].get('source', {})
    if not isinstance(source, dict) or source.get('clean') is not True or source.get('commit') != revision:
        raise ValueError('SDK source identity mismatch: ' + repr(source))
    sdk = checker.extract_verified(archive, work/'Extracted SDK', verified)
    consumer = work/'Document Starter'
    run([sys.executable, '-B', sdk/'scripts/new-app.py', consumer, '--module=example.com/private-document', '--template=document'])
    apps = [
        ('Vellum Forge', sdk/'examples/forge', '.', 'com.dshills.vellum.forge.demo'),
        ('Vellum Measure Pad', sdk/'examples/measurepad', '.', 'com.dshills.vellum.measurepad.demo'),
        ('Vellum Documents', sdk/'examples/macapps', './cmd/documents', 'com.dshills.vellum.documents.demo'),
        ('Vellum Document Starter', consumer, '.', 'com.dshills.vellum.documentstarter.demo'),
        ('Vellum Notes', sdk/'examples/macapps', './cmd/notes', 'com.dshills.vellum.notes.demo'),
    ]
    hashes = {archive.name: checksum}
    app_records = []
    for index, (name, module, package, identifier) in enumerate(apps):
        modules = checker.decode_objects(check_child_output(['go', 'list', '-m', '-json', 'all'], cwd=module, env=env, text=True, timeout=120))
        checker.check_origin(modules, sdk, verified['manifest']['release']['module'], verified['manifest'].get('native_gpu'))
        binary = work / ('app-' + str(index))
        run(['go', 'build', '-trimpath', '-o', binary, package], module)
        bundle = work / (name + '.app')
        options = []
        if name == 'Vellum Forge':
            icon = work/'Forge.png'
            run([binary, '--write-icon', icon])
            options = ['--icon', icon]
        if name in {'Vellum Forge', 'Vellum Notes', 'Vellum Document Starter'}:
            extension = 'vforge' if name == 'Vellum Forge' else 'vellumnotes'
            types = work / ('document-types-' + str(index) + '.json')
            types.write_text(json.dumps([dict(extension=extension,
                uti='com.dshills.vellum.' + extension, name='Vellum ' + extension + ' document')]))
            options += ['--document-types', types]
        run([sys.executable, '-B', sdk/'scripts/package-app.py', '--binary', binary, '--output', bundle,
             '--name', name, '--identifier', identifier, '--version', '0.1.0', '--build-version', args.build_version, '--sign-identity', '-', *options])
        run(['codesign', '--verify', '--deep', '--strict', bundle])
        zipped = out / (name.replace(' ', '-') + '.zip')
        run(['ditto', '-c', '-k', '--norsrc', '--noextattr', '--keepParent', bundle, zipped])
        extracted = work / ('ZIP check ' + str(index)); extracted.mkdir()
        run(['ditto', '-x', '-k', zipped, extracted])
        run(['codesign', '--verify', '--deep', '--strict', extracted/(name+'.app')])
        hashes[zipped.name] = digest(zipped)
        app_records.append(dict(name=name, archive=zipped.name, module=str(module.relative_to(out)), binary_sha256=digest(binary)))
    verify_snapshot(ROOT, admission)
    if digest(out/'PREFLIGHT.json') != admission_digest:
        raise ValueError('preflight receipt changed during packaging')
    if clean_source(ROOT, env) != revision:
        raise ValueError('source changed during private packaging')
    (out/'SHA256SUMS').write_text(''.join(f'{value}  {name}\n' for name,value in sorted(hashes.items())), encoding='utf-8')
    (out/'README.md').write_text('# Private Vellum consumer bundle\n\nAll apps were built from the included, independently verified extracted SDK.\nVerify SHA256SUMS before extraction. Only named archives in SHA256SUMS are delivery payloads. build/ contains intermediate SDKs, consumer source, binaries and ZIP checks for local testing; do not treat those working files as checksummed delivery payloads.\n\nForge, Notes and Measure Pad run normally. Notes uses its established Application Support/Vellum Notes directory; use --args -data-dir /absolute/private/notes-test for isolated checks. Documents requires -data-dir /absolute/private/workspace; it opens independent A/B repositories. Document Starter launches normally from Finder and uses its own module-derived Application Support directory. For isolated tests or an existing private notebook, launch it with --args -data-dir /absolute/private/document-data. Use fresh private directories for testing. No generated starter defaults to Vellum Notes storage.\n\nApps are ad-hoc signed and unnotarized. Only the recorded Apple Silicon Mac configuration is qualified; human IME/VoiceOver and downloaded-quarantine launch remain unqualified. No public release or audit approval is implied.\n\nThe extracted SDK includes source, both starter profiles, guides and known limits. Do not run prepare.sh over it. Failed producer attempts retain partial output. BUNDLE_MANIFEST.json records completed packaging; optional installation completes only when DELIVERY.json exists. No app is launched by packaging or installation.\n', encoding='utf-8')
    # Publish the completion record only after all other delivery metadata exists.
    (out/'BUNDLE_MANIFEST.json').write_text(json.dumps(dict(source_commit=revision, build_version=args.build_version, private=True, signing='ad-hoc', notarized=False,
            sdk_sha256=checksum, apps=app_records, files=hashes, preflight_sha256=digest(out/'PREFLIGHT.json')), indent=2)+'\n', encoding='utf-8')
    if args.install_to is not None:
        install_delivery(out, Path(admission['install_to']))
    print(json.dumps(dict(output=str(out), source_commit=revision, apps=len(apps), private=True)))

if __name__ == '__main__':
    main()
