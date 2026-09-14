//go:build !unix

package authority

import "context"

func installationKey(context.Context, string) (string, error) {
	return "", ErrInstallationUnsupported
}
