# Native review: compatible missing-Resource minimum

No blocking design finding. The proposed one-callback repair implements the already-satisfied, compatible case without choosing the unresolved behind/foreign versus missing precedence. Recommend proceeding within the stated five-file boundary, incorporating the two test clarifications below.

## Findings

### Low 1 — Make the broad-catch discriminator reach the Transactional callback

Plan lines64/72 require malformed-context coverage that distinguishes an overly broad missing-result catch. Specify at least one malformed **string** source position, e.g. `scopePosition: "pos-02"`, injected through the delegating real-owner wrapper with otherwise valid context. `authority-read.ts:464–470` rejects non-string source-context values before the source-condition callback. Such a non-string failure would survive even an incorrectly broad catch in `transaction-read.ts`, so it alone would not discriminate that mutant. The string above passes generic capture but fails the exact Transactional grammar at `transaction-read.ts:154–165`. Assert the resulting integrity error is neither a returned404 nor `TransactionMinimumVisibilityGap`; execute the broad-catch sensitivity control against this case. No production seam or extra file is needed.

### Low 2 — Pin owned receipt observations at their actual operation step

Plan ledger lines80–81 correctly retains the installed fixture's owned Relationship and deletes Link before source. Make its exact receipt oracle explicit: assert the three available `results` and literal canonical IDs/types/properties/endpoints, including the owned-Link result's `source` and `sourceRevision`. Do not interpret “allocated identities” as a compact-receipt `allocated` field or treat the first create-Bead result as the final source Bead image. The fixture declares `ownsOutgoing` (`transactional-collection-minimum-installed.test.ts:88`), the owned result includes source revision (`transaction-owned.test.ts:158–164`), and the batch records each outcome immediately (`transaction-batch.ts:706–707`). In the deletion receipt, the source Bead's deleted revision should equal the preceding owned-Link deletion result's `sourceRevision`; the Link's deleted identity/revision can be compared with its create result. This makes the existing two batches a strong creation/deletion witness without an extra GET or a guessed revision. No request-count adjustment is needed.

## Reviewed contract and implementation boundary

- Spec `bdp.md:1445–1450` supplies the ordinary unknown/deleted Resource404 behavior; `:5758–5787` supplies the minimum fence; `:5793–5799` requires actual observed Scope context. There is no failing fence left to adjudicate in the selected compatible/satisfied case. Hidden-policy coverage is correctly excluded from the fixed all-visible fixture.
- `authority-read.ts:977–998` materializes and checks within the same owner entry; its observed refusal retains that observation. Returning the existing refusal preserves this coherence. No second read/context reconstruction is needed.
- `transaction-read.ts:56–68` is the sufficient production seam. Snapshotting the condition and parsing context outside the special catch preserves integrity failures. Reusing `checkTransactionMinimum` retains epoch/view/position semantics. Catching only the three actual synchronous consistency-failure reasons preserves the existing missing-behind development gap and avoids introducing waits; unexpected errors/reasons must escape.
- `transaction-http.ts` already adds context to observed Problems and distinguishes visibility-gap errors from generic integrity failures. Existing conditional handling transforms successful representations only; HEAD strips the body without inventing a Resource ETag. Thus no HTTP production edit is required.
- The narrow route predicate, pre-context authorization behavior, receiving refusals, present-target wait behavior and all broader simultaneous-failure questions remain unchanged. This is not a complete minimum-position implementation or profile qualification.

## Fixture, inventory and bounds

The five-file inventory is sufficient: `transaction-read.ts`, its test, `transaction-http.test.ts`, the existing installed collection-minimum test file, and the progress document. Existing real-owner interception, injected minimum-wait observation, installed raw transport and cleanup helpers provide the necessary seams. No authority/store/runtime/parser/framework edit is justified by this plan.

The ten-exchange ledger is internally consistent: two batches (three then two members), followed by eight distinct absent/deleted Resource/property requests; one batch URL plus eight view URLs, one child, at most two live Beads and one Link. Owned-Link removal before source deletion is necessary and correctly ordered. Two nonempty committed batches imply independent pos-1/pos-2 expectations despite multiple Resource revisions inside either batch. The all-visible fixture supports the stated claims, not hidden-target policy behavior. The remaining Bead is deliberately not claimed independently GET-verified.

Keep the old collection scenario's report/count separate. The new scenario must assert exact10 completed exchanges, not merely rely on the existing helper's80 ceiling. Preserve the actual90s aggregate/120s outer,5s exchange and10s lifecycle bounds, safe failure reporting, fresh isolated store and cleanup. No extra installed child is required for source-only sensitivity controls. Frozen fresh packed proof remains mandatory; no execution evidence is claimed by this review.

## Evidence and scope

Reviewed plan SHA256: `07fa069b8519c6fe06d74cc236b3d3e941962771140c35485a777208ea92d65b`.

Source inspection used frozen callback-repair checkout `/Users/dbox/repos/wt/vickie/vickie-tx-discovery-repair`, HEAD `3f486c245bfd31f41fd42f5acc8e0f13683bad42`; worktree remains clean. The target `transaction-read.ts` is unchanged from the plan's aebf/b20 basis. The callback-receiver repair is independent and was not modified. This review ran no build, test, service, mutation, push or coordination publication. Only this external review report was written.
