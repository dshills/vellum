"""Cheap private producer admission and sampled compiler/preparation identities.

This is an early diagnostic, not a replacement for pinned preparation, module
integrity, SDK relocation, signatures or a hermetic compiler attestation.
"""
import hashlib
import json
from pathlib import Path
import re
import shutil

from child_process import check_child_output
from native_toolchain import capture as native_capture

SENTINELS = {
    'upstream': ('cogentcore.org/core', ('go.mod', 'core/scene.go', 'system/driver/desktop/window.go')),
    'typesetting': ('github.com/go-text/typesetting', ('go.mod', 'font/font.go')),
}


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        while chunk := stream.read(1 << 20):
            h.update(chunk)
    return h.hexdigest()


def location(path, root, *, new):
    requested = path.expanduser().absolute()
    if new and (requested.exists() or requested.is_symlink()):
        raise ValueError('producer output already exists; use a new directory')
    selected = requested.resolve()
    source = root.resolve()
    if selected == source or source in selected.parents:
        raise ValueError('private producer locations must be outside the source checkout')
    home = Path.home().resolve()
    for entry in ('Documents', 'Desktop', 'Library/Mobile Documents', 'Library/CloudStorage'):
        synced = (home/entry).resolve()
        if selected == synced or synced in selected.parents:
            raise ValueError('use a nonsynced producer/installation directory; copy completed archives to Documents afterward')
    return selected


def preflight(root, output, env, install_to=None):
    output = location(output, root, new=True)
    destination = location(install_to, root, new=False) if install_to is not None else None
    match = re.search(r'^go\s+(\d+\.\d+\.\d+)\s*$', (root/'go.mod').read_text(), re.M)
    if match is None:
        raise ValueError('private producer requires an exact go.mod toolchain version')
    executable = shutil.which('go', path=env.get('PATH', ''))
    if executable is None:
        raise ValueError('Go executable is unavailable; put the declared Go version on PATH')
    executable = Path(executable).resolve(strict=True)
    config = json.loads(check_child_output([str(executable), 'env', '-json', 'GOVERSION', 'GOOS',
        'GOARCH', 'CGO_ENABLED', 'GOROOT', 'GOTOOLDIR', 'CC', 'CXX',
        'CGO_CFLAGS', 'CGO_CPPFLAGS', 'CGO_CXXFLAGS', 'CGO_LDFLAGS'], cwd=root, env=env, text=True, timeout=20))
    wanted = 'go'+match.group(1)
    if config.get('GOVERSION') != wanted:
        raise ValueError(f"private producer requires {wanted}; PATH selects {config.get('GOVERSION')}")
    if (config.get('GOOS'), config.get('GOARCH'), config.get('CGO_ENABLED')) != ('darwin', 'arm64', '1'):
        raise ValueError('private producer requires darwin/arm64 with CGO_ENABLED=1')
    tools = {'go': executable}
    for name in ('compile', 'link'):
        tools[name] = Path(config['GOTOOLDIR'])/name
    identities = {}
    for name, path in tools.items():
        if not path.is_file() or path.is_symlink():
            raise ValueError('expected regular toolchain binary: '+str(path))
        identities[name] = {'path': str(path), 'sha256': digest(path)}
    driver = root/'drivers/cogent'
    if (driver/'build/.prepare-lock').exists():
        raise ValueError('native preparation is active; wait before packaging')
    manifest = driver/'native/dependencies.json'
    if not manifest.is_file():
        raise ValueError('native dependency manifest is missing')
    dependencies = {entry['name']: entry for entry in json.loads(manifest.read_text())['dependencies']}
    checked = {}
    for name, (module, files) in SENTINELS.items():
        base = driver/'build'/name
        if dependencies.get(name, {}).get('module') != module:
            raise ValueError('native preparation manifest does not match '+name)
        for relative in files:
            path = base/relative
            if not path.is_file() or path.is_symlink() or not path.stat().st_size:
                raise ValueError('incomplete native preparation: '+str(path)+'; run drivers/cogent/prepare.sh')
            checked[name+'/'+relative] = digest(path)
        if not re.search(r'^module\s+'+re.escape(module)+r'\s*$', (base/'go.mod').read_text(), re.M):
            raise ValueError('wrong prepared native module: '+name)
    return {'output': str(output), 'install_to': str(destination) if destination else None,
        'toolchain': {'configuration': config, 'binaries': identities},
        'native_toolchain': native_capture(config, env),
        'native_preparation': {'scope': 'presence/module identity and sampled sentinel hashes; full preparation/SDK verification still required',
            'manifest_sha256': digest(manifest), 'checked_files': checked}}


def verify_snapshot(root, admission, env=None):
    if 'native_toolchain' in admission:
        if env is None or native_capture(admission['toolchain']['configuration'], env) != admission['native_toolchain']:
            raise ValueError('native compiler/linker/SDK selection changed during packaging')
    for item in admission['toolchain']['binaries'].values():
        if digest(Path(item['path'])) != item['sha256']:
            raise ValueError('toolchain changed during packaging')
    for name, expected in admission['native_preparation']['checked_files'].items():
        if digest(root/'drivers/cogent/build'/name) != expected:
            raise ValueError('sampled native preparation changed during packaging')
    if digest(root/'drivers/cogent/native/dependencies.json') != admission['native_preparation']['manifest_sha256']:
        raise ValueError('native manifest changed during packaging')
