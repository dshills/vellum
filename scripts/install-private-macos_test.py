import hashlib,importlib.util,json,os,stat,tempfile,unittest,zipfile
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
if __name__=='__main__':unittest.main()
