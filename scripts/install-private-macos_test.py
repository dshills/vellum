import hashlib,importlib.util,json,os,select,stat,subprocess,sys,tempfile,unittest,zipfile
from pathlib import Path
from unittest import mock
spec=importlib.util.spec_from_file_location('installer',Path(__file__).with_name('install-private-macos.py'))
i=importlib.util.module_from_spec(spec);spec.loader.exec_module(i)
class Installation(unittest.TestCase):
 def fixture(self,root):
  bundle=root/'bundle';bundle.mkdir();archive=bundle/'Vellum-Forge.zip'
  with zipfile.ZipFile(archive,'w') as z:z.writestr('Vellum Forge.app/Contents/MacOS/Forge',b'executable')
  m=dict(source_commit='a'*40,build_version='20',private=True,apps=[dict(name='Vellum Forge',archive=archive.name)],files={archive.name:i.digest(archive)})
  (bundle/'BUNDLE_MANIFEST.json').write_text(json.dumps(m));return bundle
 def test_install_repeat_and_preserve_previous_and_data(self):
  with tempfile.TemporaryDirectory() as t,mock.patch.object(i.sys,'platform','darwin'),mock.patch.object(i,'signature') as verify:
   root=Path(t).resolve();bundle=self.fixture(root);dest=root/'Applications';dest.mkdir();prior=dest/'old';prior.mkdir();(prior/'keep').write_text('old');data=root/'Notebook';data.write_text('user bytes')
   target=i.install(bundle,dest);self.assertEqual(i.install(bundle,dest),target)
   self.assertEqual((prior/'keep').read_text(),'old');self.assertEqual(data.read_text(),'user bytes');self.assertEqual(verify.call_count,2)
   (target/'apps/Vellum Forge.app/Contents/MacOS/Forge').write_text('changed')
   with self.assertRaisesRegex(ValueError,'bytes differ'):i.install(bundle,dest)
 def test_corruption_does_not_publish(self):
  with tempfile.TemporaryDirectory() as t,mock.patch.object(i.sys,'platform','darwin'),mock.patch.object(i,'signature') as verify:
   root=Path(t).resolve();bundle=self.fixture(root);(bundle/'Vellum-Forge.zip').write_bytes(b'bad')
   with self.assertRaisesRegex(ValueError,'SHA256'):i.install(bundle,root/'install')
   self.assertFalse(list((root/'install').glob('build*')));verify.assert_not_called()
 def test_reject_traversal_links_and_foreign_root(self):
  for name,mode in [('Vellum Forge.app/../escape',0),('/absolute',0),('Other.app/file',0),('Vellum Forge.app/link',stat.S_IFLNK|0o777)]:
   with self.subTest(name=name),tempfile.TemporaryDirectory() as t:
    p=Path(t)/'app.zip'
    with zipfile.ZipFile(p,'w') as z:
     info=zipfile.ZipInfo(name);info.external_attr=mode<<16;z.writestr(info,'data')
    with self.assertRaisesRegex(ValueError,'unsafe'):i.entries(p,'Vellum Forge')
 def test_synced_and_symlink_destinations_refused(self):
  with tempfile.TemporaryDirectory() as t,mock.patch.object(i.sys,'platform','darwin'):
   root=Path(t).resolve();bundle=self.fixture(root);real=root/'real';real.mkdir();linked=root/'link';linked.symlink_to(real,target_is_directory=True)
   with self.assertRaisesRegex(ValueError,'symlink'):i.install(bundle,linked/'apps')
   with mock.patch.object(i.Path,'home',return_value=root):
    with self.assertRaisesRegex(ValueError,'nonsynced'):i.install(bundle,root/'Documents/apps')
 def test_signature_failure_cannot_publish(self):
  with tempfile.TemporaryDirectory() as t,mock.patch.object(i.sys,'platform','darwin'),mock.patch.object(i,'signature',side_effect=OSError('invalid signature')):
   root=Path(t).resolve();bundle=self.fixture(root)
   with self.assertRaises(OSError):i.install(bundle,root/'install')
   self.assertFalse(list((root/'install').glob('build*')))
 def test_failed_stages_preserve_previous_installation_and_allow_retry(self):
  for boundary in ('extract', 'tree', 'signature', 'receipt', 'stage-rename', 'publish'):
   with self.subTest(boundary=boundary),tempfile.TemporaryDirectory() as t,mock.patch.object(i.sys,'platform','darwin'),mock.patch.object(i,'signature'):
    root=Path(t).resolve();bundle=self.fixture(root);dest=root/'installed'
    previous=i.install(bundle,dest)
    snapshot={p.relative_to(previous):p.read_bytes() for p in previous.rglob('*') if p.is_file()}
    manifest=bundle/'BUNDLE_MANIFEST.json';m=json.loads(manifest.read_text());m['build_version']='21';manifest.write_text(json.dumps(m))
    target=dest/('build21-'+'a'*12)
    real_extract=i.extract;real_write=Path.write_text;real_rename=Path.rename
    def failed_extract(archive,payload):
     real_extract(archive,payload)
     raise OSError('injected extraction failure')
    def failed_write(path,*args,**kwargs):
     if path.name=='INSTALLATION.json':raise OSError('injected receipt failure')
     return real_write(path,*args,**kwargs)
    def failed_rename(path,to):
     if (boundary=='stage-rename' and path.name=='apps') or (boundary=='publish' and path.name=='ready'):
      raise OSError('injected rename failure')
     return real_rename(path,to)
    patches={'extract':mock.patch.object(i,'extract',side_effect=failed_extract),
             'tree':mock.patch.object(i,'verify_tree',side_effect=ValueError('injected tree failure')),
             'signature':mock.patch.object(i,'signature',side_effect=OSError('injected signature failure')),
             'receipt':mock.patch.object(Path,'write_text',failed_write),
             'stage-rename':mock.patch.object(Path,'rename',failed_rename),
             'publish':mock.patch.object(Path,'rename',failed_rename)}
    with patches[boundary],self.assertRaisesRegex((OSError,ValueError),'injected'):i.install(bundle,dest)
    self.assertFalse(target.exists())
    self.assertFalse(list(dest.glob('.install-*')))
    self.assertEqual(snapshot,{p.relative_to(previous):p.read_bytes() for p in previous.rglob('*') if p.is_file()})
    self.assertEqual(i.install(bundle,dest),target)
    self.assertEqual(i.install(bundle,dest),target)
    self.assertEqual(snapshot,{p.relative_to(previous):p.read_bytes() for p in previous.rglob('*') if p.is_file()})
 @unittest.skipUnless(os.name == 'posix', 'requires POSIX process termination and pipe polling')
 def test_killed_extractor_never_publishes_and_next_install_succeeds(self):
  # Kill only our synthetic installer subprocess, never an app or user process.
  with tempfile.TemporaryDirectory() as t,mock.patch.object(i.sys,'platform','darwin'),mock.patch.object(i,'signature'):
   root=Path(t).resolve();bundle=self.fixture(root);dest=root/'installed'
   previous=i.install(bundle,dest)
   snapshot={p.relative_to(previous):p.read_bytes() for p in previous.rglob('*') if p.is_file()}
   manifest=bundle/'BUNDLE_MANIFEST.json';m=json.loads(manifest.read_text());m['build_version']='21';manifest.write_text(json.dumps(m))
   code="""import importlib.util,signal,sys
from pathlib import Path
sys.path.insert(0,sys.argv[1])
spec=importlib.util.spec_from_file_location('killed_installer',Path(sys.argv[1])/'install-private-macos.py')
i=importlib.util.module_from_spec(spec);spec.loader.exec_module(i)
i.sys.platform='darwin';i.signature=lambda app: None
extract=i.extract
def pause(archive,payload):
 extract(archive,payload)
 print('extracted',flush=True)
 signal.pause()
i.extract=pause
i.install(Path(sys.argv[2]),Path(sys.argv[3]))
"""
   child=subprocess.Popen([sys.executable,'-B','-c',code,str(Path(__file__).resolve().parent),str(bundle),str(dest)],stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
   try:
    self.assertTrue(select.select([child.stdout],[],[],10)[0],'child did not reach extraction')
    self.assertEqual(child.stdout.readline().strip(),'extracted')
    child.kill();child.communicate(timeout=10)
    self.assertLess(child.returncode,0)
   finally:
    if child.poll() is None:child.kill();child.communicate(timeout=10)
   self.assertFalse((dest/('build21-'+'a'*12)).exists())
   self.assertTrue(list(dest.glob('.install-*')),'abrupt termination should retain only staging debris')
   self.assertEqual(snapshot,{p.relative_to(previous):p.read_bytes() for p in previous.rglob('*') if p.is_file()})
   target=i.install(bundle,dest)
   self.assertEqual(i.install(bundle,dest),target)
   self.assertEqual(snapshot,{p.relative_to(previous):p.read_bytes() for p in previous.rglob('*') if p.is_file()})
if __name__=='__main__':unittest.main()
