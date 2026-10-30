package schema

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"
	"unicode"
)

const forkEpochExtensionVersion = 75

// EpochMigrationCompatibilityError refuses an unrecognized migration lineage
// or an epoch format for which this binary has no loss-preserving adapter.
// Migration consent and the ordinary schema-skew escape hatch cannot authorize
// this conversion. Admission only reads metadata and bounded existence probes.
type EpochMigrationCompatibilityError struct{ Reason string }

func (e *EpochMigrationCompatibilityError) Error() string {
	return "epoch migration compatibility: " + e.Reason + "; this migration route is refused"
}

func epochCompatibility(reason string) error {
	return &EpochMigrationCompatibilityError{Reason: reason}
}

var forkEpochLineageFiles = map[int]string{
	70: "0070_add_removed_restriction.up.sql",
	71: "0071_add_epoch_minted_addresses.up.sql",
	72: "0072_add_participation_generation.up.sql",
	73: "0073_add_arrival_seq.up.sql",
	74: "0074_create_id_tombstones.up.sql",
	75: "0075_extend_epoch_survival.up.sql",
}

// checkForkEpochAdmission must run before bootstrap, ignored-state repair or
// numbered migration writes. target is the requested MAIN migration target;
// historical MigrateUpTo targets below74 may construct an old disposable store,
// but never bypass collision-range lineage validation. The caller owns the
// existing exclusive migration interval; this read is not a writer lock.
func checkForkEpochAdmission(ctx context.Context, db DBConn, target int) error {
	if target == 0 {
		target = mainSource.latest()
	}
	cursor, err := tableInCurrentDatabase(ctx, db, "schema_migrations")
	if err != nil {
		return fmt.Errorf("probing epoch migration lineage: %w", err)
	}
	epoch, err := tableInCurrentDatabase(ctx, db, "epoch_minted_addresses")
	if err != nil {
		return fmt.Errorf("probing epoch address table: %w", err)
	}
	if !cursor {
		if epoch {
			return epochCompatibility("epoch table exists without recorded lineage")
		}
		return nil
	}
	var rawMax int
	if err := db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&rawMax); err != nil {
		return fmt.Errorf("reading raw epoch migration cursor: %w", err)
	}
	if rawMax < 0 {
		return epochCompatibility("negative raw migration cursor")
	}
	if rawMax > forkEpochExtensionVersion {
		return epochCompatibility("raw migration cursor exceeds the selected epoch lineage")
	}
	if rawMax >= 70 {
		hasHash, err := showColumnExists(ctx, db, "schema_migrations", "content_hash")
		if err != nil {
			return err
		}
		if !hasHash {
			return epochCompatibility("collision-range cursor has no content_hash column")
		}
		rows, err := db.QueryContext(ctx, "SELECT version, content_hash FROM schema_migrations WHERE version BETWEEN 70 AND 75 ORDER BY version")
		if err != nil {
			return fmt.Errorf("reading collision-range lineage: %w", err)
		}
		hashes := map[int]sql.NullString{}
		for rows.Next() {
			var version int
			var hash sql.NullString
			if err := rows.Scan(&version, &hash); err != nil {
				_ = rows.Close()
				return err
			}
			if _, exists := hashes[version]; exists {
				_ = rows.Close()
				return epochCompatibility("duplicate collision-range cursor row")
			}
			hashes[version] = hash
		}
		readErr := rows.Err()
		closeErr := rows.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		for version := 70; version <= rawMax && version <= forkEpochExtensionVersion; version++ {
			hash, exists := hashes[version]
			if !exists || !hash.Valid || hash.String == "" {
				return epochCompatibility(fmt.Sprintf("migration%04d has missing or NULL lineage", version))
			}
			data, err := mainSource.files.ReadFile("migrations/" + forkEpochLineageFiles[version])
			if err != nil {
				return fmt.Errorf("reading selected epoch lineage: %w", err)
			}
			want := fmt.Sprintf("%x", sha256.Sum256(data))
			if hash.String != want {
				return epochCompatibility(fmt.Sprintf("migration%04d hash %q differs from selected fork hash %q", version, hash.String, want))
			}
		}
	}
	if rawMax < 70 && epoch {
		return epochCompatibility("epoch table predates recognized epoch lineage")
	}
	if rawMax >= 71 && !epoch {
		return epochCompatibility("recorded epoch migration has no epoch table")
	}
	if rawMax < forkEpochExtensionVersion && target < forkEpochExtensionVersion {
		return nil
	}
	if !epoch {
		return nil
	} // Successfully observed fresh/pre-epoch absence.
	// GET_LOCK serializes migrators, but already-open old clients do not
	// honor it. Even an empty table can receive an old-format row between
	// this probe and the independently durable DDL. No existing epoch
	// upgrade is safe until a separately qualified writer fence exists.
	if rawMax < forkEpochExtensionVersion {
		return epochCompatibility("existing old or partial epoch table requires a separately qualified exclusive-writer adapter")
	}
	complete, err := readForkEpochShape(ctx, db)
	if err != nil {
		return err
	}
	if !complete {
		return epochCompatibility("completed epoch extension has incomplete shape")
	}
	return nil
}

