# Independent branch integration closure

No integration discrepancy found. This audit read actual Git objects and checkout bytes; it did not execute product code, tests, builds, merge commands or publication. I authored the earlier prospective integration map, so this is an independent recomputation against the actual merge, not an additional independent product-code council.

Actual clean HEAD is 11742645bc591b3e5affa9fcc2f36c551c6fa803, tree 66b0cc09393fdbddf32d7ee19a2d74b961c7fdbb. Its ordered parents are [df95d1a3715a766fb5236805f8549da08230cbce, 5ca22ec4e710cc4417a80d27bff69e1d8424c115]. The sequence author is an ordinary direct child of df80f8cb86b7763bc7bb005492e43d8840527045; minimum5ca is an ordinary direct child of 2e1676caa0f5861d16c5741fbd412f6c3ef7c269. Their actual merge base is df80. No current-main ref was inspected.

The sequence author changes exactly the seven paths in frozen binding b937a736c8b4b16c10346b0f8d998640fe31415a6ee79a10188945cb4324c81a, and all seven committed blob SHA256 values match it. The actual merge changes exactly the 20 paths in current map 6180b009216b3be97fe37862c51d97c664407c9b0a5a11acffe030bd5096f050. Every mapped hash and byte length agrees with both the Git blob and working file. The binary base-to-merge diff matches 20cc9cb184da35e3ebd1ac831af255bdb24076f4938bfb918aee0ebade26280f.

I independently rebuilt all three shared files in memory using each parent's edits against df80, rejecting overlapping replacements and ordering simultaneous EOF additions minimum then sequence:

| Shared file | Result SHA256 | Composition |
| --- | --- | --- |
| transaction-http.ts | 4143e8a36021d60d1deba4649b1de5dfbcb3ee2d630486749932eccc1374c602 | Both sets of disjoint imports/routing/sequence handling edits |
| transaction-http.test.ts | f4353454c8f8a9dc42367395804bbe6309f7bcc51dba2b91b3176975d08b5cce | All existing minimum expectation edits plus complete 63-line minimum and 396-line sequence appends |
| transaction-implementation-progress.md | 9668b52c1d5cfdcc7f2a326cafe3468d87de6cca14438b73ae88816587b99c28 | Complete 29-line minimum and 3-line sequence appends |

Each reconstruction equals the actual merge blob byte for byte; all shared modes/types remain the baseline's. Every other union entry retains its owning parent's complete mode/type/object identity, including the five inherited ordinary Update paths. Across the 677 final recursive tree entries, all 657 outside-union entries retain the baseline mode, type and object ID; no outside additions or deletions exist.

Evidence pins read and checked: final-source.json 3ff3bdd8c165f14e7aef72aeafb9fa33d3dde0a6818aab53205ab627e5b00523; root-integration.json f9e71e93294323db3c8e55b643a67ea68a625af09384431576bbdc004f615e40; binding-verification.json 9b5ce4c60342db25bab93ec2fb9036171fcb4f276c6c51f918b9a899b1367cf6. Their identity/scope claims agree with this recomputation.

This closes branch composition only. Running source validation and subsequent control/installed evidence retain their own attribution and gates. PR141's successful CI used a different recorded synthetic tree and is not relabelled as validation of this combined tree.
