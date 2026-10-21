# Native plan review — command minimum fields

**No blocking finding. One Low workload-accounting correction is required in the measured draft ledger.** The proposed six-file slice follows the recorded operator ruling, uses the existing shared command boundary, and preserves read minimum enforcement. This is a plan review, not authoring or execution authorization.

## Finding N1 — Low: count the RU port-probe listener separately

**Location:** implementation plan line69, aggregate “fourteen listener/owner instances”; `scripts/installed-reference-observation.mjs:10–16`; existing call pattern in `apps/bdptest/src/installed-reference.test.ts:118–120`.

The proposed reuse of the existing RU configuration/launcher pattern calls `availablePort()`. That helper creates a real loopback TCP server, listens on port0, obtains its address and awaits close before the installed child starts. It is an additional listener lifetime even though it makes no HTTP request and opens no store owner. The plan correctly counts thirteen TX owner/listener opens across source tests and controls, then one RU child owner; it misses this auxiliary listener when conflating listeners with owner opens.

**Impact:** the source/control lifecycle inventory is understated by one listener. This does not affect the 89 expected/119 maximum HTTP count, one emitted child, eight store directories or six read-only SQLite handles.

**Concrete fix:** record fourteen owner opens plus one auxiliary port-probe listener, hence fifteen listener lifetimes if the existing `availablePort()` pattern is used. Bind its completed close in the draft helper-call inventory and cleanup evidence. Keep the installed TX cohort separately at one child/owner/listener and eighteen requests. Do not refactor the shared helper or add another probe to address this bookkeeping correction. If the actual draft avoids `availablePort()` using an already supported approach, record that exact source instead; do not assume its absence from the current plan.

## Reviewed authority and source

Plan: `transaction-command-minimum-implementation-plan.md`, independently hashed as `00ffec1ab3146fa62601a66c95faf1eafae9986d0bad3070b599dbd1599b8ab6`. Read the entire plan and `transaction-selector-command-operator-rulings.md`.

The direct command ruling explicitly chooses ignoring the three copied read-minimum fields on mutation POSTs while preserving authorization, idempotency, expected revision and later receipt minimum checks. The proposed normative paragraph at `docs/specs/bdp.md:3565–3578`, plus scoping the request-side minimum paragraph at5766–5788 to reads, expresses that choice without weakening response-context duties or global HTTP bounds. Empty/partial/opaque/comma-joined field values remain ordinary receivable HTTP values; malformed framing and CR/LF are explicitly excluded. The plan does not infer unresolved read-tuple, retained Snapshot/continuation or metadata Event semantics.

Actual source inspected is the frozen c02ccc3706f2e258d380805162a8c47dbb73f4e1 checkout at `/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft`, tree088290751951c353797b984f690b7d9fd96deb80. Parent reports verified main122800511ee6306e4105e1f73bf3658f35c90a14 has that same tree; no network verification or replacement source was used here.

## Production boundary and callers

`transaction-http.ts:256–303` identifies the exact canonical operation URL, checks method, target/body bounds and media, and then performs the sole command-specific `hasMinimumControls` refusal. Removing only that block is sufficient for this ruling. The twelve literal targets in `transaction-discovery.ts:8–21` all enter the same boundary.

The following sequence path at304–335 retains prohibited-header-key, full-carrier parsing, authorization, count/composition, negotiation and actual sequence execution. Its precedence becomes exactly the existing no-copied-fields precedence; preserving the removed block's accidental early authorization order would introduce a separate change. The singleton/batch path at336–398 forwards body, key, authorization and receiving options, not these three headers. No store/comparison/service change is necessary.

`transaction-runtime.ts:122–149` still validates raw target/Host/method and appends raw header values before authentication. No global stripping or weakening is needed. A joined-value test remains distinct from an actual duplicate-header wire test. `transaction-http.ts:523–547` retains the receipt parser and root/page minimum dispatch; the exported source-private `hasMinimumControls` helper at144 must remain for other read callers. Other read/schema/discovery paths remain outside the deletion.

