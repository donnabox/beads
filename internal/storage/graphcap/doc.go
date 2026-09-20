// Package graphcap contains private process-local graph claim lifetime mechanics.
// It has no production issuer, SQL, provider wiring or public Reader. Positive
// activation is package-private and currently used only by synthetic tests.
// These mechanics establish neither authority nor any E1–E8 engineering gate.
//
// A LeaseClaim represents one admission, never a reusable grant. Check requires
// the exact context returned by Begin; even WithCancel/WithTimeout descendants
// refuse. Context composition for future transaction bodies needs separate
// review. Routine extension preserves existing admissions and their original
// deadlines. Explicit revocation cancels; the production policy for replacing
// an active era on authority-changing re-arm remains unresolved and absent.
//
// Owner identity binds workspace, installation, database and default branch.
// Era identity binds Scope, authority, epoch and grant. Physical-source loss,
// qualified issuance, passive per-call witness loading, and DB clock/suspend
// behavior remain prerequisites to production wiring. Clock anchors are captured
// inside this package; no caller-supplied timestamp or evidence grants a claim.
//
// graphsession and graphmanaged own different physical-connection/process
// lifetimes and are not dependencies. Future transaction bodies may depend on
// graphcap, never the reverse; SQL observation extraction and the corresponding
// B6/lint alignment require their own reviewed change.
package graphcap
