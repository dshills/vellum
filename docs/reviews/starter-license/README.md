# Generated application license preservation

The application generator copies the MIT LICENSE with its starter Go sources.
A real generator regression in a path containing spaces verifies license bytes
and the selected application module after running go mod edit. All 25 release
regressions pass. The preceding namespace change records the full framework
check suite; this follow-up changes no Go framework source.

Prism reviewed all four changed source/specification files using Anthropic
claude-sonnet-5-5. Coverage is complete with no findings, exclusions or truncation.
The source proof records the exact staged bytes before adding these review records.
A fresh clean-commit SDK is generated and tested separately after this commit.
