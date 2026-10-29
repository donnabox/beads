# Intrinsic materialization candidates

This package assembles an immutable **uncommitted plan** from the checked fixed
`intrinsictypecatalog` publication and six candidate native allocations. It does
not allocate, authorize, install, persist, register, select defaults or implement
ordinary show/list/create. Neither assembly nor private decoding proves that a
candidate exists or belongs to the asserted authority.

`Build` requires a nonzero checked catalog, an explicit Scope/native authority/
opaque context binding, and exactly one allocation for each of four definition
and two declaration roles. Public Scope-local IDs and rooted private locators
are separate explicit inputs. Public identity is composed with graphops' literal
canonical identity helper; kind follows the role, never the spelling of the ID.
The native authority shape is 32 lowercase hexadecimal characters. Revisions
are nonempty opaque UTF-8 strings; different Resources may share a token.
Reservation against existing Resources and authority authorization remain the
future controlled initializer's responsibility. No storage layout is selected.

Four definitions, including Refines, become candidate **Beads**. Their original
shared exact own-Type affiliations, properties, schema resources and selectors
remain unchanged. Two candidate **Links** retain global exact Refines affiliation
and declaration properties; their source is the actual floating local owner and
their target is the exact local Type edition in this context. The complete Links
are captured in final BeadType/LinkType owner bodies under the global nominal
Refines bucket; Type/Refines retain empty owned state. These mirror graph facts
do not replace the shared publication's semantic Refines relationships.

Each binding separately retains its complete original global record and local
body. `Bindings` and `LookupLocal` return independent copies. `LocalFor` requires
an exact shared pair and returns only its candidate within this frozen context;
it provides no general Type alias, current/default lookup or cross-context
fallback. New contexts leave older plans unchanged. Lookup is not ordinary
Resource ownership. Donna selected ordinary local definition Beads; this package
only prepares their candidate bodies and does not implement that producer.

`EncodeContext`/`DecodeContext` describe private planning data, **not** a public
mapping carrier or adopted persistence format. A one-MiB predecode bound and
aggregate candidate/encoded-output budgets bound this fixed six-member object.
Strict JSON custody rejects duplicates, ambiguous strings, unknown/case-confused
or missing fields. Decoding rebuilds the expected full records from the checked
R1 catalog and compares the complete canonical context. A checksum detects
corruption only: recomputing it cannot authenticate altered source/record facts.
A coherent different candidate context can decode successfully and still has no
allocation, commit or authority proof.

Artifact provenance remains the R1 ba8d publication, SHA256
`73b16c93b6963f9657be3569ba9d16f722edbf6f49090803e748c907c39aab7e`.
The separate semantic-source reference is qualified BDP main
`c6dfb3c05fe707cf605030c5661123e8dda0d8f8`. This package does not rewrite
artifact provenance to the later prose commit. The private context encoding label
is not a database layout version or public protocol version.

Imports are standard library, graphops and the fixed catalog only. No package
initializer invokes the loader, compiler, filesystem, network or process APIs.
The caller explicitly obtains the checked catalog. The proposed finite tests
exercise the production assembler/codec, captures, unchanged properties and
pins, refusal paths and independent old/new contexts. Source preparation alone
is not test evidence. Canonical Bazel/nogo, supported Linux and any later
persistence/provider/CLI qualification remain separate gates.
