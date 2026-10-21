# Supplemental measurement inventory v2

This corrects an inventory omission only. All frozen source files and original reports, measure.py, measurements.json, binding.json and draft.patch remain byte-identical. No guard or source changes were made. No tests, builds, installations, services or application code ran.

## Exact omitted transformation

apps/bdptest/src/transactional-matrix.test.ts:921-935 registers the duplicate receipt-binding control. Its manifest callback replaces atomic-commit/a-point captures with binding commit-receipt, json-pointer /id. This changes the manifest even though the artifact bundle rejects it. observe applies the callback and constructs/validates the cumulative bundle at lines119-127, before temporary installed directory creation at128 and any installed launch. The test expects scenario manifest: validation failed and zero children. Therefore the source model assigns zero service launches and zero HTTP requests to this invalid-bundle invocation; these are not newly executed observations.

The supplemental data-only script independently reproduced exactly that capture replacement. Normal stringify bytes:695793; compact execution bytes:339650; nodes:18112; actions:199; assertions:2102. Canonical overhead is195 normal bytes and5 nodes. Normal derivative SHA256:d799eada4cf3794c2c72c2a0628896c406b683d52f0bd93a70201941671b3a2b; compact SHA256:1cbbf136b47895d3e24ed30a332dd9c6d743569eead0063682d0bea15ac3e9b2.

## Corrected complete inventory

38 mutated-manifest invocations +14 unchanged-manifest invocations =52 observe invocations. The unchanged14 are metadata5, abort/preabort3, deniedall1, omittedroutes2, deniedset2 and positive1. The original inventory's abort/preabort/duplicate4 group incorrectly counted this invalid manifest among unchanged bytes.

All other37 derivative entries are retained without modification. Joint maxima remain701716 normal stringify bytes and18263 nodes; canonical counts, coverage, route digest,59 expanded app cases and52 observe invocations remain unchanged. The additional derivative is below both736794-byte/19129-node hard stops. No service/HTTP cost is added to the existing source model because this was already one of its zero-launch invalid-bundle invocations.

## Original-to-corrected bindings

- Original measurements.json SHA256:316923ef6922c943ed848d67085956033d7b620030a3245c4796079c569881e3 (preserved).
- Corrected measurements-v2.json SHA256:a827bebbdab479c119d3425e16a079d3bcd9cea6da09c9ea30d95313e8d96c87.
- Original measure.py SHA256:050d0c93e40ee24da00589b11f83ee5f661070e18be6dbd7d3fe417b8f5274dc (preserved).
- Supplemental measure-v2.py SHA256:f2576954abef33c24545276a4ff255c159d68cb67ae9728816f8a118b6866dcf.
- Original measured-report.md SHA256:570ab0884523a2a6018961a599fb075c21b65b593cca836dc743e5007a5826b5 (preserved; this supplement supersedes its37/15 inventory statement only).
- Original binding.json SHA256:6cb16804ef318023dcb764bbccd0f2c6534db2062658aefc9c86093acf50c686 (preserved).
- Original draft.patch SHA256:2c74b73049bd2a5be07698199faac8b011f403f2155286d405c24b101652fbb2 (preserved).

The original source binding still applies unchanged. Both pre/post hash checks passed for the original artifacts and all four source files. New derivative is derivative-duplicate-receipt-binding.json. Root still owns any guard amendment and execution release.

An initial supplemental-script authoring command had a quoting syntax error before writing any file; the immediate attempted script invocation found no file. Both were evidence orchestration failures, not application executions. The corrected script then produced this supplement once.
