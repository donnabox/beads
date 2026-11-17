// Package bdp is bd's hand-translated Go rendering of the Beads Protocol (BDP)
// Core contract — the v0.2 conformance boundary a client and every storage
// backend both satisfy.
//
// # What this is
//
// BDP is a refinement of HTTP: a small, fixed set of operations over versioned
// resources, the same for every resource whatever its type. The slogan is
// "BDP = Beads × HTTP". The authoritative definition is the Lumen source
// core/core.lumen on gastownhall/bdp @ spec/core-v0.2; the toolchain is meant
// to generate the Go interfaces from it. This package is the first,
// human-vetted hand translation of that contract, standing in until the
// generator exists, so bd can prove it speaks v0.2 before any server work.
//
// It is a MECHANICAL translation of the spec's primitive set, not of any
// particular implementation. Where a storage backend (Dolt, Postgres, …) or
// the corelib adds surface beyond Core — atomic write-sets, a query/evaluate
// engine, batch resolve, aliases — that surface is deliberately NOT here; see
// SPELUNK.md in this package for what go-ifies and what was left out and why.
//
// # The Scope contract
//
// A Scope is one opened BDP store. [Scope] gathers the Core primitives into one
// Go interface — the shape Donna fixed as "à la carte, NOT profiles: ONE trait,
// get plus discovery the required floor, every other primitive advertised, not
// assumed" (core.lumen §3.3, gastownhall/bdp#398).
//
// Go has no default interface methods, so the à-la-carte shape is expressed two
// ways that agree:
//
//   - [Scope.Capabilities] advertises, at runtime, exactly which primitives a
//     scope offers and the limits on each — the discovery the spec requires so a
//     caller never guesses. On the wire it is itself a get, of the well-known
//     /sys/capabilities resource, not a distinct primitive.
//   - A primitive a scope does not offer returns a *[beadserrors.ErrUnsupported]
//     naming the operation (the same typed refusal the rest of bd's storage
//     surface uses), so a caller that skips discovery still classifies the
//     refusal the same way everywhere.
//
// [UnimplementedScope] supplies that refusing default for every OPTIONAL
// primitive. Embed it and you inherit a conformant minimal scope — except Get
// and Capabilities, which it deliberately does not provide, so the compiler
// makes every implementer write the two mandatory members itself.
//
// # Error model
//
// A data primitive returns its success value directly and faults with an error;
// the rich Result/Outcome split (declared success vs declared failure vs fault)
// belongs to Call alone. Faults are classified through the shared bd error
// vocabulary ([beadserrors]) plus the BDP-specific [Problem] a fault carries;
// see errors.go.
//
// # Stability
//
// EXPERIMENTAL and pre-release. This is Phase 0 of standing bd up as a BDP v0.2
// client; the contract tracks an unreleased spec (spec/core-v0.2) and will move
// with it. Nothing here is covered by any compatibility promise yet, and the
// package is internal on purpose until the boundary settles.
package bdp
