# Operator ruling: Scope changefeed reconnect from start=now

On October14,2026 the operator answered the pending question with: “Resume after the saved checkpoint (recommended)”. The question explicitly concerned reconnecting to changes/?start=now with saved Last-Event-ID.

Accepted semantics: on an otherwise valid Scope changefeed SSE request, the saved Last-Event-ID checkpoint overrides initial start=now and selects exclusive replay after that checkpoint. Preserve the existing requirement that the client use its atomically applied durable checkpoint. Without this header, start=now retains existing anchoring behavior. Validate checkpoint authority/retention through the existing after-checkpoint path. URL validation, authentication, media negotiation and repeated-header validation remain in scope as regression constraints, not new precedence decisions. This ruling does not add start=now to application Event Sources.

This unblocks the narrow serving repair assessed in transaction-event-start-now-assessment.md. That assessment used older581bad0 source, so the implementation plan must inspect currentmain5995f6b08ebd045e4343059c2da1cffd38b09b80 before freezing edits. Minimal normative Scope wording/catalog alignment is authorized by the explicit ruling; no schema member, new cursor, runner mechanism, public profile admission or release is authorized.

Incident-Link metadata Event membership and other minimum/retention/selector rulings remain unanswered. The disjoint deletion-safety conformance author continues; root coordinates shared server files and packaging. Native+Claude plan/code reviews and exact-head CI remain required before merge. This record captures the human decision; later implementation and observations must retain their own exact source provenance.
