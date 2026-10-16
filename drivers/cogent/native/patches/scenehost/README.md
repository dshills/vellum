# GPU scene host diagnostics

This owned patch adds scalar observations to the existing explicit source-details diagnostic route. Capture/Encode/Lease are disjoint portions of source Draw; Resources is nested in Lease and Input in Resources. Counted queue input attempts, returned errors, logical input bytes and submit attempts do not establish GPU execution, physical transfer or presentation. CPU fallback preserves already attempted GPU observations.

The new regression sources cover exact input lengths, real checked-write refusal and owned restoration, late lease refusal after a submit attempt, ordinary/reused routes, native reference pixels, ticket detachment and unchanged UI-work accounting. They must be compiled and executed under explicit serial native admission before qualification. Source inspection, patch application and offscreen checks alone do not qualify those native tests or the ordinary frame latency targets.
