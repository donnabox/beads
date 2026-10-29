package graphstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Construction expectations retain their historical nonready limitations. The
// separate compiled native projection binds the selected engine source and full
// table/view facts; syntax parsing or caller observation cannot select it.
//
//go:embed intrinsic_profile_expected.json
var intrinsicExpectedBytes []byte

const intrinsicExpectedSHA256 = "e8f7a3164a4c021fb0608bf74a0dd9a2ba853d794ef97fe935ee126bf4f12c57"

// Conservative logical parse custody charge per input byte: up to eight
// input/value copies, at most one24-byte token and one16-byte column string
// header per byte. This is NOT parser/driver allocation or peak-heap accounting.
const intrinsicSyntaxCopyMultiplier = 48

const intrinsicReceiver = "6dd6a81f5c733a0b13f92f787e5e16c91affbb1d"

// checkedStoreProfile has no constructor accepting caller claims. It requires
// actual binding/schema/policy correspondence to the compiled native projection.
// Eligible operation dispatch follows this fact and precedes body acquisition.
type checkedStoreProfile struct{ schemaVersion int }

// requireBinding consumes only an already checked same-transaction fact. The
// selected/persisted binding must retain that exact supported layout before
// dependent catalog/config/member writes or operation dispatch/body acquisition.
// Bootstrap DDL and binding insertion precede the initializer's fact acquisition.
func (p checkedStoreProfile) requireBinding(b Binding) error {
	if p.schemaVersion != TypesFreshSchemaVersion || p.schemaVersion != b.SchemaVersion {
		return fmt.Errorf("%w: checked intrinsic profile differs from binding", ErrInvalidStore)
	}
	return nil
}

type intrinsicNativeProjection struct {
	Status            string                         `json:"status"`
	Receiver          string                         `json:"receiver"`
	EvidenceLogSHA256 string                         `json:"evidenceLogSHA256"`
	Tables            map[string]string              `json:"tables"`
	Views             map[string]intrinsicNativeView `json:"views"`
	SchemaEnvelope    [][]any                        `json:"schemaEnvelope"`
}
type intrinsicNativeView struct {
	Fragment string  `json:"fragment"`
	Extra    string  `json:"extra"`
	SQLMode  string  `json:"sqlMode"`
	Columns  [][]any `json:"columns"`
}

type intrinsicExpectedManifest struct {
	Native *intrinsicNativeProjection `json:"nativeProjection"`

	Status      string                             `json:"status"`
	Profile     string                             `json:"profile"`
	Receiver    string                             `json:"receiver"`
	Objects     []intrinsicExpectedObject          `json:"objects"`
	Sources     map[string]intrinsicSourceIdentity `json:"sourceIdentities"`
	ObjectCount int                                `json:"applicationObjectCount"`
	TableCount  int                                `json:"tableCount"`
	ViewCount   int                                `json:"viewCount"`
	Ready       bool                               `json:"completeRuntimeComparisonReady"`
	Unfilled    []string                           `json:"honestUnfilledFacts"`
	Requirement string                             `json:"implementationRequirement"`
}

type intrinsicSourceIdentity struct {
	GitBlob string `json:"gitBlob"`
	SHA256  string `json:"sha256"`
	Bytes   int    `json:"bytes"`
}

type intrinsicExpectedObject struct {
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	Create           string   `json:"expectedCreateSQL,omitempty"`
	ViewCreate       string   `json:"expectedPersistedCreateSQL,omitempty"`
	Aliases          []string `json:"expectedExplicitAliasList,omitempty"`
	ResolvedNames    []string `json:"expectedResolvedColumnNames,omitempty"`
	Modes            []string `json:"expectedSqlModeSet,omitempty"`
	Sources          []string `json:"constructionSource"`
	Ignore           string   `json:"expectedIgnoreResult,omitempty"`
	SourceValues     string   `json:"sourceValues,omitempty"`
	ProjectionStatus string   `json:"nativeProjectionStatus"`
}

// Loading is lazy, deterministic and pure: no init(), parser, DB or module I/O.
// Digest is checked against this source-owned constant, not a caller digest.
func loadIntrinsicExpected() (intrinsicExpectedManifest, error) {
	return decodeIntrinsicExpected(intrinsicExpectedBytes)
}

