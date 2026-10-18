# Independent native Selector code and measurement review

## Findings

1. **Medium — New real-owner spy adapter fails typecheck.** `packages/server/src/transaction-submission.test.ts:1320–1321` casts the Vitest Mock directly to the overloaded `store.admit`. Its inferred last signature requires five arguments, while the target includes the three-argument overload. I identified this statically; root's separately authorized compiler run subsequently confirmed sole error TS2352 at1321:41, exit1, source unchanged (`typecheck-first.log`/`typecheck-first-exit.json`). **Correction:** preserve the real spy and zero-call assertions, but use the already established plain delegating wrapper around `admit(...args)` at the boundary (examples1150/1200), then assert that plain function as the overload type. No unknown cast, fake owner or runtime case is needed. Preserve the original compiler failure and freeze the test-only successor before execution.

2. **Low — Two accepted UTF-8/identity assertions were omitted.** `read-selector.test.ts:491–504` only checks short é, JSONPath backslash-u and percent metrics, plus equality of decoded string summaries. `transaction-set.test.ts:1037–1060` checks percent and whitespace identity differences. Neither supplies v4's multibyte/astral4096/4097 boundary specimens nor compares normalized transaction identity for outer JSON raw é versus outer `\u00e9` spellings. Existing Read Unicode tests exercise the old entry point, not these complete-analysis/body-identity assertions. **Correction:** add the independently byte-sized é+astral specimens to the existing analysis case and prepare/normalize both outer JSON spellings in the existing preparation case. Keep case/HTTP/owner counts unchanged; preserve different JSONPath source spellings as distinct semantic identity. This is a bounded missing-test correction, not evidence of a production parser defect.

**Assembly correction verified closed:** the intermediate requests.json arithmetic note used T6=253/depth13 plus a compensating extra wrapper node. Before declaring the packet frozen, the author corrected it to T6=254/depth14 and g added nodes/depth, with no node for outer filter delimiters. Final requests SHA256 `be8877eee55f12eb5d6895031a98e219652c117c8ed8337d3d541fed60b97c4a` has the correct explanation; its final metric values and request ledger were already correct. This is not an outstanding finding.

## Reviewed outcome and release boundary

No additional actionable production defect found in the frozen twelve-file diff. Complete grammar runs after complete carrier shape validation and before key/principal admission; empty primitive Selector reaches typed grammar errors with original singleton/batch pointers. The unchanged private whole-JSON depth check follows grammar. Summaries are frozen and WeakMap-bound to prepared set objects; decoded source spelling remains in identity, and no AST or summary is serialized into it. Missing summary is an internal error. Read still uses byte-first checking and the same five early AST construction checks. Fixed4096/32/256 capacity is consulted only after prior-key equality/conflict and operation count, before negotiation/receipt insertion/execution; observed refusal context is captured at owner entry. Current sequence-set exclusion and unrelated failed-receipt semantics remain intact. The spec distinguishes body versus query diagnostics and preserves the command-minimum rule.

**Current draft must not begin source runtime validation:** finding1 is an actual compiler failure and the accepted test-only repairs are pending. After corrected freeze/native delta closure and root's measured release, this review identifies no additional production blocker. This report grants no execution, control, package or installed release.

## Independent workload/command check

Read full `draft.patch`, relevant parser/preparation/normalization/owner/submission/HTTP code, all new/changed tests, supporting fixtures, and accepted v4/root-authoring acceptance. Read the completed report and packet artifacts listed below. Compared all twelve on-disk hashes to the frozen manifest before and after review.

The explicit anchored command patterns match the extracted literal full titles exactly:67 parser +6 set +13 submission +4 HTTP =90 source-positive cases. Controls select one named case each; installed selects its sole named case. Nested describe prefixes and both expanded carrier titles are included. This was data-only pattern comparison plus source inspection, not Vitest discovery or test execution. Actual execution must still verify selected result titles/counts rather than trust this model.

Independently recounted source HTTP: legacy37 =6 wrong-method +24 authorization/reuse/read +4 grammar +2 shape +1 late-shape; other affected legacy5; H1=2*2*3*3=36; H2=16. Total94. Legacy15 receiptRows calls each close/read/reopen, yielding16 owners/listeners on one directory; the other three fixtures make19 owners/listeners on4directories and17 readonly inventories. Eleven regression cases contribute14 owner opens/10directories (including four reopens); six new/revised submission fixtures add6/6. Total39 owners/20directories, not a count of logical admissions. Source fixtures register cleanup and readonly handles close in finally.

