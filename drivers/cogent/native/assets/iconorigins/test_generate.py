import hashlib, importlib.util, json, pathlib, tempfile, unittest, xml.etree.ElementTree as ET
ROOT=pathlib.Path(__file__).resolve().parent
s=importlib.util.spec_from_file_location('neutral',ROOT/'generate.py');g=importlib.util.module_from_spec(s);s.loader.exec_module(g)
class Assets(unittest.TestCase):
    def test_reproducible_full_manifest(self):
        self.assertEqual(len(g.assets()),13)
        self.assertEqual(json.loads((ROOT/'manifest.json').read_text()),g.manifest())
        for name,body in g.assets().items():self.assertEqual((ROOT/name).read_bytes(),body)
    def test_no_fonts_links_or_fixed_tints(self):
        for name,body in g.assets().items():
            root=ET.fromstring(body);self.assertEqual(root.attrib['viewBox'],'0 0 32 32')
            for e in root.iter():
                self.assertIn(e.tag.split('}')[-1],('svg','path','rect','circle'))
                self.assertFalse(set(e.attrib)&{'fill','stroke','href','style','transform'})
    def test_aliases_and_different_states(self):
        b=g.assets();self.assertEqual(b['latex.svg'],b['tex.svg']);self.assertEqual(b['make.svg'],b['makefile.svg'])
        self.assertNotEqual(b['radio_button_partial.svg'],b['toggle_mid.svg'])
if __name__=='__main__':unittest.main()
