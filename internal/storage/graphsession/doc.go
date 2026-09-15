// Package graphsession contains the private, unwired network-session exclusion
// candidate. It has no exported constructor, controller, provider caller or
// authority capability. In particular, an endpoint value is not deployment
// evidence. Only tests currently construct one.
//
// Future integration requires managed process/binary/effective-configuration and
// database-registration provenance before handshake (which can otherwise cause
// replica effects), the qualified engine drain/lock lifecycle, all participating
// routes and renewers, witness/evidence and lease/workload qualification. Client
// result completion and physical transport disposal prove neither server
// terminality nor permission to release another participant's exclusion.
//
// Methods reject overlapping use; the caller owns its call goroutine and must
// join it before closing. Result handling owns a context.AfterFunc callback
// that closes the retained transport on cancellation/deadline and is stopped
// or explicitly joined before return, including driver-initiated Rows.Close.
// Local refusals close the transport before attempting Rows.Close; unread tail
// errors are not claimed as observed. Driver I/O has contexts
// and a 250 ms write timeout, including contextless COM_QUIT. Cleanup owns the
// actual transport and always closes the one-operation pool; no pooled reuse,
// SQL session repair, reconnect, caller-injected SQL/callback or engine recovery
// exists. The pinned MySQL driver still supports process-global local-infile
// file/reader registrations even with AllowAllFiles=false. It has no per-config
// disable option: a 0xfb reply can reach those handlers. This is an additional
// driver seam to close before production admission, not a closed capability
// proved by this package. Do not add a Beads packet filter or global-registry
// reset to disguise it.
package graphsession