Controls add2 owners/listeners/directories, expected2/max52HTTP and expected0/max2 readonly handles. Combined source is41owners/21listeners/22directories, expected96/planned max146HTTP and17/max19 readonly handles; the separately proposed positive100 plus control52 ceiling is152, not an assertion of146 enforced by the transport. These are prospective structural counts.

The installed literal ledger matches25 source calls:19POST/6GET, total35865 body bytes, max20054 at the deep invalid request. I recomputed every recorded body length and SHA256; all match. B/D/N final metrics and byte lengths agree with independent arithmetic including the corrected explanation in the final packet. The sole child uses one private directory and no reopen/read-only inventory. The25-exchange fence, same-origin receipt and actual continuation validation, exact two-group/Event/postimage expectations, successful-key set, and finally PID/directory cleanup are present. Dynamic GETs are observed grants, not guessed routes. The fixed Scope is independent of loopback port. These are private bounded serving checks, not public-profile or unseen database-inventory proof.

`commands.json` pins Node/pnpm paths, one-worker explicit test argv,120s outer caps, and separates source, controls and installed packaging gates. Current setup/typecheck `executed:false` fields describe the proposal ledger, not root's later separate failed attempt; preserve that actual failure and do not blindly rerun the entire ledger. Package/consumer paths and final SHA remain gated placeholders, and are not yet a complete actual provenance authorization.

The two counterfactual files are explicitly **source-shaped illustrations for later emitted-JavaScript patches**, not executable source repairs. C1's illustration lacks the TypeScript-required development-limit message; C2 leaves an unused source import. Neither should be applied to source and compiled. The current gate correctly requires actual emitted before/after hashes and module-resolution binding after a clean positive build. Preserve the positive runtime graph and select the exact first-status oracle: C1 actual500 with development marker versus400, C2 sanitized500 without marker versus413. Exit1 alone is insufficient. No derivative graph or emitted patch has yet been reviewed as an actual execution artifact.

## Exact basis

Workspace HEAD `757aedc8008386745d59d5005498a9256d6ea76e`, parent tree `734bf9daf40fd17d0fe2bf86b085598589122f2f`; uncommitted twelve-path draft. All twelve source hashes remain exactly those in the manifest. Plan v4 SHA256 `55f240fd2dfcdc48d142ab9e9d4366068829d88cf2e3880065313f4d169cb5cb`.

| Evidence | SHA256 |
|---|---|
| source-binding.json | dca25dfc3d64137519d9eb636442d6f2d1fefdcff0954b4a793065644d691685 |
| draft.patch | 349cc7e3e66d4c71168af9de95c6034673909827813edcd863ce185ef650ea16 |
| report.md | 53bb94d70174f423644071ff33230fae1de9e72d775c59987596859a93dd536e |
| commands.json | 4f631cea1dffd38b0c1550eb88f0d0e7d73bd745f6da67b641f838f74cfefb67 |
| controls.json | aaf73f9715941a2d04d911343ba1fbb73b6bf2bba286fcfb2b09c83ba771c323 |
| requests.json | be8877eee55f12eb5d6895031a98e219652c117c8ed8337d3d541fed60b97c4a |
| selection.json | 45ee8d720f164f4ce6c42f9eb5b12782e6952121443483f852a721f5acce5ef6 |
| measurement-summary.json | c49bdcd1d50b0b0b9d4f17ebdd721882222bc385a02688d4cf34798a188b5bfc |
| grammar-development-source-proposal.patch | 749ff6b9103e37648786bad163e4c1233af101fe487b28a30c1a1882c6882716 |
| omit-fresh-capacity-source-proposal.patch | cd4703ba7861a85c20bedd414782a17cc45c84e72d5d6ad7b95371be65e9912d |
| typecheck-first.log | 6fdbf504fc5d4010250a2654d14ac8e2632bf657b29433595f3d249c2febcb33 |
| typecheck-first-exit.json | 464b002a4bb125f90b80f1210b7d060f5e2e41ac3296bf32ec55105b429c4983 |

Reviewer ran only read-only file/Git inspection and data-only JSON/hash/arithmetic/pattern checks. No product imports, tests, builds, installs, services, external reviews, source edits, pushes or coordination posts. Root-owned typecheck evidence was inspected, not rerun. This guarded report is the only reviewer write.

Final packet freeze audit: every entry in packet-hashes.json matches its file. The preliminary native report is preserved as native-code-review-before-packet-final.md (SHA256 de20d747f8752f8106adbc29ba2eb2fefc0ee4ede74ee99c219dbb454c100d9b). Only the assembly correction above was closed; findings1–2 remain pending.
