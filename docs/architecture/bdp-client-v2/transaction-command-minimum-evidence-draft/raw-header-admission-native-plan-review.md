# Native plan review — exact authorization-view header admission

No blocking finding in the frozen narrow repair plan. The proposed exact normalized-name exception, coupled to credential-handle rejection and unchanged report projection, addresses the actual conformance-input defect without bypassing protocol authorization or deciding new semantics.

Plan: raw-header-admission-failure-plan.md SHA256 db1599a1f1b38f437722b8abae8acda1847c1b1e4fd8bb2c983b118d32b2acb6.
Reviewed worktree: /Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence, base453b6dab386746bb03aac2c0a293572b8a71c969; existing five-file diff ea04e6ba34c9e1d9b724348f32407c6fd6bd78b795cad0481e7f4093b0f21762 remained unchanged at review. This is a plan/source review, not the required successor code review or execution release.

## Actual defect confirmed

Read the retained validation static-cases.exit.json and results.json. The attempt exited1 in1.836 seconds without timeout/forced cleanup, with1pass/4fail/94skipped. The app suite registered no results because bundle construction failed with six manifest issues; its two selected callbacks did not run. The cumulative failures include the expected-error assertion being masked by the same parser error. The retained supervisor records source unchanged, no remaining sampled PIDs and no owned runtime directories; those records are not independent new process observations.

At executable-manifest.ts:975–980, the broad auth classifier rejects legitimate bdp-authorization-view raw source lines. The normalized-header path at1472–1476 has the same defect. http-executor.ts:670–672 independently rejects the same authored field before credentials/dial, so parser-only repair is insufficient. The field is settled protocol context under docs/specs/bdp.md:3581–3588 and5263–5273. The repair does not relax field syntax, command authorization or minimum semantics.

## Security and boundary checks

1. The exact normalized comparison must exempt only bdp-authorization-view. Keeping the existing sensitive-name and token/secret/password/credential/auth patterns for every other name preserves credentials and lookalike refusals. Keep reserved/framing checks outside that exemption. Normalized manifest input still requires lowercase names; raw classification may lowercase for comparison while retaining original names/order/repeats/empty values.
2. Credential handle registration at http-executor.ts:469–477 currently accepts the field because of the same broad classifier. Using the corrected predicate negatively makes the exact field unavailable as a credential handle. This complementary restriction is necessary. Resolver membership/cardinality checks at698–713 then refuse unauthorized extra context keys without an added injection path.
3. No changes to report projection are needed or proposed. runner.ts:1791–1803 emits only allowlisted redacted exact-header names, while3306–3316 redacts compatibility maps. Current allowlists omit the context field. Admitting a source field does not entitle its value or name to enter reports.
4. The names-first app assertion, whole exact-request runner checks, own-data validation, field-value grammar, framing/route/body bounds and source/session binding remain. Two short module-private predicates are proportionate; an exported shared abstraction or runner API is unnecessary.
5. The normalized path repair is a consistent correction of the same existing capability, not permission to replace raw fixtures with Headers or credentials. Keeping the original five draft files byte-identical preserves the evidence target.

## Proposed regression feasibility and count

The seven new expanded callbacks are2 raw parser forms +1 normalized parser loop +1 genuine loopback serializer exchange +1 credential-registration loop +1 six-lookalike parser/serializer loop +1 mocked-transport report-privacy case. Both requests/actions and v1/v2 ordinary-header forms exist in the parser; raw remains v2-only. The existing runner setup at4687ff exposes exactRequest and report serialization for a no-socket privacy regression, and its credential-decoration test at5575ff demonstrates insertion at the authored-prefix/credential boundary.

The fifteen retained callbacks are1 normalized credential/noncanonical +1 authority/secret/control +2 raw closed-form +9 protected serializer names +1 registered credential decoration +1 credential-shaped response-report test. These titles and the nine-name protected-field expansion exist. Together:22 focused callbacks. The original seven selected static callbacks plus the historical preservation/fence callback at transactional-cumulative.test.ts:836 make8 further callbacks. Total30, not30 newly registered tests. Expected actual network workload is one loopback request/listener; zero product children, installed children or stores. Vitest/build worker processes remain separate.

Implementation details to preserve during measured draft/code review:
- The real wire test should collect a complete request rather than assume the first TCP data chunk contains the full head/body. Use the existing tracked socket/listener cleanup and bounded test/command deadlines.
- The mocked privacy case must assert scenario/assertion success as well as sentinel absence; otherwise rejected admission could make a redaction test vacuously pass. The proposed independent authored-header assertion and success requirement address this.
- Credential and lookalike negatives must fail as configuration/not-started before resolver calls. The parser and executor arms should both be exercised; positive wire coverage must still include genuine Authorization via its registered handle.
- Use exact describe-prefixed titles/parameter expansions from the final source in the command manifest, rather than a broad file run. Existing selected titles are source-verifiable, but final new titles cannot be certified before implementation.
- Preserve the original failed attempt and use new output directories. Rebuild the emitted consumers after conformance production changes before repacking/installed execution.

The existing planned installed208 expected/225maximum HTTP and26children are unchanged. The separate regression wire makes209 expected HTTP if both stages are released; it must not be counted as installed evidence. No extra installed diagnostic cohort is needed.

## Scope and conclusion

The plan expands ownership by exactly five source paths: executable-manifest.ts, http-executor.ts and their two test files plus runner.test.ts, for ten total changed paths with the original five frozen. Before-code hashes of these five additional paths:
- executable-manifest.ts: c450ae3c39fd8a8c82d83b58eb2a527fbff942dcc965ac018a01289d9f0d5c36
- http-executor.ts: faf4b4622db4a9105ac49f38f0c8eef116866d1306f8dbae83176e3e6e06c285
- executable-manifest.test.ts: ef54d9332e7965ce8c0d815680a6bddba5ab00bb57117b99c78c2ae70cc43512
- http-executor.test.ts: 83fd213b9f1897dab3b26022cce8b7cc2df2134176c1ed8ee5e05bd00c19cc9c
- runner.test.ts: fec61f2617ed060527d0e95e3f481a9b392f1485e700c01235b8bb1930d771ad

No source change, test, build, import, service, package, install or actual Claude invocation was performed for this review. The proposed repair is suitable for root-authorized implementation and subsequent frozen-code review. Root retains execution/publication gates; exact-head CI and actual Claude final-code review remain separate requirements. No public Transactional qualification or claim eligibility is inferred.
