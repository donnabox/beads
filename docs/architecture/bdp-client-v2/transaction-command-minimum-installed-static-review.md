# Preliminary static review — new installed command-minimum scenarios

**No concrete fixture/response-shape blocker found. One Low oracle tightening is recommended below.** This is targeted preliminary scrutiny of two still-changing draft files, not measured preflight, execution proof or the later frozen code council.

Inspected checkout: /Users/dbox/repos/wt/vickie/vickie-command-minimum, parent122800511ee6306e4105e1f73bf3658f35c90a14. Source may change before final binding. At completion of this read:
- apps/bdptest/src/transactional-collection-minimum-installed.test.ts SHA256 b02bbd7306887e8ed8dfb0ab126210bfcf423f4c15a7030fa8959b8570f443e0.
- apps/bdptest/src/installed-reference.test.ts SHA256 94570e6ddb713b60bcc2134295d474fe00d578513914e3e248a633ca69fa2408.
- Git diff for those two paths SHA256 fcedcbfa592bbc322c139dc8ee277333e538c6925233b48b8edf1aa3f9d2d409.

## N1 — Low: pin the known receipt key and positions

Location: new TX receipt helper at1218–1227 and its A/B/failed submission call sites; later compatible receipt GET.

The helper checks the real mutationReceipt schema, status/detail and body epoch/view against response headers, and subsequent retries compare the entire body. It does not assert idempotencyKey, requiredPosition or effectPosition against the independently known command and committed prefix. The schema checks their shape, not these particular values. Consequently an incorrectly retained but schema-valid key or position could be repeated consistently and pass this new scenario. Resource postimages and serving headers are checked correctly; this is an evidence precision gap, not a demonstrated production defect.

Smallest fix, with no extra HTTP: assert first A receipt key minimum-a and required/effect pos-1; B key minimum-b and required/effect pos-2; failed update key minimum-stale and required pos-3, retaining the schema's effectPosition prohibition (an explicit absence assertion is also clear). Keep full replay equality. On the later compatible GET, retain the original A body including required/effect pos-1 while response headers correctly report current pos-3. Do not equate a retained receipt's historical effect to its later serving position. No new fixture, capture, route or helper-wide behavior is needed.

## Fixture and response checks

TX uses the existing captured Task descriptor with a required title and owned Relationship group. The explicit singleton create/update bodies omit operation; batch and sequence members include it. The explicit limit32 collection reads fit the fixture's maximum32 and avoid its default2 pagination. The exact Resource expectations—canonical ID/type, title, metadata:{}, opaque nonempty revision and ownedLinks:{[linkType]:[]}—match existing receipt/point observations in this same file, including lines829–841 and1002–1013. The normalizer defaults omitted metadata to an empty object.

The sequence's plain results envelope with operationIndex0/outcome created/resource is consistent with transactionalSequenceResponse and transaction-sequence.ts projection. The expected idempotency-conflict, revision-mismatch receipt Problem, catch-up-timeout and foreign-view shapes match the existing Problem/schema paths. Failed receipts forbid effectPosition/results/next through the schema. The point helper expects an ETag while the finite receipt/collection/context helper requires no ETag; this is consistent with the existing current point versus finite response projections.

The new scenario has exactly18 authored exchange calls and nine URLs after the single observed receipt root is granted. Calls are capped before delegating and the final observed URL set must equal the grant set. The receipt root is checked before grant for exact canonical href, same literal Scope prefix, one nonempty ASCII token suffix, no query/fragment/userinfo and maximum8192 characters. Because both Scope and suffix are ASCII, the accepted root's UTF-8 bytes equal its character count. No hidden member receipt, discovery, Types, restart or pagination request was found. Authorization-view values continue through the existing resolver channel; no credential firewall change appears.

RU uses the real reference Task descriptor, which has no owned outgoing groups and accepts the supplied title property. Its exact Resource shape therefore correctly omits ownedLinks, unlike the TX fixture. Both singleton and sequence payloads have the proper discriminator/key location; the mutation and sequence parsers are exported by @bdp/protocol. Exact created outcome/envelope, Resource GET equality and unchanged full retry result are asserted. The six-request loop rejects extra calls and verifies no Transactional context fields or receipt envelope fields. This is emitted-factory evidence, not fresh-consumer or public profile qualification.

## Lifecycle and failure handling

The TX addition reuses withInstalled/finishObservation unchanged. The existing helper catches the primary observation error, attempts child.close and directory removal, retains combined failures, then emits the safe observation report; it uses the existing90-second abort/120-second test budget and5-second exchange bound. No new lifecycle seam or broad historical helper edit appears.

The RU test registers its directory and child with existing cleanup until launch succeeds. In its local finally it removes only its own registered entries, attempts running.close and checks actual {code:0,signal:null}, then independently attempts directory removal. Both cleanup failures are appended to the primary assertion failure and emitted in the resulting AggregateError path. The safe console summary reports outcome, PID/exit, closed flag, directory absence and six completed status observations. It does not dump credentials or request headers. Startup occurs before this local try; startup failure still relies on the existing launcher's bounded kill/reap and registered directory cleanup, and should not be misrepresented as having produced the new local post-launch summary.

No currently evident cleanup-error masking or extra launch was found in the newly added post-launch paths. Actual PID absence, force/timeout state, directory removal and error-report retention still require the measured supervisor/root checks; neither this read nor a future closed flag alone proves those facts.

Only these two additions and the supporting fixture/schema/projection/lifecycle source were inspected. Formatting may still change line numbers and hashes; the final frozen council must bind the actual successor. No tests, typecheck, build, service, install, product imports or source edits were performed. This report was written through the required current-session guard; its read-only session check was the sole network action.
