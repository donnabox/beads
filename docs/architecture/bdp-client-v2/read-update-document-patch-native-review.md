# Independent native review: document-patch observations

**Bounded pass; no blocking finding.** Reviewed frozen
`a0b44bef83308cecf57d96065df191d71249119e` against direct parent
`55045386812ab3b14042ec573024a90a1b38a829`. Author checkout
`vickie-read-update-document-patch` remained clean and untouched. Independent build
and test checkout: `bdp-document-patch-native`. Review completed without consulting
external implementation-review findings. No source or outward changes.

## Source and normative assessment

Read the approved document-patch plan, actual diff, canonical rows and governing
Property changes/Property-change values (spec2197–2267), validation-failed diagnostics
(3442–3474), and row table6724–6726. Semantic JSON comparison confirms exactly one
existing manifest scenario changed (merge-replace), one added (validation); atomicity
and every other scenario remain unchanged. Coverage changes are merge-replace,
validation and the disclosed sequence-order reason. Four files only; no runtime,
fixture/configuration, protocol, installed schema, catalog requirement or admission edits.

The both-kinds merge obligation is now supported by the new unowned Link arm:
literal complete document oracles distinguish object-member replacement from merge,
array insertion/append, document-relative paths, whole-root replacement and each
selected-object clear while the omitted object survives. Independent GETs equal the
statically checked returned records. Existing Bead seed becomes the Link source;
the target is independently created/validated. Before/after full endpoint records
prove these unowned operations do not version endpoint Beads. Missing-target/null
checks remain supporting breadth rather than another row award. This supports the
bounded whole flag for this named row, not all possible JSON Patch behavior.

The new validation scenario verifies its seed records independently with literal
properties/metadata, identities, Link references and supplied provenance before
using captured full records as rollback oracles. Each invalid document pairs an
otherwise effectful opposite-document proposal. Full GETs detect persistence of
that proposal, revision or provenance changes; Link refusals also preserve endpoints.
Properties failure may short-circuit, correctly disclosed. Positive arbitrary
metadata does not discriminate Type-schema exemption in this schema-free fixture.
The row remains whole=false, with effective-Type rejection, schema exemption and
semantic diagnostic message wording explicitly missing. Existing atomicity/no-op
provenance assertions are not weakened or cross-credited.

Root-shape diagnostics use the unchanged normative wrapper and explicit
422/validation-failed/validation/never tuple, document-root pointer, absent paired
Type/schemaLocation, and metadata discriminator. Properties omission is explicitly
one legal spelling, not a claim that explicit properties is forbidden. Six controls
run after real client/serializer observation, modify schema-valid assertion inputs,
and require exactly one application and named failed assertions. The coherent
alternative Problem removes diagnostics; Type/schema control inserts the valid pair.
They establish oracle sensitivity, not substituted wire evidence.

## Binding and claims

Independently compared client GET target-binding membership: unchanged; first-use
order differs. Existing exact route digest includes that order, explaining both
updated pins without changing allowed routes or installed contracts. The selected
aggregate test checks the new digest. Unused-route behavior remains inherited.

Coverage has129rows:28 bounded whole,13 bounded incomplete,88 omitted;101uncovered.
The validation addition changes authored observations40→41 without reducing
uncovered rows. Packed progress inspected directly: exact a0b44bef, sourceDirty=false,
claimEligible=false, zero counterfactual applications,41 actual sessions,
650client/44setup/21native Read observations, and134 outcome records comprising
41pass and93 harness-error/missing bindings. No full-profile eligibility follows.

Author packed test JSON independently inspected:90passed/0failed/0pending at the
reported installed path. Payload0f914e6d688078dab562bc84a092fd143370abd4000f8ce5c6b0d0635c24ac26;
contracte1dff401b5caafb04a4d935775d67eeefeae7bcf306d5d100b979a64eef148e7.
These are inspected author artifacts, not a second native packed run or target
attestation. Public Read sealed evidence and runtime code were untouched.

## Independent verification and limits

Pinned Node24.16.0/pnpm11.20.0, frozen install, prebuild typecheck, build: PASS.
Focused actual HTTP run of `read-update-matrix.test.ts` selecting aggregate partial
matrix plus document-patch controls: **7passed,57filtered,13.44seconds**. Aggregate
runs all41 authored scenarios; six controls each fail their named semantic assertions
as required. Logs/JSON/exit0 in `bdp-integration-logs/document-patch-native/`.
No broad duplicate suite or new public qualification/reseal was run. This review
accepts the named bounded observations and honest incompleteness; it does not close
schema-equipped validation, message-cause semantics, concurrent visibility, complete
sequence ordering, or the remaining101 required rows.
