package dolt

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	storeops "github.com/steveyegge/beads/internal/storage/issueops"
)

func TestLegacyEpochMethodsNeverReplayIndeterminateSQLCommits(t *testing.T) {
	for _, name := range []string{"bump", "mint", "remint"} {
		t.Run(name, func(t *testing.T) {
			f := newEpochFake()
			f.connector.commitErr = func() error { return testConnectionLoss }
			db := sql.OpenDB(f.connector)
			defer db.Close()
			store := &DoltStore{db: db}
			var err error
			switch name {
			case "bump":
				var got int
				got, err = store.BumpEpoch(context.Background(), epochPublishStoreID, "restore")
				if got != 2 {
					t.Fatalf("possibly applied epoch=%d,want2", got)
				}
			case "mint":
				var got string
				got, err = store.MintUnderEpoch(context.Background(), epochPublishStoreID, epochPublishMintedID)
				if got == "" {
					t.Fatal("possibly applied address discarded")
				}
			case "remint":
				var got string
				got, err = store.CurrentAddressFor(context.Background(), epochPublishStoreID, epochPublishAddress())
				if got == "" {
					t.Fatal("possibly applied remint discarded")
				}
			}
			if !errors.Is(err, ErrCommitIndeterminate) || !errors.Is(err, testConnectionLoss) {
				t.Fatalf("error=%v,want indeterminate connection loss", err)
			}
		})
	}
}

func TestLegacyEpochDeferredMethodsNeedNoConnection(t *testing.T) {
	store := &DoltStore{}
	if _, err := store.BumpEpochCarrying(context.Background(), "store", "scheme"); !errors.Is(err, storeops.ErrEpochSemanticsDeferred) {
		t.Fatal(err)
	}
	if err := store.LoseVersion(context.Background(), "store", "address"); !errors.Is(err, storeops.ErrEpochSemanticsDeferred) {
		t.Fatal(err)
	}
}
