import copy
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
spec=importlib.util.spec_from_file_location('candidate',Path(__file__).with_name('private-candidate.py'))
m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)


class Candidate(unittest.TestCase):
 def fixture(self,root):
  r=dict(schema=1,id='build24-'+'a'*12,source_commit='a'*40,build_version='24',sdk_sha256='b'*64,manifest_sha256='c'*64,
      installation_sha256='d'*64,preflight_sha256='e'*64,installation='Applications/Vellum/build24-'+'a'*12,
      packet='Documents/fixture packet',minimum_macos='27.0',inspection_host='27.2',signing='ad-hoc',notarized=False,
      native_runtime='not executed',desktop_checks='deferred',packaging_report='docs/reports/test/packaging.json',
      verification_report='docs/reports/test/verification.json',linkage_report='docs/reports/test/linkage.json')
  (root/'release').mkdir();(root/m.INDEX).write_text(json.dumps(r))
  for slot,name in m.SLOTS.items():
   p=root/name;p.parent.mkdir(parents=True,exist_ok=True)
   p.write_text('Before\n<!-- vellum-candidate:'+slot+':start -->\n<!-- vellum-candidate:'+slot+':end -->\nAfter\n')
  p=copy.deepcopy(r);p['installation']='~/'+r['installation'];p['apps']=['app'+str(i) for i in range(5)]
  v=dict(source_commit=r['source_commit'],build_version='24',sdk_sha256=r['sdk_sha256'],manifest_sha256=r['manifest_sha256'],
     installation=p['installation'],apps_verified=5,native_runtime='not executed')
  l=dict(source_commit=r['source_commit'],build_version='24',installation_sha256=r['installation_sha256'],native_runtime='not executed',
      host_version='27.2',apps=[dict(name=n,deployment=dict(minimum='27.0')) for n in p['apps']])
  for key,value in [('packaging_report',p),('verification_report',v),('linkage_report',l)]:
   path=root/r[key];path.parent.mkdir(parents=True,exist_ok=True);path.write_text(json.dumps(value))
  return r
 def test_evidence_and_generated_blocks_are_consistent_and_idempotent(self):
  with tempfile.TemporaryDirectory() as t:
   root=Path(t);r=self.fixture(root);self.assertEqual(m.load(root),r);m.evidence(root,r)
   with self.assertRaisesRegex(ValueError,'stale'):m.documents(root,r)
   m.documents(root,r,True);m.documents(root,r)
   before={name:(root/name).read_bytes() for name in m.SLOTS.values()};m.documents(root,r,True)
   self.assertEqual(before,{name:(root/name).read_bytes() for name in m.SLOTS.values()})
   r['sdk_sha256']='f'*64
   with self.assertRaisesRegex(ValueError,'differs'):m.evidence(root,r)
   with self.assertRaisesRegex(ValueError,'stale'):m.documents(root,r)
 def test_unsafe_paths_scope_and_identity_drift_refused(self):
  for field,value in [('packet','../Notebook'),('packet','Documents/$(bad)'),('installation','Applications/other'),
                      ('source_commit','b'*40),('notarized',True),('native_runtime','passed'),('minimum_macos','28.0'),
                      ('linkage_report','/outside.json')]:
   with self.subTest(field=field),tempfile.TemporaryDirectory() as t:
    root=Path(t);r=self.fixture(root);r[field]=value;(root/m.INDEX).write_text(json.dumps(r))
    with self.assertRaises(ValueError):m.load(root)
 def test_evidence_drift_and_missing_markers_do_not_mutate_guides(self):
  with tempfile.TemporaryDirectory() as t:
   root=Path(t);r=self.fixture(root);m.documents(root,r,True)
   path=root/r['linkage_report'];l=json.loads(path.read_text());l['apps'][0]['deployment']['minimum']='28.0';path.write_text(json.dumps(l))
   with self.assertRaisesRegex(ValueError,'coverage'):m.evidence(root,r)
   second=root/m.SLOTS['commands'];second.write_text('no marker')
   before=(root/m.SLOTS['facts']).read_bytes()
   with self.assertRaisesRegex(ValueError,'markers'):m.documents(root,r,True)
   self.assertEqual(before,(root/m.SLOTS['facts']).read_bytes())

if __name__=='__main__':unittest.main()
