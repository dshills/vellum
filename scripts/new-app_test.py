"""Generated document storage identity and application-owned template checks."""
import hashlib
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location('new_app', Path(__file__).with_name('new-app.py'))
generator = importlib.util.module_from_spec(spec)
spec.loader.exec_module(generator)


class Templates(unittest.TestCase):
    def test_document_identity_is_deterministic_and_separate_per_module(self):
        with tempfile.TemporaryDirectory() as tmp:
            outputs = []
            for index, module in enumerate(['example.com/first', 'example.com/first', 'example.com/second']):
                stage = Path(tmp) / str(index)
                stage.mkdir()
                generator.copy_document_template(stage, module)
                text = (stage / 'paths.go').read_text()
                identity = 'vellum.document.' + hashlib.sha256(module.encode()).hexdigest()[:32]
                self.assertIn('"' + identity + '"', text)
                self.assertIn(identity, (stage / 'README.md').read_text())
                self.assertIn('resolveDataDirectory(*directory, os.UserConfigDir)', (stage / 'main.go').read_text())
                self.assertIn(module + '/internal/notebook', (stage / 'main.go').read_text())
                self.assertTrue((stage / 'paths_test.go').is_file())
                for path in stage.rglob('*.go'):
                    self.assertNotIn('__VELLUM_', path.read_text())
                    self.assertNotIn('github.com/dshills/vellum/examples/macapps', path.read_text())
                outputs.append(text)
            self.assertEqual(outputs[0], outputs[1])
            self.assertNotEqual(outputs[0], outputs[2])

    def test_document_testdata_keeps_exact_bytes_and_provenance(self):
        with tempfile.TemporaryDirectory() as tmp:
            stage = Path(tmp)
            generator.copy_document_template(stage, 'example.com/fixture-consumer')
            original = generator.ROOT/'examples/macapps/notes/testdata'
            expected = {path.relative_to(original): path.read_bytes() for path in original.rglob('*') if path.is_file()}
            copied = stage/'internal/notebook/testdata'
            actual = {path.relative_to(copied): path.read_bytes() for path in copied.rglob('*') if path.is_file()}
            self.assertTrue(expected)
            self.assertEqual(actual, expected)
            self.assertTrue((stage/'internal/notebook/compatibility_test.go').is_file())
            self.assertTrue((stage/'internal/notebook/codec_fuzz_test.go').is_file())

    def test_document_testdata_links_are_refused(self):
        for directory_link in (False, True):
            with self.subTest(directory_link=directory_link), tempfile.TemporaryDirectory() as tmp:
                root=Path(tmp)/'root';stage=Path(tmp)/'stage';stage.mkdir()
                for relative in ['examples/macapps/notes/model.go','examples/macapps/artwork/icon.go','examples/macapps/cmd/notes/toolbar.go']:
                    target=root/relative;target.parent.mkdir(parents=True,exist_ok=True);target.write_text('package fixture')
                fixtures=root/'examples/macapps/notes/testdata'
                outside=Path(tmp)/'outside';outside.mkdir();(outside/'unrelated').write_text('preserved')
                if directory_link:fixtures.symlink_to(outside,target_is_directory=True)
                else:
                    fixtures.mkdir();(fixtures/'linked').symlink_to(outside/'unrelated')
                with mock.patch.object(generator,'ROOT',root),self.assertRaisesRegex(ValueError,'testdata'):
                    generator.copy_document_template(stage,'example.com/test')
                self.assertEqual((outside/'unrelated').read_text(),'preserved')


if __name__ == '__main__':
    unittest.main()
