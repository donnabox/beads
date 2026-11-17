# BdpScope: the Go translation — what go-ifies, what's Rust-specific

Phase 0 of standing `bd` up as a BDP v0.2 client. This is the read that ships
alongside the hand-translated Go `Scope` contract (`scope.go`, `types*.go`,
`errors.go`): where the Go trait follows the spec, where it follows the Rust
clean-room's shape, and where it deliberately parts from both.

**Sources.** Authoritative: `core/core.lumen` @ `gastownhall/bdp spec/core-v0.2`
(the 13-primitive set, §3.3). Shape reference (read-only): the Rust `BdpScope`
trait in `bdp-cleanroom-protocol/crates/bdp-primitives/src/lib.rs` (§10, ~l.566)
and Clio's `RemoteScope` client in `bdp-cleanroom-cli/.../bdp-protocol/src/client.rs`.
Translated from the spec, not the Rust.

---

## The one thing worth saying first

**In the Lumen spec there is no `BdpScope` trait.** `BdpScope` is a *URL type*
— `type BdpScope = AbsoluteUri`, "the URL of a BDP scope (a store)". The
primitives are free functions annotated `/// bdp = { kind = "primitive" }`
(`fn get`, `command create`, …), and the only thing that gathers them is the
prose "a service advertises which it offers." The *trait* that bundles the
primitives into one in-language contract is the Rust clean-room's invention
(Donna-blessed: "ONE trait, `get` the only required method"). So "translate the
`BdpScope` trait" means: translate the **primitive set** into one Go interface,
using the Rust trait's *shape decisions* (get-mandatory, à la carte, a
discovery method) as the model. That is what `bdp.Scope` is.

The spec file's own header anticipates this: the toolchain is meant to project
`core.lumen` to "the JSON Schema bundle … plus the generated Markdown reference,
**Go interfaces**, and the canonical CLI." Our hand translation is the
human-vetted stand-in until that generator exists — scoped to Core, by hand,
no generator and no match-gate (both explicitly out of scope for now).

---

## What go-ifies cleanly

