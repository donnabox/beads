# Native exact lint correction closure

No actionable blocker. The frozen successor implements exactly the released one-block correction in packages/server/src/transaction-service.test.ts, around lines 4264–4280. An unsuccessful observed close now appends the same error directly; the existing post-close ownership probe, assertion and removal occur only in the successful branch. The surrounding catch still records close-bound, probe or removal exceptions. This preserves the former locally caught throw semantics without noUnsafeFinally.

The single 1000ms bound, once-only close, clock/release/operation settlement, final counter checks, diagnostic report and final AggregateError are unchanged. Failed close skips probe/removal and preserves its directory; successful-path workload and counts are unchanged. No suppression or broader warning cleanup was introduced.

Independently checked all 679 current source hashes, exact base diff and the supplied forward/reverse snapshot deltas. Only transaction-service.test.ts differs from the formatted predecessor; the other 678 files are byte-identical. The before snapshot equals that predecessor, and inversion recovers it exactly. Production, installed historical prefix, selection, report schemas and workload remain byte-identical. Original author 43-entry, correction1 23-entry, correction2 15-entry, formatter 33-entry and current 13-entry indexes and references remain intact. Original capped and expanded lint failures retain their original attribution.

Reviewed pins:
- Source binding: b79e0ab9b750bfa0784effecabdc56e9a7bb19c484aaaa0293f0a1c736c2f264
- Correction: fe1067e3a5e9c3099de0020c8506d750715dd372ce1d84eb4f820c50d8f62b3d
- Full patch: 5c0e0522280e7054222573f7e1817275d10fb87a6f26bb1d48d23f845d2aa30f
- Author index: a596cdad4c236b8aa8ed7124b72d061c2ccdaac68787375c5ec8fc7da233a646

The completed substantive CODE review plus prior exact correction closures remain applicable to this bounded equivalent change. This is source review, not runtime approval or evidence that lint now passes. No product/helper execution, tests, imports, formatter or source edits were performed. Fresh validation and root release remain required; only the separately attributed original two bootstrap prerequisites may be reused under the reviewed gate.