type epochColumnShape struct{ typ, nullable, key string }

// readForkEpochShape recognizes only the frozen base plus additive extension
// objects. Names alone do not prove the types, index columns or CHECK meaning.
// It is called only after a successful table existence probe.
func readForkEpochShape(ctx context.Context, db DBConn) (bool, error) {
	rows, err := db.QueryContext(ctx, "SHOW COLUMNS FROM epoch_minted_addresses")
	if err != nil {
		return false, err
	}
	columns, err := rows.Columns()
	if err != nil {
		_ = rows.Close()
		return false, err
	}
	positions := map[string]int{}
	for i, name := range columns {
		positions[strings.ToLower(name)] = i
	}
	for _, name := range []string{"field", "type", "null", "key", "default", "extra"} {
		if _, ok := positions[name]; !ok {
			_ = rows.Close()
			return false, epochCompatibility("unreadable epoch column metadata")
		}
	}
	shapes := map[string]epochColumnShape{}
	cells := make([]sql.RawBytes, len(columns))
	dest := make([]any, len(columns))
	for i := range cells {
		dest[i] = &cells[i]
	}
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			_ = rows.Close()
			return false, err
		}
		name := string(cells[positions["field"]])
		if _, ok := shapes[name]; ok {
			_ = rows.Close()
			return false, epochCompatibility("duplicate epoch column metadata")
		}
		if cells[positions["default"]] != nil || len(cells[positions["extra"]]) != 0 {
			_ = rows.Close()
			return false, epochCompatibility("unexpected epoch column default or generated behavior")
		}
		shapes[name] = epochColumnShape{strings.ToLower(string(cells[positions["type"]])), string(cells[positions["null"]]), string(cells[positions["key"]])}
	}
	readErr := rows.Err()
	closeErr := rows.Close()
	if readErr != nil {
		return false, readErr
	}
	if closeErr != nil {
		return false, closeErr
	}
	base := map[string]epochColumnShape{
		"address": {"varchar(255)", "NO", "PRI"}, "store_id": {"varchar(255)", "NO", ""}, "minted_id": {"varchar(255)", "NO", ""},
		"minted_epoch": {"int", "NO", ""}, "minted_at": {"datetime", "NO", ""},
	}
	for name, want := range base {
		got, ok := shapes[name]
		if got.typ == "int(11)" {
			got.typ = "int"
		}
		if got.typ == "datetime(0)" {
			got.typ = "datetime"
		}
		if !ok || got != want {
			return false, epochCompatibility("unexpected base epoch column shape: " + name)
		}
	}
	for name, got := range shapes {
		if _, ok := base[name]; ok {
			continue
		}
		if name != "carried_from" && name != "gone_at_epoch" {
			return false, epochCompatibility("unrecognized epoch column: " + name)
		}
		want := "varchar(255)"
		if name == "gone_at_epoch" {
			want = "int"
		}
		if got.typ == "int(11)" {
			got.typ = "int"
		}
		if got.typ != want || got.nullable != "YES" || got.key != "" && !(name == "carried_from" && got.key == "MUL") {
			return false, epochCompatibility("unexpected survival column shape: " + name)
		}
	}
	indexRows, err := db.QueryContext(ctx, "SELECT INDEX_NAME, NON_UNIQUE, SEQ_IN_INDEX, COLUMN_NAME, SUB_PART FROM information_schema.statistics WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'epoch_minted_addresses' ORDER BY INDEX_NAME, SEQ_IN_INDEX")
	if err != nil {
		return false, err
	}
	primary, carryIndex := false, false
	for indexRows.Next() {
		var name, column string
		var nonUnique, sequence int
		var prefix sql.NullInt64
		if err := indexRows.Scan(&name, &nonUnique, &sequence, &column, &prefix); err != nil {
			_ = indexRows.Close()
			return false, err
		}
		if prefix.Valid {
			_ = indexRows.Close()
			return false, epochCompatibility("prefix epoch index changes identity or lookup semantics")
		}
		switch {
		case name == "PRIMARY" && column == "address" && nonUnique == 0 && sequence == 1 && !primary:
			primary = true
		case name == "idx_epoch_minted_addresses_carried_from" && column == "carried_from" && nonUnique == 1 && sequence == 1 && !carryIndex:
			carryIndex = true
		default:
			_ = indexRows.Close()
			return false, epochCompatibility("unexpected epoch index definition")
		}
	}
	readErr = indexRows.Err()
	closeErr = indexRows.Close()
	if readErr != nil {
		return false, readErr
	}
	if closeErr != nil {
		return false, closeErr
	}
	if !primary {
		return false, epochCompatibility("missing epoch address primary key")
	}
	if carryIndex {
		if _, ok := shapes["carried_from"]; !ok {
			return false, epochCompatibility("carry index without carry column")
		}
	}
	constraintRows, err := db.QueryContext(ctx, "SELECT tc.CONSTRAINT_NAME, tc.CONSTRAINT_TYPE, cc.CHECK_CLAUSE, tc.ENFORCED FROM information_schema.table_constraints tc LEFT JOIN information_schema.check_constraints cc ON cc.CONSTRAINT_SCHEMA = tc.CONSTRAINT_SCHEMA AND cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME WHERE tc.TABLE_SCHEMA = DATABASE() AND tc.TABLE_NAME = 'epoch_minted_addresses'")
	if err != nil {
		return false, err
	}
	check := false
	for constraintRows.Next() {
		var name, kind, enforced string
		var clause sql.NullString
		if err := constraintRows.Scan(&name, &kind, &clause, &enforced); err != nil {
			_ = constraintRows.Close()
			return false, err
		}
		if kind == "PRIMARY KEY" && name == "PRIMARY" && enforced == "YES" {
			continue
		}
		if kind != "CHECK" || enforced != "YES" || name != "ck_epoch_minted_addresses_gone_not_before_mint" || check || !clause.Valid || !epochCheckMatches(clause.String) {
			_ = constraintRows.Close()
			return false, epochCompatibility("unexpected epoch constraint definition")
		}
		check = true
	}
	readErr = constraintRows.Err()
	closeErr = constraintRows.Close()
	if readErr != nil {
		return false, readErr
	}
	if closeErr != nil {
		return false, closeErr
	}
	_, carry := shapes["carried_from"]
	_, gone := shapes["gone_at_epoch"]
	if check && !gone {
		return false, epochCompatibility("loss constraint without loss column")
	}
	return carry && gone && carryIndex && check, nil
}

