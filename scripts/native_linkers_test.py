"""Owned SDK linker source, target selection, attribution and tamper regressions."""
import copy
import importlib.util
import json
import os
from pathlib import Path
import shutil
import sys
import tempfile
import unittest

import native_gpu as gpu
import native_linkers as linkers
from child_process import run_child

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('owned_linkers_sdk_fixture', HERE / 'native_gpu_test.py')
fixture = importlib.util.module_from_spec(spec); spec.loader.exec_module(fixture)


class OwnedLinkers(unittest.TestCase):
    def generated(self, root):
        options = fixture.inputs(root)
        return gpu.candidate(*options.values())

    def write(self, base, files):
        for name, body in files.items():
            p = base / name
            p.parent.mkdir(parents=True, exist_ok=True)
            p.write_bytes(body)

    def test_all_five_independently_authored_sources_and_license(self):
        expected = {'android', 'darwin', 'ios', 'linux', 'windows'}
        self.assertEqual({m.rsplit('-', 1)[1] for m in linkers.sources()}, expected)
        self.assertEqual(linkers.LICENSE, (HERE.parent / 'LICENSE').read_bytes())
        with tempfile.TemporaryDirectory() as temporary:
            record, files = self.generated(Path(temporary))
            self.assertEqual(record['owned_linkers'], linkers.declaration())
            self.assertEqual(record['distribution_audit'], 'pending')
            self.assertEqual(files[gpu.DIRECTORY + '/' + gpu.LIBRARY], (Path(temporary) / 'candidate.a').read_bytes())
            for module, bodies in linkers.sources().items():
                prefix = linkers.DIRECTORIES[module] + '/'
                for name, body in bodies.items():
                    self.assertEqual(files[prefix + name], body)
                self.assertNotIn('libs.go', bodies)
            self.assertIn(b'${SRCDIR}/arm64/libwgpu_native.a', linkers.sources()[gpu.MODULE]['link.go'])
            self.assertNotIn(b'-lwgpu_native', linkers.sources()[gpu.MODULE]['link.go'])

    def test_every_conflicting_main_module_replacement_is_refused(self):
        data = b'module example.com/native\nrequire github.com/dshills/vellum/drivers/cogent v0.0.0\n'
        for module in linkers.DIRECTORIES:
            for block in [False, True]:
                with self.subTest(module=module, block=block):
                    conflict = ('\nreplace (\n ' + module + ' => ./elsewhere\n)\n' if block else '\nreplace ' + module + ' v0.0.0 => ./elsewhere\n').encode()
                    with self.assertRaisesRegex(ValueError, 'already has'):
                        gpu.overlay_modfile('examples/example/go.mod', data + conflict)
        output = gpu.overlay_modfile('examples/example/go.mod', data)
        for module in linkers.DIRECTORIES:
            self.assertIn(('replace ' + module + ' => ../../drivers/cogent/build/libs-').encode(), output)
        self.assertIsNone(gpu.overlay_modfile('examples/automation/go.mod', b'module example.com/portable\n'))

    def test_real_go_package_selection_and_named_unsupported_failures(self):
        with tempfile.TemporaryDirectory(prefix='Vellum linker target matrix ') as temporary:
            base = Path(temporary)
            for module, bodies in linkers.sources().items():
                self.write(base / module.rsplit('-', 1)[1], bodies)
            environment = fixture.release.child_environment(); environment.update(GOPROXY='off', GOSUMDB='off')
            for module in linkers.DIRECTORIES:
                target = module.rsplit('-', 1)[1]
                env = dict(environment, GOOS='darwin', GOARCH='arm64', CGO_ENABLED='1')
                listed = run_child(['go', 'list', '-json', '.'], cwd=base / target, env=env, capture_output=True, text=True, timeout=30, check=True)
                info = json.loads(listed.stdout)
                self.assertEqual(info.get('GoFiles'), ['package.go'])
                self.assertEqual(info.get('CgoFiles', []), ['link.go'] if target == 'darwin' else [])
            # Literal expected matrix, including overlapping Android/Linux and iOS/Darwin tags.
            for module, goos, goarch, cgo in [('darwin', 'darwin', 'amd64', '1'), ('darwin', 'darwin', 'arm64', '0'), ('linux', 'linux', 'arm64', '0'), ('android', 'android', 'arm64', '0'), ('ios', 'ios', 'arm64', '0'), ('windows', 'windows', 'arm64', '0')]:
                with self.subTest(module=module, goos=goos, goarch=goarch, cgo=cgo):
                    env = dict(environment, GOOS=goos, GOARCH=goarch, CGO_ENABLED=cgo)
                    listed = run_child(['go', 'list', '-json', '.'], cwd=base / module, env=env, capture_output=True, text=True, timeout=30, check=True)
                    self.assertIn('unsupported.go', json.loads(listed.stdout)['GoFiles'])
                    built = run_child(['go', 'build', '.'], cwd=base / module, env=env, capture_output=True, text=True, timeout=30)
                    self.assertNotEqual(built.returncode, 0)
                    self.assertIn('undefined: VellumSDKRequiresDarwinARM64WithCGO', built.stderr)

    def test_inventory_uses_actual_owned_graph_and_keeps_original_resolution_separate(self):
        with tempfile.TemporaryDirectory(prefix='Vellum owned graph ') as temporary:
            base = Path(temporary); root, files = fixture.source(base)
            # Unlike upstream requirements, these sources must resolve without a proxy or upstream cache.
            before = {n: p.read_bytes() for n, p in files.items()}
            record, generated = self.generated(base)
            with linkers.inventory_tree(record, generated) as context:
                env = fixture.release.child_environment(); env.update(GOPROXY='off', GOSUMDB='off')
                from unittest import mock
                with mock.patch.object(fixture.release, 'child_environment', return_value=env):
                    inventory, notices = fixture.release.license_inventory(root, linkers=context)
                selected = {m['module']: m for m in inventory['modules']}
                self.assertEqual(set(selected), set(linkers.DIRECTORIES))
                self.assertEqual(notices, {})
                for module, item in selected.items():
                    self.assertIsNone(item['version']); self.assertIsNone(item['sum'])
                    self.assertEqual(item['selected_requirement']['version'], 'v0.0.0')
                    self.assertEqual(item['effective_source']['directory'], linkers.DIRECTORIES[module])
                    self.assertEqual(item['module_license_files'], ['LICENSE'])
                    self.assertEqual(item['status'], 'owned-source-declaration')
                    self.assertEqual(item['license_files'][0]['sha256'], linkers.checksum(linkers.LICENSE))
                self.assertEqual(inventory['distribution_audit'], 'pending')
            self.assertEqual(before, {n: p.read_bytes() for n, p in files.items()})

    def test_unrelated_missing_declaration_stays_unresolved_and_owned_bytes_cannot_change(self):
        with tempfile.TemporaryDirectory() as temporary:
            base = Path(temporary); root, _ = fixture.source(base)
            record, generated = self.generated(base); other = base / 'unlicensed'; other.mkdir()
            with linkers.inventory_tree(record, generated) as context:
                modules = [{'Path': module, 'Version': 'v0.0.0', 'Replace': {'Dir': str(directory)}} for module, directory in context['directories'].items()]
                modules.append({'Path': 'example.com/unlicensed', 'Version': 'v1.0.0', 'Dir': str(other)})
                with self.assertRaisesRegex(ValueError, 'missing an owned linker'):
                    fixture.release.license_inventory(root, modules=modules[1:], linkers=context)
                inventory, _ = fixture.release.license_inventory(root, modules=modules, linkers=context)
                unrelated = next(item for item in inventory['modules'] if item['module'] == 'example.com/unlicensed')
                self.assertEqual(unrelated['status'], 'manual-review-required')
                self.assertEqual(unrelated['module_license_files'], [])
                with self.assertRaisesRegex(ValueError, 'bound'):
                    fixture.release.license_inventory(root, modules=modules, linkers=context, max_notice_bytes=1)
                directory = context['directories'][gpu.MODULE]
                (directory / 'LICENSE').write_bytes(b'other license')
                with self.assertRaisesRegex(ValueError, 'bytes changed'):
                    fixture.release.license_inventory(root, modules=modules, linkers=context)

    def test_exact_original_checksum_is_resolution_metadata_only(self):
        with tempfile.TemporaryDirectory() as temporary:
            base = Path(temporary);root, _ = fixture.source(base);record, generated = self.generated(base)
            (root / 'drivers/cogent/go.sum').write_text(gpu.MODULE + ' v0.0.0 h1:original\n')
            with linkers.inventory_tree(record, generated) as context:
                module = {'Path': gpu.MODULE, 'Version': 'v0.0.0', 'Replace': {'Dir': str(context['directories'][gpu.MODULE])}}
                item = linkers.inventory_item(root, module, context, copy_notices=True, max_notice_bytes=None)
                self.assertEqual(item['selected_requirement']['sum'], 'h1:original')
                self.assertIsNone(item['sum'])
                module['Replace']['Sum'] = 'h1:original'
                with self.assertRaisesRegex(ValueError, 'does not select'):
                    linkers.inventory_item(root, module, context, copy_notices=True, max_notice_bytes=None)

    def test_all_capsules_are_manifest_bound_and_every_origin_is_required(self):
        with tempfile.TemporaryDirectory() as temporary:
            base = Path(temporary); root, files = fixture.source(base)
            options = fixture.inputs(base)
            packed = fixture.package(root, files, base / 'sdk.tar.gz', **options)
            verified = fixture.checker.verify_archive(base / 'sdk.tar.gz', packed['sha256'])
            sdk = fixture.checker.extract_verified(base / 'sdk.tar.gz', base / 'relocated SDK', verified)
            record = gpu.sdk_candidate(sdk)
            modules = [{'Path': name, 'Replace': {'Dir': str(sdk / relative)}} for name, relative in {
                'github.com/dshills/vellum': '.', 'github.com/dshills/vellum/drivers/cogent': 'drivers/cogent',
                'cogentcore.org/core': 'drivers/cogent/build/upstream', 'github.com/go-text/typesetting': 'drivers/cogent/build/typesetting', **linkers.DIRECTORIES}.items()]
            fixture.checker.check_origin(modules, sdk, gpu=record)
            for module in linkers.DIRECTORIES:
                with self.subTest(module=module):
                    with self.assertRaisesRegex(ValueError, 'missing explicit'):
                        fixture.checker.check_origin([m for m in modules if m['Path'] != module], sdk, gpu=record)
                    path = sdk / linkers.DIRECTORIES[module] / 'LICENSE'; original = path.read_bytes();path.write_bytes(b'changed')
                    with self.assertRaisesRegex(ValueError, 'checksum mismatch'):
                        gpu.sdk_candidate(sdk)
                    path.write_bytes(original)
                    extra = path.parent / 'unexpected';extra.mkdir()
                    with self.assertRaisesRegex(ValueError, 'extra or missing'):
                        gpu.sdk_candidate(sdk)
                    extra.rmdir()
                    original_manifest = (sdk / 'SDK_MANIFEST.json').read_bytes();m = json.loads(original_manifest)
                    del m['native_gpu']['owned_linkers']['modules'][module]
                    (sdk / 'SDK_MANIFEST.json').write_text(json.dumps(m))
                    with self.assertRaisesRegex(ValueError, 'source declaration'):
                        gpu.sdk_candidate(sdk)
                    (sdk / 'SDK_MANIFEST.json').write_bytes(original_manifest)

    def test_legacy_sdk_stays_legacy_without_claiming_new_ownership(self):
        with tempfile.TemporaryDirectory() as temporary:
            base = Path(temporary); record, files = self.generated(base)
            record.pop('owned_linkers')
            old = {'go.mod': gpu.GO_MOD, 'libs.go': gpu.LIBS_GO,
                   gpu.LIBRARY: files[gpu.DIRECTORY + '/' + gpu.LIBRARY],
                   gpu.PROVENANCE: files[gpu.DIRECTORY + '/' + gpu.PROVENANCE]}
            record['files'] = {name: gpu.digest(body) for name, body in old.items()}
            hashes = {gpu.DIRECTORY + '/' + name: gpu.digest(body) for name, body in old.items()}
            hashes['drivers/cogent/go.mod'] = 'a' * 64
            record['go_mod_overlays'] = {'drivers/cogent/go.mod': {'source_sha256': 'b' * 64, 'archived_sha256': 'a' * 64}}
            gpu.validate_record(record, hashes, True)
            self.assertEqual(linkers.replacements(record), {gpu.MODULE: gpu.DIRECTORY})
            self.assertIsNone(linkers.sdk_inventory(base, record))
            hashes['drivers/cogent/build/libs-ios/LICENSE'] = gpu.digest(linkers.LICENSE)
            with self.assertRaisesRegex(ValueError, 'coverage'):
                gpu.validate_record(record, hashes, True)


if __name__ == '__main__':
    unittest.main()
