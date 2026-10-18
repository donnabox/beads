# Frozen council scope and delivery ledger

This is a selected-source review plus explicitly scoped successor diffs. It is not an entire-branch audit or a readiness declaration. Native reviewer is the current Codex subagent, independent of implementation authors. Claude and Gemini/Antigravity run through verified external binaries with configured models unchanged. No host Codex CLI or additional native reviewer was used. All review artifacts are outside source repositories.

| Packet | Frozen target / base | Supplied scope |
|---|---|---|
| 01–03 | `72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5` / `a9bcf9ed1d869aaba2cb49fe9cca10e3b8d52a94` | Selected normative clauses and full named implementation/evidence files. Exact ranges/hashes in snapshot-index.json and packet-digests.json. Not all59 changed files. |
| 02b | Same core | Complete transaction-store1–808, transaction-sequence1–136, transaction-batch1–538, spec3793–4099; replaces the tail Gemini02 reported unread. |
| 04 | `59915ea916a210f5ce36611eb6d799f7c3124056` / core72 | Full six-file alias diagnostic delta plus relevant spec clauses. |
| 05 | `21654be3cae51d50e75e817ef8ac57735b8ee4ea` /599 | Native-only review of one-file aggregate-test30s budget/comment. Later included in packet06 full diff. |
| 06 | `3a8f9d553a2944bc7d1d7812a284f0160dcbe798` /599 | Full14-file pending/interleaving, alias-witness, profile-boundary, History/no-op prose, routing and test-budget repair delta. Extra full helper/index and clauses recorded in06-final-repairs-index.json. |
| 07 | `486ad62d7c6ecb8e36dad4e596b1597fde425cdd` /3a | Full six-file contextual diagnostic and transaction storage-failure repair delta. Extra store/clauses recorded in07-contextual-and-recovery-index.json. |

All three providers received the same core prompt and staged source for each packet. coverage-status.json records actual completion, not intended completion. Raw outputs and per-process JSON retain exit codes, PIDs, times and target pins. Claude06 stdout emitted only the last assistant text block: recovered complete assistant text, with source provenance, is in06-final-repairs-claude.recovered-full.txt and recovery.json. Original stdout is unchanged.

Delivery failures remain preserved and do not count as reviews:

- Initial Gemini path-based attempts asked for command permission; no permission setting was relaxed.
- Stdin attempts did not see source because installed agy ignored ordinary stdin.
- Direct prompt delivery succeeded, but Gemini02 explicitly omitted the store tail and sequence. Packet02b closes that omission for allthree.
- First Gemini02b attempt exited0 with empty output after denied headless command permission. One constrained direct-answer retry succeeded; the empty attempt is not counted.

Fresh Read evidence: artifact SHA256 `6142cf73deae687bc3e9668cc014ea408fa028438a871e3c29d82dea4e81d10c`; constant first40hex `6142cf73deae687bc3e9668cc014ea408fa02843`. Run source `4f0ed57906525c36bc55b343b77e0f07edbb7553` is an ancestor of core72; direct checked diff contains only artifact and capability file. Packet03 supplied a derived current summary, not individual exchange rows. No manual exchange-level audit or qualification run was performed by this council. Parent's tests and evidence verification are separate evidence. Public write-profile claims remain disabled.

Native council executed a small CRLF parsing probe, a disposable injected sequence-cancellation cause reproduction and successor fix verification, and a database-free sequence byte-bound probe. It did not rerun the full suite. Repro details/output are retained outside repositories. No source, artifact, constant, shared database, pinned playground or Janet asset was modified.

Any descendant after486 requires its own explicit delta review; no automatic extension of these coverage claims.

## Successor dependency/retention supplement

Frozen `b7eefc9ecad02540f9c26576220841c78f1725cb` against486.08-scope.json partitions all20 changed files:08a covers9 Transactional/dependency/cancellation/progress files,08b covers11 retention/runtime/harness/diagnostic/smoke files. Both include extra current source and normative ranges, recorded in their indexes. There was no earlier council review of the PR84 dependency implementation. Native reports are complete. Claude completed08a then08b sequentially (both exit0;08b ended2026-10-11T05:13:28.872194Z); Gemini attempted both but each exited3 with RESOURCE_EXHAUSTED/429 (individual quota; reported reset165h24m34s). Neither Gemini successor attempt reviewed source. No account/model substitution or retry is planned. Core01–07 Gemini completion remains valid. Direct-carrier admission work after this head is excluded.

Final status: native and Claude completed08a/08b; Gemini reviewed neither. Later9e11bbb andcd22ac3 are explicitly excluded. Source atb7 remains clean and unchanged. No readiness declaration or whole-branch coverage claim.
