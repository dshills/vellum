import unittest,importlib.util
from pathlib import Path
from check import validate,CLOCKS
s=importlib.util.spec_from_file_location('fixtures',Path(__file__).resolve().parents[1]/'check_test.py');fixtures=importlib.util.module_from_spec(s)
# Parent fixture imports its own classifier, not this extension.
import sys
ours=sys.modules['check'];sys.modules['check']=__import__('check').base
s.loader.exec_module(fixtures);sys.modules['check']=ours
class Submission(unittest.TestCase):
 def data(self,mode='combined'):
  d,m=fixtures.Controls().data();d['mode']=mode
  for r in d['records']:
   a=r['acquired'];t={k:a+.003 for k in CLOCKS}
   t.update(render_commit_start=a+.001,render_commit_end=a+.002,present_request=a+.0005,present_commit_start=a+.001,present_commit_end=a+.002,gpu_start=a+.004,gpu_end=a+.005,render_status=4,present_status=4)
   if mode=='split':t.update(present_request=a+.0021,present_commit_start=a+.0022,present_commit_end=a+.0023)
   r['timing']=t
  return d,m
 def test_combined(self):self.assertEqual(validate(*self.data(),'combined')['phase_ms']['gpu_duration']['count'],120)
 def test_split(self):self.assertEqual(validate(*self.data('split'),'split')['mode'],'split')
 def test_bad_timings(self):
  for key,value in [('gpu_start',0),('gpu_end',float('nan')),('gpu_end',100),('gpu_end',1),('render_status',5),('present_status',0),('render_commit_end',0),('present_request',100),('render_complete_observed',0)]:
   with self.subTest(key=key,value=value):
    d,m=self.data();d['records'][-1]['timing'][key]=value
    with self.assertRaises(ValueError):validate(d,m,'combined')
 def test_missing_clock(self):
  d,m=self.data();del d['records'][-1]['timing']['gpu_end']
  with self.assertRaises(KeyError):validate(d,m,'combined')
 def test_split_order(self):
  d,m=self.data('split');d['records'][-1]['timing']['present_request']=1
  with self.assertRaises(ValueError):validate(d,m,'split')
 def test_mode(self):
  d,m=self.data()
  with self.assertRaises(ValueError):validate(d,m,'split')
 def test_parent_rejects_loss(self):
  d,m=self.data();d['records'][-1]['shown']=0
  with self.assertRaises(ValueError):validate(d,m,'combined')
if __name__=='__main__':unittest.main()
