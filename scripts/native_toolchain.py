"""Observe selected Apple compiler/linker/SDK inputs without compiling or changing settings."""
import hashlib
from pathlib import Path
import re
import shlex
import shutil

from child_process import run_child, check_child_output


def identity(path):
    path=Path(path).resolve(strict=True)
    if not path.is_file():raise ValueError('expected native tool/SDK settings file')
    h=hashlib.sha256()
    with path.open('rb') as stream:
        while chunk:=stream.read(1<<20):h.update(chunk)
    return dict(path=str(path),sha256=h.hexdigest())


def capture(config,env):
    def output(*args):
        return check_child_output(list(args),env=env,text=True,timeout=20).strip()
    sdk=Path(output('xcrun','--sdk','macosx','--show-sdk-path')).resolve(strict=True)
    settings=identity(sdk/'SDKSettings.json')
    binaries={};versions={};seen={}
    for key in ('CC','CXX'):
        command=shlex.split(config[key])
        if len(command)!=1:raise ValueError('native provenance requires a single compiler driver')
        selected=shutil.which(command[0],path=env.get('PATH',''))
        if not selected:raise ValueError('native compiler driver unavailable')
        actual=output('xcrun','--sdk','macosx','--find',Path(command[0]).name)
        for label,path in [(key.lower()+'_driver',selected),(key.lower(),actual)]:
            canonical=str(Path(path).resolve(strict=True))
            if canonical not in seen:seen[canonical]=identity(path)
            binaries[label]=seen[canonical]
        versions[key]=output(selected,'--version')
    linker=output('xcrun','--sdk','macosx','--find','ld')
    binaries['linker']=identity(linker)
    versions['linker']=output(linker,'-version_details')
    compiler=binaries['cc_driver']['path']
    # -### prints the driver's commands and does not invoke compiler/linker or
    # write an output. Include the effective Go cgo flags, not ambient overrides.
    probes={}
    for name,flags in [('compile',['-c']),('link',[])]:
        if name=='compile':
            args=[compiler,'-###','-arch','arm64',*shlex.split(config['CGO_CPPFLAGS']),
                  *shlex.split(config['CGO_CFLAGS']),'-x','c','/dev/null',*flags,'-o','/dev/null']
        else:
            # Treat /dev/null as linker input: compiling it in a link probe would
            # invent a random temporary object name and make snapshots unstable.
            args=[compiler,'-###','-arch','arm64','-x','none','/dev/null','-o','/dev/null',
                  *shlex.split(config['CGO_LDFLAGS'])]
        result=run_child(args,env=env,capture_output=True,text=True,check=True,timeout=20)
        text=result.stdout+result.stderr
        if len(text)>65536:raise ValueError('native compiler probe output exceeds bound')
        probes[name]=text
    triple=re.findall(r'arm64-apple-macosx([0-9]+(?:\.[0-9]+){0,2})',probes['compile'])
    if not triple:raise ValueError('native compiler did not report macOS ARM64 deployment triple')
    return dict(scope='selected drivers/tools and SDK settings; not full SDK/toolchain closure or authenticity',
        binaries=binaries,versions=versions,sdk=dict(path=str(sdk),version=output('xcrun','--sdk','macosx','--show-sdk-version'),settings=settings),
        cgo={k:config[k] for k in ('CC','CXX','CGO_CFLAGS','CGO_CPPFLAGS','CGO_CXXFLAGS','CGO_LDFLAGS')},
        overrides={k:env.get(k) for k in ('SDKROOT','MACOSX_DEPLOYMENT_TARGET','DEVELOPER_DIR')},
        compiler_default_minimum=triple[-1],driver_probes=probes)
