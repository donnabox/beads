package graphops

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
)

// These private values describe physical observations, never admission. Zero
// presence is invalid; no observation loads a witness, repairs, or grants a lease.
type observationPresence uint8

const (
	observationAbsent observationPresence = iota + 1
	observationPresent
)

// civilTimestamp holds DATETIME(6) fields, not a UTC instant. Even a reader in
// UTC cannot establish a historical writer's zone from this naive value.
type civilTimestamp string

type scopeObservation struct {
	presence         observationPresence
	url, authorityID string
	epoch            uint64
	mintedAt         civilTimestamp
}
type ledgerPoint struct {
	presence observationPresence
	seq      uint64
	hash     string
}
type ledgerObservation struct {
	requested     *uint64     // copied operand, nil is distinct from a requested absent row
	recorded, tip ledgerPoint // tip.seq is MAX(seq) under the required B4 primary key
	head          string      // Dolt HEAD, not the ledger hash or proof of ancestry
}
type leaseObservation struct {
	presence                                      observationPresence
	scopeURL, authorityID, holder, renewer, fence string
	epoch                                         uint64
	grantedAt, expiresAt, heartbeatAt             civilTimestamp
	clock                                         civilTimestamp
	zone                                          string
	queryStarted, queryFinished                   time.Time // local monotonic interval, not DB wall time
}

const scopeObservationQuery = `SELECT id, scope_url, authority_id, CAST(authority_epoch AS CHAR),
 DATE_FORMAT(minted_at, '%Y-%m-%d %H:%i:%s.%f')
FROM graph_scope LIMIT 2`

const ledgerObservationQuery = `SELECT r.seq, r.hash, tip.seq, tip.hash, DOLT_HASHOF('HEAD')
FROM (SELECT 1 AS anchor) AS a
LEFT JOIN (
 SELECT seq, hash FROM graph_ledger_events WHERE seq = ? LIMIT 2
) AS r ON TRUE
LEFT JOIN (
 SELECT seq, hash FROM graph_ledger_events ORDER BY seq DESC LIMIT 1
) AS tip ON TRUE
LIMIT 2`

const leaseObservationQuery = `SELECT l.row_present, l.id, l.scope_url, l.authority_id,
 l.holder_installation_key, l.renewer, CAST(l.authority_epoch AS CHAR),
 DATE_FORMAT(l.granted_at, '%Y-%m-%d %H:%i:%s.%f'),
 DATE_FORMAT(l.expires_at, '%Y-%m-%d %H:%i:%s.%f'),
 DATE_FORMAT(l.heartbeat_at, '%Y-%m-%d %H:%i:%s.%f'), l.fence,
 DATE_FORMAT(NOW(6), '%Y-%m-%d %H:%i:%s.%f'), @@session.time_zone
FROM (SELECT 1 AS anchor) AS a
LEFT JOIN (
 SELECT 1 AS row_present, id, scope_url, authority_id, holder_installation_key,
 renewer, authority_epoch, granted_at, expires_at, heartbeat_at, fence
 FROM graph_authority_lease LIMIT 2
) AS l ON TRUE LIMIT 2`

var errObservationOperand = errors.New("invalid graph observation operand")

// This cap bounds retained scalar data under the qualified B4 schema, not a
// driver or database/sql allocation before or during Scan for an altered schema.
const observationRowBytes = 16 * 1024

func observationContext(ctx context.Context) error {
	if ctx == nil {
		return errObservationDeadline
	}
	if _, ok := ctx.Deadline(); !ok {
		return errObservationDeadline
	}
	return ctx.Err()
}

// Read at most one logical row. The SQL LIMIT 2 and reader's overflow probe
// detect duplicates; query, scan, iteration, close and late cancellation errors
// discard every fact. An absent Scope is the sole valid zero-row result.
func observeRow[T any](ctx context.Context, tx queryer, query string, args []any, absent T, allowAbsent bool, decode func(*sql.Rows) (T, error)) (T, error) {
	var zero T
	if err := observationContext(ctx); err != nil {
		return zero, err
	}
	items, err := readRows(ctx, tx, query, args, 1, decode)
	if errors.Is(err, errRowOverflow) {
		err = errors.Join(err, corrupt(errors.New("duplicate precondition observation")))
	}
	if err != nil {
		return zero, err
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if len(items) == 0 {
		if allowAbsent {
			return absent, nil
		}
		return zero, corrupt(errors.New("missing anchored observation"))
	}
	return items[0], nil
}

// Scan checks retained scalar sizes after database/sql has obtained and copied
// the row; transient allocations before or during Scan are not bounded here.
// Native integer widths are normalized losslessly. Floats, booleans and
// time.Time are never coerced into protocol text.
func observationCells(rows *sql.Rows, count int) ([13]any, error) {
	var values [13]any
	if count < 1 || count > len(values) {
		return values, errObservationOperand
	}
	targets := make([]any, count)
	for i := range targets {
		targets[i] = &values[i]
	}
	if err := rows.Scan(targets...); err != nil {
		return [13]any{}, err
	}
	bytes := 0
	// Every target was constructed above as a pointer to its own value cell.
	for _, target := range targets {
		cell := target.(*any)
		normalized := normalizeObservationInteger(*cell)
		*cell = normalized
		switch v := normalized.(type) {
		case nil:
		case string:
			bytes += len(v)
		case []byte:
			bytes += len(v)
		case int64, uint64:
			bytes += 20
		default:
			return [13]any{}, corrupt(errors.New("unsupported observation scalar"))
		}
		if bytes > observationRowBytes {
			return [13]any{}, errBudget
		}
	}
	return values, nil
}

// Embedded rows preserve native integer widths in *any; the MySQL driver
// generally supplies int64/uint64. Widen without signing or rounding changes.
func normalizeObservationInteger(value any) any {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	default:
		return value
	}
}

