# Independent native Selector repair closure

**No actionable delta findings. Both native findings are addressed in source; actual compiler and runtime validation remain pending.**

1. The confirmed TS2352 adapter is replaced with the established plain delegating overload wrapper in `transaction-submission.test.ts`. A stable local captures the real owner, the real vi.fn still delegates to that owner, and the original zero-admission assertions remain. No unknown cast, fake owner, additional submission or execution branch was introduced. This is statically consistent with the already working local wrapper pattern; it is not a claim that a new compiler run passed. Original failed compiler evidence remains unchanged.

2. The existing parser UTF-8 case now constructs a valid literal containing é plus an astral scalar, independently pins25 UTF-8 bytes, pads4071/4072 ASCII spaces and checks grammar-analysis bytes4096/4097. This grammar-only entry must accept both; actual admission's existing boundary tests still decide4096 versus4097. The existing preparation case constructs distinct outer JSON raw-é and literal-backslash-u spellings, verifies equal decoded Selectors, and compares actual normalized identities and full normalized values. These assertions close the missing planned coverage without changing registration counts. They remain unexecuted.

I read the complete correction and report, inspected the changed tests and independently compared all twelve source hashes in both workspaces. Exactly three test files differ from the original frozen draft. The other nine paths, including all production/spec/HTTP/installed files, are byte-identical. The original checkout still matches its frozen dca25dfc manifest for all twelve files, preserving the ongoing external review target. I make no claim that that external review has completed.

Repair HEAD remains ordinary parent `757aedc8008386745d59d5005498a9256d6ea76e`, tree `734bf9daf40fd17d0fe2bf86b085598589122f2f`; the twelve-path source draft is uncommitted. No unrelated source paths appeared in status. Every finalized packet-hashes.json entry matches its file.

The new command ledger is exactly the old ledger with workspace and evidence-root paths substituted: cwd, Vitest paths, output files, derivative copies, consumer and required installed-entry environment path all point to the repair locations. No selection, gate, timeout or environment policy changed. Data-only comparison against the extracted literal full titles yields67+6+13+4=90 source positives, one selected case per control and one installed case. requests.json, controls.json and both emitted-only source-illustration patches are byte-identical to the finalized original packet. Source94HTTP/39owners and installed25HTTP/1child therefore remain the previously reviewed prospective workload; the new assertions add no owners/HTTP/listeners. The two control maxima and separately gated package/installed requirements are preserved.

This closes the native source-delta review, not execution/council acceptance. Root still owns remaining external-review disposition and bounded validation release. Reviewer performed only file/Git/hash/JSON/title-pattern inspection; no product imports, compiler, tests, builds, installs, services, source edits or network publication. This guarded report is the only reviewer write for this assignment.

## Exact bindings

- `source-binding.json` SHA256 `03a861851a195d5da505f311291bee6d3cd98f63866d4b6df756e2c3ae289afd`
- `correction.patch` SHA256 `56305491aa6d8c9c57f3bf2ec880a508a595d88403fd73f531bd2c81c0326eca`
- `draft.patch` SHA256 `4837959bceec2e16358c46c11a6a2676d113bbaad49ba8b3d184ebd34a366b24`
- `report.md` SHA256 `bc07b8392d0254a9c6e1beea2f62309e2e1cea5bca33bd23d56662e464a7bc2e`
- `commands.json` SHA256 `39997323fe8314348aababd29199b55371e3204b6e137ac010f1b972836c0683`
- `preservation.json` SHA256 `62320c47760a9febfe270765697941fb604219f70c430043f6af21b332155f76`
- `utf8-measurements.json` SHA256 `b36f394be7b1bf1c063a4c2377ccf13f60640577f42c02c89a63e13a9eec0f5e`
