import unittest,importlib.util,sys,copy
from pathlib import Path
from check import validate,source,CODE
ours=sys.modules['check'];sys.modules['check']=source
s=importlib.util.spec_from_file_location('sourcecost_fixture_methods',CODE.parent/'sourcecost/check_test.py');f=importlib.util.module_from_spec(s);s.loader.exec_module(f);sys.modules['check']=ours
class Upload(unittest.TestCase):
 def data(self,mode='on'):
  d,m=f.SourceCost().data();d['observer_mode']=mode;rows=[]
  for i,measurement in enumerate(d['measurements']['Samples']):
   r=measurement['Frame']['SourceDetails']['Records'][0];r.update(ImageWidth=m['width'],ImageHeight=m['height'],LogicalUploadRGBABytes=4*m['width']*m['height'])
   if mode=='on':
    receipt=d['before_close']['Inputs'][i]['Received'];begin=receipt+.001;rows.append(dict(ordinal=i,complete=True,status=1,stages=31,begin=begin,end=begin+5e-9,create_ns=1,map_ns=1,copy_ns=1,flush_ns=1,encode_ns=1,bytes=r['LogicalUploadRGBABytes'],width=m['width'],height=m['height'],stage_pitch=m['width']*4))
  d['native_upload_trace']=dict(enabled=mode=='on',pending=0,overflow=False,invalid_stage=False,count=len(rows),records=rows);return d,m
 def test_on(self):self.assertEqual(validate(*self.data(),'on')['joined_cohort_uploads'],248)
 def test_off(self):self.assertEqual(validate(*self.data('off'),'off')['observer_mode'],'off')
 def test_missing_failed_negative_and_oversized_stages(self):
  for key,value in [('stages',27),('complete',False),('status',2),('copy_ns',-1),('map_ns',float('nan')),('create_ns',10000),('bytes',1),('stage_pitch',1)]:
   with self.subTest(key=key):
    d,m=self.data();d['native_upload_trace']['records'][-1][key]=value
    with self.assertRaises(ValueError):validate(d,m,'on')
 def test_terminal_refusal(self):
  for key,value in [('pending',1),('overflow',True),('invalid_stage',True),('count',4097)]:
   with self.subTest(key=key):
    d,m=self.data();d['native_upload_trace'][key]=value
    with self.assertRaises(ValueError):validate(d,m,'on')
 def test_duplicate_upload(self):
  d,m=self.data();rows=d['native_upload_trace']['records'];r=copy.deepcopy(rows[-1]);r['ordinal']=len(rows);rows.append(r);d['native_upload_trace']['count']+=1
  with self.assertRaises(ValueError):validate(d,m,'on')
 def test_unknown_post_admission_upload(self):
  d,m=self.data();r=d['native_upload_trace']['records'][-1];r['begin']+=.02;r['end']+=.02
  with self.assertRaises(ValueError):validate(d,m,'on')
 def test_enclosing_API_interval(self):
  d,m=self.data();r=d['native_upload_trace']['records'][-1];r['end']+=.001
  with self.assertRaises(ValueError):validate(d,m,'on')
if __name__=='__main__':unittest.main()
