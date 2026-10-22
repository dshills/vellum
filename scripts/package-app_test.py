"""Private app layout/overwrite/dependency gates without signing or a GUI."""
import importlib.util
import errno
import os
import subprocess
import json
import hashlib
import plistlib
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest import mock
import zipfile
spec=importlib.util.spec_from_file_location("package_app",Path(__file__).with_name("package-app.py"));app=importlib.util.module_from_spec(spec);spec.loader.exec_module(app)
policy=app.load_release_policy()
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
   manifest=json.loads((out/"Contents/Resources/APP_MANIFEST.json").read_text());self.assertEqual(manifest["sdk_release"],json.loads((app.ROOT/"release/version.json").read_text())["version"]);self.assertEqual(manifest["build_version"],"1")
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
 def notice_bundle(self,root,binary):
  source=root/"notices";source.mkdir();(source/"NOTICES").mkdir();body=b"Original fixture declaration\n";(source/"NOTICES/LICENSE").write_bytes(body)
  inventory={"schema":1,"target":"darwin","artifact_sha256":hashlib.sha256(binary.read_bytes()).hexdigest(),"distribution_audit":"pending","applicability":"review-required","scope":"Explicit fixture only; completeness unverified","files":{"NOTICES/LICENSE":{"bytes":len(body),"sha256":hashlib.sha256(body).hexdigest()}},"components":[{"id":"fixture@1","notice_files":["NOTICES/LICENSE"]}]}
  (source/"INVENTORY.json").write_text(json.dumps(inventory,indent=2)+"\n");return source
 def test_mac_notices_bind_staged_bytes_and_copy_original_payloads(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"),mock.patch.object(app,"run_child") as execute:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app"
   app.package(binary,out,"Test","org.example.test","0.1.0","darwin",notice_bundle=source)
   manifest=json.loads((out/"Contents/Resources/APP_MANIFEST.json").read_text());record=manifest["component_notices"]
   self.assertEqual(record["artifact_sha256"],manifest["source_binary_sha256"]);self.assertEqual(record["files"],1);self.assertEqual(record["distribution_audit"],"pending");self.assertEqual(manifest["native_distribution_audit"],"pending")
   for name in ["INVENTORY.json","NOTICES/LICENSE"]:self.assertEqual((out/"Contents/Resources/THIRD_PARTY_NOTICES"/name).read_bytes(),(source/name).read_bytes())
   self.assertEqual((out/"Contents/Resources/THIRD_PARTY_NOTICES/NOTICES/LICENSE").stat().st_mode&0o777,0o644);execute.assert_not_called()
 def test_notice_binary_mismatch_and_payload_rejection_leave_no_output(self):
  for kind in ["binary","payload"]:
   with self.subTest(kind=kind),tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"),mock.patch.object(app,"run_child") as execute:
    root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app"
    if kind=="binary":binary.write_bytes(b"\xcf\xfa\xed\xfechanged")
    else:(source/"NOTICES/LICENSE").write_text("changed")
    with self.assertRaises(ValueError):app.package(binary,out,"Test","org.example.test","0.1.0","darwin",notice_bundle=source)
    self.assertFalse(out.exists());execute.assert_not_called();self.assertFalse(list(root.glob(".vellum-app-*")))
 def test_notice_binding_checks_copied_executable_after_valid_source_replacement(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app";actual=app.shutil.copyfile
   def replace_then_copy(src,dst):
    src.write_bytes(b"\xcf\xfa\xed\xfereplaced-valid-format");return actual(src,dst)
   with mock.patch.object(app.shutil,"copyfile",side_effect=replace_then_copy),self.assertRaisesRegex(ValueError,"executable binding"):app.package(binary,out,"Test","org.example.test","0.1.0","darwin",notice_bundle=source)
   self.assertFalse(out.exists())
 def test_notices_are_verified_before_publication_after_explicit_signing(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app"
   def execute(args,**kwargs):
    if "--sign" in args:(Path(args[-1])/"Contents/Resources/THIRD_PARTY_NOTICES/NOTICES/LICENSE").write_text("tampered")
   with mock.patch.object(app,"run_child",side_effect=execute),self.assertRaises(ValueError):app.package(binary,out,"Test","org.example.test","0.1.0","darwin","-",notice_bundle=source)
   self.assertFalse(out.exists())
 def test_final_notice_tampering_removes_only_owned_output(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app";actual=app.shutil.move
   def move_and_tamper(src,dst):
    result=actual(src,dst);(out/"Contents/Resources/THIRD_PARTY_NOTICES/NOTICES/LICENSE").write_text("tampered");return result
   with mock.patch.object(app.shutil,"move",side_effect=move_and_tamper),self.assertRaises(ValueError):app.package(binary,out,"Test","org.example.test","0.1.0","darwin",notice_bundle=source)
   self.assertFalse(out.exists());self.assertTrue((source/"NOTICES/LICENSE").exists());self.assertTrue(binary.exists())
 def test_notice_bundle_is_mac_only_and_never_implicitly_signed(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\x7fELFfixture")
   with self.assertRaisesRegex(ValueError,"only for macOS"):app.package(binary,root/"out","Test","org.example.test","0.1.0","linux",notice_bundle=root/"missing")
   self.assertFalse((root/"out").exists())
 def test_executable_and_manifest_tampering_during_publication_refuse(self):
  for kind in ["executable","manifest"]:
   with self.subTest(kind=kind),tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
    root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app";actual=app.shutil.move
    def move_and_tamper(src,dst):
     result=actual(src,dst);path=out/("Contents/MacOS/Test" if kind=="executable" else "Contents/Resources/APP_MANIFEST.json");body=bytearray(path.read_bytes());body[-1]^=1;path.write_bytes(body);return result
    with mock.patch.object(app.shutil,"move",side_effect=move_and_tamper),self.assertRaises(ValueError):app.package(binary,out,"Test","org.example.test","0.1.0","darwin",notice_bundle=source)
    self.assertFalse(out.exists());self.assertTrue(binary.exists())
 def test_pre_sign_executable_and_manifest_revalidation(self):
  for kind in ["executable","manifest"]:
   with self.subTest(kind=kind),tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"),mock.patch.object(app,"run_child") as execute:
    root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app";actual=app.normalize_layout
    def tamper(stage):
     actual(stage);path=stage/("Contents/MacOS/Test" if kind=="executable" else "Contents/Resources/APP_MANIFEST.json");body=bytearray(path.read_bytes());body[-1]^=1;path.write_bytes(body)
    with mock.patch.object(app,"normalize_layout",side_effect=tamper),self.assertRaises(ValueError):app.package(binary,out,"Test","org.example.test","0.1.0","darwin","-",notice_bundle=source)
    execute.assert_not_called();self.assertFalse(out.exists())
 def test_explicit_signing_transformation_is_bound_through_publication(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app"
   def execute(args,**kwargs):
    if "--sign" in args:
     path=Path(args[-1])/"Contents/MacOS/Test";path.write_bytes(path.read_bytes()+b"simulated signature transformation")
   with mock.patch.object(app,"run_child",side_effect=execute):result=app.package(binary,out,"Test","org.example.test","0.1.0","darwin","-",notice_bundle=source)
   manifest=json.loads((out/"Contents/Resources/APP_MANIFEST.json").read_text());final=hashlib.sha256((out/"Contents/MacOS/Test").read_bytes()).hexdigest()
   self.assertEqual(manifest["source_binary_sha256"],hashlib.sha256(binary.read_bytes()).hexdigest());self.assertEqual(result["packaged_binary_sha256"],final);self.assertNotEqual(final,manifest["source_binary_sha256"])
 def test_replacement_output_directory_is_preserved_on_failure(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");source=self.notice_bundle(root,binary);out=root/"Test.app";moved=root/"original-renamed";actual=app.shutil.move
   def replace_output(src,dst):
    result=actual(src,dst);out.rename(moved);out.mkdir();(out/"unrelated").write_bytes(b"preserve me");return result
   with mock.patch.object(app.shutil,"move",side_effect=replace_output),self.assertRaisesRegex(ValueError,"directory was replaced"):app.package(binary,out,"Test","org.example.test","0.1.0","darwin",notice_bundle=source)
   self.assertEqual((out/"unrelated").read_bytes(),b"preserve me");self.assertTrue((moved/"Contents/MacOS/Test").exists());self.assertTrue(binary.exists())
 def test_failure_before_output_identity_preserves_empty_unproven_path(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);binary=root/"binary";body=b"\xcf\xfa\xed\xfefixture";binary.write_bytes(body);source=self.notice_bundle(root,binary);out=root/"Test.app";actual=Path.lstat
   def fail_ownership_capture(path,*args,**kwargs):
    if path==out and out.exists():raise PermissionError("fixture identity acquisition failed")
    return actual(path,*args,**kwargs)
   with mock.patch.object(Path,"lstat",new=fail_ownership_capture),self.assertRaisesRegex(PermissionError,"identity acquisition failed"):app.package(binary,out,"Test","org.example.test","0.1.0","darwin",notice_bundle=source)
   self.assertTrue(out.is_dir());self.assertEqual(list(out.iterdir()),[]);self.assertEqual(binary.read_bytes(),body);self.assertFalse(list(root.glob(".vellum-app-*")))
   with self.assertRaisesRegex(ValueError,"exists"):app.package(binary,out,"Test","org.example.test","0.1.0","darwin",notice_bundle=source)
 def public_source(self, root, audit="approved"):
  checkout=root/"checkout";(checkout/"release").mkdir(parents=True)
  metadata={"version":"v9.8.7-preview.2","module":"github.com/dshills/vellum","namespace_status":"selected","license_status":"selected","third_party_distribution_audit":audit}
  (checkout/"release/version.json").write_text(json.dumps(metadata));(checkout/"LICENSE").write_text("fixture license")
  (checkout/"go.mod").write_text("module github.com/dshills/vellum\n")
  return checkout,metadata
 def test_public_requires_both_release_approval_and_binary_bound_notices(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);checkout,metadata=self.public_source(root,"pending");binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");notes=self.notice_bundle(root,binary);out=root/"Public.app"
   state={"commit":"a"*40,"clean":True}
   with mock.patch.object(app,"load_release_policy",return_value=policy),mock.patch.object(app,"ROOT",checkout),mock.patch.object(policy,"source_state",return_value=state),mock.patch.object(policy,"hidden_index_paths",return_value=[]),mock.patch.object(policy,"git_output",return_value=str(checkout)):
    with self.assertRaisesRegex(ValueError,"binary-bound notices"):app.package(binary,out,"Public","org.example.public","0.1.0","darwin",public=True)
    with self.assertRaisesRegex(ValueError,"audit is incomplete"):app.package(binary,out,"Public","org.example.public","0.1.0","darwin",notice_bundle=notes,public=True)
    self.assertFalse(out.exists())
    metadata["third_party_distribution_audit"]="approved";(checkout/"release/version.json").write_text(json.dumps(metadata))
    inventory=json.loads((notes/"INVENTORY.json").read_text());inventory["artifact_sha256"]="b"*64;(notes/"INVENTORY.json").write_text(json.dumps(inventory))
    with self.assertRaisesRegex(ValueError,"binding"):app.package(binary,out,"Public","org.example.public","0.1.0","darwin",notice_bundle=notes,public=True)
    self.assertFalse(out.exists())
    inventory["artifact_sha256"]=hashlib.sha256(binary.read_bytes()).hexdigest();(notes/"INVENTORY.json").write_text(json.dumps(inventory))
    result=app.package(binary,out,"Public","org.example.public","0.1.0","darwin",notice_bundle=notes,public=True)
   manifest=json.loads((out/"Contents/Resources/APP_MANIFEST.json").read_text())
   self.assertFalse(result["private_build"]);self.assertEqual(result["native_distribution_audit"],"approved");self.assertFalse(result["notarized"])
   self.assertEqual(manifest["sdk_release"],metadata["version"]);self.assertEqual(manifest["release_authorization"]["source"],state)
   self.assertEqual(manifest["release_authorization"]["metadata_sha256"],hashlib.sha256((checkout/"release/version.json").read_bytes()).hexdigest())
   self.assertEqual(manifest["component_notices"]["distribution_audit"],"pending","declaration collection must not approve itself")
 def test_public_keeps_source_index_and_publication_rechecks(self):
  for case in ["dirty","index","wrong-checkout","metadata-changed","source-changed"]:
   with self.subTest(case=case),tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
    root=Path(tmp);checkout,metadata=self.public_source(root);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");notes=self.notice_bundle(root,binary);out=root/"Public.app";changed=False;original=app.normalize_layout
    def state(path):return {"commit":("b" if changed and case=="source-changed" else "a")*40,"clean":case!="dirty"}
    def normalize(stage):
     nonlocal changed
     original(stage);changed=True
     if case=="metadata-changed":
      metadata["third_party_distribution_audit"]="pending";(checkout/"release/version.json").write_text(json.dumps(metadata))
    top=root if case=="wrong-checkout" else checkout
    with mock.patch.object(app,"load_release_policy",return_value=policy),mock.patch.object(app,"ROOT",checkout),mock.patch.object(policy,"source_state",side_effect=state),mock.patch.object(policy,"hidden_index_paths",return_value=["hidden"] if case=="index" else []),mock.patch.object(policy,"git_output",return_value=str(top)),mock.patch.object(app,"normalize_layout",side_effect=normalize):
     with self.assertRaises(ValueError):app.package(binary,out,"Public","org.example.public","0.1.0","darwin",notice_bundle=notes,public=True)
    self.assertFalse(out.exists());self.assertTrue(binary.exists());self.assertFalse(list(root.glob(".vellum-app-*")))
 def test_private_mode_in_non_git_copy_does_not_load_public_policy(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app,"load_release_policy",side_effect=AssertionError("private policy load")):
   root=Path(tmp);checkout=root/"no-git";(checkout/"release").mkdir(parents=True);(checkout/"release/version.json").write_text('{"version":"v9.8.7"}')
   binary=root/"binary";binary.write_bytes(b"\x7fELFfixture")
   with mock.patch.object(app,"ROOT",checkout):result=app.package(binary,root/"out.tar.gz","App","org.example.app","0.1.0","linux")
   self.assertTrue(result["private_build"])
 def test_public_authorization_drift_during_output_move_removes_owned_bundle(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);checkout,metadata=self.public_source(root);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");notes=self.notice_bundle(root,binary);out=root/"Public.app";actual=app.shutil.move
   def move_and_revoke(src,dst):
    result=actual(src,dst);metadata["third_party_distribution_audit"]="pending";(checkout/"release/version.json").write_text(json.dumps(metadata));return result
   with mock.patch.object(app,"load_release_policy",return_value=policy),mock.patch.object(app,"ROOT",checkout),mock.patch.object(policy,"source_state",return_value={"clean":True,"commit":"a"*40}),mock.patch.object(policy,"hidden_index_paths",return_value=[]),mock.patch.object(policy,"git_output",return_value=str(checkout)),mock.patch.object(app.shutil,"move",side_effect=move_and_revoke):
    with self.assertRaisesRegex(ValueError,"metadata changed"):app.package(binary,out,"Public","org.example.public","0.1.0","darwin",notice_bundle=notes,public=True)
   self.assertFalse(out.exists());self.assertTrue(binary.exists());self.assertTrue((notes/"INVENTORY.json").exists())
 def test_public_source_change_across_initial_gates_refuses(self):
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(app.platform,"system",return_value="Darwin"),mock.patch.object(app,"darwin_dependencies"):
   root=Path(tmp);checkout,_=self.public_source(root);binary=root/"binary";binary.write_bytes(b"\xcf\xfa\xed\xfefixture");notes=self.notice_bundle(root,binary);out=root/"Public.app"
   a={"clean":True,"commit":"a"*40};b={"clean":True,"commit":"b"*40}
   with mock.patch.object(app,"load_release_policy",return_value=policy),mock.patch.object(app,"ROOT",checkout),mock.patch.object(policy,"source_state",side_effect=[a,b,b]),mock.patch.object(policy,"hidden_index_paths",return_value=[]),mock.patch.object(policy,"git_output",return_value=str(checkout)):
    with self.assertRaisesRegex(ValueError,"during authorization"):app.package(binary,out,"Public","org.example.public","0.1.0","darwin",notice_bundle=notes,public=True)
   self.assertFalse(out.exists())
 def test_public_mode_does_not_expand_other_platform_support(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);binary=root/"binary";binary.write_bytes(b"\x7fELFfixture")
   with self.assertRaisesRegex(ValueError,"only for macOS"):app.package(binary,root/"out","App","org.example.app","0.1.0","linux",public=True)
   self.assertFalse((root/"out").exists())

if __name__=="__main__":unittest.main()
