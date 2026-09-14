// Package authority provides the SQL-free installation identity primitive for
// future graph authority. It has no production caller and creates no Scope,
// witness, lease, or engine lock. Its permanent application sidecar serializes
// cooperating ID creators/readers, including durability repair on existing IDs.
//
// The ID is 64 lowercase hex characters and LF. Its key hashes those characters,
// a colon, and the exact canonical native storage-path bytes. Symlink aliases of
// one directory agree; moving the directory changes the key. Copying both the
// installation ID and the same canonical path remains an undetectable copy.
//
// Empty, partial, corrupt, replaced, foreign-owned, or permissive ID/lock files
// cause refusal and remain untouched. Do not casually delete or replace a used
// ID: later authority administration must account for the resulting identity
// change. Every successful call flushes the ID and its complete resolved parent
// chain through its same-device/root anchor, even after an interrupted earlier
// call. The anchor is flushed; the next differing-device ancestor is not. Device
// IDs need not distinguish bind mounts or Darwin firmlinks, so this can flush
// conservatively through /. No sync error inside that chain is ignored.
// Process termination tests do not establish power-loss guarantees.
//
// Go 1.26 crypto/rand.Read irrecoverably terminates on an entropy source failure;
// that is not a recoverable helper error. ID creation follows entropy collection,
// but parents and the sidecar can already exist. Injected entropy errors in tests
// exercise cleanup only. Caller cancellation/deadlines are not contention Busy;
// only the helper's internal two-second lock-wait cap reports Busy.
//
// The native implementation uses Unix no-follow opens and advisory flock. Actual
// filesystem support for file/directory Sync is required; network/FUSE and other
// Unix variants need their own qualification. Windows and wasm initially refuse
// before filesystem access. No universal hardware durability or malicious
// same-user ancestor-swap protection is claimed. The caller must select trusted
// per-user storage: untrusted users must not be able to replace the ID/sidecar
// entries or ancestors leading to them, including through ACL grants. Root-owned
// ancestors and sticky temporary directories can be valid; mode 0755 alone grants
// no other-user write access. Final ID/lock ownership and private modes are checked,
// but mode bits alone cannot establish this complete parent-trust precondition.
//
// The key uses CanonicalizeExistingPath's exact spelling. Strict resolution failure
// returns no key; it does not derive a key from a best-effort fallback. Configuration
// relocation can change the default ID location; following the existing resolver
// does not guarantee identity continuity across that relocation.
package authority
