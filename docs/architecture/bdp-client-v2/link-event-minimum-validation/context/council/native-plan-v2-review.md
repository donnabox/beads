# Native Link plan-v2 review

Verdict: the bounded six-file design is feasible against the pinned source. One exact header-oracle clarification needs disposition before authoring; no additional product decision or missing fixture dependency found. This is a read-only plan review, not implementation or runtime verification. All execution/review/authoring gates remain root-owned.

## N1 — distinguish the private limit marker from the non-attesting marker

Plan D says “no ETag/Last-Modified/development marker.” Taken literally, the final phrase is incompatible with the existing runtime: `packages/server/src/transaction-runtime.ts:92` unconditionally sets `x-bdp-development: non-attesting` in `deliver`. The current Bead pair helper at `transaction-events.test.ts:1761–1763` instead requires `x-bdp-development-limit: minimum-position` on the excluded private500 responses and absence of **that limit header** on other statuses. Require exactly that behavior in Link D; preserve the non-attesting marker. Do not remove the runtime marker or assert all development headers absent. This is a narrow plan/oracle wording correction, with no request, source-file, semantics, or timeout expansion. The runtime is outside the six-file edit scope.

## Fixture and preservation closure

The optional `finiteEventMinimumFixture(waitCount, eventTypes = types)` parameter is supported by the existing `fixture(extra)` spread at service.test.ts:55. The current default `types` array is Bead-only; retaining that very array as the default and adding one separate Link-enabled array for B/C fixes the real prerequisite without changing old callers. The proposed Link descriptor matches the existing Event fixture descriptor shape; its source/target empty conformity lists and opaque external mode permit the seeded local endpoints. The service fixture Types have no properties schema, so the proposed empty Link properties do not conflict with a required schema.

The installed callback's separate descriptor array and unchanged shared schema are coherent: that schema requires exactly a string `title`, and all three proposed creations supply it. Reusing `artifacts.schemas` does not mutate it. Only the new child receives `linkArtifacts`; the 37,245-byte historical prefix is explicitly frozen (SHA e4510bf2dfdaacac981bf9583a2f6be4c536de5ceca5309c754f93fd56b6f652). No Type implementation changes are needed.

The two old sentinel changes are necessary and correctly distinguished. The old owner case has no Link identity and becomes contextual missing, not proof of known-Link waiting. The old HTTP case genuinely seeds Link l; its pair becomes503. Its total remains37, with503 count10 and500 count12, adding two bounded80ms waits. No historical success evidence can be relabeled.

## Semantics and control scope

The selected extension preserves the existing layered order: HTTP lexical receiving and route/media fences; service waiter policy/clock/abort preflight before one captured tuple; owner canonical parsing and lexical capture before entry, retained identity/authorization before semantic minimum, then current context/minimum before group scans. In particular the service classifier catch must remain restricted to source classification, leaving malformed tuple capture outside it. F's canonical malformed row detects omission of Link capture; aborted malformed detects moving capture before abort; extra-query and relative-source rows retain original owner source-error precedence. Existing selected policy/clock rows cover the unchanged shared waiter.

The owner uses retained `identities` (transaction-store.ts:693), not current resource existence. Link deletion therefore retains created/deleted membership, while empty authorization returns current-context missing before comparing the old tuple. Two own Link Events fit the unchanged maximum2 and produce next:null; endpoint incident Events are projected into Bead sources, not the Link's own page. The positive known-current and retained-page scan checks make the zero-group-read negatives meaningful. One spy is installed after setup, reset before measured reads, and restored finally; identity SQL remains allowed.

The pinned spec permits the reference remaining-read choice at5841–5843 without extending the current Resource-specific MUST. Link own membership5583–5585 and retained unknown/hidden identity rules5665–5676 support these controls. After-bearing, SSE, Last-Event-ID, media and foreign unknown policy stay strict/private; G2 remains a separate unchanged incident-metadata callback. This introduces neither S1 nor G2 policy.

## Prospective workload and lifecycle

- A:11 owner Event reads, two completed transactions/four operations and one separate policy replacement. Setup/page/cursor are real; unknown source and hidden retained source are distinct controls.
- B:two service calls/four delegated reads at0/0/5/10, two waits/probes and one separate three-operation seed. C:four service calls/five reads at0/0/5/10/15, three waits/probes, one seed, one separate post-close direct refusal. Existing helper gates require release after abort and during close; immediate promise observation plus finish/drain is retained.
- F:four isolated stores/calls, four initial clock calls, zero waits, two delegated source-error attempts. No seed is needed to prove pre-entry precedence.
- D:35 exchanges partition into503×8,500×12,400×4,404×3,200×8. The seed/delete are counted HTTP requests (five mutation operations), eight waits, one listener/store. The80ms budget and5s callback remain fixed; actual elapsed determines feasibility at validation, not this arithmetic.
- E:four exchanges200/200/503/503, one genuine three-operation POST, one child/store. Callback-local artifacts, exact allowlist and actual response URL/header/body assertions discriminate real serving. Preserve5000ms request/45000ms owned/60000ms callback and corrected close/error aggregation/conditional removal, including late rejection observation.

Selection has25 existing +8 new =33 distinct source callbacks, then four installed callbacks with11+3+4+4=22 exchanges. Expected emitted traffic128+35=163; total185 across37 callbacks/stores, four in-process listeners and four sequential installed children. These are source-derived bounds, not observed measurements. The previous21 post-bootstrap stages plus the new installed callback/audit pair are23, with two separate bootstrap/tool stages when needed. Each installed callback needs its own immediate audit; no concurrent children or aggregate timeout substitution.

Fresh exact build/binding must precede any test module collection, including direct-focused selections in files with top-level emitted imports. Fresh archives, measured consumer lock, real entry and source/dirty attribution precede external collection. Commit only after all source/installed/cleanup audits, then679-byte dirty-to-clean equivalence and exact-head CI. No execution is authorized by this review.

## Exact reviewed inputs and verification

- plan-v2.md:8ac16c7f6a9ad1cf4cb90217c6bb33c01467b0219648d595ee48629fe14e59c4
- selection-v2.json:101661631b273734e4c252c990449e610de107224d5521314b9d96855dd525e6
- plan-v2-binding.json:ad40368057bca4bd35deb7fccd176ed8d6049e90ed65f159629476d76e8c2aa9
- Source HEAD5bcd277b2a61461211ffc0c2d2f8fd701ef0a385, clean working tree; all679 tracked file hashes and all four frozen context references verified before and again immediately before this guarded report write. Pinned spec SHA67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8.

No source edits, tests, imports, builds, installs, services, Claude execution, CI claims, commits or publication were performed. Remote role session was checked by guarded-action.py before writing this sole report.
