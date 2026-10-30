package issueops

import (
	"context"
	"errors"
	"testing"
)

func TestReleasedEpochAPIsDeferSurvivalBeforeSQL(t *testing.T) {
	// A nil runner would panic if either operation attempted any SQL.
	if got, err := BumpEpochCarryingInTx(context.Background(), nil, "store", "scheme"); got != 0 || !errors.Is(err, ErrEpochSemanticsDeferred) {
		t.Fatalf("carrying=(%d,%v)", got, err)
	}
	if err := LoseVersionInTx(context.Background(), nil, "store", "address"); !errors.Is(err, ErrEpochSemanticsDeferred) {
		t.Fatalf("loss=%v", err)
	}
}
