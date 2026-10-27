#!/usr/bin/env python3
"""Build one private ad-hoc Forge app/ZIP in a fresh directory; never publishes."""
import argparse,hashlib,json,platform,sys
from pathlib import Path
sys.dont_write_bytecode=True
MODULE=Path(__file__).resolve().parent
ROOT=MODULE.parents[1]
sys.path.insert(0,str(ROOT/'scripts'))
from child_process import run_child,check_child_output
from sdk_environment import child_environment

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--output',type=Path,required=True)
    p.add_argument('--build-version',default='1')
    a=p.parse_args()
    if platform.system()!='Darwin':p.error('macOS is required')
    out=a.output.absolute()
    env=child_environment()
    def run(argv):run_child(argv,cwd=MODULE,env=env,timeout=600,check=True)
    revision=check_child_output(['git','rev-parse','HEAD'],cwd=ROOT,env=env,text=True,timeout=10).strip()
    dirty=check_child_output(['git','status','--porcelain','--untracked-files=normal'],cwd=ROOT,env=env,text=True,timeout=10)
    if dirty:raise ValueError('commit source changes before packaging Forge')
    out.mkdir()  # reserve only after source validation; never overwrite
    binary=out/'forge';icon=out/'Forge.png';bundle=out/'Vellum Forge.app'
    run(['go','build','-trimpath','-o',str(binary),'.'])
    run([str(binary),'--write-icon',str(icon)])
    run([sys.executable,'-B',str(ROOT/'scripts/package-app.py'),'--binary',str(binary),'--output',str(bundle),'--name','Vellum Forge','--identifier','com.dshills.vellum.forge.demo','--version','0.1.0','--build-version',a.build_version,'--icon',str(icon),'--sign-identity','-'])
    run(['codesign','--verify','--deep','--strict',str(bundle)])
    archive=out/('Vellum-Forge-macOS-'+platform.machine()+'.zip')
    run(['ditto','-c','-k','--norsrc','--noextattr','--keepParent',str(bundle),str(archive)])
    extract=out/'extracted';extract.mkdir();run(['ditto','-x','-k',str(archive),str(extract)])
    run(['codesign','--verify','--deep','--strict',str(extract/'Vellum Forge.app')])
    if check_child_output(['git','rev-parse','HEAD'],cwd=ROOT,env=env,text=True,timeout=10).strip()!=revision or check_child_output(['git','status','--porcelain','--untracked-files=normal'],cwd=ROOT,env=env,text=True,timeout=10):
        raise ValueError('source changed during packaging; partial output is not a completed delivery')
    digest=hashlib.sha256(archive.read_bytes()).hexdigest()
    (out/'SHA256SUMS').write_text(digest+'  '+archive.name+'\n')
    (out/'README.md').write_bytes((MODULE/'README.md').read_bytes())
    (out/'manifest.json').write_text(json.dumps({'source_commit':revision,'private':True,'notarized':False,'signing':'ad-hoc','architecture':platform.machine(),'archive':archive.name,'sha256':digest,'extracted_signature_verified':True},indent=2)+'\n')
    print(str(out))
if __name__=='__main__':main()
