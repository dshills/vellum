"""SDK archive tamper, coverage and safe extraction checks without a native display."""
import hashlib
import importlib.util
import io
import json
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest import mock
import gzip
import os
import zlib
from child_process import ChildCleanupError

spec=importlib.util.spec_from_file_location("check_sdk",Path(__file__).with_name("check-sdk.py"));sdk=importlib.util.module_from_spec(spec);spec.loader.exec_module(sdk)
PREFIX="vellum-v0.1.0-alpha.1"
ORIGINS={"github.com/dshills/vellum":".","github.com/dshills/vellum/drivers/cogent":"drivers/cogent","cogentcore.org/core":"drivers/cogent/build/upstream","github.com/go-text/typesetting":"drivers/cogent/build/typesetting"}
def artifact(path,files=None,extra=None,checksums=None,release_record=None):
 files=dict(files or {"api.go":b"package vellum\n"})
 files.setdefault("go.mod",b"module github.com/dshills/vellum\ngo 1.25.6\n")
 if checksums is not None:
  checksums=dict(checksums);checksums.setdefault("go.mod",hashlib.sha256(files["go.mod"]).hexdigest())
 manifest={"schema":1,"release":release_record if release_record is not None else {"version":"v0.1.0-alpha.1","module":"github.com/dshills/vellum"},"files":checksums if checksums is not None else {name:hashlib.sha256(data).hexdigest() for name,data in files.items()}}
 with tarfile.open(path,"w:gz") as archive:
  for name,data in [*files.items(),("SDK_MANIFEST.json",json.dumps(manifest).encode())]:
   member=tarfile.TarInfo(PREFIX+"/"+name);member.size=len(data);archive.addfile(member,io.BytesIO(data))
  if extra:archive.addfile(extra)
