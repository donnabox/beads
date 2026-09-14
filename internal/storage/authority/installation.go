package authority

import (
	"context"
	"errors"
)

var (
	// ErrInvalidIdentity reports invalid persistent bytes, kind, or permissions.
	ErrInvalidIdentity = errors.New("invalid installation identity")
	// ErrInstallationBusy reports contention beyond the bounded lock wait.
	ErrInstallationBusy = errors.New("installation identity busy")
	// ErrInstallationUnsupported reports unavailable platform or durability support.
	ErrInstallationUnsupported = errors.New("installation identity unsupported")
)

// InstallationKey binds one per-user installation ID to an existing canonical
// storage directory. A nil context is invalid. The key grants no authority.
// Cancellation and all persistence/cleanup failures return no key; a retry never
// replaces an existing identity. See the package documentation for limitations.
func InstallationKey(ctx context.Context, beadsDir string) (string, error) {
	if ctx == nil {
		return "", errors.New("installation identity requires a context")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return installationKey(ctx, beadsDir)
}
