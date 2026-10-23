"""Exercise the actual CLI failure exit/seal with a scalar fake host; no UI."""
import contextlib,io,json,runpy,signal,sys,tempfile,unittest
from pathlib import Path
from unittest.mock import patch
HOST="""import contextlib,hashlib,json
from pathlib import Path
need=lambda ok,why:None if ok else (_ for _ in ()).throw(ValueError(why))
sha=lambda p:hashlib.sha256(Path(p).read_bytes()).hexdigest()
sdk_vector=lambda:{}
def write(p,d):p.write_text(json.dumps(d))
class Lease:
 def __enter__(self):return self
 def __exit__(self,*args):pass
class lease_module:LiveLease=Lease
class c:
 @staticmethod
 def passed(run):return False
class Host:
 def __init__(self,*args):pass
 def geometry(self):return {}
 def probes(self,path):return {}
 def command(self,*args):return {'logs':{'stderr':{'bytes':0}}}
 def finish(self):pass
 def stopping(self):return False
"""
class ExitStatus(unittest.TestCase):
 def test_actual_runner_rejection_exits_nonzero_and_seals_without_UI(self):
  with tempfile.TemporaryDirectory() as directory:
   q=Path(directory);reuse=q/'reuse';(reuse/'macos-release-default-assessment').mkdir(parents=True);(reuse/'macos-release-default-assessment/assessment.py').write_text(HOST)
   (q/'plan.json').write_text(json.dumps({'pins':{},'sdk':{},'binary':'unused-no-execution','expected_match':{}}));(q/'preparation-commit.json').write_text(json.dumps({'head':'fake-reviewed-head','review_complete':True,'human_authorization':'scalar unit fixture only'}))
   argv=['startup/run.py','--artifacts',str(q),'--reuse',str(reuse),'--run-authorized']
   with patch.object(sys,'argv',argv),patch('subprocess.check_output',side_effect=['fake-reviewed-head\n','']),contextlib.redirect_stdout(io.StringIO()),self.assertRaises(SystemExit) as caught:
    runpy.run_path(str(Path(__file__).with_name('run.py')),run_name='__main__')
   self.assertEqual(caught.exception.code,1);r=json.loads((q/'result.json').read_text());self.assertEqual(r['outcome'],'REJECT');self.assertEqual(r['error']['message'],'startup command failure');self.assertTrue((q/'stop-live').exists());self.assertFalse((q/'native').exists())
 def test_actual_runner_restores_signals_when_result_write_fails(self):
  previous={sig:signal.getsignal(sig) for sig in [signal.SIGINT,signal.SIGTERM]}
  with tempfile.TemporaryDirectory() as directory:
   q=Path(directory);reuse=q/'reuse';(reuse/'macos-release-default-assessment').mkdir(parents=True)
   host=HOST.replace('def write(p,d):p.write_text(json.dumps(d))',"def write(p,d):\n if p.name=='result.json':raise OSError('simulated result write failure')\n p.write_text(json.dumps(d))")
   (reuse/'macos-release-default-assessment/assessment.py').write_text(host)
   (q/'plan.json').write_text(json.dumps({'pins':{},'sdk':{},'binary':'unused-no-execution','expected_match':{}}));(q/'preparation-commit.json').write_text(json.dumps({'head':'fake-reviewed-head','review_complete':True,'human_authorization':'scalar unit fixture only'}))
   argv=['startup/run.py','--artifacts',str(q),'--reuse',str(reuse),'--run-authorized']
   with patch.object(sys,'argv',argv),patch('subprocess.check_output',side_effect=['fake-reviewed-head\n','']),contextlib.redirect_stdout(io.StringIO()),self.assertRaisesRegex(OSError,'simulated result write failure'):
    runpy.run_path(str(Path(__file__).with_name('run.py')),run_name='__main__')
   self.assertEqual(previous,{sig:signal.getsignal(sig) for sig in previous});self.assertTrue((q/'started.json').exists());self.assertTrue((q/'stop-live').exists());self.assertFalse((q/'native').exists())
if __name__=='__main__':unittest.main()
