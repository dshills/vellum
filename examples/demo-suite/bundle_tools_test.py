import json
import sys
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

from bundle_tools import digest, identity, verify_suite
import build
import launch


class DeliveryTests(unittest.TestCase):
    def test_version_and_toolchain_come_from_matching_source(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / 'release').mkdir()
            (root / 'release/version.json').write_text('{"version":"v2.3.4"}')
            (root / 'vellum.go').write_text('const Version = "v2.3.4"')
            (root / 'go.mod').write_text('module example.com/test\n\ngo 1.25.6\n')
            self.assertEqual(identity(root), ('v2.3.4', '1.25.6'))
            (root / 'vellum.go').write_text('const Version = "v2.3.5"')
            with self.assertRaises(ValueError):
                identity(root)

    def test_output_reservation_preserves_existing_delivery_and_refuses_alias(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary).resolve()
            existing = root / 'delivery'
            existing.mkdir()
            (existing / 'keep').write_text('original')
            with self.assertRaises(FileExistsError):
                build.reserve_output(existing)
            self.assertEqual((existing / 'keep').read_text(), 'original')
            alias = root / 'alias'
            alias.symlink_to(root, target_is_directory=True)
            with self.assertRaises(ValueError):
                build.reserve_output(alias / 'new')
            self.assertFalse((root / 'new').exists())

    @unittest.skipUnless(sys.platform == "darwin", "native macOS staging")
    def test_delivery_destination_never_holds_executable_bundles(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary).resolve()
            with build.build_locations(root / 'synced-delivery') as (delivery, staging):
                self.assertEqual(delivery, root / 'synced-delivery')
                self.assertEqual(list(delivery.iterdir()), [])
                self.assertTrue(staging.is_relative_to(Path('/private/tmp')))
                self.assertFalse(staging.is_relative_to(delivery))
                self.assertEqual(staging.stat().st_mode & 0o777, 0o700)
            staging.rmdir()

    @unittest.skipUnless(sys.platform == "darwin", "native macOS staging")
    def test_failed_build_removes_only_its_staging(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary).resolve()
            with self.assertRaisesRegex(RuntimeError, 'synthetic build failure'):
                with build.build_locations(root / 'delivery') as (delivery, staging):
                    (staging / 'partial-app').write_text('generated')
                    (delivery / 'diagnostic').write_text('preserved')
                    raise RuntimeError('synthetic build failure')
            self.assertFalse(staging.exists())
            self.assertEqual((delivery / 'diagnostic').read_text(), 'preserved')

    def test_modified_selected_archive_is_refused_before_extract_or_launch(self):
        with tempfile.TemporaryDirectory() as temporary:
            archive = Path(temporary) / 'suite.zip'
            archive.write_bytes(b'original')
            expected = digest(archive)
            archive.write_bytes(b'changed')
            with patch('subprocess.run') as run:
                with self.assertRaises(ValueError):
                    launch.selected_suite({'archive':str(archive), 'archive_sha256':expected, 'suite':str(Path(temporary)/'suite')})
                run.assert_not_called()

    def test_restore_remembers_verified_location_without_launching(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / 'build').mkdir()
            receipt = {'version':'v2.3.4','source_commit':'fixture','suite':str(root/'expired')}
            (root/'build/latest-suite.json').write_text(json.dumps(receipt))
            restored = root / 'restored'
            with patch.object(launch, 'HERE', root), patch('bundle_tools.HERE', root), patch.object(launch, 'selected_suite', return_value=restored), patch('sys.argv',['launch.py','--verify-only']), patch('subprocess.run') as run:
                launch.main()
                run.assert_not_called()
            saved = json.loads((root/'build/latest-suite.json').read_text())
            self.assertEqual(saved['suite'],str(restored))
            self.assertEqual(saved['version'],receipt['version'])
            self.assertEqual(list((root/'build').glob('.suite-*')), [])

    def test_manifest_bundle_path_cannot_escape_suite(self):
        with tempfile.TemporaryDirectory() as temporary:
            suite = Path(temporary)
            path = suite / 'BUILD.json'
            path.write_text(json.dumps({'features':[{'id':'hub','name':'../outside'}]}))
            with patch('subprocess.run') as run:
                with self.assertRaises(ValueError):
                    verify_suite(suite,digest(path))
                run.assert_not_called()

    def test_changed_packaged_executable_is_refused(self):
        with tempfile.TemporaryDirectory() as temporary:
            suite = Path(temporary)
            binary = suite / 'replay'
            binary.write_bytes(b'original')
            manifest = {'features':[{'id':'replay','packaged_binary_sha256':digest(binary)}]}
            path = suite / 'BUILD.json'
            path.write_text(json.dumps(manifest))
            expected = digest(path)
            binary.write_bytes(b'changed')
            with self.assertRaises(ValueError):
                verify_suite(suite, expected)


if __name__ == '__main__':
    unittest.main()
