import tempfile,unittest
from pathlib import Path
import control
class Host:
 def __init__(self,mode):self.mode=mode;self.commands=[];self.finished=False;self.geometries=0
 def geometry(self):self.geometries+=1;return {'fixed':self.geometries if self.mode=='geometry' else 1}
 def probes(self,path):return {'display':{'logs':{'stdout':{'sha256':'drift' if self.mode=='config' and path.name=='after' else 'fixed'}}}}
 def command(self,argv,cwd,out,timeout):self.commands.append((argv,cwd,out,timeout));return {'ok':self.mode!='command','logs':{'stderr':{'bytes':0}}}
 def finish(self):self.finished=True
 def stopping(self):return self.mode=='deadline'
class Routing(unittest.TestCase):
 def execute(self,mode):
  self.host=Host(mode);self.result={};self.saved=[];self.validated=[]
  def validate(path,match):
   self.validated.append((path.name,match.name))
   if mode=='invalid':raise ValueError('startup invalid')
   return {'startup':True}
  with tempfile.TemporaryDirectory() as name:
   root=Path(name);self.root=root
   if mode=='exists':(root/'pilot-match.json').write_text('{}')
   return control.execute(root,Path('/unchanged/framework'),self.host,validate,lambda path,value:self.saved.append((path.name,value)),lambda run:run['ok'],self.result)
 def test_single_exact_pilot_no_native_reference_or_cohort(self):
  self.assertEqual(self.execute('valid'),{'startup':True});argv,cwd,out,timeout=self.host.commands[0];self.assertEqual(argv,['/unchanged/framework','pilot',str(self.root/'pilot-match.json'),'on','on']);self.assertEqual(timeout,20);self.assertEqual(len(self.host.commands),1);self.assertTrue(self.host.finished);self.assertEqual(self.validated,[('stdout','pilot-match.json')]);self.assertEqual(self.saved,[('validation.json',{'startup':True})])
 def test_refusals_keep_partial_result_and_do_not_retry(self):
  for mode in ['command','geometry','config','deadline','invalid','exists']:
   with self.subTest(mode=mode),self.assertRaises(ValueError):self.execute(mode)
   self.assertEqual(len(self.host.commands),0 if mode=='exists' else 1);self.assertNotIn('validation',self.result);self.assertFalse(self.saved)
if __name__=='__main__':unittest.main()
