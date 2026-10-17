# Independent static successor audit

Typecheck and format-check PASS. Lint FAILS raw1/acceptedfalse; the prefix stopped before build. No runtime release is recommended. The complete lint log reports1error,19warnings,1info and1diagnostic omitted by the display cap. There was no rerun to expand the cap and no warning sweep. Format says no fixes applied. This audit executed no helper/gate/test/build/install/tool capture/formatter/product import or service.

- typecheck: raw0, accepted=true, 4.997846s.
- format-check: raw0, accepted=true, 4.073100s.
- lint: raw1, accepted=false, 4.124000s.

All terminal start/exit records and command argv/cwd match frozen source identity `83037cbf38d884ba1133a3e633869e6fd3cbd6fa179ba9d435555b232571229e`, parent `b76c3b7df77ea9b33611bc9027db651833e20abb` and patch `c41151ea696977164e12196d106de182c16434033bdd12e44571f484dac143d1`. Independently rehashed all679 current source inputs and full patch; no untracked source. The finalized implementation index is unchanged and matches its inputs. Bootstrap reuse remains correctly attributed to old source identity `6d547d6c57fcc0f1017770cc97c2b0b977b859162aeb78538c61289be312fb7c`: exact three-test delta,17 dependency inputs and11 actual tool identities verified, with original install/capture log records preserved. Old successful bootstrap is a prerequisite only; original failed format was not relabeled.

All three current attempts report no forced cleanup, supervisor errors, survivors or owned bdp directories. Independently all4 sampled PIDs were absent and each current stage TMPDIR contained no bdp directory. No fresh-build attempt exists. Terminal lint failure/log are preserved unchanged.

Sole lint error: `apps/bdptest/src/transactional-event-missing-context-installed.test.ts:592`, `lint/correctness/noUnsafeFinally`, explicit `throw new Error("installed finite Event child close was not clean and timely")` in the outer finally's nested try. The nested catch currently captures that error and adds it to failures; the intended primary/cleanup aggregation is already explicit. The bounded proposal in lint-successor-plan.md replaces only this throw/catch detour with the identical push and an else guard around success/removal. It retains the failure outcome and preserves uncertain ownership/store. No suppression, budget change or warning cleanup is proposed. Native review and root authoring release remain required.

## Evidence hashes

- `run/typecheck.started.json` — `982569c8610ec6ce9babd1fe294ab3e3c2aa65f4f50c83e1426c904f9c7882aa`
- `run/typecheck.exit.json` — `7929fafd61e28ced2ee0f23a59d580488268f4f418ae295919a5ef4b93353254`
- `run/typecheck.log` — `91d54ffbee73a3ac72294d090e58d77745f8b93edbf94b862807698273a6505a`
- `run/typecheck.processes.jsonl` — `d99882dff387f67181e0ef8a73d8e17368762b5f9dc2210922151e03b9b02c2a`
- `run/format-check.started.json` — `4f5989f11067875032edf57bf1724d6b1505bee98dfb329b6c04ada8e7ed6686`
- `run/format-check.exit.json` — `e5a8243162a06a2b6f1a5c0fa7dfc4417d9050c43314594fb74aada4ce7d0dbb`
- `run/format-check.log` — `04666f7bea76742e24a3ac19abfe3df449c3e31dac98110180217a31355361b3`
- `run/format-check.processes.jsonl` — `0573af2f4c11ca504fc4549b597293e89e4a13eda1f3d78c43c100037fbd8e6e`
- `run/lint.started.json` — `3021949c7cbe605a68d75cdbe117854fbf5db028e8e408c1ea0865f7f39d02bf`
- `run/lint.exit.json` — `e9f9445906ca3a948aa8b5dcb0d6e492502599be3125869204fd7a05bd3bf265`
- `run/lint.log` — `d33d74a8112151ff71d2d977a9cc4c53cd08c3f878e15e625bb5eba886e03c2c`
- `run/lint.processes.jsonl` — `648d292c629b0e943e55e0bbb1edb0bba3b688c0c4e4eeb3851e2e01998963bd`
