import importlib.util
from pathlib import Path
import tempfile
import unittest
import wave

spec = importlib.util.spec_from_file_location("fixtures", Path(__file__).with_name("content-fixtures.py"))
fixtures = importlib.util.module_from_spec(spec)
spec.loader.exec_module(fixtures)


class ContentFixturesTest(unittest.TestCase):
    def test_pdf_offsets_audio_and_exclusive_directory(self):
        with tempfile.TemporaryDirectory() as parent:
            root = Path(parent) / "owned"
            first = fixtures.create(root)
            data = (root / "two-pages.pdf").read_bytes()
            xref = int(data.split(b"startxref\n")[1].splitlines()[0])
            self.assertEqual(data[xref:xref+4], b"xref")
            for offset in data[xref:].splitlines()[3:10]:
                at = int(offset.split()[0])
                self.assertRegex(data[at:].splitlines()[0], rb"^[1-7] 0 obj$")
            self.assertIn(b"/Count 2", data)
            with wave.open(str(root / "silent-60s.wav")) as audio:
                self.assertEqual(audio.getnframes() / audio.getframerate(), 60)
                self.assertEqual(audio.readframes(1), b"\0\0")
            with self.assertRaises(FileExistsError):
                fixtures.create(root)
            second = fixtures.create(Path(parent) / "other")
            self.assertEqual(first, second)


if __name__ == "__main__":
    unittest.main()
