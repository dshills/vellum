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
if __name__=='__main__':unittest.main()
