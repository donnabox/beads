# Operational profile capabilities after PR102

Source inspected: main627b560c plus PR102 head61ed530a, now merged as581bad0de6f5eb098b920147f6a7e09809c553ea. The PR changes tests/manifest/coverage only; runtime is identical. This is a source and existing-evidence audit, not a new execution or profile qualification. Supporting the sole design.md plan.

## Practical boundary

Both write engines have working private/development services. Normal bdptest --serve still composes the qualified Read service and public admission rejects both write profiles. The installed candidates use private launch/configuration paths. That product-activation gap is distinct from implementation and evidence gaps.

## Read+Update

Working bounded paths include Bead/Link create/read/update/delete, alias management, property and metadata patches, revision guards, schema and owned-Link validation, ordered sequences with local references and partial failures, durable idempotent retries, and resulting-state reads from the same owner. The private installed candidate also navigates canonical Scope discovery and Operation Directory with configured credentials/policies/Types. These have actual packed-package HTTP observations, not fabricated replies.

PR102 adds real refusal/unchanged-state/correction observations for schema-invalid PropertyChange on Beads and Links; it does not add runtime behavior. Existing installed evidence does not yet establish the complete profile's concurrent revision races, pending/dependent retries, expiry, disconnect/recovery, changing authorization/nondisclosure, pagination consistency and all alias/ownership cases. Some already have component or separate-process tests. Do not label all unqualified contracts unimplemented.

Public startup/activation remains unavailable. The exported development API is explicitly nonattesting; the private installed candidate is not a public CLI/profile factory.

## Transactional

Working bounded paths include atomic singleton/batch Resource mutations and owned-Link effects; valid updateWhere/deleteWhere; rollback/no-op behavior; receipt roots, pages, GET/HEAD and ordinary restart; retained retries/failures; same-owner current reads; sequences and local dependencies; finite Scope/Resource Events and live SSE; finite/SSE changefeed; durable paged Snapshot and feed convergence. Evidence varies by surface: latest PR101 specifically executed receipt pagination/reopen, while older snapshot/feed observations retain their original sources.

Complete installed qualification remains missing for pending/recovery/expiry/fencing, sequence dependency/receipt-namespace laws and cumulative inherited contracts. Separate component evidence is not a public profile claim.

Concrete unavailable or incomplete paths:

- Scope discovery, bdp.json and Operation Directory intentionally return a private development gap, preventing a cold client from discovering this service.
- Mutation minimum-position controls, Snapshot minimum controls, and collection/alias minimum reads are refused; full consistency/refusal ordering remains incomplete.
- Sequence Resource-alias operands are refused.
- Unsupported Selector/depth cases remain explicit private gaps rather than complete protocol admission behavior.
- Snapshot/feed views are fixed and all-visible with empty erasures; changing views and live erasure are not implemented.
- Incident-Link metadata-only changes into Bead Event Sources have an explicit projection gap.
- SSE start=now combined with Last-Event-ID is refused.

## Sequencing recommendation, not an adopted redesign

Do not expect a large total-effort saving merely from completing TX before working on RU. They already share Resource/alias evaluators, schema contracts, numeric handling, selectors and Read components. RU uses the recovery-store/member-executor path; TX uses transaction-store and transactional receipt/epoch/change-group semantics. Their result shapes, admission, retries and retention contracts differ. TX observations cannot be relabeled RU evidence, and public TX also needs its cumulative inherited obligations checked on its own served owner.

A TX-first product milestone is reasonable if that is the desired first consumer. Prioritize its actual launch/discovery, consistency, alias and projection gaps; keep shared fixes and small RU regression checks moving. Share test intent and expected graph states where contracts coincide, with profile-specific wire/retry/atomicity oracles. Do not wait for every RU coverage flag before advancing TX, and do not rewrite the working RU path through TX solely to reduce the evidence backlog. No quantified speedup is established.

## Source map

Paths below are under gastownhall/bdp at61ed530a (runtime unchanged at merged581bad0):

- apps/bdptest/src/main.ts and server-composition.ts: shipping Read-only composition.
- packages/server/src/index.ts:288–306: public profile admission.
- packages/server/src/development-authority.ts; installed-reference.ts; installed-reference-process.ts: mutable RU owner/private installed launch.
- packages/server/src/member-executor.ts and transaction-batch.ts: shared Resource evaluator, different execution/identity paths.
- packages/server/src/transaction-http.ts:203,297–332,391,529–548: explicit serving gaps.
- packages/server/src/transaction-service.ts:418: fixed visibility projection limit.
- packages/server/src/transaction-events.ts:257–263: incident-Link metadata projection gap.
- packages/server/src/transaction-installed.ts; transaction-runtime-configuration.ts: private installation and captured configuration.
- packages/conformance/test-support/read-update-coverage-v1.json and transactional/coverage-v1.json: evidence scope, not functionality percentages.