Read+Update `development-runtime.ts:440–477` already passes body/key/Accept into mutate or sequence without reading these fields. Its existing maximum header size/timeouts remain intact. The planned emitted-factory regression is appropriate compatibility evidence, separately labelled from the fresh packed TX consumer.

## Test and control assessment

The source ledger arithmetic is otherwise sound: A27 + B16 + C13 + D18 + E6 =80 HTTP exchanges, five fixtures and eleven owner opens. D's six receipt-row queries each close/read-only-open/reopen exactly as the existing helper at `transaction-http.test.ts:138–153` specifies. Two controls add expected1+2 requests and maximum27+6, giving83 expected/113 maximum TX HTTP, seven fixtures and thirteen owner opens. RU adds six HTTP and one owner/store/child; N1 corrects only its auxiliary listener accounting.

A crosses each field with all three carrier families and tests success plus same-key replay with only copied controls altered. B covers all twelve operation targets using valid under-limit Selectors, actual receipts/effects and exact final Bead/Link/alias state; its seed plus12 commands plus3 reads is16. Captured context and a minimum-wait spy avoid assuming alias operations advance positions. C preserves real key, revision, negotiation and sequence-header admission behavior; D preserves actual no-reservation evidence under failed authorization. E is an independent read-side guard with deterministic finite-budget seams already supported by source tests; it does not rely on command behavior as its oracle.

The old-block control must first reach A's successful-command assertion with the actual500/development marker. The receipt-dispatch-only control must reach E's first future receipt GET with incorrect200 against the named503 oracle. Both must rebuild emitted server code because this source test imports `../dist/transaction-runtime.js`. The plan expressly requires rebuilds and rejects earlier/generic failures, unexpected passes and cleanup masking. No global helper mutation is permitted. Exact response capture and assertion ordering remain draft-preflight obligations, not results inferred from this plan.

The installed sequence is exactly18 requests and nine canonical URLs: four operations, three Resource points, one explicit-limit collection and one validated receipt root. Its final complete collection excludes a conflicting extra identity; failed-revision state is deep-compared to the prior independently validated Resource. Equal retries alter only ignored fields. Actual receipt GET/HEAD minimum503, foreign-view409 and compatible200 remain discriminators after successful commands. Source E additionally covers foreign epoch410. No Types/discovery, member-receipt, pagination or restart traffic is hidden in that list.

The existing collection helper supports the declared authorization-view resolver channel and enforces5s exchanges,16KiB request headers,64KiB request body,32KiB response body and the90s scenario deadline. The local18-count/route assertion must be added without widening its shared80 guard. Existing finishObservation preserves primary and cleanup errors and can report after directory-removal failure; subsequent measured review must verify the new case uses that lifecycle. The strict receipt-root check must precede granting its route. Full receipt/sequence schemas and literal Resource/effect oracles are required as planned, with only opaque values captured.

## Scope and remaining gates

The six-file footprint is sufficient; no additional production API, runtime switch, parser, catalog, schema or public claim is warranted. Shared HTTP/spec ownership with the Selector lane is explicitly serialized. Source/control and installed executions are bounded, selected and separately attributed. The existing helper reports safe fingerprints rather than complete wire bodies; the plan correctly preserves that limitation and requires source→archive→consumer/full-dependency identity and independently checked PID/directory absence.

Before execution, root still needs the actual six-file diff, exact selected expansion, body/route ledger, N1-corrected listener inventory, valid alias/Selector shapes, concrete derivative patches and named-failure cleanup evidence design. Tests do not yet exist under the proposed titles, so this review does not certify their eventual implementation, measured sizes or compilation. The supported design is ready for root disposition after council, with no additional behavioral requirement identified.

Reviewer performed only source/document reads and hashing, then wrote this report. No test, build, installation, service, network/remote action or source mutation was performed.
