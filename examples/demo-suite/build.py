#!/usr/bin/env python3
"""Build a source-bound private demo suite from a prepared clean macOS checkout."""
import argparse
from contextlib import contextmanager
import importlib.util
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import sys
import tempfile
sys.dont_write_bytecode = True
from bundle_tools import HERE, ROOT, catalog, clean_source, digest, identity, verify_suite, publish_receipt


def reserve_output(path):
    if path is None:
        return Path(tempfile.mkdtemp(prefix='vellum-demo-build-', dir='/private/tmp'))
    path = path.expanduser().absolute()
    if path != path.resolve() or path.is_relative_to(ROOT):
        raise ValueError('output must be outside the checkout without symlink aliases')
    path.mkdir(mode=0o700)  # exclusive; never replace an earlier delivery
    return path


@contextmanager
def build_locations(output):
    """Delivery may be synced; executable bundles always stay in local staging."""
    delivery = reserve_output(output)
    staging = Path(tempfile.mkdtemp(prefix='vellum-demo-staging-', dir='/private/tmp'))
    try:
        yield delivery, staging
    except BaseException:
        shutil.rmtree(staging, ignore_errors=True)
        raise


def build(output, build_version):
    version, toolchain = identity(ROOT)
    commit = clean_source(ROOT)
    entries = catalog(ROOT, HERE / 'demos.json')
    if sys.platform != 'darwin' or platform.machine() != 'arm64':
        raise ValueError('private demo packaging requires Apple Silicon macOS')
    env = os.environ.copy()
    env.update(GOWORK='off', GOTOOLCHAIN='go' + toolchain,
               CGO_LDFLAGS_ALLOW='-Wl,-rpath,@executable_path/native')
    def run(args, cwd=None):
        subprocess.run(args, cwd=cwd, env=env, check=True)
    for name in ['upstream', 'typesetting']:
        if not (ROOT / 'drivers/cogent/build' / name / 'go.mod').is_file():
            raise ValueError('prepare native dependencies before building')
    with build_locations(output) as (out, staging):
        suite = staging / 'Vellum Demo Suite'
        suite.mkdir()
        binaries = staging / 'binaries'
        binaries.mkdir()
        sys.path.insert(0, str(ROOT / 'scripts'))
        spec = importlib.util.spec_from_file_location('app_packager', ROOT / 'scripts/package-app.py')
        packager = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(packager)
        manifest = dict(schema=1, version=version, build_version=build_version,
                        source_commit=commit, private_build=True, features=[])
        for entry in entries:
            print('Building ' + entry['name'], flush=True)
            binary = binaries / entry['id']
            run(['go', 'build', '-buildvcs=false', '-mod=readonly', *entry['flags'], '-o', str(binary), entry['target']], ROOT / entry['directory'])
            result = dict(entry, input_binary_sha256=digest(binary))
            if entry['id'] == 'replay':
                target = suite / 'replay'
                target.write_bytes(binary.read_bytes())
                target.chmod(0o755)
            else:
                bundle = suite / (entry['name'] + '.app')
                packager.package(binary, bundle, entry['name'], 'com.dshills.vellum.suite.' + entry['id'], version[1:], 'darwin', None, build_version)
                stamp = bundle / 'Contents/Resources/APP_MANIFEST.json'
                data = json.loads(stamp.read_text())
                data.update(source_commit=commit, signed_identity='-')
                stamp.write_text(json.dumps(data, indent=2) + '\n')
                run(['xattr', '-cr', str(bundle)])
                run(['codesign', '--force', '--sign', '-', str(bundle)])
                target = bundle / 'Contents/MacOS' / entry['name']
            result['packaged_binary_sha256'] = digest(target)
            manifest['features'].append(result)
        if clean_source(ROOT) != commit:
            raise ValueError('source changed during build; suite not promoted')
        (suite / 'BUILD.json').write_text(json.dumps(manifest, indent=2) + '\n')
        manifest_hash = digest(suite / 'BUILD.json')
        verify_suite(suite, manifest_hash)
        local_archive = staging / 'Vellum Demo Suite.zip'
        run(['ditto', '-c', '-k', '--norsrc', '--noextattr', '--keepParent', str(suite), str(local_archive)])
        archive = out / local_archive.name
        shutil.copyfile(local_archive, archive)
        if digest(archive) != digest(local_archive):
            raise ValueError('delivery archive changed during copy; suite not promoted')
        # Verify the delivered archive, rather than only the pre-archive working tree.
        extracted = staging / 'verified-extraction'
        extracted.mkdir()
        run(['ditto', '-x', '-k', str(archive), str(extracted)])
        verify_suite(extracted / suite.name, manifest_hash)
        receipt = dict(schema=1, version=version, build_version=build_version, source_commit=commit, suite=str(suite),
                       manifest_sha256=manifest_hash, archive=str(archive), archive_sha256=digest(archive),
                       bundles_verified=sum(e['id'] != 'replay' for e in entries), native_runtime='not executed')
        if clean_source(ROOT) != commit:
            raise ValueError('source changed during archive verification; suite not promoted')
        (out / 'DELIVERY.json').write_text(json.dumps(receipt, indent=2) + '\n')
        publish_receipt(receipt)
        print('Ready: ' + str(suite / 'Vellum Demo Suite.app'), flush=True)
        return receipt


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, help='new output outside the checkout; default: fresh /private/tmp directory')
    parser.add_argument('--build-version', default='1')
    args = parser.parse_args()
    if not args.build_version.isdecimal() or not 1 <= len(args.build_version) <= 6:
        parser.error('build version must have one to six decimal digits')
    try:
        build(args.output, args.build_version)
    except (OSError, ValueError, KeyError, TypeError, subprocess.CalledProcessError) as error:
        parser.exit(1, str(error) + '\n')


if __name__ == '__main__':
    main()
