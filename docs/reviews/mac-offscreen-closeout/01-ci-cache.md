# Bounded CI cache upload — internal review

The ac61d78 Framework attempt passed all checks and benchmarks, then reached the
existing 60-minute job limit in the combined cache action's automatic post-save.
Replace combined actions/cache with explicit restore and save in both Framework
jobs. Optional upload runs only after successful verification with five minutes
of measured headroom, is capped at two minutes, and is non-gating. Near-deadline
cold runs skip upload; they do not skip verification. Existing job deadlines,
commands, dependency integrity checks, cache keys and data scope are retained.
A cold verification run that itself exceeds its deadline still fails normally.

Internal review found and corrected a save-path typo before commit: the Framework
job must save vellum-go-cache, exactly as restored and used by env.sh. A workflow
consistency regression now checks both jobs' paths and primary save key, alongside
clock-boundary and invalid-input tests. Restore-only actions have no automatic
post-save, so a skipped upload cannot reappear during job cleanup.

Validation: three cache budget/consistency tests, two stage-dispatch tests, Ruby
YAML parse and git diff --check passed. All preexisting verification commands
remain in order. No hosted run is claimed for this unpushed checkpoint; the prior
main workflows remain green. No desktop use, native policy or stable API change.
