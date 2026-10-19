"""Compile/run actual bridge ownership with inert objects; never open a window."""
import argparse, hashlib, json, sys, tempfile
from pathlib import Path

STAGE=Path(__file__).resolve().parent
DRIVER=STAGE.parents[2]
ROOT=DRIVER.parents[1]
sys.path.insert(0,str(ROOT/'scripts'))
from child_process import run_child
from sdk_environment import child_environment

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--layer-header',type=Path,default=DRIVER/'build/upstream/system/driver/desktop/surface_layer_darwin.h')
    parser.add_argument('--input-source',type=Path,default=DRIVER/'internal/textinput/input_darwin.m')
    parser.add_argument('--accessibility-source',type=Path,default=DRIVER/'accessibility_darwin.m')
    args=parser.parse_args()
    sources={'surface_pool':args.layer_header,'input_pool':args.input_source,'accessibility_pool':args.accessibility_source}
    for source in sources.values():
        if not source.is_absolute() or not source.is_file():parser.error('Fixture sources require existing absolute paths')
    env=child_environment();result=0
    with tempfile.TemporaryDirectory(prefix='vellum-pool-fixture-')as temporary:
        for name,source in sources.items():
            fixture=STAGE/(name+'_test.m');binary=Path(temporary)/name
            selector={'surface_pool':'VELLUM_LAYER_SOURCE','input_pool':'VELLUM_INPUT_SOURCE','accessibility_pool':'VELLUM_AX_SOURCE'}[name]
            implementation=source.parent/"presentation_darwin.m" if name=="surface_pool" else None
            extra=[implementation,implementation.with_suffix(".h")] if implementation else []
            inputs={str(f):hashlib.sha256(f.read_bytes()).hexdigest()for f in [fixture,source,DRIVER/'internal/textinput/input.h',*extra]}
            argv=['/usr/bin/clang','-fobjc-arc'if name=='accessibility_pool'else'-fno-objc-arc','-Wno-unused-parameter','-I',str(DRIVER/'internal/textinput'),'-D'+selector+'="'+str(source)+'"',str(fixture),'-framework','Cocoa','-framework','QuartzCore','-framework','Carbon','-o',str(binary)]
            if implementation:argv[1:1]=["-fblocks",str(implementation),"-framework","Metal"]
            compiled=run_child(argv,env=env,capture_output=True,timeout=30)
            record={'fixture':name,'live_tests':False,'inputs':inputs,'compile_argv':argv,'compile_exit':compiled.returncode,'compile_stdout':compiled.stdout.decode(),'compile_stderr':compiled.stderr.decode(),'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest()if binary.exists()else None}
            if compiled.returncode==0:
                ran=run_child([str(binary)],env=env,capture_output=True,timeout=10)
                record.update({'run_exit':ran.returncode,'run_stdout':ran.stdout.decode(),'run_stderr':ran.stderr.decode()})
                result=max(result,int(ran.returncode!=0))
            else:result=1
            print(json.dumps(record),flush=True)
    return result

if __name__=='__main__':sys.exit(main())
