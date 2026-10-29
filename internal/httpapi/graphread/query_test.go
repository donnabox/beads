package graphread

import (
	"context"
	"errors"
	"testing"

	"github.com/steveyegge/beads/internal/resourcequery"
)

func TestQueryAdmissionPrecedesStoreAcquisition(t *testing.T) {
	var reader *Reader
	for _, raw := range []string{
		`{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"unknown"},"result":{"form":"resources"}}`,
		`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"count"},"Profile":"query"}`,
	} {
		_, err := reader.ExecuteResourceQuery(context.Background(), []byte(raw))
		if !errors.Is(err, resourcequery.ErrInvalid) {
			t.Fatalf("expected request refusal before consulting nil storage: %v", err)
		}
	}
}
