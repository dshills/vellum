import json
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest
from unittest import mock
import native_toolchain as n


class NativeInputs(unittest.TestCase):
    def exercise(self,root,config):
        tools=root/'tools';tools.mkdir();sdk=root/'MacOSX.sdk';sdk.mkdir()
        (sdk/'SDKSettings.json').write_text('{"Version":"27.0"}')
        for name in ('clang','clang++','ld'):(tools/name).write_text(name)
        def output(args,**kw):
            self.assertEqual(kw['timeout'],20)
            if args[0]=='xcrun':
                if args[-1]=='--show-sdk-path':return str(sdk)
                if args[-1]=='--show-sdk-version':return '27.0'
                if '--find' in args:return str(tools/args[-1])
            return 'synthetic tool version'
        probes=[]
        def probe(args,**kw):
            self.assertIn('-###',args);self.assertEqual(args[args.index('-o')+1],'/dev/null')
            self.assertEqual(kw['timeout'],20);probes.append(args)
            return SimpleNamespace(stdout='',stderr='"clang" "-triple" "arm64-apple-macosx27.0.0"')
        with mock.patch.object(n,'check_child_output',side_effect=output),mock.patch.object(n.shutil,'which',side_effect=lambda name,**k:str(tools/name)),mock.patch.object(n,'run_child',side_effect=probe):
            result=n.capture(config,{'PATH':str(tools)})
        return result,probes

    def test_records_selected_tools_sdk_and_nonexecuting_probes(self):
        with tempfile.TemporaryDirectory() as t:
            config=dict(CC='clang',CXX='clang++',CGO_CFLAGS='-O2 -g',CGO_CPPFLAGS='',CGO_CXXFLAGS='-O2 -g',CGO_LDFLAGS='-O2 -g')
            r,probes=self.exercise(Path(t),config)
            self.assertEqual(r['compiler_default_minimum'],'27.0.0')
            self.assertEqual(r['cgo'],config)
            self.assertEqual(r['overrides'],dict(SDKROOT=None,MACOSX_DEPLOYMENT_TARGET=None,DEVELOPER_DIR=None))
            self.assertEqual(set(r['binaries']),{'cc_driver','cc','cxx_driver','cxx','linker'})
            self.assertEqual(len(probes),2);self.assertIn('-c',probes[0]);self.assertNotIn('-c',probes[1])
            old=n.identity(Path(r['sdk']['settings']['path']));Path(old['path']).write_text('changed')
            self.assertNotEqual(n.identity(Path(old['path'])),old)

    def test_wrapped_compiler_is_refused_instead_of_misattributed(self):
        with tempfile.TemporaryDirectory() as t:
            with self.assertRaisesRegex(ValueError,'single compiler'):self.exercise(Path(t),{'CC':'ccache clang'})


if __name__=='__main__':unittest.main()
