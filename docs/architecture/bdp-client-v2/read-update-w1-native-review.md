# Native W1 council review

**No actionable findings.** The eight new whole-contract assessments are supported for their cited clauses in the current installed configuration. The ninth new row, alias deletion, correctly remains incomplete. This is not full Read+Update qualification or public admission.

Reviewed frozen `459d32b67e3d8c2c4e4e667bc7481649ca33eee3` against `4156f380af8e5a8e5e74ec29e0221561fbc4df18`. Author checkout `/Users/dbox/repos/wt/vickie/vickie-read-update-w1` remained clean at the exact SHA. Independent build/tests used detached clone `/Users/dbox/repos/wt/vickie/bdp-w1-native-review`, also clean afterward. No source edits or outward writes.

Inputs: W1 plan, native plan review, c562 external review, 4156 external review, and final `read-update-w1-report.md`. Inspected all ten changed files, canonical catalog citations, relevant normative clauses, unchanged exact client/runner/schema-validator behavior, and actual progress observations. W1b/W2/W3 not assessed or authorized here.

## Independent execution

Pinned Node24.16.0 / pnpm11.20.0; frozen install and build PASS. Five focused files **45 tests PASS**, 18.39s: installed matrix, installed Read fixture, installed Read counterfactuals, cumulative artifact binding, exact client. This includes all eight W1 response counterfactuals and existing configuration/auth/guard/no-op controls. Typecheck, dependency boundaries, and changed TypeScript format/lint PASS. Logs: `bdp-integration-logs/w1-native/{install,build,tests,checks}.log`, `tests.json`.

Normal unfiltered result independently reproduced in `w1-native/progress.json`:
- 134 applicable scenarios: 33 pass, 101 not-implemented harness errors.
- 129 required IDs, 33 authored, 25 whole flags, 104 uncovered.
- 33 same-configuration sessions; 199 actual client exchanges, 44 actual seed mutations, 21 native conditional-header observations.
- `counterfactualApplications:0`, `claimEligible:false`, clean source exact reviewed SHA.
- Catalog `03ca39f2…`, manifest `10b233f5…`, fixture `599677be…`, exact route/configuration `700da44d…`, matching author record.
- Emitted payload `d12573ef…`; author's separately recorded packed payload `6503d708…`. Different installation payload identities are expected; no identity relabeling performed.

I did not repeat the author's packed 42-test run or full 836-test conformance run. The focused run exercised actual emitted installed child processes. Process inspection found no test-owned child remaining.

## Whole-contract audit (all 25)

Each entry below was checked against its catalog-selected clauses, not just its passing status. IDs omit only the obvious `read.` or `read-update.` prefix.

| Whole row | Clause-to-observation assessment |
| --- | --- |
| read discovery.scope-service-desc | Successful Scope response, actual Link capture, row-local dereference, successful closed RU discovery JSON. Previously borrowed evidence is now local. |
| discovery.canonical-navigation | Actual discovery values resolve exactly to canonical Scope beads/links/types roots. |
| resource.link-record | Closed full record, independent id/type/endpoints/properties/metadata from real seed. |
| resource.properties-view | Both kinds return exact stored properties alone. |
| resource.default-bounded | Real Bead with incident Links; closed record plus independent identity and absence assertion. Valid-shaped embedded-plane counterfactual fails that assertion. |
| types.inventory | Closed complete id/name/describes summaries, exact independent tuples, order and complete page. Both prior corruption controls remain load-bearing. |
| types.descriptor | All thirteen actual closed descriptors, covering Bead and Link endpoint-constraint shapes. |
| collections.beads | Complete items/next page, independently associated id/type/documents and canonical order, both default links-plane absences. |
| collections.links | Complete closed records, independent endpoint/document tuples and canonical order. |
| collections.invalid-query | Repeated limit/selector and unsupported parameter refusals retained. |
| http.get-head | Retained actual GET/HEAD pairs and body/framing parity assertions. |
| http.method-405 | Actual unsupported method, exact Allow and no Problem body. OPTIONS-enabling antecedent is inapplicable to fixed CORS-disabled configuration. |
| http.cache-cors | Actual scope-bounded representations and error responses carry private/no-store. No enabled-CORS credit inferred. |
| singleton.create-bead | Explicit and allocated identities, mutation postimage compared with actual GET. Existing bounded assessment preserved. |
| singleton.document-patch-merge-replace | Both document-relative paths; whole nested-object replacement by add; insertion and append with array order; untouched sentinels; smaller root replacement and clearing each document; nonexistent-path refusals and full unchanged GET. Expectations are static values, not evaluator output. |
| singleton.document-patch-atomicity | Both valid/invalid directions refuse and retain the complete independently read version; effectful combined update; actor/agent/message/committed context independently asserted before capture. No-op reorders object keys, preserves arrays and supplies different provenance, then directly compares returned record and GET with saved record. |
| singleton.delete-bead | Incident-Link refusal, unchanged state, Link removal, final live identity without postimage, all three Bead read paths absent. Existing assessment preserved. |
| singleton.create-link | Relative endpoints resolve under `/acme/` against Scope; missing local endpoint leaves Link absent; local noncurrent and external URN pins echo exact semantic string values; full literal endpoint/document oracles and independent GETs. URN does not introduce a claim of HTTP dereferencing observation. |
| singleton.delete-link | Effectfully updated Link independently read, old guard refused, live guard succeeds, static kind/id/type plus captured final revision, no postimage and both Link read paths absent. Endpoint records unchanged in unowned fixture. |
| singleton.expected-revision | Retained matching/stale update/deletion and unguarded-current-state legs with independent afterstate. |
| singleton.idempotency-conflict | Changed semantic input refused, no effect, original retained disposition still returned. |
| singleton.result-headers | All eight actual POST targets have both no-ETag and no-Location assertions; includes Resource and alias no-op observations. Alias redirect Location remains separate. Both Resource-header and alias-header counterfactuals fail their intended absence assertions. |
| alias.put-creates | Closed literal created result, canonical target and alias URL; actual bodyless GET/HEAD307. |
| alias.put-repoints | Different keys for repoint and same-target put, updated outcomes, intermediate target witnesses, deletion/reuse and temporary redirects. |
| alias.no-version-minted | Full old/new target records unchanged at each phase; exact closed alias results; no representation; complete Bead collection and empty Link collection exclude membership. `next:null` is asserted, so a partial first page cannot masquerade as absence. |

