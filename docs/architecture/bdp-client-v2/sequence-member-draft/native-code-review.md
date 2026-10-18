# Independent native sequence-member code review

No actionable Critical, High, Medium or Low findings remain in the reviewed measured draft. The initial undefined sequence binding and embedded collection query defect are corrected. This is a source review, not a test result or publication approval.

## Exact historical review basis

Parent `18104a03b850553641f24fe83e7318493fa102ca`, tree `10810a8016a451e241ea974487681dcddd63c2e4`; accepted v8 plan `57c662f5e65fae5a4e1d103f1953a2754bbcf9a9d0e8ebb15f418805bab26f74`. Reviewed source binding `d9b68985603f168061e341ad4445b0881ac225449e9b93475b1541a4310dae79`, full diff `54947a6994071cd618bdc00c7287eb379f3060b0678a004899b24acce087cdca`.

| Source | SHA256 |
|---|---|
| manifest-v1.json | be0c229bec3441118eba469cda0c9d50ff9c9d919fcb2efbe1d79401269286cd |
| coverage-v1.json | 346e1edaa22af789f89017fdc2fef6bb01b7564ed72a65344e9fe5635887c74f |
| transactional-cumulative.test.ts | 424ff6e05e3e86c9484522e7c8f70565995560ec98b06a88e7945f46c30f179e |
| transactional-matrix.test.ts | dcf0161723a6b77dd4235111a83a6aa1ce0cbb6a1a5d79e362f992742bc503de |

I did not author this draft. I read the complete four-path diff, new manifest/coverage row, supporting runner/receipt/sequence/Event code and measured packet. Read-only Git/text inspection and standard-library JSON/hash arithmetic were used. No product imports, compiler, tests, build, services, installed consumers, external reviewers or repository mutation were performed by this reviewer.

The four hashes and complete diff matched during the substantive review. Immediately before report writing, the repeated source check stopped because root began an authorized manifest-format repair after actual format-check failure. This report deliberately preserves the reviewed pre-format snapshot; successor mechanical closure is required. It does not relabel that snapshot as current or claim the failed format-check passed. A prior binding check likewise detected the authorized fence amendment, after which review was rebound to d9b68985 above.

## Correctness and oracle closure

Corrected initial/replay requests use the existing Scope binding plus `operations/sequence`; collection path `beads/` uses query `{limit:"32"}`. The compact row SHA is `48e524596890fab7bf578cbbb653ed844178c33b568f8e0c033736d2745598ca`. Ten actions comprise seven raw requests and three receipt-helper GETs; only three URL roots are captured; there are 209 assertions.

App lines 2851–3141 check authored header names before values, exact body digest, genuine status/context, schemas and literal/absence facts before baselines. The replay ledger drives full replay comparison and performed inventory. Finish requires exact observation order and distinct receipt roots/transactions. Complete Resources, final collection, two Events and transaction linkage are checked. Independently calculated projected Event IDs are fixture/implementation-specific evidence, not normative spelling. `transaction-batch.ts:692–708,788–796` supplies the actual created Event fields. `transaction-sequence.ts:34–85` establishes projected index 1 versus retained singleton receipt index 0.

App lines 3171–3235 check actual helper exchanges, operator authority, original wire digest/root, serving session, dynamic confined GET and internal assertions. Full retained body equality comes from `transactional-observation.ts:198–313`, which stores validated original wire JSON and compares subsequent GET JSON; `runner.ts:900–988` validates that agreement and schema before client projection. The app correctly treats redacted output as `{kind:"json"}` and does not invent body recovery.

App lines 3143–3169 preserve the cross-carrier sentinel except for exact new initial/replay handoffs: literal URL/digest, complete write, absent outer key, raw indices 0/4. Cross.finish runs before initial handoff; every other ordinary response still reaches the old sentinel, and minimum observer receives every ordinary response. Helper GETs use their existing separate executor. Prior control mutations are unchanged. Full positive and both existing cross controls additionally require the new row to pass.

App lines 3237–3341 change only the first middle-delete target for the selected negative. The hook validates the genuine three-success position-3 response, exact deleted identity and absence of deleted properties/metadata. Manifest assertions stay unchanged; authored-request assertions compare against the actual changed request. Expected failures are exactly serving-position, middle-type, middle-code, middle-status, middle-retry, middle-no-outcome and middle-no-deleted. One exchange, zero client actions, no harness/transport/prerequisite/cleanup failure and two closed children are required. Actual fail-fast discrimination remains unexecuted by this reviewer; responses are not fabricated or substituted.

## Preservation and bounds

Independent JSON parsing reconstructed both complete parent files byte-for-byte by removing only the appended object: manifest `29729151cf6da641545910ec3880f9609c13b319af7bbe99d3d2ad47833dab96`, coverage `4619813e9446e0e714a64c11eac184531950f57720e32c7ef414c7e3342121af`. This verifies all 28 prior raw objects and order. The cumulative helper composes existing historical reconstructions; pagination index remains 10. WholeContract is false, and coverage explicitly excludes pending-all-keys, interleaving, crash/expiry, Changes enumeration and full-profile credit. No catalog, fixture, schema, runtime API or public admission changes occur.

Pre-format canonical measurements: disk790555/normal802361/nodes20956/coverage117702/actions220/assertions2503. Independent parsing rechecked all41 stored derivative byte/node records; joint maxima808479/21112. The final fence amendment versus source-author patch3facf550 changes only approved numerical fences plus one explanatory historical comment. Both app fences812575/21176 and canonical margins4096B/64nodes/1024coverage retain historical delta bases. Runtime/parser limits and timeouts remain unchanged.

Eight existing static callbacks and four cohort titles match the authored source. Prospective installed totals are213+1+204+205=623expectedHTTP,632maximum and23+2+23+23=71children. Accepted CI amplification remains55observe invocations/62expanded app cases. Counts are source models, not observed launches. The63-second positive assertion remains; review does not guarantee CI wall time.

## Disposition

No substantive source blocker to root-authorized bounded validation was identified. The subsequent formatting repair needs mechanical closure. Actual compilation/static results, exact positive/control outcomes, cleanup/provenance, independent Claude disposition and exact-head CI remain separate gates. No merge, release, qualification or broader execution authority is granted here.

## Final formatting successor closure

The preceding substantive review text has SHA256 `d7f1e87c749c957bf3187efaf01e61f92152169d10214aa8072fec5575a7fdb3`. Mechanical closure now binds source manifest `a6eb425d2616279fed28132c2b59258b7503b186224ba538179704b6364c9e2a` and complete patch `53c66bc9a5d07f53749e18628eaba32773458160ab506566ec1fb7960d13c313`, same parent/tree. Final manifest SHA is `a11a79f6d0f9ffa5933a680a3e4fed87987789c3deca50a38cda8563132d242e`; cumulative test SHA is `313e6a947b33a3893565ea5db2d9d4a962b46c2e926900598e9285e657a31bf2`. App and coverage hashes above remain exact.

I inspected the complete preformat-to-successor patch difference: five new-row JSON wrapping changes, disk measurement comment790555→790009, and disk fences306371→305825/794651→794105 only. Independent JSON/hash checks confirm the same compact new row and unchanged full old-row prefix; all four live source hashes and complete diff match the final binding. Disk is790009, normal802361; disk margin remains4096 and all normal/node/coverage/derivative bounds stay unchanged. Thus no action, body, assertion, identity, control or test behavior changed. Original format failure remains historical evidence; this source review does not claim any successor command passed.

The mechanical closure introduces no actionable finding. The no-substantive-blocker conclusion applies to this exact final successor. Runtime/council/CI gates remain root-owned as stated above.
