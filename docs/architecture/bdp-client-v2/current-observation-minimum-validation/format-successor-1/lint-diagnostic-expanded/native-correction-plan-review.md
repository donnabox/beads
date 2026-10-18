# Native single-block lint correction plan review

Accept the proposed one-file correction. No suppression, warning sweep, new test or additional council round is needed for this exact change, subject to native inverse/delta closure and actual fresh static/runtime validation. Source authoring and execution remain root-owned.

Independently read actual expanded log7508175c665e5b41664607e003ff64bcb65638c7e7685ab61f106e26519cb862, exit and reportb998cb59, verified their index and rehashed all679 current formatted source bytes. Expanded run is diagnostic-only raw1, qualificationPass=false,0.45343354204669595s; it exposes exactly one error lint/correctness/noUnsafeFinally at transaction-service.test.ts:4267. Full summary remains1 error/165 warnings/9 infos with no truncation. Original capped lint failure remains unchanged. No build or source result is created by this diagnostic.

The relevant C-finally block currently obtains a settled close Outcome under currentObservationBound(close(), "current observation drain timeout"). On ok:false it throws outcome.error into the immediately enclosing catch, which pushes the same error and skips the ownership probe/removal. On ok:true it increments postCloseProbes, asserts the owner's returned entry-refused/not-accepting and removes the directory; any timeout/probe/removal exception enters that same catch. The proposed rewrite preserves those paths exactly:

    const outcome = await currentObservationBound(close(), "current observation drain timeout");
    if (!outcome.ok) {
      failures.push(outcome.error);
    } else {
      postCloseProbes++;
      expect(f.owner.withRead(() => true)).toMatchObject({
        kind: "entry-refused",
        reason: "not-accepting",
      });
      directoryAbsent = f.remove();
    }

Keep the surrounding try/catch unchanged. Keep clock advancement, release, observed operation settlement, once-only close(), existing1000ms wrapper/timer cleanup, final counter assertions, diagnostic and post-finally AggregateError unchanged. Failed close must still skip probe/removal and preserve its directory. Direct push records the same error once, then the existing count assertion may append its independent failed-count evidence just as before. No late success, retry, extra deadline or false closed flag is introduced. This removes the syntactic explicit throw inside finally without weakening failure propagation: the final AggregateError still fails the callback.

Exactly packages/server/src/transaction-service.test.ts changes, with before/after snapshots and exact inverse reconstruction. All678 other tracked bytes, selection97, nine-new/28-legacy report contracts, service/page/HTTP/wait/clock/mutation/probe counts, installed22/new37 and production must remain unchanged. Existing successful typecheck/format checks remain predecessor evidence; rerun all19 later stages plus4 required checks on the new source, reusing only the original explicitly mapped bootstrap2 with unchanged17 inputs/11 tools. Preserve both raw lint failures and their original attribution.

The completed full actual CODE review and F1/F2 closure remain applicable because this is a mechanically equivalent handled-error branch correction. Native exact-delta closure is sufficient to disposition this additional tooling defect; another full actual CODE review would add no needed semantic scope. Any unexpected larger change must stop for review. No source edit, helper import or product execution occurred in this native plan review.
