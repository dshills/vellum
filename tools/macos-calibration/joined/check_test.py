import copy,importlib.util,sys,unittest
from pathlib import Path
from check import validate,upload,CODE
ours=sys.modules['check'];sys.modules['check']=upload
s=importlib.util.spec_from_file_location('joined_upload_fixture',CODE.parent/'upload/check_test.py');fixtures=importlib.util.module_from_spec(s);s.loader.exec_module(fixtures);sys.modules['check']=ours
class Joined(unittest.TestCase):
 def data(self):
  d,m=fixtures.Upload().data();d['joined_mode']='absolute-raster-upload-schedule-v1';d['request_schedule_phases']=[];d['request_schedule_anchors']=[];d['receipt_lateness_ns']=[2000]*248;rasters=[]
  for i,measurement in enumerate(d['measurements']['Samples']):
   f=measurement['Frame'];f.update({k:1 for k in ['InputNS','OwnerNS','ReconcileNS','LayoutPaintNS','GatherNS','EncodeSubmitNS','QueuedNS','AcquireNS','PresentCallNS']});f['TotalUIWorkNS']=106
   receipt=d['before_close']['Inputs'][i]['Received'];rasters.append(dict(Ordinal=i,Target=dict(Min=dict(X=0,Y=0),Max=dict(X=m['width'],Y=m['height'])),Begin=receipt+.0005,End=receipt+.0005+20e-9,Begun=True,Complete=True,Returned=True))
   d['request_schedule_phases'].append(dict(Sequence=i+1,WakeLatenessNS=100,ForegroundCheckNS=100,PostCheckLatenessNS=1000));d['request_schedule_anchors'].append(dict(Sequence=i+1,WakeHost=receipt-.00003,CheckedHost=receipt-.00002,PostCheckHost=receipt-.00001,TargetFromReceiptStartNS=-2000,ReceiptBracketNS=100))
  for j,p in enumerate(d['before_close']['Samples']):p['DrawableID']=j
  d['after_close']['Samples']=copy.deepcopy(d['before_close']['Samples']);d['raster_trace']=dict(Configured=True,Overflow=False,Invalid=False,Pending=0,Records=rasters);return d,m
 def test_valid_partition_and_same_ticket_rank(self):
  r=validate(*self.data());self.assertEqual(len(r['joined_rows']),248);self.assertEqual(r['measured'],120);self.assertFalse(r['release_qualified']);self.assertEqual(r['p95_rank_ticket']['phase'],'measured')
  for row in r['joined_rows']:self.assertAlmostEqual(sum(row['intervals_ns'].values()),row['wake_to_shown_ns'],delta=1000)
 def test_sparse_opaque_tags_are_joined_by_identity(self):
  d,m=self.data()
  for r in d['before_close']['Inputs']:r['Tag']=1000+3*r['Tag']
  for sample in d['measurements']['Samples']:sample['Frame']['Tag']=1000+3*sample['Frame']['Tag']
  for r in d['before_close']['Samples']:
   if r['Tag']:r['Tag']=1000+3*r['Tag']
  d['after_close']['Samples']=copy.deepcopy(d['before_close']['Samples'])
  self.assertEqual(len(validate(d,m)['joined_rows']),248)
 def test_refuses_broken_raster_terminal(self):
  for field,value in [('Configured',False),('Overflow',True),('Invalid',True),('Pending',1)]:
   d,m=self.data();d['raster_trace'][field]=value
   with self.subTest(field=field),self.assertRaises(ValueError):validate(d,m)
 def test_refuses_raster_and_source_corruption(self):
  mutations=[lambda d:d['raster_trace']['Records'][-1].update(Begun=False),lambda d:d['raster_trace']['Records'].pop(),lambda d:d['raster_trace']['Records'][-1].update(Returned=False),lambda d:d['raster_trace']['Records'][-1].update(Begin=float('nan')),lambda d:d['raster_trace']['Records'][-1]['Target']['Max'].update(X=1),lambda d:d['raster_trace']['Records'][-1].update(End=1000),lambda d:d['measurements']['Samples'][-1]['Frame'].update(TotalUIWorkNS=1),lambda d:d['raster_trace']['Records'].append(copy.deepcopy(d['raster_trace']['Records'][-1]))]
  def extra(d,shift):
   row=copy.deepcopy(d['raster_trace']['Records'][-1]);row['Ordinal']=len(d['raster_trace']['Records']);row['Begin']+=shift;row['End']+=shift;d['raster_trace']['Records'].append(row)
  mutations.extend([lambda d:extra(d,0),lambda d:extra(d,.005),lambda d:d['raster_trace']['Records'][-1].update(End=d['raster_trace']['Records'][-1]['Begin']+.00001)])
  for i,change in enumerate(mutations):
   d,m=self.data();change(d)
   with self.subTest(case=i),self.assertRaises(ValueError):validate(d,m)
 def test_refuses_schedule_and_parent_accounting_corruption(self):
  mutations=[lambda d:d['request_schedule_anchors'].pop(),lambda d:d['request_schedule_anchors'][-1].update(WakeHost=1000),lambda d:d['request_schedule_anchors'][-1].update(TargetFromReceiptStartNS=0),lambda d:d['request_schedule_anchors'][-1].update(ReceiptBracketNS=-1),lambda d:d['request_schedule_phases'][-1].update(Sequence=1),lambda d:d['request_schedule_phases'][-1].update(PostCheckLatenessNS=20_000_001),lambda d:d['native_upload_trace']['records'][-1].update(copy_ns=-1),lambda d:d['after_close']['Scopes'][0].update(Closed=False)]
  for i,change in enumerate(mutations):
   d,m=self.data();change(d)
   with self.subTest(case=i),self.assertRaises(ValueError):validate(d,m)
if __name__=='__main__':unittest.main()
