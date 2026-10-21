# PR121 base-observation adjustment review

Accept the proposed bounded adjustment. No definite unsafe gap is introduced by allowing PR base metadata to report either original7e46 or verified PR120 main454, provided the actual remote-main, tested-tree, current-head/check and final-ancestry gates remain mandatory. Do not substitute the metadata value for the actual base used by the merge-tree/final-parent checks.

Independently inspected local PR120 merge object `454e521c1bf966aa1d430e86596f3924ac7d46d0`: tree `e870145ffa8f49821ea3a1fb59c1c06506b8bf17`, exact parents [7e46cdc93b2fd02094ba7e15e57e743d2ef81daa,40806dd3ac3e74ff94c07330c89ee47dc730080c]. Local merge-tree(454,31e74290fb43974ce6f7b15ed1c421badf180e4c) independently returns `5c14bcc20ec7a72118bc07d2ed5896703211a660`.

Retained CI metadata reports successful run38625916119 on exact31e. Its log identifies synthetic checkout `75c6358221003c0c5c9bcde8c4c9d0f5f228ae8e`, terminal9261passed/1skip across172files and separate7preflight. One read-only GitHub commit API call independently confirms this synthetic commit has exact parents [old7e46,31e] and tree5c14. Thus the actual new-main merge result is byte-for-byte the tested combined tree; this is not an assumption based solely on equal PR heads.

The amended script should keep actual B fixed from the validated PR120 merge artifact454 and permit metadata base only in [old7e46,B] at both PR observations. Require actual remote main==B, fetch and verify local B, merge-tree(B,H)==5c14, exact head31e and successful current checks, then merge with head match. Continue root's serialized main ownership and final verification of actual new main tree5c14 and exact ordinary parents [454,31e]. Stop if actual main/head/checks move. These separately sampled operations are not atomic; the pre-existing limited race and final verification remain, and stale PR metadata is not permission to merge against an unverified base.

No artificial source commit, rebase or CI retry is needed merely to refresh base metadata when these exact tree checks hold. This does not waive any platform-required check or authorize another merge mode. Root still executes the mutation only after all existing gates.

Read-only local object/log inspection and one read-only GitHub commit lookup; no scripts, tests, builds, installs, services, source changes, pushes or PR mutations executed. This report records a gate decision, not an already-completed PR121 merge.
