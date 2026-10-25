"""Text-runner compilation/execution boundaries without opening a desktop."""
import importlib.util
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest import mock

spec=importlib.util.spec_from_file_location("text_backend",Path(__file__).with_name("check-text-backend.py"))
runner=importlib.util.module_from_spec(spec);spec.loader.exec_module(runner)

class TextRunnerDeadline(unittest.TestCase):
 def fixture(self, failure=None, platform=None, missing=False):
  temporary=tempfile.TemporaryDirectory(prefix="text runner test ")
  self.addCleanup(temporary.cleanup)
  root=Path(temporary.name)
  for name in ["upstream","typesetting"]:
   p=root/"drivers/cogent/build"/name;p.mkdir(parents=True)
   (p/"go.mod").write_text("module example.com/test\n")
   (p/"go.sum").write_text("")
  calls=[]
  def child(argv,**options):
   calls.append((argv,options))
   if failure is not None and len(calls)==failure:
    raise subprocess.CalledProcessError(1,argv)
   if "-c" in argv and not missing:Path(argv[argv.index("-o")+1]).write_bytes(b"fake owned test binary")
  with mock.patch.object(runner,"ROOT",root),mock.patch.object(runner,"run_child",side_effect=child),mock.patch.object(runner.sys,"platform",platform or runner.sys.platform):
   if missing:
    with self.assertRaisesRegex(ValueError,"required text test binary missing"):runner.main()
   elif failure:
    with self.assertRaises(subprocess.CalledProcessError):runner.main()
   else:runner.main()
  return calls
 def test_cold_build_is_separate_from_unchanged_execution_budget(self):
  calls=self.fixture()
  self.assertEqual(len(calls),9)
  self.assertEqual([a[-1] for a,_ in calls[1::2]],list(runner.TEXT_PACKAGES))
  for (build,bopts),(execute,eopts) in zip(calls[1::2],calls[2::2]):
   self.assertIn("-c",build);self.assertIn("-race",build);self.assertIn("offscreen",build)
   self.assertEqual(execute[0],build[build.index("-o")+1])
   self.assertEqual(execute[1:],["-test.run="+runner.TEXT_TESTS,"-test.count=1","-test.timeout=90s"])
   self.assertEqual(bopts["timeout"],300);self.assertEqual(eopts["timeout"],120)
   self.assertEqual(eopts["cwd"],bopts["cwd"]/build[-1].removeprefix("./"))
   self.assertTrue(bopts["check"] and eopts["check"])
 def test_windows_binary_uses_executable_suffix(self):
  calls=self.fixture(platform="win32")
  self.assertTrue(all(argv[0].endswith(".exe") for argv,_ in calls[2::2]))
 def test_missing_required_binary_refuses_instead_of_skipping(self):
  self.assertEqual(len(self.fixture(missing=True)),2)
 def test_compile_failure_never_executes_binary(self):
  self.assertEqual(len(self.fixture(failure=2)),2)
 def test_test_failure_stops_remaining_packages(self):
  self.assertEqual(len(self.fixture(failure=3)),3)

if __name__=="__main__":unittest.main()
