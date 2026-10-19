"""Synthetic join/schema controls. No native execution or presentation evidence."""
from pathlib import Path
import copy,hashlib,importlib.util,json,tempfile,unittest,contextlib,io,os,sys
from unittest.mock import patch
p=Path(__file__).with_name('check-presentation.py');s=importlib.util.spec_from_file_location('check_presentation',p);m=importlib.util.module_from_spec(s);s.loader.exec_module(m)
SOURCE='7'*40
BINARY='a'*64

def fixture(tree=False):
 config={'Method':'routed-tree-scroll'if tree else'routed-grid-scroll','Samples':120,'Warmup':64,'DataSize':10000,'Variable':False,'Deep':False,'Novel':False,'GPURequested':False}
 tags=list(range(1,187))if tree else list(range((1<<32)+1,(1<<32)+65))+list(range(1,121))
 frames=[];inputs=[];samples=[]
 for index,tag in enumerate(tags):
  at=10+index*.016
  frame={key:0 for key in m.FRAME_KEYS if key.endswith('NS')};frame.update(Tag=tag,Window='perf',Generation=index+9,Completed=True,Failure='')
  frames.append(frame);inputs.append({'Scope':1,'Tag':tag,'Received':at})
  if index==0:samples.append({'Sequence':1,'Scope':1,'Tag':tag,'Generation':0,'Composition':1,'Armed':False,'DrawableID':0,'Acquired':at+.0001,'Presented':at+.005,'State':'presented'})
  samples.append({'Sequence':len(samples)+1,'Scope':1,'Tag':tag,'Generation':index+9,'Composition':len(samples)+1,'Armed':True,'DrawableID':0,'Acquired':at+.001,'Presented':at+.01,'State':'presented'})
 snapshot={'Scopes':[{'ID':1,'Window':'perf','Closed':False}],'Inputs':inputs,'Samples':samples,'Pending':0,'Skipped':0,'Unavailable':0,'Invalid':0,'Retired':0,'Overflow':False,'Failed':False,'Failure':''}
 before={'Frames':{'Samples':frames,'Dropped':0,'Pending':0,'Failure':''},'Presentation':snapshot}
 after=copy.deepcopy(before);after['Presentation']['Scopes'][0]['Closed']=True
 tail={'Sequence':len(samples)+1,'Scope':1,'Tag':0,'Generation':0,'Composition':len(samples)+1,'Armed':False,'DrawableID':0,'Acquired':samples[-1]['Acquired']+1,'Presented':0,'State':'retired'}
 after['Presentation']['Samples'].append(tail);after['Presentation']['Retired']=1
 return {'Schema':'vellum-presentation-capture-v1','Result':'CAPTURED_DIAGNOSTIC','Window':'perf','RunError':'','BinarySHA256':BINARY,'CutoffMethod':'acquisition-sequence prefix at completed CPU cohort','Config':config,'Build':{'vcs':'git','vcs.revision':SOURCE,'vcs.modified':'false','go_version':'go1.25.6','GOOS':'darwin','GOARCH':'arm64','CGO_ENABLED':'1'},'ContinuousWork':False,'WorkerRetired':True,'RuntimeReturned':True,'AcquisitionCutoff':len(samples),'ReceiptCutoff':len(inputs),'BeforeClose':before,'AfterClose':after}

def mutate_path(j,path,value):
 node=j
 for item in path[:-1]:node=node[item]
 node[path[-1]]=value

def replace_after_prefix(j):
 native=j['BeforeClose']['Presentation']['Samples']
 for index,sample in enumerate(native,1):sample['Sequence']=index
 j['AcquisitionCutoff']=len(native)
 tail=j['AfterClose']['Presentation']['Samples'][-1]
 tail['Sequence']=len(native)+1
 j['AfterClose']['Presentation']['Samples']=copy.deepcopy(native)+[tail]

def composition_cases():
 for kind in ['reused-final','regressed','changed-generation','changed-armed','changed-tag']:
  j=fixture();native=j['BeforeClose']['Presentation']['Samples']
  if kind=='reused-final':native[2]['Composition']=native[1]['Composition']
  if kind=='regressed':native[2]['Composition']=1
  if kind=='changed-generation':native[1]['Composition']=native[0]['Composition']
  if kind=='changed-armed':native[0]['Generation']=native[1]['Generation'];native[1]['Composition']=native[0]['Composition']
  if kind=='changed-tag':
   clone=copy.deepcopy(native[0]);clone.update(Tag=0,Acquired=10.00005,Presented=10.004);native.insert(0,clone)
  replace_after_prefix(j)
  yield kind,j

