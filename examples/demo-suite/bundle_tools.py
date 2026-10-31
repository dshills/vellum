"""Shared local demo identity and verification helpers; no desktop actions."""
import hashlib
import json
from pathlib import Path
import plistlib
import re
import subprocess
import tempfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]


def digest(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def identity(root):
    release = json.loads((root / 'release/version.json').read_text())
    version = release['version']
    if not re.fullmatch(r'v\d+\.\d+\.\d+', version):
        raise ValueError('numeric stable framework version required')
    source = (root / 'vellum.go').read_text()
    if not re.search(r'const Version = "' + re.escape(version) + '"', source):
        raise ValueError('framework and release versions disagree')
    toolchain = re.search(r'^go (\d+\.\d+\.\d+)$', (root / 'go.mod').read_text(), re.M)
    if not toolchain:
        raise ValueError('exact Go version required')
    return version, toolchain[1]


def git(root, *args):
    return subprocess.check_output(['git', '-C', str(root), *args], text=True).strip()


def clean_source(root):
    if Path(git(root, 'rev-parse', '--show-toplevel')).resolve() != root.resolve():
        raise ValueError('build from the source checkout, not an extracted SDK')
    if git(root, 'status', '--porcelain'):
        raise ValueError('commit source changes before building a source-bound suite')
    return git(root, 'rev-parse', 'HEAD')


def catalog(root, path):
    entries = json.loads(path.read_text())
    seen = set()
    for entry in entries:
        key = entry['id']
        if not re.fullmatch(r'[a-z][a-z0-9-]*', key) or key in seen:
            raise ValueError('duplicate or invalid demo id')
        seen.add(key)
        directory = Path(entry['directory'])
        if directory.is_absolute() or '..' in directory.parts or not (root / directory / 'go.mod').is_file():
            raise ValueError('demo directory must name a checkout module')
        if not re.fullmatch(r'[A-Za-z0-9 ]+', entry['name']):
            raise ValueError('invalid demo name')
        if not isinstance(entry['flags'], list) or not all(isinstance(v, str) for v in entry['flags']):
            raise ValueError('invalid build flags')
        if not isinstance(entry['target'], str) or not entry['target'].startswith('.'):
            raise ValueError('invalid module target')
    if not {'hub', 'replay'} <= seen:
        raise ValueError('hub and replay entries required')
    return entries


def verify_suite(suite, expected_manifest):
    if digest(suite / 'BUILD.json') != expected_manifest:
        raise ValueError('suite manifest differs from selected receipt')
    manifest = json.loads((suite / 'BUILD.json').read_text())
    for entry in manifest['features']:
        if entry['id'] == 'replay':
            binary = suite / 'replay'
        else:
            if not isinstance(entry.get('name'), str) or not re.fullmatch(r'[A-Za-z0-9 ]+', entry['name']):
                raise ValueError('invalid demo name in manifest')
            bundle = suite / (entry['name'] + '.app')
            info = plistlib.loads((bundle / 'Contents/Info.plist').read_bytes())
            if info['CFBundleShortVersionString'] != manifest['version'].removeprefix('v') or info['CFBundleVersion'] != manifest['build_version']:
                raise ValueError('app version differs from suite')
            binary = bundle / 'Contents/MacOS' / entry['name']
            stamp = json.loads((bundle / 'Contents/Resources/APP_MANIFEST.json').read_text())
            if stamp.get('source_commit') != manifest['source_commit']:
                raise ValueError('app source differs from suite')
            subprocess.run(['codesign', '--verify', '--strict', str(bundle)], check=True, capture_output=True)
        if digest(binary) != entry['packaged_binary_sha256']:
            raise ValueError('packaged executable differs: ' + entry['id'])
    return manifest


def publish_receipt(receipt):
    pointer = HERE / 'build'
    pointer.mkdir(exist_ok=True)
    temporary = None
    try:
        with tempfile.NamedTemporaryFile(mode='w', prefix='.suite-', dir=pointer, delete=False) as stream:
            temporary = Path(stream.name)
            stream.write(json.dumps(receipt, indent=2) + '\n')
        temporary.replace(pointer / 'latest-suite.json')
    finally:
        if temporary is not None and temporary.exists():
            temporary.unlink()
