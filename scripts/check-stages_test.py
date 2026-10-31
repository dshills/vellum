"""Exercise stage dispatch without launching compilers, native windows or network."""
from pathlib import Path
import os
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
STAGES = ['tooling','portable','native','dependencies','examples','offscreen']
class Stages(unittest.TestCase):
 def invoke(self, script, *args, env=None):
  return subprocess.run(['sh', str(script), *args],env=env,text=True,capture_output=True,timeout=15)
 def test_listing_and_invalid_stage_never_initialize_dependencies(self):
  with tempfile.TemporaryDirectory() as directory:
   script=Path(directory)/'check.sh';shutil.copy2(ROOT/'scripts/check.sh',script)
   result=self.invoke(script,'--list-stages');self.assertEqual(result.returncode,0,result.stderr);self.assertEqual(result.stdout.splitlines(),STAGES)
   for args in [('--stage','all'),('--stage','unknown'),('--stage',),('--list-stages','extra')]:
    result=self.invoke(script,*args);self.assertEqual(result.returncode,2);self.assertIn('Usage:',result.stderr);self.assertNotIn('env.sh',result.stderr)
 def test_all_dispatch_equals_ordered_stages_without_losing_commands(self):
  with tempfile.TemporaryDirectory() as directory:
   root=Path(directory);(root/'scripts').mkdir();shutil.copy2(ROOT/'scripts/check.sh',root/'scripts/check.sh')
   for name in ['upstream','typesetting']:
    target=root/'drivers/cogent/build'/name;target.mkdir(parents=True);(target/'go.mod').touch()
   (root/'drivers/cogent/env.sh').write_text(':\n')
   (root/'scripts/check-example.sh').write_text('#!/bin/sh\necho example-suite >> "$VELLUM_STAGE_LOG"\n');(root/'scripts/check-example.sh').chmod(0o755)
   for name in ['sdk-contract','replay','grid','loadedgrid','loadedtree','placement','demo-suite','personal-mac','forge']:(root/'examples'/name).mkdir(parents=True)
   tools=root/'bin';tools.mkdir()
   for name in ['go','python3']:
    executable=tools/name;executable.write_text('#!/bin/sh\n'+('if [ "$1 $2" = "env GOOS" ]; then echo darwin; exit 0; fi\n' if name=='go' else '')+'echo "'+name+' $*" >> "$VELLUM_STAGE_LOG"\n');executable.chmod(0o755)
   log=root/'calls';env=dict(os.environ,PATH=str(tools)+os.pathsep+os.environ['PATH'],VELLUM_STAGE_LOG=str(log))
   result=self.invoke(root/'scripts/check.sh',env=env);self.assertEqual(result.returncode,0,result.stderr);complete=log.read_text().splitlines();log.unlink()
   for stage in STAGES:
    result=self.invoke(root/'scripts/check.sh','--stage',stage,env=env);self.assertEqual(result.returncode,0,result.stderr)
   self.assertEqual(complete,log.read_text().splitlines())
   self.assertIn('go test -race ./...',complete);self.assertIn('go vet ./...',complete);self.assertIn('example-suite',complete)
   self.assertIn('go test -race -tags offscreen -timeout=20m ./...',complete)
   self.assertIn('python3 -B bundle_tools_test.py',complete)
   self.assertTrue(any('sceneplan' in line and 'gpuscene' in line for line in complete))
   self.assertTrue(any('TestMainQueuePublicationDuringStartupDispatch' in line for line in complete))
if __name__=='__main__':unittest.main()
