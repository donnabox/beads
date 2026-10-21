# Independent installed-data audit

No actionable finding. Saved positive and intended distinct-key negative outcomes match the exact clean candidate and reviewed callback. This is local evidence acceptance only; focused code council, exact-head CI and publication remain separate root gates.

Audited `18f5428686dd84fb139c9dbbcbe39d5958012323` / tree `9c78087e1036dd26ca908e1d8feaa91263513717`, sole parent `98c600767e8e91a79e109cb741beb91e23b9bebe`; exactly the unchanged two authored files. I helped author the accepted plan and performed the native full-source review; I did not author this installed test or execute its cohorts. This audit used read-only standard-library parsing/hashing, archive/Git reads and process inspection. No product import, test, service or rerun.

## Actual results

All20 sequential start/exit records match command argv, cwd, expected exit and unchanged source hashes. Nineteen exits are0; `installed-second-key-distinct` is intentionally1, retained as1, and its separate adapter exits0. No timeout, forced cleanup, supervisor error or source mutation.

Positive: exactly one named callback passed in298.658ms; six completed200 responses on six distinct sockets, maximum2 live. All15 performed checks completed, including whole initial terminal-pair equality. Requests1/2/3/6 have equal721-byte receipt fingerprints (`a9189a0cc94a87f1fb85244801abbc380e76acd8362ed3d5fd67e6e5f49a7f89`), one common receipt/transaction/Resource/revision, effect=required=pos-1; all six serving contexts arepos-1. Complete Resource collection and Event checks executed; respective response fingerprints are263/651bytes.

Control: exactly one named callback failed in277.432ms after exactly two completed200 responses. Own keys differ; receipt IDs, transactions and Resource IDs differ; effect/required positions arepos-1 andpos-2 and each serving position is at least its own effect and at mostpos-2. Six prerequisite checks completed before the sole `same-key-overlap:shared-receipt-id` AssertionError at test685. Primary failure1, cleanup failure0. Earlier transport/schema failure did not substitute. Raw process/supervisor exits remain1, adapter acceptance0.

Both observations contain22 ordered pair milestones. Both invoked/end-called events precede first settled response; both connect and write-finish events precede first response-start in these actual samples. Distinct sockets and complete responses are observed. Both initial receipts were terminal: no pending branch or server pending-join was exercised.

## Provenance and cleanup

The preceding read-only pass rehashed every historical build input671, package output312 and emitted file440 against current bytes. Build remains attributed to dirty98c600; root bridge `d82d6a94e17591e360b3e22f1500d4237de862c0f134b5327ba502ce583aea34` establishes byte equality after ordinary18f5428 commit, without relabelling the earlier build.

Independently compared all313 archive members (protocol63/server250) with actual external installed bytes and source outputs. Only expected server package.json workspace:*→0.0.0 rewrite applies; shared LICENSE explains312 unique source inventory entries. Archive hashes: protocol `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`, server `6ca3ee46dbd61c130ed3b69d0c6a3b4242d356d4e72de4eb7eb0282843285303`. Consumer lock `f334e1cc232a3026bce514d3c6afdceff98a551934c70e6b42207ffe8cc5874d`, manifest and settings hashes match. Exactly eight production package versions and every declared dependency edge resolve inside the external consumer: protocol/server0.0.0, ajv8.17.1, ajv-formats3.0.1, fast-deep-equal3.1.3, fast-uri3.1.5, json-schema-traverse1.0.0, require-from-string2.0.2. No workspace fallback.

Reproduced actual path-sensitive recursive payload hash without importing product: `c46a251cdf51f274ec9bf4f0f401cedadcbaa7748cfb059d3009de34186753b9`,14 package visits/473 file visits; both child reports match. External entry SHA `bcf153170f953f9e28c6c89f1153be907f64df289e64f3ad929cad4a4a54c47c` and consumer-contained realpath match. Actual served identities remain distinct from separately named credential-free projection.

Children58489 and59018 each exit0/no signal and were sampled. Both owned store directories are absent and both runtime temp roots empty. Independently verified all81 sampled/reported PIDs and all20 supervisor groups absent at `2026-10-15T20:01:01.644825+00:00`. Ten other stage temp roots retain only node-compile-cache. Supervisor labels these entries unclassified; this audit classifies them as preserved tooling caches, not runtime leaks. An initial audit-only predicate incorrectly required that field empty, stopped before writing a report, and was corrected after inspecting the actual entries; no product execution was repeated. Sampling can miss short-lived/detached descendants.

## Evidence pins and limits

Installed provenance `f017b438df13d2b8bfe594aaa31b674b00e58b2008179b4afc79f2c57f6404c8`; tar/source provenance `7370c480eb33a44c997a433ede2154df6b40b2a9bf05752ae268311b1d816c23`; build inventory `7ad9c4ea84f900c395b22664927295b81e252d280e6ad0fc11c1ad7c15cd156f`.

```json
{
  "positive": {
    "responses": 6,
    "child": 58489,
    "caseMs": 298.657667,
    "observedMs": 297.297084,
    "milestones": 22,
    "performed": 15,
    "observationSha256": "4c5885527eb753386c2259108ed6358021e990e1bc35bac2c04d34987413cf1d",
    "resultSha256": "0419ea9919798f736ebc7c5ef06011126dc540688c485f4f10fd74a793938a2a",
    "logSha256": "01c21171fcfcd0007dd9a2cef3efd58ab0c9e24460b2fd3b49f4598615b72527"
  },
  "second-key-distinct": {
    "responses": 2,
    "child": 59018,
    "caseMs": 277.43208300000003,
    "observedMs": 273.90345800000006,
    "milestones": 22,
    "performed": 6,
    "observationSha256": "b65e909899c2a966a41e7365695e4e3c033204ff5de41c76c7c72af2a40ec7f2",
    "resultSha256": "a809462f5b01c348efd335eef841d1dcf486670bd8b0c34ddcfeb56a26ea807f",
    "logSha256": "30c9684ac581c864a0a3d8727cfc78e326f7d0e6fbf688b4ed734adefdb016be"
  }
}
```

Raw logs remain unchanged. Their single safe stdout objects agree with observation JSON except elapsed serialization time. Full receipt/Resource/Event comparisons and schema validation occurred in memory; saved hashes/lengths are fingerprints, not retained complete wire bodies. This audit neither reconstructs them nor reruns schemas. It establishes fixed-fixture client dispatch/lifetime overlap and two-key counterfactual sensitivity, not server concurrency, internal pending join, universal atomicity, singleton retained replay or profile qualification. `claimEligible` remainsfalse. Caller sourceCommit is externally bridged attribution, not child-attested Git identity.
