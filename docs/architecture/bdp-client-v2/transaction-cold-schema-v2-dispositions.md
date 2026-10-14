# Local schema serving v2 — review dispositions

Status: revised after completed Claude review and root dispositions; **independent closure and root acceptance pending**. No source authoring or execution is authorized. Root's `transaction-cold-schema-root-plan-dispositions.md` controls these decisions. The complete external report is `transaction-cold-schema-claude-plan-review.txt` (exec94022 exit0, original basis ed359/tree7544). This revision uses frozen90cc834c3672b02840c68dcd7abb35794f038887/tree97db367180fb49a20e831fba859e0ac09482966b for new source citations, with no assumption of merge or authoring-parent assignment.

The original audit remains unchanged: `transaction-cold-schema-serving-audit.md`, SHA256 e00fa92025a26e131420476fb948b4dcc3f5146776775724a8597c8da8bb2cd4. Original native review60aa08ec and external review history are not rewritten. Before revision, both initial drafts were preserved byte-for-byte in `transaction-cold-schema-plan-v2-initial-draft/`: plan SHA256251287fda46c1a902ab2204e8ba2d94853a6252f4a8f385ba7c8406dfc42a7c1; dispositions SHA256a2597110c70aaab4d0e97d6a863a72a4daefabd3a0fb12059c386761766ed12b.

## Native dispatch correction — accepted

Runtime credential parsing is not authentication rejection, and the runtime does not already execute the HTTP handler's minimum gate. After existing raw-target/Host/method/body/closing/Origin checks, intercept only authenticated GET/HEAD exact manifest hits with none of the three minimum fields present. Empty/malformed/partial values count as present. All excluded requests fall through the unchanged handler without a schema observation. Never call the handler and override its404 or other refusal. Retain source cases for authentication, conditional and Accept interactions, and Origin/method precedence. The initial seven-file/no-export recommendation is superseded specifically by root's acceptance of external finding3 below.

## External 1 — retrieval URI versus canonical $id (High)

**Reachability concern accepted; proposed new startup refusal rejected by root.** Graph canonical IDs and retrieval identities differ legitimately under current offline compilation. Serving only exact eligible captured retrieval identities leaves canonical/embedded-ID aliases and differing-$id reference bases as explicit cold-traversal gaps. V2 records those limitations and requires source compatibility cases: accepted differing-$id configuration compiles and starts, offline mutation validation remains usable, eligible retrieval URLs are served, and otherwise unrouted canonical aliases remain explicitly unserved rather than fabricated. No adapter-construction/startup/configuration refusal or URI equality rule is added.

The installed positive has a deliberately aligned fixture: independently check actual root $id equals its captured retrieval URI, resolve actual root $ref against actual $id, and require the resulting target to equal the independent transitive capture identity in the exact served manifest. This tests a working supported traversal and does not claim all accepted closures are traversable. Broader cold reachability remains work for the complete profile milestone.

## External 2 — namespace completeness (High)

**Residual coverage concern accepted; proposed new startup refusal rejected by root.** V2 explicitly covers the schemas/ retrieval-identity subset, not every Scope-local installed closure. Accepted outside-namespace and query-bearing schemas remain available to existing offline validation, with unchanged configuration/startup and HTTP fallthrough. Required source cases preserve acceptance, observe exclusion and unchanged handler outcomes, and prove no shadowing of real Resource routes. No wider URI grant, new proxy, configuration restriction or normative exemption is introduced. These remain outstanding cold-client deployment cases.

## External 3 — shared minimum gate (Medium)

**Accepted through exact helper reuse.** Scope expands from seven to eight files only by adding `packages/server/src/transaction-http.ts` to export its existing hasMinimumControls at90cc:144–148. The adapter imports and calls that helper. Its implementation, field list and existing handler behavior remain unchanged. There is no new list, package/index export, consistency parser refactor or general abstraction. Service tests stay in scope. Presence semantics, including empty values, and the full behavioral exclusion matrix remain required.

## External 4 — context authority (Medium)

**Partially accepted for precision.** Actual spec5768 states “Transactional Scope-bounded responses carry” before the three fields; the external report's characterization as merely syntax is not adopted. The general rule does not itself classify every static/external schema route. V2 explicitly chooses to observe these authenticated exact local schema responses under this Scope owner and emit their actually observed current tuple as a conservative local serving contract consistent with the rule. It neither asserts a universal schema classification nor declares schemas exempt.

Retain separate narrow schemaObserved under existing run/withRead, with owner readContext as the sole authority; preserve sequenceRefusalObserved unchanged. Retain lifetime, entry-fault and current-position tests. No unobserved-byte-only split, dropped context oracle, minimum support claim or operator ruling is introduced. The service/progress overlap remains gated on root releasing the PR118 repair owner.

## External 5 — raw writer (Low)

**Accepted.** Extend the existing private runtime deliver with an optional exact representation bytes/string argument instead of adding a parallel sender. Use supplied raw representation only for selected200, without JSON serialization. Keep its destroyed guard, development marker and final defensive HEAD strip in one place. Adapter results keep selected HTTP metadata separate from raw bytes; public HttpResponse is unchanged. HEAD200 reports exact captured byte length without transmission. Selected304/406/412 receive no raw representation/body and preserve selector metadata, including absent304 Content-Length. Required transport tests cover the marker for200/304/406/412, destroyed/disconnected delivery and HEAD suppression.

## External 6 — citation drift (Low)

**Accepted with historical identities preserved.** New v2 citations are checked from frozen90cc Git objects. Relevant anchors: transaction-http.ts:144–148 minimum helper,387 nonmutation authentication,546–557 minimum gate,560–565 withheld discovery,566 Read fallthrough; transaction-runtime.ts:86–93 sender,115–164 pre-dispatch gates,165–170 handler dispatch,311–330 owner/service/handler construction; development-contracts.ts:56–90 capture and byte reconstruction; transaction-service.ts:130–145 lifetime registration,527–536 sequence observation pattern,578–595 draining close. Original ed359 audit/review citations are historical and unchanged.

## Bounds, ownership and closure

Eight files: new private schema adapter and test; runtime; service and service test; installed schema test; progress document; existing HTTP module solely for the helper export. No other source scope is authorized. PR118's active two-file repair retains sole ownership of overlaps; actual parent assignment is still root-owned.

Retain14 positive requests/one child and predecessor negative stopping at3 requests/one child:17 expected requests/two children, maximum28/two. Source compatibility and gating cases must be measured separately before execution; they add no installed requests. Retain all exchange/body/header/target/observation/time bounds, finite cleanup, exact bytes, independent reference resolution, current-context controls, source/package/route bindings, named negative failure and independent postprocessing. No hidden fetches, broad installed matrix, fabricated discovery, profile advertisement or cumulative qualification claim.

All six external findings now have an explicit root-directed disposition in the coherent plan. This is plan revision, not implementation or evidence of passing tests. No product source, tests, build, service, install, network or outward record was changed; only the two requested supporting Markdown files and their initial-draft snapshots were written.
