# Bounded two-test correction plan (not execution authority)

Basis: frozen current-observation correction-successor-1 source binding9c75bdd9409a3930ae75aaa539f08e0bd74904da6fed618f9b8c0206b4757310, fullpatchca59dc9f280ed0ac281f5b6ec00ed6d0fab1197fae7ef3b6a281bf817476eb55, completed actual CODE findings F1/F2. Root must release source changes separately.

Exactly two replacements, each uniquely constrained to the new callback/fixture:

1. packages/server/src/transaction-service.test.ts, currentObservationFixture.invoke:
   `return service.readAtMinimum({ kind: "authenticated", principal }, minimum).perform(request);`
   becomes
   `return service.readAtMinimum({ kind: "authenticated", principal: { id: principal } }, minimum).perform(request);`
   Leave receiptPage string principal, discovery, all route construction and method/counter/schedule logic untouched. Do not change production types or suppress typing errors.

2. packages/server/src/transaction-http.test.ts, the new current-observation callback's label === "types" expectation:
   order `{ id: beadType, name: "Task", describes: "bead" }` before `{ id: linkType, name: "Relation", describes: "link" }`.
   Change the expectation only. Leave runtime types configuration in its existing order, all validation/page/next/condition/context/body oracles and installed F's artifact-sorted Relation,Task expectation untouched. Do not normalize by sorting either side.

Preserve predecessor packets and all failed/completed review evidence. Save successor before/after snapshots, exact correction and inverse, full base diff, all679 source binding, updated author/report/schema/workload references, and explicit no-execution attribution. Independent native closure must verify two files changed/677 identical relative to this successor, exact inverse restoration, original16 owned scope, unchanged production and installed bytes, exact97 selection and unchanged9-new/28-legacy report contracts.

No new request, test title, callback, method, owner probe, operation, wait, clock, mutation, timeout, schema or assertion removal is authorized. Source E37, installed22/new37 and sibling5 fullCI-only remain. D retains its corrected ownership/equality/failure schema and separate added probe. Actual static checks may still identify formatting defects: do not claim a formatting failure has already run or use this release for repository-wide formatting.

After native exact-delta closure, root may disposition the completed full CODE findings as fixed without another full council. Rebind the already reviewed closed validation/required-check packets to the new source; all unexecuted2bootstrap+19 stages plus4 checks remain fresh. Retain actual tool/build/consumer gate discipline and fresh-build-before-module-collection. Do not copy or manufacture earlier successful results. Runtime and publication require their own root release and actual evidence.