func decodeIntrinsicExpected(raw []byte) (intrinsicExpectedManifest, error) {
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != intrinsicExpectedSHA256 {
		return intrinsicExpectedManifest{}, fmt.Errorf("%w: intrinsic expectation source bytes", ErrInvalidStore)
	}
	var m intrinsicExpectedManifest
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return intrinsicExpectedManifest{}, fmt.Errorf("%w: expectation JSON: %v", ErrInvalidStore, err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return intrinsicExpectedManifest{}, fmt.Errorf("%w: expectation trailing JSON", ErrInvalidStore)
	}
	if m.Receiver != intrinsicReceiver || m.Ready || len(m.Unfilled) != 4 || len(m.Sources) != 65 ||
		m.ObjectCount != 44 || m.TableCount != 42 || m.ViewCount != 2 || len(m.Objects) != 44 {
		return intrinsicExpectedManifest{}, fmt.Errorf("%w: unqualified expectation contract", ErrInvalidStore)
	}
	seen := make(map[string]bool)
	tables, views := 0, 0
	for _, object := range m.Objects {
		if !databaseName.MatchString(object.Name) || seen[object.Name] || len(object.Sources) == 0 || !strings.HasPrefix(object.ProjectionStatus, "UNQUALIFIED:") {
			return intrinsicExpectedManifest{}, fmt.Errorf("%w: expectation object", ErrInvalidStore)
		}
		seen[object.Name] = true
		for _, path := range object.Sources {
			if _, ok := m.Sources[path]; !ok && path != "private:R3-NATIVE-PRODUCER-PLAN02.md@90f2f535" {
				return intrinsicExpectedManifest{}, fmt.Errorf("%w: expectation source correspondence", ErrInvalidStore)
			}
		}
		switch object.Kind {
		case "BASE TABLE":
			tables++
			if object.Create == "" || object.ViewCreate != "" || (object.Ignore != "Ignore" && object.Ignore != "DontIgnore") {
				return intrinsicExpectedManifest{}, fmt.Errorf("%w: expectation table", ErrInvalidStore)
			}
		case "VIEW":
			views++
			if object.ViewCreate == "" || object.Create != "" || len(object.Aliases) != 0 || len(object.ResolvedNames) == 0 {
				return intrinsicExpectedManifest{}, fmt.Errorf("%w: expectation view", ErrInvalidStore)
			}
		default:
			return intrinsicExpectedManifest{}, fmt.Errorf("%w: expectation kind", ErrInvalidStore)
		}
	}
	if tables != 42 || views != 2 {
		return intrinsicExpectedManifest{}, fmt.Errorf("%w: expectation census", ErrInvalidStore)
	}
	return m, nil
}

// Budget is operation-owned and shared with any later body acquisition. It
// counts logical retained cells/copies only; native SHOW already materializes
// cells before application bounds, and the parser cannot be preempted by ctx.
type intrinsicMetadataBudget struct{ remaining, metadata int }

func newIntrinsicMetadataBudget() *intrinsicMetadataBudget {
	return &intrinsicMetadataBudget{remaining: PreviewCurrentReadByteLimit}
}

// chargeBody shares the operation allowance without spending its distinct
// metadata sublimit. All debits precede database byte acquisition/copy creation;
// this is logical custody, not a claim about engine/driver peak allocation.
func (b *intrinsicMetadataBudget) chargeBody(n int) error {
	if b == nil || b.remaining < 0 || b.remaining > PreviewCurrentReadByteLimit || n < 0 || n > b.remaining {
		return fmt.Errorf("%w: intrinsic operation custody", ErrLimitExceeded)
	}
	b.remaining -= n
	return nil
}

func (b *intrinsicMetadataBudget) charge(n int) error {
	if b == nil || b.remaining < 0 || b.remaining > PreviewCurrentReadByteLimit || n < 0 || n > b.remaining || n > (4<<20)-b.metadata {
		return fmt.Errorf("%w: intrinsic metadata shared acquisition", ErrLimitExceeded)
	}
	b.remaining -= n
	b.metadata += n
	return nil
}