class SDKChecks(unittest.TestCase):
 def test_native_benchmark_reports_geometry_isolation(self):
  report={"event":"vellum.native.ready","geometry_persistence_disabled_in_fixture":True}
  self.assertEqual(sdk.native_benchmark_report("diagnostic output\n"+json.dumps(report)+"\n"),report)
 def test_native_benchmark_rejects_missing_false_or_coerced_geometry_flag(self):
  for value in (None,False,1,"true"):
   report={"event":"vellum.native.ready"}
   if value is not None:report["geometry_persistence_disabled_in_fixture"]=value
   with self.subTest(value=value),self.assertRaisesRegex(ValueError,"did not disable window geometry"):
    sdk.native_benchmark_report(json.dumps(report))
 def test_native_benchmark_requires_exactly_one_report(self):
  report=json.dumps({"event":"vellum.native.ready","geometry_persistence_disabled_in_fixture":True})
  for output in ("diagnostic only",report+"\n"+report):
   with self.subTest(output=output),self.assertRaisesRegex(ValueError,"exactly one verified GPU report"):
    sdk.native_benchmark_report(output)
 def test_exact_hash_coverage_and_immutable_extraction(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);path=root/"sdk.tar.gz";artifact(path)
   verified=sdk.verify_archive(path);path.write_bytes(b"replaced after verification")
   extracted=sdk.extract_verified(path,root/"SDK with spaces",verified)
   self.assertEqual((extracted/"api.go").read_bytes(),b"package vellum\n")
   with self.assertRaisesRegex(ValueError,"exists"):sdk.extract_verified(path,root/"SDK with spaces",verified)
 def test_tampered_member_fails(self):
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"sdk.tar.gz";artifact(path,checksums={"api.go":"0"*64})
   with self.assertRaisesRegex(ValueError,"checksum mismatch"):sdk.verify_archive(path)
 def test_unrecorded_member_fails(self):
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"sdk.tar.gz";artifact(path,checksums={})
   with self.assertRaisesRegex(ValueError,"coverage mismatch"):sdk.verify_archive(path)
 def test_archive_authenticity_requires_trusted_hash(self):
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"sdk.tar.gz";artifact(path)
   with self.assertRaisesRegex(ValueError,"SHA-256 mismatch"):sdk.verify_archive(path,"0"*64)
 def test_unsafe_members_and_duplicates_fail(self):
  for name,kind in [(PREFIX+"/../escape",tarfile.REGTYPE),(PREFIX+"/link",tarfile.SYMTYPE),(PREFIX+"/hard",tarfile.LNKTYPE),(PREFIX+"/api.go",tarfile.REGTYPE),("/outside",tarfile.REGTYPE)]:
   with self.subTest(name=name),tempfile.TemporaryDirectory() as tmp:
    path=Path(tmp)/"sdk.tar.gz";member=tarfile.TarInfo(name);member.type=kind;member.linkname="/outside"
    checksums={"api.go":hashlib.sha256(b"package vellum\n").hexdigest()}
    if name!=PREFIX+"/api.go":checksums[name.removeprefix(PREFIX+"/")]=hashlib.sha256(b"").hexdigest()
    artifact(path,extra=member,checksums=checksums)
    expected="duplicate SDK archive member" if name==PREFIX+"/api.go" else "unsafe SDK archive member"
    with self.assertRaisesRegex(ValueError,expected):sdk.verify_archive(path)
 def test_verification_rejects_all_resource_budgets(self):
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"sdk.tar.gz";artifact(path)
   for constant,limit,message in [("MAX_COMPRESSED_BYTES",1,"compressed"),("MAX_MEMBER_BYTES",1,"member size"),("MAX_MEMBERS",1,"member count"),("MAX_EXPANDED_BYTES",1,"expanded size"),("MAX_MANIFEST_BYTES",1,"manifest size")]:
    with self.subTest(constant=constant),mock.patch.object(sdk,constant,limit):
     with self.assertRaisesRegex(ValueError,message):sdk.verify_archive(path)
 def test_extended_metadata_is_bounded_before_tarfile_parsing(self):
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"sdk.tar.gz";header=tarfile.TarInfo("pax");header.type=tarfile.XHDTYPE;header.size=sdk.MAX_METADATA_BYTES+1
   path.write_bytes(gzip.compress(header.tobuf()))
   with self.assertRaisesRegex(ValueError,"metadata size"):
    sdk.verify_archive(path)
 def test_pax_framing_overrides_cannot_bypass_physical_preflight(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp)
   for key,value in (("size","512"),("GNU.sparse.size","512")):
    record=f"{key}={value}\n".encode();length=len(record)+2
    while length!=len(record)+len(str(length))+1:length=len(record)+len(str(length))+1
    payload=str(length).encode()+b" "+record
    pax=tarfile.TarInfo("pax");pax.type=tarfile.XHDTYPE;pax.size=len(payload)
    file=tarfile.TarInfo(PREFIX+"/zero.go");file.size=0
    huge=tarfile.TarInfo("huge");huge.type=tarfile.XHDTYPE;huge.size=sdk.MAX_METADATA_BYTES+1
    raw=pax.tobuf()+payload+b"\0"*((-len(payload))%512)+file.tobuf()+b"\0"*512+huge.tobuf()
    path=root/(key+".tar.gz");path.write_bytes(gzip.compress(raw))
    with self.subTest(key=key),self.assertRaisesRegex(ValueError,"unsupported SDK PAX metadata key"):
     sdk.verify_archive(path)
 def test_producer_path_pax_remains_supported(self):
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"sdk.tar.gz";name="directory/"+"long-name-"*25+".go"
   artifact(path,files={name:b"package longname"})
   verified=sdk.verify_archive(path)
   self.assertIn(name,verified["manifest"]["files"])
 def test_ordinary_members_are_hashed_in_bounded_chunks(self):
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"sdk.tar.gz";artifact(path,files={"large.go":b"x"*500})
   sizes=[];original=tarfile.TarFile.extractfile
   class Reader:
    def __init__(self,source):self.source=source
    def read(self,size=-1):sizes.append(size);return self.source.read(size)
   def tracked(archive,member):
    source=original(archive,member)
    return Reader(source) if member.name.endswith("/large.go") else source
   with mock.patch.object(sdk,"CHUNK_BYTES",7),mock.patch.object(tarfile.TarFile,"extractfile",tracked):sdk.verify_archive(path)
   self.assertGreater(len(sizes),10);self.assertEqual(set(sizes),{7})
 def test_manifest_module_is_validated_before_archive_execution(self):
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"sdk.tar.gz"
   for value in (None,"",17,"../outside"):
    record={"version":"v0.1.0-alpha.1"}
    if value is not None:record["module"]=value
    artifact(path,release_record=record)
    with self.subTest(module=value),mock.patch.object(sdk,"run_child") as execute,self.assertRaisesRegex(ValueError,"invalid SDK manifest module path"):
     sdk.check(path,expected_sha256=hashlib.sha256(path.read_bytes()).hexdigest())
    execute.assert_not_called()
   artifact(path,release_record={"version":"v0.1.0-alpha.1","module":"example.com/other"})
   with self.assertRaisesRegex(ValueError,"differs from archived root go.mod"):sdk.verify_archive(path)
 def test_child_environment_preserves_build_settings_and_drops_unrelated_secrets(self):
  fixture={"PATH":"toolchain", "HOME":"home", "GOWORK":"original", "GOFLAGS":"-modfile=outside", "GOENV":"outside-env", "GOTOOLCHAIN":"auto", "GONOSUMDB":"*", "GOCACHE":"cache", "GOROOT":"outside-toolchain", "PKG_CONFIG_PATH":"outside-native-metadata", "CGO_LDFLAGS_ALLOW":".*", "CC":"outside-compiler", "CGO_CFLAGS":"-include outside", "HTTPS_PROXY":"transport-proxy", "SSL_CERT_FILE":"certificate", "DISPLAY":"display", "DBUS_SESSION_BUS_ADDRESS":"desktop-session", "GPU_DEVICE":"selected", "ANTHROPIC_API_KEY":"fixture-token", "AWS_SECRET_ACCESS_KEY":"fixture-cloud", "GITHUB_TOKEN":"fixture-ci"}
  with mock.patch.dict(sdk.os.environ,fixture,clear=True):
   environment=sdk.child_environment()
  self.assertEqual(environment,{name:value for name,value in fixture.items() if name in sdk.CHILD_ENVIRONMENT}|sdk.GO_CONFIGURATION)
  self.assertFalse(any(name in environment for name in ("ANTHROPIC_API_KEY","AWS_SECRET_ACCESS_KEY","GITHUB_TOKEN")))
  self.assertNotIn("CC",environment);self.assertNotIn("CGO_CFLAGS",environment)
  self.assertNotIn("GOROOT",environment);self.assertNotIn("PKG_CONFIG_PATH",environment)
  self.assertNotIn("DISPLAY",environment);self.assertNotIn("DBUS_SESSION_BUS_ADDRESS",environment);self.assertNotIn("GPU_DEVICE",environment)
  with mock.patch.dict(sdk.os.environ,fixture,clear=True):
   runtime_environment=sdk.child_environment(runtime=True)
  self.assertEqual(runtime_environment["DISPLAY"],"display");self.assertEqual(runtime_environment["GPU_DEVICE"],"selected")
  self.assertNotIn("ANTHROPIC_API_KEY",runtime_environment)
  self.assertEqual(environment["CGO_LDFLAGS_ALLOW"],r"^-Wl,-rpath,@executable_path/native$")
  self.assertEqual(environment["HTTPS_PROXY"],"transport-proxy");self.assertEqual(environment["SSL_CERT_FILE"],"certificate")
 def test_native_runtime_check_enforces_geometry_report(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp)
   modules="\n".join(json.dumps({"Path":name,"Replace":{"Dir":str(root/relative)}}) for name,relative in ORIGINS.items())
   verified={"archive_sha256":"digest","manifest":{"release":{"module":"github.com/dshills/vellum","version":"v0.1.0-alpha.1"},"files":{}}}
   for flag in (None,False,1,"true",True):
    report={"event":"vellum.native.ready"}
    if flag is not None:report["geometry_persistence_disabled_in_fixture"]=flag
    with self.subTest(flag=flag),mock.patch.object(sdk,"verify_archive",return_value=verified),mock.patch.object(sdk,"extract_verified",return_value=root),mock.patch.object(sdk,"run_child",return_value=mock.Mock(stdout=json.dumps(report))) as commands,mock.patch.object(sdk,"check_child_output",return_value=modules):
     if flag is True:
      self.assertEqual(sdk.check(Path("own-archive"),expected_sha256="digest",native_runtime=True)["native_runtime"],report)
     else:
      with self.assertRaisesRegex(ValueError,"did not disable window geometry"):
       sdk.check(Path("own-archive"),expected_sha256="digest",native_runtime=True)
     self.assertEqual(Path(commands.call_args.args[0][0]).name,"native-diagnostic")
     self.assertEqual(commands.call_args.args[0][1:],["--benchmark","--idle=1s","--samples=3"])
     self.assertEqual(commands.call_args.kwargs["timeout"],sdk.RUNTIME_TIMEOUT_SECONDS)
 def test_execution_requires_digest_or_explicit_self_attestation(self):
  with mock.patch.object(sdk,"verify_archive") as verify,mock.patch.object(sdk,"run_child") as execute:
   with self.assertRaisesRegex(ValueError,"trusted producer or explicit"):
    sdk.check(Path("missing"))
   verify.assert_not_called();execute.assert_not_called()
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);names=["github.com/dshills/vellum","github.com/dshills/vellum/drivers/cogent","cogentcore.org/core","github.com/go-text/typesetting"]
   modules="\n".join(json.dumps({"Path":name,"Replace":{"Dir":str(root/ORIGINS[name])}}) for name in names)
   verified={"archive_sha256":"digest","manifest":{"release":{"module":"github.com/dshills/vellum","version":"v0.1.0-alpha.1"},"files":{}}}
   with mock.patch.object(sdk,"verify_archive",return_value=verified),mock.patch.object(sdk,"extract_verified",return_value=root),mock.patch.object(sdk,"run_child") as commands,mock.patch.object(sdk,"check_child_output",return_value=modules) as listing:
    self.assertFalse(sdk.check(Path("own-archive"),allow_self_attested=True)["digest_verified_against_expected"])
    self.assertTrue(sdk.check(Path("trusted-archive"),expected_sha256="trusted-digest")["digest_verified_against_expected"])
    self.assertEqual(listing.call_args.kwargs["timeout"],sdk.COMMAND_TIMEOUT_SECONDS)
    for call in commands.call_args_list:self.assertIn(call.kwargs["timeout"],(sdk.COMMAND_TIMEOUT_SECONDS,sdk.GENERATOR_TIMEOUT_SECONDS))
 def test_command_timeout_fails_before_acceptance(self):
  import subprocess
  verified={"archive_sha256":"digest","manifest":{"release":{"module":"github.com/dshills/vellum","version":"v0.1.0-alpha.1"},"files":{}}}
  with tempfile.TemporaryDirectory() as tmp,mock.patch.object(sdk,"verify_archive",return_value=verified),mock.patch.object(sdk,"extract_verified",return_value=Path(tmp)),mock.patch.object(sdk,"run_child",side_effect=subprocess.TimeoutExpired("generator",sdk.GENERATOR_TIMEOUT_SECONDS)) as commands:
   with self.assertRaises(subprocess.TimeoutExpired):sdk.check(Path("own-archive"),expected_sha256="digest")
   self.assertEqual(commands.call_count,1)
   self.assertEqual(commands.call_args.kwargs["timeout"],sdk.GENERATOR_TIMEOUT_SECONDS)
 def test_missing_or_relative_replacement_directory_fails_closed_inside_sdk(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);original=Path.cwd()
   try:
    os.chdir(root)
    for directory in (None,"", "."):
     replacement={} if directory is None else {"Dir":directory}
     with self.subTest(directory=directory),self.assertRaisesRegex(ValueError,"no absolute local directory"):
      sdk.check_origin([{"Path":"github.com/dshills/vellum","Replace":replacement}],root)
   finally:os.chdir(original)
 def test_replacement_inside_sdk_must_match_exact_module_layout(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp)
   modules=[{"Path":name,"Replace":{"Dir":str(root/relative)}} for name,relative in ORIGINS.items()]
   self.assertEqual(sdk.check_origin(modules,root),ORIGINS)
   for index in range(len(modules)):
    wrong=[{"Path":module["Path"],"Replace":dict(module["Replace"])} for module in modules]
    wrong[index]["Replace"]["Dir"]=str(root/"some other SDK directory")
    with self.subTest(module=wrong[index]["Path"]),self.assertRaisesRegex(ValueError,"wrong directory"):
     sdk.check_origin(wrong,root)
 def test_cli_invalid_archive_reports_json_failure_without_traceback_or_success(self):
  import subprocess,sys
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"invalid.tar.gz";path.write_bytes(b"invalid gzip bytes")
   result=subprocess.run([sys.executable,str(Path(sdk.__file__)),"--archive",str(path),"--allow-self-attested"],capture_output=True,text=True)
   self.assertEqual(result.returncode,1);self.assertEqual(result.stdout,"")
   failure=json.loads(result.stderr);self.assertEqual(failure["event"],"vellum.sdk.check_failed")
   self.assertIn(failure["error_type"],("BadGzipFile","ReadError"))
 def test_corrupt_deflate_cli_is_structured(self):
  import subprocess,sys
  with tempfile.TemporaryDirectory() as tmp:
   path=Path(tmp)/"corrupt.gz"
   path.write_bytes(b"\x1f\x8b\x08\x00"+b"\x00"*6+b"\xff"*10)
   result=subprocess.run([sys.executable,str(Path(sdk.__file__)),"--archive",str(path),"--allow-self-attested"],capture_output=True,text=True)
   self.assertEqual(result.returncode,1);self.assertEqual(result.stdout,"")
   failure=json.loads(result.stderr);self.assertEqual(failure["event"],"vellum.sdk.check_failed");self.assertEqual(failure["error_type"],"error")
 def test_cli_expected_validation_and_child_failures_are_structured(self):
  from contextlib import redirect_stderr,redirect_stdout
  import subprocess
  self.assertTrue(issubclass(ChildCleanupError,OSError))
  failures=[ChildCleanupError("descendant termination unknown"),zlib.error("corrupt deflate"),ValueError("bad manifest"),EOFError("short archive"),tarfile.ReadError("bad tar"),OSError("file unavailable"),UnicodeDecodeError("utf-8",b"\xff",0,1,"invalid"),subprocess.CalledProcessError(7,["go","test"]),subprocess.TimeoutExpired("generator",60)]
  for error in failures:
   with self.subTest(error=type(error).__name__),mock.patch.object(sdk.sys,"argv",["check-sdk.py","--archive","unused","--allow-self-attested"]),mock.patch.object(sdk,"check",side_effect=error),redirect_stderr(io.StringIO()) as stderr,redirect_stdout(io.StringIO()) as stdout:
    with self.assertRaises(SystemExit) as exit:sdk.main()
    self.assertEqual(exit.exception.code,1);self.assertEqual(stdout.getvalue(),"")
    failure=json.loads(stderr.getvalue());self.assertEqual(failure["event"],"vellum.sdk.check_failed")
    if isinstance(error,subprocess.TimeoutExpired):self.assertEqual(failure["timeout_seconds"],60)
    else:self.assertEqual(failure["error_type"],type(error).__name__)
 def test_replacements_must_stay_in_sdk(self):
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp)/"sdk";root.mkdir()
   names=["github.com/dshills/vellum","github.com/dshills/vellum/drivers/cogent","cogentcore.org/core","github.com/go-text/typesetting"]
   modules=[{"Path":name,"Replace":{"Dir":str(root/ORIGINS[name])}} for name in names]
   self.assertEqual(set(sdk.check_origin(modules,root)),set(names))
   modules[-1]["Replace"]["Dir"]=str(Path(tmp)/"original checkout")
   with self.assertRaisesRegex(ValueError,"outside relocated SDK"):sdk.check_origin(modules,root)
   with self.assertRaisesRegex(ValueError,"missing explicit"):sdk.check_origin(modules[:2],root)
if __name__=="__main__":unittest.main()