func observationText(value any, maxBytes int) (string, error) {
	var text string
	switch v := value.(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		return "", corrupt(errors.New("missing or nontext observation scalar"))
	}
	if len(text) > maxBytes {
		return "", errBudget
	}
	if text == "" || !utf8.ValidString(text) {
		return "", corrupt(errors.New("invalid observation text"))
	}
	return text, nil
}

func observationUint(value any, max uint64) (uint64, error) {
	var text string
	switch v := value.(type) {
	case int64:
		if v <= 0 {
			return 0, corrupt(errors.New("nonpositive observation integer"))
		}
		text = strconv.FormatInt(v, 10)
	case uint64:
		text = strconv.FormatUint(v, 10)
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		return 0, corrupt(errors.New("missing or noninteger observation scalar"))
	}
	if len(text) == 0 || len(text) > 20 || text[0] == '0' {
		return 0, corrupt(errors.New("noncanonical observation integer"))
	}
	for i := range text {
		if text[i] < '0' || text[i] > '9' {
			return 0, corrupt(errors.New("nondecimal observation integer"))
		}
	}
	n, err := strconv.ParseUint(text, 10, 64)
	if err != nil || n > max {
		return 0, corrupt(errors.New("observation integer out of range"))
	}
	return n, nil
}

func observationLabel(value any, width int, last byte) (string, error) {
	text, err := observationText(value, observationRowBytes)
	if err != nil {
		return "", err
	}
	if len(text) != width {
		return "", corrupt(errors.New("invalid observation label width"))
	}
	for i := range text {
		c := text[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= last) {
			return "", corrupt(errors.New("invalid observation label alphabet"))
		}
	}
	return text, nil
}

func observationScopeURL(value any) (string, error) {
	// VARCHAR(2048) is a character bound; utf8mb4 may require 8192 bytes.
	text, err := observationText(value, 8192)
	if err != nil {
		return "", err
	}
	if err := graph.ValidatePersistedScopeURL(text); err != nil {
		return "", corrupt(err)
	}
	return text, nil
}

func observationCivil(value any) (civilTimestamp, error) {
	text, err := observationText(value, observationRowBytes)
	if err != nil {
		return "", err
	}
	const layout = "2006-01-02 15:04:05.000000"
	parsed, err := time.Parse(layout, text)
	if err != nil || len(text) != 26 || parsed.Year() < 1 || parsed.Format(layout) != text {
		return "", corrupt(errors.New("invalid observation civil timestamp"))
	}
	return civilTimestamp(text), nil
}

func observeScopeInTx(ctx context.Context, tx queryer) (scopeObservation, error) {
	return observeRow(ctx, tx, scopeObservationQuery, nil, scopeObservation{presence: observationAbsent}, true, func(rows *sql.Rows) (scopeObservation, error) {
		values, err := observationCells(rows, 5)
		if err != nil {
			return scopeObservation{}, err
		}
		id, err := observationUint(values[0], 1)
		if err != nil || id != 1 {
			return scopeObservation{}, corrupt(errors.New("invalid Scope singleton id"))
		}
		url, urlErr := observationScopeURL(values[1])
		authorityID, idErr := observationLabel(values[2], 32, 'f')
		epoch, epochErr := observationUint(values[3], ^uint64(0))
		mintedAt, timeErr := observationCivil(values[4])
		if err := errors.Join(urlErr, idErr, epochErr, timeErr); err != nil {
			return scopeObservation{}, err
		}
		return scopeObservation{observationPresent, url, authorityID, epoch, mintedAt}, nil
	})
}

func decodeLedgerPoint(seq, hash any) (ledgerPoint, error) {
	if seq == nil && hash == nil {
		return ledgerPoint{presence: observationAbsent}, nil
	}
	n, seqErr := observationUint(seq, graph.MaxLedgerSeq)
	label, hashErr := observationLabel(hash, 64, 'f')
	if err := errors.Join(seqErr, hashErr); err != nil {
		return ledgerPoint{}, err
	}
	return ledgerPoint{observationPresent, n, label}, nil
}

