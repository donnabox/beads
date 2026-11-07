package main

import "fmt"

// This describes the existing local mutation executors. It does not advertise
// public BDP Write, select a workspace, or change a backend's capabilities.
type canonicalMutationLeafDescriptor struct {
	Operation string
	Path []string
	Flags []string
	MinimumPositionals int
	MaximumPositionals int
	Create bool
	Link bool
	Delete bool
	Batch bool
}

// The profile router consumes these fresh descriptors using its existing
// prohibited-Value guard. No flag setter, command clone or argv parser belongs
// here. Unsupported flags must be rejected even when explicitly set to false.
func lookupCanonicalMutationLeaf(path []string) (canonicalMutationLeafDescriptor, bool) {
	d := canonicalMutationLeafDescriptor{}
	switch {
	case len(path) == 1 && path[0] == "create":
		d.Operation, d.Create = "createBead", true
	case len(path) == 1 && path[0] == "update":
		d.Operation = "updateBead"
	case len(path) == 1 && path[0] == "delete":
		d.Operation, d.Delete = "deleteBead", true
	case len(path) == 2 && path[0] == "link" && path[1] == "create":
		d.Operation, d.Create, d.Link = "createLink", true, true
	case len(path) == 2 && path[0] == "link" && path[1] == "update":
		d.Operation, d.Link = "updateLink", true
	case len(path) == 2 && path[0] == "link" && path[1] == "delete":
		d.Operation, d.Delete, d.Link = "deleteLink", true, true
	case len(path) == 1 && path[0] == "batch":
		d.Operation, d.Batch = "batch", true
	default:
		return canonicalMutationLeafDescriptor{}, false
	}
	d.Path = append([]string(nil), path...)
	d.Flags = []string{"directory", "graph-mode", "readonly", "quiet", "no-color", "actor", "json", "scope", "dry-run-mode"}
	if d.Batch {
		d.Flags = append(d.Flags, "file", "input-format", "batch-version", "batch-arg", "idempotency-key", "retry-file", "output-format", "include", "omit")
		return d, true
	}
	if d.Create {
		d.Flags = append(d.Flags, "id", "id-root", "batch-bind")
		if d.Link {
			d.Flags = append(d.Flags, "from", "to")
			d.MaximumPositionals = 2
		}
	} else {
		d.MinimumPositionals, d.MaximumPositionals = 1, 1
		d.Flags = append(d.Flags, "if-revision")
	}
	if d.Delete {
		d.Flags = append(d.Flags, "force")
	} else {
		d.Flags = append(d.Flags, "type", "properties", "metadata", "revision-provenance")
		if !d.Create {
			d.Flags = append(d.Flags, "replace-properties", "patch-properties", "set-property", "unset-property", "replace-metadata", "patch-metadata", "set-metadata", "unset-metadata")
		}
	}
	return d, true
}

type canonicalMutationLeafRefusal struct {
	Code string
	ExitCode int
	Message string
}

func (e *canonicalMutationLeafRefusal) Error() string { return e.Code + ": " + e.Message }

func canonicalMutationLeafInvalid(message string) error {
	return &canonicalMutationLeafRefusal{Code: "invalid_selector", ExitCode: 2, Message: message}
}

func canonicalMutationLeafUnavailable(message string) error {
	return &canonicalMutationLeafRefusal{Code: "capability_unavailable", ExitCode: 5, Message: message}
}

func (d canonicalMutationLeafDescriptor) admitPositionals(count int) error {
	if count < d.MinimumPositionals || count > d.MaximumPositionals {
		return canonicalMutationLeafInvalid(fmt.Sprintf("%s requires %d to %d positional References", d.Operation, d.MinimumPositionals, d.MaximumPositionals))
	}
	return nil
}
