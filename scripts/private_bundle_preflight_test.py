import hashlib
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import private_bundle_preflight as p


class Preflight(unittest.TestCase):
    def fixture(self, directory):
        base = Path(directory).resolve()
        root = base/'source'; root.mkdir()
        (root/'go.mod').write_text('module example.com/test\n\ngo 1.25.6\n')
        driver = root/'drivers/cogent'; (driver/'native').mkdir(parents=True)
        deps = []
        for name, (module, files) in p.SENTINELS.items():
            deps.append(dict(name=name, module=module))
            for relative in files:
                path = driver/'build'/name/relative
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text('module '+module+'\n' if relative=='go.mod' else 'prepared source')
        (driver/'native/dependencies.json').write_text(json.dumps(dict(dependencies=deps)))
        tools = base/'tools'; tools.mkdir()
        for name in ('go','compile','link'): (tools/name).write_bytes(name.encode())
        config = dict(GOVERSION='go1.25.6', GOOS='darwin', GOARCH='arm64', CGO_ENABLED='1',
                      GOROOT=str(base/'go-root'), GOTOOLDIR=str(tools))
        return base, root, tools, config

    def invoke(self, root, tools, config, output, install=None):
        with mock.patch.object(p.shutil,'which',return_value=str(tools/'go')), \
             mock.patch.object(p,'check_child_output',return_value=json.dumps(config)) as child:
            result = p.preflight(root, output, {'PATH':str(tools)}, install)
            self.assertEqual(child.call_args.args[0][:3], [str(tools/'go'),'env','-json'])
            self.assertEqual(child.call_count,1)
            return result

    def test_success_records_exact_go_tools_without_creating_output(self):
        with tempfile.TemporaryDirectory() as d:
            base,root,tools,config=self.fixture(d)
            result=self.invoke(root,tools,config,base/'result',base/'installed')
            self.assertFalse((base/'result').exists())
            self.assertFalse((base/'installed').exists())
            self.assertEqual(result['toolchain']['configuration'],config)
            self.assertEqual(result['toolchain']['binaries']['compile']['sha256'],hashlib.sha256(b'compile').hexdigest())
            p.verify_snapshot(root,result)
            (tools/'compile').write_bytes(b'changed')
            with self.assertRaisesRegex(ValueError,'toolchain changed'):p.verify_snapshot(root,result)

    def test_wrong_version_target_and_cgo_fail_before_output(self):
        for key,value in [('GOVERSION','go1.27.3'),('GOOS','linux'),('GOARCH','amd64'),('CGO_ENABLED','0')]:
            with self.subTest(key=key),tempfile.TemporaryDirectory() as d:
                base,root,tools,config=self.fixture(d);config[key]=value
                with self.assertRaisesRegex(ValueError,'requires'):
                    self.invoke(root,tools,config,base/'result')
                self.assertFalse((base/'result').exists())

    def test_missing_and_active_native_preparation_fail_early(self):
        for failure in ('missing','active','module'):
            with self.subTest(failure=failure),tempfile.TemporaryDirectory() as d:
                base,root,tools,config=self.fixture(d);driver=root/'drivers/cogent'
                if failure=='missing':(driver/'build/upstream/core/scene.go').unlink()
                if failure=='active':(driver/'build/.prepare-lock').mkdir()
                if failure=='module':(driver/'build/typesetting/go.mod').write_text('module wrong\n')
                with self.assertRaises(ValueError):self.invoke(root,tools,config,base/'result')
                self.assertFalse((base/'result').exists())

    def test_sampled_native_drift_is_not_accepted_after_build(self):
        with tempfile.TemporaryDirectory() as d:
            base,root,tools,config=self.fixture(d)
            result=self.invoke(root,tools,config,base/'result')
            (root/'drivers/cogent/build/upstream/core/scene.go').write_text('changed')
            with self.assertRaisesRegex(ValueError,'native preparation changed'):p.verify_snapshot(root,result)

    def test_synced_alias_existing_and_checkout_output_refused_before_go(self):
        with tempfile.TemporaryDirectory() as d:
            base,root,tools,config=self.fixture(d);home=base/'home';(home/'Documents').mkdir(parents=True)
            (base/'alias').symlink_to(home/'Documents',target_is_directory=True)
            existing=base/'existing';existing.mkdir();(existing/'keep').write_text('preserved')
            with mock.patch.object(p.Path,'home',return_value=home),mock.patch.object(p,'check_child_output') as child:
                for output in [home/'Documents/out',base/'alias/out',root/'build',existing]:
                    with self.subTest(output=output),self.assertRaises(ValueError):p.preflight(root,output,{'PATH':str(tools)})
                with self.assertRaisesRegex(ValueError,'nonsynced'):p.preflight(root,base/'new',{'PATH':str(tools)},home/'Documents/install')
                child.assert_not_called()
            self.assertEqual((existing/'keep').read_text(),'preserved')


if __name__=='__main__':unittest.main()