| Rust / spec | Go rendering | Why |
|---|---|---|
| `trait BdpScope: Send + Sync` | `type Scope interface` | The contract. A client and a backend both satisfy it. |
| `async fn` everywhere | `ctx context.Context` first param, synchronous return | Go has no `async`; the idiom is a context for cancellation + a blocking call. Matches every existing `internal/storage` signature. |
| default trait method → `not_implemented()` | embeddable `UnimplementedScope` returning `*ErrUnsupported` | Go has no default interface methods. The gRPC `UnimplementedServer` pattern is the standard analog — and it already matches beads' own "à la carte + typed unsupported" storage surface (`backend/`, `RunUnsupportedContract`). |
| `capabilities() -> Vec<PrimitiveCapabilities>` | `Capabilities(ctx) (Capabilities, error)` | Discovery is a mandatory member (no default), exactly as in Rust. Condensed the per-op limit *enum-of-structs* into one `PrimitiveCapability` with the knobs a Core caller branches on (Selector, IfMatch, JSONL/SSE, Async, MaxBatch). |
| `ref<T>` / `rep<T>` with `PhantomData<fn()->T>` | plain `ResourceID` / `Representation`, **no type param** | The spec says `T` is "checker-only and erases on the wire." Go phantom generics buy nothing at runtime and make every call site noisier; dropping `T` is faithful and idiomatic. |
| `RawJson(Bytes)` zero-copy | `json.RawMessage` | Core does not interpret `body`/`metadata`/events; keep them undecoded. |
| `enum BdpError { NotFound, RevisionMismatch, … }` | `error` return + sentinels (`ErrNotFound`, `ErrRevisionMismatch`, `ErrAlreadyExists`, `ErrValidation`) + `FaultError{Problem}` | `errors.Is`/`errors.As` is the Go classification idiom. `FaultError` carries the RFC 9457 `Problem` and `Unwrap`s to a sentinel, so one returned value answers both "which fault kind" and "what did the server say." |
| `Result<T,E>` / `Outcome<T,E>` (call's rich result) | `Reply{ Outcome{ Success/Failure/Fault/Canceled/Unknown } }` | Kept to `call` alone, per the spec's error-model ruling. The open non-domain set (fault·canceled) gets an explicit `Unknown json.RawMessage` arm so an evolving outcome set is tolerated, not dropped — the Go stand-in for "a reader MUST tolerate an unrecognized arm." |
| `enum Scope { This, Web, Named }` | `Namespace{ Store string; URI bool }` (zero value = local) | Renamed to free `Scope` for the interface. The common case (local) is the zero value, which reads well in Go. (Rust's `Web` = the spec's `"uri"` arm.) |

**`fn` vs `command`.** The Lumen surface distinguishes safe reads (`fn get`,
`fn list`, `fn pull`, `fn call`, `fn export`) from effectful writes (`command
create`, `idempotent command put/delete`, `command patch/push/cancel/import`).
Go has no such keyword, so the distinction lives in doc comments and in the
`Capabilities` advertisement, not the signature. Worth a generator annotation
later (it maps to the Operation effect markers `pure`/`safe`/`idempotent`).

---

## What is Rust/storage-specific and was deliberately left OUT of core `Scope`

The Rust trait is a *storage-facing* surface ("the scope's READ + STATE-MUTATION
interface"), so it carries members that are not Core primitives. Translating the
**spec**, I left these out and noted them here rather than smuggle them into the
conformance boundary:

- **`apply(&[Write])`** — one atomic write-set / transaction (log-append + edge
  re-derive + acyclicity + changefeed advance). Not a `core.lumen` primitive;
  it's the Dolt/graph backend's transactional convenience (`ApplyCaps{max_writes,
  acyclicity}`). Clio's client even routes it to a custom method `POST /:apply`.
  A real capability, but a *corelib/backend* one — out of Core.
- **`evaluate(Expr)`** — the query/eval engine (`bdp_ast`, `lumen_ast`, …). The
  spec explicitly parks "full filter / query / search / count / graph" as
  "corelib capabilities, advertised, not base." Out of Core.
- **`resolve(&[ResourceId])`** — batch `get` returning per-slot
  present/absent/erased. Ergonomic sugar over `get`; not a spec primitive. A
  caller loops `Get`, or a future `resolve` capability lands as an extension.
- **`resolve_alias(name)`** — the `aliases` capability. The spec parks aliases
  entirely ("aliases & Transfer" under open/parked). Out of Core.
- **`changes(...)` as a distinct method** — the Rust trait splits the changefeed
  out from general `pull`. The spec models the changefeed as the well-known
  channel `/sys/changes` read by the one `pull` primitive ("namespace /
  well-known-name routing lives ABOVE this layer," says the Rust module doc
  itself). So Go has one `Pull`; the changefeed is just the canonical channel.

## What the Rust trait OMITS that the spec keeps (and Go keeps)

The Rust trait is not a 1:1 of the spec's 13 primitives — it drops three that
`bd`'s Go contract restores, because they are Core:

- **`revisions`** — one resource's history (spec §3.1, full tier). Absent from
  the Rust trait; present in Go `Scope.Revisions`.
- **`push`** — the channel write primitive, dual of `pull` (spec §2.2). Absent
  in Rust (a reader-first trait); present in Go `Scope.Push`.
- **`cancel`** — stop a running async `call` (spec §2.3). Absent in Rust;
  present in Go `Scope.Cancel`.

Net: Go `Scope` = **exactly the 13 spec primitives + `Capabilities`**, no Rust
extensions. That is the deliberate call — the conformance boundary tracks the
spec, and backend/corelib surface attaches as advertised capabilities later.

---

## Beads-norm fit (not just Rust translation)

- **`context.Context` first, `(T, error)` return** — every method, matching
  `internal/storage`.
- **Typed `*beadserrors.ErrUnsupported{Op, Backend}`** — the repo's existing
  "this backend doesn't do that" idiom (`backend/errors.go`,
  `RunUnsupportedContract`). The à-la-carte shape is *already* how beads models
  its Dolt storage contract (core `Storage` + capability sub-interfaces answered
  by typed unsupported). BdpScope lands in a surface the codebase already speaks.
- **Domain-neutral sentinels, no issue coupling.** `beadserrors` has neutral
  `ErrNotFound`/`ErrValidation`; it has **no** neutral revision-mismatch or
  already-exists (issue-flavored twins live in `issueops`). BDP defines its own
  protocol-level `ErrRevisionMismatch`/`ErrAlreadyExists` rather than couple the
  firewall boundary to the issue domain. (A candidate to promote into
  `beadserrors` if a second neutral consumer ever wants them.)
- **Heavy package/type doc comments** — matches `backend/`, `beadserrors/`.
- **`internal/bdp`** — internal until the boundary settles; the public extension
  door is `backend/`, and this isn't that yet.

---

## Spec gaps / questions flagged for Donna (before v0.2 closes)

These are things the mechanical translation bumped into. None blocked the trait;
all are worth a ruling.

1. **Is there meant to be an in-language aggregation of the primitives?** The
   spec defines `BdpScope` as a *URL* and the primitives as loose `fn`/`command`
   forms; the "one contract a client and backend satisfy" lives only in prose +
   the Rust trait. If the generated **Go interfaces** the header promises are to
   be deterministic, the spec needs to say *what gathers the primitives* (a
   `service`/`trait` form? an annotation the generator keys off?). Right now a
   generator has to invent the grouping, exactly as I did by hand.

2. **`>DB?` markers still open in `core.lumen`** that touch primitive signatures:
   - l.834 — `create`: may it admit a *content* resource (media-type `type` +
     registry authority), or is content `put`-only? Changes whether `CreateInput`
     should carry content at all.
   - l.884 — `delete` returning a bare `Rev` (tombstone) instead of a pinned
     `ref`, breaking the "writes return a pinned ref" symmetry on purpose.
     Confirmed-looking, but still a `>DB?`. Go followed it (`Delete → (Rev,
     error)`).
   - l.693 — `Outcome` left as a declared-open union (no explicit `other{kind}`
     arm). Go modeled the openness with an `Unknown` arm; a ruling here fixes the
     wire encoding.

3. **`fn` vs `command` has no wire/trait projection.** The safe/effectful split
   is real in the Lumen surface and in the Operation effect markers, but a
   generated Go interface erases it (both become `(ctx, …) (T, error)`). If that
   split should survive into generated clients (e.g. to gate retries on `safe`),
   the spec should say how it projects. Noted, not blocking.

4. **Discovery shape is Shelly's draft, not yet spec.** The module doc says the
   per-op limit records (`GetCaps`, `ListCaps`, …) are "Shelly's draft, Lydia
   BDP-ifies them into the spec." So `Capabilities`/`PrimitiveCapability` here is
   a best-effort condensation; it will need re-alignment once discovery is
   actually in `core.lumen`.
