# Ordinary Update plan — actual PR137 merge reanchor

No reanchor blocker found. All 23 committed pins were compared from 73fc0daed2e5fa6843e581de9bf7a46e519562ab to 67a2842e3f1918e70978f745587cf11eaf0df5a8: **22 unchanged, specification only changed; all six proposed source paths remain byte-identical**. This is a reanchor audit, not authoring or execution authorization. The active actual Claude review retains its frozen inputs.

The fetched merge object has tree 596abe870220fb61ce606f8382a9da52d554d063 and ordered parents 73fc0daed2e5fa6843e581de9bf7a46e519562ab, c18c71202c90fa2558f34107ad342befabf750fe. Its tree equals the reviewed integrated candidate. The first read preceded the other worker's fetch and reported unavailable object; the final audit reads the now-present object directly. No fetch or repository mutation was performed here. Retained API/remote verification is separately attributed and hashed in the JSON.

## Exact delta and applicability

Only docs/specs/bdp.md changed among the 23 pins: f036103ffdf5f89fa3a2033570e1e987c4cd203cd80b1c633cb44615123f892e → 63e93be562334c84c1e967da2fce212e8ef515aa3c9448924c5f9a272ea7b80b. The JSON retains the complete specification diff and every before/after file hash. PR137 changed 11 paths overall; the other ten are Transactional implementation/tests and historical progress, outside the proposed six Update paths and other pinned dependencies.

At merged spec lines 4627–4657, Transactional directory/receipt/page status tables now explicitly include malformed-minimum 400 and receiving validation before disclosure lookup. Lines 5819–5832 require all three fields exactly once, nonempty and lexically valid; receiving refusal has no projected context, cannot become 304/412/406/SSE, and HEAD retains GET metadata without body. Existing semantic consistency decisions remain distinct. The enclosing profile boundary at 5772–5776 permits ordinary Read/Read+Update to skip minimum-position/strict replication behavior; the request paragraph at 5810 explicitly identifies a Transactional client. The existing mutation-command ignore rule at 5841–5845 remains unchanged.

The proposed scenario explicitly requires read-update. Its ten raw HTTP actions contain empty authored header lists. Its eleven client actions use the unchanged exact-client construction (apps/bdptest/test-support/read-update-exact-client.ts:129–143): GET emits no authored headers; POST emits content-type and optional idempotency-key. None sends the three minimum fields. No new Snapshot, receipt or Transactional context obligation is imported into this ordinary Update witness. The new rule requires **no scenario, expected status/body, control, count, route-order digest or workload change**.

## Preserved plan and next binding

Frozen plan SHA 116b2350049f30ddab90248152691256b2608aba61f34f27a205ebe018633ecb and native closure 1df153e3685256a5b0db7790d9ba64df3514e228a6e9de365c83dbf2ccf0b33d remain unchanged. All six canonical/test paths, fixture/continuation/index, schema, runner/executor/client and ordinary installed composition pins match the original committed bytes. Historical raw-object preservation and measured data/route inputs remain anchored to identical source bytes; this is not a fresh runtime measurement.

After disposition of the active plan review, create a successor authoring binding to 67a2842e3f1918e70978f745587cf11eaf0df5a8/tree 596abe870220fb61ce606f8382a9da52d554d063 with the new specification hash. Preserve the existing 73fc review packet and evidence chronology. No runtime caps, semantics or scope expansion is needed for this reanchor.

No product commands, tests, builds, services, imports, extra reviewers or source changes were run. Artifact JSON SHA: 1472b702972da0718a7c1b42dac3bfdb4babe11232b36052ac842be08d223c47.
