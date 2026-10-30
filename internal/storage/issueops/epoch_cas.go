package issueops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/steveyegge/beads/internal/storage/dberrors"
)

// Released fork epoch semantics remain the production default.
// The separate Survival functions retain the incoming complete-format contract;
// no schema inspection or migration automatically selects that contract.
func epochAddress(storeID, id string, epoch int) string {
	return fmt.Sprintf("epch:%d:%s:%d:%s:%d", len(storeID), storeID, len(id), id, epoch)
}

// validateEpochAddressInputs rejects a storeID or id containing ":" before
// MintUnderEpochInTx mints an address from them (item B2b above).
func validateEpochAddressInputs(storeID, id string) error {
	if strings.Contains(storeID, ":") {
		return fmt.Errorf("%w: storeID %q", ErrEpochAddressSeparator, storeID)
	}
	if strings.Contains(id, ":") {
		return fmt.Errorf("%w: id %q", ErrEpochAddressSeparator, id)
	}
	return nil
}

// ensureStoreEpochRow lazily initializes store_epoch's singleton row
// (migration 0067; the table starts empty) to epoch 1 on first use, and
// reports the current epoch either way. Only BumpEpochInTx and
// MintUnderEpochInTx call this: both are about to write regardless, so
// creating the row here is not a surprising extra side effect. Every
// read-only path calls readLegacyStoreEpochInTx instead (gastownhall/beads#6664,
// bee-ghosttrack review 5268699223, item B4): a write from what callers
// reasonably expect to be a read is surprising at best, and a hard failure
// against a genuinely read-only connection at worst.
func ensureStoreEpochRow(ctx context.Context, tx DBTX) (int, error) {
	var epoch int
	err := tx.QueryRowContext(ctx, `SELECT epoch FROM store_epoch WHERE id = 1`).Scan(&epoch)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO store_epoch (id, epoch) VALUES (1, 1)`); err != nil {
			return 0, fmt.Errorf("initialize store_epoch: %w", err)
		}
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read store_epoch: %w", err)
	}
	return epoch, nil
}

// readLegacyStoreEpochInTx reports the current epoch without writing:
// store_epoch's singleton row starting absent (migration 0067) and its
// epoch being 1 are the same state, so a read-only caller can answer from
// that default instead of initializing the row the way ensureStoreEpochRow
// does (item B4 above). CurrentEpochInTx, StillServesInTx and
// ResolveEpochInTx all read this way; only BumpEpochInTx and
// MintUnderEpochInTx, which are about to write regardless, use
// ensureStoreEpochRow.
func readLegacyStoreEpochInTx(ctx context.Context, tx DBTX) (int, error) {
	var epoch int
	err := tx.QueryRowContext(ctx, `SELECT epoch FROM store_epoch WHERE id = 1`).Scan(&epoch)
	if errors.Is(err, sql.ErrNoRows) {
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read store_epoch: %w", err)
	}
	return epoch, nil
}

// legacyEpochMintedAddress is one row as read from epoch_minted_addresses.
type legacyEpochMintedAddress struct {
	storeID     string
	mintedID    string
	mintedEpoch int
}

// readEpochMintedAddressInTx reads address's row, if any. found is false
// (with a zero-value row and nil error) when address was never minted —
// distinguishing "not found" from an actual query error is the caller's
// job, the same way readExpectedRevisionRowInTx's callers do it.
func readEpochMintedAddressInTx(ctx context.Context, tx DBTX, address string) (legacyEpochMintedAddress, bool, error) {
	var row legacyEpochMintedAddress
	err := tx.QueryRowContext(ctx,
		`SELECT store_id, minted_id, minted_epoch FROM epoch_minted_addresses WHERE address = ?`, address,
	).Scan(&row.storeID, &row.mintedID, &row.mintedEpoch)
	if errors.Is(err, sql.ErrNoRows) {
		return legacyEpochMintedAddress{}, false, nil
	}
	if err != nil {
		return legacyEpochMintedAddress{}, false, fmt.Errorf("epoch CAS: read minted address %s: %w", address, err)
	}
	return row, true, nil
}

// upsertEpochMintedAddressInTx writes address's row. exists (from the
// caller's own prior read) says whether one already does: if so, this is a
// pure no-op (gastownhall/beads#6664, bee-ghosttrack review 5268699223,
// item B2c) — epochAddress is now injective (item B2a), so an existing row
// at this exact address provably already carries this exact (storeID,
// mintedID, mintedEpoch) triple, and the old exists-branch UPDATE (which
// had no store_id/minted_id guard) could only ever have mattered by
// rewriting a DIFFERENT triple's row out from under it on an address
// collision the new encoding no longer allows. The not-exists branch treats
// a duplicate-key error from the INSERT as a benign race rather than a hard
// failure (item B3): two concurrent minters computing the same address can
// both attempt the same insert, and since it would insert the identical row
// the loser would otherwise have upserted anyway, treating its
// duplicate-key error as success is correct, not merely convenient.
func upsertEpochMintedAddressInTx(ctx context.Context, tx DBTX, address, storeID, mintedID string, mintedEpoch int, exists bool) error {
	if exists {
		return nil
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO epoch_minted_addresses (address, store_id, minted_id, minted_epoch, minted_at) VALUES (?, ?, ?, ?, ?)`,
		address, storeID, mintedID, mintedEpoch, time.Now().UTC(),
	); err != nil {
		if dberrors.IsDuplicateKey(err) {
			return nil
		}
		return fmt.Errorf("epoch CAS: insert minted address %s: %w", address, err)
	}
	return nil
}

