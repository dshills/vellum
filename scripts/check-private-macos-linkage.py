#!/usr/bin/env python3
"""Inspect the installed private five-app system-library route without execution."""
import argparse
import importlib.util
import json
from pathlib import Path, PurePosixPath
import platform
import plistlib
import re
import subprocess
import sys

sys.dont_write_bytecode = True
from child_process import check_child_output
from sdk_environment import child_environment

spec = importlib.util.spec_from_file_location('private_verifier', Path(__file__).with_name('verify-private-macos.py'))
v = importlib.util.module_from_spec(spec); spec.loader.exec_module(v)
LOADS = {'LC_LOAD_DYLIB', 'LC_LOAD_WEAK_DYLIB', 'LC_REEXPORT_DYLIB',
         'LC_LOAD_UPWARD_DYLIB', 'LC_LAZY_LOAD_DYLIB'}
MACH = {b'\xcf\xfa\xed\xfe', b'\xfe\xed\xfa\xcf', b'\xce\xfa\xed\xfe',
        b'\xfe\xed\xfa\xce', b'\xca\xfe\xba\xbe', b'\xbe\xba\xfe\xca',
        b'\xca\xfe\xba\xbf', b'\xbf\xba\xfe\xca'}


def version(value):
    if not isinstance(value, str) or not re.fullmatch(r'\d{1,3}(?:\.\d{1,3}){0,2}', value):
        raise ValueError('invalid macOS version')
    return tuple([int(x) for x in value.split('.')] + [0]*(3-len(value.split('.'))))


def field(block, key, *, path=False):
    values = re.findall(r'^\s*'+re.escape(key)+r'\s+(.+?)\s*$', block, re.M)
    if len(values) != 1:
        raise ValueError('missing or duplicate load-command field: '+key)
    value = values[0]
    if path:
        match = re.fullmatch(r'(.+) \(offset \d+\)', value)
        if not match:
            raise ValueError('invalid load-command path')
        value = match[1]
    return value


def system_dependency(name):
    parts = PurePosixPath(name).parts
    return (name.startswith(('/System/Library/', '/usr/lib/'))
            and not any(p in ('', '.', '..') for p in name.split('/')[1:])
            and '\\' not in name and not any(ord(c)<32 for c in name)
            and '..' not in parts)


def parse_commands(output, host):
    dependencies = []; deployment = []; dyld = []
    blocks = re.split(r'(?m)^Load command \d+\s*$', output)
    if len(blocks) < 2:
        raise ValueError('missing Mach-O load commands')
    for block in blocks[1:]:
        cmd = field(block, 'cmd')
        if cmd in LOADS:
            name = field(block, 'name', path=True)
            if not system_dependency(name):
                raise ValueError('unbundled or non-system dependency: '+name)
            dependencies.append(dict(kind=cmd, path=name))
        elif cmd == 'LC_RPATH':
            raise ValueError('unexpected rpath in system-library-only private route: '+field(block, 'path', path=True))
        elif cmd == 'LC_DYLD_ENVIRONMENT':
            raise ValueError('embedded dyld environment override')
        elif cmd == 'LC_LOAD_DYLINKER':
            dyld.append(field(block, 'name', path=True))
        elif cmd in {'LC_BUILD_VERSION', 'LC_VERSION_MIN_MACOSX'}:
            if cmd == 'LC_BUILD_VERSION' and field(block, 'platform') not in {'1', 'macos', 'MACOS'}:
                raise ValueError('Mach-O deployment platform is not macOS')
            minimum = field(block, 'minos' if cmd=='LC_BUILD_VERSION' else 'version')
            sdk = field(block, 'sdk')
            if version(sdk) < version(minimum):
                raise ValueError('SDK version precedes deployment minimum')
            deployment.append(dict(command=cmd, minimum=minimum, sdk=sdk))
    if not dependencies or len(deployment)!=1 or dyld!=['/usr/lib/dyld']:
        raise ValueError('missing/ambiguous dependency, deployment or dynamic-loader metadata')
    if version(deployment[0]['minimum']) > version(host):
        raise ValueError('deployment minimum exceeds this inspection host')
    return dict(dependencies=dependencies, deployment=deployment[0], rpaths=[], dynamic_loader=dyld[0])


