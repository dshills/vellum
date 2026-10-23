import unittest
import context


class SourceTransformation(unittest.TestCase):
    def test_refuses_missing_repeated_and_already_patched_sources(self):
        # Exact initial anchors for the two compiled consumer files.
        main = '\t\t\tif !good(state) {\n\t\t\t\treturn errors.New("foreground/configuration changed during cohort")\n\t\t\t}'
        observer = '  NSData *data=[NSJSONSerialization dataWithJSONObject:d'
        for original, patch in [(main, context.patch_main), (observer, context.patch_observer)]:
            transformed = patch(original)
            for invalid in ['', original * 2, transformed]:
                with self.subTest(patch=patch.__name__, invalid=invalid), self.assertRaises(ValueError):
                    patch(invalid)


if __name__ == '__main__':
    unittest.main()
