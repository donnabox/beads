// Package graphmanaged contains private, SQL-free input inspection and owned
// process machinery for a future managed graph adapter. It is intentionally
// unwired: no approved production adapter, exported endpoint, startup API or
// graph authority is provided here. Inspection is not source qualification.
//
// A future caller must first qualify the adapter's complete initialization,
// security and mutation closure, non-forking lifetime, exact artifact, and the
// host's exclusive child-wait/no-auto-reap behavior. Process fixtures do not
// establish these facts. The existing launcher and storage semantics are not
// used or changed. No engine, SQL repair, process adoption or restart occurs.
//
// The process owner completes every signal before its sole Wait starts. This
// ordering also avoids the Signal/Wait PID-reuse race documented by Go's Darwin
// fallback. Report loss invalidates independently of log EOF and reaping. Only
// a prepared/activate/activated exchange completes fixture startup. Neither an
// in-memory sequence nor successful startup is a graph capability.
//
// Inputs require stable, available local files and trusted mutable leaves and
// parents. Privileged replacement and blocked filesystem syscalls are outside
// the cooperative deadline guarantee. Engine-owned files are hashed as opaque
// bytes, never parsed as storage state. Resource limits are administrative
// defaults, independent of BDP limits. Windows and other unsupported platforms
// refuse before inspecting files or starting children.
package graphmanaged
