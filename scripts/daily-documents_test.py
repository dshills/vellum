"""Safety checks for the persistent evaluation directory."""
import importlib.util
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location('daily_documents', Path(__file__).with_name('daily-documents.py'))
daily = importlib.util.module_from_spec(spec)
spec.loader.exec_module(daily)


class DirectoryTests(unittest.TestCase):
    def test_check_does_not_create_and_launch_preparation_is_private(self):
        with tempfile.TemporaryDirectory() as root:
            path = Path(root).resolve() / 'evaluation'
            daily.private_directory(path)
            self.assertFalse(path.exists())
            daily.private_directory(path, create=True)
            self.assertEqual(path.stat().st_mode & 0o777, 0o700)

    def test_rejects_link_ancestor_and_public_existing_directory(self):
        with tempfile.TemporaryDirectory() as root:
            root = Path(root).resolve()
            link = root / 'link'
            link.symlink_to(root, target_is_directory=True)
            with self.assertRaises(ValueError):
                daily.private_directory(link / 'evaluation', create=True)
            public = root / 'public'
            public.mkdir(mode=0o755)
            public.chmod(0o755)
            with self.assertRaises(ValueError):
                daily.private_directory(public, create=True)
            self.assertEqual(public.stat().st_mode & 0o777, 0o755)


if __name__ == '__main__':
    unittest.main()
