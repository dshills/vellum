import argparse
import json
import sys
import zipfile
from types import SimpleNamespace
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest import mock
spec=importlib.util.spec_from_file_location('bundle',Path(__file__).with_name('private-macos-bundle.py'))
bundle=importlib.util.module_from_spec(spec);spec.loader.exec_module(bundle)
class Admission(unittest.TestCase):
 def test_dirty_source_refuses(self):
  with mock.patch.object(bundle,'check_child_output',side_effect=['a'*40,' M source.go']):
   with self.assertRaisesRegex(ValueError,'commit source'):bundle.clean_source(Path('/unused'),{})
 def test_clean_source_identity(self):
  with mock.patch.object(bundle,'check_child_output',side_effect=['b'*40+'\n','']):self.assertEqual(bundle.clean_source(Path('/unused'),{}),'b'*40)
 def test_checksum_is_of_exact_bytes(self):
  with tempfile.TemporaryDirectory() as d:
   p=Path(d)/'app.zip';p.write_bytes(b'abc');self.assertEqual(bundle.digest(p),'ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad')
class Delivery(unittest.TestCase):
 def setUp(self):
  self.installer = bundle.load_script('install-private-macos')
 def fixture(self, root):
  out = root/'bundle'; out.mkdir()
  sdk = out/'Vellum-SDK.tar.gz'; sdk.write_bytes(b'fixture SDK')
  hashes = {sdk.name: bundle.digest(sdk)}; apps = []
  for name in sorted(self.installer.APP_NAMES):
   archive = out/(name.replace(' ', '-')+'.zip')
   with zipfile.ZipFile(archive, 'w') as z:
    z.writestr(name+'.app/Contents/MacOS/App', b'fixture executable')
   apps.append(dict(name=name, archive=archive.name))
   hashes[archive.name] = bundle.digest(archive)
  manifest = dict(source_commit='a'*40, build_version='23', private=True,
                  signing='ad-hoc', notarized=False, sdk_sha256=hashes[sdk.name], apps=apps, files=hashes)
  (out/'BUNDLE_MANIFEST.json').write_text(json.dumps(manifest))
  (out/'SHA256SUMS').write_text(''.join(f'{v}  {k}\n' for k,v in sorted(hashes.items())))
  return out, manifest
 def checker(self, commit='a'*40):
  return SimpleNamespace(verify_archive=mock.Mock(return_value={'manifest': {'source': {'commit': commit, 'clean': True}}}))
 def loader(self, checker):
  return lambda name: {'install-private-macos': self.installer, 'check-sdk': checker}[name]
 def test_delivery_installs_verified_bytes_and_records_no_runtime_claim(self):
  with tempfile.TemporaryDirectory() as d, mock.patch.object(sys, 'platform', 'darwin'), mock.patch.object(self.installer, 'signature') as signature:
   root=Path(d).resolve();out,manifest=self.fixture(root)
   old=root/'installed/old';old.mkdir(parents=True);(old/'keep').write_text('previous app')
   data=root/'Notebook';data.write_text('untouched user data')
   checker=self.checker()
   with mock.patch.object(bundle, 'load_script', side_effect=self.loader(checker)):
    result=bundle.install_delivery(out, root/'installed')
    self.assertEqual(result['source_commit'], manifest['source_commit'])
    self.assertEqual(result['build_version'], '23')
    self.assertEqual(result['native_runtime'], 'not executed')
    target=Path(result['installation'])
    self.assertEqual(result['installation_sha256'], bundle.digest(target/'INSTALLATION.json'))
    self.assertEqual(json.loads((out/'DELIVERY.json').read_text()), result)
    self.assertEqual(signature.call_count, 5)
    with self.assertRaisesRegex(ValueError, 'already exists'):
     bundle.install_delivery(out, root/'installed')
   self.assertEqual((old/'keep').read_text(), 'previous app')
   self.assertEqual(data.read_text(), 'untouched user data')
   checker.verify_archive.assert_called_with(out/'Vellum-SDK.tar.gz', manifest['sdk_sha256'])
 def test_corruption_and_source_mismatch_refuse_before_install(self):
  for failure in ('SDK', 'app', 'sums', 'source', 'incomplete'):
   with self.subTest(failure=failure),tempfile.TemporaryDirectory() as d:
    root=Path(d).resolve();out,m=self.fixture(root)
    if failure=='SDK': (out/'Vellum-SDK.tar.gz').write_bytes(b'corrupt')
    if failure=='app': (out/m['apps'][0]['archive']).write_bytes(b'corrupt')
    if failure=='incomplete':
     m['apps'].pop();(out/'BUNDLE_MANIFEST.json').write_text(json.dumps(m))
    if failure=='sums': (out/'SHA256SUMS').write_text('incorrect')
    checker=self.checker('b'*40 if failure=='source' else 'a'*40)
    with mock.patch.object(bundle, 'load_script', side_effect=self.loader(checker)),mock.patch.object(self.installer, 'install') as install:
     with self.assertRaises(ValueError): bundle.install_delivery(out, root/'installed')
     install.assert_not_called()
    self.assertFalse((out/'DELIVERY.json').exists())
 def test_signature_failure_never_publishes_delivery(self):
  with tempfile.TemporaryDirectory() as d,mock.patch.object(sys, 'platform', 'darwin'),mock.patch.object(self.installer, 'signature', side_effect=OSError('signature failed')):
   root=Path(d).resolve();out,_=self.fixture(root)
   with mock.patch.object(bundle, 'load_script', side_effect=self.loader(self.checker())):
    with self.assertRaisesRegex(OSError,'signature failed'):bundle.install_delivery(out,root/'installed')
   self.assertFalse((out/'DELIVERY.json').exists())
   self.assertFalse(list((root/'installed').glob('build*')))
 def test_manifest_drift_never_publishes_delivery(self):
  with tempfile.TemporaryDirectory() as d,mock.patch.object(sys, 'platform', 'darwin'),mock.patch.object(self.installer, 'signature'):
   root=Path(d).resolve();out,_=self.fixture(root);real_install=self.installer.install
   def drift(*args):
    target=real_install(*args)
    with (out/'BUNDLE_MANIFEST.json').open('a') as stream:stream.write(' ')
    return target
   with mock.patch.object(bundle,'load_script',side_effect=self.loader(self.checker())),mock.patch.object(self.installer,'install',side_effect=drift):
    with self.assertRaisesRegex(ValueError,'identity changed'):bundle.install_delivery(out,root/'installed')
   self.assertFalse((out/'DELIVERY.json').exists())
 def test_build_version_admission(self):
  self.assertEqual(bundle.build_version('23'),'23')
  for value in ('', '-1', '1/2', '1234567', 'v23'):
   with self.assertRaises(argparse.ArgumentTypeError):bundle.build_version(value)

if __name__=='__main__':unittest.main()
