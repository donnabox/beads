# Graph export discovery council

Native Codex and actual Claude reviewed the discovery/acceptance plan; Gemini
was excluded as requested. [Native review and repair confirmation](native.txt)
and [external raw review](claude.txt) are preserved. The external invocation
completed successfully; this is a plan review, not an exporter readiness review.

Both native findings were repaired: a controlled writer-interleaving snapshot
proof is required, and transaction/connection cleanup must succeed before any
publication, with rollback/close fault checks.

## External findings disposition

- H1/H2/H3: added all existing projection seams, their divergent aliases and
  attribution shapes, and the per-member disposition table. The internal reader
  deliberately retains source values; the file projection still requires owner
  coordination and a fidelity test. BDP Read DTOs are not an export default.
- H4: current-only deleted/reserved omission and the source never-reuse limitation
  are explicit. Disagree that exporting current state necessarily violates
  never-reuse: source authority cannot be recreated by a new authority. A future
  importer must not interpret omitted source identities as certified free.
- H5: corrected graph Memory to title/body and excluded the distinct classic
  configuration-key Memory plane pending a contrary contract decision.
- H6: documented complete Memory inclusion and private file mode as the proposal;
  policy remains a coordinated decision, not an implemented privacy default.
- H7: made the extra same-transaction reader mandatory. The literal assertion
  that comments cannot be read inside a transaction is inaccurate: native
  transaction-aware SQL readers exist. The public IssueComments method alone
  is insufficient for snapshot assembly; that consistency finding is accepted.
- H8: made graph/native dual identity and approved source-binding provenance
  explicit. Do not export canonical filesystem WorkspaceID or credentials.
- M1/M2: explicit mode-independent storage reader, static-credential ordinary
  server qualification, current shared/proxied refusals unchanged.
- M3/M4/M5/M6: explicit dispatch/admission/flag/proxy guards, complete-reader
  refusal bounds, strict canonical encode/decode oracle and owned/top-level Link
  consistency. Do not configure a paginated list as an export reader.
- M7/M8: explicit lease disposition and separate formatVersion proposal. No final
  format semantics are inferred from the review.
- M9/M10: atomic rename is reader atomicity, not power-loss durability; clean
  owned temp files on ordinary failure and disclose crash residue. Whole-store
  mapping validity is a stated current-reader precondition.
- M11/L1: preserve graph execution gates and avoid the two shared graph files
  touched by PR86. Disagree with forbidding every shared non-graph file edit:
  early graph dispatch in export.go necessarily shares a command file. Guard
  the branch and run ordinary-export regressions rather than claim zero risk.
- M12: determinism is unchanged source state yielding identical bytes, including
  observed timestamps; no fabricated export-time clock.
- L2/L3: corrected DOT/HTML (Mermaid is another command), and explicitly require
  defining cancellation/timeout coverage through publication.

Several external findings describe text already expanded while the read-only
review ran. The final document consolidates those clarifications. Format and
CLI decisions remain open on the bus; no format, protocol, runtime or release
qualification is claimed. Next independent slice is internal bounded acquisition
of records/Types/comments/exact-prefix/deleted count in one rollback-only read.
