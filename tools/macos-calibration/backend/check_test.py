from pathlib import Path
import unittest,importlib.util,sys
from check import validate,base
s=importlib.util.spec_from_file_location('parent_fixtures',Path(__file__).resolve().parents[1]/'check_test.py');f=importlib.util.module_from_spec(s)
ours=sys.modules['check'];sys.modules['check']=base;s.loader.exec_module(f);sys.modules['check']=ours
class Backend(unittest.TestCase):
 def data(self,mode='on'):
  d,m=f.Controls().data('framework');d['observer_mode']=mode;rows=[]
  if mode=='on':
   for frame in d['before_close']['Samples']:
    a=frame['Acquired']
    rows.append(dict(ordinal=len(rows),kind=0,queue=10,commit_start=a+.001,commit_end=a+.0011,present_request=0,scheduled_observed=a+.002,complete_observed=a+.004,gpu_start=a+.002,gpu_end=a+.003,shown=0,status=4))
    rows.append(dict(ordinal=len(rows),kind=2,queue=10,commit_start=a+.004,commit_end=a+.0041,present_request=a+.0035,scheduled_observed=a+.005,complete_observed=a+.006,gpu_start=0,gpu_end=0,shown=frame['Presented'],status=4))
  d['backend_trace']=dict(enabled=mode=='on',pending=0,overflow=False,count=len(rows),records=rows);return d,m
 def test_on(self):self.assertEqual(validate(*self.data(),'on')['phase_ms']['gpu_end_to_shown']['count'],120)
 def test_off(self):self.assertEqual(validate(*self.data('off'),'off')['observer_mode'],'off')
 def test_record_failures(self):
  for key,value in [('gpu_start',0),('gpu_end',float('nan')),('gpu_end',100),('status',5),('ordinal',0),('queue',99),('commit_end',0),('scheduled_observed',0)]:
   with self.subTest(key=key):
    d,m=self.data();d['backend_trace']['records'][-2][key]=value
    with self.assertRaises(ValueError):validate(d,m,'on')
 def test_drain_and_overflow(self):
  for key,value in [('pending',1),('overflow',True),('count',4097),('enabled',False)]:
   with self.subTest(key=key):
    d,m=self.data();d['backend_trace'][key]=value
    with self.assertRaises(ValueError):validate(d,m,'on')
 def test_ambiguous_feedback(self):
  d,m=self.data();d['backend_trace']['records'][-1]['shown']=d['backend_trace']['records'][-3]['shown']
  with self.assertRaises(ValueError):validate(d,m,'on')
 def test_missing_feedback(self):
  d,m=self.data();d['backend_trace']['records'][-1]['shown']=0
  with self.assertRaises(ValueError):validate(d,m,'on')
 def test_reordered_commit(self):
  d,m=self.data();d['backend_trace']['records'][-1]['commit_start']=1
  with self.assertRaises(ValueError):validate(d,m,'on')
 def test_disabled_rows(self):
  d,m=self.data();d['observer_mode']='off';d['backend_trace']['enabled']=False
  with self.assertRaises(ValueError):validate(d,m,'off')
 def test_parent_focus_failure(self):
  d,m=self.data();d['end']['losses']=1
  with self.assertRaises(ValueError):validate(d,m,'on')
if __name__=='__main__':unittest.main()
