#!/usr/bin/env python3
"""Build a private SDK and four apps from that extracted SDK; never publishes."""
import argparse
import hashlib
import importlib.util
import json
import platform
from pathlib import Path
import sys
sys.dont_write_bytecode = True
from child_process import run_child, check_child_output
from sdk_environment import child_environment
ROOT = Path(__file__).resolve().parents[1]


def clean_source(root, env):
    revision = check_child_output(['git', 'rev-parse', 'HEAD'], cwd=root, env=env, text=True, timeout=20).strip()
    if check_child_output(['git', 'status', '--porcelain', '--untracked-files=normal'], cwd=root, env=env, text=True, timeout=20):
        raise ValueError('commit source changes before private packaging')
    return revision


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', required=True, type=Path)
    args = parser.parse_args()
    if platform.system() != 'Darwin' or platform.machine() != 'arm64':
        parser.error('requires an Apple Silicon Mac')
    env = child_environment()
    revision = clean_source(ROOT, env)
    out = args.output.absolute()
    out.mkdir()  # exclusive: failed attempts retain partial output
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
        run([sys.executable, '-B', sdk/'scripts/package-app.py', '--binary', binary, '--output', bundle,
             '--name', name, '--identifier', identifier, '--version', '0.1.0', '--build-version', '19', '--sign-identity', '-', *options])
        run(['codesign', '--verify', '--deep', '--strict', bundle])
        zipped = out / (name.replace(' ', '-') + '.zip')
        run(['ditto', '-c', '-k', '--norsrc', '--noextattr', '--keepParent', bundle, zipped])
        extracted = work / ('ZIP check ' + str(index)); extracted.mkdir()
        run(['ditto', '-x', '-k', zipped, extracted])
        run(['codesign', '--verify', '--deep', '--strict', extracted/(name+'.app')])
        hashes[zipped.name] = digest(zipped)
        app_records.append(dict(name=name, archive=zipped.name, module=str(module.relative_to(out)), binary_sha256=digest(binary)))
    if clean_source(ROOT, env) != revision:
        raise ValueError('source changed during private packaging')
    (out/'SHA256SUMS').write_text(''.join(f'{value}  {name}\n' for name,value in sorted(hashes.items())), encoding='utf-8')
    (out/'BUNDLE_MANIFEST.json').write_text(json.dumps(dict(source_commit=revision, private=True, signing='ad-hoc', notarized=False,
            sdk_sha256=checksum, apps=app_records, files=hashes), indent=2)+'\n', encoding='utf-8')
    (out/'README.md').write_text('# Private Vellum consumer bundle\n\nAll apps were built from the included, independently verified extracted SDK.\nVerify SHA256SUMS before extraction. Only named archives in SHA256SUMS are delivery payloads. build/ contains intermediate SDKs, consumer source, binaries and ZIP checks for local testing; do not treat those working files as checksummed delivery payloads.\n\nForge and Measure Pad run normally. Documents requires -data-dir /absolute/private/workspace; it opens independent A/B repositories. Document Starter requires -data-dir /absolute/private/document-data. Use fresh private directories for testing. No app defaults to Vellum Notes storage.\n\nApps are ad-hoc signed and unnotarized. Only the recorded Apple Silicon Mac configuration is qualified; human IME/VoiceOver and downloaded-quarantine launch remain unqualified. No public release or audit approval is implied.\n\nThe extracted SDK includes source, both starter profiles, guides and known limits. Do not run prepare.sh over it. Failed producer attempts retain partial output and never write BUNDLE_MANIFEST.json as a successful result.\n', encoding='utf-8')
    print(json.dumps(dict(output=str(out), source_commit=revision, apps=len(apps), private=True)))

if __name__ == '__main__':
    main()
