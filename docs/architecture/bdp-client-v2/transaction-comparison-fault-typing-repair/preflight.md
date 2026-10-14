# Maintenance-fault guard typing repair: preflight delta

Status: source preparation only; no tests, build, setup, install, services, commits or pushes in this repair checkout. Root/native gate requested before execution. Original draft remains completely frozen for Claude42482.

Repair checkout `/Users/dbox/repos/wt/vickie/vickie-comparison-fault-typing-repair`; branch `codex/vickie-comparison-fault-typing-repair`; ordinary unchanged parent `427cfef6decbb422b393024e36218affc26007fa`, tree `9075897867d88344a8477e1f4a8e1a00a50f1228`. Parent remains unmerged; no green-CI claim. This is an isolated five-file successor of the same parent, not a rewrite of the frozen draft or its evidence.

## Only correction

The initial typecheck failed TS2352 at both direct Mock-to-owner-overload casts. Root approved exactly two wrapper replacements patterned on the established existing test boundary: `admit: ((...args: Parameters<typeof f.store.admit>) => admit(...args)) as typeof f.store.admit`. The actual real-owner-backed vi.fn spy remains the delegate, and both zero-call assertions remain. No unknown cast, helper, production change, new test or workload. One TS file formatted through the parent's existing Biome (22ms, fixed1), without dependency setup.

`correction.patch` is the complete diff against reviewed draft69c3c4fd; SHA256 `86e9a8cf6d06eb80df2aab46e42e9a5ba0f2cfe261d792e0f23b9659ea8242ca`. The other four files are byte-identical to the reviewed draft. Complete new `draft.patch` SHA256 `a86a67e85b925a0fc1bcda18f0a1c797be5012fb7f733b00dd83f9adf14717a2`. Current five-file source snapshots and exact hashes are in `source/` and `source-binding.json`. All original read-only caller/config/test hashes are unchanged.

The original frozen checkout's actual git diff was checked against SHA256 `69c3c4fd2e60ff1ec069755efaac37920b55d51402d89f620f4eeb1faf17dd6c`. Original setup/build passes and typecheck failure remain under ../transaction-comparison-fault-guard; first-attempt-stop.json SHA256 `c360859e1378f9d4b2c7f5838572b5004da703803cf303a90bf12f136a037d64`. They are never relabeled as execution against this repair.

## Unchanged measured preflight

The full prior preflight remains authoritative for workload and oracles: ../transaction-comparison-fault-guard/preflight.md SHA256 `56056217cf2257dbebc89f2fc2ebe5732ad171bc712ed1d727c733a2734eb4de`. This delta supersedes only its checkout, source/diff hashes and admission-boundary syntax. `selection.json` and `requests.json` are byte-for-byte copies, SHA256 `4bf4631e67436f30f62bcd985931033f831be5b83a307ba9b68c3396722576f4` and `117449bd402a40533e995c224108c0c506b50c463bc87fb979b89a55ff5cf52a` respectively.

Exactly14 positive cases + one separately built old-guard HTTP derivative: 15 directories,16 owner opens,2 readonly inventory handles,7 HTTP,4 listener lifetimes,zero installed children. No extra cases/probes/retries. Same two POSTs, named selective interval/fetch go/no-go, exact handler500/header/body assertions, actual seed receipt inventory and finally/error-preserving cleanup. The control still requires actual503 at the named internal500 assertion plus full cleanup; setup/typecheck/timer failures do not count as sensitivity. Source production guard is unchanged from reviewed draft.

No tests were run in the original draft, so the planned14 positives and one control remain entirely unconsumed. After gate, this new checkout needs one explicitly approved offline frozen setup (it currently has no node_modules), then typecheck/fresh build and the remaining static/Read gates before the one positive run and control. Existing timeouts, maxWorkers1 and120s supervision remain. Preserve any failure and stop without rerun. A control uses its own isolated derivative and build, not mutations of either frozen draft. The optional fresh healthy installed14/1 remains separately held until source results and signed clean source freeze; no pack/install/installed execution now.

Root/native must confirm this narrow typing repair before any execution. A later docs-only factual outcome update is separate from this source-review binding.
