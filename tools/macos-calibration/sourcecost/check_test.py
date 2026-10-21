from pathlib import Path
import unittest,importlib.util,sys,copy
from check import validate,base,cost,CODE
def module(n,p):
 s=importlib.util.spec_from_file_location(n,p);m=importlib.util.module_from_spec(s);s.loader.exec_module(m);return m
ours=sys.modules['check'];sys.modules['check']=base;fixtures=module('source_fixtures',CODE.parent/'check_test.py');sys.modules['check']=ours
helper=module('source_details_fixtures',CODE.parents[2]/'scripts/benchmark-grid_test.py')
class SourceCost(unittest.TestCase):
 def data(self):
  d,m=fixtures.Controls().data('framework');d['diagnostic_mode']='source-cost-only'
  for v in d['measurements']['Samples']:v['Frame'].update(RasterUploadNS=100,SourceDetails=dict(Compositions=1,UploadFailuresAvailable=False,Records=[helper.source_record()]))
  return d,m
 def test_valid_nested_cost(self):
  r=validate(*self.data());self.assertEqual(r['totals']['LogicalUploadRGBABytes'],120*800);self.assertFalse(r['latency_gate_qualified'])
 def test_missing_detail(self):
  d,m=self.data();del d['measurements']['Samples'][-1]['Frame']['SourceDetails']
  with self.assertRaises(ValueError):validate(d,m)
 def test_nested_and_counter_failures(self):
  for k,v in [('RasterNS',100),('CopyNS',100),('LogicalUploadRGBABytes',801),('ImageCacheHits',1),('ImageWidth',-1),('SetFromGoImageNS',99)]:
   with self.subTest(k=k):
    d,m=self.data();d['measurements']['Samples'][-1]['Frame']['SourceDetails']['Records'][0][k]=v
    with self.assertRaises(ValueError):validate(d,m)
 def test_warmup_validated(self):
  d,m=self.data();d['measurements']['Samples'][0]['Frame']['SourceDetails']['Compositions']=33
  with self.assertRaises(ValueError):validate(d,m)
 def test_marker(self):
  d,m=self.data();del d['diagnostic_mode']
  with self.assertRaises(ValueError):validate(d,m)
if __name__=='__main__':unittest.main()
