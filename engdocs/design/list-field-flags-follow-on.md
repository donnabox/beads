# List field flags and exact Count

Status, October 29: source authored on `codex/melinda-list-field-flags-20261029`, based on List361 at `6d3903f893b376c008d8960ebce1f168c8ee4784`. Formatting, normal generation and bounded source qualification have run; the first forty-root execution had one failure, followed by a qualified four-root correction. Their exact subjects and failure history remain separate. The corrected native embedded workflow, including its SQL-text differential assertions, passed and was independently accepted by Janet. A standalone native playground binary exists. This fixed-field follow-on has not yet been committed or published as a product PR. Installed global Query/List remains source361.

Donna clarified that fixed field flags are the intended everyday List surface. Raw `--ast` is an optional advanced input and implementation bridge. The actual spellings come from canonical CLI PR82 at `dcaad407b1ef59479fb3fd6a771c0caee81f4b59`:

```sh
bd list --type memory --property-field status=active --has-metadata-key reviewed_by
bd link list --metadata-field team=compiler --has-property-key note
bd list --type memory --count-only
bd link list --from ISSUE --property-field title=Evidence --count-only
```

These flags are present in the standalone native playground binary at `/Users/dbox/Documents/Codex/2026-10-29/query-playground/bin/bd` (SHA256 `a3e2b7c791c7debcf680eaac854286db6e61f673bae47e25b41ce5cb982a23fe`). Its normal build, signature and help checks passed. The later native workflow executed the Bazel-injected binary at SHA256 `cb2a2a115b56f9ee72d7641538e8fc2778496975f83828928d9a99b319311999`; these are distinct builds. The global installed binary has not been replaced with this follow-on.

## One candidate selection

`--type T` includes T and its refinements; `--exact-type T` selects T only. The chosen Type conditions AND with the existing endpoint selectors, an optional candidate-local AST, and every fixed field condition. Bead List retains the Issue-family default and `--type any`; Link List retains its existing population defaults. Type affiliation uses captured Type definitions; it does not authorize arbitrary traversal.

The field flags are `--property-field KEY=VALUE`, `--has-property-key KEY`, `--metadata-field KEY=VALUE`, and `--has-metadata-key KEY`. Property presence and equality flags repeat; existing metadata equality remains repeatable and metadata presence keeps its existing scalar spelling. Every supplied condition ANDs. Repeated conflicting equalities remain separate conditions and cannot overwrite one another. Keys such as `gc.routed_to` address one literal top-level bag member, never a nested path. Values retain all text following the first equals sign, including empty text or additional equals signs. Keys follow the existing metadata-key validator. Properties and metadata are separate namespaces.

A missing member fails equality and presence. An explicit JSON null is present. Fixed flags inspect only the candidate's properties or metadata; they cannot dereference a Link endpoint. Optional AST input is recursively checked against the List profile before acquisition, including short-circuited branches. Walks, endpoint dereferencing, reference-set acquisition, nested/global selections and any/all subqueries refuse. A Link's stored endpoint reference remains its own data. The four existing representation-root refusals remain.

Existing dependency-mode metadata flags retain their existing parser and provider. Property flags and raw List AST still require graph mode. The additive flat dependency-mode `bd list --count-only --output-format json` adapter reuses the existing List selection and reader, removes delivery paging, and emits a count only after a complete read and successful cleanup. Max-row safeguards remain active. Explicit canonical Type selectors, conflicting metadata equalities, continuation, watch and hierarchical/dependency rendering are refused where this adapter cannot honor them; these are remaining audited gaps, not approved exceptions. In graph mode field flags select the additive canonical List path; no new text expression grammar is introduced.

Remaining dependency-mode property/AST filtering and canonical Type-authority gaps are explicit cross-mode audit items, not approved permanent exceptions or claims of full canonical/compat support. The flat count adapter described below is additive and does not close those gaps. Existing dependency-mode selection behavior must stay unchanged. Cosmetic output normalization may wait; selection, exact counts, identity/revision, completeness and errors remain substantive.

## Settled compatibility rule

Donna explicitly selected existing DIRECT LIST flags as the authority, rather than the old Query OR helper. These paths disagree for nulls and container formatting; their union would silently broaden matching. The source decision and alternatives are recorded in the local compatibility source review.

The explicit `LegacyListFields` adapter follows direct List's `JSON_UNQUOTE(JSON_EXTRACT(...))` text comparison. It keeps the shared AST's equality typed. JSON string `"1"` and number `1` can both match flag text `1`; captured numeric `1.0` normalizes to the admitted graph number `1`. JSON null matches text `null`, not empty text. Empty string still matches empty text. Containers use the pinned SQL formatter's spaced separators and length/byte key ordering: `[1, 2]` and `{"b": 2, "aa": 1}`. SQL UNQUOTE decodes string escapes even inside containers. Numbers use the pinned formatter's integer special case or fixed decimal output, so `1e21` becomes `1000000000000000000000` in comparison text. No caller value is automatically parsed as JSON.

