# Item 2: Work Log from the exported SDK

The repository recipe generates a standalone Work Log application from a verified
SDK archive, not checkout replacements. The actual consumer used build25's SDK
v1.1.1, source ed4f71f49bbf623dd8e205571669539212068c62. It is not relabeled as
v1.1.2. No runtime change distinguishes those application APIs in this task.
Generated application-owned sources preserve the tested document starter; the
customization is the Work Log window title, module-derived storage identity and
usage guide. It is a small notebook for dated work entries, not a new framework
feature. Initial example notes remain until the user deletes them.

All module origins were independently checked against the extracted SDK. Consumer
race tests, vet and build passed with Go 1.25.6. The first invocation failed due to
an ambient GOROOT pointing at Go 1.27.3; unsetting it resolved the mismatch. No SDK
or dependency version was changed to mask the problem.

The private ad-hoc signed app passed strict signature checks and native creation,
Unicode editing, insertion Undo/Redo, Save, normal Quit/restart and native Open
Notebook checks. Exact final bytes were asserted independently; the structured
result records their hash. The reopened note and seven-note count matched.
No framework workflow gap was found in this bounded scenario. IME, VoiceOver,
quarantine admission and sustained human use remain unqualified.

Default Prism completed full coverage. Its medium path-alias finding is fixed by
resolving paths before rejecting outputs inside the checkout. UTF-8 is explicit.
Its cleanup suggestion was declined: preserving partial generated sources aids
diagnosis and avoids deleting output; the documented retry uses a fresh path.
The generated source and app are reproducible from the recipe and pinned archive;
/private/tmp paths are temporary acceptance locations, not permanent artifacts.

The corrected generator successfully accepted /tmp as a harmless alias; rerunning against that existing output failed without changing main.go. The installed app at ~/Applications/Vellum/work-log-0.1.0/Vellum Work Log.app independently passed strict signature verification. The generated source/SDK is preserved in the workspace delivery archive; its absolute replacement paths require regeneration or deliberate edits after relocation.

The final default Prism review also completed full coverage. Its repeated partial-output finding remains a documented intentional policy: CONSUMER.json is published only after generation/customization completes; a failed directory is diagnostic, and exclusive retry refusal protects earlier source. Automatic recursive deletion was not adopted. Low findings about hypothetical stdlib-shadow files, pre-3.9 Python, missing parent directories and an inferred machine-readable stdout contract do not reflect the documented workflow: checked-in scripts are trusted local dependencies, the current tooling already uses modern pathlib APIs, callers create the parent and choose a fresh leaf, and stdout is human progress output. No pre-3.9 or machine-readable CLI contract is claimed. These suggestions are recorded rather than adding unrelated compatibility or cleanup behavior.
