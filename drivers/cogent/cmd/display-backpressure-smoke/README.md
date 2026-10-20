# Display-backpressure lifecycle diagnostic

Run from `drivers/cogent` on macOS with the prepared native backend:

```sh
go run -tags=vellum_native_checks ./cmd/display-backpressure-smoke -case UNTAGGED_ORDINARY_CLOSE
go run -tags=vellum_native_checks ./cmd/display-backpressure-smoke -case UNTAGGED_CPU_BATCH_DEADLINE
go run -tags=vellum_native_checks ./cmd/display-backpressure-smoke -case UNTAGGED_PRE_SHOW_CPU_BATCH_DEADLINE
```

Each command uses a fresh process and a small disposable window. Geometry saving
is disabled through the shared initialized fixture. A cooperative eight-second
deadline and ten-second hard watchdog bound the application; a failed watchdog
does not establish normal cleanup. Native test controllers should additionally
own, retire and observe the whole process group.

The ordinary case requires a normal open and close. The other cases perform one 250 ms
application `MeshCanvas.Prepare` operation under a 100 ms experimental display
deadline, after Show or during the first Prepare before Show. They require the
typed deadline failure, `Driver.Run` return, worker join and terminal service and
window inventory observations. The pre-Show mode also emits a detached JSON entry
before its deliberate work, followed by the final JSON report if Run returns.

Both metrics recorders remain disabled. These are native lifecycle diagnostics
built with the `vellum_native_checks` tag, using public application/render APIs
and internal cleanup/geometry fixtures. They do not suppress presentation
callbacks or prove behavior after a missing callback, native deallocation,
physical input, responsiveness, endurance or release acceptance. Display
backpressure remains an explicit experiment; the ordinary driver default is
unchanged.
