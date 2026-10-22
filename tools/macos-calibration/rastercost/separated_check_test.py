import unittest,copy
from separated_check import validate,CONDITIONS
class SeparatedCheck(unittest.TestCase):
    def test_conditions_and_refusals(self):
        doc=dict(kind='offscreen-separated-raster',cpu=2,target=[1519,1266],pressure_bytes=32<<20,cache_flush_proven=False,release_qualified=False,gate_changed=False,elapsed_ns=2_000_000_000,samples=[])
        for round in range(8):
            for offset in range(4):
                name=CONDITIONS[(round+offset)%4];doc['samples'].append(dict(Round=round,Condition=name,ReplayNS=200_000,GapNS=121_000_000 if name.startswith('separated') else 2000,PressureNS=1000 if 'pressure' in name else 0,Color=dict(R=255 if round%2==0 else 0,G=0 if round%2==0 else 255,B=0,A=255)))
        self.assertEqual(len(validate(doc)),4)
        mutations=[lambda d:d['samples'].pop(),lambda d:d.update(cpu=14),lambda d:d.update(cache_flush_proven=True),lambda d:d.update(elapsed_ns=10_000_000_000),lambda d:d['samples'][0].update(ReplayNS=0),lambda d:d['samples'][1].update(GapNS=10),lambda d:d['samples'][0].update(PressureNS=1),lambda d:d['samples'][0].update(Round=1),lambda d:d['samples'][0].update(Color=dict(R=0,G=255,B=0,A=255))]
        for i,mutate in enumerate(mutations):
            bad=copy.deepcopy(doc);mutate(bad)
            with self.subTest(case=i),self.assertRaises(ValueError):validate(bad)
if __name__=='__main__':unittest.main()
