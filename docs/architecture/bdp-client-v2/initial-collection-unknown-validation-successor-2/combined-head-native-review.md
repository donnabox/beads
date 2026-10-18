# Native combined-head scope and ancestry review

No blocking scope or composition findings. Independently verified clean HEAD d4f3f4be88c7837b71d6646518ca3589d464ba3d, tree d71a5753aef97444db36d044a74d3f99c168dc29, and exact ordinary parents 1aeed4dec5dc7563c59b136d4461a6091da052ee and e34a06d9033b26488be4a03ff124b4c8665f2c68. The collection commit has sole parent 2dfd2c6cc84a28b1dd336f4184c1a5f910ce8a96.

Full recursive Git tree comparison, including modes and blob IDs, proves an exact disjoint union: four collection files and six target files, with every other entry unchanged from 2dfd. All four collection blobs match the tested successor-1 SHA256 bindings; all six target blobs are byte-identical to e34a06d. No reviewed bytes were altered by composition. The complete ancestry above 2dfd contains only the eight expected collection, conditional, diagnostic, Event and ordinary merge commits; no newer main lineage entered.

The sole production modification remains transaction-read.ts: initial Bead/Link collection classification and matched-context unknown-minimum waitability. Target changes are conditional test/data assertions, the isolated Event restart witness, and bounded Read failure diagnostics. They do not alter production runtime, shared fixture helpers, build/dependency configuration, or the collection callback. The collection change preserves current-resource and strict/continuation branches; no concrete new semantic interaction requiring another council review was identified. Exact combined-head CI remains required before integration; separately tested inputs are not a claim that this merged head passed CI.

No tests, imports, source changes, commits or publication were performed for this review.
