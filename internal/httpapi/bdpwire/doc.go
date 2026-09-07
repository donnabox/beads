// Package bdpwire holds the Go wire types for the BDP v0 Read profile — the
// JSON envelopes `bd --graph-mode link serve` will answer with (P2) and a BDP
// client will decode — pinned to one upstream commit of the protocol's
// normative schema bundle.
//
// THE PIN. The bundle is vendored verbatim at schema/bdp-v0.schema.json from
// gastownhall/bdp commit 0b7d86e7 (the plan's §0 pin), with the Read-profile
// fixtures and the executable Read matrix from the same commit beside it.
// schema/PIN names every vendored file with its sha256 and its upstream git
// blob sha1, and pin_test.go recomputes both from the bytes on disk, so a
// re-pin is an edit to PIN that review can see and a drifted fixture is a
// failing test — never a network fetch. Nothing in this package touches the
// network, at build time or in tests.
//
// HAND-WRITTEN, VALIDATED AGAINST THE BUNDLE. The types here are written by
// hand and welded to the bundle by schema_parity_test.go, which parses the
// vendored bytes and asserts, for every `$defs` entry: two-way equality
// between the schema's property list and the struct's JSON tags; that a
// required member never carries omitempty and an optional one always does;
// that a closed envelope carries no extension carrier and the one open
// envelope does; that every enum vocabulary matches constant for constant;
// and that the problem code → family/status/retry table equals the bundle's
// conditional rows. GENERATOR.md records why no generator was used (spec
// Part D.3), with the evidence, and what would change that. The parity tests
// run under `go test ./internal/httpapi/...`, which is the second half of
// `make api-check`, so the existing OpenAPI drift gate covers this contract
// too and no Makefile change was needed.
//
// SHAPE, NOT VALIDATION. A type here says what a conforming document looks
// like on the wire: which members exist, which are required, which are
// closed. DECISION: it does not validate URL grammar, canonical-ID spelling,
// Type contracts, or that a required member was present — those are model
// laws and belong to the graph leaf (graphops laws.go, per
// BDP_GRAPH_ARCHITECTURE.md §3), not to transport. The exceptions are the
// structural facts the bundle states about the bytes themselves and that a Go
// value could not otherwise carry faithfully: Reference's string-or-object
// sum (an object arm needs a nonempty revision, or the Go value would read as
// the string arm), and the closed problem table (ReadProblem.Validate).
//
// DECODING POSTURE. DECISION: Unmarshal and Decode are strict — an unknown
// member in a closed envelope is an error — because the spec says the
// normative schemas alone decide where additional members are allowed ("no
// implicit minor-version rule") and every public envelope closes its
// protocol-owned members; the spec does not say what a CLIENT should do with
// a stranger, and rejecting is the reading a conformance check wants. The two
// open places — a Resource's `properties` document and RFC 9457 extension
// members on a problem — are carried through byte for byte. A caller that
// wants a lenient read uses encoding/json directly.
//
// The package imports the standard library and nothing else, and
// imports_test.go keeps it that way: it sits beneath internal/httpapi and,
// later, a client package, and must pull neither's dependencies into the
// other. It must never import graphops or any internal/storage package.
package bdpwire
