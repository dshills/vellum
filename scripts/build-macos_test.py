"""Owned compiler pairing, sampled drift and descriptor-relative failure tests."""
import copy
import contextlib
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import struct
import subprocess
import tempfile
import sys
import unittest
from unittest import mock

from build_output import BuildOutput

spec = importlib.util.spec_from_file_location("build_macos",Path(__file__).with_name("build-macos.py"))
build = importlib.util.module_from_spec(spec)
spec.loader.exec_module(build)


class BuildChecks(unittest.TestCase):
    def run_fixture(self, base, mode="ok", hook=None):
        root=base/"source";root.mkdir(exist_ok=True)
        output=base/"Output parent"/"Result with spaces, 'quotes'"
        before={"schema":1,"runtime_executed":False,"distribution_audit":"pending","source":{"commit":"fixture","clean":True},"build_configuration":{"GOOS":"darwin","GOARCH":"arm64","CGO_ENABLED":"1"},"sdk_files":[{"file":"fixture.go","sha256":"original"}]}
        calls=[]
        def collect(root, **options):
            value=copy.deepcopy(before)
            if options.get("binary") is not None:
                binary,link_map=options['binary'],options['link_map']
                if mode=="output-drift":binary.write_bytes(binary.read_bytes()+b'changed')
                observation=build.inventory.native_link_observation(binary,link_map,{}, {"native_link_candidates":[]})
                value['native_link_observation']=observation
                if mode=="late-output-drift":binary.write_bytes(binary.read_bytes()+b'changed')
                if mode in {"source-drift","config-drift"}:
                    if mode=="source-drift":value['sdk_files'][0]['sha256']='changed'
                    else:value['build_configuration']['GOARCH']='amd64'
                if hook:hook(output)
            return value
        def compile(argv, **options):
            calls.append((argv,options))
            self.assertEqual(argv[1],"build")
            self.assertEqual(argv[-1],"./cmd/workbench")
            self.assertEqual(options['log_limit_bytes'],4*1024*1024)
            options['stdout'].write(b'compiler stdout\n');options['stderr'].write(b'compiler stderr\n')
            if mode=="timeout":raise subprocess.TimeoutExpired(argv,600)
            if mode=="nonzero":return subprocess.CompletedProcess(argv,7)
            binary=Path(argv[argv.index('-o')+1]);link_map=Path(argv[2].split('-Wl,-map,',1)[1])
            self.assertFalse(binary.exists());self.assertFalse(link_map.exists())
            binary.write_bytes(b'\xcf\xfa\xed\xfe'+struct.pack('<I',0x100000c)+b'not-executable-fixture')
            if mode!="missing-map":link_map.write_text('# Path: /intermediate/linker/output\n# Arch: '+('x86_64'if mode=='bad-arch'else'arm64')+'\n# Object files:\n[  0] linker synthesized\n# Sections:\n')
            result = subprocess.CompletedProcess(argv,0)
            result.log_summaries = {channel: {"bytes":len(data),"sha256":hashlib.sha256(data).hexdigest()} for channel,data in [('stdout',b'compiler stdout\n'),('stderr',b'compiler stderr\n')]}
            return result
        tool=(Path('/fixture/go'),{'file':'go','bytes':7,'sha256':'fixture'})
        identities=[tool,tool]if mode!='tool-drift'else[tool,(tool[0],{**tool[1],'sha256':'changed'})]
        env = build.child_environment()
        environments = [env, env if mode != 'env-drift' else {**env, 'GOMAXPROCS': 'changed'}]
        with mock.patch.object(build.platform,'system',return_value='Darwin'),mock.patch.object(build,'go_identity',side_effect=identities),mock.patch.object(build.inventory,'collect',side_effect=collect),mock.patch.object(build,'run_child',side_effect=compile),mock.patch.object(build, 'child_environment', side_effect=environments):
            report=build.build(root,output,include_untracked=True)
        return report,output,calls

    def test_success_binds_artifacts_and_keeps_collector_scope(self):
        with tempfile.TemporaryDirectory()as temporary:
            report,output,calls=self.run_fixture(Path(temporary))
            self.assertEqual(len(calls),1)
            self.assertEqual(report['event'],'vellum.macos.build_complete')
            self.assertFalse(report['runtime_executed'])
            receipt=json.loads((output/build.RECEIPT).read_text())
            self.assertEqual(report['receipt_sha256'],hashlib.sha256((output/build.RECEIPT).read_bytes()).hexdigest())
            self.assertEqual(receipt['artifact_pair_binding'],'producer-observed-same-build')
            inventory=json.loads((output/'inventory.json').read_text())
            self.assertEqual(inventory['native_link_observation']['artifact_pair_binding'],'caller-supplied-unverified')
            self.assertEqual(set(os.listdir(output)),set(receipt['artifacts'])|{build.RECEIPT})
            for name,value in receipt['artifacts'].items():
                self.assertEqual(value['sha256'],hashlib.sha256((output/name).read_bytes()).hexdigest())
            self.assertEqual((output/build.BINARY).stat().st_mode&0o777,0o755)
            self.assertEqual(receipt['command']['exit_code'],0)
            self.assertTrue(receipt['command']['outputs_initially_absent'])

    def test_failures_retain_diagnostics_without_success_receipt(self):
        cases={'nonzero':subprocess.CalledProcessError,'timeout':subprocess.TimeoutExpired,'missing-map':ValueError,'source-drift':ValueError,'config-drift':ValueError,'output-drift':ValueError,'late-output-drift':ValueError,'tool-drift':ValueError,'env-drift':ValueError,'bad-arch':ValueError}
        for mode,error in cases.items():
            with self.subTest(mode=mode),tempfile.TemporaryDirectory()as temporary:
                base=Path(temporary)
                with self.assertRaises(error):self.run_fixture(base,mode)
                output=base/"Output parent"/"Result with spaces, 'quotes'"
                self.assertFalse((output/build.RECEIPT).exists())
                failure=json.loads((output/'build-failed.json').read_text())
                self.assertEqual(failure['event'],'vellum.macos.build_failed')
                self.assertEqual((output/'build.stdout').read_bytes(),b'compiler stdout\n')

    def test_parent_and_output_replacement_never_write_into_replacements(self):
        for parent in [False,True]:
            with self.subTest(parent=parent),tempfile.TemporaryDirectory()as temporary:
                def replace(output):
                    victim=output.parent if parent else output
                    victim.rename(victim.with_name(victim.name+'-retired'))
                    output.mkdir(parents=True)
                    (output/'sentinel').write_bytes(b'keep')
                with self.assertRaisesRegex(ValueError,'replaced'):self.run_fixture(Path(temporary),hook=replace)
                output=Path(temporary)/"Output parent"/"Result with spaces, 'quotes'"
                self.assertEqual(list(output.iterdir()),[output/'sentinel'])
                self.assertEqual((output/'sentinel').read_bytes(),b'keep')

    def test_existing_inside_symlink_and_wrong_host_refuse_before_collection(self):
        with tempfile.TemporaryDirectory()as temporary:
            base=Path(temporary);root=base/'source';root.mkdir();existing=base/'existing';existing.mkdir()
            (base/'link').symlink_to(existing,target_is_directory=True)
            with mock.patch.object(build.platform,'system',return_value='Darwin'),mock.patch.object(build.inventory,'collect')as collect:
                for output in [existing,base/'link',root/'inside']:
                    with self.assertRaises(ValueError):build.build(root,output)
                collect.assert_not_called()
            with mock.patch.object(build.platform,'system',return_value='Linux'):
                with self.assertRaisesRegex(ValueError,'macOS host'):build.build(root,base/'new')
            self.assertFalse((base/'new').exists())

    def test_late_publication_failure_removes_owned_receipt(self):
        original=BuildOutput.write_json
        def partial(output,name,value):
            if name==build.RECEIPT:
                with output.create(name)as stream:stream.write(b'incomplete')
                raise OSError('injected receipt write failure')
            return original(output,name,value)
        with tempfile.TemporaryDirectory()as temporary,mock.patch.object(BuildOutput,'write_json',partial):
            with self.assertRaisesRegex(OSError,'injected receipt'):self.run_fixture(Path(temporary))
            output=Path(temporary)/"Output parent"/"Result with spaces, 'quotes'"
            self.assertFalse((output/build.RECEIPT).exists())
            self.assertTrue((output/'build-failed.json').is_file())

    def test_initialization_parent_redirection_refuses_before_output_creation(self):
        with tempfile.TemporaryDirectory() as temporary:
            base = Path(temporary)
            root, parent = base/'source', base/'parent'
            root.mkdir(); parent.mkdir()
            original = Path.mkdir
            def redirect(path, *args, **options):
                if path == parent:
                    parent.rmdir()
                    parent.symlink_to(root, target_is_directory=True)
                return original(path, *args, **options)
            with mock.patch.object(Path, 'mkdir', redirect):
                with self.assertRaisesRegex(ValueError, 'parent changed'):
                    BuildOutput(root, parent/'out', 100)
            self.assertEqual(list(root.iterdir()), [])

    def test_failed_receipt_retirement_reports_cleanup_and_preserves_original(self):
        original_write, original_remove = BuildOutput.write_json, BuildOutput.remove_created
        original = OSError('original receipt publication failed')
        def partial(output, name, value):
            if name == build.RECEIPT:
                with output.create(name) as stream: stream.write(b'incomplete')
                raise original
            return original_write(output, name, value)
        def refuse(output, name):
            if name == build.RECEIPT: raise PermissionError('injected receipt removal refusal')
            return original_remove(output, name)
        with tempfile.TemporaryDirectory() as temporary, mock.patch.object(BuildOutput, 'write_json', partial), mock.patch.object(BuildOutput, 'remove_created', refuse):
            with self.assertRaises(OSError) as failure: self.run_fixture(Path(temporary))
            self.assertIs(failure.exception, original)
            cleanup = original.build_cleanup_error
            self.assertEqual(cleanup['phase'], 'receipt-retirement')
            self.assertFalse(cleanup['owned_receipt_retirement_confirmed'])
            self.assertTrue(cleanup['output_completion_unconfirmed'])
            self.assertEqual(cleanup['type'], 'PermissionError')
        stderr = io.StringIO()
        with mock.patch.object(build, 'build', side_effect=original), mock.patch.object(sys, 'argv', ['build-macos.py', '--output', '/unused']), contextlib.redirect_stderr(stderr):
            self.assertEqual(build.main(), 1)
        self.assertEqual(json.loads(stderr.getvalue())['error']['cleanup_error'], cleanup)

    def test_replaced_or_edited_metadata_never_becomes_receipt_truth(self):
        for name in ['inputs-before.json','build.stdout']:
            for replace in [False,True]:
                with self.subTest(name=name,replace=replace),tempfile.TemporaryDirectory()as temporary:
                    def corrupt(output):
                        target=output/name
                        if replace:target.unlink()
                        target.write_bytes(b'forged metadata')
                    with self.assertRaises(ValueError):self.run_fixture(Path(temporary),hook=corrupt)
                    output=Path(temporary)/"Output parent"/"Result with spaces, 'quotes'"
                    self.assertFalse((output/build.RECEIPT).exists())

    def test_published_binary_drift_never_commits(self):
        original=BuildOutput.copy
        def corrupt(output,source,name,mode=0o644):
            original(output,source,name,mode)
            if name==build.BINARY:(output.path/name).write_bytes(b'changed binary')
        with tempfile.TemporaryDirectory()as temporary,mock.patch.object(BuildOutput,'copy',corrupt):
            with self.assertRaisesRegex(ValueError,'artifact, JSON or log'):self.run_fixture(Path(temporary))
            output=Path(temporary)/"Output parent"/"Result with spaces, 'quotes'"
            self.assertFalse((output/build.RECEIPT).exists())

    def test_altered_completion_receipt_refuses_and_is_retired(self):
        original = BuildOutput.write_json
        def corrupt(output, name, value):
            original(output, name, value)
            if name == build.RECEIPT:
                (output.path/name).write_bytes(b'forged receipt')
        with tempfile.TemporaryDirectory() as temporary, mock.patch.object(BuildOutput, 'write_json', corrupt):
            with self.assertRaisesRegex(ValueError, 'receipt differs'):
                self.run_fixture(Path(temporary))
            output=Path(temporary)/"Output parent"/"Result with spaces, 'quotes'"
            self.assertFalse((output/build.RECEIPT).exists())

    def test_output_bounds_symlink_fifo_and_duplicate_preservation(self):
        with tempfile.TemporaryDirectory()as temporary:
            base=Path(temporary);root=base/'source';root.mkdir();output=BuildOutput(root,base/'out',32)
            try:
                output.write_json('keep.json',{'value':'keep'})
                before=(output.path/'keep.json').read_bytes()
                with self.assertRaises(FileExistsError):output.write_json('keep.json',{'value':'other'})
                self.assertEqual((output.path/'keep.json').read_bytes(),before)
                (output.path/'keep.json').rename(output.path/'retired.json')
                (output.path/'keep.json').write_bytes(before)
                with self.assertRaisesRegex(ValueError, 'file was replaced'):
                    output.identity('keep.json')
                source=base/'large';source.write_bytes(b'x'*33)
                with self.assertRaisesRegex(ValueError,'file bound'):output.copy(source,'large')
                self.assertFalse((output.path/'large').exists())
                (output.path/'alias').symlink_to(source)
                with self.assertRaises(OSError):output.identity('alias')
                os.mkfifo(output.path/'fifo')
                with self.assertRaisesRegex(ValueError,'regular'):output.identity('fifo')
                for name in ['../outside','/outside','']:
                    with self.assertRaises(ValueError):output.identity(name)
            finally:output.close()


if __name__=='__main__':unittest.main()