func intrinsicCellString(raw any) (string, error) {
	switch v := raw.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	}
	return "", fmt.Errorf("%w: metadata text encoding", ErrInvalidStore)
}

func intrinsicUnsigned(raw any) (uint64, error) {
	// The selected native engine returns SQL INT as int32, BIGINT as int64,
	// boolean-system scalars as int8 and unsigned aggregates as uint64.
	// Widen only exact integral representations; never float/string coercion.
	switch v := raw.(type) {
	case int8:
		return intrinsicUnsigned(int64(v))
	case int16:
		return intrinsicUnsigned(int64(v))
	case int32:
		return intrinsicUnsigned(int64(v))
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	}
	if v, ok := raw.(int64); ok {
		if v >= 0 {
			return uint64(v), nil
		}
		return 0, fmt.Errorf("%w: negative metadata scalar", ErrInvalidStore)
	}
	text, err := intrinsicCellString(raw)
	if err != nil || text == "" || len(text) > 20 {
		return 0, fmt.Errorf("%w: metadata integer", ErrInvalidStore)
	}
	for _, c := range text {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("%w: metadata integer", ErrInvalidStore)
		}
	}
	v, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: metadata integer: %v", ErrInvalidStore, err)
	}
	return v, nil
}

func intrinsicQuote(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

// Only package-owned constant queries or manifest-literal identifiers reach
// this helper. It is NOT a public arbitrary SQL facility. Every exit joins
// driver EOF/Scan/Close errors; no truncated acquisition can succeed.
func readIntrinsicMetadataRows(ctx context.Context, tx *sql.Tx, b *intrinsicMetadataBudget, query string, columns, maximum int, args ...any) (out [][]any, err error) {
	if tx == nil || columns < 1 || columns > 9 || maximum < 0 || maximum > 256 {
		return nil, fmt.Errorf("%w: metadata query contract", ErrInvalidStore)
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Join(ErrInvalidStore, err)
	}
	defer func() {
		closeErr := rows.Close()
		err = errors.Join(err, rows.Err(), closeErr)
		if err != nil {
			out = nil
		}
	}()
	names, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if len(names) != columns {
		return nil, fmt.Errorf("%w: metadata result columns", ErrInvalidStore)
	}
	for rows.Next() {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		if len(out) == maximum {
			return nil, fmt.Errorf("%w: extra metadata row", ErrInvalidStore)
		}
		values := make([]any, columns)
		pointers := make([]any, columns)
		for i := range values {
			pointers[i] = &values[i]
		}
		if err = rows.Scan(pointers...); err != nil {
			return nil, err
		}
		if err = b.charge(256); err != nil {
			return nil, err
		}
		for i, value := range values {
			switch v := value.(type) {
			case []byte:
				if !utf8.Valid(v) {
					return nil, fmt.Errorf("%w: metadata UTF-8", ErrInvalidStore)
				}
				if len(v) > intrinsicStatementLimit {
					return nil, fmt.Errorf("%w: metadata cell", ErrLimitExceeded)
				}
				if err = b.charge(len(v)); err != nil {
					return nil, err
				}
				values[i] = string(v)
			case string:
				if !utf8.ValidString(v) {
					return nil, fmt.Errorf("%w: metadata UTF-8", ErrInvalidStore)
				}
				if len(v) > intrinsicStatementLimit {
					return nil, fmt.Errorf("%w: metadata cell", ErrLimitExceeded)
				}
				if err = b.charge(len(v)); err != nil {
					return nil, err
				}
				values[i] = strings.Clone(v)
			case nil, bool, int8, int16, int32, int64, uint8, uint16, uint32, uint64: // fixed driver scalar forms only
			default:
				return nil, fmt.Errorf("%w: metadata scalar encoding %T in %s", ErrInvalidStore, value, query)
			}
		}
		out = append(out, values)
	}
	if rows.NextResultSet() {
		return nil, fmt.Errorf("%w: extra metadata result set", ErrInvalidStore)
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func intrinsicOne(rows [][]any, columns int) ([]any, error) {
	if len(rows) != 1 || len(rows[0]) != columns {
		return nil, fmt.Errorf("%w: metadata singleton", ErrInvalidStore)
	}
	return rows[0], nil
}

const intrinsicBindingQuery = `SELECT workspace,scope_url,authority_id,schema_version,writer_token FROM graph_preview_scope WHERE singleton=1 AND OCTET_LENGTH(workspace)=? AND OCTET_LENGTH(scope_url)=?`

func readIntrinsicBinding(ctx context.Context, tx *sql.Tx, o Options, b *intrinsicMetadataBudget) error {
	rows, err := readIntrinsicMetadataRows(ctx, tx, b, intrinsicBindingQuery, 5, 1, len(o.Binding.WorkspaceID), len(o.Binding.ScopeURL))
	if err != nil {
		return err
	}
	row, err := intrinsicOne(rows, 5)
	if err != nil {
		return err
	}
	workspace, e1 := intrinsicCellString(row[0])
	scope, e2 := intrinsicCellString(row[1])
	authority, e3 := intrinsicCellString(row[2])
	version, e4 := intrinsicUnsigned(row[3])
	token, e5 := intrinsicCellString(row[4])
	if err = errors.Join(e1, e2, e3, e4, e5); err != nil {
		return err
	}
	if workspace != o.Binding.WorkspaceID || scope != o.Binding.ScopeURL || authority != o.Binding.AuthorityID || version != uint64(TypesFreshSchemaVersion) || version != uint64(o.Binding.SchemaVersion) || !authorityID.MatchString(authority) || !authorityID.MatchString(token) {
		return fmt.Errorf("%w: actual persisted intrinsic binding", ErrInvalidStore)
	}
	return nil
}

func readIntrinsicAggregate(ctx context.Context, tx *sql.Tx, b *intrinsicMetadataBudget, query string, maxRows, maxBytes uint64) error {
	rows, err := readIntrinsicMetadataRows(ctx, tx, b, query, 2, 1, maxBytes+1)
	if err != nil {
		return err
	}
	row, err := intrinsicOne(rows, 2)
	if err != nil {
		return err
	}
	count, e1 := intrinsicUnsigned(row[0])
	size, e2 := intrinsicUnsigned(row[1])
	if err = errors.Join(e1, e2); err != nil {
		return err
	}
	if count > maxRows || size > maxBytes {
		return fmt.Errorf("%w: metadata aggregate", ErrLimitExceeded)
	}
	return nil
}

// acquireIntrinsicMetadata performs no intrinsic context/member/body query.
// Syntax alone never authenticates engine equivalence. A selected compiled
// projection additionally requires exact table, persisted-view and typed-column
// correspondence; unknown or historical candidates cannot emit a checked fact.
func acquireIntrinsicMetadata(ctx context.Context, tx *sql.Tx, o Options, m intrinsicExpectedManifest, b *intrinsicMetadataBudget) error {
	db := intrinsicQuote(o.Database)
	rows, err := readIntrinsicMetadataRows(ctx, tx, b, "SELECT @@sql_mode, @@dolt_show_system_tables", 2, 1)
	if err != nil {
		return err
	}
	row, err := intrinsicOne(rows, 2)
	if err != nil {
		return err
	}
	mode, e1 := intrinsicCellString(row[0])
	system, e2 := intrinsicUnsigned(row[1])
	if err = errors.Join(e1, e2); err != nil {
		return err
	}
	if err = intrinsicSQLModes(mode); err != nil {
		return err
	}
	if system != 0 {
		return fmt.Errorf("%w: system table enumeration mode", ErrInvalidStore)
	}
	rows, err = readIntrinsicMetadataRows(ctx, tx, b, "SHOW FULL TABLES FROM "+db, 2, 44)
	if err != nil {
		return err
	}
	if len(rows) != 44 {
		return fmt.Errorf("%w: application object census", ErrInvalidStore)
	}
	objects := make(map[string]intrinsicExpectedObject, len(m.Objects))
	for _, object := range m.Objects {
		objects[object.Name] = object
	}
	seen := make(map[string]bool)
	for _, r := range rows {
		name, e1 := intrinsicCellString(r[0])
		kind, e2 := intrinsicCellString(r[1])
		if err = errors.Join(e1, e2); err != nil {
			return err
		}
		object, ok := objects[name]
		if !ok || seen[name] || object.Kind != kind {
			return fmt.Errorf("%w: literal application object", ErrInvalidStore)
		}
		seen[name] = true
	}
	for _, object := range m.Objects {
		if object.Kind != "BASE TABLE" {
			continue
		}
		rows, err = readIntrinsicMetadataRows(ctx, tx, b, "SHOW CREATE TABLE "+db+"."+intrinsicQuote(object.Name), 2, 1)
		if err != nil {
			return err
		}
		row, err = intrinsicOne(rows, 2)
		if err != nil {
			return err
		}
		name, e1 := intrinsicCellString(row[0])
		ddl, e2 := intrinsicCellString(row[1])
		if err = errors.Join(e1, e2); err != nil {
			return err
		}
		if name != object.Name {
			return fmt.Errorf("%w: returned DDL identity", ErrInvalidStore)
		}
		if err = b.charge(intrinsicSyntaxCopyMultiplier * len(ddl)); err != nil {
			return err
		}
		syntax, parseErr := parseIntrinsicSyntax(ctx, ddl, name, object.Kind)
		if parseErr != nil {
			return fmt.Errorf("%s %s: %w", object.Kind, name, parseErr)
		}
		if intrinsicNativeSelected(m) && m.Native.Tables[syntax.name] != ddl {
			return fmt.Errorf("%w: fixed native table correspondence %s", ErrInvalidStore, name)
		}
	}
	// Compare the full selected envelope and resolved view type/null/collation
	// facts below; historical construction expectations alone are insufficient.
	rows, err = readIntrinsicMetadataRows(ctx, tx, b, "SHOW FULL COLUMNS FROM "+db+".`dolt_schemas`", 9, 5)
	if err != nil {
		return err
	}
	if len(rows) != 5 {
		return fmt.Errorf("%w: schema envelope", ErrInvalidStore)
	}
	if intrinsicNativeSelected(m) && !reflect.DeepEqual(rows, m.Native.SchemaEnvelope) {
		return fmt.Errorf("%w: fixed native schema envelope", ErrInvalidStore)
	}
	for i, name := range []string{"type", "name", "fragment", "extra", "sql_mode"} {
		field, err := intrinsicCellString(rows[i][0])
		if err != nil || field != name {
			return fmt.Errorf("%w: schema envelope order", ErrInvalidStore)
		}
	}
	const schemaAggregate = "SELECT COUNT(*),CAST(LEAST(COALESCE(SUM(COALESCE(OCTET_LENGTH(type),0)+COALESCE(OCTET_LENGTH(name),0)+COALESCE(OCTET_LENGTH(fragment),0)+COALESCE(OCTET_LENGTH(extra),0)+COALESCE(OCTET_LENGTH(sql_mode),0)),0),?) AS UNSIGNED) FROM dolt_schemas"
	if err = readIntrinsicAggregate(ctx, tx, b, schemaAggregate, 2, 2*intrinsicStatementLimit); err != nil {
		return err
	}
	rows, err = readIntrinsicMetadataRows(ctx, tx, b, "SELECT type,name,fragment,extra,sql_mode FROM "+db+".`dolt_schemas` ORDER BY BINARY type,BINARY name", 5, 2)
	if err != nil {
		return err
	}
	if len(rows) != 2 {
		return fmt.Errorf("%w: persisted view census", ErrInvalidStore)
	}
	viewSeen := make(map[string]bool)
	for _, r := range rows {
		kind, e1 := intrinsicCellString(r[0])
		name, e2 := intrinsicCellString(r[1])
		fragment, e3 := intrinsicCellString(r[2])
		sqlMode, e4 := intrinsicCellString(r[4])
		if err = errors.Join(e1, e2, e3, e4); err != nil {
			return err
		}
		object, ok := objects[name]
		if kind != "view" || !ok || object.Kind != "VIEW" || viewSeen[name] {
			return fmt.Errorf("%w: persisted view identity", ErrInvalidStore)
		}
		viewSeen[name] = true
		if intrinsicNativeSelected(m) {
			expected, exists := m.Native.Views[name]
			extra, e := intrinsicCellString(r[3])
			if !exists || e != nil || fragment != expected.Fragment || extra != expected.Extra || sqlMode != expected.SQLMode {
				return fmt.Errorf("%w: fixed native view correspondence %s", ErrInvalidStore, name)
			}
		}

		if err = intrinsicSQLModes(sqlMode); err != nil {
			return err
		}
		if err = b.charge(intrinsicSyntaxCopyMultiplier * len(fragment)); err != nil {
			return err
		}
		syntax, parseErr := parseIntrinsicSyntax(ctx, fragment, name, "VIEW")
		if parseErr != nil {
			return parseErr
		}
		resolved, err := readIntrinsicMetadataRows(ctx, tx, b, "SHOW FULL COLUMNS FROM "+db+"."+intrinsicQuote(syntax.name), 9, 128)
		if err != nil {
			return err
		}
		if intrinsicNativeSelected(m) && !reflect.DeepEqual(resolved, m.Native.Views[syntax.name].Columns) {
			return fmt.Errorf("%w: fixed native resolved view %s", ErrInvalidStore, name)
		}
		if len(resolved) != len(object.ResolvedNames) {
			return fmt.Errorf("%w: resolved view census", ErrInvalidStore)
		}
		for i, want := range object.ResolvedNames {
			got, err := intrinsicCellString(resolved[i][0])
			if err != nil || got != want {
				return fmt.Errorf("%w: resolved view field", ErrInvalidStore)
			}
		}
	}
	const ignoreAggregate = "SELECT COUNT(*),CAST(LEAST(COALESCE(SUM(COALESCE(OCTET_LENGTH(pattern),0)+COALESCE(OCTET_LENGTH(ignored),0)),0),?) AS UNSIGNED) FROM dolt_ignore"
	if err = readIntrinsicAggregate(ctx, tx, b, ignoreAggregate, 256, 64<<10); err != nil {
		return err
	}
	rows, err = readIntrinsicMetadataRows(ctx, tx, b, "SELECT pattern,ignored FROM "+db+".`dolt_ignore` ORDER BY BINARY pattern", 2, 256)
	if err != nil {
		return err
	}
	return checkIntrinsicIgnorePolicy(m.Objects, rows)
}

// Fresh initialization consumes this fact after bootstrap DDL/binding insertion
// and before dependent catalog/config/member writes. Types14 admission consumes
// it before operation dispatch/body acquisition. Original11 diagnosis is unchanged.
// This function never calls
// checkBinding (which loads old installed descriptors), checkIntrinsicState,
// currentSnapshot or a source/member/record/owned-Link hydration function.
func checkTypesFreshProfileInTx(ctx context.Context, tx *sql.Tx, o Options, b *intrinsicMetadataBudget) (checkedStoreProfile, error) {
	if o.Binding.SchemaVersion != TypesFreshSchemaVersion || !databaseName.MatchString(o.Database) || tx == nil {
		return checkedStoreProfile{}, fmt.Errorf("%w: intrinsic metadata request", ErrInvalidStore)
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	m, err := loadIntrinsicExpected()
	if err != nil {
		return checkedStoreProfile{}, err
	}
	if err = readIntrinsicBinding(ctx, tx, o, b); err != nil {
		return checkedStoreProfile{}, err
	}
	if err = acquireIntrinsicMetadata(ctx, tx, o, m, b); err != nil {
		return checkedStoreProfile{}, err
	}
	if intrinsicNativeSelected(m) {
		return checkedStoreProfile{schemaVersion: TypesFreshSchemaVersion}, nil
	}
	// Unknown native projections remain a hard stop. Syntax acquisition or a
	// caller boolean/digest cannot promote an unselected source manifest.
	return checkedStoreProfile{}, fmt.Errorf("%w: Types14 canonical schema/view correspondence remains unqualified", ErrInvalidStore)
}

// This discriminator names a compiled, source-owned exact projection, not a
// caller readiness switch. The complete byte-bound projection is compared by
// acquireIntrinsicMetadata; unknown projections cannot construct a checked fact.
func intrinsicNativeSelected(m intrinsicExpectedManifest) bool {
	return m.Native != nil && m.Native.Status == "FIXED_SELECTED_ENGINE_PROJECTION_V1" && m.Native.Receiver == intrinsicReceiver && len(m.Native.Tables) == 42 && len(m.Native.Views) == 2 && len(m.Native.SchemaEnvelope) == 5
}