func observeLedgerInTx(ctx context.Context, tx queryer, requested *uint64) (ledgerObservation, error) {
	var operand any
	var requestedCopy *uint64
	if requested != nil {
		n := *requested
		if n == 0 || n > graph.MaxLedgerSeq {
			return ledgerObservation{}, errObservationOperand
		}
		requestedCopy = &n
		operand = strconv.FormatUint(n, 10)
	}
	return observeRow(ctx, tx, ledgerObservationQuery, []any{operand}, ledgerObservation{}, false, func(rows *sql.Rows) (ledgerObservation, error) {
		values, err := observationCells(rows, 5)
		if err != nil {
			return ledgerObservation{}, err
		}
		recorded, recordedErr := decodeLedgerPoint(values[0], values[1])
		tip, tipErr := decodeLedgerPoint(values[2], values[3])
		head, headErr := observationLabel(values[4], 32, 'v')
		if err := errors.Join(recordedErr, tipErr, headErr); err != nil {
			return ledgerObservation{}, err
		}
		if recorded.presence == observationPresent && (requestedCopy == nil || recorded.seq != *requestedCopy || tip.presence != observationPresent || recorded.seq > tip.seq || recorded.seq == tip.seq && recorded.hash != tip.hash) {
			return ledgerObservation{}, corrupt(errors.New("inconsistent recorded ledger observation"))
		}
		return ledgerObservation{requestedCopy, recorded, tip, head}, nil
	})
}

func observeLeaseInTx(ctx context.Context, tx queryer) (leaseObservation, error) {
	started := time.Now()
	result, err := observeRow(ctx, tx, leaseObservationQuery, nil, leaseObservation{}, false, decodeLeaseObservation)
	if err != nil {
		return leaseObservation{}, err
	}
	finished := time.Now()
	if err := ctx.Err(); err != nil {
		return leaseObservation{}, err
	}
	result.queryStarted, result.queryFinished = started, finished
	return result, nil
}

func decodeLeaseObservation(rows *sql.Rows) (leaseObservation, error) {
	values, err := observationCells(rows, 13)
	if err != nil {
		return leaseObservation{}, err
	}
	clock, clockErr := observationCivil(values[11])
	zone, zoneErr := observationText(values[12], 256)
	if err := errors.Join(clockErr, zoneErr); err != nil {
		return leaseObservation{}, err
	}
	result := leaseObservation{presence: observationAbsent, clock: clock, zone: zone}
	if values[0] == nil {
		for _, value := range values[1:11] {
			if value != nil {
				return leaseObservation{}, corrupt(errors.New("partial absent lease"))
			}
		}
		return result, nil
	}
	marker, markerErr := observationUint(values[0], 1)
	id, idErr := observationUint(values[1], 1)
	if markerErr != nil || idErr != nil || marker != 1 || id != 1 {
		return leaseObservation{}, corrupt(errors.New("invalid lease singleton identity"))
	}
	result.presence = observationPresent
	result.scopeURL, err = observationScopeURL(values[2])
	var authorityErr, holderErr, renewerErr, epochErr, grantErr, expiryErr, heartbeatErr, fenceErr error
	result.authorityID, authorityErr = observationLabel(values[3], 32, 'f')
	result.holder, holderErr = observationLabel(values[4], 64, 'f')
	result.renewer, renewerErr = observationLabel(values[5], 32, 'f')
	result.epoch, epochErr = observationUint(values[6], ^uint64(0))
	result.grantedAt, grantErr = observationCivil(values[7])
	result.expiresAt, expiryErr = observationCivil(values[8])
	result.heartbeatAt, heartbeatErr = observationCivil(values[9])
	result.fence, fenceErr = observationLabel(values[10], 32, 'f')
	if err := errors.Join(err, authorityErr, holderErr, renewerErr, epochErr, grantErr, expiryErr, heartbeatErr, fenceErr); err != nil {
		return leaseObservation{}, err
	}
	return result, nil
}

// preconditionObservations is still only data. It does not compare any fact to
// a witness/claim or confer authority on the subsequent row body.
type preconditionObservations struct {
	scope  scopeObservation
	ledger ledgerObservation
	lease  leaseObservation
	state  stateObservation
}

func observePreconditionsInTx(ctx context.Context, tx queryer, recordedSeq *uint64) (preconditionObservations, error) {
	if err := observationContext(ctx); err != nil {
		return preconditionObservations{}, err
	}
	if recordedSeq != nil {
		n := *recordedSeq
		if n == 0 || n > graph.MaxLedgerSeq {
			return preconditionObservations{}, errObservationOperand
		}
		recordedSeq = &n
	}
	var result preconditionObservations
	var err error
	result.scope, err = observeScopeInTx(ctx, tx)
	if err != nil {
		return preconditionObservations{}, err
	}
	result.ledger, err = observeLedgerInTx(ctx, tx, recordedSeq)
	if err != nil {
		return preconditionObservations{}, err
	}
	result.lease, err = observeLeaseInTx(ctx, tx)
	if err != nil {
		return preconditionObservations{}, err
	}
	result.state, err = observeStateInTx(ctx, tx)
	if err != nil {
		return preconditionObservations{}, err
	}
	return result, nil
}
