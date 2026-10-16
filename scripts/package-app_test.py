"""Private app layout/overwrite/dependency gates without signing or a GUI."""
import importlib.util
import errno
import os
import subprocess
import json
import plistlib
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest import mock
import zipfile
spec=importlib.util.spec_from_file_location("package_app",Path(__file__).with_name("package-app.py"));app=importlib.util.module_from_spec(spec);spec.loader.exec_module(app)
class AppChecks(unittest.TestCase):
 def test_windows_and_linux_layouts(self):
  for target,magic in [("windows",b"MZ\x00\x00"),("linux",b"\x7fELF")]:
   with self.subTest(target=target),tempfile.TemporaryDirectory() as tmp:
    root=Path(tmp);binary=root/"binary";binary.write_bytes((b"MZ"+b"\x00"*58+(64).to_bytes(4,"little")+b"PE\x00\x00") if target=="windows" else magic+b"fixture");out=root/"app.archive"
    result=app.package(binary,out,"Test App","org.example.app","0.1.0",target)
    self.assertTrue(result["private_build"]);self.assertFalse(result["notarized"])
    if target=="windows":
     with zipfile.ZipFile(out) as archive:
      self.assertIn("Test App/Test App.exe",archive.namelist());manifest=json.loads(archive.read("Test App/APP_MANIFEST.json"))
    else:
     with tarfile.open(out) as archive:
      self.assertEqual(archive.getmember("Test App/Test App").mode,0o755);manifest=json.loads(archive.extractfile("Test App/APP_MANIFEST.json").read())
    self.assertEqual(manifest["target_runtime"],"unverified")
    with self.assertRaisesRegex(ValueError,"exists"):app.package(binary,out,"Test App","org.example.app","0.1.0",target)
 def test_mac_bundle_is_private_and_unsigned_by_default(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"),mock.patch.object(app,"run_child") as execute:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");out=root/"Test.app"
   app.package(binary,out,"Test","org.example.test","0.1.0","darwin")
   self.assertTrue((out/"Contents/Info.plist").is_file());self.assertTrue((out/"Contents/MacOS/Test").is_file());execute.assert_not_called()
   info=plistlib.loads((out/"Contents/Info.plist").read_bytes());self.assertEqual(info["CFBundleShortVersionString"],"0.1.0");self.assertEqual(info["CFBundleVersion"],"1")
   manifest=json.loads((out/"Contents/Resources/APP_MANIFEST.json").read_text());self.assertEqual(manifest["sdk_release"],"v0.1.0-alpha.2");self.assertEqual(manifest["build_version"],"1")
 def test_bundle_versions_are_distinct_and_numeric(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"),mock.patch.object(app,"run_child"):
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture")
   app.package(binary,root/"Good.app","Good","org.example.good","0.1.0","darwin",build_version="42.3")
   info=plistlib.loads((root/"Good.app/Contents/Info.plist").read_bytes());self.assertEqual(info["CFBundleVersion"],"42.3");self.assertEqual(info["CFBundleShortVersionString"],"0.1.0")
   with self.assertRaisesRegex(ValueError,"numeric major.minor.patch"):app.package(binary,root/"Bad.app","Bad","org.example.bad","v0.1.0-alpha.1","darwin")
   for value in ("v1","1.2.3.4","1-alpha",""):
    with self.subTest(build=value),self.assertRaisesRegex(ValueError,"build version"):app.package(binary,root/"Bad.app","Bad","org.example.bad","0.1.0","darwin",build_version=value)
 def test_universal_headers_are_not_dependencies(self):
  output="universal (architecture arm64):\n\t/System/Library/Frameworks/AppKit.framework/Versions/C/AppKit (compatibility version 1)\nuniversal (architecture x86_64):\n\t/usr/lib/libSystem.B.dylib (compatibility version 1)\n"
  with mock.patch.object(app,"check_child_output",return_value=output) as execute:
   app.darwin_dependencies(Path("universal"))
   self.assertEqual(execute.call_args.kwargs["timeout"],app.NATIVE_TOOL_TIMEOUT_SECONDS)
   self.assertEqual(execute.call_args.kwargs["env"],app.child_environment())
  with mock.patch.object(app,"check_child_output",return_value=output+"\t@rpath/libmissing.dylib (compatibility version 1)\n"):
   with self.assertRaisesRegex(ValueError,"audited bundle closure"):app.darwin_dependencies(Path("universal"))
 def test_archive_on_filesystem_without_hardlinks(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\x7fELFfixture")
   with mock.patch.object(app.os,"link",side_effect=OSError(errno.ENOTSUP,"no hardlinks")):
    app.package(binary,root/"app.tar.gz","App","org.example.app","0.1.0","linux")
   with tarfile.open(root/"app.tar.gz") as archive:self.assertTrue(archive.getmember("App/APP_MANIFEST.json"))
 def test_archives_normalize_umask_modes_and_builder_metadata(self):
  for target in ("linux","windows"):
   with self.subTest(target=target),tempfile.TemporaryDirectory() as tmp:
    root=Path(tmp);binary=root/"binary";binary.write_bytes((b"MZ"+b"\x00"*58+(64).to_bytes(4,"little")+b"PE\x00\x00") if target=="windows" else b"\x7fELFfixture")
    original=os.umask(0o077)
    try:
     first=root/"first";app.package(binary,first,"App","org.example.app","0.1.0",target)
     os.umask(0o022);second=root/"second";app.package(binary,second,"App","org.example.app","0.1.0",target)
    finally:os.umask(original)
    self.assertEqual(first.read_bytes(),second.read_bytes())
    if target=="linux":
     with tarfile.open(first) as archive:
      for member in archive:
       self.assertEqual((member.uid,member.gid,member.uname,member.gname,member.mtime),(0,0,"","",0))
       self.assertEqual(member.mode,0o755 if member.isdir() or member.name=="App/App" else 0o644)
    else:
     with zipfile.ZipFile(first) as archive:
      for member in archive.infolist():
       self.assertEqual(member.date_time,(1980,1,1,0,0,0))
       self.assertEqual((member.external_attr>>16)&0o777,0o755 if member.is_dir() or member.filename.endswith(".exe") else 0o644)
 def test_explicit_signing_verifies_final_bundle_and_requires_app_suffix(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"),mock.patch.object(app,"run_child") as execute:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");output=root/"Signed.app"
   app.package(binary,output,"Signed","org.example.signed","0.1.0","darwin","-")
   self.assertEqual(execute.call_args.args[0],["codesign","--verify","--strict",str(output)])
   self.assertEqual(execute.call_count,3)
   for call in execute.call_args_list:
    self.assertEqual(call.kwargs["timeout"],app.NATIVE_TOOL_TIMEOUT_SECONDS)
    self.assertEqual(call.kwargs["env"],app.child_environment())
   with self.assertRaisesRegex(ValueError,".app suffix"):app.package(binary,root/"invalid","Signed","org.example.signed","0.1.0","darwin")
 def test_failed_final_signature_verification_removes_owned_bundle(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");output=root/"Signed.app"
   invocations=[]
   final_verify=["codesign","--verify","--strict",str(output)]
   def execute(arguments,**kwargs):
    invocations.append(arguments)
    if arguments==final_verify:raise subprocess.CalledProcessError(1,arguments)
   with mock.patch.object(app,"run_child",side_effect=execute),self.assertRaises(subprocess.CalledProcessError):app.package(binary,output,"Signed","org.example.signed","0.1.0","darwin","-")
   self.assertEqual(len(invocations),3)
   self.assertIn("--sign",invocations[0]);self.assertNotEqual(invocations[0][-1],str(output))
   self.assertEqual(invocations[1],["codesign","--verify","--strict",invocations[0][-1]])
   self.assertEqual(invocations[2],final_verify)
   self.assertFalse(output.exists());self.assertTrue(binary.exists())
 def test_names_are_portable_executable_path_components(self):
  for value in ("App", "Test App", "Vellum-Editor", "editor.2"):
   self.assertTrue(app.valid_name(value),value)
  for value in ("App\n", "App\x7f", "App*", "App?", 'App"', "App<", "App>", "App|", "App ", "App.", "-App", "CON", "con.txt", "LPT1", "COM9", "NUL", "APP_MANIFEST.json", "app_manifest.JSON"):
   self.assertFalse(app.valid_name(value),value)
 def test_manifest_filename_cannot_overwrite_linux_executable(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\x7fELFfixture")
   with self.assertRaisesRegex(ValueError,"invalid app name"):app.package(binary,root/"output","APP_MANIFEST.json","org.example.app","0.1.0","linux")
   self.assertFalse((root/"output").exists());self.assertEqual(binary.read_bytes(),b"\x7fELFfixture")
 def test_archive_modes_do_not_depend_on_posix_chmod_support(self):
  for target in ("windows","linux"):
   with self.subTest(target=target),tempfile.TemporaryDirectory() as tmp,mock.patch.object(Path,"chmod",return_value=None):
    root=Path(tmp);binary=root/"binary";binary.write_bytes((b"MZ"+b"\x00"*58+(64).to_bytes(4,"little")+b"PE\x00\x00") if target=="windows" else b"\x7fELFfixture")
    output=root/"app.archive";app.package(binary,output,"App","org.example.app","0.1.0",target)
    if target=="windows":
     with zipfile.ZipFile(output) as archive:
      self.assertEqual(archive.getinfo("App/").external_attr>>16,0o040755)
      self.assertEqual(archive.getinfo("App/App.exe").external_attr>>16,0o100755)
      self.assertEqual(archive.getinfo("App/APP_MANIFEST.json").external_attr>>16,0o100644)
    else:
     with tarfile.open(output) as archive:
      self.assertEqual(archive.getmember("App").mode,0o755)
      self.assertEqual(archive.getmember("App/App").mode,0o755)
      self.assertEqual(archive.getmember("App/APP_MANIFEST.json").mode,0o644)
 def test_system_library_prefix_does_not_allow_parent_traversal(self):
  for path in ("/usr/lib/../../tmp/evil.dylib","/System/Library/../../Users/x/evil.dylib","/usr/lib/../local/evil.dylib"):
   with self.subTest(path=path),mock.patch.object(app,"check_child_output",return_value="binary:\n\t"+path+" (compatibility version 1)\n"):
    with self.assertRaisesRegex(ValueError,"audited bundle closure"):app.darwin_dependencies(Path("binary"))
 def test_missing_native_closure_is_not_accepted(self):
  with mock.patch.object(app,"check_child_output",return_value="binary:\n\t@rpath/libmissing.dylib (compatibility version 1)\n"):
   with self.assertRaisesRegex(ValueError,"audited bundle closure"):app.darwin_dependencies(Path("fixture"))
 def test_invalid_names_signing_target_and_dangling_output(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\x7fELFfixture")
   with self.assertRaisesRegex(ValueError,"invalid app name"):app.package(binary,root/"app","../escape","org.example.app","0.1.0","linux")
   with self.assertRaisesRegex(ValueError,"signing"):app.package(binary,root/"app","App","org.example.app","0.1.0","linux","-")
   output=root/"app";output.symlink_to(root/"missing")
   with self.assertRaisesRegex(ValueError,"exists"):app.package(binary,output,"App","org.example.app","0.1.0","linux")
   self.assertTrue(output.is_symlink())
 def test_copied_format_is_validated_after_source_mutation(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\x7fELFfixture");output=root/"app.archive"
   actual_copy=app.shutil.copyfile
   def replace_then_copy(source,destination):
    source.write_bytes(b"not an ELF executable")
    return actual_copy(source,destination)
   with mock.patch.object(app.shutil,"copyfile",side_effect=replace_then_copy),self.assertRaisesRegex(ValueError,"binary format"):
    app.package(binary,output,"App","org.example.app","0.1.0","linux")
   self.assertFalse(output.exists())
 def test_dependency_audit_reads_staged_bytes_before_publication(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"):
   root=Path(tmp);binary=root/"binary";content=b"\xcf\xfa\xed\xfefixture";binary.write_bytes(content);output=root/"Audit.app"
   inspected=[]
   def inspect(executable):
    inspected.append(executable)
    self.assertNotEqual(executable,binary)
    self.assertEqual(executable.read_bytes(),content)
    self.assertFalse(output.exists())
    raise ValueError("fixture staged dependency rejection")
   with mock.patch.object(app,"darwin_dependencies",side_effect=inspect),self.assertRaisesRegex(ValueError,"staged dependency rejection"):
    app.package(binary,output,"Audit","org.example.app","0.1.0","darwin")
   self.assertEqual(len(inspected),1);self.assertFalse(output.exists())
if __name__=="__main__":unittest.main()