The three whole alias rows each contain their own full workflow. Duplication does not borrow coverage across rows. Alias.delete-releases stays false: unknown/live-query404 and reuse are observed, fragment targets are not. The no-aliases antecedent is correctly identified as inapplicable to RU, which requires alias discovery. Effective-schema, owned-Link, authorization, sequence and History obligations are not inferred from this fixture.

## Schema and negative controls

Independent `w1-native/schema-probe.mjs` checked **all 158 original `$defs` and all other canonical root values unchanged**. Exactly six harness-only wrappers are added. For each wrapper, an actual observed valid response passes, missing body fails, an invalid normative required member fails, and an added body member has exactly the same acceptance as the underlying canonical definition. In particular, RFC Problem extension openness is preserved; it must not be mistaken for a weakened closed Resource/result schema. Probe output: `schema-probe.log`.

The first draft of this reviewer probe incorrectly assumed Problem objects were closed and failed on an extra Problem member; correcting the probe to compare canonical openness resolved that reviewer assumption. No implementation defect or source change resulted.

W1's eight committed counterfactuals each execute actual HTTP first, preserve the original observation, then substitute only the assertion input. Each executes exactly once and fails the named assertion: array order, entire retained envelope, local pin, deleted final revision, alias target, Resource ETag, alias Location, closed mutation result. These controls do not claim altered bytes originated from the authority. The independent schema probe extends sensitivity checks to every wrapper kind, beyond the committed mutation-result example.

## Prior review dispositions and boundaries

All five c562 lows are addressed: default-bounded identity/schema plus absence control; row-local discovery dereference; semantic/raw paths honor exact default credential with real anonymous/operator wire tests; redundant handwritten Read route list removed in favor of finite derived/native and required client routes; seed comparison accepts JSON object reordering and rejects changed values. No arbitrary target or credential surface was introduced.

The three 4156 follow-ups are addressed: both collection items reject injected links; proxy framing is explicitly reserialization, with prohibited synthetic204 Content-Length removed; launch is inside cleanup scope and parse/mutation failure returns502. The latter cleanup/error branches were inspected, not newly fault-injected in this review.

Configuration identity still comes from actual installed capture/ready values, compared across all scenario processes. Credentials remain separately bound from the exact route digest. The unchanged exact client preserves real serializer/body observations and only allows fixed routes or validated canonical captured Bead identities. New Link and alias observations use finite fixture routes. Captured operation-directory values are used to address all eight mutation targets. Source SHA remains launcher/reviewer provenance; the target does not attest it. The wrapper helper is reviewed harness source, not silently added to target payload identity. Public admission, production evidence, normative definitions and runtime semantics remain unchanged.
