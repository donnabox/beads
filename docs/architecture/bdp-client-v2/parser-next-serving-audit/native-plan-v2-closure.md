# Native closure of finite Scope Event plan v2

Both findings from native-plan-review.md are closed. No remaining actionable concern in this narrow comparison. This is plan-review closure only; actual Claude review and root authoring/execution release remain outstanding.

- B/C now specify a local facade overriding only readEvents, forwarding unchanged to the real owner. Each wait uses real owner.withRead(() => true), asserting both successful entry and value. Probes are separately counted; inspect() is explicitly prohibited because it scans groups. No public instrumentation or added Event traffic is introduced.
- Cancellation now explicitly observes rejection, waits for first wait-entry, aborts, releases the deferred gate, and only then awaits rejection. Drain begins after cancellation settles. Its separate fixed clock observes elapsed0/5/10, close remains pending during its waits, and new service work is refused before owner access. Finally releases gates before awaiting settlements and close.
- Counts reconcile: B has2 service calls,4 delegated Event reads,2 waits/probes; C has4 service calls,5 delegated Event reads,3 waits/probes, plus one separate refused post-close owner call. The while-closing service refusal contributes zero owner attempts. A remains8 owner calls; D32 emitted HTTP; E3 installed HTTP; five stores, one listener and one child. No timeout, predicate, six-file scope, parser/auth precedence, cleanup ownership requirement or excluded policy changed.

Independently regenerated the unified v1→v2 diff and compared exact bytes with plan-v2.delta.patch. Changes are confined to the accepted seam/schedule/count clarifications and immutable-context/commit chronology. Source implementation remains untouched.

Independently verified current clean HEAD b76c3b7df77ea9b33611bc9027db651833e20abb, parent d4f3f4be88c7837b71d6646518ca3589d464ba3d, tree4d9f8ac00702d32a45c3eb38cf50fac55ac96f2d. All679 tracked working files AND committed blobs exactly match the predecessor frozen successor-1 allTrackedFilesAfter hashes. All12 source/spec/helper pins and7 immutable context pins match. The frozen design snapshot is byte-identical to the earlier bound design content (c5ab2d5bd89cf576cee9213460ee82000a73fd77bcd7038d4c423374f7b9a961); its original live path is recorded as provenance, not a review input or competing plan. The preparation binding truthfully retains its historical pending-commit state; the separate reanchor supplies the actual confirmed commit without rewriting it.

Reviewed artifact SHA256 values:

- plan-v2.md: 4db6021ec190c3ec6f7559234abba49568f6664fb25f294f628aa5999e6f6fe2
- plan-v2-binding.json: a6125d0519e965d3abe9afdc1cdbff66aeffacc2e1d6e3b190b5ef1d407fd053
- plan-v2.delta.patch: db9b65d7214cd47483dd69b74af82b5017b3df2408d87baab3f06cb0b2d66b6a
- plan-v2-commit-reanchor.json: 5dd23371ae3d3c22d88a198bf03de1aaf07750879bfb4a34678b53331b2e96f3

No source edits, product imports, tests, builds, installs, services or review launch occurred. Only this guarded report was written.
