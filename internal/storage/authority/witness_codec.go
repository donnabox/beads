package authority

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strconv"
	"unicode/utf8"
)

type Phase string

const (
	Begun          Phase = "begun"
	LocalCommitted Phase = "local_committed"
	Published      Phase = "published"
	ConfigWritten  Phase = "config_written"
)

type priorBinding struct {
	Presence Presence `json:"presence"`
	Key      string   `json:"key"`
	Token    string   `json:"token"`
}
type pendingRecord struct {
	OperationKey    string            `json:"operation_installation_key"`
	Prior           priorBinding      `json:"prior_binding"`
	OperationID     string            `json:"op_id"`
	Phase           Phase             `json:"phase"`
	Intent          IntentFields      `json:"intent"`
	Prepared        PreparationFields `json:"prepared"`
	Observation     Ownership         `json:"observation"`
	OperationCommit string            `json:"op_commit"`
}
type envelope struct {
	Format          uint64         `json:"format"`
	InstallationKey string         `json:"installation_key"`
	Generation      uint64         `json:"generation"`
	Witness         *WitnessFields `json:"witness"`
	Pending         *pendingRecord `json:"pending"`
}

// Snapshot is an immutable diagnostic, never an authority assertion.
type Snapshot struct{ encoded string }

func (s Snapshot) Bytes() []byte { return []byte(s.encoded) }
func (s Snapshot) Absent() bool  { return s.encoded == "" }
func (s Snapshot) Token() string {
	if s.Absent() {
		return digest([]byte("witness:absent"))
	}
	return digest([]byte(s.encoded))
}
func (s Snapshot) record() envelope {
	var e envelope
	if json.Unmarshal([]byte(s.encoded), &e) != nil {
		return envelope{}
	}
	return e
}
func (s Snapshot) Witness() (Witness, bool) {
	w := s.record().Witness
	if w == nil {
		return Witness{}, false
	}
	return Witness{*w}, true
}
func (s Snapshot) InstallationKey() string { return s.record().InstallationKey }
func (s Snapshot) Generation() uint64      { return s.record().Generation }
func (s Snapshot) Pending() bool           { return s.record().Pending != nil }
func (s Snapshot) OperationID() string {
	p := s.record().Pending
	if p == nil {
		return ""
	}
	return p.OperationID
}
func (s Snapshot) Phase() Phase {
	p := s.record().Pending
	if p == nil {
		return ""
	}
	return p.Phase
}
func (s Snapshot) Intent() Intent {
	p := s.record().Pending
	if p == nil {
		return Intent{}
	}
	i, err := NewIntent(p.Intent)
	if err != nil {
		return Intent{}
	}
	return i
}
func (s Snapshot) Preparation() PreparationFields {
	p := s.record().Pending
	if p == nil {
		return PreparationFields{}
	}
	return p.Prepared
}
func (s Snapshot) OperationInstallationKey() string {
	p := s.record().Pending
	if p == nil {
		return ""
	}
	return p.OperationKey
}
func (s Snapshot) OperationCommit() string {
	p := s.record().Pending
	if p == nil {
		return ""
	}
	return p.OperationCommit
}

