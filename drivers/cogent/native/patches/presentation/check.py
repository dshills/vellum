"""Compile/run controlled drawable callbacks; no window or GPU device is used."""
import argparse,hashlib,json,sys,tempfile
from pathlib import Path
STAGE=Path(__file__).resolve().parent
DRIVER=STAGE.parents[2]
sys.path.insert(0,str(DRIVER.parents[1]/'scripts'))
from child_process import run_child
from sdk_environment import child_environment

def main():
 parser=argparse.ArgumentParser(description=__doc__)
 parser.add_argument('--source',type=Path,default=DRIVER/'build/upstream/system/driver/desktop/presentation_darwin.m')
 args=parser.parse_args();source=args.source
 if not source.is_absolute()or not source.is_file():parser.error('Existing absolute prepared source required')
 fixture=STAGE/'presentation_fixture.m';header=source.with_suffix('.h')
 pins={str(f):hashlib.sha256(f.read_bytes()).hexdigest()for f in [fixture,source,header]}
 with tempfile.TemporaryDirectory(prefix='vellum-presentation-fixture-')as temporary:
  binary=Path(temporary)/'presentation'
  argv=['/usr/bin/clang','-fno-objc-arc','-fblocks','-Wall','-Wextra','-Werror','-DVELLUM_PRESENTATION_SOURCE="'+str(source)+'"',str(fixture),'-framework','Foundation','-framework','QuartzCore','-framework','Metal','-o',str(binary)]
  compiled=run_child(argv,env=child_environment(),capture_output=True,timeout=30)
  result={'source_pins':pins,'compile_exit':compiled.returncode,'compile_stderr':compiled.stderr.decode(),'native_window':False,'real_gpu_drawables':0,'actual_presentation':False}
  if compiled.returncode==0:
   ran=run_child([str(binary)],env=child_environment(),capture_output=True,timeout=10)
   result.update(run_exit=ran.returncode,run_stdout=ran.stdout.decode(),run_stderr=ran.stderr.decode())
  print(json.dumps(result))
  return int(compiled.returncode!=0 or result.get('run_exit')!=0)
if __name__=='__main__':sys.exit(main())
