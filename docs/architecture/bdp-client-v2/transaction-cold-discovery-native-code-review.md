# C3 native code review — b20c929

## Finding

**P2 — Preserve pagination callback receivers when freezing configuration.** `packages/server/src/transaction-service.ts:76` creates a frozen spread of `supplied.pagination`. `createAuthorityReadPlane` explicitly binds both `clock` and `generateOpaqueToken` to its supplied pagination object (`authority-read.ts:395–397`); the new caller therefore changes their receiver from the original configuration to the frozen copy. Previously supported stateful methods now fail. This is a regression in the private service’s existing Read path, independent of C4/public discovery.

A bounded differential probe uses a real owner and `clock() { return this.ticks++; }` with a numeric `ticks` property. The exact-parent service returns a successful empty Bead collection and increments original ticks1000→1001. The frozen C3 service instead returns observed503 `temporarily-unavailable` and leaves ticks1000. Assignment to the frozen callback receiver throws and becomes the existing pagination failure response. The same receiver change applies to token methods; the executed minimal reproduction exercises the clock, not token allocation.

**Minimal correction:** capture the original pagination object and callback values once; preserve their original receiver with bound callbacks while freezing the scalar configuration. Keep all nested-limit capture/enforcement safeguards. Add a stateful clock/token receiver regression alongside the existing mutation-after-construction tests. Do not broadly redesign callback semantics or relax the immutable limits. Recheck invalid callback handling through the existing validation boundary.

Proof: `transaction-cold-discovery-native-probes/pagination-receiver-probe.mjs` and `.log`. The counterfactual holds all compiled modules constant and substitutes only the exact parent service module. Parent service TypeScript from aebf656 was byte-compared with prior frozen0fafcf1 source whose retained compiled module supplies that parent variant. Both actual SQLite owners close and temporary stores are removed in finally. This is an isolated behavioral differential, not a claimed execution of the entire parent branch or a public deployment outage.

No additional actionable finding in the reviewed C3 delta. The accepted page/Selector-only candidate is expressly not wire-ready; diagnostic limits and public activation remain C4 prerequisites, not defects silently added to this slice.

## Exact scope

Reviewed all14 changed files in `aebf656506d187ed226302b68b526d34088537c6..b20c92914b48e66b33b66e9e7434dbc3467fad10`, frozen tree `6a84e1cc2916ee5fb4cc49ad7cfd2a65562ef260`, checkout `/Users/dbox/repos/wt/vickie/vickie-tx-discovery-assembly`. Confirmed clean source and every14 file byte/hash binding in author source.json. Read accepted C3v3 plan SHA1ff0c0cdd954249bb2ad743614c7aa7e4b4d9dce89298e4954d78d9b34e6a0ae, root acceptance, author report, source, summary, exits and frozen pack harness.

Additional source inspection covered snapshotProtocolRecord/json-values capture, RU discovery types/parser pattern, authority-read configuration validation, owner withRead lifetime and service run/close. This is a complete C3 diff review with bounded dependency inspection, not an independent review of the inherited runtime or all protocol clauses.

## Reviewed invariants

- Both new protocol parsers snapshot detached/frozen JSON before lazy schema validation; canonical Scope and all nine navigation members are additionally parsed. Optional capabilities remain accepted structurally; private candidate omission does not become a generic parser prohibition. The new exports are the approved two parser/two type seam.
- Owner observation copies only epoch/view/position inside one real synchronous entry. Assembly and lazy parsing follow that entry; refused entry and structural validation failure remain distinct. run/check/close preserve owner lifetime. No reader escapes.
- Five advertised limit scalars plus private TTL are captured once and validation/enforcement use that capture. Candidate omits TTL and uses the captured values; callback receiver defect above is the remaining preservation issue.
- Operation map is shared with HTTP target construction. Independent literals pin all12 targets, actual route classifiers pin non-operation roots, and canonical handler Scope must equal service Scope.
- Floor pos-0 is distinct from checkpoint genesis. Real source tests retain complete group pages through completed-receipt detail expiry and durable reopen; the result includes unchanged epoch/view/head/floor and exact original pages. No runtime pruning scan or codec rewrite was introduced.
- Private routes remain refused, authenticated versus anonymous behavior stays separate, and protocol structural parsers do not authorize public serving. Doc wording correctly discloses this.

## Independently executed checks

Pinned executable `/Users/dbox/.nvm/versions/node/v24.16.0/bin/node`:

1. `transaction-cold-discovery-native-probes/installed-parser-probe.mjs` — exit0,9 named checks against the author’s actual retained installed protocol/server packages: valid parsers/detachment, getter single-capture positive and hostile negative, canonical Scope/nine URLs, exact directory rejection, valid optional capabilities/freeze, nested mutation isolation, forbidden retention.idempotency, public-root/private-deep-import denial. No server, child or network request.
2. `transaction-cold-discovery-native-probes/pagination-receiver-probe.mjs` — exit0 because it asserts the differential reproduction above: parent success/current503. This is **confirmation of a defect**, not a passing product test. Two disposable owners, no server/child/network; finally closes/removes both stores.
3. Python source/evidence audit — exit0: exact HEAD/tree/parent, clean source,14 file hashes and both retained tarball hashes matched; author TX child PID57850 absent. Results and exact digests are retained in `transaction-cold-discovery-native-probes/audit.json`.

No author checkout edits, builds, test-suite reruns, outward actions, credentials, live/shared databases or artifacts were changed. The first parser-probe version also passed; its final retained version additionally directly checks absent root exports alongside deep-import denial.

## Author execution evidence inspected, not rerun

Actual JSON reports confirm213/213 combined tests,246/246 Read/RU regressions,1 fresh packed TX test, and2 fresh packed RU tests with200 intentionally filtered. These results do not eliminate the reproduced receiver regression, which existing closure-based callbacks do not exercise.

The frozen external pack harness verifies clean exact HEAD before/after build/pack/offline install and supplies explicit installed entry/source variables. TX log contains12 completed raw exchanges/one child: six500 authenticated discovery refusals, three401 anonymous refusals, batch200, independent ResourceGET200 and exact retained retry200. No emitted-only substitution: the entry points into the retained external installation. Source is declared launcher metadata plus independently inspected package/source bindings, not target attestation. Claim flags remain false.

Verified package SHA256s: protocol `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`; server `10cb809a2f74a5b42dd62bd582bf78be76b993a621de3c429153316215d9ec04`. Observed TX payload `da8ccc6f5c5f36074f4f56e15e787705073335e996aa347e2f02f4363f1b9272`, contracts `184330942591336c6ad84ef3d32e2704f67006b8ff64882a300488ab4698608a`, runtime configuration `7335b088e633ec32982476e99bd60f24ed105e089a26fe2115b1af3610671ff4`.

Author report SHA256 `1d3a5ab81ac089a57337c0dcbf5cba8c2d2c3ce84b436c750d9bdd01fd3e7c9c`; source.json `c5f76c188b6931c9b04efd41983e12dd59162cdd4cba59375301df0b8377fe6f`; summary.json `b8cae4efae349797c0aaad5bf3adbd3e78352d222538c9fe93b610f7b56fba85`; exits.json `13570d00d4ed8366f4152a8cd9a1c845514120ce9f5289e153939018a271a9c0`.

The author correctly preserves prior preparation failures and identifies corrected fixture assumptions. No historical evidence was relabeled. No native-owned child process exists; retained probe module copies/scripts are review artifacts only. Root owns disposition, repair and subsequent integration; this report does not approve readiness.