class PresentationJournalTests(unittest.TestCase):
 def test_both_methods_all_tags_and_zero_drawable_id(self):
  for tree in [False,True]:
   j=fixture(tree);out=m.validate(j,SOURCE,BINARY)
   self.assertEqual(out['measured_input_latency_ms']['count'],120)
   self.assertAlmostEqual(out['measured_input_latency_ms']['p95'],10,places=8)
   self.assertEqual(len(out['latencies']),len(j['BeforeClose']['Frames']['Samples']))
   self.assertEqual(out['tail_records'],1)
   self.assertFalse(out['cadence_qualified']);self.assertFalse(out['physical_latency_qualified'])
 def test_generation_intermediate_and_presentation_order(self):
  j=fixture();j['BeforeClose']['Presentation']['Samples'][0]['Presented']=10.015
  j['AfterClose']['Presentation']['Samples'][0]['Presented']=10.015
  out=m.validate(j,SOURCE,BINARY);self.assertTrue(all(i>=0 for i in out['prefix_intervals_ms']))
  j['BeforeClose']['Presentation']['Samples'][0]['Presented']=10.01
  j['AfterClose']['Presentation']['Samples'][0]['Presented']=10.01
  self.assertEqual(m.validate(j,SOURCE,BINARY)['equal_presentation_times'],1)
  # Wrong final generation must fail even when another prefix carries the tag.
  j['BeforeClose']['Presentation']['Samples'][1]['Generation']+=1
  with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_external_modfile_without_vcs_requires_binary_binding(self):
  j=fixture()
  for key in ['vcs','vcs.revision','vcs.modified']:j['Build'].pop(key)
  out=m.validate(j,SOURCE,BINARY)
  self.assertIn('VCS stamp absent',out['source_binding'])
  with self.assertRaises(ValueError):m.validate(j,SOURCE,'b'*64)
 def test_schema_clock_identity_loss_and_scope_controls(self):
  controls=[
   (['Schema'],'unknown'),(['Result'],'FAILED_DIAGNOSTIC'),(['RunError'],'canceled'),(['WorkerRetired'],False),(['RuntimeReturned'],False),(['ContinuousWork'],True),(['CutoffMethod'],'successful samples only'),
   (['AcquisitionCutoff'],0),(['AcquisitionCutoff'],9999),(['ReceiptCutoff'],0),(['Config','Samples'],True),(['Config','Warmup'],65),(['Config','DataSize'],3),(['Config','Deep'],True),(['Config','Method'],'physical-pointer'),(['Config','GPURequested'],1),
   (['Build','vcs.modified'],'true'),(['Build','vcs.revision'],'8'*40),(['Build','go_version'],'go1.27.2'),(['BinarySHA256'],'b'*64),(['Build','GOOS'],'linux'),(['Build','GOARCH'],'amd64'),(['Build','CGO_ENABLED'],'0'),(['Build','-tags'],'offscreen'),(['Build','-race'],'true'),(['Build','-gcflags'],'-N -l'),
   (['BeforeClose','Frames','Dropped'],1),(['BeforeClose','Frames','Pending'],1),(['BeforeClose','Frames','Failure'],'refused'),(['BeforeClose','Frames','Samples',0,'Completed'],False),(['BeforeClose','Frames','Samples',0,'Tag'],True),(['BeforeClose','Frames','Samples',0,'Generation'],0),(['BeforeClose','Frames','Samples',0,'Window'],'other'),(['BeforeClose','Frames','Samples',0,'Unavailable'],True),(['BeforeClose','Frames','Samples',0,'SourceDetails'],{}),(['BeforeClose','Frames','Samples',0,'TotalUIWorkNS'],1),
   (['BeforeClose','Presentation','Scopes',0,'Closed'],True),(['BeforeClose','Presentation','Scopes',0,'ID'],2),(['BeforeClose','Presentation','Failed'],True),(['BeforeClose','Presentation','Overflow'],True),(['BeforeClose','Presentation','Failure'],'refused'),(['BeforeClose','Presentation','Pending'],1),
   (['BeforeClose','Presentation','Inputs',0,'Received'],float('nan')),(['BeforeClose','Presentation','Inputs',0,'Tag'],True),(['BeforeClose','Presentation','Inputs',0,'Scope'],2),(['BeforeClose','Presentation','Inputs',1,'Received'],1),
   (['BeforeClose','Presentation','Samples',0,'Sequence'],2),(['BeforeClose','Presentation','Samples',0,'Scope'],2),(['BeforeClose','Presentation','Samples',0,'Composition'],0),(['BeforeClose','Presentation','Samples',0,'Acquired'],0),(['BeforeClose','Presentation','Samples',0,'Presented'],9),(['BeforeClose','Presentation','Samples',1,'State'],'pending'),(['BeforeClose','Presentation','Samples',1,'State'],'skipped'),(['BeforeClose','Presentation','Samples',1,'State'],'retired'),(['BeforeClose','Presentation','Samples',1,'State'],'unavailable'),(['BeforeClose','Presentation','Samples',1,'State'],'invalid'),(['BeforeClose','Presentation','Samples',1,'Armed'],False),
   (['AfterClose','Presentation','Scopes',0,'Closed'],False),(['AfterClose','Presentation','Samples',0,'DrawableID'],9),(['AfterClose','Presentation','Samples',-1,'Tag'],1),(['AfterClose','Presentation','Pending'],1),(['AfterClose','Frames','Samples',0,'Generation'],999)
  ]
  for path,value in controls:
   with self.subTest(path=path,value=value):
    j=fixture();mutate_path(j,path,value)
    with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_missing_duplicate_unknown_and_capacity_controls(self):
  for kind in ['missing-cpu','duplicate-cpu','missing-receipt','duplicate-receipt','extra-armed','unknown-field','unknown-nested','native-record-limit','tagged-tail','cutoff-filters-final']:
   with self.subTest(kind=kind):
    j=fixture();p=j['BeforeClose']
    if kind=='missing-cpu':p['Frames']['Samples'].pop()
    if kind=='duplicate-cpu':p['Frames']['Samples'].append(copy.deepcopy(p['Frames']['Samples'][0]))
    if kind=='missing-receipt':p['Presentation']['Inputs'].pop()
    if kind=='duplicate-receipt':p['Presentation']['Inputs'].append(copy.deepcopy(p['Presentation']['Inputs'][0]))
    if kind=='extra-armed':p['Presentation']['Samples'][0]['Armed']=True;p['Presentation']['Samples'][0]['Generation']=9
    if kind=='unknown-field':j['passes60Hz']=True
    if kind=='unknown-nested':p['Presentation']['Samples'][0]['extra']=True
    if kind=='native-record-limit':p['Presentation']['Samples']=[p['Presentation']['Samples'][0]]*65537
    if kind=='tagged-tail':j['AfterClose']['Presentation']['Samples'][-1].update(Tag=1,Generation=1)
    if kind=='cutoff-filters-final':j['AcquisitionCutoff']-=1
    with self.assertRaises(ValueError):m.validate(j,SOURCE,BINARY)
 def test_composition_identity_cannot_be_reused_or_changed(self):
  for kind,j in composition_cases():
   with self.subTest(kind=kind):
    with self.assertRaisesRegex(ValueError,'composition'):m.validate(j,SOURCE,BINARY)
 def test_same_composition_may_repeat_with_identical_stamp(self):
  j=fixture();native=j['BeforeClose']['Presentation']['Samples']
  clone=copy.deepcopy(native[0]);clone['Acquired']+=.00001;clone['Presented']+=.00001
  native.insert(1,clone);replace_after_prefix(j)
  self.assertEqual(m.validate(j,SOURCE,BINARY)['prefix_presentations'],len(native))
 def test_cli_validates_binary_hash_and_remains_nonqualifying(self):
  with tempfile.TemporaryDirectory()as directory:
   path=Path(directory)/'journal.json';raw=json.dumps(fixture()).encode();path.write_bytes(raw)
   args=['check-presentation.py',str(path),'--sha256',hashlib.sha256(raw).hexdigest(),'--source',SOURCE,'--binary-sha256',BINARY]
   for bad in [None,'hash','source','binary']:
    argv=list(args)
    if bad=='hash':argv[argv.index('--sha256')+1]='0'*64
    if bad=='source':argv[argv.index('--source')+1]='0'*40
    if bad=='binary':argv[argv.index('--binary-sha256')+1]='0'*64
    stdout=io.StringIO()
    with patch.object(sys,'argv',argv),contextlib.redirect_stdout(stdout):code=m.main()
    result=json.loads(stdout.getvalue())
    if bad is None:self.assertEqual(code,0);self.assertFalse(result['cadence_qualified'])
    else:self.assertEqual(code,1);self.assertEqual(result['result'],'REFUSED_JOURNAL')
 def test_nonregular_and_symlink_input_refused_without_blocking(self):
  with tempfile.TemporaryDirectory()as directory:
   root=Path(directory);target=root/'target';target.write_text('{}')
   link=root/'link';link.symlink_to(target)
   fifo=root/'fifo';os.mkfifo(fifo)
   for path in [root,link,fifo]:
    with self.assertRaises((ValueError,OSError)):m.load(path,'0'*64)
 def test_bounded_strict_file_loading(self):
  with tempfile.TemporaryDirectory()as directory:
   path=Path(directory)/'journal.json'
   for raw in [b'{"x":1,"x":2}',b'{"x":NaN}',b'{"x":Infinity}',b'\xff',b'{"x":']:
    path.write_bytes(raw)
    with self.assertRaises((ValueError,UnicodeError)):m.load(path,hashlib.sha256(raw).hexdigest())
   raw=json.dumps(fixture()).encode();path.write_bytes(raw)
   self.assertEqual(m.load(path,hashlib.sha256(raw).hexdigest())['Schema'],'vellum-presentation-capture-v1')
   with self.assertRaises(ValueError):m.load(path,'0'*64)
   # The byte guard refuses before parsing, using a small limit for this control.
   old=m.MAX_BYTES
   try:
    m.MAX_BYTES=8
    with self.assertRaises(ValueError):m.load(path,hashlib.sha256(raw).hexdigest())
   finally:m.MAX_BYTES=old
if __name__=='__main__':unittest.main()
