# Regression repair review

Prism reviewed all three decoded Go files and, separately, four preparation/runner/report paths. Both final reports have complete coverage and zero findings, exclusions, truncation or skipped inputs. The initial medium finding about case-insensitive Children decoding is fixed and covered by a regression. The manifest verifies canonical prepared sources match the reviewed snapshot and hashes the chosen raw test logs.

The full canonical text suite passes with a five-minute package budget; the TeX corpus takes about four minutes. No golden images were changed. Earlier overlapping diagnostic TeX runs and their output are not passing acceptance evidence. The pre-fix failure control and final parser race, editor, core and transport race results are included. G1 remains open for the separate live/platform/resource gates.
