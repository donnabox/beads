# PR64 native Read enforcement repair review

**One blocking finding on the reviewed commit; no other actionable findings.** Exact source: `21522d4573fcc3afd815a8fe68f86ff5a8f6fbbe`, direct child of `c1712f2d26f889f723ffb79c01146c87883a65df`, from `/Users/dbox/repos/wt/vickie/bdp-pr64-read-metadata-repair`. Bounded scope: its 13 changed files, 385 additions/45 deletions. Author checkout unchanged.

## Finding

1. **P2 / required-check blocker — `scripts/generate-read-cohort.mts:81`.** `createCurrentReadWireValidator`, a function, was added as an entry of `IN_PROCESS_ONLY_CAPABILITIES`, declared `Set<string>`. It is an observer factory, not a capability name. The required typecheck includes this script and fails with TS2769: the function type is not assignable to string. Remove only that array entry; retain the import and the correct `validateReadWire: createCurrentReadWireValidator(schemaValidator)` wiring. This prevents claiming the reviewed head passes all required checks. Reproduced independently in the exact-commit archive; `native-read-typecheck.log`, exit 1. Parent notified immediately.

## Reviewed behavior without further findings

- Current record/collection assertions use `currentBeadRecord`, `currentLinkRecord`, `currentBeadCollection` and `currentLinkCollection`. Controlled page checks also use current schemas. The compatibility definitions remain reachable in the sealed projection rather than being removed or redefined.
- Counterfactuals run the actual matrix against a real local HTTP proxy. For each Bead record, Link record, Bead collection, Link collection and inline owned Link, the unmodified or explicitly complete counterpart passes. Removing metadata fails the row and its schema assertion, with no other failed exchange assertion. The exact same observed body remains accepted by the public compatibility parser. Thus the negative controls prove wire enforcement rather than failing for a missing unrelated fixture or silently testing a projected object.
- The client-action test proves the same missing-metadata response succeeds without the observer and fails with it, and counts both modified responses. The observer runs on the transport's bounded parsed JSON body before `ReadSession`'s legacy Resource parser. It does not perform an additional Fetch, clone/read a response body, or modify the public client package API.
- The wrapper forwards URL, scope and AbortSignal to the existing transport and lets that transport retain body limits, timeout and cleanup ownership. Validation occurs after the transport finishes its read/cleanup and only on success. The existing client-action `finally` blocks still close each client even when validation throws.
- URL classification limits strict Resource checks to same-Scope current records/collections and incident Link views. Properties views and revision-addressed historical representations remain distinct. The helper is test support; current wrapper use is wired into both matrix targets and packaged client actions.
- The new observer source is included in the fixed-order harness binding, while packaged bindings already include the cohort generator and matrix bindings include their respective suites. No evidence artifact was edited. Direct schema roots become 16 from 20 because legacy roots are now reached through the strict current roots; the sealed 47-definition closure and digest are unchanged, as independently checked.
- Adding `metadata` to the existing Selector root allowlist exposes the already approved common document without changing JSONPath parsing or traversal. Tests cover both Resource kinds, separation from `properties.metadata`, quoted/slash/tilde keys, negative array indexing, null versus missing members, absence on legacy-shaped candidates, and no candidate mutation. Server selection receives complete Resource records, so this is not only a parser allowance.

## Independent validation

Node v24.16.0; external exact-commit archive `native-read-source/`.

- `native-read-tests.log`: **81 tests passed in 3 files** — reference Read matrix including real wire counterfactuals and full checked-in plans, Read Selector, and Read schema projection.
- `native-read-typecheck.log`: **failed**, exactly the TS2769 capability-set entry above.

No full suite, packaged cohort run or fresh seal is claimed. The source requires the one-line correction and a successful typecheck before the finding can be closed; subsequent evidence qualification remains the parent workstream's responsibility.

## Closure at `3e7b6724fac41277c1d7d0e7220f854930d6af0d`

Independently inspected the one-line child commit: it removes only the validator factory from the capability-token Set and retains the import and strict observer wiring. Applied that exact committed script to the external review archive (the rest of the tree is identical) and reran the required typecheck: **exit 0**, `native-read-typecheck-fixed.log`. Finding 1 is closed. Earlier 81 passing tests still describe the unchanged behavioral source; no full cohort is claimed.