// CurrentEpochInTx reports storeID's current epoch generation. store_epoch
// has no store_id column: migration 0067 established it as one physical
// counter per database, matching production reality (one logical store per
// database already) — storeID is accepted here only to satisfy
// EpochFixture's hook signature.
func releasedCurrentEpochInTx(ctx context.Context, tx DBTX, storeID string) (int, error) {
	epoch, err := readLegacyStoreEpochInTx(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf("epoch CAS: current epoch for %s: %w", storeID, err)
	}
	return epoch, nil
}

// BumpEpochInTx advances storeID's epoch generation by one and records why
// (R20-a: a bump is triggered only by restore, destructive-reinit, or
// token-scheme-change — the leg adapter converts the fixture's
// conformance.EpochBumpTrigger to this plain string via trigger.String(),
// keeping that vocabulary out of this file per the package doc above).
func releasedBumpEpochInTx(ctx context.Context, tx DBTX, storeID, reason string) (int, error) {
	if _, err := ensureStoreEpochRow(ctx, tx); err != nil {
		return 0, fmt.Errorf("epoch CAS: bump epoch for %s: %w", storeID, err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE store_epoch SET epoch = epoch + 1, bumped_at = ?, bumped_reason = ? WHERE id = 1`,
		time.Now().UTC(), reason,
	); err != nil {
		return 0, fmt.Errorf("epoch CAS: bump epoch for %s: %w", storeID, err)
	}
	var newEpoch int
	if err := tx.QueryRowContext(ctx, `SELECT epoch FROM store_epoch WHERE id = 1`).Scan(&newEpoch); err != nil {
		return 0, fmt.Errorf("epoch CAS: read epoch after bump for %s: %w", storeID, err)
	}
	return newEpoch, nil
}

// MintUnderEpochInTx mints id's address under storeID's CURRENT epoch,
// deterministically (epochAddress): minting the same id again under the
// same epoch reproduces the same address and is an idempotent no-op upsert
// of the same row.
func releasedMintUnderEpochInTx(ctx context.Context, tx DBTX, storeID, id string) (string, error) {
	if err := validateEpochAddressInputs(storeID, id); err != nil {
		return "", fmt.Errorf("epoch CAS: mint %s under epoch for %s: %w", id, storeID, err)
	}
	epoch, err := ensureStoreEpochRow(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("epoch CAS: mint %s under epoch for %s: %w", id, storeID, err)
	}
	address := epochAddress(storeID, id, epoch)
	_, found, err := readEpochMintedAddressInTx(ctx, tx, address)
	if err != nil {
		return "", err
	}
	if err := upsertEpochMintedAddressInTx(ctx, tx, address, storeID, id, epoch, found); err != nil {
		return "", err
	}
	return address, nil
}

// mintedIDHasAddressAtEpochInTx reports whether mintedID's mapping was
// CARRIED FORWARD from priorEpoch through to epoch (storeID's current
// epoch) — R20-n's retained-mapping exception: a version that survives an
// epoch transition keeps its prior-epoch address resolving once
// CurrentAddressForInTx (or a fresh MintUnderEpochInTx) has carried its id
// forward into the current epoch. A bare "does mintedID have ANY row at
// epoch" existence check is not that (gastownhall/beads#6664,
// bee-ghosttrack review 5268699223, item B1): a mapping that lapses at an
// intermediate epoch and is only later reused — an unrelated later mint
// that happens to share the same id — would satisfy it too, wrongly
// retaining an already-superseded address across the gap. Nor is a
// single-hop check enough (gastownhall/beads#6664 FINDING 1): a mapping
// carried forward through every intervening epoch across MORE than one hop
// is just as validly retained, and epoch == priorEpoch + 1 wrongly rejects
// it. Requiring an exact COUNT(DISTINCT minted_epoch) match against the
// full range width (priorEpoch, epoch] pins this to an UNBROKEN CHAIN
// across every intervening epoch: since epochAddress is deterministic per
// (storeID, id, epoch) and upsertEpochMintedAddressInTx's exists-branch is
// a no-op, there is at most one row per (storeID, mintedID, minted_epoch),
// so a row exists at every one of the range's (epoch-priorEpoch) possible
// values iff the distinct count equals that width — an exact pigeonhole
// match, not a bound, so it still correctly rejects a gap (no epoch in the
// range is missing from the count, no matter how many epochs follow it)
// while now accepting any chain length >= 1, subsuming the old width=1
// single-hop case. mintedID is never recomputed from a bumped store_epoch
// (see the package doc above), so this is a lookup across a range of
// later rows sharing the same id, not a derivation.
func mintedIDHasAddressAtEpochInTx(ctx context.Context, tx DBTX, storeID, mintedID string, priorEpoch, epoch int) (bool, error) {
	var count int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT minted_epoch) FROM epoch_minted_addresses WHERE store_id = ? AND minted_id = ? AND minted_epoch > ? AND minted_epoch <= ?`,
		storeID, mintedID, priorEpoch, epoch,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("epoch CAS: check retained mapping for %s across epochs (%d, %d]: %w", mintedID, priorEpoch, epoch, err)
	}
	return count == epoch-priorEpoch, nil
}

// StillServesInTx reports whether address is still served under storeID's
// CURRENT epoch: an address minted under an earlier epoch is no longer
// served once the epoch has moved past it, UNLESS its underlying id was
// carried forward into the current epoch by a retained mapping (R20-n),
// even though the row itself is never deleted (ResolveEpochInTx must still
// be able to answer for it). An address from a different store, or one
// never minted, is not served either.
func releasedStillServesInTx(ctx context.Context, tx DBTX, storeID, address string) (bool, error) {
	row, found, err := readEpochMintedAddressInTx(ctx, tx, address)
	if err != nil {
		return false, err
	}
	if !found || row.storeID != storeID {
		return false, nil
	}
	epoch, err := readLegacyStoreEpochInTx(ctx, tx)
	if err != nil {
		return false, fmt.Errorf("epoch CAS: still serves %s for %s: %w", address, storeID, err)
	}
	if row.mintedEpoch == epoch {
		return true, nil
	}
	retained, err := mintedIDHasAddressAtEpochInTx(ctx, tx, storeID, row.mintedID, row.mintedEpoch, epoch)
	if err != nil {
		return false, fmt.Errorf("epoch CAS: still serves %s for %s: %w", address, storeID, err)
	}
	return retained, nil
}

// ResolveEpochInTx answers R20's epoch-only restriction for address: Live
// while its minting epoch is still current OR its id was carried forward
// into the current epoch by a retained mapping (R20-n), GoneReorganization
// once the epoch has moved past it with no such mapping (a reorganization,
// not a retention or erasure outcome; RetentionFixture/R17 states are out
// of scope here), and Unknown for an address this store never minted.
// ProducingStore is always storeID: this file has no lineage/replica model
// to attribute a foreign store to (unlike RetentionFixture's cross-store
// answers).
func releasedResolveEpochInTx(ctx context.Context, tx DBTX, storeID, address string) (EpochResolveResult, error) {
	row, found, err := readEpochMintedAddressInTx(ctx, tx, address)
	if err != nil {
		return EpochResolveResult{}, err
	}
	if !found || row.storeID != storeID {
		return EpochResolveResult{Restriction: EpochRestrictionUnknown, ProducingStore: storeID}, nil
	}
	epoch, err := readLegacyStoreEpochInTx(ctx, tx)
	if err != nil {
		return EpochResolveResult{}, fmt.Errorf("epoch CAS: resolve %s for %s: %w", address, storeID, err)
	}
	if row.mintedEpoch == epoch {
		return EpochResolveResult{Restriction: EpochRestrictionLive, ProducingStore: storeID, Epoch: &epoch}, nil
	}
	retained, err := mintedIDHasAddressAtEpochInTx(ctx, tx, storeID, row.mintedID, row.mintedEpoch, epoch)
	if err != nil {
		return EpochResolveResult{}, fmt.Errorf("epoch CAS: resolve %s for %s: %w", address, storeID, err)
	}
	if retained {
		return EpochResolveResult{Restriction: EpochRestrictionLive, ProducingStore: storeID, Epoch: &epoch}, nil
	}
	return EpochResolveResult{Restriction: EpochRestrictionGoneReorganization, ProducingStore: storeID, Epoch: &epoch}, nil
}

// CurrentAddressForInTx re-mints oldAddress's underlying id under storeID's
// CURRENT epoch, giving callers a live address to move to once oldAddress
// stops being served. oldAddress must be one storeID has actually minted.
func releasedCurrentAddressForInTx(ctx context.Context, tx DBTX, storeID, oldAddress string) (string, error) {
	row, found, err := readEpochMintedAddressInTx(ctx, tx, oldAddress)
	if err != nil {
		return "", err
	}
	if !found || row.storeID != storeID {
		return "", fmt.Errorf("epoch CAS: current address for %s: %w: %s", storeID, ErrEpochAddressNotFound, oldAddress)
	}
	return releasedMintUnderEpochInTx(ctx, tx, storeID, row.mintedID)
}

// EpochMintDirtyTables includes the singleton initialized by a first mint.
func EpochMintDirtyTables() ChangedTables { return SurvivalEpochMintDirtyTables() }

// ErrEpochSemanticsDeferred identifies APIs whose incoming survival contract
// has not been adopted for released workspaces. Refusal occurs before SQL.
var ErrEpochSemanticsDeferred = errors.New("explicit epoch survival semantics are deferred for released workspaces")

func BumpEpochCarryingInTx(context.Context, DBTX, string, string) (int, error) {
	return 0, ErrEpochSemanticsDeferred
}

func LoseVersionInTx(context.Context, DBTX, string, string) error {
	return ErrEpochSemanticsDeferred
}

// ErrEpochFormatNotReleased rejects formats whose rows cannot safely be
// interpreted using the released epoch contract, including nonmigrating opens.
var ErrEpochFormatNotReleased = errors.New("epoch format is incompatible with released semantics")

const releasedEpochColumnsQuery = `SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'epoch_minted_addresses'`

// requireReleasedEpochShape binds interpretation to the operation's runner.
// It is read-only; schema migration separately checks lineage and constraints.
func requireReleasedEpochShape(ctx context.Context, tx DBTX) error {
	rows, err := tx.QueryContext(ctx, releasedEpochColumnsQuery)
	if err != nil {
		return fmt.Errorf("read epoch format: %w", err)
	}
	defer func() { _ = rows.Close() }()
	expected := map[string]string{"address": "varchar(255)", "store_id": "varchar(255)", "minted_id": "varchar(255)", "minted_epoch": "int", "minted_at": "datetime"}
	seen := make(map[string]bool)
	for rows.Next() {
		var name, typ, nullable, extra string
		var defaultValue sql.NullString
		if err := rows.Scan(&name, &typ, &nullable, &defaultValue, &extra); err != nil {
			return fmt.Errorf("read epoch column: %w", err)
		}
		typ = strings.ToLower(typ)
		if typ == "int(11)" {
			typ = "int"
		}
		if typ == "datetime(0)" {
			typ = "datetime"
		}
		want, ok := expected[name]
		if !ok || seen[name] || typ != want || nullable != "NO" || defaultValue.Valid || extra != "" {
			return fmt.Errorf("%w: unexpected column %q", ErrEpochFormatNotReleased, name)
		}
		seen[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read epoch columns: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close epoch columns: %w", err)
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("%w: missing epoch columns", ErrEpochFormatNotReleased)
	}
	return nil
}

// CurrentEpochInTx applies the released contract only to its compatible format.
func CurrentEpochInTx(ctx context.Context, tx DBTX, storeID string) (int, error) {
	if err := requireReleasedEpochShape(ctx, tx); err != nil {
		return 0, err
	}
	return releasedCurrentEpochInTx(ctx, tx, storeID)
}

// BumpEpochInTx applies the released contract only to its compatible format.
func BumpEpochInTx(ctx context.Context, tx DBTX, storeID, reason string) (int, error) {
	if err := requireReleasedEpochShape(ctx, tx); err != nil {
		return 0, err
	}
	return releasedBumpEpochInTx(ctx, tx, storeID, reason)
}

// MintUnderEpochInTx applies the released contract only to its compatible format.
func MintUnderEpochInTx(ctx context.Context, tx DBTX, storeID, id string) (string, error) {
	if err := validateEpochAddressInputs(storeID, id); err != nil {
		return "", err
	}
	if err := requireReleasedEpochShape(ctx, tx); err != nil {
		return "", err
	}
	return releasedMintUnderEpochInTx(ctx, tx, storeID, id)
}

// StillServesInTx applies the released contract only to its compatible format.
func StillServesInTx(ctx context.Context, tx DBTX, storeID, address string) (bool, error) {
	if err := requireReleasedEpochShape(ctx, tx); err != nil {
		return false, err
	}
	return releasedStillServesInTx(ctx, tx, storeID, address)
}

// ResolveEpochInTx applies the released contract only to its compatible format.
func ResolveEpochInTx(ctx context.Context, tx DBTX, storeID, address string) (EpochResolveResult, error) {
	if err := requireReleasedEpochShape(ctx, tx); err != nil {
		return EpochResolveResult{}, err
	}
	return releasedResolveEpochInTx(ctx, tx, storeID, address)
}

// CurrentAddressForInTx applies the released contract only to its compatible format.
func CurrentAddressForInTx(ctx context.Context, tx DBTX, storeID, oldAddress string) (string, error) {
	if err := requireReleasedEpochShape(ctx, tx); err != nil {
		return "", err
	}
	return releasedCurrentAddressForInTx(ctx, tx, storeID, oldAddress)
}
