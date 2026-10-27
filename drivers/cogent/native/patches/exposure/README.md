# 0108 — bounded native window exposure presentation

Raising a surviving window after another closes, or receiving native focus/show,
must present that window’s retained scene even when no model value changes.
Previously those paths could schedule paint but the unchanged-scene fast path
returned before composition. A native backing surface that needed restoration
could therefore retain stale contents until an unrelated edit.

Use the existing per-window forced-frame bit and paint coalescer. A request is
consumed only after visibility and native-window ownership admit composition;
a hidden pass retains one request. No model mutation, forced command, timer loop,
new rendering backend or experimental policy option is added. The normal idle
path still skips unchanged work without an exposure request.

The pinned stage follows0107 and changes core/renderwindow.go plus an offscreen
regression file. Input/output hashes and patch SHA are checked by prepare.py.
Regressions use the real retained scene/software compositor and simulate stale
native destination pixels: close-other, raise, focus and show restore the owner’s
pixels without a model edit; repeated/busy raises coalesce; hidden passes retain
the request. Native Mac Shell observations remain separate from these tests.
