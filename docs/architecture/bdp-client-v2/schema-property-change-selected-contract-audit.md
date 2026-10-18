# Schema PropertyChange: R selected-contract audit

## Findings and decision recommendation

**No remaining missing partition identified in R's five selected requirements at c57f5a7a.** On the accepted finite representative-partition standard, its row-local evidence now supports a later separate whole-row promotion decision. The schema-invalid Property Change gap is closed for both Beads and Links by actual refusal, complete unchanged preimage, admitted effectful correction and independently expected current GET. This conclusion does not rely merely on adding tests or observing a green suite.

R remains `wholeContract:false`. No promotion, source edit, test run, network, service launch or outward action occurred in this audit. Only this external report was written. Root must separately decide and authorize any coverage change. Current inventory stays33 reviewed-whole/96 not-whole out of129 required,54 authored and75 omitted; public qualification remains false.

## Exact basis and interpretation

Frozen source `c57f5a7a804b1b7ababf2d9195d7fdccbd37dee8`, tree `b7ffb7fb34ea5014ba824be945026a353eb94a70`, direct parent `627b560c0daf1a76b292615718c86670255ac9af`, clean at inspection. Checkout `/Users/dbox/repos/wt/vickie/vickie-schema-property-change`.

Read all five canonical requirement objects in `packages/conformance/catalog/numeric-model-v1.json:38–68`, relevant spec context, all96 authored R actions/assertions and ordered coverage bindings, the installed Type cohort, control implementation, author report/summary, and actual standalone and genuine continuation reports. The result reports store five source-anchor references rather than selectedText objects; those references exactly match the canonical objects in order, including the two separate Revisions selections. The canonical objects themselves remain unchanged; the selected contract is not narrowed to the row title.

`docs/specs/bdp.md:2638–2669` supplies the validation-list preamble: each singleton or sequence member is validated independently; when reached, its resulting state is checked before the next operation. The Type/schema list item must therefore cover update validation, not only creation. The earlier zero-action proposal's create-only schema witness did not close that distinction; its favorable native recommendation is not treated as authority here. The present audit specifically verifies the now-executed schema update path. This does not import every neighboring list item (authorization, aggregate constraints, endpoint policy, revisions, limits) or their separate conformance rows. No A/B/N or baseline-row executions substitute for R witnesses.

## Five canonical clauses and row-local mapping

### Clause 1

Source: `docs/specs/bdp.md#revisions`. Exact selectedText:

> a number literal is **admissible** if and only if converting its exact decimal value to the nearest IEEE 754 binary64 value and serializing that value back in its shortest round-trip form — the ECMAScript `Number::toString` form that [RFC 8785 Section 3.2.2.3](https://www.rfc-editor.org/rfc/rfc8785.html#section-3.2.2.3) adopts — yields a literal denoting the same exact decimal value.

Both kinds execute `nr-{bead,link}-{integer,precision,overflow}-{create,patch}-refused` and their corrected requests. Exact raw literal bodies preserve9007199254740993,0.12345678901234567891 and1e309 before binary64 parsing; independently authored corrections use9007199254740992/9007199254740994,0.125/0.25 and1e300/1e299. Schema-backed success/current GET assertions distinguish actual admissible stored values from blanket refusal or rounding bad input. Integer metadata whole/change refusals additionally exercise1e-400 and long fractional precision with admitted companion properties. **Supported:** representative admitted/refused partitions of the iff definition; not exhaustive enumeration of decimal literals.

### Clause 2

Source: `docs/specs/bdp.md#revisions`. Exact selectedText:

> `1e300` and its expanded form are admissible; `-0.0` denotes zero, is admissible, and is the same value as `0`; `9007199254740993` and a decimal of twenty significant digits are not admissible.

The corrected raw creation bodies explicitly retain1e300 and its301-digit expanded spelling,1.0 and-0.0. Result and current GET independently require1e300 for both large/expanded and0 for zero. Refusals explicitly include9007199254740993 and the long fractional decimal. Both Resource kinds execute these requests. **Supported:** all named examples represented; no need to borrow B's distinct no-op revision-retention law to show that admitted-0.0 yields zero.

### Clause 3

Source: `docs/specs/bdp.md#numeric-admission`. Exact selectedText:

> An authority MUST refuse at admission a `properties` document — supplied whole or through a Property Change — that contains an inadmissible literal at any depth: nothing changes, and the operation fails as the write profiles' `validation-failed` with a diagnostic naming the offending member.

For both kinds, whole properties create refusals are followed by absence404; Property Changes are followed by full Resource equality against an independently validated prior GET capture, including revision/provenance/metadata and Link endpoints. The precision cases nest the bad value in an object/array under escaped key `a/b~c`; whole-document diagnostic pins `/value/items/0/a~1b~0c`, patch diagnostic names `/propertiesChange/0/value/items/0/a~1b~0c` in the message. Each bad patch also carries an effectful companion metadata change, so full preimage equality checks no partial commit. Exact422/validation-failed/diagnostic Problem assertions and effectful admitted corrections/current GETs discriminate refusal from arbitrary failure. Link endpoint/final source-target reads remain unchanged. **Supported:** whole/change, shallow/nested, member naming and nothing-changes conjuncts. Finite depth is representative, not a claim of enumerating arbitrary depth.

### Clause 4

Source: `docs/specs/bdp.md#validation-and-results`. Exact selectedText:

> Type and properties-schema constraints, and the admissibility of every number literal under [Revisions](#revisions);

Numeric cases cover admissibility in both properties and metadata. R's own `nr-{bead,link}-schema-create-refused` declares the installed child Type with admissible `n:-1`, enabled:true/detail:ok; exact Problem identifies inherited base-Type minimum, and absence GET proves no create. Corrected n:0 creates the same declared child Type. New `-schema-patch-refused` applies admissible-1 through propertiesChange to an existing valid record, with metadata.phase→updated. Exact minimum diagnostic, `-schema-patch-unchanged` full captured-preimage equality, fresh-key n:1 correction and `-schema-updated-read` independently pinned properties/metadata prove authority-side resulting-state schema enforcement on update. Both kinds execute the whole chain. **Supported:** this now addresses the actual missing update partition without recasting the selected Type/schema clause as merely numeric. The installed child inherits base minimum0 and has its own enabled/detail requirements; the failed requests satisfy the child requirements, so their specified refusal cannot be explained by numeric inadmissibility or malformed carrier syntax.

### Clause 5

Source: `docs/design/requirements.md#protocol-requirements`. Exact selectedText:

> An authority MUST refuse at admission any number literal in `properties`, at any depth, whose exact decimal value does not round-trip through IEEE 754 binary64 unchanged, as `validation-failed` with a diagnostic naming the member

R's own whole/change shallow/nested numeric refusals, escaped-path complete Problems, absence or full unchanged reads, effectful admitted corrections and current GETs described for clause3 cover this duplicate PROTO016 obligation. **Supported:** no separate N execution is counted and no selected text is dropped.

## New schema-update chain: independent oracles

For each kind, the four new actions are `nr-<kind>-schema-patch-refused`, `-schema-patch-unchanged`, `-schema-patch-corrected`, `-schema-updated-read` (Bead indices82–85; Link90–93). Four existing actions change only URL capture or conversion of the validating initial GET to the client family;84 prior actions remain unchanged.

The admitted creation response yields the canonical URL. The following initial GET independently asserts fixed Resource identity, declared child Type, properties `{n:0,enabled:true,detail:"ok"}`, metadata `{phase:"created"}`, absent ownedLinks/provenance, and fixed source/target for Link. That actual GET is captured as `nr-<kind>-schema.before`. A schema refusal then must preserve the entire record, not just n or status. Because all those initial assertions passed in the actual positive reports, the capture is independently anchored rather than circular snapshot equality.

The refused patch has n:-1 and metadata.phase:updated. The correction differs in the properties numeric leaf(-1→1), with a fresh key; both document planes now change. Exact configured refusal diagnostics include base Type URI, schema URI `#/properties/n/minimum` and instanceLocation `/n`. Corrected response and current GET independently expect n:1, phase:updated, retained child Type/identity/endpoints and absence of ownedLinks/provenance; corrected responses also retain absence of source/sourceRevision. This rejects a stale successful read or silently dropped metadata patch. The corrected response's numeric/schema assertions are not generated from the returned body.

Four new controls in `apps/bdptest/src/read-update-matrix.test.ts:5123–5211` alter assertions' inputs only after actual HTTP: Bead schemaLocation→only complete-problem fails; Link refused metadata phase→only postimage fails; Bead corrected n→only properties fails; Link corrected metadata phase→only metadata fails. Each checks one application, exact sole failing assertion across the whole row and cleanup. The raw control retains observed422/complete-write; client controls independently inspect the original200/complete-write and original field. All four are passed tests in the final installed23-test report. They are sensitivity checks, not four extra protocol behaviors or injected authority faults.

R's earlier pointer/metadata-label/create-schema-location controls remain and pass. The combined installed invocation also executes N and shared controls, but this audit credits only R's own observations/controls. A selected row need not run every schema keyword or Type-graph permutation to have finite complete-clause evidence; that would be a schema-language test project, not a newly identified missing conjunct.

## Independently inspected actual reports

Both report bodies identify clean c57 and claimEligible:false. I recomputed these file hashes and reconciled every authored action ID and every ordered assertion ID with the actual row; no missing/extra action or assertion was found:

| Actual report | SHA256 | Bytes | Actual R |
| --- | --- | ---: | --- |
| `schema-property-change/numeric-admission.round-trip.json` | `905d62c98639fb09b3ec6fdd5b0143466b69e7ee03e7ca0302de850c0daceb0c` |264384|pass;43 raw+53 client;687/687 assertions pass; cleanup true |
| `schema-property-change/artifact-set-progress.json.reference-continuation.json` | `1c7a9a3d62a5d4ad2558e4f300e4d927a07b8e243ed0b8a8119816879a341f67` |1027195|same96-action/687-assertion all-pass row in genuine cumulative run; cleanup true |

R has36 raw mutation exchanges. Its coverage binding list exactly equals the96 authored actions'687 assertion IDs in order. The author summary's96/687 is therefore checked against actual records, not used as proof by itself.

Actual target binding: payload `1b12629fe54c39c8f7864dbb13bdba16bcbe4298b0a5bd420fa19daaa4fb7ec2`, installed contracts `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`, configuration identity `1720b63b49a7be3299d4b2865de4db9fcbe47560f217a32eea6397db7ea54ca5`; fixture `f89932d68e3bf5867d79eb3de680af4bc67b761b3c3bce5da127f6ca75e27765`, Scope `https://reference.invalid/acme/`, Read+Update, Historyfalse. Frozen source identity is separately caller-bound; payload identity is actual captured installed bytes. Continuation manifest `24d754618b51f255b5885d83c53ccf9a06593031e1812739a4f2b504c4a1df8d`, coverage `ee39ca50dc8f48dd062fa8f2e1364c1842aa3bb2d1f732152dd42646ffa154b1`.

Existing final evidence, read not rerun:23 installed tests pass/179 filtered,29.844s; genuine aggregate54 observed pass plus75 omissions,18.451600833s under90s. Static22 and final focused20 also pass. Preparation typecheck/static/audit-bookkeeping failures are explicitly preserved in the author report; no fresh execution is claimed here. Author cleanup reports no remaining installed children. Continuation899400 bytes/23502 nodes leaves only2495 bytes/44 nodes under unchanged continuation ceilings; no capacity or deadline change is proposed by this audit.

## Explicit residuals and next action

The raw effectful correction's response revision is not linked by opaque-token equality to the later GET revision. That is a real evidence limit, but not a missing conjunct of these five selected requirements: the numeric definition does not select every paragraph under Revisions, and the Type/schema list item does not select the later complete-result coherence paragraph. Conversely the selected “nothing changes” requirement does cover failed-update revision stability; complete preimage equality actually observes that requirement, including revision. These two scopes must not be conflated.

The findings do not certify every schema language construct, installation/hosting/retention law, Type lifecycle, every mutation operation, owned-Link schema fanout, concurrency, History or public profile conformance. No neighboring requirements were narrowed or silently credited; no baseline/A/B/N evidence filled R gaps. Representative whole/create and PropertyChange/update validation is now present for both Resource kinds, including a genuine inherited constraint. I identify no further distinct selected partition requiring implementation before a separate R whole decision.

Root may now consider the separately reviewed promotion with accurate limited reason, canonical allowlist and count updates. This report makes none of them. A single authorized promotion would produce34 whole/95 not-whole, not an automatic public claim. Preserve this c57 evidence as c57 evidence when constructing any successor.

## Complete authored R action accounting

Setup0–9; numeric Bead10–37; numeric Link38–77; schema Bead78–85; schema Link86–93; unchanged endpoint finals94–95. Each action below was reconciled to its recorded ID/assertion list in both actual reports.

| Index | Action | Transport | Assertions |
| --- | --- | --- | ---: |
|0|`scope`|http|2|
|1|`discovery`|http|10|
|2|`directory`|http|2|
|3|`directory-captures`|client|1|
|4|`task`|http|4|
|5|`relates`|http|6|
|6|`nr-create-source`|client|12|
|7|`nr-source-initial`|client|10|
|8|`nr-create-target`|client|12|
|9|`nr-target-initial`|client|10|
|10|`nr-bead-integer-create-refused`|http|5|
|11|`nr-bead-integer-absent`|client|4|
|12|`nr-bead-integer-meta-create-refused`|http|5|
|13|`nr-bead-integer-meta-absent`|client|4|
|14|`nr-bead-integer-create-corrected`|http|13|
|15|`nr-bead-integer-created-read`|client|9|
|16|`nr-bead-integer-patch-refused`|http|5|
|17|`nr-bead-integer-patch-refused-read`|client|4|
|18|`nr-bead-integer-meta-patch-refused`|http|5|
|19|`nr-bead-integer-meta-unchanged`|client|4|
|20|`nr-bead-integer-patch-corrected`|http|13|
|21|`nr-bead-integer-updated-read`|client|9|
|22|`nr-bead-precision-create-refused`|http|5|
|23|`nr-bead-precision-absent`|client|4|
|24|`nr-bead-precision-create-corrected`|http|13|
|25|`nr-bead-precision-created-read`|client|9|
|26|`nr-bead-precision-patch-refused`|http|5|
|27|`nr-bead-precision-patch-refused-read`|client|4|
|28|`nr-bead-precision-patch-corrected`|http|13|
|29|`nr-bead-precision-updated-read`|client|9|
|30|`nr-bead-overflow-create-refused`|http|5|
|31|`nr-bead-overflow-absent`|client|4|
|32|`nr-bead-overflow-create-corrected`|http|13|
|33|`nr-bead-overflow-created-read`|client|9|
|34|`nr-bead-overflow-patch-refused`|http|5|
|35|`nr-bead-overflow-patch-refused-read`|client|4|
|36|`nr-bead-overflow-patch-corrected`|http|13|
|37|`nr-bead-overflow-updated-read`|client|9|
|38|`nr-link-integer-create-refused`|http|5|
|39|`nr-link-integer-absent`|client|4|
|40|`nr-link-integer-create-endpoints-source`|client|4|
|41|`nr-link-integer-create-endpoints-target`|client|4|
|42|`nr-link-integer-meta-create-refused`|http|5|
|43|`nr-link-integer-meta-absent`|client|4|
|44|`nr-link-integer-create-corrected`|http|15|
|45|`nr-link-integer-created-read`|client|11|
|46|`nr-link-integer-patch-refused`|http|5|
|47|`nr-link-integer-patch-refused-read`|client|4|
|48|`nr-link-integer-patch-endpoints-source`|client|4|
|49|`nr-link-integer-patch-endpoints-target`|client|4|
|50|`nr-link-integer-meta-patch-refused`|http|5|
|51|`nr-link-integer-meta-unchanged`|client|4|
|52|`nr-link-integer-patch-corrected`|http|15|
|53|`nr-link-integer-updated-read`|client|11|
|54|`nr-link-precision-create-refused`|http|5|
|55|`nr-link-precision-absent`|client|4|
|56|`nr-link-precision-create-endpoints-source`|client|4|
|57|`nr-link-precision-create-endpoints-target`|client|4|
|58|`nr-link-precision-create-corrected`|http|15|
|59|`nr-link-precision-created-read`|client|11|
|60|`nr-link-precision-patch-refused`|http|5|
|61|`nr-link-precision-patch-refused-read`|client|4|
|62|`nr-link-precision-patch-endpoints-source`|client|4|
|63|`nr-link-precision-patch-endpoints-target`|client|4|
|64|`nr-link-precision-patch-corrected`|http|15|
|65|`nr-link-precision-updated-read`|client|11|
|66|`nr-link-overflow-create-refused`|http|5|
|67|`nr-link-overflow-absent`|client|4|
|68|`nr-link-overflow-create-endpoints-source`|client|4|
|69|`nr-link-overflow-create-endpoints-target`|client|4|
|70|`nr-link-overflow-create-corrected`|http|15|
|71|`nr-link-overflow-created-read`|client|11|
|72|`nr-link-overflow-patch-refused`|http|5|
|73|`nr-link-overflow-patch-refused-read`|client|4|
|74|`nr-link-overflow-patch-endpoints-source`|client|4|
|75|`nr-link-overflow-patch-endpoints-target`|client|4|
|76|`nr-link-overflow-patch-corrected`|http|15|
|77|`nr-link-overflow-updated-read`|client|11|
|78|`nr-bead-schema-create-refused`|http|5|
|79|`nr-bead-schema-absent`|http|3|
|80|`nr-bead-schema-create-corrected`|http|13|
|81|`nr-bead-schema-read`|client|9|
|82|`nr-bead-schema-patch-refused`|http|5|
|83|`nr-bead-schema-patch-unchanged`|client|4|
|84|`nr-bead-schema-patch-corrected`|http|13|
|85|`nr-bead-schema-updated-read`|client|9|
|86|`nr-link-schema-create-refused`|http|5|
|87|`nr-link-schema-absent`|http|3|
|88|`nr-link-schema-create-corrected`|http|15|
|89|`nr-link-schema-read`|client|11|
|90|`nr-link-schema-patch-refused`|http|5|
|91|`nr-link-schema-patch-unchanged`|client|4|
|92|`nr-link-schema-patch-corrected`|http|15|
|93|`nr-link-schema-updated-read`|client|11|
|94|`nr-final-source`|client|4|
|95|`nr-final-target`|client|4|