// normalizeEpochCheck keeps grouping. Only redundant wrappers around an entire
// expression or either complete OR arm may disappear; arbitrary internal
// parentheses can change SQL comparison precedence and must not be erased.
func normalizeEpochCheck(clause string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '`' {
			return -1
		}
		return unicode.ToLower(r)
	}, clause)
}

func epochCheckMatches(clause string) bool {
	text := normalizeEpochCheck(clause)
	stripOuter := func(s string) string {
		for len(s) > 1 && s[0] == '(' && s[len(s)-1] == ')' {
			depth := 0
			whole := true
			for i, r := range s {
				if r == '(' {
					depth++
				}
				if r == ')' {
					depth--
				}
				if depth < 0 || depth == 0 && i != len(s)-1 {
					whole = false
					break
				}
			}
			if !whole || depth != 0 {
				break
			}
			s = s[1 : len(s)-1]
		}
		return s
	}
	text = stripOuter(text)
	depth, split := 0, -1
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth < 0 {
			return false
		}
		if depth == 0 && i+1 < len(text) && text[i:i+2] == "or" {
			if split != -1 {
				return false
			}
			split = i
			i++
		}
	}
	if depth != 0 || split < 0 {
		return false
	}
	return stripOuter(text[:split]) == "gone_at_epochisnull" && stripOuter(text[split+2:]) == "gone_at_epoch>=minted_epoch"
}
