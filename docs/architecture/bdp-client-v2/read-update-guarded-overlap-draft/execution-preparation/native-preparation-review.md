# Independent RU execution preparation review

1. **Low — stale example preparation pins** (`execution-gate.example.json:26,33`). The closed example still pins the former gate and commands bytes (113ecbd… and a6a04b…), while the actual frozen packet uses 87d8d7d… and bfddc36e…. `load_gate` checks every preparation pin, so a release copied from this example will stop before its first command even after filling the source/merge placeholders. Replace those two example values with the current hashes during the already-required mechanical rebind. No execution is currently enabled; this is a fail-closed preparation defect.

No other actionable finding from this bounded read-only audit.

The 22-stage ordering stops dependent work after an unexpected exit, failed selected test adjudication, source drift, timeout or incomplete owned cleanup. The actual merged parent and two-path repository source binding remain deliberately unresolved; the hard pending binding and missing live gate prevent use of the external preview as executed repository source. Later packaging requires a clean ordinary child, exact two-path committed scope and equality with the recorded build inputs/outputs.

The direct Node/Vitest selectors match the exact existing preview title, including the literal plus sign. Positive and negative select one callback each. The negative adjudicator binds one failed case to the retained named oracle, two schema-valid 200 responses, six completed requests, no winner and clean teardown. In the frozen preview, `namedOracle` is assigned immediately before the sole `expect(valid, namedOracle).toBe(true)` and cleared only after it succeeds; its failure count/termination/cleanup checks distinguish this counterfactual from earlier transport/startup/schema failures. The postprocessor does not independently reconstruct full Resource bodies or parse the assertion text; the frozen callback supplies those full checks and root must still read the actual failure message, as the report states.

Package verification compares archive, source and installed bytes/manifests, the exact eight-package dependency closure, lock and entry identity. Derivative preparation confines symlinks before and after the independent copy, preserves copied topology and all regular bytes except one exact emitted expectedRevision guard, and checks the actual installed entry-to-resource-evaluator chain plus dependency resolution. The selected guard is the shared evaluator's real mismatch refusal, not an unrelated diagnostic branch. Its removal leaves each candidate update valid, making the two-success negative oracle feasible. The later Node dependency check and actual child payload comparison complement that byte binding; the derivative payload helper is correctly described as reused rather than independently implemented.

The supervisor drains its process group, records sampled PIDs and owned temporary roots, preserves unexpected entries for review and fails on forced cleanup. Sampling cannot prove that no short-lived detached child ever existed; the packet explicitly retains that limitation. The selected callback's normal child exit, actual PID absence and removed store provide separate cleanup evidence. A successful future run remains a private ordinary Read+Update witness, not public/cumulative qualification or a proof of simultaneous server admission.

## Basis and limits

Independent of the packet author. Fully read all eleven files below, the complete external preview test and relevant actual installed payload/evaluator/import-chain source. Performed only filesystem/Git source reads and data-only JSON/hash/selector inspection. No packet import, product import, command execution, setup, build, package, install, service, test, remote mutation or source edit. PR128 CI monitoring is a separate task. Original packet remains unchanged. Root retains release and evidence disposition.

```json
{
  "report.md": "10bb843c6a590bc5c74a19ac1b2a9cab86c7d6bdd16c5c9f687975b73fafab04",
  "commands.json": "bfddc36e0d78a8f36ca262e314b293b1a0420aacca7095f7d30bf4769badc7ca",
  "gate.py": "87d8d7dc5b6bcbadc697e568e3acc73f74435d5f5a46bab7f3e9e02c18462502",
  "supervise.py": "6a67e588e01e319467ddc82fd2e84a2ae99441948dff1997f50dcaed65c94c4c",
  "packages.py": "29d90c563efc6a072f0cbdb9ada74547afdb76153cedfa9635f063bbf503cc48",
  "derivative.py": "77c3fc32cba5826614a7fcf7e9e6f87161d8ec9e8cac329ae18364bfad9394c2",
  "derivative-payload.mjs": "6d6d0e999b8507835a411d25da60fee6cdb61a4ba56effaf3b7e6179ecdc46b4",
  "review.py": "2cdbc04ed6a75b4e133e869bb5b9de390868b77f47b1b0976c08b04da5444117",
  "inventory.py": "0fa979cd18f043396214787e20168c9c60df0af33a029a56daec8383eb680cdc",
  "execution-gate.example.json": "231d997a106d1f11c51e0aef4e52bc39e649cd3f746af737d16cfe1210f1734e",
  "preparation-hashes.json": "684af4c2ea6b66668965bf281f3281c3f82ba3cbdb816bb264671327c9708213"
}
```
