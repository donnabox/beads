# Root council disposition — branch-only CI

Accepted exact one-line patch9109f4e5 after native review597d0031 and actual Claude reviewb9392b68 (natural exit0,214.896s, unchanged input). Root read both complete reports. No product/test/spec changes and no main reconciliation.

Claude1: the asserted current-minimum-receiving base is not the observed PR144 base. Actual PR144 base is codex/vickie-minimum-current-resource; both it and our head belong to the intentionally selected role branch family. This includes integration branches. Existing main trigger entries are retained byte-for-byte; verifying current main/default is unnecessary and prohibited here. No claim about current main state is made.

Claude2 accepted disclosure: push workflows are defined by the pushed branch. Future authorized repository writers could change their jobs/permissions; branch-name filtering does not constrain those writers. This literal patch adds no permissions, secrets or jobs. No repository security policy changes are needed or authorized for this bounded trigger repair.

Claude3: root searched the entire local .github/workflows directory at11742645 for workflow_run/workflow_call: zero hits. This establishes the frozen branch contents only; it does not claim remote/default-branch policy knowledge. No privileged follow-on is introduced by this patch.

Claude4 accepted: require API event=push/head_sha=C and checkout log SHA=C. Tree equality is derived from verified Git commit identity and the locally recorded tree, not independently measured runner file bytes.

Claude5: expected Read evidence verification is exit0, not2. Existing actual combined source run/read-seal.exit.json records exit0 at11742645 on2026-10-16T05:29:15Z, and unchanged89-row historical Read evidence. The verifier uses recorded artifact runHead ancestry; root source search confirms no origin/main/current-main/GITHUB reference. No repeated baseline run needed. CI must still execute its unchanged gate.

Claude6: preserve failures/cancellations and exact attempts. Wait for C CI before a subsequent push to this same branch. Any later justified manual rerun must retain event/head identity and separate attempt attribution; no automatic retries or force-push planned. Local commit objects and evidence remain preserved if branch advances.

Claude7: git apply --check succeeded on frozen11742645. No recount needed. Root authorizes applying the literal patch, normal child commit and guarded ordinary push; no main merge or public admission. Validate exact one-path bytes/parent/tree and remote head, then inspect the actual push run.
