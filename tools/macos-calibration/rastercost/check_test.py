import unittest
from check import parse,CASES
class RasterCostCheck(unittest.TestCase):
    def test_complete_and_refusals(self):
        text='\n'.join('BenchmarkRasterReplayBreakdown/'+n+'-2 100 123.4 ns/op 0 B/op 0 allocs/op' for n in CASES for _ in range(3))+'\nPASS\nok\n'
        self.assertEqual(len(parse(text)),5)
        cases=[text.replace('-2 ','-14 ',1),text.replace('123.4 ns/op','NaN ns/op',1),text.replace('0 allocs/op','1 allocs/op',1),text.replace('0 B/op','1 B/op',1),text.replace('100 123.4','0 123.4',1),text.replace('full_replay','unknown',1),text.replace('PASS','FAIL'),text.split('\n',1)[1],text+text.splitlines()[0]+'\n']
        for i,bad in enumerate(cases):
            with self.subTest(case=i),self.assertRaises(ValueError):parse(bad)
if __name__=='__main__':unittest.main()
