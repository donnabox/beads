// Package epochfixture prepares the deliberately non-default epoch contract
// only in newly owned disposable test databases. Production opens never use it.
package epochfixture

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"regexp"
)

//go:embed 0075_extend_epoch_survival.up.sql
var SQL string

// Conn is the pinned connection used to prepare a private fixture.
type Conn interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// Statements derives the four fresh-only DDL operations from the preserved
// guarded source. It does not replay or convert any retained address rows.
func Statements() ([]string, error) {
	re := regexp.MustCompile(`(?s)SET @sql = IF\(@epoch_[a-z]+_needs_add = 1,\s*'([^']+)',\s*'SELECT 1'\);`)
	matches := re.FindAllStringSubmatch(SQL, -1)
	if len(matches) != 4 {
		return nil, fmt.Errorf("deferred fixture has %d DDL steps, want four", len(matches))
	}
	statements := make([]string, len(matches))
	for i, match := range matches {
		statements[i] = match[1]
	}
	return statements, nil
}

// Prepare requires a normal74 database with no minted addresses. The test
// author owns this private database; this is neither a user-store converter nor
// a writer fence. The production cursor remains74 and no ordinary reopen is
// permitted after choosing the explicit Survival test adapter.
func Prepare(ctx context.Context, db Conn) error {
	var version, rows int
	if err := db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version); err != nil {
		return err
	}
	if version != 74 {
		return fmt.Errorf("private survival fixture requires production cursor74, got%d", version)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM epoch_minted_addresses").Scan(&rows); err != nil {
		return err
	}
	if rows != 0 {
		return fmt.Errorf("private survival fixture refuses %d retained addresses", rows)
	}
	statements, err := Statements()
	if err != nil {
		return err
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("private survival fixture DDL: %w", err)
		}
	}
	checks := []string{
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'epoch_minted_addresses' AND ((column_name = 'carried_from' AND column_type = 'varchar(255)' AND is_nullable = 'YES') OR (column_name = 'gone_at_epoch' AND column_type IN ('int', 'int(11)') AND is_nullable = 'YES'))",
		"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'epoch_minted_addresses' AND index_name = 'idx_epoch_minted_addresses_carried_from' AND column_name = 'carried_from' AND seq_in_index = 1 AND sub_part IS NULL",
		"SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'epoch_minted_addresses' AND constraint_name = 'ck_epoch_minted_addresses_gone_not_before_mint' AND constraint_type = 'CHECK' AND enforced = 'YES'",
	}
	for i, query := range checks {
		var got int
		if err := db.QueryRowContext(ctx, query).Scan(&got); err != nil {
			return err
		}
		want := 1
		if i == 0 {
			want = 2
		}
		if got != want {
			return fmt.Errorf("private survival shape check%d got%d,want%d", i, got, want)
		}
	}
	if _, err := db.ExecContext(ctx, "CALL DOLT_ADD('epoch_minted_addresses')"); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, "CALL DOLT_COMMIT('-m', 'prepare private epoch survival fixture')"); err != nil {
		return err
	}
	return nil
}
