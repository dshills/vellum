import tempfile,unittest
from pathlib import Path
import control
class Host:
 def __init__(self,mode):self.mode=mode;self.commands=[];self.finished=False;self.geometries=0
 def geometry(self):self.geometries+=1;return {'fixed':self.geometries if self.mode=='geometry' else 1}
 def probes(self,path):return {'display':{'logs':{'stdout':{'sha256':'drift' if self.mode=='config' and path.name=='after' else 'fixed'}}}}
 def command(self,argv,cwd,out,timeout):self.commands.append((argv,cwd,out,timeout));return {'ok':self.mode!='command','logs':{'stderr':{'bytes':int(self.mode=='stderr')}}}
 def finish(self):self.finished=True
 def stopping(self):return self.mode=='deadline'
class Routing(unittest.TestCase):
 def execute(self,mode):
  self.host=Host(mode);self.result={};self.saved=[];self.validated=[]
  def validate(path,match):
   self.validated.append((path.name,match.name))
   if mode=='invalid':raise ValueError('strict phase rejection')
   return {'framework':True}
  with tempfile.TemporaryDirectory() as name:
   root=Path(name);self.root=root;match=root/'match.json'
   if mode!='missing':match.write_text('{}')
   return control.execute(root,Path('/unchanged/framework'),match,self.host,validate,lambda path,value:self.saved.append((path.name,value)),lambda run:run['ok'],self.result)
 def test_one_exact_timed_command_with_forty_second_watchdog(self):
  self.assertEqual(self.execute('valid'),{'framework':True});argv,cwd,out,timeout=self.host.commands[0];self.assertEqual(argv,['/unchanged/framework','timed',str(self.root/'match.json'),'on','on']);self.assertEqual(timeout,40);self.assertEqual(len(self.host.commands),1);self.assertTrue(self.host.finished);self.assertEqual(self.validated,[('stdout','match.json')]);self.assertEqual(self.saved,[('validation.json',{'framework':True})])
 def test_refusals_preserve_partial_result_and_do_not_retry(self):
  for mode in ['command','stderr','geometry','config','deadline','invalid','missing']:
   with self.subTest(mode=mode),self.assertRaises(ValueError):self.execute(mode)
   self.assertEqual(len(self.host.commands),0 if mode=='missing' else 1);self.assertNotIn('validation',self.result);self.assertFalse(self.saved)
if __name__=='__main__':unittest.main()