func decodeEnvelope(b []byte) (Snapshot, error) {
	if len(b) == 0 || len(b) > MaxWitnessBytes || !utf8.Valid(b) || !pairedEscapes(b) {
		return Snapshot{}, ErrInvalidRecord
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := walkJSON(dec, 0); err != nil {
		return Snapshot{}, errors.Join(ErrInvalidRecord, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return Snapshot{}, ErrInvalidRecord
	}
	var shape any
	sd := json.NewDecoder(bytes.NewReader(b))
	sd.UseNumber()
	if err := sd.Decode(&shape); err != nil {
		return Snapshot{}, ErrInvalidRecord
	}
	if !exactShape(shape, reflect.TypeFor[envelope]()) {
		return Snapshot{}, ErrInvalidRecord
	}
	var e envelope
	dec = json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return Snapshot{}, ErrInvalidRecord
	}
	if err := validateEnvelope(e); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{string(b)}, nil
}
func encodeEnvelope(e envelope) (Snapshot, error) {
	if err := validateEnvelope(e); err != nil {
		return Snapshot{}, err
	}
	b, err := json.Marshal(e)
	if err != nil {
		return Snapshot{}, ErrInvalidRecord
	}
	return decodeEnvelope(append(b, '\n'))
}
func validateEnvelope(e envelope) error {
	if e.Format != 1 {
		return ErrWitnessUnsupported
	}
	if e.Generation == 0 || !hexLength(e.InstallationKey, 64) || (e.Witness == nil && e.Pending == nil) {
		return ErrInvalidRecord
	}
	if e.Witness != nil && validateWitness(*e.Witness) != nil {
		return ErrInvalidRecord
	}
	p := e.Pending
	if p == nil {
		return nil
	}
	if !hexLength(p.OperationKey, 64) || !hexLength(p.OperationID, 32) || validateIntent(p.Intent) != nil || validatePreparation(p.Intent, p.Prepared, p.OperationKey, p.Observation) != nil {
		return ErrInvalidRecord
	}
	if prepareMatchesPriorFields(e.Witness, p.Intent, p.Prepared) != nil {
		return ErrInvalidRecord
	}
	if e.Witness == nil && admitBegin(Snapshot{}, p.OperationKey, p.Intent) != nil {
		return ErrInvalidRecord
	}
	switch p.Phase {
	case Begun, LocalCommitted, Published, ConfigWritten:
	default:
		return ErrWitnessUnsupported
	}
	if e.Witness == nil {
		if p.Prior.Presence != Absent || p.Prior.Key != "" || p.Prior.Token != (Snapshot{}).Token() || e.InstallationKey != p.OperationKey {
			return ErrInvalidRecord
		}
	} else {
		if p.Prior.Presence != Present || p.Prior.Key != e.InstallationKey || !hexLength(p.Prior.Token, 64) {
			return ErrInvalidRecord
		}
	}
	if e.InstallationKey != p.OperationKey {
		if e.Witness == nil || p.Intent.Kind != Rotate && (p.Intent.Kind != Promote || p.Intent.PromoteMode != Steal) {
			return ErrInvalidRecord
		}
	}
	eventFree := p.Intent.Kind == Adopt || p.Intent.Kind == LedgerApply || p.Intent.Kind == Promote && p.Intent.PromoteMode == SelfRegrant
	if p.Phase == Begun || eventFree {
		if p.OperationCommit != "" {
			return ErrInvalidRecord
		}
	} else if !validToken(p.OperationCommit) {
		return ErrInvalidRecord
	}
	return nil
}
func walkJSON(d *json.Decoder, depth int) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	delimiter, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	depth++
	if depth > 8 {
		return ErrInvalidRecord
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			s, ok := key.(string)
			if !ok || seen[s] {
				return ErrInvalidRecord
			}
			seen[s] = true
			if err := walkJSON(d, depth); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := walkJSON(d, depth); err != nil {
				return err
			}
		}
	default:
		return ErrInvalidRecord
	}
	_, err = d.Token()
	return err
}

// All disk members are explicit, including null union arms and false flags.
func exactShape(v any, t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		if v == nil {
			return true
		}
		return exactShape(v, t.Elem())
	}
	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok || len(m) != t.NumField() {
			return false
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			x, ok := m[f.Tag.Get("json")]
			if !ok || !exactShape(x, f.Type) {
				return false
			}
		}
		return true
	case reflect.Array:
		a, ok := v.([]any)
		if !ok || len(a) != t.Len() {
			return false
		}
		for _, x := range a {
			if !exactShape(x, t.Elem()) {
				return false
			}
		}
		return true
	case reflect.String:
		_, ok := v.(string)
		return ok
	case reflect.Bool:
		_, ok := v.(bool)
		return ok
	case reflect.Uint64:
		n, ok := v.(json.Number)
		if !ok {
			return false
		}
		_, err := strconv.ParseUint(string(n), 10, 64)
		return err == nil
	}
	return false
}

// encoding/json replaces unpaired escaped surrogates; reject that lossy input.
func pairedEscapes(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if b[i] != '\\' {
			continue
		}
		i++
		if i >= len(b) {
			return false
		}
		if b[i] != 'u' {
			continue
		}
		if i+4 >= len(b) {
			return false
		}
		n, e := strconv.ParseUint(string(b[i+1:i+5]), 16, 16)
		if e != nil {
			return false
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return false
		}
		if n >= 0xd800 && n <= 0xdbff {
			if i+6 >= len(b) || b[i+1] != '\\' || b[i+2] != 'u' {
				return false
			}
			low, e := strconv.ParseUint(string(b[i+3:i+7]), 16, 16)
			if e != nil || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}
