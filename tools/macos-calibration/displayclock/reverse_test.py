from pathlib import Path
import tempfile,unittest
from reverse_control import execute
class Reversed(unittest.TestCase):
 def run_case(self,fail=None):
  temp=tempfile.TemporaryDirectory();self.addCleanup(temp.cleanup);root=Path(temp.name);calls=[];cases=[];saved=[]
  class Host:
   def __init__(self,out):self.mode=out.name;self.probe_count=0;self.geo_count=0
   def geometry(self):
    self.geo_count+=1
    return 'changed' if fail=='geometry' and self.mode=='on' and self.geo_count==2 else 'stable'
   def probes(self,out):
    self.probe_count+=1
    value='changed' if (fail=='local-config' and self.mode=='on' and self.probe_count==2) or (fail=='cross-config' and self.mode=='off') else 'stable'
    return {'display':{'logs':{'stdout':{'sha256':value}}}}
   def command(self,argv,cwd,out,timeout):
    calls.append((argv,cwd,out,timeout));return {'ok':not(fail=='native' and self.mode=='on'),'logs':{'stderr':{'bytes':1 if fail=='stderr' and self.mode=='on' else 0}}}
   def finish(self):pass
   def stopping(self):return fail=='stopping' and self.mode=='on'
  def validate(path,mode):
   if fail=='validator' and mode=='on':raise ValueError('invalid observation')
   return {'mode':mode,'accepted':True}
  def run():execute(root,Path('BINARY'),Path('CWD'),Path('MATCH'),Host,validate,lambda p,d:saved.append((p,d)),lambda r:r['ok'],cases)
  return run,calls,cases,saved
 def test_literal_order_arguments_and_retention(self):
  run,calls,cases,saved=self.run_case();run()
  self.assertEqual([c[0] for c in calls],[['BINARY','timed','MATCH','on','on'],['BINARY','timed','MATCH','on','off']]);self.assertEqual([c['mode'] for c in cases],['on','off']);self.assertEqual(len(saved),2);self.assertTrue(all(c[3]==40 for c in calls))
 def test_first_native_or_stderr_failure_stops(self):
  for failure in ['native','stderr']:
   run,calls,cases,saved=self.run_case(failure)
   with self.subTest(failure=failure),self.assertRaises(ValueError):run()
   self.assertEqual(len(calls),1);self.assertEqual(len(cases),1);self.assertFalse(saved)
 def test_validation_and_stop_failures_do_not_run_off(self):
  for failure in ['validator','stopping']:
   run,calls,cases,saved=self.run_case(failure)
   with self.subTest(failure=failure),self.assertRaises(ValueError):run()
   self.assertEqual(len(calls),1);self.assertEqual(len(cases),1)
 def test_local_drift_stops_before_off(self):
  for failure in ['geometry','local-config']:
   run,calls,cases,saved=self.run_case(failure)
   with self.subTest(failure=failure),self.assertRaises(ValueError):run()
   self.assertEqual(len(calls),1);self.assertFalse(saved)
 def test_cross_case_drift_retains_both_and_refuses(self):
  run,calls,cases,saved=self.run_case('cross-config')
  with self.assertRaisesRegex(ValueError,'cross-case'):run()
  self.assertEqual(len(calls),2);self.assertEqual(len(cases),2);self.assertEqual(len(saved),2)
if __name__=='__main__':unittest.main()