def tool(*args):
    return check_child_output(list(args), env=child_environment(), text=True, timeout=20)


def inspect_app(app,name,identifier,build,host):
    v.no_links(app)
    executable=app/'Contents/MacOS'/name;v.no_links(executable)
    # Inspect only the reviewed route: exactly one Mach-O main executable.
    binaries=[];count=0
    for path in app.rglob('*'):
        count+=1
        if count>2000 or path.is_symlink():
            raise ValueError('app file count or symlink outside admitted route')
        if not path.is_file() and not path.is_dir():
            raise ValueError('special file outside admitted app route')
        if path.is_file():
            v.installer.regular(path, v.installer.MAX_ARCHIVE)
            with path.open('rb') as stream:
                if stream.read(4) in MACH:binaries.append(path)
    if binaries!=[executable]:
        raise ValueError('unexpected embedded Mach-O closure')
    info=plistlib.loads(v.read_regular(app/'Contents/Info.plist'))
    v.require_fields(info,dict(CFBundleExecutable=name,CFBundleIdentifier=identifier,CFBundleVersion=build),'app plist')
    if tool('lipo','-archs',str(executable)).split()!=['arm64']:
        raise ValueError('expected thin ARM64 executable')
    header=tool('otool','-arch','arm64','-hv',str(executable))
    if not re.search(r'^MH_MAGIC_64\s+ARM64\s+\S+\s+\S+\s+EXECUTE\s',header,re.M):
        raise ValueError('expected ARM64 Mach-O executable header')
    observed=parse_commands(tool('otool','-arch','arm64','-l',str(executable)),host)
    return dict(name=name,executable_sha256=v.installer.digest(executable),architectures=['arm64'],**observed)


def inspect(installation, receipt_sha256):
    if sys.platform!='darwin':
        raise ValueError('static Mach-O inspection requires macOS tools')
    installation=installation.expanduser().absolute();v.no_links(installation)
    receipt, receipt_hash=v.record(installation/'INSTALLATION.json')
    if not re.fullmatch('[0-9a-f]{64}', receipt_sha256) or receipt_hash!=receipt_sha256:
        raise ValueError('installation receipt pin mismatch')
    names=receipt.get('apps')
    if (not isinstance(names,list) or len(names)!=len(v.IDENTIFIERS)
            or any(not isinstance(n,str) for n in names) or set(names)!=set(v.IDENTIFIERS)
            or not isinstance(receipt.get('source_commit'),str) or not re.fullmatch('[0-9a-f]{40}',receipt['source_commit'])
            or not isinstance(receipt.get('build_version'),str) or not re.fullmatch('[0-9]{1,6}',receipt['build_version'])):
        raise ValueError('unexpected private installation identity')
    host=platform.mac_ver()[0];version(host)
    apps=[]
    v.no_links(installation/'apps')
    for name in names:
        apps.append(inspect_app(installation/'apps'/(name+'.app'),name,v.IDENTIFIERS[name],receipt['build_version'],host))
    return dict(source_commit=receipt['source_commit'],build_version=receipt['build_version'],
                installation_sha256=receipt_hash,host_version=host,host_meets_declared_minima=True,
                apps=apps,native_runtime='not executed',qualification='static metadata only; no support-matrix extension',
                scope='current private five-app system-library-only route; dlopen/plugins/transitive system closure not proven')


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--installation',required=True,type=Path)
    parser.add_argument('--installation-sha256',required=True)
    args=parser.parse_args()
    try:
        report=inspect(args.installation,args.installation_sha256)
    except (ValueError,OSError,subprocess.CalledProcessError) as error:
        parser.exit(1,str(error)+'\n')
    print(json.dumps(report,indent=2))


if __name__=='__main__':main()
