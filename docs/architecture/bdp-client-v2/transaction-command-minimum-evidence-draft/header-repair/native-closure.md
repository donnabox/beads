# Native narrow header-repair closure

No unresolved finding in this exact repair. Confirmed the complete delta consists of the two reviewed hunks only: authored-prefix plus boolean no-BDP-suffix check, and explicit whole-json replay-mode guard.

- Preserved before.ts SHA256 7bcb736e933c26d603ca40244e08eabef62f53a3894f988a2235751441796fe9.
- Successor app SHA256 421eeff168aa54b4d97893c81ebe1b9f790f9f4cf2e015a9fc6aac552b0aafb0.
- Direct before/after unified diff SHA256 70255304bc53601ee3cd58fee0eff8c977ad21204d5b5a2dbb923a43120e5429 (absolute-path headers).
- Base remains453b6dab386746bb03aac2c0a293572b8a71c969.
- The other four files match every hash in final-code-source.json exactly; only the same five authorized product paths are dirty.

The prefix comparison at app2364 matches the serializer's actual authored-first order and retains authored Accept/empty fields. The suffix check converts its scan to a boolean before assertion, so it does not dump credential header values; unexpected bdp-* fields still fail. It permits the actual generated default headers that caused F1. Existing runner request-authored-headers assertions still enforce the entire credential/default/framing order, count and values; request-authored-body and write-state checks are unchanged.

The pair[2] guard immediately before baseline lookup/equality makes whole-json explicit at use, using the existing named ledger error. Baseline validation, complete body equality, performed-pair recording and finish inventory are unchanged.

No transaction-field change or weakening of Resource/Event equality. N1 onObserved retention, all report bindings/control failing-ID lists,192manifest assertions,11actions, exact source fences and63000/90000/120000/5000/10000ms limits remain unchanged. No new callback registration, request, child, runtime/API/config change or qualification credit.

This is static closure of F1 and accepted F4 hardening, not actual test/wire success. No tests, imports, build, services or Claude launch performed; root owns the separate actual supplemental review and later execution release.
