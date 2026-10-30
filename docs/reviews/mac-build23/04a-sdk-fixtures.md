# Generated document test fixtures — integration correction/internal review

The final source-81b3f19 producer correctly stopped during independent SDK
relocation: the generated document copied compatibility_test.go and
codec_fuzz_test.go but lacked their testdata/compatibility files. The generator
previously copied only top-level Go files. All historical/fuzz-seed tests then
reported missing fixture paths. No app was packaged or installed from that
incomplete attempt; the initial green-source build23 remains intact.

Copy regular Notes recipe testdata recursively into internal/notebook/testdata,
retaining exact bytes and provenance. Do not apply UI/import string replacements
to serialized fixtures or generator source records. Refuse links/special files;
the existing staged generator publishes the application only after completion.
This is a packaging integration fix, not a feature/API/format change.

Internal review checked fixture completeness, byte-preserving copy, correct
package-relative placement, no modification of original fixture/user files,
recursive link refusal and preservation of the generator's staged publication.
The default minimal template and app storage identity/arguments are unchanged.

Validation: three generator tests passed, including the existing deterministic
storage/import checks, exact recursive fixture byte comparison and link refusal.
An independent freshly generated document module passed go test -race ./... and
go vet ./... with Go 1.25.6, including its historical/fuzz seed notebook tests.
No native app entry point or window was launched. The corrected committed source
will be rebuilt through the same full SDK producer check; the failed attempt is
retained as diagnostic evidence, not a successful candidate. Desktop deferral,
feature/backend freeze and signing/audit policies remain unchanged.
