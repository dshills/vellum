import copy,importlib.util,sys,unittest
from check import validate,joined,backend,CODE
ours=sys.modules['check'];sys.modules['check']=joined
s=importlib.util.spec_from_file_location('joined_backend_fixtures',CODE.parent/'joined/check_test.py');fixtures=importlib.util.module_from_spec(s);s.loader.exec_module(fixtures);sys.modules['check']=ours
class JoinedBackend(unittest.TestCase):
 def data(self):
  d,m=fixtures.Joined().data();d['joined_backend_mode']='same-input-cpu-submit-gpu-shown-v1';rows=[]
  for frame in d['before_close']['Samples']:
   a=frame['Acquired']
   for offset in [0,.00005]:
    rows.append(dict(ordinal=len(rows),kind=0,queue=10,drawable_id=0,drawable_recorded=False,commit_start=a+.001+offset,commit_end=a+.00101+offset,present_request=0,scheduled_observed=a+.002,complete_observed=a+.004,gpu_start=a+.002+offset,gpu_end=a+.003+offset,shown=0,status=4))
   rows.append(dict(ordinal=len(rows),kind=2,queue=10,drawable_id=frame['DrawableID'],drawable_recorded=True,commit_start=a+.0012,commit_end=a+.0013,present_request=a+.0011,scheduled_observed=a+.005,complete_observed=a+.006,gpu_start=0,gpu_end=0,shown=frame['Presented'],status=4))
  d['backend_trace']=dict(enabled=True,pending=0,overflow=False,count=len(rows),records=rows);return d,m
 def test_same_input_partition_and_overlapping_request(self):
  r=validate(*self.data());self.assertEqual(len(r['joined_rows']),248);self.assertFalse(r['release_qualified'])
  for row in r['joined_rows']:
   self.assertEqual(len(row['backend_render_ordinals']),2);self.assertLess(row['present_request_minus_render_gpu_end_ns'],0)
   self.assertAlmostEqual(sum(row['post_acquire_ns'].values()),row['intervals_ns']['acquire_to_shown'],delta=1000)
  self.assertEqual(r['p95_rank_ticket']['phase'],'measured')
 def test_sparse_tags(self):
  d,m=self.data()
  for folder in ['before_close','after_close']:
   for r in d[folder]['Inputs']:r['Tag']=r['Tag']*3+1000
   for r in d[folder]['Samples']:
    if r['Armed']:r['Tag']=r['Tag']*3+1000
  for r in d['measurements']['Samples']:r['Frame']['Tag']=r['Frame']['Tag']*3+1000
  self.assertEqual(len(validate(d,m)['joined_rows']),248)
 def test_corruption_refused(self):
  mutations=[lambda d:d.update(joined_backend_mode='wrong'),lambda d:d['backend_trace'].update(pending=1),lambda d:d['backend_trace'].update(overflow=True),lambda d:d['backend_trace']['records'][-2].update(gpu_start=0),lambda d:d['backend_trace']['records'][-2].update(gpu_end=1000),lambda d:d['backend_trace']['records'][-1].update(drawable_id=99999),lambda d:d['backend_trace']['records'][-1].update(drawable_recorded=False),lambda d:d['backend_trace']['records'][-1].update(shown=0),lambda d:d['backend_trace']['records'][-2].update(complete_observed=0),lambda d:d['backend_trace']['records'][-1].update(gpu_start=1,gpu_end=0),lambda d:d['raster_trace'].update(Pending=1),lambda d:d['native_upload_trace'].update(overflow=True),lambda d:d['end'].update(losses=1),lambda d:d['backend_trace']['records'][-1].update(drawable_id=d['backend_trace']['records'][-4]['drawable_id']),lambda d:d['backend_trace']['records'][-1].update(gpu_start=1000,gpu_end=1001),lambda d:[r.update(kind=1) for r in d['backend_trace']['records'][-3:-1]]]
  for i,change in enumerate(mutations):
   d,m=self.data();change(d)
   with self.subTest(case=i),self.assertRaises(ValueError):validate(d,m)
 def test_trailing_submission_refused(self):
  d,m=self.data();row=copy.deepcopy(d['backend_trace']['records'][-2]);row.update(ordinal=len(d['backend_trace']['records']),commit_start=39,commit_end=39.001,scheduled_observed=39.002,complete_observed=39.004,gpu_start=39.002,gpu_end=39.003);d['backend_trace']['records'].append(row);d['backend_trace']['count']+=1
  with self.assertRaisesRegex(ValueError,'trailing'):validate(d,m)
 def test_commit_before_drawable_refused(self):
  d,m=self.data();r=d['backend_trace']['records'][-3];r['commit_start']-=.002;r['commit_end']-=.002
  with self.assertRaisesRegex(ValueError,'partition order'):validate(d,m)
 def test_unrelated_intercohort_presentation_refused(self):
  d,m=self.data();frames=d['before_close']['Samples'];last=frames[-1];a=last['Acquired'];extra=copy.deepcopy(last)
  extra.update(Armed=False,Tag=0,Generation=0,Acquired=a-.06,Presented=a-.05)
  last['DrawableID']+=1;last['Sequence']+=1;frames.insert(-1,extra);d['after_close']['Samples']=copy.deepcopy(frames)
  trace=d['backend_trace']['records'];trace[-1]['drawable_id']=last['DrawableID'];new=copy.deepcopy(trace[-3:])
  for r in new:
   for k in ['commit_start','commit_end','scheduled_observed','complete_observed','gpu_start','gpu_end','present_request']:
    if r[k]:r[k]-=.06
   if r['kind']==2:r.update(drawable_id=extra['DrawableID'],shown=extra['Presented'])
  trace[-3:-3]=new
  for i,r in enumerate(trace):r['ordinal']=i
  d['backend_trace']['count']=len(trace)
  self.assertEqual(len(joined.validate(d,m)['joined_rows']),248)
  self.assertEqual(backend.validate(d,m,'on')['measured'],120)
  with self.assertRaisesRegex(ValueError,'outside startup'):validate(d,m)
if __name__=='__main__':unittest.main()
