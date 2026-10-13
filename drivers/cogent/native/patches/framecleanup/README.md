# Owned frame failure cleanup

Patch 0048 shares checked pass preparation and submission between ordinary and
measured drawing. Encoder/pass/finish refusal releases partial handles and aborts
the acquired surface. Only a successfully finished and submitted buffer may be
presented or read as an image. Queue submission is synchronous admission; it does
not establish successful GPU execution. Legacy void End logs refusal, while the
checked End and drawer paths return it.

Surface retirement detaches owned view/texture references before release, releases
view before texture, and unlocks once. Idle abort/present is a no-op. Surface
Release also retires an outstanding acquisition before normal teardown. The frame
owner must serialize acquisition, encoding, presentation/abort and teardown;
concurrent or stale owners, arbitrary abandoned passes and broad panic recovery
are not supported by this checkpoint. RenderTexture has no acquired surface lock.

The new tests inject nil/partial creation and finish failures into the actual
shared control functions, check counted reference order and the actual Surface
mutex, and exercise both production End methods with an unsubmitted frame.
They do not inject failures into a physical GPU/backend, qualify device recovery,
or prove native graphics memory reclamation. Prepared input/output/patch hashes
are pinned; the module cache is unchanged. Modified upstream files retain their
BSD-style notices; new Vellum source/tests carry MIT notices.
