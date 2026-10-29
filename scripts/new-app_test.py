"""Generated document storage identity and application-owned template checks."""
import hashlib
import importlib.util
from pathlib import Path
import tempfile
import unittest

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


if __name__ == '__main__':
    unittest.main()
