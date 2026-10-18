# Native closure: maintenance-fault shortcut plan v2

## Verdict

**No actionable findings.** Revised v2 closes external F1–F8 and native N1 within the original five-file serving repair. This is plan closure only: no source authoring or execution is released, and PR120/main ancestry and later measured preflight remain root gates. No product ruling is required for the settled unexpected-fault classification.

Exact plan SHA256 `72a06f8e370bc6ed5df10d835464cb102c3c8cad09076e0b7b879b365ea8ee85`; revision map SHA256 `5156bfd9c425896abf7418390abde166680550b1c36ecdd42876147aeeb22079`. Reviewed against frozen `70ac666002708202d2f93195eed90dfe8541300d`, tree `1353e8f4fb0b1149b3f4ff158335804bf3f1cecc`, `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair`.

Read the revised plan/map, full root dispositions, native v1 review and full external review. Inspected actual store/owner failure guard and timer, observed/unobserved submission test fixtures, selected dependency/alias/service facilities, and emitted-runtime HTTP fixture/handler. No tests, builds, installs, services, network or source changes were performed. The original maintenance plan and frozen source were not modified.

## Findings closure

- **F1/native N1:** the proposal explicitly uses spread adapters with `vi.fn` delegates for `admit`/`admitObserved`, preserving real closure-backed epoch/fault state. This matches existing submission tests and avoids replacing frozen owner properties. Both admission delegates remain independently checked for zero calls.
- **F2:** the actual valid authorized POST fault is caught by `transaction-http.ts:584–617`; `empty(500)` at614, non-attesting marker at616 and POST `vary: Accept` at617 are the deterministic path. V2 excludes the runtime fallback and pins500, empty body, length0, private/no-store, marker/vary plus absent Problem/content-type/retry/location/limit/context and no sentinel anywhere on the wire. This is stronger than a generic500 and leaves malformed/auth/media ordering unchanged.
- **F3:** new submission maintenance variants use plain owners, without dependency reservations or `cancelSequence` on a faulted owner. Fixture close is safe after sticky fault. Timer restoration, primary/cleanup failure aggregation and no faulted-owner teardown entry are explicit drafting requirements.
- **F4/F5:** corrected anchors match the current fixture76–175, override spread91, inventory109–124, runtime seams51–65, stop52–55 and final observation1726–1740. No source API expansion is needed.
- **F6:** test clock returns its own literal1000; fixture `setTime` is expressly excluded because the override replaces that closure. The inventory oracle retains the genuine parsed seed200 receipt, with separate independently authored seed facts, rather than guessing opaque IDs/revisions/timestamps.
- **F7:** selective interval/fetch compatibility is now the first seed POST in the already planned HTTP case, not a new probe. Production timer reachability is source-grounded; bundled-fetch compatibility is correctly unproved until the approved run. Failure preserves the attempt and stops, without sleep, timeout increase or replacement request.
- **F8:** the ledger distinguishes directories, owners and read-only handles. It removes the loose24-store ceiling and preserves one controlled HTTP test, one old-guard derivative, and the existing bounded regressions. Details independently checked below.

## Independent static workload reconciliation

Three selected store cases each call their ordinary fixture once and do not reopen. Four proposed submission variants each use a new plain fixture once. The two existing submission cases each use one fixture and a sequence reservation inside that owner, not a second owner. The selected service finite-budget case, dependency static-mismatch case and alias bounded-deadline case each use one fixture and do not reopen. This gives12 positive cases/directories/owner opens before HTTP.

The new HTTP positive adds one directory/owner, two requests/one listener, and one post-close read-only inventory connection. The existing fencing HTTP case at867–890 has exactly POST, GET, reopen, POST: one directory, two owner opens and two listener lifetimes, three requests. Thus14 positive cases =14 directories/15 owners/1 read-only handle/5HTTP/3listeners. The single old-guard derivative adds1 directory/1owner/1read-only handle/2HTTP/1listener. Total **15 case executions,15 directories,16 owner opens,2 read-only inventory handles,7HTTP,4 listener lifetimes,zero installed children**. This is a static proposal, not a measured outcome. The18 transaction-file SQLite handle lifetimes include the2 inspection handles; they are not18 stores. Later draft must bind actual expanded case titles, bytes/routes and exact cleanup paths before execution.

The new case's receipt-inventory helper first closes the runtime, opens one read-only SQLite connection, and closes it in finally. Calling that helper exactly once also on the old-guard assertion-failure path is feasible. Global fixture teardown must still restore/remove its registered resources; v2 appropriately demands preservation of the primary named503 mismatch and separately adjudicated close/inventory completion. It does not infer full graph preservation from receipt inventory alone.

## Correctness and remaining gates

The minimal production predicate is sound: only benign closed owners return undefined; retained ownerFailure/maintenanceFailure fall through to existing `enter`, which throws before reading metadata. The maintenance callback stores its failure before setting closed synchronously, so this does not invent a race or touch a closed database. Explicit normal close and unobserved healthy result shapes stay unchanged. The one production comparisonEpoch caller and the observed submission adapter both reach this closure; dependency/alias comparisons exercise the same seam. Sequence execution already enters the owner and is not rewritten.

The baseline derivative must restore only the old unconditional closed shortcut and reach the named second-response500 assertion with actual503, after passing seed/timer checks, while preserving cleanup evidence. Generic setup/timer/transport failure is not sensitivity. No installed fault-injection seam, fake live epoch or synthetic context is proposed. The optional separately accepted healthy installed14/1 remains a regression only; it cannot establish fault qualification.

Root must supply verified-main-after-PR120 ancestry (or an explicit later re-anchor), release the five paths, accept the normal-format measured draft and authorize execution. Historical30035/70ac18/five and27/two evidence remains historical; the progress residual will be superseded by a separately source-bound note. This review grants no public profile/qualification/merge claim.
