# Composed campaign and effective occlusion

The 33-action composed fixture completed 1,000 native window lifecycles on
`c078eb8e20f0e493ee063949f5fe7ee85e51bbe4`, using the recorded 6,227 prepared
source files. Its raw trace contains 33,000 completed actions and 34,042 JSON
events. All owner/native completion, dense-label, tree, ListBox and named-canvas
predicates passed. The process ran for 11,584.145 seconds, including inspection
and publication waits; this is diagnostic wall time, not input/frame latency.
There were 33,271 attempts/admissions and 271 retried steps. Their cause was not
established; queue admission is separate from accepted owner completion.

The 41 forced-GC safety samples recorded Go heap allocations of
30,136,112–50,996,720 bytes and process RSS of 187,809,792–396,443,648 bytes.
The 512 MiB Go heap and 1 GiB RSS guards passed. These sampled safety bounds do
not establish professional memory budgets, absence of leaks or native/GPU
allocation. Terminal tracked services and callbacks retired; one hidden,
inactive, untitled 64×64 Cocoa window remained. The child exited zero and its
owned process group was observed absent. No zero-Cocoa-window or original-menu
restoration claim is made.

The following idle-soak preflight failed while establishing occlusion, after the
visible phase. A separately built diagnostic reproduction also failed. It
recorded equal target/cover outer frames, an opaque cover ordered in front, and
the target still reporting `OcclusionVisible=true`. This does not establish a
framework root cause. Apple's [occlusion documentation](https://developer.apple.com/documentation/appkit/nswindow/occlusionstate-swift.struct/visible?language=objc)
explicitly permits a nonrectangular window to count as visible when its bounding
box intersects a visible region. Equal outer frames therefore provide no
sufficient occlusion proof; the specific reason for this run remains unobserved.

The fixture now requests a cover with 32 window-manager points of margin around
the unchanged composed target. It records both windows' actual native frames,
ordering, levels and opacity at each transition, retaining the last valid pair
and final query error on failure. Occluded-phase admission requires an observed
opaque, fully containing cover in front and the target's strict native
`OcclusionVisible=false`; geometry never substitutes for that predicate.

The corrected 40-second preflight completed four phases, 33 native actions,
16 one-second quiet samples and four allocating-worker completions in 54.822
seconds overall. Actual target/cover frames were `[2048,0,3072,1173]` and
`[2016,-32,3136,1237]` at occluded admission. All four occluded samples retained
the strict native state. Tracked services retired and the owned group was
observed absent; a hidden inactive untitled 500×500 Cocoa window remained.
The source is `621aca2` plus the exact five reviewed fixture/specification leaves
in [qualification](records/qualification.json), separate from the earlier
`c078eb8` campaign. Harness race tests, cover rejection predicates, inert Cocoa
wire checks and vet checks passed at their recorded endpoints.

After those observations, Prism requested clearer callback variable names and
UTF-8-safe bounded error text. Inert prefix regressions cover the final logging
correction; native observations retain their exact pre-review source hashes.

Zero preflight samples qualify the 60-second idle interval. Background means
allocating-worker completion in a visible unfocused window, not OS application
background. App-wide wakeups/presents, continuous window-state observation,
native/GPU memory, physical input/displays, human accessibility, overnight soak
and complete P4/P6/P11/Mac release acceptance remain open. These fixtures request
no input-source, VoiceOver, clipboard or OS preference changes.

Acceptance requires evidence records preserving both failed attempts, the exact
successful source contexts, raw logs, validators, build associations, reviews
and cleanup. The sealed archive must exclude executable binaries; those remain
private observed outputs.
