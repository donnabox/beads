# Materialization closure followup

Base `d907a1e1576c5795c17edb5f15528f47edfd7d6b`. Checkout `/Users/dbox/repos/wt/vickie/vickie-materialization-closure`, branch `codex/vickie-materialization-closure`. Frozen clean head: `3148287745c351574752a89f4407c54305aba361`. Only two test files and relevant progress prose change; no production tuple, batch semantics, store/runtime/service/HTTP or Event test edits.

## Outcome

The newly measured **31 explicit owned updates + 32 set matches** workload passes the predeclared 120-second/1-GiB supervisor gate through execution, replay, receipt/page/Event reads and reopen. It reaches **886,912 KiB child self maxRSS / 832,480 KiB supervisor sampled peak**. This is appreciably less margin than set-only, not a universal heap or all-workload headroom claim. Root received this result before any consideration of another tuple change; tuple remains M32/E256.

All new runs import/assert the actual emitted tuple and record it with source and emitted-byte/script identities. Every exact named script/helper/supervisor is preserved alongside its output. No retrospective provenance is asserted for overwritten historical scripts.

## Dispositions of all eight findings

1. Correct present-tense finite group bound from253493248 to169345024. The larger value remains only in the explicitly superseded M64 table.
2. Correct the permanent256-Event claim to past tense under the then-current M64 tuple. Current128/129 lowered-cap observation is separately stated.
3. Execute mixed full operation-budget plus full match-budget case. Seed63 large owned pairs;31 Links have n=1 and are explicitly updated to2;32 have n=0 and are selected by updateWhere and updated to1. Actual request is65536 bytes total. Explicit operation-array serialization3287 bytes; remaining set patch12403 numeric values,272985 normalized bytes. Source sizes262144, Link sizes261849–261850. No impossible full-Q-per-explicit-operation assumption. Receipt64 entries (31 explicit +1 summary +32 members), group50568556 bytes, receipt16507992 bytes. Assert actual explicit and set afterstate, retained receipt digest, page, finite Link Event and corresponding reopen outcomes. Root notified of the result; no further budget edit.
4. Every new stress imports TRANSACTION_RUNTIME_LIMITS from the same emitted batch module used with its owner; asserts matches32, Events256, operations32, R262144 and request65536 and logs the entire actual tuple. Provenance records Node version, root/source commit/status, source batch/store hashes, emitted server/protocol JS/JSON file hashes and combined digest, exact script and shared helper digests. Compatibility imports both old/current tuples, asserts64/32 and logs actual values, not hard-coded labels. It records matching before/after database digests and original/replayed receipt digests; it does not print a fabricated before-maintenance flag.
5. Each fresh variant has an immutable named script: mixed-m32.mjs, dense-m32.mjs, noop-m32.mjs, compatibility-m64-to-m32.mjs, plus provenance.mjs/supervisor.py. Supervisor creates logs exclusively and records script/supervisor digests; child and supervisor exits are separate. Old intermediate outputs remain historical, but their overwritten scripts cannot be reconstructed by declaration. New noop script removes the dead dense branch. No older outputs/scripts were changed.
6. Replace the wrapper loop with one exact two-entry array: Link upsert equals the independently read Link record, source upsert equals the independently read source record. This rejects missing/extra entries and wrong records, not only unexpected wrapper fields.
7. Pin all four normalized numeric literals:1e20→100000000000000000000, -9e20→-900000000000000000000,9e9→9000000000 and -0.0000010000000000000002 unchanged. Keep factor-seven checks and actual stored Event normalization observation.
8. Explicitly state that E256 remains a validated configuration bound whose default workload refusal is not observed by the128 evaluator seam. Do not spend on a speculative cascading-delete workload or relabel historical default-boundary evidence.

## Actual emitted tuple and provenance

All three fresh stress runs recorded this same tuple:

```
maximumOperations:32, inducedEvents:256, matchedResources:32,
representationBytes:262144, scopeSourceBytes:65536,
requestBodyBytes:65536, propertiesBytes:65536,
diagnosticBytes:524288, diagnosticCount:16
```

Recorded source head is d907a1e1576c5795c17edb5f15528f47edfd7d6b with the test/prose working changes explicitly present in sourceStatus. Production source is unchanged by this successor. Build completed before runs. All three record emitted server/protocol JS/JSON tree digest `73b5a9866729b2254c23d47995ab50c7cb301a27dde923d35c6d31d36236ce1a`; exact constituent hashes are in each log. This is an emitted-byte/source witness, not whole-system or third-party-dependency attestation.

| Fresh run | Elapsed s | Self maxRSS KiB | Sampled peak KiB | Group bytes | Receipt bytes | Exit |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| mixed-m32 |8.073|886912|832480|50568556|16507992|0|
| dense-m32 |4.015|551504|523040|35207784|8385225|0|
| noop-m32 |0.428|288544|261840|0|8392388|0|

Supervisor samples every100ms; its sampled metric, not the child's high-water, drives the1-GiB gate. All disposable directories were removed and no stress child remains. No inspect() or full-history database read was used; only the already-approved post-close scalar length query for one actual effect-position group. The finite observation is a Link Event page, not exhaustive Scope projection or replica qualification.

Compatibility uses the actual old emitted owner at source8569b85d785f96325ca948580106f454c48ed034 (tuple64) and the new emitted source above (tuple32). New open refuses installation mismatch, before/after file digests match, and old owner successfully reopens and replays the exact serialized receipt. Both tuples/source/module/tree/script digests are in compatibility-m64-to-m32.log. Current success is derived from executed assertions and observed digests, not nominal labels.

## Verification

Pinned Node24.16.0 / pnpm11.20.0. Frozen install and build pass. **51 tests pass in2 files** (transaction-owned and transaction-materialization), including actual emitted-process cases already in the owned suite. Typecheck, dependency boundaries, format check and lint all exit0. Fresh mixed, dense, noop and compatibility runs each exit0. Actual command exit files are beside each log; machine-readable focused.json records the test results.

This successor changes tests/prose only, so no repeat full suite or packed smokes is claimed. The preceding integrated source's broader evidence is separate. No tests were stopped or failed in this slice, and no public evidence artifact was regenerated. Frozen source checkouts were not edited; no outward actions occurred.

## Remaining limits

Passing the mixed case removes the specific missing explicit-plus-set measurement, not all heap uncertainty. Complete collection materialization, larger accumulated graphs/history, other patch/object shapes, principal data, transient allocation copies and sampling gaps remain. The high measured mixed RSS is reported plainly. No public profile/capacity or full normative-row claim follows. Keep current gate/proof/default-boundary distinctions and preserve historical output-only reproducibility limitations.
