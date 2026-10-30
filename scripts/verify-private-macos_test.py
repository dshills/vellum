"""Fictional delivery fixtures; mock only native signing and SDK archive parsing."""
import importlib.util
import json
from pathlib import Path
import plistlib
import sys
import tempfile
import unittest
from unittest import mock
import zipfile

spec = importlib.util.spec_from_file_location('verify_delivery', Path(__file__).with_name('verify-private-macos.py'))
v = importlib.util.module_from_spec(spec); spec.loader.exec_module(v)


class Verification(unittest.TestCase):
    def fixture(self, root):
        bundle = root/'bundle'; bundle.mkdir()
        sdk = bundle/'Vellum-SDK.tar.gz'; sdk.write_bytes(b'fictional SDK')
        files = {sdk.name: v.installer.digest(sdk)}; apps = []
        for name, identifier in v.IDENTIFIERS.items():
            archive = bundle/(name.replace(' ', '-')+'.zip')
            binary_hash = 'c'*64
            with zipfile.ZipFile(archive, 'w') as z:
                prefix = name+'.app/Contents/'
                z.writestr(prefix+'MacOS/'+name, b'fictional executable')
                z.writestr(prefix+'Info.plist', plistlib.dumps(dict(CFBundleName=name, CFBundleExecutable=name,
                    CFBundleIdentifier=identifier, CFBundleVersion='23')))
                z.writestr(prefix+'Resources/APP_MANIFEST.json', json.dumps(dict(name=name, identifier=identifier,
                    build_version='23', source_binary_sha256=binary_hash, private_build=True, signed_identity='-', notarized=False)))
            files[archive.name] = v.installer.digest(archive)
            apps.append(dict(name=name, archive=archive.name, binary_sha256=binary_hash))
        preflight = bundle/'PREFLIGHT.json'; preflight.write_text(json.dumps(dict(source_commit='a'*40)))
        m = dict(source_commit='a'*40, build_version='23', private=True, signing='ad-hoc', notarized=False,
                 sdk_sha256=files[sdk.name], apps=apps, files=files, preflight_sha256=v.installer.digest(preflight))
        (bundle/'BUNDLE_MANIFEST.json').write_text(json.dumps(m))
        (bundle/'SHA256SUMS').write_text(''.join(f'{h}  {n}\n' for n, h in sorted(files.items())))
        target = v.installer.install(bundle, root/'installed')
        (bundle/'DELIVERY.json').write_text(json.dumps(dict(source_commit=m['source_commit'], build_version='23',
            sdk_sha256=m['sdk_sha256'], manifest_sha256=v.installer.digest(bundle/'BUNDLE_MANIFEST.json'),
            installation=str(target), installation_sha256=v.installer.digest(target/'INSTALLATION.json'),
            private=True, signing='ad-hoc', notarized=False, native_runtime='not executed', preflight_sha256=m['preflight_sha256'])))
        return bundle, target

    def setUp(self):
        self.platform = mock.patch.object(sys, 'platform', 'darwin'); self.platform.start(); self.addCleanup(self.platform.stop)
        self.sign = mock.patch.object(v.installer, 'signature'); self.signature = self.sign.start(); self.addCleanup(self.sign.stop)
        self.sdk = mock.patch.object(v.checker, 'verify_archive', return_value={'manifest': {'source': {'commit': 'a'*40, 'clean': True}}})
        self.sdk_check = self.sdk.start(); self.addCleanup(self.sdk.stop)

    def test_success_is_read_only_and_does_not_install_or_launch(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary).resolve(); bundle, target = self.fixture(root)
            snapshot = {p.relative_to(root): p.read_bytes() for p in root.rglob('*') if p.is_file()}
            self.signature.reset_mock()
            with mock.patch.object(v.installer, 'install', side_effect=AssertionError('must not install')):
                result = v.verify(bundle, target, v.installer.digest(bundle/'BUNDLE_MANIFEST.json'))
            self.assertEqual(result['apps_verified'], 5)
            self.assertEqual(result['native_runtime'], 'not executed')
            self.assertFalse(result['independent_authenticity_verified'])
            self.assertEqual(self.signature.call_count, 10)
            self.assertEqual(snapshot, {p.relative_to(root): p.read_bytes() for p in root.rglob('*') if p.is_file()})

    def test_corruption_refused(self):
        for kind in ('SDK', 'archive', 'sums', 'delivery', 'receipt', 'preflight', 'installed', 'extra-file', 'link', 'source', 'pin', 'signature'):
            with self.subTest(kind=kind), tempfile.TemporaryDirectory() as temporary:
                root = Path(temporary).resolve(); bundle, target = self.fixture(root)
                self.sdk_check.return_value = {'manifest': {'source': {'commit': 'b'*40 if kind=='source' else 'a'*40, 'clean': True}}}
                if kind=='SDK': (bundle/'Vellum-SDK.tar.gz').write_bytes(b'changed')
                if kind=='archive': (bundle/'Vellum-Forge.zip').write_bytes(b'changed')
                if kind=='sums': (bundle/'SHA256SUMS').write_text('changed')
                if kind in ('delivery','receipt','preflight'):
                    path = {'delivery':bundle/'DELIVERY.json', 'receipt':target/'INSTALLATION.json', 'preflight':bundle/'PREFLIGHT.json'}[kind]
                    data=json.loads(path.read_text()); data['source_commit']='b'*40; path.write_text(json.dumps(data))
                if kind=='installed': (target/'apps/Vellum Forge.app/Contents/MacOS/Vellum Forge').write_bytes(b'changed')
                if kind=='extra-file': (target/'apps/extra').write_text('unrecorded')
                if kind=='link': (target/'apps/alias').symlink_to(root)
                if kind=='signature': self.signature.side_effect=ValueError('bad signature')
                try:
                    with self.assertRaises((ValueError, OSError)):
                        v.verify(bundle, target, '0'*64 if kind=='pin' else None)
                finally:
                    self.signature.side_effect=None

    def test_inconsistent_app_metadata_is_refused_after_matching_hashes(self):
        with tempfile.TemporaryDirectory() as temporary:
            root=Path(temporary).resolve(); bundle,target=self.fixture(root)
            # Metadata check is separate from byte integrity and signature verification.
            item=json.loads((bundle/'BUNDLE_MANIFEST.json').read_text())['apps'][0]
            app=target/'apps'/(item['name']+'.app')
            p=app/'Contents/Info.plist'; info=plistlib.loads(p.read_bytes());info['CFBundleVersion']='99';p.write_bytes(plistlib.dumps(info))
            with self.assertRaisesRegex(ValueError,'app plist'):v.verify_app(app,item,'23')
            info['CFBundleVersion']='23';p.write_bytes(plistlib.dumps(info))
            p=app/'Contents/Resources/APP_MANIFEST.json';m=json.loads(p.read_text());m['source_binary_sha256']='d'*64;p.write_text(json.dumps(m))
            with self.assertRaisesRegex(ValueError,'app manifest'):v.verify_app(app,item,'23')

    def test_incomplete_or_malformed_manifest_and_symlink_roots_refused(self):
        for kind in ('incomplete','duplicate','nonobject','unsafe-archive','symlink-root'):
            with self.subTest(kind=kind),tempfile.TemporaryDirectory() as temporary:
                root=Path(temporary).resolve(); bundle,target=self.fixture(root)
                p=bundle/'BUNDLE_MANIFEST.json';m=json.loads(p.read_text())
                if kind=='incomplete':m['apps'].pop()
                if kind=='duplicate':m['apps'][-1]=m['apps'][0]
                if kind=='nonobject':m=[]
                if kind=='unsafe-archive':m['apps'][0]['archive']='../escape.zip'
                if kind=='symlink-root':
                    alias=root/'alias';alias.symlink_to(target,target_is_directory=True);target=alias
                p.write_text(json.dumps(m))
                with self.assertRaises(ValueError):v.verify(bundle,target)


if __name__ == '__main__': unittest.main()
