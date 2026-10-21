from pathlib import Path
import unittest,importlib.util,sys
from check import validate,base
s=importlib.util.spec_from_file_location('parent_fixtures',Path(__file__).resolve().parents[1]/'check_test.py');f=importlib.util.module_from_spec(s)
ours=sys.modules['check'];sys.modules['check']=base;s.loader.exec_module(f);sys.modules['check']=ours
class Backend(unittest.TestCase):
 def data(self,mode='on'):
  d,m=f.Controls().data('framework');d['observer_mode']=mode;rows=[]
  if mode=='on':
   for j,frame in enumerate(d['before_close']['Samples']):
    frame['DrawableID']=j+6;d['after_close']['Samples'][j]['DrawableID']=j+6
    a=frame['Acquired']
    rows.append(dict(ordinal=len(rows),kind=0,queue=10,drawable_id=0,drawable_recorded=False,commit_start=a+.001,commit_end=a+.0011,present_request=0,scheduled_observed=a+.002,complete_observed=a+.004,gpu_start=a+.002,gpu_end=a+.003,shown=0,status=4))
    rows.append(dict(ordinal=len(rows),kind=2,queue=10,drawable_id=frame['DrawableID'],drawable_recorded=True,commit_start=a+.004,commit_end=a+.0041,present_request=a+.0035,scheduled_observed=a+.005,complete_observed=a+.006,gpu_start=0,gpu_end=0,shown=frame['Presented'],status=4))
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
  d,m=self.data();d['backend_trace']['records'][-1]['drawable_id']=d['backend_trace']['records'][-3]['drawable_id']
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
 def test_sparse_tags_use_measurement_membership(self):
  d,m=self.data()
  for folder in ['before_close','after_close']:
   for r in d[folder]['Samples']:
    if r['Armed']:r['Tag']=2*r['Tag']+2
   for r in d[folder]['Inputs']:r['Tag']=2*r['Tag']+2
  for r in d['measurements']['Samples']:r['Frame']['Tag']=2*r['Frame']['Tag']+2
  self.assertEqual(validate(d,m,'on')['phase_ms']['receipt_to_submit']['count'],120)
 def with_prefix(self):
  d,m=self.data();rows=d['backend_trace']['records'];prefix=[]
  for i in range(6):
   a=.01+i*.01
   prefix.append(dict(ordinal=i,kind=2,queue=10,drawable_id=i,drawable_recorded=True,commit_start=a,commit_end=a+.0001,present_request=a-.001,scheduled_observed=a+.001,complete_observed=a+.002,gpu_start=0,gpu_end=0,shown=0 if i<2 else a+.003,status=4))
  for r in rows:r['ordinal']+=6
  d['backend_trace']['records']=prefix+rows;d['backend_trace']['count']+=6;return d,m
 def test_accounted_pre_admission_prefix(self):
  result=validate(*self.with_prefix(),'on');self.assertEqual(result['retained_pre_admission_prefix'],dict(count=6,zero_feedback=2,positive_feedback=4,drawable_ids=list(range(6))))
 def test_unmatched_after_admission_refused(self):
  d,m=self.with_prefix();d['backend_trace']['records'][0]['drawable_id']=999
  with self.assertRaises(ValueError):validate(d,m,'on')
 def test_late_startup_prefix_refused(self):
  d,m=self.with_prefix();d['backend_trace']['records'][0]['present_request']=10
  with self.assertRaises(ValueError):validate(d,m,'on')
 def test_unrecorded_zero_id_refused(self):
  d,m=self.with_prefix();d['backend_trace']['records'][0]['drawable_recorded']=False
  with self.assertRaises(ValueError):validate(d,m,'on')
 def test_armed_setup_without_measured_receipt(self):
  d,m=self.data()
  for folder in ['before_close','after_close']:
   startup=d[folder]['Samples'][1];startup.update(Tag=9999,Generation=9,Armed=True)
  self.assertEqual(validate(d,m,'on')['phase_ms']['receipt_to_submit']['count'],120)
if __name__=='__main__':unittest.main()
