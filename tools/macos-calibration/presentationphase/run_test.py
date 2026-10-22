import tempfile,unittest
from pathlib import Path
import control_run
class Host:
 def __init__(self,root,log,fail):self.root=root;self.log=log;self.fail=fail
 def geometry(self):return {'same':True}
 def probes(self,root):return {'display':{'logs':{'stdout':{'sha256':'fixed'}}}}
 def command(self,args,cwd,out,timeout):self.log.append((args,timeout));return {'ok':not self.fail,'logs':{'stderr':{'bytes':0}}}
 def finish(self):pass
 def stopping(self):return False
class Routing(unittest.TestCase):
 def run_case(self,fail=False,reject=False):
  log=[];cases=[];saved=[];self.log=log;self.cases=cases;commands={'native':['N','MATCH'],'framework':['F','timed','MATCH','on','on']}
  def validate(path,kind):
   if reject:raise ValueError('invalid first journal')
   return {'kind':kind}
  with tempfile.TemporaryDirectory() as directory:
   (Path(directory)/'native').write_text('binary');(Path(directory)/'framework').write_text('binary')
   result=control_run.execute(Path(directory),commands,lambda root:Host(root,log,fail),validate,lambda a,b: {'pair':[a['kind'],b['kind']]},lambda path,value:saved.append(path.name),lambda run:run['ok'],cases)
  return result,log,cases,saved
 def test_exact_commands_order_timeout_and_pair(self):
  result,log,cases,saved=self.run_case();self.assertEqual(log,[(['N','MATCH'],40),(['F','timed','MATCH','on','on'],40)]);self.assertEqual(result,{'pair':['native','framework']});self.assertEqual(saved,['validation.json','validation.json','pair-validation.json'])
 def test_stop_first_invalid_without_second_native(self):
  for failure in ['native','validation']:
   with self.subTest(failure=failure):
    with self.assertRaises(ValueError):self.run_case(fail=failure=='native',reject=failure=='validation')
    self.assertEqual(len(self.log),1);self.assertEqual(len(self.cases),1);self.assertIn('native',self.cases[0]);self.assertNotIn('validation',self.cases[0])
if __name__=='__main__':unittest.main()
