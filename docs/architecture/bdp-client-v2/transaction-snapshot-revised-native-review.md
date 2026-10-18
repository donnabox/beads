# Revised S1 Snapshot plan: native council review

Reviewed exact plan SHA-256 `e870f3456a2e255c2027df63fea70518e6a4e9b9a677a2e34bf7319c3fb46cbf` in `transaction-snapshot-plan.md` against frozen `d907a1e1576c5795c17edb5f15528f47edfd7d6b`. Read the prior native and Claude plan reviews and the revised disposition ledger; did not consult the new revised external review. Source/spec/schema inspection only, no tests or source changes.

**Verdict: bounded private capture and finite-stream authoring can start under root authorization while S1-Q1 remains pending. No further product ruling is needed for this first boundary.** The plan does not establish full Snapshot or Transactional conformance. Retained minimum requests must remain explicitly outside the implemented private contract until Q1 is answered; an internal gap is not a normative response or a passing conformance observation.

## Required earlier corrections are incorporated

- Completeness oracles are now correctly separated: a generic client detects observable graph/stream contradictions; durable storage checks prove its own recorded plan; independent fixture inventory detects coherent omissions that the wire alone cannot reveal. No new manifest count/digest field is invented.
- Snapshot access depends on independently authorized current Scope/view, not receipt-style creator isolation. An authorized second principal in the same fixed view is a positive case. Supplied foreign tokens do not prove mutable policy rotation.
- The post-commit/pre-delivery failure case explicitly permits an undisclosed durable capture consuming capacity until fixed expiry. Pre-commit rollback and uncertain-commit fencing are distinct tests.
- The expired-issued/unknown distinction is preserved by bounded lifetime disposition retention rather than an unsupported forgetting analogy. Bodies expire; compact issued dispositions do not silently disappear.

## Capacity, representation and serving boundaries

The conjunctive 4,096-row/32-MiB per-capture and eight-live/64-MiB aggregate ceilings are coherent. They are maxima applied together, not simultaneous entitlements. The separate 65,536 issued identities at at most 1 KiB compact metadata each bounds lifetime disposition storage to at most 64 MiB of encoded disposition metadata, excluding SQLite overhead; a finite installation can stop new creation without evicting existing promised handles. The plan also bounds captured page-plan metadata by the bounded row count and explicit envelopes. Keep actual encoded accounting separate from heap/RSS measurements as stated.

Permanent row/byte/issuance refusal uses existing size-family `limit-exceeded` 413 with no invented limit identifier. Temporary live/body pressure uses existing `temporarily-unavailable` 503 and Retry-After. These are reference capacity choices, not new protocol error codes or universal numerical promises. No capture ceiling is reused as mutation admission.

The corrected full-record R premise is supported by actual source: `resource-evaluator.ts:1008–1014` materializes subject/source records, `:1124–1165` includes inline owned Links, then `:1050–1058` serializes and enforces representationBytes. Thus the suggested extra 8R expansion was a false premise. The near-R owned source and one-byte overflow controls remain important implementation evidence; preserving the exact canonical record avoids an unaccounted read expansion.

The proposed page inequality has sufficient conservative headroom: R=262,144 plus 6*(65,536+1,024)+1,024 is 662,528 bytes, below 4 MiB. It is a first-item feasibility argument, not permission to skip actual `{items,next}` accounting. Two independently bounded initial pages plus a separately bounded 2-MiB manifest envelope gives the proposed 10-MiB maximum. Implementation must derive the exact bounded field/token contributions and enforce actual serialization as the plan requires, without allocating a giant proof string. Empty erasures are valid only for the current no-erasure installation.

## Lifetime, ordering and integration

Capture is confined to one guarded owner entry/SQLite transaction. Count and aggregate-length checks precede bounded identity enumeration, per-row encoded-length checks precede loading bodies, and canonical-ID ordering must use the actual UTF-16 code-unit comparator rather than assuming SQLite byte order. Exact immutable rows, ordinals and page ranges avoid coupling future pages to mutable Resources or new configuration limits.

Configuration/format/capture compatibility remains inside locked writable open before BDP maintenance. Necessary SQLite hot-journal recovery remains allowed before metadata is readable; byte-identical normal incompatible-file refusal is a distinct assertion. No migration, reseed, identity-policy rewrite or historical evidence relabeling is authorized by this slice.

The no-pruning group store already retains every suffix needed by a valid Snapshot. Reuse the real G1 checkpoint codec; `genesis` is an exclusive checkpoint and `pos-0` a position, with neither claiming served discovery. Store capture, service paired observation and existing private HTTP composition are the correct seams. H1 should consume this same owner/configuration later, using a configured creation target and actual returned continuations; this is not cold discovery evidence.

No-ETag conditionals are supported by the existing shared implementation: wildcard If-None-Match can yield 304 and unmatched If-Match yields 412. Current context on those responses and all applicable auth/expiry/integrity fences remain required. Q1 still concerns which position a minimum on an immutable retained body constrains; importing a receipt-only rule is an interpretation, not established Snapshot law. Recording both captured anchor and current same-entry observation keeps the eventual ruling implementable without changing capture storage.

## Authoring acceptance boundary

Proceed with capture tables/configuration, bounded canonical row capture, immutable finite stream reads, expiry/disposition accounting, owner failure/restart controls and the three distinct replica-oracle classes. Creation with settled complete minima may reuse the existing finite wait loop. Retained reads without minimum controls may be implemented privately with the stated current observation/immutable body distinction. Do not claim retained minimum conformance, public discovery, erasure/policy rotation, SSE, installed evidence before H1 execution, or exhaustive heap capacity.

No new blocking issue found. The old introductory sentence saying original G1 external review remains pending is historical provenance, superseded by the now-completed component review; it does not affect the implementation boundary or need to delay authoring. Final source still requires independent native/external review and actual executed evidence at its own exact head.
