"""Static metadata fixtures; never load or execute an application."""
import importlib.util
from pathlib import Path
import hashlib
import json
import plistlib
import tempfile
import unittest
from unittest import mock

spec=importlib.util.spec_from_file_location('linkage',Path(__file__).with_name('check-private-macos-linkage.py'))
m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)

BASE='''fixture:
Load command 0
 cmd LC_LOAD_DYLIB
 cmdsize 56
 name /usr/lib/libSystem.B.dylib (offset 24)
Load command 1
 cmd LC_LOAD_DYLINKER
 cmdsize 32
 name /usr/lib/dyld (offset 12)
Load command 2
 cmd LC_BUILD_VERSION
 cmdsize 32
 platform 1
 minos 27.0
 sdk 27.0
 ntools 1
 tool 3
 version 27037.1
'''


class Linkage(unittest.TestCase):
    def test_current_host_metadata_is_not_old_host_qualification(self):
        result=m.parse_commands(BASE,'27.2')
        self.assertEqual(result['deployment']['minimum'],'27.0')
        self.assertEqual(result['rpaths'],[])
        with self.assertRaisesRegex(ValueError,'exceeds'):m.parse_commands(BASE,'26.6.2')

    def test_all_non_system_dependency_forms_refused(self):
        for kind in m.LOADS:
            for name in ['/private/tmp/build/libX.dylib','/opt/homebrew/lib/libX.dylib',
                         '@rpath/libX.dylib','@loader_path/libX.dylib','libX.dylib',
                         '/usr/lib/../../private/tmp/x','/System/Library/../x','/usr/lib//x']:
                with self.subTest(kind=kind,name=name),self.assertRaisesRegex(ValueError,'dependency'):
                    m.parse_commands(BASE.replace('LC_LOAD_DYLIB',kind).replace('/usr/lib/libSystem.B.dylib',name),'27.2')

    def test_unsafe_loader_rpaths_and_embedded_environment_refused(self):
        for output in [BASE.replace('/usr/lib/dyld','/private/tmp/dyld'),
                       BASE+'Load command 3\n cmd LC_RPATH\n path /private/tmp/build (offset 12)\n',
                       BASE+'Load command 3\n cmd LC_DYLD_ENVIRONMENT\n name DYLD_INSERT_LIBRARIES=x (offset 12)\n']:
            with self.subTest(output=output),self.assertRaises(ValueError):m.parse_commands(output,'27.2')

    def test_missing_duplicate_and_foreign_deployment_refused(self):
        for output in [BASE.replace('LC_BUILD_VERSION','LC_UUID'),BASE+BASE[BASE.index('Load command 2'):],
                       BASE.replace('platform 1','platform 2'),BASE.replace('minos 27.0','minos banana'),
                       BASE.replace('minos 27.0','minos 28.0'),BASE.replace('sdk 27.0','sdk 26.0'),
                       BASE.replace('minos 27.0','minos 27.0\n minos 26.0')]:
            with self.subTest(output=output),self.assertRaises(ValueError):m.parse_commands(output,'27.2')

    def test_legacy_macos_minimum_is_parsed_explicitly(self):
        output=BASE[:BASE.index('Load command 2')]+'''Load command 2
 cmd LC_VERSION_MIN_MACOSX
 cmdsize 16
 version 13.0
 sdk 27.0
'''
        self.assertEqual(m.parse_commands(output,'27.2')['deployment']['minimum'],'13.0')
        for value in ['', '27.2x','1.2.3.4','-1','9999.0']:
            with self.assertRaises(ValueError):m.version(value)


    def test_installed_orchestration_is_pinned_and_uses_only_static_tools(self):
        with tempfile.TemporaryDirectory() as temporary,mock.patch.object(m.sys,'platform','darwin'),mock.patch.object(m.platform,'mac_ver',return_value=('27.2',(),'')):
            root=Path(temporary).resolve()
            receipt=dict(source_commit='a'*40,build_version='24',apps=list(m.v.IDENTIFIERS))
            raw=json.dumps(receipt).encode();(root/'INSTALLATION.json').write_bytes(raw)
            for name,identifier in m.v.IDENTIFIERS.items():
                contents=root/'apps'/(name+'.app')/'Contents';(contents/'MacOS').mkdir(parents=True)
                (contents/'MacOS'/name).write_bytes(b'\xcf\xfa\xed\xfe' + b'fictional executable')
                (contents/'Resources').mkdir();(contents/'Resources/APP_MANIFEST.json').write_text('{}')
                (contents/'Info.plist').write_bytes(plistlib.dumps(dict(CFBundleExecutable=name,CFBundleIdentifier=identifier,CFBundleVersion='24')))
            calls=[]
            def static_tool(*args):
                calls.append(args)
                self.assertTrue(Path(args[-1]).is_relative_to(root/'apps'))
                if args[:2]==('lipo','-archs'):return 'arm64\n'
                if args[:4]==('otool','-arch','arm64','-hv'):return 'MH_MAGIC_64 ARM64 ALL 0x00 EXECUTE 3 128 NOUNDEFS\n'
                if args[:4]==('otool','-arch','arm64','-l'):return BASE
                self.fail('unexpected tool invocation '+str(args))
            with mock.patch.object(m,'tool',side_effect=static_tool):
                report=m.inspect(root,hashlib.sha256(raw).hexdigest())
                self.assertEqual(len(report['apps']),5)
                self.assertEqual(report['native_runtime'],'not executed')
                self.assertEqual(len(calls),15)
                calls.clear()
                with self.assertRaisesRegex(ValueError,'pin mismatch'):m.inspect(root,'0'*64)
                self.assertEqual(calls,[])
                (root/'apps/Vellum Forge.app/Contents/extra').write_bytes(b'\xcf\xfa\xed\xfe')
                with self.assertRaisesRegex(ValueError,'embedded Mach-O'):m.inspect(root,hashlib.sha256(raw).hexdigest())


if __name__=='__main__':unittest.main()
