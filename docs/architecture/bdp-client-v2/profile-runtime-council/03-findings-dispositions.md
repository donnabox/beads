# Packet03: independent disposition of Claude's findings

Source packet remains frozen at `72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5`. Claude exited 0 and reports all supplied source complete. The full raw review is preserved. These dispositions distinguish current observations, evidence limitations and requirements proposals; no source changes were made.

| Claude finding | Disposition |
|---|---|
| A1 targetProcess digest includes ephemeral paths | Confirmed limitation of recipe reproducibility. `generate-read-cohort.mts:590–607` hashes baseEnv including temporary HOME/config paths; the digest is run provenance, not independently reproducible from the commit alone. The comment at320–329 overstates reproducibility. This is preexisting binding design, not an invalid current observation. Normalize a future recipe binding only under an explicit provenance change/reseal plan. |
| A2 anti-relabel claim | The comment at `cohort-verification.ts:323–326` overstates protection against an editor who fabricates all packaged-only digests and recomputes the artifact seal. Those three digests are checked for shape. This verifier is not a proof against a malicious committer. Reject the suggested automatic ban on packaged self-certifiable rows: `cohort-artifact.ts:355–377` defines rows that MAY run in-process and the current admission rule is one-way. A stronger provenance requirement is a separate proposal. No evidence of actual relabeling in the current artifact. |
| A3 missing development marker on empty/redirect responses | Confirmed low consistency improvement: `empty()` and alias 307 omit the marker. It is not the profile admission mechanism and its absence cannot itself turn this private development server into attested production evidence. Follow-up may uniformly add it without claiming a protocol requirement. |
| A4 unknown development POST returns405 | Already repaired and reviewed in packet06. Query-bearing mutation targets remain deliberately strict; supporting query variants is not established by this finding. |
| A5 limits unreachable | Properties input budget can be masked by equal total body budget; this is a declared unqualified boundary, not false coverage in the current plans. The broader assertion that representation/diagnostic limits must be below input cap is unsupported: they govern generated output too. No new public semantics or limits changes proposed. |
| A6 authority removal | Intentional canonical public origin behavior at `server/src/index.ts:414–432`: transport authority never defines protocol identity. Development rejects `//` separately. No demonstrated security defect. |
| A7 empty query target refused | Verified helper's strict comparison rejects a bare trailing '?'. Extracted behavior predates this branch. No requirement requiring acceptance identified; record strict raw-target scope rather than relax normalization automatically. |
| A8 empty actions suppress requests | Rejected by upstream manifest parser: `executable-manifest.ts:520–540` requires exactly one of requests/actions. |
| A9 malformed pointer | Rejected upstream by `readJsonPointer` at2305–2321 (leading slash and escape syntax). Internal lookup receives validated pointers. |
| A10 runtime digest shape vs gate | Declared build/CI evidence architecture, not a runtime cryptographic verifier. `read-conformance-capability.ts:11–16` explicitly names evidence:verify as the gate. A forged source constant without running required gates is outside the trusted-build assertion. |
| A11 ignored rejecting finally | Hypothetical catch-handler throw, no reachable double-response trigger provided. No confirmed process crash; keep as defensive bookkeeping follow-up only if reproduced. |
| A12 epoch/view conflation | Private development authority's bounded lifecycle design, not a public Transactional or authorization-identity claim. No diagnostic currently claims epoch/view independence. |

## Omitted-context questions

- **Q-A:** Fresh observations are real: runHead `4f0ed57906525c36bc55b343b77e0f07edbb7553`, artifact SHA256 `6142cf73deae687bc3e9668cc014ea408fa028438a871e3c29d82dea4e81d10c`, corresponding constant first40. Run-head→core72 changes only the artifact and capability constant. Parent independently verified current evidence. The capability comment at18–22 is stale: it describes historical schema-projection migrations, not this fresh run.
- **Q-B:** Cross-packet owner review covers comparison-before-admission and acceptability. Missing context in packet03 is not missing implementation. Invalid carrier/body refusals do not establish a 406 precedence violation.
- **Q-C:** Shared protocol types distinguish Problem Details `code` from success/sequence envelope; focused tests cover the envelope rendering. No demonstrated misclassification.
- **Q-D:** Development factory is excluded from the public conformance index and package API. Diagnostic kind/Read+Update rows do not certify normative rows. Shared immutable-bundle provenance is not profile admission. A malicious source deep-import with invented catalog bytes is not a current public write claim.
- **Q-E:** Canonical artifact verification and fatal UTF-8 gate cover the actual CLI path. Catalog/manifest byte bindings and derived IDs supply identity beyond informational catalogId. No accepted mismatch demonstrated.
- **Q-F:** `vitest.evidence.config.ts` supplies 1,800,000ms test timeout and120,000ms hooks; the alleged 5s default is inapplicable.
- **Q-G:** `read-harness-bindings.ts` binds generator bytes plus observer support. Generator contains the extra external-endpoint seed recipe; workspace digest alone is not the full provenance chain. Installed-payload transitive closure remains explicitly unclaimed.
- **Q-H:** setup.requires controls fixture actions and applicability.requires gates scenarios; distinct roles alone are not contradiction. No concrete misclassified current row supplied.
- **Q-I:** No missing required ETag target demonstrated. The review covered current Read helper behavior, not exhaustive optional endpoint validation.
- **Q-J:** Recovery store startup and retention configuration are covered by packet02; private independently configurable retention mismatch is recorded as future construction hardening, not current deployment failure.

## Safe treatment of stale capability prose

Suggested replacement for historical projection-migration sentences:

> The gate recomputes the Read schema projection bound by each segment as `schemaReadProjection`, using the sealed definition set in its declared order. The current artifact records observations from its `runHead`; its digest determines this constant.

**Do not apply that comment alone under the current seal.** `scripts/read-cohort-evidence.mjs:120–163` identifies evidenceCommit by the latest commit touching either artifact or capability file. A comment-only edit moves that commit forward; runHead→new evidenceCommit would then contain intervening runtime/diagnostic files and fail the confined-delta gate. The file is not part of the explicit observer-support hash list, but it is an evidence-path trigger. Defer to a legitimate future reseal or separately review evidence-commit identification rules. No constants, artifact bytes, or verifier rules were changed by this review.