The small pure formatter is source-bound to existing go-mysql-server `e5eafe0da809`, `sql/types/json_encode.go` and `sql/expression/function/json/json_unquote.go`; it imports no engine. This is a named CLI compatibility boundary, not an added portable AST operator. Data remains within Core's existing JSON number admission domain; no integer/float distinction is introduced. Per-request field counts/bytes and cumulative comparison work are bounded, repeated field conversions are cached per candidate, cancellation is observed, and a failure returns no partial selection/count. The new error mappings preserve existing AST classifications: operational budget/cancellation refusals are capability unavailable, corrupt data is invalid store, and malformed keys/syntax remain invalid selectors.

The executed native differential compares old SQL-backed Issue-search metadata predicates with new List counts against the same disposable Issue and embedded engine. Its seven cases cover numbers, null versus empty text, arrays and object formatting. These checks passed inside the selected embedded workflow. They do not establish universal collation, server-mode or independent-backend parity.

## Count is the same selection

Donna confirmed `bd list --count-only` and `bd link list --count-only`. Exact Count applies to the complete intersection above, before page boundaries or delivery limits. Valid Resource field masks cannot change its value; only envelope delivery carries to the aggregate. Aggregate output uses existing `clioutput` rendering with `result_kind: aggregate` and a protected `count` field. Existing legacy `bd count`/grouped-count behavior is not changed in this slice.

For graph-mode List, the evaluator already supports a `count` result form, but the retained public List AST remains a Resource-selection document. The CLI's count-only result operation counts the complete List selection including Type and endpoint selectors; it does not count a returned page or discard a refusal. An acquisition bound/error is never an exact partial total. Current execution still captures and evaluates the checked inventory: this is not a claim of SQL COUNT pushdown or improved storage latency.

## Review and qualification

Cleo confirmed no overlapping active List flag/parser implementation. Her existing shared root-cause foundations and tests must be preserved; she remains paused. Sybil reviews shared request composition, Trish owns native Type/capture authority and receiving, and Janet owns all formatter, generation, Go/Bazel, build, engine, installed and integration reservations. No current runtime slot or receiving baseline is inferred from earlier queue order or candidate publication.

Source review found one operational-error mapping issue; the source applies the same translation to field admission/evaluation and AST evaluation. The first forty-root execution then found a real empty-argument defect: pflag's `GetStringArray` CSV round trip hid a sole explicit empty argument. The correction reads the registered `SliceValue.GetSlice()` and preserves literal empty/comma/quote/equals arguments. Four affected CLI roots passed afterward, together with normal generation fixed point and formatting/nogo checks. The original one-failure execution is preserved rather than relabelled as a fresh forty-root pass.

The bounded executions cover recursive List AST admission, literal keys and namespace separation, repeated AND conditions, presence/null and SQL-text formatting, cancellation/budgets, graph-only opt-in versus dependency metadata preservation, count before paging/masks, and input/error admission. The later native execution selected exactly `TestListASTInstalledWorkflow` and its `embedded` subtest, with two named RUN/PASS/XML cases across sixteen uncached shards and no named skips/failures. Its assertions cover Type/endpoint/AST/field intersections, count and seven direct-SQL differential cases. The execution used a disposable test-owned graph; server-mode behavior was not selected. The 17 child processes exited naturally and were reaped, source postchecks matched, and Janet independently accepted the evidence. This does not relabel the earlier forty-root failure as a fresh forty-root pass. The superseded native proposal stayed UNRUN; no passed test was replayed to repair offline analysis.

Published user CLI documentation remains pinned to release `v1.2.2` through `docs/cli-docs.pin`. This unreleased candidate does not move that pin or hand-edit generated CLI reference files. Its source help and this contributor note describe the proposed follow-on.

Separate pure Type-affiliation hardening is published in fork draft PR365 at `d32230778fa74b69ac7d357893d06cb64e71a342`, based on this same361 commit. Its four files do not overlap this follow-on's Go edits. It is not silently adopted here: receiving composition and native acquisition remain separate owner-reviewed steps.

## Dependency-mode count follow-on

The flat dependency Count adapter uses `IssueReader.List` through the shared complete-List helper. It preserves visibility, ready, parent, label, metadata and Issue/wisp selection policy; the legacy Count reader has different population/readiness behavior and is not substituted. Delivery limits/cursors are removed, Resource masks are validated before their aggregate effect is discarded, and malformed, duplicate-ID, truncated, capped, canceled or failed reads produce an error without a partial count. Existing `bd count` remains unchanged.

Qualification uses separate exact-source results: WorkAPI four roots/nine named cases from the initial cohort, eighteen passing command roots from the next cohort, then the corrected error-helper root containing all six plain-loop rows plus fmt/CI-nogo-race and two normal generation fixed points. The two historical failed cohorts remain recorded as failures. The final cause-specific run passed one named RUN/PASS/XML root and fmt, with 92 settled natural-zero children, 184 raw streams and unchanged 6,163 source/index facts. The six rows are not six separately logged subtests. This does not establish a new whole-cohort pass, native/store/server CLI behavior, installation, complete canonical coverage or main readiness.

The public documentation pin remains unchanged. This contributor note records development source behavior and evidence; it does not advertise support in the pinned release.
